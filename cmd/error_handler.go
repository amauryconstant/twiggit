package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"twiggit/internal/domain"
)

// ExitCode defines the exit codes used by the application
type ExitCode int

const (
	// ExitCodeSuccess indicates successful execution
	ExitCodeSuccess ExitCode = 0
	// ExitCodeError indicates a general error occurred
	ExitCodeError ExitCode = 1
	// ExitCodeUsage indicates incorrect command-line usage
	ExitCodeUsage ExitCode = 2
)

// ErrorCategory defines categories of errors for consistent handling
type ErrorCategory int

const (
	// ErrorCategoryCobra represents Cobra argument/flag validation errors
	ErrorCategoryCobra ErrorCategory = iota
	// ErrorCategoryValidation represents input validation errors
	ErrorCategoryValidation
	// ErrorCategoryService represents service operation errors
	ErrorCategoryService
	// ErrorCategoryGit represents git operation errors
	ErrorCategoryGit
	// ErrorCategoryConfig represents configuration errors
	ErrorCategoryConfig
	// ErrorCategoryGeneric represents all other errors
	ErrorCategoryGeneric
)

// HandleCLIError is a pure function that maps errors to CLI output and returns exit code
func HandleCLIError(err error) ExitCode {
	// Check if this is a Cobra argument validation error
	if IsCobraArgumentError(err) {
		// Print Cobra's argument validation error since we silenced it in the command
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeUsage
	}

	// Format and print the error
	formatter := NewErrorFormatter()
	formattedError := formatter.Format(err)
	fmt.Fprint(os.Stderr, formattedError)

	// Return appropriate exit code based on error category
	return GetExitCodeForError(err)
}

// GetExitCodeForError maps errors to appropriate exit codes
func GetExitCodeForError(err error) ExitCode {
	category := CategorizeError(err)

	switch category {
	case ErrorCategoryCobra:
		return ExitCodeUsage
	case ErrorCategoryValidation, ErrorCategoryService, ErrorCategoryGit, ErrorCategoryConfig:
		return ExitCodeError
	default:
		return ExitCodeError
	}
}

// CategorizeError determines the category of an error for consistent handling
func CategorizeError(err error) ErrorCategory {
	// Check for Cobra argument errors first
	if IsCobraArgumentError(err) {
		return ErrorCategoryCobra
	}

	// Check for specific domain error types using errors.As for wrapped error support
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return ErrorCategoryValidation
	}

	var worktreeServiceErr *domain.WorktreeServiceError
	if errors.As(err, &worktreeServiceErr) {
		return ErrorCategoryService
	}

	var projectServiceErr *domain.ProjectServiceError
	if errors.As(err, &projectServiceErr) {
		return ErrorCategoryService
	}

	var serviceErr *domain.ServiceError
	if errors.As(err, &serviceErr) {
		return ErrorCategoryService
	}

	var gitRepoErr *domain.GitRepositoryError
	if errors.As(err, &gitRepoErr) {
		return ErrorCategoryGit
	}

	var gitWorktreeErr *domain.GitWorktreeError
	if errors.As(err, &gitWorktreeErr) {
		return ErrorCategoryGit
	}

	var gitCmdErr *domain.GitCommandError
	if errors.As(err, &gitCmdErr) {
		return ErrorCategoryGit
	}

	var configErr *domain.ConfigError
	if errors.As(err, &configErr) {
		return ErrorCategoryConfig
	}

	// Check for validation-related errors by message content
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "validation") {
		return ErrorCategoryValidation
	}

	return ErrorCategoryGeneric
}

// IsCobraArgumentError checks if the error is a Cobra argument validation error
func IsCobraArgumentError(err error) bool {
	errStr := err.Error()

	// Common Cobra error patterns for argument validation
	cobraPatterns := []string{
		"accepts",
		"requires",
		"received",
		"unknown shorthand flag",
		"unknown flag",
		"flag needs an argument",
		"required flag(s)",
	}

	for _, pattern := range cobraPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
