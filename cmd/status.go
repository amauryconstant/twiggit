package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of all worktrees",
		Long: `Display a comprehensive status overview of all worktrees.

Shows information about:
- Branch name and status (clean/dirty)
- Last commit information
- Uncommitted changes
- Creation date and last activity

Examples:
  twiggit status
  twiggit status --global
  twiggit status --project myproject`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement status logic
			fmt.Println("Status command - not yet implemented")
			return nil
		},
	}

	// Add flags specific to status
	cmd.Flags().Bool("global", false, "Show status across all projects in workspace")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml)")
	cmd.Flags().Bool("dirty-only", false, "Show only worktrees with uncommitted changes")

	return cmd
}
