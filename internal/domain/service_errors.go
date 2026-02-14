package domain

import (
	"fmt"
	"strings"
)

// ServiceError represents a general service operation error
type ServiceError struct {
	Service   string // Service name (e.g., "WorktreeService", "ProjectService")
	Operation string // Operation name (e.g., "CreateWorktree", "DiscoverProject")
	Message   string // Error message
	Cause     error  // Underlying cause
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s.%s failed: %s", e.Service, e.Operation, e.Message)
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(service, operation, message string, cause error) *ServiceError {
	return &ServiceError{
		Service:   service,
		Operation: operation,
		Message:   message,
		Cause:     cause,
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
	if len(e.suggestions) > 0 {
		var sb strings.Builder
		sb.WriteString(baseMsg)
		for _, suggestion := range e.suggestions {
			sb.WriteString("\n💡 ")
			sb.WriteString(suggestion)
		}
		return sb.String()
	}
	return baseMsg
}

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
	Cause        error
}

func (e *WorktreeServiceError) Error() string {
	if e.BranchName != "" {
		return fmt.Sprintf("worktree service operation '%s' failed for %s (branch: %s): %s", e.Operation, e.WorktreePath, e.BranchName, e.Message)
	}
	return fmt.Sprintf("worktree service operation '%s' failed for %s: %s", e.Operation, e.WorktreePath, e.Message)
}

func (e *WorktreeServiceError) Unwrap() error {
	return e.Cause
}

// IsNotFound returns true if the error indicates the worktree was not found.
func (e *WorktreeServiceError) IsNotFound() bool {
	lowerMsg := strings.ToLower(e.Message)
	return strings.Contains(lowerMsg, "not found") ||
		strings.Contains(lowerMsg, "does not exist")
}

// NewWorktreeServiceError creates a new worktree service error
func NewWorktreeServiceError(worktreePath, branchName, operation, message string, cause error) *WorktreeServiceError {
	return &WorktreeServiceError{
		WorktreePath: worktreePath,
		BranchName:   branchName,
		Operation:    operation,
		Message:      message,
		Cause:        cause,
	}
}

// ProjectServiceError represents project service specific errors
type ProjectServiceError struct {
	ProjectName string
	ProjectPath string
	Operation   string
	Message     string
	Cause       error
}

func (e *ProjectServiceError) Error() string {
	if e.ProjectName != "" {
		return fmt.Sprintf("project service operation '%s' failed for project '%s': %s", e.Operation, e.ProjectName, e.Message)
	}
	return fmt.Sprintf("project service operation '%s' failed for %s: %s", e.Operation, e.ProjectPath, e.Message)
}

func (e *ProjectServiceError) Unwrap() error {
	return e.Cause
}

// NewProjectServiceError creates a new project service error
func NewProjectServiceError(projectName, projectPath, operation, message string, cause error) *ProjectServiceError {
	return &ProjectServiceError{
		ProjectName: projectName,
		ProjectPath: projectPath,
		Operation:   operation,
		Message:     message,
		Cause:       cause,
	}
}

// NavigationServiceError represents navigation service specific errors
type NavigationServiceError struct {
	Target    string
	Context   string
	Operation string
	Message   string
	Cause     error
}

func (e *NavigationServiceError) Error() string {
	return fmt.Sprintf("navigation service operation '%s' failed for target '%s' (context: %s): %s", e.Operation, e.Target, e.Context, e.Message)
}

func (e *NavigationServiceError) Unwrap() error {
	return e.Cause
}

// NewNavigationServiceError creates a new navigation service error
func NewNavigationServiceError(target, context, operation, message string, cause error) *NavigationServiceError {
	return &NavigationServiceError{
		Target:    target,
		Context:   context,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// ResolutionError represents path resolution errors
type ResolutionError struct {
	Target      string
	Context     string
	Message     string
	Suggestions []string // Optional suggestions for resolution
	Cause       error
}

func (e *ResolutionError) Error() string {
	baseMsg := fmt.Sprintf("resolution failed for target '%s' (context: %s): %s", e.Target, e.Context, e.Message)

	if len(e.Suggestions) > 0 {
		baseMsg += fmt.Sprintf("\nsuggestions: %v", e.Suggestions)
	}

	return baseMsg
}

func (e *ResolutionError) Unwrap() error {
	return e.Cause
}

// NewResolutionError creates a new resolution error
func NewResolutionError(target, context, message string, suggestions []string, cause error) *ResolutionError {
	return &ResolutionError{
		Target:      target,
		Context:     context,
		Message:     message,
		Suggestions: suggestions,
		Cause:       cause,
	}
}

// ConflictError represents operation conflict errors
type ConflictError struct {
	Resource   string // Resource type (e.g., "worktree", "branch")
	Identifier string // Resource identifier
	Operation  string // Operation that conflicted
	Message    string // Conflict description
	Cause      error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict during %s operation on %s '%s': %s", e.Operation, e.Resource, e.Identifier, e.Message)
}

func (e *ConflictError) Unwrap() error {
	return e.Cause
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, identifier, operation, message string, cause error) *ConflictError {
	return &ConflictError{
		Resource:   resource,
		Identifier: identifier,
		Operation:  operation,
		Message:    message,
		Cause:      cause,
	}
}
