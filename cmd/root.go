package cmd

import (
	"github.com/amaury/twiggit/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for twiggit
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "twiggit",
		Short: "Simple Git worktree and project management",
		Long: `twiggit is a fast and intuitive tool for managing Git worktrees and projects.

It provides simple commands to switch between projects and worktrees,
create new worktrees, list existing ones, and clean up when done.

Perfect for developers who work with multiple branches across different projects.`,
		Version: version.Version(),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	// Add subcommands (ordered by usage frequency)
	rootCmd.AddCommand(NewSwitchCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewDeleteCmd())

	return rootCmd
}
