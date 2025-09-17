// Package domain contains core business entities and interfaces for twiggit
package domain

import "errors"

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
		return errors.New("path cannot be empty")
	}
	if w.Branch == "" {
		return errors.New("branch cannot be empty")
	}
	return nil
}

// GitClient interface defines Git operations for worktree management
type GitClient interface {
	// Repository operations
	IsGitRepository(path string) (bool, error)
	IsMainRepository(path string) (bool, error)
	GetRepositoryRoot(path string) (string, error)

	// Worktree operations
	ListWorktrees(repoPath string) ([]*WorktreeInfo, error)
	CreateWorktree(repoPath, branch, targetPath string) error
	RemoveWorktree(repoPath, worktreePath string, force bool) error
	GetWorktreeStatus(worktreePath string) (*WorktreeInfo, error)

	// Branch operations
	GetCurrentBranch(repoPath string) (string, error)
	GetAllBranches(repoPath string) ([]string, error)
	GetRemoteBranches(repoPath string) ([]string, error)
	BranchExists(repoPath, branch string) bool

	// Status operations
	HasUncommittedChanges(repoPath string) bool
}
