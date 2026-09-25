// Path: cmd/root.go
package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/cmdutil"
)

// NewRootCommand builds the command tree from f. No package globals and no
// init(), so each test builds a fresh tree.
func NewRootCommand(f *cmdutil.Factory) *cobra.Command {
	var verbose, debug bool

	root := &cobra.Command{
		Use:           "myapp",
		Short:         "Manage worktrees",
		SilenceUsage:  true, // runtime errors print one line, not the usage text
		SilenceErrors: true, // main.go formats every error once
		Args:          cmdutil.NoSubcommand,
		RunE:          func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			f.IOStreams.Verbose = verbose
			if debug || os.Getenv("MYAPP_DEBUG") != "" {
				f.IOStreams.Logger = slog.New(slog.NewTextHandler(f.IOStreams.ErrOut,
					&slog.HandlerOptions{Level: slog.LevelDebug}))
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show progress detail")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "log internal diagnostics to stderr")
	root.SetFlagErrorFunc(cmdutil.FlagError)

	root.AddCommand(
		NewCmdCreate(f, nil),
		NewCmdList(f, nil),
		NewCmdVersion(f),
	)
	return root
}
