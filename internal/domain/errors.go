package domain

import (
	"errors"
	"fmt"
)

// DomainErrorType represents different categories of domain errors.
// This unified type provides consistent error handling across all domain operations
// including projects, worktrees, and workspaces. Each error type corresponds to
// a specific category of failure that can occur during system operations.
type DomainErrorType int

const (
	// ErrUnknown represents an unknown error type
	ErrUnknown DomainErrorType = iota

	// ErrInvalidProjectName represents an invalid project name error
	ErrInvalidProjectName
	// ErrInvalidGitRepoPath represents an invalid git repository path error
	ErrInvalidGitRepoPath
	// ErrProjectNotFound represents a project not found error
	ErrProjectNotFound
	// ErrProjectAlreadyExists represents a project already exists error
	ErrProjectAlreadyExists
	// ErrProjectValidation represents a project validation error
	ErrProjectValidation

	// ErrNotRepository represents a not a git repository error
	ErrNotRepository
	// ErrCurrentDirectory represents a current directory operation error
	ErrCurrentDirectory
	// ErrUncommittedChanges represents an uncommitted changes detected error
	ErrUncommittedChanges
	// ErrWorktreeExists represents a worktree already exists error
	ErrWorktreeExists
	// ErrWorktreeNotFound represents a worktree not found error
	ErrWorktreeNotFound
	// ErrInvalidBranchName represents an invalid branch name error
	ErrInvalidBranchName
	// ErrInvalidPath represents an invalid path error
	ErrInvalidPath
	// ErrPathNotWritable represents a path not writable error
	ErrPathNotWritable
	// ErrGitCommand represents a git command failed error
	ErrGitCommand

	// ErrWorkspaceInvalidPath represents a workspace path invalid error
	ErrWorkspaceInvalidPath
	// ErrWorkspaceProjectNotFound represents a workspace project not found error
	ErrWorkspaceProjectNotFound
	// ErrWorkspaceProjectAlreadyExists represents a workspace project already exists error
	ErrWorkspaceProjectAlreadyExists
	// ErrWorkspaceWorktreeNotFound represents a workspace worktree not found error
	ErrWorkspaceWorktreeNotFound
	// ErrWorkspaceInvalidConfiguration represents a workspace configuration invalid error
	ErrWorkspaceInvalidConfiguration
	// ErrWorkspaceDiscoveryFailed represents a workspace discovery failed error
	ErrWorkspaceDiscoveryFailed
	// ErrWorkspaceValidationFailed represents a workspace validation failed error
	ErrWorkspaceValidationFailed

	// ErrValidation represents a validation error
	ErrValidation
)

// String returns a human-readable representation of the domain error type
func (et DomainErrorType) String() string {
	switch et {
	// Project error types
	case ErrInvalidProjectName:
		return "invalid project name"
	case ErrInvalidGitRepoPath:
		return "invalid git repository path"
	case ErrProjectNotFound:
		return "project not found"
	case ErrProjectAlreadyExists:
		return "project already exists"
	case ErrProjectValidation:
		return "project validation error"

	// Worktree error types
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

	// Workspace error types
	case ErrWorkspaceInvalidPath:
		return "workspace path invalid"
	case ErrWorkspaceProjectNotFound:
		return "workspace project not found"
	case ErrWorkspaceProjectAlreadyExists:
		return "workspace project already exists"
	case ErrWorkspaceWorktreeNotFound:
		return "workspace worktree not found"
	case ErrWorkspaceInvalidConfiguration:
		return "workspace configuration invalid"
	case ErrWorkspaceDiscoveryFailed:
		return "workspace discovery failed"
	case ErrWorkspaceValidationFailed:
		return "workspace validation failed"

	// Generic error types
	case ErrValidation:
		return "validation error"
	default:
		return "unknown error"
	}
}

// DomainError represents a comprehensive error with context for all domain operations.
// This unified type replaces ProjectError, WorktreeError, and WorkspaceError.
// It provides rich context including error type, message, path, cause, suggestions,
// and entity type for better error handling and user experience.
type DomainError struct {
	Type        DomainErrorType
	Message     string
	Path        string
	Cause       error
	Suggestions []string
	EntityType  string // "project", "worktree", "workspace" for context
}

// Error implements the error interface.
// Returns a formatted error message that includes the error type, message,
// and path if available. This provides consistent error string representation
// across all domain operations.
func (e *DomainError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error.
// This enables error wrapping and allows callers to inspect the root cause
// using errors.Is() and errors.As() functions from the standard library.
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithSuggestion adds a suggestion to help resolve the error.
// Suggestions are displayed to users in formatted error messages to provide
// actionable guidance for resolving the error. Multiple suggestions can be added.
func (e *DomainError) WithSuggestion(suggestion string) *DomainError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithEntityType adds entity type context for better error categorization.
// Entity type helps categorize errors by domain concept ("project", "worktree", "workspace")
// and enables more targeted error handling and user messaging.
func (e *DomainError) WithEntityType(entityType string) *DomainError {
	e.EntityType = entityType
	return e
}

// IsDomainErrorType checks if an error is of a specific DomainError type.
// This function enables type-safe error checking and is useful for error handling
// logic that needs to respond differently based on error categories.
// Returns true if the error matches the specified type, false otherwise.
func IsDomainErrorType(err error, errType DomainErrorType) bool {
	domainErr := &DomainError{}
	if errors.As(err, &domainErr) {
		return domainErr.Type == errType
	}
	return false
}

// Domain-specific constructors that preserve ubiquitous language

// NewProjectError creates a new DomainError with project context.
// Use this function for all project-related errors to ensure consistent
// error handling and user experience. Automatically sets entity type to "project".
// Accepts an optional cause error for wrapping underlying errors.
//
// Example:
//
//	err := NewProjectError(ErrProjectNotFound, "project not found", path)
//	    .WithSuggestion("Check if project exists in projects directory")
//
//	// With underlying cause:
//	err := NewProjectError(ErrProjectValidation, "validation failed", path, underlyingErr)
//	    .WithSuggestion("Check project configuration")
func NewProjectError(errType DomainErrorType, message string, path string, cause ...error) *DomainError {
	var err error
	if len(cause) > 0 {
		err = cause[0]
	}

	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       err,
		Suggestions: make([]string, 0),
		EntityType:  "project",
	}
}

// NewWorktreeError creates a new DomainError with worktree context.
// Use this function for all worktree-related errors to ensure consistent
// error handling and user experience. Automatically sets entity type to "worktree".
// Accepts an optional cause error for wrapping underlying errors.
//
// Example:
//
//	err := NewWorktreeError(ErrWorktreeExists, "worktree already exists", path)
//	    .WithSuggestion("Use 'twiggit switch' to navigate to existing worktree")
//
//	// With underlying cause:
//	err := NewWorktreeError(ErrGitCommand, "git command failed", path, gitErr)
//	    .WithSuggestion("Check git installation and repository state")
func NewWorktreeError(errType DomainErrorType, message string, path string, cause ...error) *DomainError {
	var err error
	if len(cause) > 0 {
		err = cause[0]
	}

	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       err,
		Suggestions: make([]string, 0),
		EntityType:  "worktree",
	}
}

// NewWorkspaceError creates a new DomainError with workspace context.
// Use this function for all workspace-related errors to ensure consistent
// error handling and user experience. Automatically sets entity type to "workspace".
// Accepts an optional cause error for wrapping underlying errors.
//
// Example:
//
//	err := NewWorkspaceError(ErrWorkspaceDiscoveryFailed, "discovery failed")
//	    .WithSuggestion("Check workspace configuration and permissions")
//
//	// With underlying cause:
//	err := NewWorkspaceError(ErrWorkspaceDiscoveryFailed, "discovery failed", underlyingErr)
//	    .WithSuggestion("Check workspace directory permissions")
func NewWorkspaceError(errType DomainErrorType, message string, cause ...error) *DomainError {
	var err error
	if len(cause) > 0 {
		err = cause[0]
	}

	return &DomainError{
		Type:        errType,
		Message:     message,
		Cause:       err,
		Suggestions: make([]string, 0),
		EntityType:  "workspace",
	}
}
