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
	var force, mergedOnly, changeDir bool

	cmd := &cobra.Command{
		Use:   "delete <project>/<branch> | <worktree-path>",
		Short: "Delete a worktree",
		Long: `Delete a worktree with safety checks.
By default, prevents deletion of worktrees with uncommitted changes.
Use --force to override safety checks.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeDelete(config, cmd, args[0], force, mergedOnly, changeDir)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVar(&mergedOnly, "merged-only", false, "Only delete if branch is merged")
	cmd.Flags().BoolVarP(&changeDir, "change-dir", "C", false, "Change directory after deletion (outputs path to stdout)")

	return cmd
}

// executeDelete executes the delete command with the given configuration
func executeDelete(config *CommandConfig, cmd *cobra.Command, target string, force, mergedOnly, changeDir bool) error {
	ctx := context.Background()

	currentCtx, worktreePath, err := resolveWorktreeTarget(config, target)
	if err != nil {
		return err
	}

	err = validateWorktreeStatus(ctx, config, worktreePath, force, changeDir, currentCtx)
	if err != nil {
		return err
	}

	err = validateMergedOnly(ctx, config, worktreePath, mergedOnly)
	if err != nil {
		return err
	}

	return deleteWorktree(ctx, config, cmd, worktreePath, force, changeDir, currentCtx)
}

func resolveWorktreeTarget(config *CommandConfig, target string) (*domain.Context, string, error) {
	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return nil, "", fmt.Errorf("context detection failed: %w", err)
	}

	resolution, err := config.Services.ContextService.ResolveIdentifier(target)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve target %s: %w", target, err)
	}

	if resolution.Type == domain.PathTypeInvalid {
		return nil, "", fmt.Errorf("invalid target format: %s", resolution.Explanation)
	}

	return currentCtx, resolution.ResolvedPath, nil
}

func validateWorktreeStatus(ctx context.Context, config *CommandConfig, worktreePath string, force, changeDir bool, currentCtx *domain.Context) error {
	if force {
		return nil
	}

	status, err := config.Services.WorktreeService.GetWorktreeStatus(ctx, worktreePath)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "worktree not found") ||
			strings.Contains(errStr, "invalid git repository") ||
			strings.Contains(errStr, "repository does not exist") ||
			strings.Contains(errStr, "no such file or directory") {
			fmt.Printf("Deleted worktree: %s (already removed)\n", worktreePath)
			if changeDir {
				fmt.Println(currentCtx.Path)
			}
			return fmt.Errorf("worktree not found: %s", worktreePath)
		}
		return fmt.Errorf("failed to check worktree status: %w", err)
	}

	if !status.IsClean {
		return errors.New("worktree has uncommitted changes (use --force to override)")
	}

	return nil
}

func validateMergedOnly(ctx context.Context, config *CommandConfig, worktreePath string, mergedOnly bool) error {
	if !mergedOnly {
		return nil
	}

	worktrees, err := config.Services.GitClient.ListWorktrees(ctx, worktreePath)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	var branchName string
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			branchName = wt.Branch
			break
		}
	}

	if branchName == "" {
		return fmt.Errorf("could not determine branch name for worktree: %s", worktreePath)
	}

	isMerged, err := config.Services.GitClient.IsBranchMerged(ctx, worktreePath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch '%s' is merged: %w", branchName, err)
	}

	if !isMerged {
		return fmt.Errorf("branch '%s' is not merged (cannot delete with --merged-only)", branchName)
	}

	return nil
}

func deleteWorktree(ctx context.Context, config *CommandConfig, cmd *cobra.Command, worktreePath string, force, changeDir bool, currentCtx *domain.Context) error {
	logv(cmd, 1, "Deleting worktree at %s", worktreePath)

	logv(cmd, 2, "  project: %s", currentCtx.ProjectName)

	worktrees, _ := config.Services.GitClient.ListWorktrees(ctx, worktreePath)
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			logv(cmd, 2, "  branch: %s", wt.Branch)
			break
		}
	}
	logv(cmd, 2, "  force: %t", force)

	req := &domain.DeleteWorktreeRequest{
		WorktreePath: worktreePath,
		Force:        force,
		Context:      currentCtx,
	}

	err := config.Services.WorktreeService.DeleteWorktree(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	fmt.Printf("Deleted worktree: %s\n", worktreePath)

	if changeDir {
		fmt.Println(currentCtx.Path)
	}

	return nil
}
