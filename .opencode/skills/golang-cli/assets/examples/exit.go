// Path: internal/cmdutil/exit.go
package cmdutil

import (
	"errors"

	"github.com/you/myapp/internal/core"
)

type ExitCode int

const (
	ExitOK        ExitCode = 0
	ExitError     ExitCode = 1   // any runtime failure
	ExitUsage     ExitCode = 2   // invalid flags, arguments, or missing input
	ExitInterrupt ExitCode = 130 // 128 + SIGINT; returned by run() in main.go
)

// ExitCodeFor maps a command error to an exit code. The error type carries the
// meaning for formatting; the code stays a blunt success/usage/failure signal.
func ExitCodeFor(err error) ExitCode {
	var usage *core.UsageError
	switch {
	case err == nil:
		return ExitOK
	case errors.As(err, &usage):
		return ExitUsage
	default:
		return ExitError
	}
}
