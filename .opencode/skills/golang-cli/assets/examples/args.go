// Path: internal/cmdutil/args.go
package cmdutil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/core"
)

// Cobra reports flag, argument, and unknown-command mistakes as untyped
// errors. These helpers turn them into *core.UsageError so they exit 2.

// FlagError is the root's flag error func: root.SetFlagErrorFunc(FlagError).
// Subcommands inherit it; UsageArgs reuses it for argument errors.
func FlagError(cmd *cobra.Command, err error) error {
	return &core.UsageError{
		Message: fmt.Sprintf("%v\nRun '%s --help' for usage.", err, cmd.CommandPath()),
		Cause:   err,
	}
}

// UsageArgs wraps a Cobra argument validator: Args: UsageArgs(cobra.ExactArgs(1)).
func UsageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return FlagError(cmd, err)
		}
		return nil
	}
}

// NoSubcommand is the Args validator for commands that only group subcommands
// (root, `myapp worktree`). Pair it with
// RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }
// so the command is runnable and Cobra calls this instead of its own untyped
// "unknown command" check.
func NoSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if s := cmd.SuggestionsFor(args[0]); len(s) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(s, "\n\t")
	}
	return &core.UsageError{Message: msg}
}
