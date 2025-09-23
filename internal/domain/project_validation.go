package domain

// ProjectValidationResult is a legacy type alias for backward compatibility during migration
// Use unified ValidationResult from errors.go
type ProjectValidationResult = ValidationResult

// ValidateProjectName validates a project name according to business rules
func ValidateProjectName(projectName string) *ProjectValidationResult {
	return ValidateNotEmpty(projectName, "project name", ErrInvalidProjectName, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewProjectError(errorType, message, context).WithSuggestion("Provide a valid project name")
	})
}

// ValidateGitRepoPath validates a git repository path according to business rules
func ValidateGitRepoPath(gitRepoPath string) *ProjectValidationResult {
	return ValidateNotEmpty(gitRepoPath, "git repository path", ErrInvalidGitRepoPath, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewProjectError(errorType, message, context).WithSuggestion("Provide a valid git repository path")
	})
}

// ValidateProjectCreation performs comprehensive validation for project creation
func ValidateProjectCreation(projectName, gitRepoPath string) *ProjectValidationResult {
	return MergeValidationResults(
		ValidateProjectName(projectName),
		ValidateGitRepoPath(gitRepoPath),
	)
}

// ValidateProjectHealth validates the health status of a project (domain-only validation)
func ValidateProjectHealth(project *Project) *ProjectValidationResult {
	result := NewProjectValidationResult()

	// Basic validation - check if git repo path is not empty
	if project.GitRepo == "" {
		result.AddError(NewProjectError(
			ErrInvalidGitRepoPath,
			"git repository path cannot be empty",
			"",
		).WithSuggestion("Provide a valid git repository path"))
	}

	// Note: Infrastructure-specific validation (like path validation) is now handled
	// by the service layer. Domain validation only checks basic business rules.

	return result
}
