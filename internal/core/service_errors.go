package core

// Demoted service-error constructors. The legacy per-type taxonomy
// (ServiceError, WorktreeServiceError, ProjectServiceError,
// NavigationServiceError, ResolutionError, ConflictError) collapses to
// ValidationError; the original type names remain as constructor entry
// points to minimize call-site churn. Op / Entity / Field carry the
// distinguishing context that the previous per-type fields held.

// NewServiceError demoted: returns *ValidationError.
func NewServiceError(service, operation, message string, err error) *ValidationError {
	return &ValidationError{
		Op:      service,
		Entity:  operation,
		Message: message,
	}
}

// NewWorktreeServiceError demoted: returns *ValidationError.
func NewWorktreeServiceError(worktreePath, branchName, operation, message string, err error) *ValidationError {
	ve := &ValidationError{
		Op:      "worktree.service",
		Entity:  worktreePath,
		Message: message,
	}
	if branchName != "" {
		ve.Field = "branch:" + branchName
	}
	if err != nil {
		ve.Value = err.Error()
	}
	return ve
}

// NewProjectServiceError demoted: returns *ValidationError.
func NewProjectServiceError(projectName, projectPath, operation, message string, err error) *ValidationError {
	ve := &ValidationError{
		Op:      "project.service",
		Message: message,
	}
	if projectName != "" {
		ve.Entity = projectName
	} else {
		ve.Entity = projectPath
	}
	return ve
}

// NewNavigationServiceError demoted: returns *ValidationError.
func NewNavigationServiceError(target, context, operation, message string, err error) *ValidationError {
	return &ValidationError{
		Op:      "navigation.service",
		Entity:  target,
		Field:   "context:" + context,
		Message: message,
	}
}

// NewResolutionError demoted: returns *ValidationError.
func NewResolutionError(target, context, message string, suggestions []string, err error) *ValidationError {
	return &ValidationError{
		Op:          "resolution",
		Entity:      target,
		Field:       "context:" + context,
		Message:     message,
		Suggestions: suggestions,
	}
}

// NewConflictError demoted: returns *ValidationError.
func NewConflictError(resource, identifier, operation, message string, err error) *ValidationError {
	return &ValidationError{
		Op:      "conflict",
		Entity:  resource,
		Field:   identifier,
		Message: message,
	}
}
