package infrastructure

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"twiggit/internal/domain"
)

// GoGitClientImpl implements GoGitClient using go-git library
type GoGitClientImpl struct {
	cache        map[string]*git.Repository // Simple in-memory cache
	cacheEnabled bool
}

// NewGoGitClient creates a new GoGitClient implementation
func NewGoGitClient(cacheEnabled ...bool) *GoGitClientImpl {
	enabled := true
	if len(cacheEnabled) > 0 {
		enabled = cacheEnabled[0]
	}

	return &GoGitClientImpl{
		cache:        make(map[string]*git.Repository),
		cacheEnabled: enabled,
	}
}

// OpenRepository opens git repository (pure function, idempotent)
func (c *GoGitClientImpl) OpenRepository(path string) (*git.Repository, error) {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, domain.NewGitRepositoryError(path, "failed to get absolute path", err)
	}

	// Check cache first
	if repo, exists := c.cache[absPath]; exists {
		return repo, nil
	}

	// Open repository
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, domain.NewGitRepositoryError(path, "failed to open git repository", err)
	}

	// Cache the repository
	c.cache[absPath] = repo

	return repo, nil
}

// ListBranches lists all branches in repository (idempotent)
func (c *GoGitClientImpl) ListBranches(_ context.Context, repoPath string) ([]domain.BranchInfo, error) {
	repo, err := c.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to list branches", err)
	}

	var branchInfos []domain.BranchInfo

	// Get current branch reference
	headRef, err := repo.Head()
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to get HEAD reference", err)
	}

	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		if !strings.HasPrefix(branchName, "refs/") {
			branchInfo := domain.BranchInfo{
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
				if remoteRef, err := repo.Reference(plumbing.ReferenceName(remoteTrackingBranch), false); err == nil {
					branchInfo.Remote = "origin/" + branchName
					_ = remoteRef // Avoid unused variable warning
				}
			}

			branchInfos = append(branchInfos, branchInfo)
		}
		return nil
	})

	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to iterate branches", err)
	}

	return branchInfos, nil
}

// BranchExists checks if branch exists (idempotent)
func (c *GoGitClientImpl) BranchExists(_ context.Context, repoPath, branchName string) (bool, error) {
	repo, err := c.OpenRepository(repoPath)
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
		return false, domain.NewGitRepositoryError(repoPath, "failed to check branch "+branchName, err)
	}

	return true, nil
}

// GetRepositoryStatus returns repository status (idempotent)
func (c *GoGitClientImpl) GetRepositoryStatus(_ context.Context, repoPath string) (domain.RepositoryStatus, error) {
	repo, err := c.OpenRepository(repoPath)
	if err != nil {
		return domain.RepositoryStatus{}, err
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return domain.RepositoryStatus{}, domain.NewGitRepositoryError(repoPath, "failed to get worktree", err)
	}

	// Get status
	status, err := worktree.Status()
	if err != nil {
		return domain.RepositoryStatus{}, domain.NewGitRepositoryError(repoPath, "failed to get repository status", err)
	}

	// Get current branch
	headRef, err := repo.Head()
	if err != nil {
		return domain.RepositoryStatus{}, domain.NewGitRepositoryError(repoPath, "failed to get HEAD reference", err)
	}

	repoStatus := domain.RepositoryStatus{
		IsClean: status.IsClean(),
		Branch:  headRef.Name().Short(),
		Commit:  headRef.Hash().String(),
	}

	// Categorize files
	for file, entry := range status {
		switch entry.Worktree {
		case git.Modified:
			repoStatus.Modified = append(repoStatus.Modified, file)
		case git.Added:
			repoStatus.Added = append(repoStatus.Added, file)
		case git.Deleted:
			repoStatus.Deleted = append(repoStatus.Deleted, file)
		}

		if entry.Staging == git.Untracked {
			repoStatus.Untracked = append(repoStatus.Untracked, file)
		}
	}

	return repoStatus, nil
}

// ValidateRepository checks if path contains valid git repository (pure function)
func (c *GoGitClientImpl) ValidateRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err != nil {
		return domain.NewGitRepositoryError(path, "not a valid git repository", err)
	}
	return nil
}

// GetRepositoryInfo returns comprehensive repository information
func (c *GoGitClientImpl) GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
	_, err := c.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get basic info
	info := &domain.GitRepository{
		Path:   repoPath,
		IsBare: false, // go-git doesn't expose IsBare directly, assume false for worktrees
	}

	// Get branches
	branches, err := c.ListBranches(ctx, repoPath)
	if err == nil {
		info.Branches = branches
	}

	// Get remotes
	remotes, err := c.ListRemotes(ctx, repoPath)
	if err == nil {
		info.Remotes = remotes
	}

	// Get status
	status, err := c.GetRepositoryStatus(ctx, repoPath)
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
func (c *GoGitClientImpl) ListRemotes(_ context.Context, repoPath string) ([]domain.RemoteInfo, error) {
	repo, err := c.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to list remotes", err)
	}

	remoteInfos := make([]domain.RemoteInfo, 0, len(remotes))

	for _, remote := range remotes {
		remoteInfo := domain.RemoteInfo{
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

// GetCommitInfo returns information about a specific commit
func (c *GoGitClientImpl) GetCommitInfo(_ context.Context, repoPath, commitHash string) (*domain.CommitInfo, error) {
	repo, err := c.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Parse commit hash
	hash := plumbing.NewHash(commitHash)

	// Get commit object
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to get commit "+commitHash, err)
	}

	commitInfo := &domain.CommitInfo{
		Hash:      commit.Hash.String(),
		ShortHash: commit.Hash.String()[:7],
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Date:      commit.Author.When,
		Message:   commit.Message,
	}

	return commitInfo, nil
}
