package cmd

import (
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
