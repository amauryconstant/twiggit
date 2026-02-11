package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"twiggit/internal/domain"
)

// NewCDCommand creates a new cd command
func NewCDCommand(config *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cd <project|project/branch>",
		Short: "Change directory to a worktree",
		Long: `Change directory to the specified worktree.
If no target is provided, changes to the default worktree for the current project.
The command outputs the path to be used by shell integration.

Examples:
  twiggit cd                    # Change to default worktree for current project
  twiggit cd myproject          # Change to main worktree of myproject
  twiggit cd myproject/feature  # Change to feature branch worktree
  twiggit cd feature            # Change to feature branch (relative to current project)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			return executeCD(cmd, config, target)
		},
	}

	return cmd
}

// executeCD executes the cd command with the given configuration
func executeCD(cmd *cobra.Command, config *CommandConfig, target string) error {
	ctx := context.Background()

	// Detect current context
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	// If no target specified, use default behavior
	if target == "" {
		switch currentCtx.Type {
		case domain.ContextWorktree:
			// If in a worktree, use the current branch as target
			target = currentCtx.BranchName
		case domain.ContextProject:
			// If in project, use main as default target
			target = "main"
		default:
			return errors.New("no target specified and no default worktree in context")
		}
	}

	// Resolve path
	req := &domain.ResolvePathRequest{
		Target:  target,
		Context: currentCtx,
		Search:  false,
	}

	result, err := config.Services.NavigationService.ResolvePath(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to resolve path for %s: %w", target, err)
	}

	logv(cmd, 1, "Navigating to worktree")
	logv(cmd, 2, "  target: %s", target)
	logv(cmd, 2, "  worktree path: %s", result.ResolvedPath)
	if currentCtx.ProjectName != "" {
		logv(cmd, 2, "  resolved project: %s", currentCtx.ProjectName)
	}

	// Validate that the resolved path exists
	if err := config.Services.NavigationService.ValidatePath(ctx, result.ResolvedPath); err != nil {
		if result.Type == domain.PathTypeWorktree {
			return fmt.Errorf("worktree '%s' not found", target)
		}
		return fmt.Errorf("project '%s' not found", target)
	}

	// Output the resolved path for shell integration
	_, err = fmt.Fprintln(cmd.OutOrStdout(), result.ResolvedPath)
	if err != nil {
		return fmt.Errorf("failed to output path: %w", err)
	}
	return nil
}
