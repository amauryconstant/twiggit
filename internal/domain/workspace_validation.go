package domain

// WorkspaceValidationResult is a legacy type alias for backward compatibility during migration
// Use unified ValidationResult from errors.go
type WorkspaceValidationResult = ValidationResult

// ValidateWorkspacePath validates a workspace path
func ValidateWorkspacePath(path string) *ValidationResult {
	return ValidateNotEmpty(path, "workspace path", ErrWorkspaceInvalidPath, func(errorType DomainErrorType, message, _ string) *DomainError {
		return NewWorkspaceError(errorType, message)
	})
}

// ValidateWorkspaceProjectName validates a project name for addition to a workspace
func ValidateWorkspaceProjectName(projectName string, workspace *Workspace) *ValidationResult {
	// Validate project name is not empty
	result := ValidateNotEmpty(projectName, "project name", ErrWorkspaceInvalidConfiguration, func(errorType DomainErrorType, message, _ string) *DomainError {
		return NewWorkspaceError(errorType, message)
	})

	// If project name is empty, return early
	if result.HasErrors() {
		return result
	}

	// Check for duplicate project names
	if workspace != nil {
		for _, project := range workspace.Projects {
			if project.Name == projectName {
				result.AddError(NewWorkspaceError(
					ErrWorkspaceProjectAlreadyExists,
					"project '"+projectName+"' already exists in workspace",
				))
				break
			}
		}
	}

	return result
}

// ValidateWorkspaceProjectExists validates that a project exists in the workspace
func ValidateWorkspaceProjectExists(projectName string, workspace *Workspace) *ValidationResult {
	// Validate workspace is not nil - use direct check since ValidateNotNil seems to have issues
	if workspace == nil {
		result := NewWorkspaceValidationResult()
		result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "workspace cannot be nil"))
		return result
	}

	// Validate project name is not empty
	nameResult := ValidateNotEmpty(projectName, "project name", ErrWorkspaceInvalidConfiguration, func(errorType DomainErrorType, message, _ string) *DomainError {
		return NewWorkspaceError(errorType, message)
	})
	if nameResult.HasErrors() {
		return nameResult
	}

	// Check if project exists
	result := NewWorkspaceValidationResult()
	found := false
	for _, project := range workspace.Projects {
		if project.Name == projectName {
			found = true
			break
		}
	}

	if !found {
		result.AddError(NewWorkspaceError(
			ErrWorkspaceProjectNotFound,
			"project '"+projectName+"' not found in workspace",
		))
	}

	return result
}

// ValidateWorkspaceCreation validates workspace creation parameters
func ValidateWorkspaceCreation(path string) *ValidationResult {
	return ValidateWorkspacePath(path)
}

// ValidateWorkspaceHealth validates the health of a workspace (domain-only validation)
func ValidateWorkspaceHealth(workspace *Workspace) *ValidationResult {
	result := NewWorkspaceValidationResult()

	if workspace == nil {
		result.AddError(NewWorkspaceError(
			ErrWorkspaceInvalidConfiguration,
			"workspace cannot be nil",
		))
		return result
	}

	// Basic validation - check if workspace path is not empty
	if workspace.Path == "" {
		result.AddError(NewWorkspaceError(
			ErrWorkspaceInvalidPath,
			"workspace path cannot be empty",
		))
	}

	// Note: Infrastructure-specific validation (like path validation) is now handled
	// by the service layer. Domain validation only checks basic business rules.

	return result
}
