// Package git provides Git client implementations for twiggit
package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		return false, errors.New("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("path does not exist: %s", path)
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("context cancelled: %w", ctx.Err())
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
		return false, errors.New("path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Open the repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	// Try to get the worktree - if it fails, it's likely a bare repository
	_, err = repo.Worktree()
	if err != nil {
		// Check if the error indicates it's a bare repository
		if strings.Contains(err.Error(), "bare repository") {
			return true, nil
		}
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	return false, nil
}

// IsMainRepository checks if the path is a main git repository (not a worktree)
func (c *Client) IsMainRepository(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, errors.New("path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("context cancelled: %w", ctx.Err())
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
		return false, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get remotes
	remotes, err := repo.Remotes()
	if err != nil {
		return false, fmt.Errorf("failed to get remotes: %w", err)
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

	// Origin doesn't point to another repository in the same workspace - this is a main repository
	return true, nil
}

// ListWorktrees returns all worktrees for the given repository
func (c *Client) ListWorktrees(ctx context.Context, repoPath string) ([]*domain.WorktreeInfo, error) {
	if repoPath == "" {
		return nil, errors.New("repository path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
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
		return nil, fmt.Errorf("failed to read directory: %w", err)
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
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Get the worktree for the main repository
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get current status
	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
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
		return errors.New("repository path cannot be empty")
	}
	if branch == "" {
		return errors.New("branch name cannot be empty")
	}
	if targetPath == "" {
		return errors.New("target path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Ensure target directory doesn't exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Open the source repository first
	sourceRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open source repository: %w", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Initialize a new git repository in the target path
	worktreeRepo, err := git.PlainInit(targetPath, false)
	if err != nil {
		return fmt.Errorf("failed to initialize worktree repository: %w", err)
	}

	// Add the source repository as a remote
	_, err = worktreeRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoPath},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Fetch all branches from the source repository
	fetchOpts := &git.FetchOptions{
		RemoteName: "origin",
	}
	if err := worktreeRepo.Fetch(fetchOpts); err != nil {
		return fmt.Errorf("failed to fetch from source repository: %w", err)
	}

	// Get the worktree
	wt, err := worktreeRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
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
			return fmt.Errorf("failed to checkout existing branch: %w", err)
		}

		// Set up local branch to track remote
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create local branch: %w", err)
		}
	} else {
		// Get the HEAD reference from source repository
		head, err := sourceRepo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD from source repository: %w", err)
		}

		// Create the branch in the source repository first
		sourceWt, err := sourceRepo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get source worktree: %w", err)
		}

		err = sourceWt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
			Hash:   head.Hash(),
		})
		if err != nil {
			return fmt.Errorf("failed to create branch in source repository: %w", err)
		}

		// Switch back to the original branch in source repository
		if head.Name().IsBranch() {
			err = sourceWt.Checkout(&git.CheckoutOptions{
				Branch: head.Name(),
			})
			if err != nil {
				return fmt.Errorf("failed to switch back to original branch: %w", err)
			}
		}

		// Now create and checkout the branch in the worktree
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Create: true,
			Hash:   head.Hash(),
		})
		if err != nil {
			return fmt.Errorf("failed to create and checkout new branch in worktree: %w", err)
		}
	}

	return nil
}

// RemoveWorktree removes an existing worktree
func (c *Client) RemoveWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	if repoPath == "" {
		return errors.New("repository path cannot be empty")
	}
	if worktreePath == "" {
		return errors.New("worktree path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Check if worktree path exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}

	// For now, we'll use a simple approach: remove the directory
	// In a real implementation, we would use go-git to properly remove the worktree
	if force {
		if err := os.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	} else {
		// Check if worktree has uncommitted changes
		if c.HasUncommittedChanges(ctx, worktreePath) {
			return errors.New("worktree has uncommitted changes, use force flag to remove")
		}
		if err := os.RemoveAll(worktreePath); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}

// GetWorktreeStatus returns the status of a specific worktree
func (c *Client) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	if worktreePath == "" {
		return nil, errors.New("worktree path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}

	// Check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if worktree is git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("worktree is not a git repository: %s", worktreePath)
	}

	return c.getWorktreeStatusGoGit(ctx, worktreePath)
}

// getWorktreeStatusGoGit gets worktree status using go-git library
func (c *Client) getWorktreeStatusGoGit(_ context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	// Open the repository
	repo, err := git.PlainOpen(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get current status
	status, err := wt.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
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

// GetRepositoryRoot finds and returns the root directory of the git repository
func (c *Client) GetRepositoryRoot(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
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

	return "", errors.New("not a git repository or unable to find root: repository does not exist")
}

// GetCurrentBranch returns the name of the currently checked out branch
func (c *Client) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	if repoPath == "" {
		return "", errors.New("repository path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
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
		return nil, errors.New("repository path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get branches iterator
	branchesIter, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	var branches []string
	err = branchesIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branches, nil
}

// GetRemoteBranches returns all remote branches in the repository
func (c *Client) GetRemoteBranches(ctx context.Context, repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, errors.New("repository path cannot be empty")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get remote references
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
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
