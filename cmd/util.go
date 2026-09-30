package cmd

import (
	"fmt"
	"io"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"

	"github.com/spf13/cobra"
)

// verbosef is the verbose-output gate used by every command. It is
// a thin wrapper around iostreams.IOStreams.Verbosef that tolerates
// a nil ios for early-startup call sites and keeps the call shape
// short. Per cli-verbose-output, the previous two-level
// (-v / -vv) distinction is collapsed into a single boolean Verbose
// flag: -vv is treated as -v and the level-2 indentation prefix is
// removed from message strings.
func verbosef(ios *iostreams.IOStreams, format string, args ...any) {
	if ios == nil {
		return
	}
	ios.Verbosef(format, args...)
}

// ProgressReporter provides progress feedback for bulk operations.
// It honours the iostreams.Quiet gate rather than reading the flag
// directly so the constructor composes cleanly with the test-only
// iostreams.Test() helper.
type ProgressReporter struct {
	quiet bool
	out   io.Writer
}

// NewProgressReporter creates a new progress reporter; pass ios to
// inherit its Quiet flag and ErrOut writer.
func NewProgressReporter(ios *iostreams.IOStreams) *ProgressReporter {
	if ios == nil {
		return &ProgressReporter{quiet: false, out: io.Discard}
	}
	return &ProgressReporter{quiet: ios.Quiet, out: ios.ErrOut}
}

// Report outputs a progress message if not in quiet mode
func (p *ProgressReporter) Report(format string, args ...any) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(writeOrIgnore(p.out), format+"\n", args...)
}

// ignoreWriter wraps an io.Writer so its Write always reports success.
//
// Per the swallowed-error policy (modernization sweep, task 16.5),
// terminal writes to user-visible streams are intentionally best-effort:
// a broken pipe or closed TTY cannot be meaningfully recovered from
// inside a CLI, and logging the error via slog.Error would re-target
// the same broken stream. Use writeOrIgnore to wrap the destination so
// the call site does not need a `_, _ =` lint pattern.
type ignoreWriter struct{ io.Writer }

// Write delegates to the wrapped writer and discards the error.
func (w ignoreWriter) Write(p []byte) (int, error) {
	n, _ := w.Writer.Write(p)
	return n, nil
}

// writeOrIgnore returns an io.Writer whose Write never errors.
// See ignoreWriter docs.
func writeOrIgnore(w io.Writer) io.Writer { return ignoreWriter{w} }

// wrapArgsValidator wraps a cobra.PositionalArgs validator so any error it
// returns is converted to *core.UsageError. This routes args-shape failures
// (e.g., cobra.ExactArgs(1) on `create` with 0 args) through the same exit-code
// dispatch path as flag-parse errors, so they yield ExitCodeUsage (2) instead
// of the default ExitCodeError (1). Cobra v1.10 does not expose an
// args-equivalent of SetFlagErrorFunc, so we wrap at each command's Args field.
func wrapArgsValidator(v cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := v(cmd, args); err != nil {
			return core.UsageWrap(err)
		}
		return nil
	}
}
