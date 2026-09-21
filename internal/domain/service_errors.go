package domain

import (
	"fmt"
)

// ServiceError represents a general service operation error
type ServiceError struct {
	Service   string // Service name (e.g., "WorktreeService", "ProjectService")
	Operation string // Operation name (e.g., "CreateWorktree", "DiscoverProject")
	Message   string // Error message
	Err       error  // Underlying cause
}

func (e *ServiceError) Error() string {
	// Return user-friendly message without internal operation names
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new service error
func NewServiceError(service, operation, message string, err error) *ServiceError {
	return &ServiceError{
		Service:   service,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// ValidationError represents a validation error for service requests
type ValidationError struct {
	field       string   // Field name that failed validation
	value       string   // Field value that failed validation
	message     string   // Validation error message
	request     string   // Request type name
	suggestions []string // Helpful suggestions for fixing the error
	context     string   // Additional context information
}

func (e *ValidationError) Error() string {
	baseMsg := fmt.Sprintf("validation failed for %s.%s: %s (value: %s)", e.request, e.field, e.message, e.value)
	return baseMsg
}

// Unwrap returns nil: ValidationError is a terminal error type with no
// underlying cause to chain to.
func (e *ValidationError) Unwrap() error { return nil }

// NewValidationError creates a new validation error
func NewValidationError(request, field, value, message string) *ValidationError {
	return &ValidationError{
		field:       field,
		value:       value,
		message:     message,
		request:     request,
		suggestions: []string{},
		context:     "",
	}
}

// WithSuggestions returns a new ValidationError with suggestions (immutable)
func (e *ValidationError) WithSuggestions(suggestions []string) *ValidationError {
	newVE := *e // Copy
	newVE.suggestions = make([]string, len(suggestions))
	copy(newVE.suggestions, suggestions)
	return &newVE
}

// WithContext returns a new ValidationError with context (immutable)
func (e *ValidationError) WithContext(context string) *ValidationError {
	newVE := *e // Copy
	newVE.context = context
	return &newVE
}

// Pure getter methods

// Field returns the validation field that failed
func (e *ValidationError) Field() string { return e.field }

// Value returns the value that caused the validation failure
func (e *ValidationError) Value() string { return e.value }

// Message returns the validation error message
func (e *ValidationError) Message() string { return e.message }

// Request returns the request context for the validation error
func (e *ValidationError) Request() string { return e.request }

// Suggestions returns suggestions for fixing the validation error
func (e *ValidationError) Suggestions() []string {
	result := make([]string, len(e.suggestions))
	copy(result, e.suggestions)
	return result
}

// Context returns the context information for the validation error
func (e *ValidationError) Context() string { return e.context }

// WorktreeServiceError represents worktree service specific errors
type WorktreeServiceError struct {
	WorktreePath string
	BranchName   string
	Operation    string
	Message      string
	Err          error
}

func (e *WorktreeServiceError) Error() string {
	// Return user-friendly message without internal operation names
	if e.BranchName != "" {
		return fmt.Sprintf("%s for worktree '%s' (branch: %s)", e.Message, e.WorktreePath, e.BranchName)
	}
	return fmt.Sprintf("%s for worktree '%s'", e.Message, e.WorktreePath)
}

func (e *WorktreeServiceError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *WorktreeServiceError) Is(target error) bool {
	return target == ErrWorktreeNotFound
}

// NewWorktreeServiceError creates a new worktree service error
func NewWorktreeServiceError(worktreePath, branchName, operation, message string, err error) *WorktreeServiceError {
	return &WorktreeServiceError{
		WorktreePath: worktreePath,
		BranchName:   branchName,
		Operation:    operation,
		Message:      message,
		Err:          err,
	}
}

// ProjectServiceError represents project service specific errors
type ProjectServiceError struct {
	ProjectName string
	ProjectPath string
	Operation   string
	Message     string
	Err         error
}

func (e *ProjectServiceError) Error() string {
	// Return user-friendly message without internal operation names
	if e.ProjectName != "" {
		return fmt.Sprintf("%s for project '%s'", e.Message, e.ProjectName)
	}
	return fmt.Sprintf("%s for '%s'", e.Message, e.ProjectPath)
}

func (e *ProjectServiceError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *ProjectServiceError) Is(target error) bool {
	return target == ErrProjectNotFound
}

// NewProjectServiceError creates a new project service error
func NewProjectServiceError(projectName, projectPath, operation, message string, err error) *ProjectServiceError {
	return &ProjectServiceError{
		ProjectName: projectName,
		ProjectPath: projectPath,
		Operation:   operation,
		Message:     message,
		Err:         err,
	}
}

// NavigationServiceError represents navigation service specific errors
type NavigationServiceError struct {
	Target    string
	Context   string
	Operation string
	Message   string
	Err       error
}

func (e *NavigationServiceError) Error() string {
	// Return user-friendly message without internal operation names
	if e.Context != "" {
		return fmt.Sprintf("%s for target '%s' (context: %s)", e.Message, e.Target, e.Context)
	}
	return fmt.Sprintf("%s for target '%s'", e.Message, e.Target)
}

func (e *NavigationServiceError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *NavigationServiceError) Is(target error) bool {
	return target == ErrResolutionNotFound
}

// NewNavigationServiceError creates a new navigation service error
func NewNavigationServiceError(target, context, operation, message string, err error) *NavigationServiceError {
	return &NavigationServiceError{
		Target:    target,
		Context:   context,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// ResolutionError represents path resolution errors
type ResolutionError struct {
	Target      string
	Context     string
	Message     string
	Suggestions []string // Optional suggestions for resolution
	Err         error
}

func (e *ResolutionError) Error() string {
	baseMsg := fmt.Sprintf("resolution failed for target '%s' (context: %s): %s", e.Target, e.Context, e.Message)

	if len(e.Suggestions) > 0 {
		baseMsg += fmt.Sprintf("\nsuggestions: %v", e.Suggestions)
	}

	return baseMsg
}

func (e *ResolutionError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *ResolutionError) Is(target error) bool {
	return target == ErrResolutionNotFound
}

// NewResolutionError creates a new resolution error
func NewResolutionError(target, context, message string, suggestions []string, err error) *ResolutionError {
	return &ResolutionError{
		Target:      target,
		Context:     context,
		Message:     message,
		Suggestions: suggestions,
		Err:         err,
	}
}

// ConflictError represents operation conflict errors
type ConflictError struct {
	Resource   string // Resource type (e.g., "worktree", "branch")
	Identifier string // Resource identifier
	Operation  string // Operation that conflicted
	Message    string // Conflict description
	Err        error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict during %s operation on %s '%s': %s", e.Operation, e.Resource, e.Identifier, e.Message)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, identifier, operation, message string, err error) *ConflictError {
	return &ConflictError{
		Resource:   resource,
		Identifier: identifier,
		Operation:  operation,
		Message:    message,
		Err:        err,
	}
}
