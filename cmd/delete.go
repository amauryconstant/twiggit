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

Flags:
  -f, --force      Force deletion even with uncommitted changes
  --merged-only    Only delete if branch is merged
  -C, --cd         Output navigation target path to stdout (for shell wrapper)`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return executeDelete(c, config, args[0], force, mergedOnly, changeDir)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVar(&mergedOnly, "merged-only", false, "Only delete if branch is merged")
	cmd.Flags().BoolVarP(&changeDir, "cd", "C", false, "Change directory after deletion (outputs path to stdout)")

	return cmd
}

func executeDelete(c *cobra.Command, config *CommandConfig, target string, force, mergedOnly, changeDir bool) error {
	ctx := context.Background()

	currentCtx, worktreePath, err := resolveWorktreeTarget(config, target)
	if err != nil {
		return err
	}

	err = validateWorktreeStatus(ctx, config, c, worktreePath, force, changeDir, currentCtx)
	if err != nil {
		return err
	}

	err = validateMergedOnly(ctx, config, worktreePath, mergedOnly)
	if err != nil {
		return err
	}

	return deleteWorktree(ctx, config, c, worktreePath, force, changeDir, currentCtx)
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

func validateWorktreeStatus(ctx context.Context, config *CommandConfig, c *cobra.Command, worktreePath string, force, changeDir bool, currentCtx *domain.Context) error {
	if force {
		return nil
	}

	status, err := config.Services.WorktreeService.GetWorktreeStatus(ctx, worktreePath)
	if err != nil {
		var worktreeErr *domain.WorktreeServiceError
		var gitRepoErr *domain.GitRepositoryError
		var gitWorktreeErr *domain.GitWorktreeError
		if errors.As(err, &worktreeErr) && strings.Contains(worktreeErr.Error(), "not found") ||
			errors.As(err, &gitRepoErr) && strings.Contains(gitRepoErr.Error(), "does not exist") ||
			errors.As(err, &gitWorktreeErr) && strings.Contains(gitWorktreeErr.Error(), "not found") {
			if changeDir {
				navigationTarget := getDeleteNavigationTarget(ctx, config, worktreePath, currentCtx)
				if navigationTarget != "" {
					_, _ = fmt.Fprintln(c.OutOrStdout(), navigationTarget)
				}
			} else {
				_, _ = fmt.Fprintf(c.OutOrStdout(), "Deleted worktree: %s (already removed)\n", worktreePath)
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

func getDeleteNavigationTarget(ctx context.Context, config *CommandConfig, _ string, currentCtx *domain.Context) string {
	if currentCtx.Type == domain.ContextWorktree {
		req := &domain.ResolvePathRequest{
			Target:  "main",
			Context: currentCtx,
		}
		resolution, err := config.Services.NavigationService.ResolvePath(ctx, req)
		if err == nil && resolution.ResolvedPath != "" {
			return resolution.ResolvedPath
		}
	}
	return ""
}

func deleteWorktree(ctx context.Context, config *CommandConfig, c *cobra.Command, worktreePath string, force, changeDir bool, currentCtx *domain.Context) error {
	logv(c, 1, "Deleting worktree at %s", worktreePath)

	logv(c, 2, "  project: %s", currentCtx.ProjectName)

	worktrees, _ := config.Services.GitClient.ListWorktrees(ctx, worktreePath)
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			logv(c, 2, "  branch: %s", wt.Branch)
			break
		}
	}
	logv(c, 2, "  force: %t", force)

	req := &domain.DeleteWorktreeRequest{
		WorktreePath: worktreePath,
		Force:        force,
		Context:      currentCtx,
	}

	err := config.Services.WorktreeService.DeleteWorktree(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	if changeDir {
		navigationTarget := getDeleteNavigationTarget(ctx, config, worktreePath, currentCtx)
		if navigationTarget != "" {
			_, _ = fmt.Fprintln(c.OutOrStdout(), navigationTarget)
		}
	} else {
		_, _ = fmt.Fprintf(c.OutOrStdout(), "Deleted worktree: %s\n", worktreePath)
	}

	return nil
}
