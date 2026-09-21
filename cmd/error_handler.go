package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"twiggit/internal/domain"
)

// ErrFlagUsage is the sentinel for cmd-internal usage errors that arise
// after flag parsing (e.g. "init --config requires --install"). cmd
// wrappers can do `fmt.Errorf("%w: <message>", ErrFlagUsage)` to mark an
// error as a usage failure, dispatching it to ExitCodeUsage via
// IsCobraUsageError.
var ErrFlagUsage = errors.New("cmd: flag usage error")

// ExitCode defines the exit codes used by the application.
//
// The canonical mapping is:
//
//	0 ExitCodeSuccess — clean run
//	1 ExitCodeError   — any non-usage failure, runtime error, or recovered panic
//	2 ExitCodeUsage   — cobra/pflag typed usage error
//
// Per-resource NotFound categories share ExitCodeError; they are
// distinguished in the formatter hint layer.
type ExitCode int

const (
	// ExitCodeSuccess indicates clean exit.
	ExitCodeSuccess ExitCode = 0
	// ExitCodeError is the catch-all for non-usage failures.
	ExitCodeError ExitCode = 1
	// ExitCodeUsage indicates a typed cobra/pflag usage error.
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

// IsCobraUsageError reports whether err is a typed pflag usage error.
// Arg-shape failures (cobra.Args validators) are blocked before RunE runs,
// so the typed walk here covers only flag-value errors emitted by pflag
// during flag parsing plus the cmd-internal ErrFlagUsage sentinel.
func IsCobraUsageError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrFlagUsage) {
		return true
	}
	var vre *pflag.ValueRequiredError
	if errors.As(err, &vre) {
		return true
	}
	var ive *pflag.InvalidValueError
	if errors.As(err, &ive) {
		return true
	}
	var ise *pflag.InvalidSyntaxError
	if errors.As(err, &ise) {
		return true
	}
	return false
}

// IsCobraArgumentError is retained as an alias for callers that still
// expect the historical name. New code should call IsCobraUsageError.
func IsCobraArgumentError(err error) bool {
	return IsCobraUsageError(err)
}

// flagParseError wraps cobra's flag-parsing errors in a typed error so
// IsCobraUsageError can match them without falling back to substring
// matching.
type flagParseError struct {
	err error
}

func (w *flagParseError) Error() string {
	if w.err == nil {
		return ""
	}
	return w.err.Error()
}

func (w *flagParseError) Unwrap() error { return w.err }

func (w *flagParseError) Is(target error) bool {
	return target == ErrFlagUsage
}

// WrapFlagError wraps a cobra-emitted flag-parsing error in a
// flagParseError so it dispatches to ExitCodeUsage.
func WrapFlagError(err error) error {
	if err == nil {
		return nil
	}
	return &flagParseError{err: err}
}
