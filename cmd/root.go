package cmd

import (
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
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
// the cached Config and GitClient fields. The IOStreams and the
// persistent --output / --quiet / --verbose flags are bound on the
// root and stashed on cmd's context so error_handler.go and
// completion helpers can recover them via cmdutil.IOStreamsFromCmd.
func NewRootCommand(f *CommandConfig) *cobra.Command {
	globalOpts := &cmdutil.GlobalOptions{}
	f.GlobalOptions = globalOpts

	cmd := &cobra.Command{
		Use:   "twiggit",
		Short: "A pragmatic tool for managing git worktrees",
		Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees
across multiple projects.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if f == nil {
				return core.NewUsageError("cmd: factory not provided", nil)
			}
			// Stash IOStreams on cmd's context so error paths and
			// completion helpers can recover them without carrying
			// the Factory reference through every signature.
			cmdutil.SetIOStreams(cmd, f.IOStreams)
			// Mirror the persistent --output / --quiet / --verbose
			// flags onto the IOStreams gates so Verbosef and the
			// quiet-aware formatter honour the user-supplied values.
			globalOpts.ApplyToIOS(f.IOStreams)
			return nil
		},
	}

	// Wrap flag-parse errors in core.UsageError so IsCobraUsageError
	// dispatches them to ExitCodeUsage without substring matching.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return core.UsageWrap(err)
	})

	// Persistent --output / --quiet / --verbose (slice 10 contract).
	cmdutil.AddPersistentFlags(cmd, globalOpts)

	// Add subcommands
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdCreate(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	cmd.AddCommand(NewCmdPrune(f, nil))
	cmd.AddCommand(NewCmdCd(f, nil))
	cmd.AddCommand(NewCmdInit(f, nil))
	cmd.AddCommand(NewCmdVersion(f, nil))

	carapace.Gen(cmd)

	// Replace Cobra's default completion command with Carapace's version
	if completionCmd, _, err := cmd.Find([]string{"completion"}); err == nil {
		cmd.RemoveCommand(completionCmd)
	}
	cmd.AddCommand(newCompletionCommand(cmd))

	return cmd
}
