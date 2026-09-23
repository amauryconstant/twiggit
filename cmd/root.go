package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
)

// CommandConfig is the configuration handle every command constructor
// receives. After slice 9 it is an alias for the cmdutil.Factory: the
// composition root (main.go) builds a Factory once and passes it to
// every command, so runX functions can lazy-resolve git client,
// config, IO streams, and logger without service-layer indirection.
type CommandConfig = cmdutil.Factory

// NewRootCommand creates a new root command with the given factory.
//
// Every subcommand receives the same Factory pointer so they all share
// the cached Config and GitClient fields.
func NewRootCommand(f *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twiggit",
		Short: "A pragmatic tool for managing git worktrees",
		Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees
across multiple projects.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if f == nil {
				return core.NewUsageError("cmd: factory not provided", nil)
			}
			return nil
		},
	}

	// Wrap flag-parse errors in core.UsageError so IsCobraUsageError
	// dispatches them to ExitCodeUsage without substring matching.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return core.UsageWrap(err)
	})

	// Add persistent verbose flag
	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (can be used multiple times: -v, -vv)")

	// Add persistent quiet flag
	cmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	// Add subcommands
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewCreateCommand(f))
	cmd.AddCommand(NewDeleteCommand(f))
	cmd.AddCommand(NewPruneCommand(f))
	cmd.AddCommand(NewCDCommand(f))
	cmd.AddCommand(NewInitCmd(f))
	cmd.AddCommand(NewVersionCommand(f))

	carapace.Gen(cmd)

	// Replace Cobra's default completion command with Carapace's version
	if completionCmd, _, err := cmd.Find([]string{"completion"}); err == nil {
		cmd.RemoveCommand(completionCmd)
	}
	cmd.AddCommand(newCompletionCommand(cmd))

	return cmd
}
