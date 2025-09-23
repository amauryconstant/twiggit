package domain

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// validBranchNameRegex matches valid git branch names
	// Based on git-check-ref-format rules
	validBranchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
)

// WorktreeValidationResult is a legacy type alias for backward compatibility during migration
// Use unified ValidationResult from errors.go
type WorktreeValidationResult = ValidationResult

// ValidateBranchName validates a Git branch name according to git naming rules
func ValidateBranchName(branchName string) *ValidationResult {
	// Validate not empty
	result := ValidateNotEmpty(branchName, "branch name", ErrInvalidBranchName, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewWorktreeError(errorType, message, context).WithSuggestion("Provide a valid branch name")
	})
	if result.HasErrors() {
		return result
	}

	// Validate length
	lengthResult := ValidateStringLength(branchName, "branch name", MinBranchNameLength, MaxBranchNameLength, ErrInvalidBranchName, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewWorktreeError(errorType, message, context).WithSuggestion("Use a branch name with appropriate length")
	})
	if lengthResult.HasErrors() {
		return lengthResult
	}

	// Validate characters
	charResult := ValidateCharacters(branchName, "branch name", InvalidBranchChars, ErrInvalidBranchName, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewWorktreeError(errorType, message, context).WithSuggestion("Remove invalid characters from branch name")
	})
	if charResult.HasErrors() {
		return charResult
	}

	// Check for reserved names
	if strings.HasPrefix(branchName, "-") {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name cannot start with a hyphen",
			"",
		).WithSuggestion("Choose a branch name that doesn't start with '-'"))
	}

	if strings.HasSuffix(branchName, ".lock") {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name cannot end with '.lock'",
			"",
		).WithSuggestion("Choose a branch name that doesn't end with '.lock'"))
	}

	// Check UTF-8 validity
	if !utf8.ValidString(branchName) {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name contains invalid UTF-8 characters",
			"",
		).WithSuggestion("Use only valid UTF-8 characters in branch name"))
	}

	// Check against regex pattern (if no other errors)
	if result.Valid && !validBranchNameRegex.MatchString(branchName) {
		result.AddError(NewWorktreeError(
			ErrInvalidBranchName,
			"branch name format is invalid",
			"",
		).WithSuggestion("Use alphanumeric characters, dots, slashes, and hyphens only"))
	}

	return result
}

// ValidatePath validates a filesystem path for worktree operations
func ValidatePath(path string) *ValidationResult {
	// Validate not empty
	result := ValidateNotEmpty(path, "path", ErrInvalidPath, func(errorType DomainErrorType, message, context string) *DomainError {
		return NewWorktreeError(errorType, message, context).WithSuggestion("Provide a valid filesystem path")
	})
	if result.HasErrors() {
		return result
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path must be absolute",
			path,
		).WithSuggestion("Use an absolute path starting with '/'"))
	}

	// Check path length (filesystem dependent, but 4096 is a safe limit)
	if len(path) > MaxPathLength {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			fmt.Sprintf("path too long (maximum %d characters)", MaxPathLength),
			path,
		).WithSuggestion("Use a shorter path"))
	}

	// Check for invalid characters in path
	if strings.Contains(path, "\x00") {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path contains null character",
			path,
		).WithSuggestion("Remove null characters from path"))
	}

	// Check UTF-8 validity
	if !utf8.ValidString(path) {
		result.AddError(NewWorktreeError(
			ErrInvalidPath,
			"path contains invalid UTF-8 characters",
			path,
		).WithSuggestion("Use only valid UTF-8 characters in path"))
	}

	return result
}

// ValidatePathWritable checks if a path is writable (pure business logic only)
// Note: This function only validates the path format, not actual filesystem access
// For full validation including filesystem checks, use services.ValidationService.ValidatePathWritable
func ValidatePathWritable(path string) *ValidationResult {
	// First validate the path format
	pathResult := ValidatePath(path)
	if !pathResult.Valid {
		return pathResult
	}

	// Pure business logic validation only
	// Actual filesystem checks are handled by the service layer

	return NewValidationResult()
}

// ValidateWorktreeCreation performs comprehensive validation for worktree creation
// Note: This function only validates the format, not actual filesystem access
// For full validation including filesystem checks, use services.ValidationService.ValidateWorktreeCreation
func ValidateWorktreeCreation(branchName, targetPath string) *ValidationResult {
	return MergeValidationResults(
		ValidateBranchName(branchName),
		ValidatePathWritable(targetPath),
	)
}
