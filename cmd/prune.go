package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

// NewPruneCommand creates a new prune command for deleting merged worktrees.
func NewPruneCommand(config *CommandConfig) *cobra.Command {
	var force, deleteBranches, allProjects, dryRun bool

	cmd := &cobra.Command{
		Use:   "prune [project/branch]",
		Short: "Prune merged worktrees",
		Long: `Delete merged worktrees for post-merge cleanup.

By default, prunes merged worktrees in the current project context.
Use flags to customize behavior:

  --dry-run          Preview what would be deleted without making changes
  --force            Bypass uncommitted changes safety checks
  --delete-branches  Also delete the corresponding git branches
  --all              Prune across all projects (requires confirmation)

Examples:
  twiggit prune                       Prune merged worktrees in current project
  twiggit prune --dry-run             Preview what would be deleted
  twiggit prune --all                 Prune across all projects
  twiggit prune myproject/feature     Prune a specific worktree
  twiggit prune --delete-branches     Prune and delete branches`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			var specificWorktree string
			if len(args) > 0 {
				specificWorktree = args[0]
			}
			return executePrune(c, config, force, deleteBranches, allProjects, dryRun, specificWorktree)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVar(&deleteBranches, "delete-branches", false, "Delete branches after worktree removal")
	cmd.Flags().BoolVarP(&allProjects, "all", "a", false, "Prune across all projects")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview only, no actual deletion")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(config, infrastructure.WithExistingOnly()),
	)

	return cmd
}

func executePrune(c *cobra.Command, config *CommandConfig, force, deleteBranches, allProjects, dryRun bool, specificWorktree string) error {
	ctx := context.Background()

	currentCtx, err := config.Services.ContextService.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	if allProjects && !force && !dryRun {
		confirmed, err := confirmBulkPrune(c)
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(c.ErrOrStderr(), "Prune cancelled.")
			return nil
		}
	}

	req := &domain.PruneWorktreesRequest{
		Context:          currentCtx,
		Force:            force,
		DeleteBranches:   deleteBranches,
		DryRun:           dryRun,
		AllProjects:      allProjects,
		SpecificWorktree: specificWorktree,
	}

	result, err := config.Services.WorktreeService.PruneMergedWorktrees(ctx, req)
	if err != nil {
		return fmt.Errorf("prune failed: %w", err)
	}

	outputPruneResults(c, result, dryRun)

	if result.NavigationPath != "" {
		_, _ = fmt.Fprintln(c.OutOrStdout(), result.NavigationPath)
	}

	return nil
}

func confirmBulkPrune(c *cobra.Command) (bool, error) {
	_, _ = fmt.Fprint(c.ErrOrStderr(), "This will prune merged worktrees across all projects. Continue? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func outputPruneResults(c *cobra.Command, result *domain.PruneWorktreesResult, dryRun bool) {
	errOut := c.OutOrStderr()

	if dryRun {
		_, _ = fmt.Fprintln(errOut, "Dry run - no changes made:")
	}

	if len(result.DeletedWorktrees) > 0 {
		if dryRun {
			_, _ = fmt.Fprintf(errOut, "\nWould delete %d worktree(s):\n", len(result.DeletedWorktrees))
		} else {
			_, _ = fmt.Fprintf(errOut, "\nDeleted %d worktree(s):\n", len(result.DeletedWorktrees))
		}
		for _, wt := range result.DeletedWorktrees {
			_, _ = fmt.Fprintf(errOut, "  %s (%s/%s)\n", wt.WorktreePath, wt.ProjectName, wt.BranchName)
			if wt.BranchDeleted {
				_, _ = fmt.Fprintf(errOut, "    branch deleted: %s\n", wt.BranchName)
			}
			if wt.Error != nil {
				_, _ = fmt.Fprintf(errOut, "    warning: %v\n", wt.Error)
			}
		}
	}

	if len(result.UnmergedSkipped) > 0 {
		_, _ = fmt.Fprintf(errOut, "\nSkipped %d unmerged worktree(s):\n", len(result.UnmergedSkipped))
		for _, wt := range result.UnmergedSkipped {
			_, _ = fmt.Fprintf(errOut, "  %s/%s\n", wt.ProjectName, wt.BranchName)
		}
	}

	if len(result.ProtectedSkipped) > 0 {
		_, _ = fmt.Fprintf(errOut, "\nSkipped %d protected branch(es):\n", len(result.ProtectedSkipped))
		for _, wt := range result.ProtectedSkipped {
			_, _ = fmt.Fprintf(errOut, "  %s/%s\n", wt.ProjectName, wt.BranchName)
		}
	}

	if len(result.SkippedWorktrees) > 0 {
		_, _ = fmt.Fprintf(errOut, "\nSkipped %d worktree(s):\n", len(result.SkippedWorktrees))
		for _, wt := range result.SkippedWorktrees {
			_, _ = fmt.Fprintf(errOut, "  %s/%s: %s\n", wt.ProjectName, wt.BranchName, wt.SkipReason)
		}
	}

	_, _ = fmt.Fprintf(errOut, "\nSummary: %d deleted, %d skipped", result.TotalDeleted, result.TotalSkipped)
	if result.TotalBranchesDeleted > 0 {
		_, _ = fmt.Fprintf(errOut, ", %d branches deleted", result.TotalBranchesDeleted)
	}
	_, _ = fmt.Fprintln(errOut)
}
