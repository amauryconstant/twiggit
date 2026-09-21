package cmd

import (
	"errors"
	"fmt"
	"strings"

	"twiggit/internal/domain"
)

// asType extracts a typed error from the chain. Mirrors the semantics of
// errors.AsType[T] (Go 1.26+); the local helper exists because the project
// targets go 1.25.5.
func asType[T error](err error) (T, bool) {
	var target T
	if !errors.As(err, &target) {
		return target, false
	}
	return target, true
}

// matcherFunc is a function that checks if an error matches a specific type
type matcherFunc func(error) bool

// formatterFunc is a function that formats an error into a user-friendly string
type formatterFunc func(error) string

func isValidationError(err error) bool {
	var target *domain.ValidationError
	return errors.As(err, &target)
}

func isShellAlreadyInstalledError(err error) bool {
	var target *domain.ShellAlreadyInstalledError
	return errors.As(err, &target)
}

func isShellNotInstalledError(err error) bool {
	var target *domain.ShellNotInstalledError
	return errors.As(err, &target)
}

func isShellInvalidTypeError(err error) bool {
	var target *domain.ShellInvalidTypeError
	return errors.As(err, &target)
}

func isShellInferenceError(err error) bool {
	var target *domain.ShellInferenceError
	return errors.As(err, &target)
}

func isShellDetectionError(err error) bool {
	var target *domain.ShellDetectionError
	return errors.As(err, &target)
}

func isShellWrapperError(err error) bool {
	var target *domain.ShellWrapperError
	return errors.As(err, &target)
}

func isShellConfigError(err error) bool {
	var target *domain.ShellConfigError
	return errors.As(err, &target)
}

func isGitRepositoryError(err error) bool {
	var target *domain.GitRepositoryError
	return errors.As(err, &target)
}

func isGitWorktreeError(err error) bool {
	var target *domain.GitWorktreeError
	return errors.As(err, &target)
}

func isGitCommandError(err error) bool {
	var target *domain.GitCommandError
	return errors.As(err, &target)
}

func isNavigationServiceError(err error) bool {
	var target *domain.NavigationServiceError
	return errors.As(err, &target)
}

func isResolutionError(err error) bool {
	var target *domain.ResolutionError
	return errors.As(err, &target)
}

func isConflictError(err error) bool {
	var target *domain.ConflictError
	return errors.As(err, &target)
}

func isWorktreeError(err error) bool {
	var target *domain.WorktreeServiceError
	return errors.As(err, &target)
}

func isProjectError(err error) bool {
	var target *domain.ProjectServiceError
	return errors.As(err, &target)
}

func isServiceError(err error) bool {
	var target *domain.ServiceError
	return errors.As(err, &target)
}

// hintFor returns the resource-specific hint for a NotFound sentinel, or
// an empty string if the error does not match any registered sentinel.
func hintFor(err error) string {
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		return "Use 'twiggit list --all' to see available projects"
	case errors.Is(err, domain.ErrWorktreeNotFound):
		return "Use 'twiggit list' to see available worktrees"
	case errors.Is(err, domain.ErrResolutionNotFound):
		return "Use 'twiggit list' to see available navigation targets"
	case errors.Is(err, domain.ErrGitRepoNotFound):
		return "Verify the repository path"
	}
	return ""
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

// NewErrorFormatterWithOptions creates a new error formatter with options.
//
// Registration order is specific-before-generic per the
// cli-error-formatting spec's Type-matched dispatch requirement: the
// outermost concrete wrapper in the error chain must win over a
// terminal type the chain passes through (e.g. a WorktreeServiceError
// wrapping a ValidationError should render via the worktree
// formatter, not the validation formatter).
func NewErrorFormatterWithOptions(quiet bool) *ErrorFormatter {
	formatter := &ErrorFormatter{
		quiet: quiet,
	}

	// Shell subtypes (most specific)
	formatter.register(isShellAlreadyInstalledError, formatShellAlreadyInstalledError)
	formatter.register(isShellNotInstalledError, formatShellNotInstalledError)
	formatter.register(isShellInvalidTypeError, formatShellInvalidTypeError)
	formatter.register(isShellInferenceError, formatShellInferenceError)
	formatter.register(isShellDetectionError, formatShellDetectionError)
	formatter.register(isShellWrapperError, formatShellWrapperError)
	formatter.register(isShellConfigError, formatShellConfigError)
	// Git errors
	formatter.register(isGitRepositoryError, formatGitRepositoryError)
	formatter.register(isGitWorktreeError, formatGitWorktreeError)
	formatter.register(isGitCommandError, formatGitCommandError)
	// Navigation/resolution/conflict
	formatter.register(isNavigationServiceError, formatNavigationServiceError)
	formatter.register(isResolutionError, formatResolutionError)
	formatter.register(isConflictError, formatConflictError)
	// Service wrappers (specific)
	formatter.register(isWorktreeError, formatWorktreeError)
	formatter.register(isProjectError, formatProjectError)
	// Validation (terminal, checked after service wrappers so a
	// Worktree{Err: Validation} chain renders via the worktree formatter)
	formatter.register(isValidationError, formatValidationError)
	// Generic fallback
	formatter.register(isServiceError, formatServiceError)

	return formatter
}

// register registers a matcher-formatter pair.
// Matchers are checked in registration order.
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
			lines := strings.Split(output, "\n")
			filtered := make([]string, 0, len(lines))
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
	if err == nil {
		return ""
	}
	for _, mf := range ef.matchers {
		if mf.matcher(err) {
			return mf.formatter(err)
		}
	}
	return ef.formatGenericError(err)
}

func formatValidationError(err error) string {
	validationErr, ok := asType[*domain.ValidationError](err)
	if !ok {
		return ""
	}
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Error: %s\n", validationErr.Message()))
	for _, suggestion := range validationErr.Suggestions() {
		output.WriteString(fmt.Sprintf("Hint: %s\n", suggestion))
	}
	if context := validationErr.Context(); context != "" {
		output.WriteString(fmt.Sprintf("Context: %s\n", context))
	}
	return output.String()
}

func formatShellAlreadyInstalledError(err error) string {
	if _, ok := asType[*domain.ShellAlreadyInstalledError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func formatShellNotInstalledError(err error) string {
	if _, ok := asType[*domain.ShellNotInstalledError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func formatShellInvalidTypeError(err error) string {
	if _, ok := asType[*domain.ShellInvalidTypeError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\nHint: Supported shells: bash, zsh, fish\n", err.Error())
}

func formatShellInferenceError(err error) string {
	if _, ok := asType[*domain.ShellInferenceError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func formatShellDetectionError(err error) string {
	if _, ok := asType[*domain.ShellDetectionError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func formatShellWrapperError(err error) string {
	if _, ok := asType[*domain.ShellWrapperError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\nHint: Check the wrapper script and shell config file\n", err.Error())
}

func formatShellConfigError(err error) string {
	if _, ok := asType[*domain.ShellConfigError](err); !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

func formatGitRepositoryError(err error) string {
	gitErr, ok := asType[*domain.GitRepositoryError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", gitErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatGitWorktreeError(err error) string {
	gitErr, ok := asType[*domain.GitWorktreeError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", gitErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatGitCommandError(err error) string {
	gitErr, ok := asType[*domain.GitCommandError](err)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", gitErr.Error())
}

func formatNavigationServiceError(err error) string {
	navErr, ok := asType[*domain.NavigationServiceError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", navErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatResolutionError(err error) string {
	resErr, ok := asType[*domain.ResolutionError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", resErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatConflictError(err error) string {
	conflictErr, ok := asType[*domain.ConflictError](err)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Error: %s\n", conflictErr.Error())
}

func formatWorktreeError(err error) string {
	worktreeErr, ok := asType[*domain.WorktreeServiceError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", worktreeErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatProjectError(err error) string {
	projectErr, ok := asType[*domain.ProjectServiceError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", projectErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatServiceError(err error) string {
	serviceErr, ok := asType[*domain.ServiceError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", serviceErr.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	} else {
		output += "Hint: Check your configuration and try again\n"
	}
	return output
}

// formatGenericError formats any error with basic plain text formatting
func (ef *ErrorFormatter) formatGenericError(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}
