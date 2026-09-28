package cmd

import (
	"context"
	"errors"
	"fmt"
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

	// Per-command fields.
	Target     string
	Force      bool
	MergedOnly bool
	ChangeDir  bool
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
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Target = args[0]
			if runF != nil {
				return runF(opts)
			}
			return runDelete(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVarP(&opts.MergedOnly, "merged-only", "m", false, "Only delete if branch is merged")
	cmd.Flags().BoolVarP(&opts.ChangeDir, "cd", "C", false, "Change directory after deletion (outputs path to stdout)")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// runDelete implements the orchestration previously in
// navigationService + worktreeService.DeleteWorktree + worktreeService.GetWorktreeStatus.
// It composes the context detector, resolver, and Client directly.
//
//nolint:gocyclo // orchestration function: branching reflects CLI safety checks (status, merged-only, idempotent not-found), not duplicated logic.
func runDelete(opts *DeleteOptions) error {
	ctx := opts.Ctx

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	client, err := opts.GitClient()
	if err != nil {
		return fmt.Errorf("git client init failed: %w", err)
	}
	gitClient, _ := client.(*git.Client)

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

	resolution, err := resolver.ResolveIdentifier(currentCtx, opts.Target)
	if err != nil {
		return fmt.Errorf("failed to resolve target %s: %w", opts.Target, err)
	}

	if resolution.Type == core.PathTypeInvalid {
		return fmt.Errorf("invalid target format: %s", resolution.Explanation)
	}

	if resolution.Type == core.PathTypeWorktree && resolution.ResolvedPath == "" {
		return core.NewOpValidationError("runDelete", "ResolvedPath", "", "resolved path cannot be empty")
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

	// navTarget returns the project path for -C navigation when the
	// deletion leaves the user stranded inside the deleted worktree.
	navTarget := func() string {
		return getDeleteNavigationTarget(currentCtx, cfg.ProjectsDirectory, worktreePath)
	}

	// Status safety check
	if !opts.Force {
		status, err := getWorktreeStatus(ctx, gitClient, worktreePath)
		if err != nil {
			if errors.Is(err, core.ErrWorktreeNotFound) {
				// Idempotent: the desired post-state (worktree gone)
				// is already true. Emit the navigation target only
				// when the caller is inside the deleted worktree;
				// outside-git callers asked for deletion by explicit
				// reference and stay in place, so nothing is printed.
				if opts.ChangeDir {
					if nav := navTarget(); nav != "" {
						_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), nav)
					}
				}
				return nil
			}
			return fmt.Errorf("failed to check worktree status: %w", err)
		}
		if !status.IsClean {
			return errors.New("worktree has uncommitted changes (use --force to override)")
		}
	}

	// merged-only check
	if opts.MergedOnly {
		wtInfo, err := clientGetWorktreeByPath(ctx, gitClient, projectPath, worktreePath)
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

	verbosef(opts.IO, "Deleting worktree at %s", worktreePath)
	verbosef(opts.IO, "project: %s", currentCtx.ProjectName)
	verbosef(opts.IO, "branch: %s", resolution.BranchName)
	verbosef(opts.IO, "force: %t", opts.Force)

	if err := gitClient.DeleteWorktree(ctx, projectPath, worktreePath, opts.Force); err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}

	if opts.ChangeDir {
		nav := navTarget()
		if nav != "" {
			_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), nav)
		}
	} else if !opts.IO.Quiet {
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
// when the deleted worktree was a ContextWorktree. The worktree path
// is unused: navigation targets the project root so the shell wrapper
// can cd the user out of the just-deleted worktree.
func getDeleteNavigationTarget(currentCtx *core.Context, projectsDir, _ string) string {
	if currentCtx.Type == core.ContextWorktree {
		return filepath.Join(projectsDir, currentCtx.ProjectName)
	}
	return ""
}
