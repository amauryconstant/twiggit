package types

import "fmt"

// WorktreeInfo represents information about a Git worktree
type WorktreeInfo struct {
	// Path is the filesystem path to the worktree
	Path string
	// Branch is the branch name checked out in the worktree
	Branch string
	// Commit is the current commit hash
	Commit string
	// Clean indicates if the worktree has no uncommitted changes
	Clean bool
}

// Validate checks if the WorktreeInfo is valid
func (w *WorktreeInfo) Validate() error {
	if w.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if w.Branch == "" {
		return fmt.Errorf("branch cannot be empty")
	}
	return nil
}

// GitClient interface defines Git operations for worktree management
type GitClient interface {
	// IsGitRepository checks if the given path is a Git repository
	IsGitRepository(path string) (bool, error)

	// ListWorktrees returns all worktrees for the given repository
	ListWorktrees(repoPath string) ([]*WorktreeInfo, error)

	// CreateWorktree creates a new worktree from the specified branch
	CreateWorktree(repoPath, branch, targetPath string) error

	// RemoveWorktree removes an existing worktree
	RemoveWorktree(repoPath, worktreePath string) error

	// GetWorktreeStatus returns the status of a specific worktree
	GetWorktreeStatus(worktreePath string) (*WorktreeInfo, error)
}
