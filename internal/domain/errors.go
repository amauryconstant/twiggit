package domain

import (
	"errors"
	"fmt"
)

// DomainErrorType represents different categories of domain errors
// This unified type replaces ProjectErrorType, ErrorType, and WorkspaceErrorType
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

// DomainError represents a comprehensive error with context for all domain operations
// This unified type replaces ProjectError, WorktreeError, and WorkspaceError
type DomainError struct {
	Type        DomainErrorType
	Message     string
	Path        string
	Cause       error
	Suggestions []string
	Code        string
	EntityType  string // "project", "worktree", "workspace" for context
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithSuggestion adds a suggestion to help resolve the error
func (e *DomainError) WithSuggestion(suggestion string) *DomainError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithCode adds an error code for programmatic handling
func (e *DomainError) WithCode(code string) *DomainError {
	e.Code = code
	return e
}

// WithEntityType adds entity type context for better error categorization
func (e *DomainError) WithEntityType(entityType string) *DomainError {
	e.EntityType = entityType
	return e
}

// IsDomainErrorType checks if an error is of a specific DomainError type
func IsDomainErrorType(err error, errType DomainErrorType) bool {
	domainErr := &DomainError{}
	if errors.As(err, &domainErr) {
		return domainErr.Type == errType
	}
	return false
}

// Domain-specific constructors that preserve ubiquitous language

// NewProjectError creates a new DomainError with project context
func NewProjectError(errType DomainErrorType, message string, path string) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Suggestions: make([]string, 0),
		EntityType:  "project",
	}
}

// NewWorktreeError creates a new DomainError with worktree context
func NewWorktreeError(errType DomainErrorType, message string, path string) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Suggestions: make([]string, 0),
		EntityType:  "worktree",
	}
}

// NewWorkspaceError creates a new DomainError with workspace context
func NewWorkspaceError(errType DomainErrorType, message string) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Suggestions: make([]string, 0),
		EntityType:  "workspace",
	}
}

// Wrapping functions that preserve entity context

// WrapProjectError wraps an existing error with project context
func WrapProjectError(errType DomainErrorType, message string, path string, cause error) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       cause,
		Suggestions: make([]string, 0),
		EntityType:  "project",
	}
}

// WrapWorktreeError wraps an existing error with worktree context
func WrapWorktreeError(errType DomainErrorType, message string, path string, cause error) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       cause,
		Suggestions: make([]string, 0),
		EntityType:  "worktree",
	}
}

// WrapWorkspaceError wraps an existing error with workspace context
func WrapWorkspaceError(errType DomainErrorType, message string, cause error) *DomainError {
	return &DomainError{
		Type:        errType,
		Message:     message,
		Cause:       cause,
		Suggestions: make([]string, 0),
		EntityType:  "workspace",
	}
}

// Legacy compatibility functions - these will be deprecated later

// IsProjectErrorType checks if an error is of a specific project error type
// This function maintains backward compatibility during migration
func IsProjectErrorType(err error, errType DomainErrorType) bool {
	return IsDomainErrorType(err, errType)
}

// IsWorktreeErrorType checks if an error is of a specific worktree error type
// This function maintains backward compatibility during migration
func IsWorktreeErrorType(err error, errType DomainErrorType) bool {
	return IsDomainErrorType(err, errType)
}

// IsWorkspaceErrorType checks if an error is of a specific workspace error type
// This function maintains backward compatibility during migration
func IsWorkspaceErrorType(err error, errType DomainErrorType) bool {
	return IsDomainErrorType(err, errType)
}

// WrapError is a legacy function for worktree error wrapping
// This function maintains backward compatibility during migration
func WrapError(errType DomainErrorType, message string, path string, cause error) *DomainError {
	return WrapWorktreeError(errType, message, path, cause)
}
