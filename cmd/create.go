package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCreateCmd creates the create command
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [branch-name]",
		Short: "Create a new Git worktree",
		Long: `Create a new Git worktree for the specified branch.

If no branch name is provided, an interactive selection will be presented.
The worktree will be created in the configured workspace directory.

Examples:
  twiggit create feature/new-auth
  twiggit create --from=main hotfix/critical-bug
  twiggit create  # Interactive mode`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement create logic
			fmt.Println("Create command - not yet implemented")
			return nil
		},
	}

	// Add flags specific to create
	cmd.Flags().String("from", "", "Create worktree from specific branch or commit")
	cmd.Flags().String("template", "", "Use project template for worktree setup")
	cmd.Flags().Bool("open", false, "Open worktree in default IDE after creation")

	return cmd
}
