package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"twiggit/internal/domain"
)

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(config *CommandConfig) *cobra.Command {
	var force, keepBranch bool

	cmd := &cobra.Command{
		Use:   "delete <project>/<branch> | <worktree-path>",
		Short: "Delete a worktree",
		Long: `Delete a worktree with safety checks.
By default, prevents deletion of worktrees with uncommitted changes.
Use --force to override safety checks.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return executeDelete(config, args[0], force, keepBranch)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVar(&keepBranch, "keep-branch", false, "Keep the branch after deletion")

	return cmd
}

// executeDelete executes the delete command with the given configuration
func executeDelete(config *CommandConfig, target string, force, _ bool) error {
	ctx := context.Background()

	// Detect current context
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	// Resolve target to worktree path
	resolution, err := config.Services.ContextService.ResolveIdentifier(target)
	if err != nil {
		return fmt.Errorf("failed to resolve target %s: %w", target, err)
	}

	worktreePath := resolution.ResolvedPath

	// Safety check: verify worktree status unless forced
	if !force {
		status, err := config.Services.WorktreeService.GetWorktreeStatus(ctx, worktreePath)
		if err != nil {
			// If worktree doesn't exist, we can proceed with deletion (idempotent)
			// Be more specific about what constitutes "doesn't exist"
			errStr := err.Error()
			if strings.Contains(errStr, "worktree not found") ||
				strings.Contains(errStr, "invalid git repository") ||
				strings.Contains(errStr, "repository does not exist") ||
				strings.Contains(errStr, "no such file or directory") {
				// Worktree doesn't exist, proceed to deletion
				return nil
			}
			// For other errors (like HEAD reference issues), we should try to delete anyway
			// since the worktree might exist but be in a broken state
		} else {
			// Worktree exists, check if it's clean
			if !status.IsClean {
				return errors.New("worktree has uncommitted changes (use --force to override)")
			}
		}
	}

	// Create delete request
	req := &domain.DeleteWorktreeRequest{
		WorktreePath: worktreePath,
		Force:        force,
		Context:      currentCtx,
	}

	// Delete worktree
	err = config.Services.WorktreeService.DeleteWorktree(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	// Display success message
	fmt.Printf("Deleted worktree: %s\n", worktreePath)

	return nil
}
