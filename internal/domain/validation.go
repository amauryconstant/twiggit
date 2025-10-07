package domain

import (
	"regexp"
	"strings"
)

// ValidationFunc is a pure function that validates input and returns Result
type ValidationFunc[T any] func(T) Result[bool]

// ValidationPipeline composes multiple validation functions
type ValidationPipeline[T any] struct {
	validations []ValidationFunc[T]
}

// NewValidationPipeline creates a new validation pipeline with the given validations
func NewValidationPipeline[T any](validations ...ValidationFunc[T]) *ValidationPipeline[T] {
	return &ValidationPipeline[T]{
		validations: validations,
	}
}

// Validate runs all validations in the pipeline, returning early on first error
func (vp *ValidationPipeline[T]) Validate(input T) Result[bool] {
	for _, validation := range vp.validations {
		if result := validation(input); result.IsError() {
			return result
		}
	}
	return NewResult(true)
}

// Pure validation functions for branch names

// ValidateBranchNameNotEmpty checks if branch name is not empty or whitespace only
func ValidateBranchNameNotEmpty(branchName string) Result[bool] {
	if strings.TrimSpace(branchName) == "" {
		return NewErrorResult[bool](
			NewValidationError("Validation", "BranchName", branchName, "branch name is required").
				WithSuggestions([]string{"Provide a valid branch name"}),
		)
	}
	return NewResult(true)
}

// ValidateBranchNameFormat checks if branch name contains only valid characters
func ValidateBranchNameFormat(branchName string) Result[bool] {
	// Git branch names should follow: no spaces, no @, no #, etc.
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(branchName) {
		return NewErrorResult[bool](
			NewValidationError("Validation", "BranchName", branchName, "branch name format is invalid").
				WithSuggestions([]string{"Use only alphanumeric characters, dots, hyphens, and underscores"}),
		)
	}
	return NewResult(true)
}

// ValidateBranchNameLength checks if branch name is within reasonable length
func ValidateBranchNameLength(branchName string) Result[bool] {
	if len(branchName) > 255 {
		return NewErrorResult[bool](
			NewValidationError("Validation", "BranchName", branchName, "branch name is too long").
				WithSuggestions([]string{"Branch names should be 255 characters or less"}),
		)
	}
	return NewResult(true)
}

// ValidateBranchName composes all branch name validations
func ValidateBranchName(branchName string) Result[bool] {
	pipeline := NewValidationPipeline(
		ValidateBranchNameNotEmpty,
		ValidateBranchNameFormat,
		ValidateBranchNameLength,
	)
	return pipeline.Validate(branchName)
}

// Pure validation functions for project names

// ValidateProjectNameNotEmpty checks if project name is not empty or whitespace only
func ValidateProjectNameNotEmpty(projectName string) Result[bool] {
	if strings.TrimSpace(projectName) == "" {
		return NewErrorResult[bool](
			NewValidationError("Validation", "ProjectName", projectName, "project name is required").
				WithSuggestions([]string{"Provide a valid project name"}),
		)
	}
	return NewResult(true)
}

// ValidateProjectNameFormat checks if project name contains only valid characters
func ValidateProjectNameFormat(projectName string) Result[bool] {
	// Project names should be simpler than branch names
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(projectName) {
		return NewErrorResult[bool](
			NewValidationError("Validation", "ProjectName", projectName, "project name format is invalid").
				WithSuggestions([]string{"Use only alphanumeric characters, hyphens, and underscores"}),
		)
	}
	return NewResult(true)
}

// ValidateProjectName composes all project name validations
func ValidateProjectName(projectName string) Result[bool] {
	pipeline := NewValidationPipeline(
		ValidateProjectNameNotEmpty,
		ValidateProjectNameFormat,
	)
	return pipeline.Validate(projectName)
}

// Pure validation functions for shell types

// ValidateShellTypeNotEmpty checks if shell type is not empty or whitespace only
func ValidateShellTypeNotEmpty(shellType string) Result[bool] {
	if strings.TrimSpace(shellType) == "" {
		return NewErrorResult[bool](
			NewValidationError("Validation", "ShellType", shellType, "shell type is required").
				WithSuggestions([]string{"Provide a valid shell type (bash, zsh, fish)"}),
		)
	}
	return NewResult(true)
}

// ValidateShellTypeFormat checks if shell type has no leading/trailing whitespace
func ValidateShellTypeFormat(shellType string) Result[bool] {
	if strings.TrimSpace(shellType) != shellType {
		return NewErrorResult[bool](
			NewValidationError("Validation", "ShellType", shellType, "shell type format is invalid").
				WithSuggestions([]string{"Shell type should not contain leading or trailing whitespace"}),
		)
	}
	return NewResult(true)
}

// ValidateShellTypeSupported checks if shell type is supported
func ValidateShellTypeSupported(shellType string) Result[bool] {
	supportedShells := map[string]bool{
		"bash": true,
		"zsh":  true,
		"fish": true,
	}

	if !supportedShells[shellType] {
		return NewErrorResult[bool](
			NewValidationError("Validation", "ShellType", shellType, "unsupported shell type").
				WithSuggestions([]string{"Supported shells: bash, zsh, fish"}),
		)
	}
	return NewResult(true)
}

// ValidateShellType composes all shell type validations
func ValidateShellType(shellType string) Result[bool] {
	pipeline := NewValidationPipeline(
		ValidateShellTypeNotEmpty,
		ValidateShellTypeFormat,
		ValidateShellTypeSupported,
	)
	return pipeline.Validate(shellType)
}
