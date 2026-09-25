package cmdutil

import (
	"errors"
	"strconv"
	"twiggit/internal/core"
)

// ExitCode is the typed process exit code. The three values below are the
// only codes any caller of ExitCodeFor can observe; signal-driven exits
// (SIGINT → 130, SIGTERM → 143) bypass this type and write the code
// directly via os.Exit at the main boundary.
type ExitCode int

const (
	// ExitOK signals successful command completion (0).
	ExitOK ExitCode = 0
	// ExitError signals a non-usage runtime failure (1).
	ExitError ExitCode = 1
	// ExitUsage signals an invocation-level usage failure (2): invalid
	// flag combination, missing required value, malformed argument.
	ExitUsage ExitCode = 2
)

// ExitCodeFor maps an error to the exit code the main loop should return.
// UsageError is the only type that resolves to ExitUsage; nil resolves to
// ExitOK; every other error is ExitError. Wrapped UsageErrors are unwrapped
// via errors.As before classification.
//
// UsageError.Unwrap returns nil (per the core-errors spec), so a
// UsageError buried under a fmt.Errorf("%w") wrap still matches because
// errors.As walks the full chain.
func ExitCodeFor(err error) ExitCode {
	if err == nil {
		return ExitOK
	}
	if _, ok := errors.AsType[*core.UsageError](err); ok {
		return ExitUsage
	}
	return ExitError
}

// String renders the exit code as its integer text ("0", "1", "2"). Kept
// stable so logs, scripts, and tests can compare exact text rather than
// re-implementing the formatting inline.
func (c ExitCode) String() string {
	return strconv.Itoa(int(c))
}
