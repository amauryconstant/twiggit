package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amaury/twiggit/pkg/types"
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
		return false, nil
	}

	// Check if it's a git repository
	cmd := exec.Command(c.gitCommand, "-C", path, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil, nil
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

	// Get worktree list
	cmd := exec.Command(c.gitCommand, "-C", repoPath, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

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

	// Create the worktree
	cmd := exec.Command(c.gitCommand, "-C", repoPath, "worktree", "add", targetPath, branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	return nil
}

// RemoveWorktree removes an existing worktree
func (c *Client) RemoveWorktree(repoPath, worktreePath string) error {
	if repoPath == "" {
		return fmt.Errorf("repository path cannot be empty")
	}
	if worktreePath == "" {
		return fmt.Errorf("worktree path cannot be empty")
	}

	// Remove the worktree
	cmd := exec.Command(c.gitCommand, "-C", repoPath, "worktree", "remove", worktreePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
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

	// Get branch name
	branchCmd := exec.Command(c.gitCommand, "-C", worktreePath, "branch", "--show-current")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch name: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get current commit
	commitCmd := exec.Command(c.gitCommand, "-C", worktreePath, "rev-parse", "HEAD")
	commitOutput, err := commitCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	commit := strings.TrimSpace(string(commitOutput))

	// Check if working tree is clean
	statusCmd := exec.Command(c.gitCommand, "-C", worktreePath, "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
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
