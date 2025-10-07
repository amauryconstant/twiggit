package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"twiggit/internal/domain"
)

// ExitCode defines the exit codes used by the application
type ExitCode int

const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
	ExitCodeUsage   ExitCode = 2
)

// ErrorCategory defines categories of errors for consistent handling
type ErrorCategory int

const (
	ErrorCategoryCobra      ErrorCategory = iota // Cobra argument/flag validation errors
	ErrorCategoryValidation                      // Input validation errors
	ErrorCategoryService                         // Service operation errors
	ErrorCategoryGit                             // Git operation errors
	ErrorCategoryConfig                          // Configuration errors
	ErrorCategoryGeneric                         // All other errors
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

	// Check for specific domain error types using reflection
	errType := reflect.TypeOf(err)

	switch errType {
	case reflect.TypeOf(&domain.ValidationError{}):
		return ErrorCategoryValidation
	case reflect.TypeOf(&domain.WorktreeServiceError{}):
		return ErrorCategoryService
	case reflect.TypeOf(&domain.ProjectServiceError{}):
		return ErrorCategoryService
	case reflect.TypeOf(&domain.ServiceError{}):
		return ErrorCategoryService
	case reflect.TypeOf(&domain.GitRepositoryError{}):
		return ErrorCategoryGit
	case reflect.TypeOf(&domain.GitWorktreeError{}):
		return ErrorCategoryGit
	case reflect.TypeOf(&domain.GitCommandError{}):
		return ErrorCategoryGit
	}

	// Check for configuration-related errors by message content
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "config") || strings.Contains(errStr, "toml") {
		return ErrorCategoryConfig
	}

	// Check for validation-related errors by message content
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
