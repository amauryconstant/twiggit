// Package domain contains core business entities and interfaces for twiggit
package domain

import (
	"context"
	"errors"
	"time"
)

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
	// CommitTime is the timestamp when the commit was created
	CommitTime time.Time
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
	IsGitRepository(ctx context.Context, path string) (bool, error)
	IsBareRepository(ctx context.Context, path string) (bool, error)
	IsMainRepository(ctx context.Context, path string) (bool, error)
	GetRepositoryRoot(ctx context.Context, path string) (string, error)

	// Worktree operations
	ListWorktrees(ctx context.Context, repoPath string) ([]*WorktreeInfo, error)
	CreateWorktree(ctx context.Context, repoPath, branch, targetPath string) error
	RemoveWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error
	GetWorktreeStatus(ctx context.Context, worktreePath string) (*WorktreeInfo, error)

	// Branch operations
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)
	GetAllBranches(ctx context.Context, repoPath string) ([]string, error)
	GetRemoteBranches(ctx context.Context, repoPath string) ([]string, error)
	BranchExists(ctx context.Context, repoPath, branch string) bool

	// Status operations
	HasUncommittedChanges(ctx context.Context, repoPath string) bool
}
