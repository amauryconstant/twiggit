package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// DeleteOptions captures every input to runDelete.
type DeleteOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	// Per-command fields.
	Target       string
	IsForce      bool
	IsMergedOnly bool
	IsChangeDir  bool
}

// NewCmdDelete creates a new delete command.
//
// runF is the optional override used by tests; pass nil to install
// the default runDelete body.
func NewCmdDelete(f *cmdutil.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			opts.Target = args[0]
			if runF != nil {
				return runF(opts)
			}
			return runDelete(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.IsForce, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVarP(&opts.IsMergedOnly, "merged-only", "m", false, "Only delete if branch is merged")
	cmd.Flags().BoolVarP(&opts.IsChangeDir, "cd", "C", false, "Change directory after deletion (outputs path to stdout)")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// runDelete implements the orchestration previously in
// navigationService + worktreeService.DeleteWorktree + worktreeService.GetWorktreeStatus.
// It composes the context detector, resolver, and Client directly.
//

func runDelete(opts *DeleteOptions) error {
	ctx := opts.Ctx
	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	target, err := resolveDeleteTarget(ctx, currentCtx, cfg, opts)
	if err != nil {
		return err
	}

	if err := validateDeleteSafety(ctx, gitClient, target, opts); err != nil {
		if errors.Is(err, core.ErrWorktreeNotFound) {
			return emitDeleteResult(opts, currentCtx, cfg.ProjectsDirectory, target.WorktreePath)
		}
		return err
	}

	return performDelete(ctx, gitClient, cfg, currentCtx, target, opts)
}

// DeleteTarget is the resolved coordinates of a delete invocation.
// Produced by resolveDeleteTarget and consumed by validateDeleteSafety
// and performDelete so the runDelete entry point stays a flat
// composition rather than a 100+ line procedure.
type DeleteTarget struct {
	ProjectPath  string
	WorktreePath string
	BranchName   string
}

// resolveDeleteTarget parses opts.Target through the resolver chain
// and derives the (projectPath, worktreePath) pair the delete
// command will operate on. The two-level path resolution (resolver
// for worktreePath, projectPath derivation for outside-git callers)
// used to be inlined in the runDelete body; pulling it out makes
// the cmd/ run function a three-call composition.
func resolveDeleteTarget(_ context.Context, currentCtx *core.Context, cfg *core.Config, opts *DeleteOptions) (DeleteTarget, error) {
	resolver := git.NewContextResolver(cfg, opts.gitClient(), opts.gitClient())

	resolution, err := resolver.ResolveIdentifier(currentCtx, opts.Target)
	if err != nil {
		return DeleteTarget{}, fmt.Errorf("failed to resolve target %s: %w", opts.Target, err)
	}
	if resolution.Type == core.PathTypeInvalid {
		return DeleteTarget{}, fmt.Errorf("invalid target format: %s", resolution.Explanation)
	}
	if resolution.Type == core.PathTypeWorktree && resolution.ResolvedPath == "" {
		return DeleteTarget{}, core.NewOpValidationError("resolveDeleteTarget", "ResolvedPath", "", "resolved path cannot be empty")
	}

	worktreePath := resolution.ResolvedPath

	// Resolve the project path (the git repo root) used to issue
	// git worktree commands. From a worktree or project context
	// currentCtx.Path already is the project. From outside-git
	// currentCtx.Path is the CWD which is not a project, so we
	// fall back to ProjectsDirectory/ProjectName based on the
	// resolved identifier; cross-project references like
	// "test/feature-1" still resolve to the right project.
	projectPath := currentCtx.Path
	if projectPath == "" || currentCtx.Type == core.ContextOutsideGit {
		if resolution.ProjectName != "" {
			projectPath = filepath.Join(cfg.ProjectsDirectory, resolution.ProjectName)
		}
	}

	return DeleteTarget{
		ProjectPath:  projectPath,
		WorktreePath: worktreePath,
		BranchName:   resolution.BranchName,
	}, nil
}

// gitClient is a typed accessor that lifts opts.GitClient (which
// returns cmdutil.Client = any) to the concrete *git.Client used
// downstream. Centralised so every helper reads through one seam
// rather than each one calling opts.GitClient() with a type-assert.
func (opts *DeleteOptions) gitClient() *git.Client {
	if opts.GitClient == nil {
		return nil
	}
	c, err := opts.GitClient()
	if err != nil || c == nil {
		return nil
	}
	gc, _ := c.(*git.Client)
	return gc
}

// validateDeleteSafety runs the pre-delete guards: uncommitted-changes
// safety (skipped under --force) and merged-only branch check.
// Returns core.ErrWorktreeNotFound when the worktree was already gone
// at status-check time so runDelete can short-circuit to the
// idempotent navigation/ack emission.
func validateDeleteSafety(ctx context.Context, gitClient *git.Client, target DeleteTarget, opts *DeleteOptions) error {
	if !opts.IsForce {
		status, err := getWorktreeStatus(ctx, gitClient, target.WorktreePath)
		if err != nil {
			if errors.Is(err, core.ErrWorktreeNotFound) {
				return err
			}
			return fmt.Errorf("failed to check worktree status: %w", err)
		}
		if !status.IsClean {
			return core.NewUncommittedChangesError(target.WorktreePath)
		}
	}

	if opts.IsMergedOnly {
		wtInfo, err := clientGetWorktreeByPath(ctx, gitClient, target.ProjectPath, target.WorktreePath)
		if err != nil {
			return fmt.Errorf("failed to get worktree info: %w", err)
		}
		merged, err := gitClient.IsBranchMerged(ctx, target.WorktreePath, wtInfo.Branch)
		if err != nil {
			return fmt.Errorf("failed to check if branch '%s' is merged: %w", wtInfo.Branch, err)
		}
		if !merged {
			return fmt.Errorf("branch '%s' is not merged (cannot delete with --merged-only)", wtInfo.Branch)
		}
	}

	return nil
}

// performDelete issues the actual git worktree delete and renders the
// post-delete output (navigation target, deleted-ack, or quiet ack).
// The git.ErrNotFound race branch falls through to the same
// emission shape as the success branch via emitDeleteResult.
func performDelete(ctx context.Context, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context, target DeleteTarget, opts *DeleteOptions) error {
	verbosef(opts.IO, "Deleting worktree at %s", target.WorktreePath)
	verbosef(opts.IO, "project: %s", currentCtx.ProjectName)
	verbosef(opts.IO, "branch: %s", target.BranchName)
	verbosef(opts.IO, "force: %t", opts.IsForce)

	if err := gitClient.DeleteWorktree(ctx, target.ProjectPath, target.WorktreePath, opts.IsForce); err != nil {
		if errors.Is(err, git.ErrNotFound) {
			// Race: worktree vanished between the status check and
			// the delete invocation. The desired post-state holds,
			// so fall through to the navigation/quiet branch below
			// instead of surfacing a failure.
			return emitDeleteResult(opts, currentCtx, cfg.ProjectsDirectory, target.WorktreePath)
		}
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	return emitDeleteResult(opts, currentCtx, cfg.ProjectsDirectory, target.WorktreePath)
}

// emitDeleteResult renders the post-delete navigation (or quiet ack)
// path. Extracted so the race-condition ErrNotFound branch can fall
// through to the same emission shape as the success branch.
func emitDeleteResult(opts *DeleteOptions, currentCtx *core.Context, projectsDir, worktreePath string) error {
	if opts.IsChangeDir {
		nav := getDeleteNavigationTarget(currentCtx, projectsDir, worktreePath)
		if nav != "" {
			_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), nav)
		}
		return nil
	}
	if !opts.IO.IsQuiet {
		_, _ = fmt.Fprintf(writeOrIgnore(opts.IO.Out), "Deleted worktree: %s\n", worktreePath)
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

	repoStatus, err := client.RepositoryStatus(ctx, worktreePath)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "status.worktree",
			Entity:  worktreePath,
			Message: "failed to get repository status",
			Cause:   err,
		}
	}

	branchStatus := "up-to-date"
	switch {
	case repoStatus.Ahead > 0 && repoStatus.Behind > 0:
		branchStatus = "diverged"
	case repoStatus.Ahead > 0:
		branchStatus = "ahead"
	case repoStatus.Behind > 0:
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
func clientGetWorktreeByPath(ctx context.Context, client *git.Client, projectPath, worktreePath string) (*core.Worktree, error) {
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
// when the deleted worktree was a ContextWorktree. The worktree path
// is unused: navigation targets the project root so the shell wrapper
// can cd the user out of the just-deleted worktree.
func getDeleteNavigationTarget(currentCtx *core.Context, projectsDir, _ string) string {
	if currentCtx.Type == core.ContextWorktree {
		return filepath.Join(projectsDir, currentCtx.ProjectName)
	}
	return ""
}
