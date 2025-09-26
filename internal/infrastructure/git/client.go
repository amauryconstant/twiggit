// Package git provides Git client implementations for twiggit
package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// Client implements the GitClient interface using the go-git library
type Client struct{}

// NewClient creates a new Git client
func NewClient() *Client {
	return &Client{}
}

// IsGitRepository checks if the given path is a Git repository
func (c *Client) IsGitRepository(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, domain.NewWorktreeError(domain.ErrInvalidPath, "path cannot be empty", "")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, domain.NewWorktreeError(domain.ErrInvalidPath, "path does not exist", path)
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Check if it's a git repository using go-git
	_, err := git.PlainOpen(path)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// IsBareRepository checks if the given path is a bare Git repository
func (c *Client) IsBareRepository(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, domain.NewWorktreeError(domain.ErrInvalidPath, "path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", path, err)
	}

	// Try to get the worktree - if it fails, it's likely a bare repository
	_, err = repo.Worktree()
	if err != nil {
		// Check if the error indicates it's a bare repository
		if strings.Contains(err.Error(), "bare repository") {
			return true, nil
		}
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get worktree", path, err)
	}

	return false, nil
}

// IsMainRepository checks if the path is a main git repository (not a worktree)
func (c *Client) IsMainRepository(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, domain.NewWorktreeError(domain.ErrInvalidPath, "path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Check if it's a git repository first
	isRepo, err := c.IsGitRepository(ctx, path)
	if err != nil || !isRepo {
		return false, err
	}

	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", path, err)
	}

	// Get remotes
	remotes, err := repo.Remotes()
	if err != nil {
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get remotes", path, err)
	}

	// Check if there's an origin remote
	var originRemote *config.RemoteConfig
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			originRemote = remote.Config()
			break
		}
	}

	// If no origin remote, it's a main repository
	if originRemote == nil {
		return true, nil
	}

	// Check if origin remote points to another repository in the same workspace
	workspaceDir := filepath.Dir(path)
	for _, url := range originRemote.URLs {
		// Only check local file URLs (not remote URLs like GitHub/GitLab)
		isRemoteURL := strings.Contains(url, "://") || strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")

		if !isRemoteURL {
			// Remove file:// prefix if present
			cleanURL := url
			if strings.HasPrefix(url, "file://") {
				cleanURL = strings.TrimPrefix(url, "file://")
			}

			// Convert to absolute path if it's relative
			absURL := cleanURL
			if !filepath.IsAbs(cleanURL) {
				absURL = filepath.Join(path, cleanURL)
			}

			// Check if the origin URL points to a directory within the same workspace
			if strings.HasPrefix(absURL, workspaceDir) && absURL != path {
				// Origin points to another repository in the same workspace - this is a worktree
				return false, nil
			}
		}
	}

	// Origin doesn't point to another repository in the same workspace - this is a main repository
	// This includes repositories with remote origins (like GitHub/GitLab) and local repositories
	return true, nil
}

// ListWorktrees returns all worktrees for the given repository
func (c *Client) ListWorktrees(ctx context.Context, repoPath string) ([]*domain.WorktreeInfo, error) {
	if repoPath == "" {
		return nil, domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return nil, domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", repoPath, err)
	}

	var result []*domain.WorktreeInfo

	// Add main repository as a worktree
	mainWorktree, err := c.getMainWorktreeInfo(ctx, repo, repoPath)
	if err == nil {
		result = append(result, mainWorktree)
	}

	// Look for worktrees in the same directory as the repository
	parentDir := filepath.Dir(repoPath)
	worktrees, err := c.findWorktreesInDirectory(ctx, repoPath, parentDir)
	if err != nil {
		return nil, err
	}

	result = append(result, worktrees...)
	return result, nil
}

// findWorktreesInDirectory scans a directory for worktrees related to the main repository.
// It returns a slice of WorktreeInfo for all valid worktrees found, or an error if the directory cannot be read.
func (c *Client) findWorktreesInDirectory(ctx context.Context, mainRepoPath, dirPath string) ([]*domain.WorktreeInfo, error) {
	var result []*domain.WorktreeInfo

	if _, err := os.Stat(dirPath); err != nil {
		return result, nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to read directory", dirPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidatePath := filepath.Join(dirPath, entry.Name())

		// Skip the main repository itself
		if candidatePath == mainRepoPath {
			continue
		}

		// Check if this is a worktree of the main repository
		if worktreeInfo, err := c.checkCandidateWorktree(ctx, candidatePath, mainRepoPath); err == nil && worktreeInfo != nil {
			result = append(result, worktreeInfo)
		}
	}

	return result, nil
}

// checkCandidateWorktree checks if a candidate directory is a worktree of the main repository.
// It returns WorktreeInfo if the candidate is a valid worktree, nil if it's not a worktree,
// or an error if the check fails.
func (c *Client) checkCandidateWorktree(ctx context.Context, candidatePath, mainRepoPath string) (*domain.WorktreeInfo, error) {
	// Check if this is a git repository
	isGitRepo, err := c.IsGitRepository(ctx, candidatePath)
	if err != nil || !isGitRepo {
		return nil, nil
	}

	// Check if this repository has the main repo as origin
	if !c.hasOriginRemote(ctx, candidatePath, mainRepoPath) {
		return nil, nil
	}

	// Get worktree status
	worktreeInfo, err := c.GetWorktreeStatus(ctx, candidatePath)
	if err != nil {
		return nil, err
	}

	return worktreeInfo, nil
}

// hasOriginRemote checks if a repository has the specified path as its origin remote
func (c *Client) hasOriginRemote(_ context.Context, repoPath, expectedOriginPath string) bool {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return false
	}

	// Get remotes
	remotes, err := repo.Remotes()
	if err != nil {
		return false
	}

	// Check if any remote matches the expected origin path
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			for _, url := range remote.Config().URLs {
				// Compare paths, handling both file:// and direct paths
				if url == expectedOriginPath || url == "file://"+expectedOriginPath {
					return true
				}
			}
		}
	}

	return false
}

// getMainWorktreeInfo gets information about the main repository worktree
func (c *Client) getMainWorktreeInfo(ctx context.Context, repo *git.Repository, repoPath string) (*domain.WorktreeInfo, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Get the worktree for the main repository
	wt, err := repo.Worktree()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get worktree", repoPath, err)
	}

	// Get current status
	status, err := wt.Status()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get status", repoPath, err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get HEAD", repoPath, err)
	}

	// Extract branch name from HEAD reference
	branch := head.Name().Short()
	if head.Name().IsBranch() {
		branch = strings.TrimPrefix(head.Name().String(), "refs/heads/")
	}

	// Get commit timestamp
	commitTime := time.Now()
	if commitObj, err := repo.CommitObject(head.Hash()); err == nil {
		commitTime = commitObj.Author.When
	}

	return &domain.WorktreeInfo{
		Path:       repoPath,
		Branch:     branch,
		Commit:     head.Hash().String(),
		Clean:      status.IsClean(),
		CommitTime: commitTime,
	}, nil
}

// CreateWorktree creates a new worktree from the specified branch
func (c *Client) CreateWorktree(ctx context.Context, repoPath, branch, targetPath string) error {
	if repoPath == "" {
		return domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}
	if branch == "" {
		return domain.NewWorktreeError(domain.ErrValidation, "branch name cannot be empty", "")
	}
	if targetPath == "" {
		return domain.NewWorktreeError(domain.ErrInvalidPath, "target path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Ensure target directory doesn't exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return domain.NewWorktreeError(domain.ErrWorktreeExists, "target path already exists", targetPath)
	}

	// Open the source repository first
	sourceRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to open source repository", repoPath, err)
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return domain.NewWorktreeError(domain.ErrPathNotWritable, "failed to create target directory", targetPath, err)
	}

	// Initialize a new git repository in the target path
	worktreeRepo, err := git.PlainInit(targetPath, false)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to initialize worktree repository", targetPath, err)
	}

	// Add the source repository as a remote
	_, err = worktreeRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoPath},
	})
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to add remote", targetPath, err)
	}

	// Fetch all branches from the source repository
	fetchOpts := &git.FetchOptions{
		RemoteName: "origin",
	}
	if err := worktreeRepo.Fetch(fetchOpts); err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to fetch from source repository", targetPath, err)
	}

	// Get the worktree
	wt, err := worktreeRepo.Worktree()
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to get worktree", targetPath, err)
	}

	// Check if branch exists in the source repository
	branchExists := false
	branchesIter, err := sourceRepo.Branches()
	if err == nil {
		_ = branchesIter.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsBranch() && ref.Name().Short() == branch {
				branchExists = true
				return nil
			}
			return nil
		})
	}

	if branchExists {
		// Checkout existing branch from remote
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewRemoteReferenceName("origin", branch),
			Create: false,
		})
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to checkout existing branch", targetPath, err)
		}

		// Set up local branch to track remote
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
		})
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to create local branch", targetPath, err)
		}
	} else {
		// Get the HEAD reference from source repository
		head, err := sourceRepo.Head()
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to get HEAD from source repository", repoPath, err)
		}

		// Create the branch in the source repository first
		sourceWt, err := sourceRepo.Worktree()
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to get source worktree", repoPath, err)
		}

		err = sourceWt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
			Hash:   head.Hash(),
		})
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to create branch in source repository", repoPath, err)
		}

		// Switch back to the original branch in source repository
		if head.Name().IsBranch() {
			err = sourceWt.Checkout(&git.CheckoutOptions{
				Branch: head.Name(),
			})
			if err != nil {
				return domain.NewWorktreeError(domain.ErrGitCommand, "failed to switch back to original branch", repoPath, err)
			}
		}

		// Now create and checkout the branch in the worktree
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
			Hash:   head.Hash(),
		})
		if err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to create and checkout new branch in worktree", targetPath, err)
		}
	}

	return nil
}

// RemoveWorktree removes an existing worktree
func (c *Client) RemoveWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	if repoPath == "" {
		return domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}
	if worktreePath == "" {
		return domain.NewWorktreeError(domain.ErrInvalidPath, "worktree path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Check if worktree path exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return domain.NewWorktreeError(domain.ErrWorktreeNotFound, "worktree path does not exist", worktreePath)
	}

	// For now, we'll use a simple approach: remove the directory
	// In a real implementation, we would use go-git to properly remove the worktree
	if force {
		if err := os.RemoveAll(worktreePath); err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to remove worktree directory", worktreePath, err)
		}
	} else {
		// Check if worktree has uncommitted changes
		if c.HasUncommittedChanges(ctx, worktreePath) {
			return domain.NewWorktreeError(domain.ErrUncommittedChanges, "worktree has uncommitted changes, use force flag to remove", worktreePath)
		}
		if err := os.RemoveAll(worktreePath); err != nil {
			return domain.NewWorktreeError(domain.ErrGitCommand, "failed to remove worktree directory", worktreePath, err)
		}
	}

	return nil
}

// GetWorktreeStatus returns the status of a specific worktree
func (c *Client) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	if worktreePath == "" {
		return nil, domain.NewWorktreeError(domain.ErrInvalidPath, "worktree path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return nil, domain.NewWorktreeError(domain.ErrWorktreeNotFound, "worktree path does not exist", worktreePath)
	}

	// Check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check if worktree is git repository", worktreePath, err)
	}
	if !isRepo {
		return nil, domain.NewWorktreeError(domain.ErrNotRepository, "worktree is not a git repository", worktreePath)
	}

	// Try go-git first
	result, err := c.getWorktreeStatusGoGit(ctx, worktreePath)
	if err == nil {
		return result, nil
	}

	// If go-git fails, try CLI fallback
	return c.getWorktreeStatusCLI(ctx, worktreePath)
}

// getWorktreeStatusGoGit gets worktree status using go-git library
func (c *Client) getWorktreeStatusGoGit(_ context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	// Open the repository
	repo, err := git.PlainOpen(worktreePath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", worktreePath, err)
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get worktree", worktreePath, err)
	}

	// Get current status
	status, err := wt.Status()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get status", worktreePath, err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get HEAD", worktreePath, err)
	}

	// Extract branch name from HEAD reference
	branch := head.Name().Short()
	if head.Name().IsBranch() {
		branch = strings.TrimPrefix(head.Name().String(), "refs/heads/")
	}

	// Get commit timestamp
	commitTime := time.Now()
	if commitObj, err := repo.CommitObject(head.Hash()); err == nil {
		commitTime = commitObj.Author.When
	}

	return &domain.WorktreeInfo{
		Path:       worktreePath,
		Branch:     branch,
		Commit:     head.Hash().String(),
		Clean:      status.IsClean(),
		CommitTime: commitTime,
	}, nil
}

// getWorktreeStatusCLI gets worktree status using git CLI commands as fallback
func (c *Client) getWorktreeStatusCLI(_ context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	// Get current branch name
	branchCmd := exec.Command("git", "-C", worktreePath, "symbolic-ref", "--short", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get branch name", worktreePath, err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get commit hash
	commitCmd := exec.Command("git", "-C", worktreePath, "rev-parse", "HEAD")
	commitOutput, err := commitCmd.Output()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get commit hash", worktreePath, err)
	}
	commit := strings.TrimSpace(string(commitOutput))

	// Get commit timestamp
	timeCmd := exec.Command("git", "-C", worktreePath, "log", "-1", "--format=%ct", "HEAD")
	timeOutput, err := timeCmd.Output()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get commit time", worktreePath, err)
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(timeOutput)), 10, 64)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to parse commit time", worktreePath, err)
	}
	commitTime := time.Unix(timestamp, 0)

	// Check if working directory is clean
	statusCmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get status", worktreePath, err)
	}
	clean := len(strings.TrimSpace(string(statusOutput))) == 0

	return &domain.WorktreeInfo{
		Path:       worktreePath,
		Branch:     branch,
		Commit:     commit,
		Clean:      clean,
		CommitTime: commitTime,
	}, nil
}

// GetRepositoryRoot finds and returns the root directory of the git repository
func (c *Client) GetRepositoryRoot(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", domain.NewWorktreeError(domain.ErrInvalidPath, "path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", domain.NewWorktreeError(domain.ErrInvalidPath, "path does not exist", path)
	}

	// Traverse up the directory structure to find the repository root
	currentPath := path
	for {
		// Check if current path is a git repository
		_, err := git.PlainOpen(currentPath)
		if err == nil {
			// Found the repository root
			return currentPath, nil
		}

		// Get the parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// We've reached the root directory and haven't found a repository
			break
		}
		currentPath = parentPath
	}

	return "", domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository or unable to find root", path)
}

// GetCurrentBranch returns the name of the currently checked out branch
func (c *Client) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	if repoPath == "" {
		return "", domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return "", domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return "", domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", repoPath, err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return "", domain.NewWorktreeError(domain.ErrGitCommand, "failed to get HEAD", repoPath, err)
	}

	// Extract branch name from HEAD reference
	branch := head.Name().Short()
	if head.Name().IsBranch() {
		branch = strings.TrimPrefix(head.Name().String(), "refs/heads/")
	}

	return branch, nil
}

// GetAllBranches returns all local branches in the repository
func (c *Client) GetAllBranches(ctx context.Context, repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return nil, domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", repoPath, err)
	}

	// Get branches iterator
	branchesIter, err := repo.Branches()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get branches", repoPath, err)
	}

	var branches []string
	err = branchesIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})

	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to iterate branches", repoPath, err)
	}

	return branches, nil
}

// GetRemoteBranches returns all remote branches in the repository
func (c *Client) GetRemoteBranches(ctx context.Context, repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return nil, domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", repoPath, err)
	}

	// Get remote references
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, domain.NewWorktreeError(domain.ErrGitCommand, "failed to get remotes", repoPath, err)
	}

	var branches []string
	for _, remote := range remotes {
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			continue // Skip remotes that can't be listed
		}

		for _, ref := range refs {
			if ref.Name().IsRemote() && !strings.Contains(ref.Name().String(), "HEAD") {
				branches = append(branches, ref.Name().Short())
			}
		}
	}

	return branches, nil
}

// BranchExists checks if a branch exists in the repository
func (c *Client) BranchExists(ctx context.Context, repoPath, branch string) bool {
	if repoPath == "" || branch == "" {
		return false
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil || !isRepo {
		return false
	}

	// Check local branches first
	branches, err := c.GetAllBranches(ctx, repoPath)
	if err == nil {
		for _, b := range branches {
			if b == branch {
				return true
			}
		}
	}

	// Check remote branches
	remoteBranches, err := c.GetRemoteBranches(ctx, repoPath)
	if err == nil {
		for _, b := range remoteBranches {
			// Extract branch name from remote/branch format
			parts := strings.Split(b, "/")
			if len(parts) >= 2 && parts[len(parts)-1] == branch {
				return true
			}
		}
	}

	return false
}

// HasUncommittedChanges checks if the repository has any uncommitted changes
func (c *Client) HasUncommittedChanges(ctx context.Context, repoPath string) bool {
	if repoPath == "" {
		return false
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil || !isRepo {
		return false
	}

	return c.hasUncommittedChangesGoGit(ctx, repoPath)
}

// hasUncommittedChangesGoGit checks for uncommitted changes using go-git
func (c *Client) hasUncommittedChangesGoGit(_ context.Context, repoPath string) bool {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return false
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		return false
	}

	// Get status
	status, err := wt.Status()
	if err != nil {
		return false
	}

	// Check if status is clean
	return !status.IsClean()
}

// DeleteBranch deletes a branch from the repository
func (c *Client) DeleteBranch(ctx context.Context, repoPath, branch string) error {
	if repoPath == "" {
		return domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}
	if branch == "" {
		return domain.NewWorktreeError(domain.ErrValidation, "branch name cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Check if branch exists
	if !c.BranchExists(ctx, repoPath, branch) {
		return domain.NewWorktreeError(domain.ErrValidation, "branch does not exist", branch)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to open repository", repoPath, err)
	}

	// Get the current branch to avoid deleting the currently checked out branch
	currentBranch, err := c.GetCurrentBranch(ctx, repoPath)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to get current branch", repoPath, err)
	}

	if currentBranch == branch {
		return domain.NewWorktreeError(domain.ErrValidation, "cannot delete currently checked out branch", branch)
	}

	// Delete the branch using go-git
	branchRef := plumbing.NewBranchReferenceName(branch)
	err = repo.Storer.RemoveReference(branchRef)
	if err != nil {
		return domain.NewWorktreeError(domain.ErrGitCommand, "failed to delete branch", branch, err)
	}

	return nil
}

// IsBranchMerged checks if a branch has been merged into the current branch
func (c *Client) IsBranchMerged(ctx context.Context, repoPath, branch string) (bool, error) {
	if repoPath == "" {
		return false, domain.NewWorktreeError(domain.ErrInvalidPath, "repository path cannot be empty", "")
	}
	if branch == "" {
		return false, domain.NewWorktreeError(domain.ErrValidation, "branch name cannot be empty", "")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, domain.NewWorktreeError(domain.ErrValidation, "context cancelled", "", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check repository", repoPath, err)
	}
	if !isRepo {
		return false, domain.NewWorktreeError(domain.ErrNotRepository, "not a git repository", repoPath)
	}

	// Check if branch exists
	if !c.BranchExists(ctx, repoPath, branch) {
		return false, domain.NewWorktreeError(domain.ErrValidation, "branch does not exist", branch)
	}

	// Use git CLI to check if branch is merged (more reliable than go-git for this operation)
	cmd := exec.Command("git", "-C", repoPath, "merge-base", "--is-ancestor", branch, "HEAD")
	err = cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			// Exit code 1 means branch is not merged
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, domain.NewWorktreeError(domain.ErrGitCommand, "failed to check if branch is merged", branch, err)
	}

	// If command succeeded, branch is merged
	return true, nil
}
