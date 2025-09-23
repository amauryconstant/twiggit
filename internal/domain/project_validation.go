package domain

import (
	"strings"
)

// ProjectValidationResult represents the result of a project validation operation
type ProjectValidationResult struct {
	Valid    bool
	Errors   []*ProjectError
	Warnings []string
}

// AddError adds an error to the project validation result
func (vr *ProjectValidationResult) AddError(err *ProjectError) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, err)
}

// AddWarning adds a warning to the project validation result
func (vr *ProjectValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// HasErrors returns true if there are validation errors
func (vr *ProjectValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// FirstError returns the first validation error, or nil if none
func (vr *ProjectValidationResult) FirstError() error {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return nil
}

// NewProjectValidationResult creates a new ProjectValidationResult
func NewProjectValidationResult() *ProjectValidationResult {
	return &ProjectValidationResult{
		Valid:    true,
		Errors:   make([]*ProjectError, 0),
		Warnings: make([]string, 0),
	}
}

// ValidateProjectName validates a project name according to business rules
func ValidateProjectName(projectName string) *ProjectValidationResult {
	result := NewProjectValidationResult()

	trimmedName := strings.TrimSpace(projectName)
	if trimmedName == "" {
		result.AddError(NewProjectError(
			ErrInvalidProjectName,
			"project name cannot be empty",
			"",
		).WithSuggestion("Provide a valid project name"))
		return result
	}

	// Additional validation rules can be added here as needed
	// For now, we accept any non-empty trimmed string as valid

	return result
}

// ValidateGitRepoPath validates a git repository path according to business rules
func ValidateGitRepoPath(gitRepoPath string) *ProjectValidationResult {
	result := NewProjectValidationResult()

	trimmedPath := strings.TrimSpace(gitRepoPath)
	if trimmedPath == "" {
		result.AddError(NewProjectError(
			ErrInvalidGitRepoPath,
			"git repository path cannot be empty",
			"",
		).WithSuggestion("Provide a valid git repository path"))
		return result
	}

	// Additional validation rules can be added here as needed
	// For now, we accept any non-empty trimmed string as valid
	// Filesystem-specific validation is handled by the infrastructure layer

	return result
}

// ValidateProjectCreation performs comprehensive validation for project creation
func ValidateProjectCreation(projectName, gitRepoPath string) *ProjectValidationResult {
	result := NewProjectValidationResult()

	// Validate project name
	nameResult := ValidateProjectName(projectName)
	result.Errors = append(result.Errors, nameResult.Errors...)
	result.Warnings = append(result.Warnings, nameResult.Warnings...)

	// Validate git repository path
	pathResult := ValidateGitRepoPath(gitRepoPath)
	result.Errors = append(result.Errors, pathResult.Errors...)
	result.Warnings = append(result.Warnings, pathResult.Warnings...)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// ValidateProjectHealth validates the health status of a project using path validator
func ValidateProjectHealth(project *Project, pathValidator PathValidator) *ProjectValidationResult {
	result := NewProjectValidationResult()

	// Basic validation - check if git repo path is not empty
	if project.GitRepo == "" {
		result.AddError(NewProjectError(
			ErrInvalidGitRepoPath,
			"git repository path cannot be empty",
			"",
		).WithSuggestion("Provide a valid git repository path"))
	}

	// Additional validation - check if git repo path looks valid
	if project.GitRepo != "" && !pathValidator.IsValidGitRepoPath(project.GitRepo) {
		result.AddError(NewProjectError(
			ErrInvalidGitRepoPath,
			"git repository not validated",
			project.GitRepo,
		).WithSuggestion("Check that the path points to a valid git repository"))
	}

	return result
}
