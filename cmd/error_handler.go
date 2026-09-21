package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"twiggit/internal/domain"
)

// ExitCode defines the exit codes used by the application.
//
// The canonical mapping is:
//
//	0 ExitCodeSuccess — clean run
//	1 ExitCodeError   — any non-usage failure, runtime error, or recovered panic
//	2 ExitCodeUsage   — typed domain.UsageError
//
// Per-resource NotFound categories share ExitCodeError; they are
// distinguished in the formatter hint layer.
type ExitCode int

const (
	// ExitCodeSuccess indicates clean exit.
	ExitCodeSuccess ExitCode = 0
	// ExitCodeError is the catch-all for non-usage failures.
	ExitCodeError ExitCode = 1
	// ExitCodeUsage indicates a typed UsageError (cobra/pflag wrapped).
	ExitCodeUsage ExitCode = 2
)

// ErrorCategory groups errors for dispatch in the formatter and exit-code
// mapping. The three remaining categories map onto the three exit codes.
type ErrorCategory int

const (
	// ErrorCategoryCobra marks cobra usage errors.
	ErrorCategoryCobra ErrorCategory = iota
	// ErrorCategoryService marks service-class runtime errors.
	ErrorCategoryService
	// ErrorCategoryGeneric marks unclassified errors.
	ErrorCategoryGeneric
)

// HandleCLIError is a pure function that maps errors to CLI output and returns exit code
func HandleCLIError(err error) ExitCode {
	return HandleCLIErrorWithCommand(nil, err)
}

// HandleCLIErrorWithCommand maps errors to CLI output and returns exit code, respecting quiet mode from command
func HandleCLIErrorWithCommand(cmd *cobra.Command, err error) ExitCode {
	if IsCobraUsageError(err) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return ExitCodeUsage
	}

	quiet := false
	if cmd != nil {
		quiet = isQuiet(cmd)
	}

	formatter := NewErrorFormatterWithOptions(quiet)
	formattedError := formatter.Format(err)
	fmt.Fprint(os.Stderr, formattedError)

	return GetExitCodeForError(err)
}

// GetExitCodeForError maps errors to the three-code exit contract.
func GetExitCodeForError(err error) ExitCode {
	if err == nil {
		return ExitCodeSuccess
	}
	if IsCobraUsageError(err) {
		return ExitCodeUsage
	}
	return ExitCodeError
}

// CategorizeError determines the category of an error for consistent handling
func CategorizeError(err error) ErrorCategory {
	if IsCobraUsageError(err) {
		return ErrorCategoryCobra
	}
	if errors.Is(err, domain.ErrGitRepoNotFound) ||
		errors.Is(err, domain.ErrWorktreeNotFound) ||
		errors.Is(err, domain.ErrProjectNotFound) ||
		errors.Is(err, domain.ErrResolutionNotFound) {
		return ErrorCategoryService
	}
	if errors.As(err, new(*domain.ValidationError)) ||
		errors.As(err, new(*domain.WorktreeServiceError)) ||
		errors.As(err, new(*domain.ProjectServiceError)) ||
		errors.As(err, new(*domain.ServiceError)) {
		return ErrorCategoryService
	}
	return ErrorCategoryGeneric
}

// IsCobraUsageError reports whether err is a typed domain UsageError.
// cmd/root.go's SetFlagErrorFunc wraps cobra/pflag flag-parse errors
// in *domain.UsageError before they propagate, so this single
// errors.As match covers all flag-validation paths.
func IsCobraUsageError(err error) bool {
	if err == nil {
		return false
	}
	var ue *domain.UsageError
	return errors.As(err, &ue)
}

// IsCobraArgumentError is retained as an alias for callers that still
// expect the historical name. New code should call IsCobraUsageError.
func IsCobraArgumentError(err error) bool {
	return IsCobraUsageError(err)
}
