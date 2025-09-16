package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSwitchCmd creates the switch command
func NewSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [worktree-path]",
		Short: "Switch to an existing worktree",
		Long: `Switch to an existing Git worktree.

If no path is provided, an interactive selection will be presented.
This command changes your shell working directory to the selected worktree.

Examples:
  twiggit switch /path/to/worktree
  twiggit switch  # Interactive mode`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement switch logic
			fmt.Println("Switch command - not yet implemented")
			return nil
		},
	}

	return cmd
}
