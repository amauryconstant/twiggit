package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// SyncOptions captures every input to runSync.
type SyncOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	IsAll       bool
	Remote      string
	Branch      string
	IsRebase    bool
	IsFetchOnly bool
	Target      string
	SyncFunc    func(ctx context.Context, opts *SyncOptions) (*core.SyncResult, error)
	RebaseOpts  *RebaseOptions
}

// NewCmdSync creates the sync command. runF is the optional test
// override; pass nil to install the default runSync body.
func NewCmdSync(f *cmdutil.Factory, runF func(*SyncOptions) error) *cobra.Command {
	opts := &SyncOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "sync [project]",
		Short: "Sync a project's tracking refs from its remote",
		Long: `Fetch the remote tracking refs for the current (or named)
project's default branch, optionally followed by a rebase of every
worktree onto the refreshed base.

By default, sync fetches the default remote (origin) for the current
project. Use flags to control scope and behaviour:

  --all, -a            Sync every project in the workspace
  --remote <name>      Remote to fetch from (default: config sync.default_remote)
  --branch <name>      Branch to fetch (default: project default branch)
  --rebase, -r         Run the rebase walk after a successful fetch
  --fetch-only, -F     Skip the rebase even if config sync.rebase_after_sync is set

Examples:
  twiggit sync                       Fetch origin's tracking refs
  twiggit sync --all                 Sync every project
  twiggit sync --remote upstream     Fetch from a non-default remote
  twiggit sync --rebase              Fetch and rebase every worktree
  twiggit sync --fetch-only          Skip the rebase walk
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			if len(args) > 0 {
				opts.Target = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runSync(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.IsAll, "all", "a", false, "Sync every project in the workspace")
	cmd.Flags().StringVar(&opts.Remote, "remote", "", "Remote to fetch from")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch to fetch (default: project default branch)")
	cmd.Flags().BoolVarP(&opts.IsRebase, "rebase", "r", false, "Run the rebase walk after fetch")
	cmd.Flags().BoolVarP(&opts.IsFetchOnly, "fetch-only", "F", false, "Skip the rebase walk")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// runSync dispatches: resolve project(s), fetch, then optionally
// invoke the shared rebase walk.
func runSync(opts *SyncOptions) error {
	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	result, err := runSyncWalk(opts, gitClient, cfg, currentCtx)
	if err != nil {
		return err
	}

	emitSyncOutput(opts, result)
	return nil
}

// runSyncWalk fetches the resolved remote tracking ref and (when
// --rebase is set, --fetch-only is not) calls the shared rebase
// walk to rebase every worktree onto the freshly fetched base.
func runSyncWalk(opts *SyncOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.SyncResult, error) {
	if opts.SyncFunc != nil {
		return opts.SyncFunc(opts.Ctx, opts)
	}

	result := &core.SyncResult{
		SyncedBranches: []*core.SyncedBranch{},
	}

	project, err := resolveSyncTargets(opts, gitClient, cfg, currentCtx)
	if err != nil {
		return nil, err
	}

	remote := opts.Remote
	if remote == "" {
		remote = cfg.Sync.DefaultRemote
		if remote == "" {
			remote = "origin"
		}
	}
	branch := opts.Branch
	if branch == "" {
		branch = cfg.DefaultSourceBranch
		if branch == "" {
			branch = "main"
		}
	}

	oldTip, _ := gitClient.RepositoryStatus(opts.Ctx, project.GitRepoPath)
	_ = oldTip

	if err := gitClient.Fetch(opts.Ctx, project.GitRepoPath, remote, branch); err != nil {
		return nil, &core.OperationError{
			Op:      "sync.fetch",
			Entity:  project.GitRepoPath,
			Message: "git fetch failed",
			Cause:   err,
		}
	}

	result.SyncedBranches = append(result.SyncedBranches, &core.SyncedBranch{
		ProjectName: project.Name,
		BranchName:  branch,
		RemoteName:  remote,
	})
	result.TotalSynced++

	if opts.IsFetchOnly {
		return result, nil
	}
	if !opts.IsRebase && !cfg.Sync.RebaseAfterSync {
		return result, nil
	}

	// Delegate the rebase walk to the shared machinery.
	rebaseOpts := &RebaseOptions{
		IO:            opts.IO,
		Config:        opts.Config,
		GitClient:     opts.GitClient,
		Ctx:           opts.Ctx,
		GlobalOptions: opts.GlobalOptions,
		Logger:        opts.Logger,
		IsAll:         true,
	}
	rebaseResult, err := runRebaseWalk(rebaseOpts, gitClient, cfg, currentCtx)
	if err != nil {
		// Sync reports the rebase conflict but does not propagate
		// the error — sync itself succeeded; the rebase is a
		// follow-up the user can drive. Capture any partial
		// result for the caller.
		if rebaseResult != nil {
			result.RebasedBranches = rebaseResult.RebasedWorktrees
			result.TotalRebased = rebaseResult.TotalRebased
			result.TotalConflicts = rebaseResult.TotalConflicts
		}
		return result, nil
	}
	if rebaseResult != nil {
		result.RebasedBranches = rebaseResult.RebasedWorktrees
		result.TotalRebased = rebaseResult.TotalRebased
		result.TotalConflicts = rebaseResult.TotalConflicts
	}
	return result, nil
}

// resolveSyncTargets picks the project to sync. With --all, fan out
// across every project in cfg.ProjectsDirectory; with a positional
// argument, target that project; otherwise, the current context's
// project. Outside git with no argument is a UsageError.
func resolveSyncTargets(opts *SyncOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.ProjectInfo, error) {
	if opts.IsAll {
		dirEntries, err := os.ReadDir(cfg.ProjectsDirectory)
		if err != nil {
			return nil, &core.OperationError{
				Op:      "sync.all",
				Entity:  cfg.ProjectsDirectory,
				Message: "failed to list projects",
				Cause:   err,
			}
		}
		if len(dirEntries) == 0 {
			return nil, core.NewUsageError("sync --all: no projects found", nil)
		}
		if len(dirEntries) > 1 {
			return nil, core.NewUsageError("sync --all: multiple projects; run from inside a project context", nil)
		}
		return discoverProject(opts.Ctx, gitClient, cfg, dirEntries[0].Name(), nil)
	}
	if opts.Target != "" {
		return discoverProject(opts.Ctx, gitClient, cfg, opts.Target, currentCtx)
	}
	if currentCtx == nil {
		return nil, core.NewUsageError("sync: not inside a git context; provide a [project] argument or use --all", nil)
	}
	// discoverProject returns ValidationError when no name and no
	// usable ctx; convert that into a UsageError so the call
	// site surfaces exit-2 rather than exit-1.
	project, err := discoverProject(opts.Ctx, gitClient, cfg, "", currentCtx)
	if err != nil {
		if _, ok := errors.AsType[*core.ValidationError](err); ok {
			return nil, core.NewUsageError("sync: project name required; provide a [project] argument or use --all", nil)
		}
		return nil, err
	}
	return project, nil
}

func emitSyncOutput(opts *SyncOptions, result *core.SyncResult) {
	if result == nil {
		return
	}
	errOut := writeOrIgnore(opts.IO.ErrOut)
	if result.TotalSynced > 0 {
		_, _ = fmt.Fprintf(errOut, "Synced %d branch(es)\n", result.TotalSynced)
	}
	if result.TotalRebased > 0 {
		_, _ = fmt.Fprintf(errOut, "Rebased %d worktree(s)\n", result.TotalRebased)
	}
	if result.TotalConflicts > 0 {
		_, _ = fmt.Fprintf(errOut, "%d conflict(s)\n", result.TotalConflicts)
	}
}
