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

	// Suppress Cobra's built-in `completion` subcommand so carapace's
	// auto-registered one is the only completion surface. The legacy
	// Find/RemoveCommand dance below is now redundant.
	cmd.CompletionOptions.DisableDefaultCmd = true

	// Persistent --output / --quiet / --verbose (slice 10 contract).
	cmdutil.AddPersistentFlags(cmd, globalOpts)

	// Shell completion for the persistent --output flag. The legacy
	// "text" and the dropped "jsonl" values are intentionally absent
	// so completion-driven callers cannot reintroduce them. The
	// completion function MUST be registered on the root command
	// because pflag does not propagate Flag.Annotations to
	// subcommand copies of persistent flags; a per-subcommand
	// RegisterFlagCompletionFunc would silently be ignored.
	_ = cmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "table", "plain"}, cobra.ShellCompDirectiveNoFileComp
	})

	// AddGroup labels surface in `twiggit --help` to keep the
	// command tree navigable as it grows. Cobra does not
	// retroactively assign groups, so register the group
	// definitions first and set GroupID on each subcommand before
	// AddCommand.
	cmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Core:"},
		&cobra.Group{ID: "navigation", Title: "Navigation:"},
		&cobra.Group{ID: "setup", Title: "Setup:"},
		&cobra.Group{ID: "meta", Title: "Meta:"},
	)

	// Subcommand factories. Each command is assigned to its group
	// via GroupID so `twiggit --help` shows the four group headers.
	listCmd := NewCmdList(f, nil)
	listCmd.GroupID = "core"
	createCmd := NewCmdCreate(f, nil)
	createCmd.GroupID = "core"
	deleteCmd := NewCmdDelete(f, nil)
	deleteCmd.GroupID = "core"
	pruneCmd := NewCmdPrune(f, nil)
	pruneCmd.GroupID = "core"
	rebaseCmd := NewCmdRebase(f, nil)
	rebaseCmd.GroupID = "core"
	syncCmd := NewCmdSync(f, nil)
	syncCmd.GroupID = "core"
	cdCmd := NewCmdCd(f, nil)
	cdCmd.GroupID = "navigation"
	initCmd := NewCmdInit(f, nil)
	initCmd.GroupID = "setup"
	versionCmd := NewCmdVersion(f, nil)
	versionCmd.GroupID = "meta"
	completionCmd := newCompletionCommand(cmd)
	completionCmd.GroupID = "meta"
	cmd.AddCommand(
		listCmd,
		createCmd,
		deleteCmd,
		pruneCmd,
		rebaseCmd,
		syncCmd,
		cdCmd,
		initCmd,
		versionCmd,
		completionCmd,
	)

	carapace.Gen(cmd)

	return cmd
}
