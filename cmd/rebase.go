package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// RebaseOptions captures every input to runRebase. The runF test seam
// lets unit tests inject a stub for the inner walk without rebuilding
// the cobra tree.
type RebaseOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	IsAll      bool
	IsFetch    bool
	IsContinue bool
	IsAbort    bool
	IsForce    bool
	SetBase    string
	Target     string
	RebaseFunc func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error)
}

// NewCmdRebase creates the rebase command. runF is the optional test
// override; pass nil to install the default runRebase body.
func NewCmdRebase(f *cmdutil.Factory, runF func(*RebaseOptions) error) *cobra.Command {
	opts := &RebaseOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "rebase [project/branch]",
		Short: "Rebase a worktree onto its tracked base",
		Long: `Rebase the current (or named) worktree onto its tracked base branch.

By default, rebase is offline and operates on the current worktree.
Use flags to control scope and behaviour:

  --all, -a           Rebase every worktree in the current project
  --fetch, -f         Fetch the tracked base from origin before rebasing
  --continue, -c      Resume a paused rebase (after conflict resolution)
  --abort, -A         Abort a paused rebase
  --set-base <branch> Persist the tracked base without rebasing
  --force, -F         Bypass the dirty-worktree refusal check (use
                      with care — git will still fail if the rebase
                      itself conflicts)

Examples:
  twiggit rebase                      Rebase the current worktree
  twiggit rebase myproject/feature    Rebase a specific worktree
  twiggit rebase --all                Rebase every worktree in the project
  twiggit rebase --fetch              Fetch the base, then rebase
  twiggit rebase --continue feature   Resume a paused rebase
  twiggit rebase --abort feature      Abort a paused rebase
  twiggit rebase --set-base develop   Persist a new tracked base`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(rebaseArgsValidator),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			if len(args) > 0 {
				opts.Target = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runRebase(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.IsAll, "all", "a", false, "Rebase every worktree in the project")
	cmd.Flags().BoolVarP(&opts.IsFetch, "fetch", "f", false, "Fetch the tracked base before rebasing")
	cmd.Flags().BoolVarP(&opts.IsContinue, "continue", "c", false, "Resume a paused rebase")
	cmd.Flags().BoolVarP(&opts.IsAbort, "abort", "A", false, "Abort a paused rebase")
	cmd.Flags().StringVar(&opts.SetBase, "set-base", "", "Persist a new tracked base without rebasing")
	cmd.Flags().BoolVarP(&opts.IsForce, "force", "F", false, "Bypass the dirty-worktree refusal check")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// rebaseArgsValidator returns a *core.UsageError when --all and a
// positional argument are both supplied. Cobra's MaximumNArgs accepts
// the combination; the spec pins UsageError so the call site can
// recover from a typo.
func rebaseArgsValidator(cmd *cobra.Command, args []string) error {
	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return core.UsageWrap(err) //nolint:wrapcheck // validator returns UsageError directly
	}
	isAll, _ := cmd.Flags().GetBool("all")
	if isAll && len(args) > 0 {
		return core.NewUsageError("rebase: --all and a positional argument are mutually exclusive", nil)
	}
	return nil
}

// runRebase composes the rebase walk. The flow: detect context,
// dispatch the short-circuit verbs (--continue / --abort / --set-base)
// or fall through to the full walk.
func runRebase(opts *RebaseOptions) error {
	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	// Short-circuit verbs.
	if opts.SetBase != "" {
		return runRebaseSetBase(opts, gitClient, cfg, currentCtx)
	}
	if opts.IsContinue {
		return runRebaseContinue(opts, gitClient, cfg, currentCtx)
	}
	if opts.IsAbort {
		return runRebaseAbort(opts, gitClient, cfg, currentCtx)
	}

	// Full walk: either fan out (--all) or single-target.
	var result *core.RebaseResult
	if opts.RebaseFunc != nil {
		result, err = opts.RebaseFunc(opts.Ctx, opts)
	} else {
		result, err = runRebaseWalk(opts, gitClient, cfg, currentCtx)
	}
	if err != nil {
		return err
	}

	emitRebaseOutput(opts, result, currentCtx)
	return nil
}

// runRebaseSetBase persists a new tracked base. The target worktree
// must resolve from either the positional argument or the current
// context. Outside git with no positional is a UsageError.
func runRebaseSetBase(opts *RebaseOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) error {
	target, err := resolveRebaseTarget(opts, gitClient, cfg, currentCtx)
	if err != nil {
		return err
	}

	base := opts.SetBase
	if base == "" {
		return core.NewUsageError("rebase --set-base requires a non-empty base branch", nil)
	}

	if err := gitClient.SetTrackedBase(opts.Ctx, target.Path, base); err != nil {
		return &core.OperationError{
			Op:      "rebase.set-base",
			Entity:  target.Path,
			Message: "failed to set tracked base",
			Cause:   err,
		}
	}

	_, _ = fmt.Fprintf(writeOrIgnore(opts.IO.Out), "%s\n", target.Path)
	return nil
}

// runRebaseContinue resumes a paused rebase. Returns the post-continue
// outcome so the user sees whether the resume completed or hit another
// conflict.
func runRebaseContinue(opts *RebaseOptions, gitClient *git.Client, _ *core.Config, currentCtx *core.Context) error {
	target, err := resolveRebaseTarget(opts, gitClient, nil, currentCtx)
	if err != nil {
		return err
	}

	outcome, err := gitClient.Continue(opts.Ctx, target.Path)
	if err != nil {
		return &core.OperationError{
			Op:      "rebase.continue",
			Entity:  target.Path,
			Message: "git rebase --continue failed",
			Cause:   err,
		}
	}

	_, _ = fmt.Fprintf(writeOrIgnore(opts.IO.Out), "continue: %s (%s)\n", outcome, target.Path)
	if outcome == core.RebaseOutcomeConflicted {
		return &core.OperationError{
			Op:      "rebase.continue",
			Entity:  target.Path,
			Message: "rebase hit a conflict during continue",
			Cause:   core.ErrRebaseConflict,
		}
	}
	return nil
}

// runRebaseAbort aborts a paused rebase. Returns nil even when no
// rebase is in progress (idempotent), surfacing only on real git errors.
func runRebaseAbort(opts *RebaseOptions, gitClient *git.Client, _ *core.Config, currentCtx *core.Context) error {
	target, err := resolveRebaseTarget(opts, gitClient, nil, currentCtx)
	if err != nil {
		return err
	}

	if err := gitClient.Abort(opts.Ctx, target.Path); err != nil {
		return &core.OperationError{
			Op:      "rebase.abort",
			Entity:  target.Path,
			Message: "git rebase --abort failed",
			Cause:   err,
		}
	}

	_, _ = fmt.Fprintf(writeOrIgnore(opts.IO.Out), "aborted: %s\n", target.Path)
	return nil
}

// resolveRebaseTarget picks the single worktree the user is
// operating on. The priority chain: explicit positional
// [project/branch] > current worktree (if any) > current project
// (if any). Outside git with no positional is a UsageError.
func resolveRebaseTarget(opts *RebaseOptions, _ *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.Worktree, error) {
	if opts.Target != "" {
		parts := strings.SplitN(opts.Target, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, core.NewUsageError("rebase: target must be in <project>/<branch> format", nil)
		}
		projectPath := filepath.Join(cfg.ProjectsDirectory, parts[0])
		worktreePath := filepath.Join(cfg.WorktreesDirectory, parts[0], parts[1])
		_ = projectPath
		return &core.Worktree{Path: worktreePath, Branch: parts[1]}, nil
	}

	if currentCtx == nil {
		return nil, core.NewUsageError("rebase: not inside a git context; provide a [project/branch] argument", nil)
	}
	if currentCtx.Type == core.ContextWorktree {
		return &core.Worktree{Path: currentCtx.Path, Branch: currentCtx.BranchName}, nil
	}
	if currentCtx.Type == core.ContextProject {
		mainRepo := git.FindMainRepoByTraversal(currentCtx.Path)
		if mainRepo == "" {
			return nil, core.NewUsageError("rebase: cannot locate main repository from project context", nil)
		}
		return &core.Worktree{Path: mainRepo, Branch: "main"}, nil
	}
	return nil, core.NewUsageError("rebase: outside any git context; provide a [project/branch] argument", nil)
}

// dummy usage to silence unused-import linter during early build phases.
var (
	_ = errors.Is
	_ = time.Second
	_ = filepath.Join
	_ = os.Getenv
)
