package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/core"
	"twiggit/internal/git"
)

// NewDeleteCommand creates a new delete command
func NewDeleteCommand(f *CommandConfig) *cobra.Command {
	var force, mergedOnly, changeDir bool

	cmd := &cobra.Command{
		Use:     "delete <project>/<branch> | <worktree-path>",
		Aliases: []string{"rm"},
		Short:   "Delete a worktree",
		Long: `Delete a worktree with safety checks.
By default, prevents deletion of worktrees with uncommitted changes.

Examples:
  twiggit delete feature/my-feature         Delete specific worktree
  twiggit rm feature/my-feature            Same as delete (alias)
  twiggit delete feature --force           Delete even with uncommitted changes
  twiggit delete feature --merged-only      Only delete if branch is merged
  twiggit delete feature -C                 Delete and output navigation path`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.ExactArgs(1)),
		RunE: func(c *cobra.Command, args []string) error {
			return executeDelete(c, f, args[0], force, mergedOnly, changeDir)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVarP(&mergedOnly, "merged-only", "m", false, "Only delete if branch is merged")
	cmd.Flags().BoolVarP(&changeDir, "cd", "C", false, "Change directory after deletion (outputs path to stdout)")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// executeDelete implements the orchestration previously in
// navigationService + worktreeService.DeleteWorktree + worktreeService.GetWorktreeStatus.
// It composes the context detector, resolver, and Client directly.
func executeDelete(c *cobra.Command, f *CommandConfig, target string, force, mergedOnly, changeDir bool) error {
	ctx := context.Background()

	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	gitClient, err := f.GitClient()
	if err != nil {
		return fmt.Errorf("git client init failed: %w", err)
	}

	detector, err := git.NewContextDetector(cfg)
	if err != nil {
		return fmt.Errorf("context detector init failed: %w", err)
	}

	resolver := git.NewContextResolver(cfg, gitClient, gitClient)

	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	currentCtx, err := detector.DetectContext(wd)
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	resolution, err := resolver.ResolveIdentifier(currentCtx, target)
	if err != nil {
		return fmt.Errorf("failed to resolve target %s: %w", target, err)
	}

	if resolution.Type == core.PathTypeInvalid {
		return fmt.Errorf("invalid target format: %s", resolution.Explanation)
	}

	if resolution.Type == core.PathTypeWorktree && resolution.ResolvedPath == "" {
		return core.NewOpValidationError("executeDelete", "ResolvedPath", "", "resolved path cannot be empty")
	}

	worktreePath := resolution.ResolvedPath

	// Status safety check
	if !force {
		status, err := getWorktreeStatus(ctx, gitClient, worktreePath)
		if err != nil {
			if errors.Is(err, core.ErrWorktreeNotFound) {
				if changeDir {
					nav := getDeleteNavigationTarget(currentCtx, worktreePath)
					if nav != "" {
						_, _ = fmt.Fprintln(c.OutOrStdout(), nav)
					}
				} else {
					_, _ = fmt.Fprintf(c.OutOrStdout(), "Deleted worktree: %s (already removed)\n", worktreePath)
				}
				return &core.OperationError{
					Op:      "delete.worktree",
					Entity:  worktreePath,
					Message: "worktree not found",
				}
			}
			return fmt.Errorf("failed to check worktree status: %w", err)
		}
		if !status.IsClean {
			return errors.New("worktree has uncommitted changes (use --force to override)")
		}
	}

	// merged-only check
	if mergedOnly {
		wtInfo, err := clientGetWorktreeByPath(ctx, gitClient, currentCtx.Path, worktreePath)
		if err != nil {
			return fmt.Errorf("failed to get worktree info: %w", err)
		}
		merged, err := gitClient.IsBranchMerged(ctx, worktreePath, wtInfo.Branch)
		if err != nil {
			return fmt.Errorf("failed to check if branch '%s' is merged: %w", wtInfo.Branch, err)
		}
		if !merged {
			return fmt.Errorf("branch '%s' is not merged (cannot delete with --merged-only)", wtInfo.Branch)
		}
	}

	logv(c, 1, "Deleting worktree at %s", worktreePath)
	logv(c, 2, "  project: %s", currentCtx.ProjectName)
	logv(c, 2, "  force: %t", force)

	if err := gitClient.DeleteWorktree(ctx, currentCtx.Path, worktreePath, force); err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	if changeDir {
		nav := getDeleteNavigationTarget(currentCtx, worktreePath)
		if nav != "" {
			_, _ = fmt.Fprintln(c.OutOrStdout(), nav)
		}
	} else if !isQuiet(c) {
		_, _ = fmt.Fprintf(c.OutOrStdout(), "Deleted worktree: %s\n", worktreePath)
	}

	return nil
}

// getWorktreeStatus mirrors worktreeService.GetWorktreeStatus.
func getWorktreeStatus(ctx context.Context, client *git.Client, worktreePath string) (*core.WorktreeStatus, error) {
	if worktreePath == "" {
		return nil, core.NewOpValidationError("GetWorktreeStatus", "worktreePath", "", "worktree path cannot be empty")
	}
	if err := client.ValidateRepository(worktreePath); err != nil {
		return nil, &core.OperationError{
			Op:      "status.worktree",
			Entity:  worktreePath,
			Message: "invalid git repository",
			Cause:   err,
		}
	}

	repoStatus, err := client.GetRepositoryStatus(ctx, worktreePath)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "status.worktree",
			Entity:  worktreePath,
			Message: "failed to get repository status",
			Cause:   err,
		}
	}

	branchStatus := "up-to-date"
	if repoStatus.Ahead > 0 && repoStatus.Behind > 0 {
		branchStatus = "diverged"
	} else if repoStatus.Ahead > 0 {
		branchStatus = "ahead"
	} else if repoStatus.Behind > 0 {
		branchStatus = "behind"
	}

	return &core.WorktreeStatus{
		RepositoryStatus:      &repoStatus,
		LastChecked:           now(),
		IsClean:               repoStatus.IsClean,
		HasUncommittedChanges: !repoStatus.IsClean,
		BranchStatus:          branchStatus,
	}, nil
}

// now returns the current time. Pulled into a helper so future slices
// can stub it without rewriting the call sites.
func now() time.Time {
	return time.Now()
}

// clientGetWorktreeByPath returns the worktree info for worktreePath
// under projectPath, or a NotFound sentinel.
func clientGetWorktreeByPath(ctx context.Context, client *git.Client, projectPath, worktreePath string) (*core.WorktreeInfo, error) {
	worktrees, err := client.ListWorktrees(ctx, projectPath)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "delete.worktree",
			Entity:  worktreePath,
			Message: "failed to list worktrees",
			Cause:   err,
		}
	}
	for i := range worktrees {
		if worktrees[i].Path == worktreePath {
			return &worktrees[i], nil
		}
	}
	return nil, &core.OperationError{
		Op:      "delete.worktree",
		Entity:  worktreePath,
		Message: "worktree not found",
	}
}

// getDeleteNavigationTarget returns the project path for navigation
// when the deleted worktree was a ContextWorktree.
func getDeleteNavigationTarget(currentCtx *core.Context, _ string) string {
	if currentCtx.Type == core.ContextWorktree {
		return currentCtx.Path
	}
	return ""
}
