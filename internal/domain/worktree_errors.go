package domain

import (
	"errors"
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	// ErrUnknown represents an unknown error type
	ErrUnknown ErrorType = iota
	// ErrNotRepository indicates the path is not a git repository
	ErrNotRepository
	// ErrCurrentDirectory indicates an operation cannot be performed on the current directory
	ErrCurrentDirectory
	// ErrUncommittedChanges indicates there are uncommitted changes that prevent an operation
	ErrUncommittedChanges
	// ErrWorktreeExists indicates a worktree already exists at the given location
	ErrWorktreeExists
	// ErrWorktreeNotFound indicates a worktree was not found
	ErrWorktreeNotFound
	// ErrInvalidBranchName indicates an invalid branch name was provided
	ErrInvalidBranchName
	// ErrInvalidPath indicates an invalid filesystem path
	ErrInvalidPath
	// ErrPathNotWritable indicates a path is not writable
	ErrPathNotWritable
	// ErrGitCommand indicates a git command failed
	ErrGitCommand
	// ErrValidation indicates a validation error
	ErrValidation
)

// String returns a human-readable representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrNotRepository:
		return "not a git repository"
	case ErrCurrentDirectory:
		return "current directory operation error"
	case ErrUncommittedChanges:
		return "uncommitted changes detected"
	case ErrWorktreeExists:
		return "worktree already exists"
	case ErrWorktreeNotFound:
		return "worktree not found"
	case ErrInvalidBranchName:
		return "invalid branch name"
	case ErrInvalidPath:
		return "invalid path"
	case ErrPathNotWritable:
		return "path not writable"
	case ErrGitCommand:
		return "git command failed"
	case ErrValidation:
		return "validation error"
	default:
		return "unknown error"
	}
}

// WorktreeError represents a comprehensive error with context for worktree operations
type WorktreeError struct {
	Type        ErrorType
	Message     string
	Path        string
	Cause       error
	Suggestions []string
	Code        string
}

// Error implements the error interface
func (e *WorktreeError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error
func (e *WorktreeError) Unwrap() error {
	return e.Cause
}

// WithSuggestion adds a suggestion to help resolve the error
func (e *WorktreeError) WithSuggestion(suggestion string) *WorktreeError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithCode adds an error code for programmatic handling
func (e *WorktreeError) WithCode(code string) *WorktreeError {
	e.Code = code
	return e
}

// NewWorktreeError creates a new WorktreeError with the given type and message
func NewWorktreeError(errType ErrorType, message string, path string) *WorktreeError {
	return &WorktreeError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Suggestions: make([]string, 0),
	}
}

// WrapError wraps an existing error with worktree context
func WrapError(errType ErrorType, message string, path string, cause error) *WorktreeError {
	return &WorktreeError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       cause,
		Suggestions: make([]string, 0),
	}
}

// IsErrorType checks if an error is of a specific WorktreeError type
func IsErrorType(err error, errType ErrorType) bool {
	wtErr := &WorktreeError{}
	if errors.As(err, &wtErr) {
		return wtErr.Type == errType
	}
	return false
}
