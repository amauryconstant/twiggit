package cmd

import (
	"github.com/amaury/twiggit/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for twiggit
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "twiggit",
		Short: "Modern Git worktree management",
		Long: `twiggit is a modern, fast, and intuitive Git worktree management tool
that replaces complex bash scripts with a maintainable Go application.

Features:
  - Interactive worktree creation and management
  - Rich terminal UI with colors and progress indicators
  - XDG-compliant configuration with environment variable support
  - Shell completion for multiple shells
  - Template-based workflows`,
		Version: version.Version(),
	}

	// Add persistent flags
	rootCmd.PersistentFlags().String("workspace", "", "Override workspace directory")
	rootCmd.PersistentFlags().String("project", "", "Override project name")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	// Add subcommands
	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewSwitchCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewListCmd())

	return rootCmd
}
