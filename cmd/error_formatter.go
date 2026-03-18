package cmd

import (
	"errors"
	"fmt"
	"strings"

	"twiggit/internal/domain"
)

// matcherFunc is a function that checks if an error matches a specific type
type matcherFunc func(error) bool

// formatterFunc is a function that formats an error into a user-friendly string
type formatterFunc func(error) string

// isValidationError checks if error is a ValidationError using errors.As()
func isValidationError(err error) bool {
	var target *domain.ValidationError
	return errors.As(err, &target)
}

// isWorktreeError checks if error is a WorktreeServiceError using errors.As()
func isWorktreeError(err error) bool {
	var target *domain.WorktreeServiceError
	return errors.As(err, &target)
}

// isProjectError checks if error is a ProjectServiceError using errors.As()
func isProjectError(err error) bool {
	var target *domain.ProjectServiceError
	return errors.As(err, &target)
}

// isServiceError checks if error is a ServiceError using errors.As()
func isServiceError(err error) bool {
	var target *domain.ServiceError
	return errors.As(err, &target)
}

// ErrorFormatter is a composable error formatter using explicit strategy pattern
type ErrorFormatter struct {
	matchers []struct {
		matcher   matcherFunc
		formatter formatterFunc
	}
	quiet bool // Suppress hint messages in quiet mode
}

// NewErrorFormatter creates a new error formatter with registered formatters
func NewErrorFormatter() *ErrorFormatter {
	return NewErrorFormatterWithOptions(false)
}

// NewErrorFormatterWithOptions creates a new error formatter with options
func NewErrorFormatterWithOptions(quiet bool) *ErrorFormatter {
	formatter := &ErrorFormatter{
		quiet: quiet,
	}

	// Register formatters using explicit strategy pattern
	// Order matters: more specific matchers should come first
	formatter.register(isValidationError, formatValidationError)
	formatter.register(isWorktreeError, formatWorktreeError)
	formatter.register(isProjectError, formatProjectError)
	formatter.register(isServiceError, formatServiceError)

	return formatter
}

// register registers a matcher-formatter pair
// Matchers are checked in registration order
func (ef *ErrorFormatter) register(matcher matcherFunc, formatter formatterFunc) {
	ef.matchers = append(ef.matchers, struct {
		matcher   matcherFunc
		formatter formatterFunc
	}{matcher, ef.withQuietMode(formatter)})
}

// withQuietMode wraps a formatter to respect quiet mode
// In quiet mode, removes hint lines from the output
func (ef *ErrorFormatter) withQuietMode(formatter formatterFunc) formatterFunc {
	return func(err error) string {
		output := formatter(err)
		if ef.quiet {
			// Remove hint lines in quiet mode
			lines := strings.Split(output, "\n")
			var filtered []string
			for _, line := range lines {
				if !strings.HasPrefix(line, "Hint:") {
					filtered = append(filtered, line)
				}
			}
			return strings.Join(filtered, "\n")
		}
		return output
	}
}

// Format formats an error according to its type using explicit strategy pattern
func (ef *ErrorFormatter) Format(err error) string {
	// Iterate through matchers in registration order
	for _, mf := range ef.matchers {
		if mf.matcher(err) {
			return mf.formatter(err)
		}
	}
	return ef.formatGenericError(err)
}

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

	// Add suggestions if available (quiet mode is handled by wrapper)
	for _, suggestion := range validationErr.Suggestions() {
		output.WriteString(fmt.Sprintf("Hint: %s\n", suggestion))
	}

	// Add context if available
	if context := validationErr.Context(); context != "" {
		output.WriteString(fmt.Sprintf("Context: %s\n", context))
	}

	return output.String()
}

// formatWorktreeError formats WorktreeServiceError with actionable hints
func formatWorktreeError(err error) string {
	worktreeErr := func() *domain.WorktreeServiceError {
		target := &domain.WorktreeServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	// The Error() method already provides user-friendly messages
	output.WriteString(fmt.Sprintf("Error: %s\n", worktreeErr.Error()))

	// Add helpful suggestion based on error type (quiet mode is handled by wrapper)
	if worktreeErr.IsNotFound() {
		output.WriteString("Hint: Use 'twiggit list' to see available worktrees\n")
	} else {
		output.WriteString("Hint: Check that worktree exists and you have permission\n")
	}

	return output.String()
}

// formatProjectError formats ProjectServiceError with actionable hints
func formatProjectError(err error) string {
	projectErr := func() *domain.ProjectServiceError {
		target := &domain.ProjectServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	// The Error() method already provides user-friendly messages
	output.WriteString(fmt.Sprintf("Error: %s\n", projectErr.Error()))

	// Add helpful suggestion (quiet mode is handled by wrapper)
	output.WriteString("Hint: Use 'twiggit list --all' to see available projects\n")

	return output.String()
}

// formatServiceError formats ServiceError with actionable hints
func formatServiceError(err error) string {
	serviceErr := func() *domain.ServiceError {
		target := &domain.ServiceError{}
		_ = errors.As(err, &target)
		return target
	}()
	var output strings.Builder

	// The Error() method now returns just the message without operation names
	output.WriteString(fmt.Sprintf("Error: %s\n", serviceErr.Error()))

	// Add a generic helpful suggestion (quiet mode is handled by wrapper)
	output.WriteString("Hint: Check your configuration and try again\n")

	return output.String()
}

// formatGenericError formats any error with basic plain text formatting
func (ef *ErrorFormatter) formatGenericError(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}
