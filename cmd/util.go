package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"
)

// iosFromCmd returns the *iostreams.IOStreams associated with cmd.
// Thin wrapper around cmdutil.IOStreamsFromCmd so error_handler.go and
// any local helper can share one lookup path.
func iosFromCmd(cmd *cobra.Command) *iostreams.IOStreams {
	return cmdutil.IOStreamsFromCmd(cmd)
}

// isQuiet is kept as a small accessor for run paths that need the
// persistent --quiet flag before opts is fully populated. Prefers the
// flag value when present; falls back to ios.Quiet.
func isQuiet(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if cmd.Flags().Lookup("quiet") != nil {
		q, _ := cmd.Flags().GetBool("quiet")
		return q
	}
	return iosFromCmd(cmd).Quiet
}

// activeVerboseLevel mirrors the running --verbose count.
// NewCmd* functions install the Factory's *GlobalOptions value via
// root's PersistentPreRunE; tests that drive runX directly can call
// setActiveVerboseLevel for the duration of the test (t.Cleanup
// restores). Zero is the silent default.
var activeVerboseLevel int

// setActiveVerboseLevel pins the verbose count. Returns a restore
// function for t.Cleanup-style usage. The active count gates every
// direct ios.Verbosef call wrapped by verbosef below.
func setActiveVerboseLevel(level int) func() {
	prev := activeVerboseLevel
	activeVerboseLevel = level
	return func() { activeVerboseLevel = prev }
}

// verbosef is the verbose-output gate used by every command. It
// honours both the iostreams-supplied --quiet/--verbose gates and
// the level-aware count installed by root.PersistentPreRunE.
// -v sets activeVerboseLevel = 1; -vv sets it = 2. With no flag
// set, verbosef is a no-op. When ios is nil the call is silently
// dropped.
func verbosef(ios *iostreams.IOStreams, level int, format string, args ...any) {
	if ios == nil {
		return
	}
	if level > activeVerboseLevel {
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
	fmt.Fprintf(writeOrIgnore(p.out), format+"\n", args...)
}

// ReportProgress outputs progress for bulk operations
func (p *ProgressReporter) ReportProgress(current, total int, item string) {
	if p.quiet {
		return
	}
	fmt.Fprintf(writeOrIgnore(p.out), "[%d/%d] Processing %s\n", current, total, item)
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

// resolveNavigationTarget is kept as a small helper used by tests that
// exercise the cd-style resolution flow. The actual cd command no
// longer uses it — see executeCD in cmd/cd.go.
func resolveNavigationTarget(_ context.Context, _ *CommandConfig, _ string) (*core.Context, *core.ResolutionResult, error) {
	return nil, nil, fmt.Errorf("resolveNavigationTarget: moved to executeCD")
}
