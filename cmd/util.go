package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"twiggit/internal/core"
)

// isQuiet checks if quiet mode is enabled
func isQuiet(cmd *cobra.Command) bool {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}

func logv(cmd *cobra.Command, level int, format string, args ...any) {
	verbosity, _ := cmd.Flags().GetCount("verbose")

	if verbosity < level {
		return
	}

	prefix := ""
	if level > 1 {
		prefix = "  "
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s%s\n", prefix, msg)
}

// ProgressReporter provides progress feedback for bulk operations
type ProgressReporter struct {
	quiet bool
	out   io.Writer
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(quiet bool, out io.Writer) *ProgressReporter {
	return &ProgressReporter{
		quiet: quiet,
		out:   out,
	}
}

// Report outputs a progress message if not in quiet mode
func (p *ProgressReporter) Report(format string, args ...any) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(p.out, format+"\n", args...)
}

// ReportProgress outputs progress for bulk operations
func (p *ProgressReporter) ReportProgress(current, total int, item string) {
	if p.quiet {
		return
	}
	_, _ = fmt.Fprintf(p.out, "[%d/%d] Processing %s\n", current, total, item)
}

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

// keep filepath import live for any helpers that may use it.
var _ = filepath.Abs
