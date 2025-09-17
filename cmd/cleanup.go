// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCleanupCmd creates the cleanup command
func NewCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove unused or stale worktrees",
		Long: `Clean up worktrees that are no longer needed.

This command helps identify and remove:
- Worktrees with merged branches
- Stale worktrees that haven't been used recently
- Worktrees pointing to deleted branches

Safety features:
- Interactive confirmation before deletion
- Backup option for uncommitted changes
- Dry-run mode to preview actions

Examples:
  twiggit cleanup
  twiggit cleanup --dry-run
  twiggit cleanup --force --merged-only`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// TODO: Implement cleanup logic
			fmt.Println("Cleanup command - not yet implemented")
			return nil
		},
	}

	// Add flags specific to cleanup
	cmd.Flags().Bool("dry-run", false, "Show what would be cleaned up without actually doing it")
	cmd.Flags().Bool("force", false, "Skip interactive confirmation")
	cmd.Flags().Bool("merged-only", false, "Only clean up worktrees with merged branches")
	cmd.Flags().Duration("older-than", 0, "Only clean up worktrees older than specified duration (e.g., 30d, 1w)")

	return cmd
}
