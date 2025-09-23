package domain

import (
	"errors"
	"fmt"
)

// ProjectErrorType represents different categories of project errors
type ProjectErrorType int

const (
	// ErrUnknownProjectError represents an unknown project error type
	ErrUnknownProjectError ProjectErrorType = iota
	// ErrInvalidProjectName indicates an invalid project name was provided
	ErrInvalidProjectName
	// ErrInvalidGitRepoPath indicates an invalid git repository path was provided
	ErrInvalidGitRepoPath
	// ErrProjectNotFound indicates a project was not found
	ErrProjectNotFound
	// ErrProjectAlreadyExists indicates a project already exists
	ErrProjectAlreadyExists
	// ErrProjectValidation indicates a project validation error
	ErrProjectValidation
)

// String returns a human-readable representation of the project error type
func (et ProjectErrorType) String() string {
	switch et {
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
	default:
		return "unknown project error"
	}
}

// ProjectError represents a comprehensive error with context for project operations
type ProjectError struct {
	Type        ProjectErrorType
	Message     string
	Path        string
	Cause       error
	Suggestions []string
	Code        string
}

// Error implements the error interface
func (e *ProjectError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Message, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ProjectError) Unwrap() error {
	return e.Cause
}

// WithSuggestion adds a suggestion to help resolve the error
func (e *ProjectError) WithSuggestion(suggestion string) *ProjectError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithCode adds an error code for programmatic handling
func (e *ProjectError) WithCode(code string) *ProjectError {
	e.Code = code
	return e
}

// NewProjectError creates a new ProjectError with the given type and message
func NewProjectError(errType ProjectErrorType, message string, path string) *ProjectError {
	return &ProjectError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Suggestions: make([]string, 0),
	}
}

// WrapProjectError wraps an existing error with project context
func WrapProjectError(errType ProjectErrorType, message string, path string, cause error) *ProjectError {
	return &ProjectError{
		Type:        errType,
		Message:     message,
		Path:        path,
		Cause:       cause,
		Suggestions: make([]string, 0),
	}
}

// IsProjectErrorType checks if an error is of a specific ProjectError type
func IsProjectErrorType(err error, errType ProjectErrorType) bool {
	projErr := &ProjectError{}
	if errors.As(err, &projErr) {
		return projErr.Type == errType
	}
	return false
}
