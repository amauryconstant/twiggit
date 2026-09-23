package cmdutil

import (
	"context"
	"errors"
)

// SignalExitCode inspects a signal-aware context and reports the
// conventional shell exit code (128+N) plus a boolean indicating
// whether the binary should exit immediately.
//
//	(0,   false) — ctx is nil or has no error; caller proceeds normally.
//	(130, true ) — ctx.Err() == context.Canceled (SIGINT).
//	(143, true ) — any other non-nil ctx.Err() (SIGTERM, deadline).
//
// main.go calls this after rootCmd.Execute() and dispatches to
// os.Exit before output.FormatError runs, per cli-error-formatting
// and cli-main-entry-point MODIFIED requirements. The formatter
// MUST NOT be invoked for signal-cancelled runs.
func SignalExitCode(ctx context.Context) (int, bool) {
	if ctx == nil {
		return 0, false
	}
	err := ctx.Err()
	if err == nil {
		return 0, false
	}
	if errors.Is(err, context.Canceled) {
		return 130, true
	}
	return 143, true
}
