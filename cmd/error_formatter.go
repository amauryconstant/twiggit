package cmd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"twiggit/internal/domain"
)

// ErrorFormatter is a composable error formatter using functional composition
type ErrorFormatter struct {
	formatters map[reflect.Type]func(error) string
}

// NewErrorFormatter creates a new error formatter with registered formatters
func NewErrorFormatter() *ErrorFormatter {
	formatter := &ErrorFormatter{
		formatters: make(map[reflect.Type]func(error) string),
	}

	// Register formatters using functional composition
	formatter.registerFormatter((*domain.ValidationError)(nil), formatValidationError)
	formatter.registerFormatter((*domain.WorktreeServiceError)(nil), formatWorktreeError)
	formatter.registerFormatter((*domain.ProjectServiceError)(nil), formatProjectError)
	formatter.registerFormatter((*domain.ServiceError)(nil), formatServiceError)

	return formatter
}

// registerFormatter registers a formatter for a specific error type
func (ef *ErrorFormatter) registerFormatter(errorType error, formatter func(error) string) {
	ef.formatters[reflect.TypeOf(errorType)] = formatter
}

// Format formats an error according to its type using functional composition
func (ef *ErrorFormatter) Format(err error) string {
	if formatter, exists := ef.formatters[reflect.TypeOf(err)]; exists {
		return formatter(err)
	}
	return formatGenericError(err)
}

// Pure formatting functions

// formatValidationError formats ValidationError with emoji indicators and suggestions
func formatValidationError(err error) string {
	validationErr := func() *domain.ValidationError {
		target := &domain.ValidationError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	// Error message with plain text indicator
	output.WriteString(fmt.Sprintf("Error: %s\n", validationErr.Message()))

	// Add suggestions if available
	for _, suggestion := range validationErr.Suggestions() {
		output.WriteString(fmt.Sprintf("Hint: %s\n", suggestion))
	}

	// Add context if available
	if context := validationErr.Context(); context != "" {
		output.WriteString(fmt.Sprintf("Context: %s\n", context))
	}

	return strings.TrimSpace(output.String())
}

// formatWorktreeError formats WorktreeServiceError with emoji indicators
func formatWorktreeError(err error) string {
	worktreeErr := func() *domain.WorktreeServiceError {
		target := &domain.WorktreeServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	if worktreeErr.BranchName != "" {
		output.WriteString(fmt.Sprintf("Error: worktree '%s' not found\n", worktreeErr.BranchName))
	} else {
		output.WriteString(fmt.Sprintf("Error: %s\n", worktreeErr.Message))
	}

	// Add helpful suggestion
	output.WriteString("Hint: Use 'twiggit list' to see available worktrees\n")

	return strings.TrimSpace(output.String())
}

// formatProjectError formats ProjectServiceError with emoji indicators
func formatProjectError(err error) string {
	projectErr := func() *domain.ProjectServiceError {
		target := &domain.ProjectServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	if projectErr.ProjectName != "" {
		output.WriteString(fmt.Sprintf("Error: project '%s' not found\n", projectErr.ProjectName))
	} else {
		output.WriteString(fmt.Sprintf("Error: %s\n", projectErr.Message))
	}

	// Add helpful suggestion
	output.WriteString("Hint: Use 'twiggit list --all' to see available projects\n")

	return strings.TrimSpace(output.String())
}

// formatServiceError formats ServiceError with emoji indicators
func formatServiceError(err error) string {
	serviceErr := func() *domain.ServiceError {
		target := &domain.ServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Error: %s\n", serviceErr.Message))

	// Add contextual suggestions based on operation
	switch serviceErr.Operation {
	case "GetCurrentContext":
		output.WriteString("Hint: Make sure you're in a git repository or worktree directory\n")
	case "DiscoverProject":
		output.WriteString("Hint: Check if the project exists and is accessible\n")
	case "ResolvePath":
		output.WriteString("Hint: Verify the target worktree or project exists\n")
	default:
		output.WriteString("Hint: Check your configuration and try again\n")
	}

	return strings.TrimSpace(output.String())
}

// formatGenericError formats any error with basic plain text formatting
func formatGenericError(err error) string {
	return "Error: " + err.Error()
}
