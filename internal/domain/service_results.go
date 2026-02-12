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

// NavigationResult represents the result of a navigation operation
type NavigationResult struct {
	ResolutionResult *ResolutionResult
	Suggestions      []*ResolutionSuggestion
	ExactMatch       bool
	SearchPerformed  bool
	Context          *Context
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid   bool
	Message   string
	Warnings  []string
	Context   *Context
	Timestamp time.Time
}

// OperationResult represents a generic operation result
type OperationResult struct {
	Success   bool
	Message   string
	Timestamp time.Time
	Duration  time.Duration
	Context   *Context
	Metadata  map[string]any
}

// ListResult represents a generic list result with pagination
type ListResult[T any] struct {
	Items      []T
	TotalCount int
	HasMore    bool
	Context    *Context
	Timestamp  time.Time
}

// WorktreeOperationResult represents the result of a worktree operation
type WorktreeOperationResult struct {
	WorktreeInfo    *WorktreeInfo
	OperationResult *OperationResult
	BranchCreated   bool
	PathCreated     bool
}

// ProjectOperationResult represents the result of a project operation
type ProjectOperationResult struct {
	ProjectInfo       *ProjectInfo
	OperationResult   *OperationResult
	WorktreesAffected []string
	BranchesAffected  []string
}

// NavigationOperationResult represents the result of a navigation operation
type NavigationOperationResult struct {
	NavigationResult *NavigationResult
	OperationResult  *OperationResult
	Alternatives     []*ResolutionSuggestion
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ServiceName string
	IsHealthy   bool
	Message     string
	LastCheck   time.Time
	Metadata    map[string]any
}

// BatchResult represents the result of a batch operation
type BatchResult[T any] struct {
	Results      []Result[T]
	SuccessCount int
	ErrorCount   int
	TotalCount   int
	Duration     time.Duration
	Context      *Context
}

// NewBatchResult creates a new batch result from individual results
func NewBatchResult[T any](results []Result[T]) BatchResult[T] {
	successCount := 0
	errorCount := 0

	for _, result := range results {
		if result.IsSuccess() {
			successCount++
		} else {
			errorCount++
		}
	}

	return BatchResult[T]{
		Results:      results,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		TotalCount:   len(results),
	}
}
