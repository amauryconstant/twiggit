package git

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"twiggit/internal/core"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	lru "github.com/hashicorp/golang-lru/v2"
)

// reader implements the read-side git operations (OpenRepository,
// ListBranches, BranchExists, RepositoryStatus, ListRemotes,
// Commit, Repository, ValidateRepository). It owns the
// LRU cache that keeps go-git Repository handles hot.
//
// reader is an internal collaborator of Client; consumers interact
// with the read-side methods through the embedded *Client.
type reader struct {
	cache          *lru.Cache[string, *git.Repository]
	isCacheEnabled bool
}

// goGitCacheFactory builds an LRU cache. Indirected so tests can inject
// failure modes for cache allocator coverage.
type goGitCacheFactory func(size int) (*lru.Cache[string, *git.Repository], error)

func defaultGoGitCacheFactory(size int) (*lru.Cache[string, *git.Repository], error) {
	return lru.New[string, *git.Repository](size)
}

// OpenRepository opens git repository (pure function, idempotent)
func (r *reader) OpenRepository(path string) (*git.Repository, error) {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, NewRepoError("open", "failed to get absolute path", err)
	}

	// Check cache first
	if repo, exists := r.cache.Get(absPath); exists {
		return repo, nil
	}

	// Open repository
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, NewRepoError("open", "failed to open git repository", err)
	}

	// Cache the repository
	r.cache.Add(absPath, repo)

	return repo, nil
}

// ListBranches lists all branches in repository (idempotent)
func (r *reader) ListBranches(_ context.Context, repoPath string) ([]core.Branch, error) {
	repo, err := r.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, NewRepoError("list-branches", "failed to list branches", err)
	}

	branchInfos := make([]core.Branch, 0)

	// Get current branch reference
	headRef, err := repo.Head()
	if err != nil {
		return nil, NewRepoError("list-branches", "failed to get HEAD reference", err)
	}

	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		if !strings.HasPrefix(branchName, "refs/") {
			branchInfo := core.Branch{
				Name:      branchName,
				IsCurrent: ref.Name() == headRef.Name(),
			}

			// Get commit info
			if commit, err := repo.CommitObject(ref.Hash()); err == nil {
				branchInfo.Commit = commit.Hash.String()
				branchInfo.Author = commit.Author.Name
				branchInfo.Date = commit.Author.When
			}

			// Check for remote tracking branch
			if ref.Name().IsBranch() {
				remoteTrackingBranch := "refs/remotes/origin/" + branchName
				if _, err := repo.Reference(plumbing.ReferenceName(remoteTrackingBranch), false); err == nil {
					branchInfo.Remote = "origin/" + branchName
				}
			}

			branchInfos = append(branchInfos, branchInfo)
		}
		return nil
	})
	if err != nil {
		return nil, NewRepoError("list-branches", "failed to iterate branches", err)
	}

	return branchInfos, nil
}

// BranchExists checks if branch exists (idempotent)
func (r *reader) BranchExists(_ context.Context, repoPath, branchName string) (bool, error) {
	repo, err := r.OpenRepository(repoPath)
	if err != nil {
		return false, err
	}

	// Try to get branch reference
	branchRefName := plumbing.ReferenceName("refs/heads/" + branchName)
	_, err = repo.Reference(branchRefName, false)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return false, nil
	}
	if err != nil {
		return false, NewRepoError("branch-exists", "failed to check branch "+branchName, err)
	}

	return true, nil
}

// RepositoryStatus returns the current working-tree status. The
// returned RepositoryStatus carries file-name slices (Modified, Added,
// Deleted, Untracked) populated from immutable strings; the slices
// themselves are owned by the returned value and may be retained by
// the caller for the lifetime of that value. Callers MUST NOT keep a
// RepositoryStatus across any subsequent RepositoryStatus call (or
// any other write to the underlying repository): the new snapshot
// supersedes the old one and stale slices must be discarded to avoid
// reporting status that no longer reflects the on-disk state.
func (r *reader) RepositoryStatus(_ context.Context, repoPath string) (core.RepositoryStatus, error) {
	repo, err := r.OpenRepository(repoPath)
	if err != nil {
		return core.RepositoryStatus{}, err
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return core.RepositoryStatus{}, NewRepoError("status", "failed to get worktree", err)
	}

	// Get status
	status, err := worktree.Status()
	if err != nil {
		return core.RepositoryStatus{}, NewRepoError("status", "failed to get repository status", err)
	}

	// Workaround for go-git worktree issue:
	// Filter out files that only appear in index (staged) but not in working tree.
	// These are baseline files from HEAD, not actual changes.
	filteredStatus := make(map[string]git.StatusCode)
	for file, entry := range status {
		if entry.Staging == git.Added && entry.Worktree == git.Unmodified {
			// File is only in index (staged), not modified in working tree
			// This is a go-git bug for worktrees - skip it
			continue
		}
		filteredStatus[file] = entry.Worktree
	}

	// Get current branch
	headRef, err := repo.Head()
	if err != nil {
		// Workaround for go-git worktree issue:
		// In worktrees, repo.Head() may fail with "reference not found"
		// even though HEAD file exists. We can still return a valid status
		// without branch/commit info.
		//nolint:nilerr // Deliberate fallback: callers get a partial status instead of a hard failure.
		return core.RepositoryStatus{
			IsClean: len(filteredStatus) == 0,
			Branch:  "unknown",
			Commit:  "",
		}, nil
	}

	repoStatus := core.RepositoryStatus{
		IsClean: len(filteredStatus) == 0,
		Branch:  headRef.Name().Short(),
		Commit:  headRef.Hash().String(),
	}

	// Categorize files (using filtered status)
	for file, worktreeStatus := range filteredStatus {
		switch worktreeStatus {
		case git.Modified:
			repoStatus.Modified = append(repoStatus.Modified, file)
		case git.Added:
			repoStatus.Added = append(repoStatus.Added, file)
		case git.Deleted:
			repoStatus.Deleted = append(repoStatus.Deleted, file)
		}

		if status[file].Staging == git.Untracked {
			repoStatus.Untracked = append(repoStatus.Untracked, file)
		}
	}

	return repoStatus, nil
}

// ValidateRepository checks if path contains valid git repository (pure function)
func (r *reader) ValidateRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err != nil {
		return NewRepoError("validate", "not a valid git repository", err)
	}
	return nil
}

// Repository returns comprehensive repository information
func (r *reader) Repository(ctx context.Context, repoPath string) (*core.Repository, error) {
	_, err := r.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get basic info
	info := &core.Repository{
		Path:   repoPath,
		IsBare: false, // go-git doesn't expose IsBare directly, assume false for worktrees
	}

	// Get branches
	branches, err := r.ListBranches(ctx, repoPath)
	if err == nil {
		info.Branches = branches
	}

	// Get remotes
	remotes, err := r.ListRemotes(ctx, repoPath)
	if err == nil {
		info.Remotes = remotes
	}

	// Get status
	status, err := r.RepositoryStatus(ctx, repoPath)
	if err == nil {
		info.Status = status
	}

	// Determine default branch
	for _, branch := range info.Branches {
		if branch.Name == "main" || branch.Name == "master" {
			info.DefaultBranch = branch.Name
			break
		}
	}

	return info, nil
}

// ListRemotes lists all remotes in repository
func (r *reader) ListRemotes(_ context.Context, repoPath string) ([]core.Remote, error) {
	repo, err := r.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, NewRepoError("list-remotes", "failed to list remotes", err)
	}

	remoteInfos := make([]core.Remote, 0, len(remotes))

	for _, remote := range remotes {
		remoteInfo := core.Remote{
			Name: remote.Config().Name,
		}

		// Get URLs
		if len(remote.Config().URLs) > 0 {
			remoteInfo.FetchURL = remote.Config().URLs[0]
			remoteInfo.PushURL = remote.Config().URLs[0]
		}

		remoteInfos = append(remoteInfos, remoteInfo)
	}

	return remoteInfos, nil
}

// Commit returns information about a specific commit
func (r *reader) Commit(_ context.Context, repoPath, commitHash string) (*core.Commit, error) {
	repo, err := r.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Parse commit hash
	hash := plumbing.NewHash(commitHash)

	// Get commit object
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, NewRepoError("commit-info", "failed to get commit "+commitHash, err)
	}

	hashStr := commit.Hash.String()
	commitInfo := &core.Commit{
		Hash:      hashStr,
		ShortHash: hashStr[:min(7, len(hashStr))],
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Date:      commit.Author.When,
		Message:   commit.Message,
	}

	return commitInfo, nil
}
