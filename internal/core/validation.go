package core

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

// validateBranchErr constructs a ValidationError for branch validation.
// Centralises the Op / Field assignment so the per-rule helpers stay
// terse and the literal field name only appears once.
func validateBranchErr(value, message, suggestion string) *ValidationError {
	return &ValidationError{
		Op:          "Validation",
		Field:       "BranchName",
		Value:       value,
		Message:     message,
		Suggestions: []string{suggestion},
	}
}

func validateProjectErr(value, message, suggestion string) *ValidationError {
	return &ValidationError{
		Op:          "Validation",
		Field:       "ProjectName",
		Value:       value,
		Message:     message,
		Suggestions: []string{suggestion},
	}
}

func validateShellErr(value, message, suggestion string) *ValidationError {
	return &ValidationError{
		Op:          "Validation",
		Field:       "ShellType",
		Value:       value,
		Message:     message,
		Suggestions: []string{suggestion},
	}
}

// Pure validation functions for branch names

// ValidateBranchNameNotEmpty checks if branch name is not empty or whitespace only
func ValidateBranchNameNotEmpty(branchName string) Result[bool] {
	if strings.TrimSpace(branchName) == "" {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name is required", "Provide a valid branch name"))
	}
	return NewResult(true)
}

// ValidateBranchNameReserved checks if branch name is a git reserved name
func ValidateBranchNameReserved(branchName string) Result[bool] {
	reservedNames := map[string]bool{
		"head":             true,
		"main":             true,
		"master":           true,
		"orig_head":        true,
		"fetch_head":       true,
		"merge_head":       true,
		"merge_state":      true,
		"cherry_pick_head": true,
		"revert_head":      true,
	}

	if reservedNames[strings.ToLower(branchName)] {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name format is invalid", "This is a git reserved branch name"))
	}
	return NewResult(true)
}

// ValidateBranchNameLeadingChars checks if branch name starts with invalid characters
func ValidateBranchNameLeadingChars(branchName string) Result[bool] {
	if len(branchName) > 0 && (branchName[0] == '.' || branchName[0] == '-') {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name format is invalid", "Branch names cannot start with . or -"))
	}
	return NewResult(true)
}

// ValidateBranchNameTrailingChars checks if branch name ends with invalid characters
func ValidateBranchNameTrailingChars(branchName string) Result[bool] {
	trimmed := strings.TrimRight(branchName, ".-_")
	if len(trimmed) != len(branchName) {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name format is invalid", "Branch names cannot end with ., -, or _"))
	}
	return NewResult(true)
}

// ValidateBranchNameFormat checks if branch name contains only valid characters
func ValidateBranchNameFormat(branchName string) Result[bool] {
	// Git branch names should follow: no spaces, no @, no #, etc.
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(branchName) {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name format is invalid", "Use only alphanumeric characters, dots, hyphens, and underscores"))
	}
	return NewResult(true)
}

// ValidateBranchNameLength checks if branch name is within reasonable length
func ValidateBranchNameLength(branchName string) Result[bool] {
	if len(branchName) > 255 {
		return NewErrResult[bool](validateBranchErr(branchName, "branch name is too long", "Branch names should be 255 characters or less"))
	}
	return NewResult(true)
}

// ValidateBranchName composes all branch name validations
func ValidateBranchName(branchName string) Result[bool] {
	pipeline := NewValidationPipeline(
		ValidateBranchNameNotEmpty,
		ValidateBranchNameReserved,
		ValidateBranchNameLeadingChars,
		ValidateBranchNameFormat,
		ValidateBranchNameTrailingChars,
		ValidateBranchNameLength,
	)
	return pipeline.Validate(branchName)
}

// Pure validation functions for project names

// ValidateProjectNameNotEmpty checks if project name is not empty or whitespace only
func ValidateProjectNameNotEmpty(projectName string) Result[bool] {
	if strings.TrimSpace(projectName) == "" {
		return NewErrResult[bool](validateProjectErr(projectName, "project name is required", "Provide a valid project name"))
	}
	return NewResult(true)
}

// ValidateProjectNameFormat checks if project name contains only valid characters
func ValidateProjectNameFormat(projectName string) Result[bool] {
	// Project names should be simpler than branch names
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(projectName) {
		return NewErrResult[bool](validateProjectErr(projectName, "project name format is invalid", "Use only alphanumeric characters, hyphens, and underscores"))
	}
	return NewResult(true)
}

// ValidateProjectNamePathTraversal checks if project name contains path traversal sequences
func ValidateProjectNamePathTraversal(projectName string) Result[bool] {
	if strings.Contains(projectName, "..") {
		return NewErrResult[bool](validateProjectErr(projectName, "project name contains invalid path traversal sequence", "Remove '..' from project name"))
	}
	return NewResult(true)
}

// ValidateProjectName composes all project name validations
func ValidateProjectName(projectName string) Result[bool] {
	pipeline := NewValidationPipeline(
		ValidateProjectNameNotEmpty,
		ValidateProjectNameFormat,
		ValidateProjectNamePathTraversal,
	)
	return pipeline.Validate(projectName)
}

// Pure validation functions for shell types

// ValidateShellTypeNotEmpty checks if shell type is not empty or whitespace only
func ValidateShellTypeNotEmpty(shellType string) Result[bool] {
	if strings.TrimSpace(shellType) == "" {
		return NewErrResult[bool](validateShellErr(shellType, "shell type is required", "Provide a valid shell type (bash, zsh, fish)"))
	}
	return NewResult(true)
}

// ValidateShellTypeFormat checks if shell type has no leading/trailing whitespace
func ValidateShellTypeFormat(shellType string) Result[bool] {
	if strings.TrimSpace(shellType) != shellType {
		return NewErrResult[bool](validateShellErr(shellType, "shell type format is invalid", "Shell type should not contain leading or trailing whitespace"))
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
		return NewErrResult[bool](validateShellErr(shellType, "unsupported shell type", "Supported shells: bash, zsh, fish"))
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
