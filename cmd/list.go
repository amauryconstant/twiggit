package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available worktrees",
		Long: `List all Git worktrees in the current project or workspace.

Shows detailed information about each worktree including:
- Path and branch name
- Status (clean/dirty)
- Last commit information
- Creation date

Examples:
  twiggit list
  twiggit list --all-projects
  twiggit list --format=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement list logic
			fmt.Println("List command - not yet implemented")
			return nil
		},
	}

	// Add flags specific to list
	cmd.Flags().Bool("all-projects", false, "List worktrees from all projects in workspace")
	cmd.Flags().String("format", "table", "Output format (table, json, yaml, simple)")
	cmd.Flags().String("sort", "name", "Sort order (name, date, branch, status)")

	return cmd
}
