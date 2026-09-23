package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"twiggit/internal/core"
	"twiggit/internal/output"
)

// ExitCode defines the exit codes used by the application.
//
// The canonical mapping is:
//
//	0 ExitCodeSuccess — clean run
//	1 ExitCodeError   — any non-usage failure, runtime error, or recovered panic
//	2 ExitCodeUsage   — typed core.UsageError
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

// HandleCLIError is a pure function that maps errors to CLI output and returns exit code.
// It writes through the system IOStreams so non-root callers (early
// startup failures, pre-cobra errors) still see formatted output.
func HandleCLIError(err error) ExitCode {
	return HandleCLIErrorWithCommand(nil, err)
}

// HandleCLIErrorWithCommand maps errors to CLI output and returns exit code.
// Formatting runs through output.FormatError so colour, hints, and the
// TWIGGIT_DEBUG chain dump all funnel through the same renderer used
// by main.go. The iostreams.IOStreams is recovered from cmd's
// context via cmdutil.IOStreamsFromCmd; see cmd/util.go:iosFromCmd for
// the lookup helper.
func HandleCLIErrorWithCommand(cmd *cobra.Command, err error) ExitCode {
	ios := iosFromCmd(cmd)

	// Usage errors get a clean "Error: <message>" line so cobra's
	// pflag-wrapped usage failures do not stack with the structured
	// formatter's multi-line hints.
	if IsCobraUsageError(err) {
		fmt.Fprintf(ios.ErrOut, "Error: %s\n", err.Error())
		return ExitCodeUsage
	}

	quiet := isQuiet(cmd)
	if quiet {
		ios.Quiet = true
	}

	output.FormatError(ios.ErrOut, err, ios)
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
	if errors.Is(err, core.ErrGitRepoNotFound) ||
		errors.Is(err, core.ErrWorktreeNotFound) ||
		errors.Is(err, core.ErrProjectNotFound) ||
		errors.Is(err, core.ErrResolutionNotFound) {
		return ErrorCategoryService
	}
	if errors.As(err, new(*core.ValidationError)) ||
		errors.As(err, new(*core.OperationError)) ||
		errors.As(err, new(*core.NotFoundError)) {
		return ErrorCategoryService
	}
	return ErrorCategoryGeneric
}

// IsCobraUsageError reports whether err is a typed domain UsageError.
// cmd/root.go's SetFlagErrorFunc wraps cobra/pflag flag-parse errors
// in *core.UsageError before they propagate, so this single
// errors.As match covers all flag-validation paths.
func IsCobraUsageError(err error) bool {
	if err == nil {
		return false
	}
	var ue *core.UsageError
	return errors.As(err, &ue)
}

// IsCobraArgumentError is retained as an alias for callers that still
// expect the historical name. New code should call IsCobraUsageError.
func IsCobraArgumentError(err error) bool {
	return IsCobraUsageError(err)
}
