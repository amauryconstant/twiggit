package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/amaury/twiggit/pkg/types"
	"github.com/go-git/go-git/v5"
)

// Client implements the GitClient interface using the git command-line tool
type Client struct {
	gitCommand string
}

// NewClient creates a new Git client
func NewClient() *Client {
	return &Client{
		gitCommand: "git",
	}
}

// IsGitRepository checks if the given path is a Git repository
func (c *Client) IsGitRepository(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("path does not exist: %s", path)
	}

	// Check if it's a git repository using go-git
	_, err := git.PlainOpen(path)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// IsMainRepository checks if the path is a main git repository (not a worktree)
func (c *Client) IsMainRepository(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("path cannot be empty")
	}

	// Check if it's a git repository first
	isRepo, err := c.IsGitRepository(path)
	if err != nil || !isRepo {
		return false, err
	}

	// Check if .git is a directory (main repo) or file (worktree)
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false, err
	}

	// If .git is a directory, it's a main repository
	// If .git is a file, it's a worktree
	return info.IsDir(), nil
}

// ListWorktrees returns all worktrees for the given repository
func (c *Client) ListWorktrees(repoPath string) ([]*types.WorktreeInfo, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repository path cannot be empty")
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// List worktrees using git command
	cmd := exec.Command(c.gitCommand, "-C", repoPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Parse the output
	return c.parseWorktreeList(string(output))
}

// CreateWorktree creates a new worktree from the specified branch
func (c *Client) CreateWorktree(repoPath, branch, targetPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repository path cannot be empty")
	}
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if targetPath == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	// Ensure target directory doesn't exist
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("target path already exists: %s", targetPath)
	}

	// Check if branch exists to determine the correct command
	branchExists := c.BranchExists(repoPath, branch)

	var cmd *exec.Cmd
	if branchExists {
		// Branch exists, create worktree from existing branch
		cmd = exec.Command(c.gitCommand, "-C", repoPath, "worktree", "add", targetPath, branch)
	} else {
		// Branch doesn't exist, create new branch from current HEAD
		cmd = exec.Command(c.gitCommand, "-C", repoPath, "worktree", "add", "-b", branch, targetPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %v, output: %s", err, string(output))
	}

	return nil
}

// RemoveWorktree removes an existing worktree
func (c *Client) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	if repoPath == "" {
		return fmt.Errorf("repository path cannot be empty")
	}
	if worktreePath == "" {
		return fmt.Errorf("worktree path cannot be empty")
	}

	// Check if worktree path exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}

	// Build the command with optional force flag
	args := []string{"-C", repoPath, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	cmd := exec.Command(c.gitCommand, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %v, output: %s", err, string(output))
	}

	return nil
}

// GetWorktreeStatus returns the status of a specific worktree
func (c *Client) GetWorktreeStatus(worktreePath string) (*types.WorktreeInfo, error) {
	if worktreePath == "" {
		return nil, fmt.Errorf("worktree path cannot be empty")
	}

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}

	// Check if it's a git repository
	isRepo, err := c.IsGitRepository(worktreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if worktree is git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("worktree is not a git repository: %s", worktreePath)
	}

	// Get current branch using git CLI (more reliable for worktrees)
	branchCmd := exec.Command(c.gitCommand, "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = worktreePath
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get current commit hash
	commitCmd := exec.Command(c.gitCommand, "rev-parse", "HEAD")
	commitCmd.Dir = worktreePath
	commitOutput, err := commitCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}
	commit := strings.TrimSpace(string(commitOutput))

	// Check if working tree is clean
	statusCmd := exec.Command(c.gitCommand, "status", "--porcelain")
	statusCmd.Dir = worktreePath
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status: %w", err)
	}
	clean := len(strings.TrimSpace(string(statusOutput))) == 0

	return &types.WorktreeInfo{
		Path:   worktreePath,
		Branch: branch,
		Commit: commit,
		Clean:  clean,
	}, nil
}

// parseWorktreeList parses the output of `git worktree list --porcelain`
func (c *Client) parseWorktreeList(output string) ([]*types.WorktreeInfo, error) {
	lines := strings.Split(output, "\n")
	var worktrees []*types.WorktreeInfo
	var current *types.WorktreeInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				worktrees = append(worktrees, current)
			}
			current = &types.WorktreeInfo{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		}
	}

	// Add the last worktree if exists
	if current != nil {
		worktrees = append(worktrees, current)
	}

	// Get status for each worktree
	for _, wt := range worktrees {
		if status, err := c.GetWorktreeStatus(wt.Path); err == nil {
			wt.Clean = status.Clean
		}
	}

	return worktrees, nil
}

// GetRepositoryRoot finds and returns the root directory of the git repository
func (c *Client) GetRepositoryRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	// Use git rev-parse to find the repository root
	cmd := exec.Command(c.gitCommand, "rev-parse", "--show-toplevel")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository or unable to find root: %w", err)
	}

	root := strings.TrimSpace(string(output))
	return root, nil
}

// GetCurrentBranch returns the name of the currently checked out branch
func (c *Client) GetCurrentBranch(repoPath string) (string, error) {
	if repoPath == "" {
		return "", fmt.Errorf("repository path cannot be empty")
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Use git command to get current branch
	cmd := exec.Command(c.gitCommand, "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	// If we're in detached HEAD state, branch will be "HEAD"
	// In this case, we could try other methods, but for now we'll return "HEAD"

	return branch, nil
}

// GetAllBranches returns all local branches in the repository
func (c *Client) GetAllBranches(repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repository path cannot be empty")
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Use git command to list all local branches
	cmd := exec.Command(c.gitCommand, "branch", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// GetRemoteBranches returns all remote branches in the repository
func (c *Client) GetRemoteBranches(repoPath string) ([]string, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repository path cannot be empty")
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Use git command to list all remote branches
	cmd := exec.Command(c.gitCommand, "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// If there are no remotes, this is not an error
		if strings.Contains(err.Error(), "fatal: No remote refs") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" && !strings.Contains(branch, "HEAD ->") {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// BranchExists checks if a branch exists in the repository
func (c *Client) BranchExists(repoPath, branch string) bool {
	if repoPath == "" || branch == "" {
		return false
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil || !isRepo {
		return false
	}

	// Check local branches first
	branches, err := c.GetAllBranches(repoPath)
	if err == nil {
		for _, b := range branches {
			if b == branch {
				return true
			}
		}
	}

	// Check remote branches
	remoteBranches, err := c.GetRemoteBranches(repoPath)
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
func (c *Client) HasUncommittedChanges(repoPath string) bool {
	if repoPath == "" {
		return false
	}

	// First check if it's a git repository
	isRepo, err := c.IsGitRepository(repoPath)
	if err != nil || !isRepo {
		return false
	}

	// Use git status --porcelain to check for changes
	cmd := exec.Command(c.gitCommand, "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		// If we can't check status, assume there are no changes
		return false
	}

	// If output is not empty, there are uncommitted changes
	return strings.TrimSpace(string(output)) != ""
}
