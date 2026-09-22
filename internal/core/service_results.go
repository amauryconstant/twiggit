package core

import (
	"reflect"
	"time"
)

// Result represents a generic result type following the Result/Either pattern
type Result[T any] struct {
	Value T
	Error error
}

// NewResult creates a new successful result.
// When T is a slice type, the underlying backing array is cloned to
// prevent callers from mutating internal state through the returned struct.
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: cloneSlice(value), Error: nil}
}

// NewErrResult creates a new error result
func NewErrResult[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Error: err}
}

// cloneSlice returns a copy of the slice when value is a slice type;
// non-slice values pass through untouched. Defensive copy avoids the
// backing-array-aliasing trap when the caller mutates the slice after
// the result has been stored.
func cloneSlice[T any](value T) T {
	v := reflect.ValueOf(value)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return value
	}
	if v.IsNil() {
		return value
	}
	dst := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
	reflect.Copy(dst, v)
	return dst.Interface().(T) //nolint:errcheck // reflect.Copy guarantees non-nil dst; type assertion to T cannot fail
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

// ProjectSummary represents lightweight project information without expensive git data
type ProjectSummary struct {
	Name        string
	Path        string
	GitRepoPath string
}

// CreateWorktreeResult represents the result of a worktree creation operation
type CreateWorktreeResult struct {
	Worktree   *WorktreeInfo
	HookResult *HookResult
}
