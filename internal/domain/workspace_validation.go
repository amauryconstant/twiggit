package domain

import (
	"strings"
)

// WorkspaceValidationResult represents the result of workspace validation
type WorkspaceValidationResult struct {
	Errors []WorkspaceError
}

// IsValid returns true if the validation result contains no errors
func (vr *WorkspaceValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}

// GetErrorCount returns the number of validation errors
func (vr *WorkspaceValidationResult) GetErrorCount() int {
	return len(vr.Errors)
}

// GetErrorsByType returns all errors of the specified type
func (vr *WorkspaceValidationResult) GetErrorsByType(errorType WorkspaceErrorType) []WorkspaceError {
	result := make([]WorkspaceError, 0)
	for i := range vr.Errors {
		if IsWorkspaceErrorType(&vr.Errors[i], errorType) {
			result = append(result, vr.Errors[i])
		}
	}
	return result
}

// GetFirstError returns the first error in the validation result, or nil if there are no errors
func (vr *WorkspaceValidationResult) GetFirstError() *WorkspaceError {
	if len(vr.Errors) == 0 {
		return nil
	}
	return &vr.Errors[0]
}

// AddError adds a validation error to the result
func (vr *WorkspaceValidationResult) AddError(err WorkspaceError) {
	vr.Errors = append(vr.Errors, err)
}

// Merge merges another validation result into this one, returning a new result
func (vr *WorkspaceValidationResult) Merge(other WorkspaceValidationResult) WorkspaceValidationResult {
	merged := NewWorkspaceValidationResult()
	merged.Errors = append(vr.Errors, other.Errors...)
	return merged
}

// ToError returns the first validation error as an error interface, or nil if there are no errors
func (vr *WorkspaceValidationResult) ToError() error {
	if len(vr.Errors) == 0 {
		return nil
	}
	return &vr.Errors[0]
}

// NewWorkspaceValidationResult creates a new workspace validation result with optional initial errors
func NewWorkspaceValidationResult(errors ...WorkspaceError) WorkspaceValidationResult {
	if len(errors) == 0 {
		return WorkspaceValidationResult{
			Errors: make([]WorkspaceError, 0),
		}
	}
	return WorkspaceValidationResult{
		Errors: errors,
	}
}

// ValidateWorkspacePath validates a workspace path
func ValidateWorkspacePath(path string) WorkspaceValidationResult {
	result := NewWorkspaceValidationResult()

	if strings.TrimSpace(path) == "" {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidPath,
			Message: "workspace path cannot be empty",
		})
	}

	return result
}

// ValidateWorkspaceProjectName validates a project name for addition to a workspace
func ValidateWorkspaceProjectName(projectName string, workspace *Workspace) WorkspaceValidationResult {
	result := NewWorkspaceValidationResult()

	if strings.TrimSpace(projectName) == "" {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidConfiguration,
			Message: "project name cannot be empty",
		})
		return result
	}

	if workspace != nil {
		for _, project := range workspace.Projects {
			if project.Name == projectName {
				result.AddError(WorkspaceError{
					Type:    WorkspaceErrorProjectAlreadyExists,
					Message: "project '" + projectName + "' already exists in workspace",
				})
				break
			}
		}
	}

	return result
}

// ValidateWorkspaceProjectExists validates that a project exists in the workspace
func ValidateWorkspaceProjectExists(projectName string, workspace *Workspace) WorkspaceValidationResult {
	result := NewWorkspaceValidationResult()

	if workspace == nil {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidConfiguration,
			Message: "workspace cannot be nil",
		})
		return result
	}

	if strings.TrimSpace(projectName) == "" {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidConfiguration,
			Message: "project name cannot be empty",
		})
		return result
	}

	found := false
	for _, project := range workspace.Projects {
		if project.Name == projectName {
			found = true
			break
		}
	}

	if !found {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorProjectNotFound,
			Message: "project '" + projectName + "' not found in workspace",
		})
	}

	return result
}

// ValidateWorkspaceCreation validates workspace creation parameters
func ValidateWorkspaceCreation(path string) WorkspaceValidationResult {
	return ValidateWorkspacePath(path)
}

// ValidateWorkspaceHealth validates the health of a workspace
func ValidateWorkspaceHealth(workspace *Workspace, validator PathValidator) WorkspaceValidationResult {
	result := NewWorkspaceValidationResult()

	if workspace == nil {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidConfiguration,
			Message: "workspace cannot be nil",
		})
		return result
	}

	if validator == nil {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorInvalidConfiguration,
			Message: "path validator cannot be nil",
		})
		return result
	}

	if !validator.IsValidWorkspacePath(workspace.Path) {
		result.AddError(WorkspaceError{
			Type:    WorkspaceErrorValidationFailed,
			Message: "workspace path '" + workspace.Path + "' is not valid",
		})
	}

	return result
}
