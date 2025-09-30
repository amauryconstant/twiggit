// Package cmd contains the CLI commands for twiggit.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "twiggit",
	Short: "Pragmatic git worktree management tool",
	Long: `Twiggit is a pragmatic tool for managing git worktrees 
with a focus on rebase workflows.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
