package domain

import (
	"errors"
	"fmt"
)

// WorkspaceErrorType represents the type of workspace error
type WorkspaceErrorType int

const (
	// WorkspaceErrorInvalidPath indicates an invalid workspace path
	WorkspaceErrorInvalidPath WorkspaceErrorType = iota + 1

	// WorkspaceErrorProjectNotFound indicates a project was not found in the workspace
	WorkspaceErrorProjectNotFound

	// WorkspaceErrorProjectAlreadyExists indicates a project with the same name already exists
	WorkspaceErrorProjectAlreadyExists

	// WorkspaceErrorWorktreeNotFound indicates a worktree was not found in the workspace
	WorkspaceErrorWorktreeNotFound

	// WorkspaceErrorInvalidConfiguration indicates invalid workspace configuration
	WorkspaceErrorInvalidConfiguration

	// WorkspaceErrorDiscoveryFailed indicates project discovery failed
	WorkspaceErrorDiscoveryFailed

	// WorkspaceErrorValidationFailed indicates workspace validation failed
	WorkspaceErrorValidationFailed
)

// String returns the string representation of the error type
func (et WorkspaceErrorType) String() string {
	switch et {
	case WorkspaceErrorInvalidPath:
		return "WorkspaceErrorInvalidPath"
	case WorkspaceErrorProjectNotFound:
		return "WorkspaceErrorProjectNotFound"
	case WorkspaceErrorProjectAlreadyExists:
		return "WorkspaceErrorProjectAlreadyExists"
	case WorkspaceErrorWorktreeNotFound:
		return "WorkspaceErrorWorktreeNotFound"
	case WorkspaceErrorInvalidConfiguration:
		return "WorkspaceErrorInvalidConfiguration"
	case WorkspaceErrorDiscoveryFailed:
		return "WorkspaceErrorDiscoveryFailed"
	case WorkspaceErrorValidationFailed:
		return "WorkspaceErrorValidationFailed"
	default:
		return "UnknownWorkspaceError"
	}
}

// WorkspaceError represents a workspace-specific error
type WorkspaceError struct {
	Type       WorkspaceErrorType
	Message    string
	Underlying error
}

// NewWorkspaceError creates a new WorkspaceError
func NewWorkspaceError(errType WorkspaceErrorType, message string) *WorkspaceError {
	return &WorkspaceError{
		Type:    errType,
		Message: message,
	}
}

// Error returns the error message
func (e *WorkspaceError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Underlying)
	}
	return e.Message
}

// IsWorkspaceErrorType checks if the error is of the specified WorkspaceErrorType
func IsWorkspaceErrorType(err error, errType WorkspaceErrorType) bool {
	var workspaceErr *WorkspaceError
	if errors.As(err, &workspaceErr) {
		return workspaceErr.Type == errType
	}
	return false
}

// Unwrap returns the underlying error
func (e *WorkspaceError) Unwrap() error {
	return e.Underlying
}
