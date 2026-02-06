package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"twiggit/internal/domain"
)

// NewListCommand creates a new list command
func NewListCommand(config *CommandConfig) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List worktrees",
		Long: `List worktrees for the current project or all projects.
By default, lists worktrees for the detected project context.
Use --all to list worktrees from all projects.`,
		Args: cobra.NoArgs, // Reject any positional arguments
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeList(cmd, config, all)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "List worktrees from all projects")

	return cmd
}

// executeList executes the list command with the given configuration
func executeList(cmd *cobra.Command, config *CommandConfig, all bool) error {
	ctx := context.Background()

	// Detect current context
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	// Build list request
	req := &domain.ListWorktreesRequest{
		Context:         currentCtx,
		IncludeMain:     false, // By default, don't include main worktree
		ListAllProjects: all,   // Use --all flag to list worktrees from all projects
	}

	// If not listing all, use project name from context
	if !all && currentCtx.ProjectName != "" {
		req.ProjectName = currentCtx.ProjectName
	}

	// List worktrees
	worktrees, err := config.Services.WorktreeService.ListWorktrees(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Display results
	if err := displayWorktrees(cmd.OutOrStdout(), worktrees); err != nil {
		return err
	}

	return nil
}

// displayWorktrees displays the worktrees in a user-friendly format
func displayWorktrees(out io.Writer, worktrees []*domain.WorktreeInfo) error {
	if len(worktrees) == 0 {
		_, err := fmt.Fprintln(out, "No worktrees found")
		if err != nil {
			return fmt.Errorf("failed to display no worktrees message: %w", err)
		}
		return nil
	}

	for _, wt := range worktrees {
		status := ""
		if wt.Modified {
			status = " (modified)"
		}
		if wt.IsDetached {
			status += " (detached)"
		}

		if _, err := fmt.Fprintf(out, "%s -> %s%s\n", wt.Branch, wt.Path, status); err != nil {
			return fmt.Errorf("failed to display worktree: %w", err)
		}
	}
	return nil
}
