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
	var output string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List worktrees",
		Long: `List worktrees for the current project or all projects.
By default, lists worktrees for the detected project context.
Use --all to list worktrees from all projects.`,
		Args: cobra.NoArgs, // Reject any positional arguments
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate output format
			if output != "" && output != "text" && output != "json" {
				return fmt.Errorf("invalid output format '%s': must be 'text' or 'json'", output)
			}
			return executeList(cmd, config, all, output)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "List worktrees from all projects")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text or json)")

	return cmd
}

// executeList executes the list command with the given configuration
func executeList(cmd *cobra.Command, config *CommandConfig, all bool, output string) error {
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

	logv(cmd, 1, "Listing worktrees")
	if all {
		logv(cmd, 2, "  repository: all projects")
	} else if currentCtx.ProjectName != "" {
		logv(cmd, 2, "  project: %s", currentCtx.ProjectName)
	}
	logv(cmd, 2, "  including main worktree: %t", req.IncludeMain)

	// List worktrees
	worktrees, err := config.Services.WorktreeService.ListWorktrees(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Select formatter based on output flag
	var formatter OutputFormatter
	if output == "json" {
		formatter = &JSONFormatter{}
	} else {
		formatter = &TextFormatter{}
	}

	// Display results
	if err := displayWorktrees(cmd.OutOrStdout(), worktrees, formatter); err != nil {
		return err
	}

	return nil
}

// displayWorktrees displays the worktrees using the specified formatter
func displayWorktrees(out io.Writer, worktrees []*domain.WorktreeInfo, formatter OutputFormatter) error {
	formatted := formatter.FormatWorktrees(worktrees)
	if _, err := fmt.Fprint(out, formatted); err != nil {
		return fmt.Errorf("failed to display worktrees: %w", err)
	}
	return nil
}
