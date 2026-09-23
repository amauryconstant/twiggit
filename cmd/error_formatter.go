package cmd

import (
	"errors"
	"fmt"
	"strings"

	"twiggit/internal/core"
)

// asType extracts a typed error from the chain. Mirrors the semantics of
// errors.AsType[T] (Go 1.26+); the local helper exists because the project
// targets Go 1.25.
func asType[T error](err error) (T, bool) {
	var target T
	if !errors.As(err, &target) {
		return target, false
	}
	return target, true
}

// hintFor returns the resource-specific hint for a NotFound sentinel, or
// an empty string if the error does not match any registered sentinel.
func hintFor(err error) string {
	switch {
	case errors.Is(err, core.ErrProjectNotFound):
		return "Use 'twiggit list --all' to see available projects"
	case errors.Is(err, core.ErrWorktreeNotFound):
		return "Use 'twiggit list' to see available worktrees"
	case errors.Is(err, core.ErrResolutionNotFound):
		return "Use 'twiggit list' to see available navigation targets"
	case errors.Is(err, core.ErrGitRepoNotFound):
		return "Verify the repository path"
	}
	return ""
}

// ErrorFormatter is a composable error formatter using explicit strategy pattern.
type ErrorFormatter struct {
	matchers []struct {
		matcher   matcherFunc
		formatter formatterFunc
	}
	quiet bool
}

// NewErrorFormatter creates a new error formatter with registered formatters.
func NewErrorFormatter() *ErrorFormatter {
	return NewErrorFormatterWithOptions(false)
}

// NewErrorFormatterWithOptions creates a new error formatter with options.
//
// Dispatch order: OperationError → NotFoundError → ValidationError.
// OperationError is checked first so a wrapper that embeds a
// ValidationError in its Cause still renders via the wrapper's own
// Message + Op context. NotFoundError catches the four canonical
// not-found sentinels before falling through to ValidationError,
// which renders suggestions and (legacy) field context for leaf
// validation failures.
func NewErrorFormatterWithOptions(quiet bool) *ErrorFormatter {
	formatter := &ErrorFormatter{
		quiet: quiet,
	}

	formatter.register(isOperationError, formatOperationError)
	formatter.register(isNotFoundError, formatNotFoundError)
	formatter.register(isValidationError, formatValidationError)

	return formatter
}

func isValidationError(err error) bool {
	var target *core.ValidationError
	return errors.As(err, &target)
}

func isNotFoundError(err error) bool {
	var target *core.NotFoundError
	return errors.As(err, &target)
}

func isOperationError(err error) bool {
	var target *core.OperationError
	return errors.As(err, &target)
}

// matcherFunc is a function that checks if an error matches a specific type
type matcherFunc func(error) bool

// formatterFunc is a function that formats an error into a user-friendly string
type formatterFunc func(error) string

// register registers a matcher-formatter pair.
func (ef *ErrorFormatter) register(matcher matcherFunc, formatter formatterFunc) {
	ef.matchers = append(ef.matchers, struct {
		matcher   matcherFunc
		formatter formatterFunc
	}{matcher, ef.withQuietMode(formatter)})
}

// withQuietMode wraps a formatter to respect quiet mode.
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
	ve, ok := asType[*core.ValidationError](err)
	if !ok {
		return ""
	}
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Error: %s\n", ve.Error()))
	for _, suggestion := range ve.Suggestions {
		output.WriteString(fmt.Sprintf("Hint: %s\n", suggestion))
	}
	if hint := hintFor(err); hint != "" {
		output.WriteString(fmt.Sprintf("Hint: %s\n", hint))
	}
	return output.String()
}

func formatNotFoundError(err error) string {
	nfe, ok := asType[*core.NotFoundError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", nfe.Error())
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

func formatOperationError(err error) string {
	oe, ok := asType[*core.OperationError](err)
	if !ok {
		return ""
	}
	output := fmt.Sprintf("Error: %s\n", oe.Error())
	var outputSb164 strings.Builder
	for _, suggestion := range oe.Suggestions {
		outputSb164.WriteString(fmt.Sprintf("Hint: %s\n", suggestion))
	}
	output += outputSb164.String()
	if hint := hintFor(err); hint != "" {
		output += fmt.Sprintf("Hint: %s\n", hint)
	}
	return output
}

// formatGenericError formats any error with basic plain text formatting
func (ef *ErrorFormatter) formatGenericError(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}
