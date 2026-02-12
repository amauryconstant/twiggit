package domain

import (
	"time"
)

// Result represents a generic result type following the Result/Either pattern
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a new successful result
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: value, Error: nil}
}

// NewErrorResult creates a new error result
func NewErrorResult[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Error: err}
}

// IsSuccess returns true if the result is successful
func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

// IsError returns true if the result contains an error
func (r Result[T]) IsError() bool {
	return r.Error != nil
}

// WorktreeStatus represents the status of a worktree
type WorktreeStatus struct {
	WorktreeInfo          *WorktreeInfo
	RepositoryStatus      *RepositoryStatus
	LastChecked           time.Time
	IsClean               bool
	HasUncommittedChanges bool
	BranchStatus          string // "ahead", "behind", "diverged", "up-to-date"
}

// ProjectInfo represents comprehensive project information
type ProjectInfo struct {
	Name          string
	Path          string
	GitRepoPath   string
	Worktrees     []*WorktreeInfo
	Branches      []*BranchInfo
	Remotes       []*RemoteInfo
	DefaultBranch string
	IsBare        bool
	LastModified  time.Time
}
