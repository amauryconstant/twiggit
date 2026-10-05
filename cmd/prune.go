package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// PruneOptions captures every input to runPrune.
type PruneOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	// Per-command fields.
	IsForce          bool
	IsYes            bool
	IsDeleteBranches bool
	IsAllProjects    bool
	IsDryRun         bool
	SpecificWorktree string
}

// NewCmdPrune creates a new prune command.
//
// runF is the optional override used by tests; pass nil to install
// the default runPrune body.
func NewCmdPrune(f *cmdutil.Factory, runF func(*PruneOptions) error) *cobra.Command {
	opts := &PruneOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "prune [project/branch]",
		Short: "Prune merged worktrees",
		Long: `Delete merged worktrees for post-merge cleanup.

By default, prunes merged worktrees in the current project context.
Use flags to customize behavior:

  --dry-run          Preview what would be deleted without making changes
  --force            Bypass uncommitted changes safety checks
  --yes, -y          Auto-confirm prompts (keeps safety checks)
  --delete-branches  Also delete the corresponding git branches
  --all              Prune across all projects (requires confirmation unless --yes or --force)

Examples:
  twiggit prune                       Prune merged worktrees in current project
  twiggit prune --dry-run             Preview what would be deleted
  twiggit prune --all                 Prune across all projects
  twiggit prune --all --yes           Prune across all projects without confirmation
  twiggit prune myproject/feature     Prune a specific worktree
  twiggit prune --delete-branches     Prune and delete branches`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			if len(args) > 0 {
				opts.SpecificWorktree = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runPrune(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.IsForce, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVarP(&opts.IsYes, "yes", "y", false, "Auto-confirm prompts (keeps safety checks)")
	cmd.Flags().BoolVarP(&opts.IsDeleteBranches, "delete-branches", "d", false, "Delete branches after worktree removal")
	cmd.Flags().BoolVarP(&opts.IsAllProjects, "all", "a", false, "Prune across all projects")
	cmd.Flags().BoolVarP(&opts.IsDryRun, "dry-run", "n", false, "Preview only, no actual deletion")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// runPrune composes the prune walk: optional bulk-preview +
// confirmation, the actual walk, and the post-walk output. After
// slice 9 the worktree iteration, skip logic, branch deletion, and
// result aggregation all live in cmd/.
func runPrune(opts *PruneOptions) error {
	ctx := opts.Ctx

	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	var preview *core.PruneWorktreesResult
	if opts.IsAllProjects && !opts.IsForce && !opts.IsYes && !opts.IsDryRun {
		preview, err = buildPrunePreview(ctx, opts, gitClient, cfg, currentCtx)
		if err != nil {
			return err
		}
	}

	result, err := runPruneWithConfirm(ctx, opts, gitClient, cfg, currentCtx, preview)
	if err != nil {
		return err
	}

	emitPruneOutput(opts, result, currentCtx)
	return nil
}

// pruneRequest assembles a *core.PruneWorktreesRequest from opts +
// the current context, with dryRun overriding opts.IsDryRun so the
// preview walk can run a forced IsDryRun=true pass without
// mutating opts.
func pruneRequest(opts *PruneOptions, currentCtx *core.Context, dryRun bool) *core.PruneWorktreesRequest {
	return &core.PruneWorktreesRequest{
		Context:          currentCtx,
		IsForce:          opts.IsForce,
		IsDeleteBranches: opts.IsDeleteBranches,
		IsAllProjects:    opts.IsAllProjects,
		IsDryRun:         dryRun,
		SpecificWorktree: opts.SpecificWorktree,
	}
}

// buildPrunePreview runs the prune walk in dry-run mode and prints
// the preview so the user can decide whether to proceed. Returns
// the preview result so runPrune can feed it into the interactive
// confirm step.
func buildPrunePreview(ctx context.Context, opts *PruneOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context) (*core.PruneWorktreesResult, error) {
	reporter := NewProgressReporter(opts.IO)
	reporter.Report("Previewing prune operation...")

	return runPruneWalk(ctx, gitClient, opts.IO.Logger, cfg, pruneRequest(opts, currentCtx, true), currentCtx)
}

// runPruneWithConfirm gates the real prune behind the interactive
// confirm prompt when a preview was built. With preview == nil the
// confirm is skipped (caller already short-circuited on
// --yes/--force/--dry-run/--specific). Returns nil + nil when the
// user declines; callers must surface "Prune cancelled." on their
// own if they care to print it.
func runPruneWithConfirm(ctx context.Context, opts *PruneOptions, gitClient *git.Client, cfg *core.Config, currentCtx *core.Context, preview *core.PruneWorktreesResult) (*core.PruneWorktreesResult, error) {
	if preview != nil {
		outputPruneResults(opts.IO, preview, true)

		confirmed, err := confirmBulkPrune(opts.IO)
		if err != nil {
			return nil, err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.ErrOut), "Prune cancelled.")
			return nil, nil
		}
	}

	if opts.IsAllProjects || opts.SpecificWorktree == "" {
		reporter := NewProgressReporter(opts.IO)
		reporter.Report("Pruning merged worktrees...")
	}

	result, err := runPruneWalk(ctx, gitClient, opts.IO.Logger, cfg, pruneRequest(opts, currentCtx, opts.IsDryRun), currentCtx)
	if err != nil {
		return nil, fmt.Errorf("prune failed: %w", err)
	}
	return result, nil
}

// emitPruneOutput prints the aggregated prune result, the
// progress-reporting tail (when applicable), and the navigation
// path for single-target prunes. Tolerates a nil result: a
// cancelled walk returns (nil, nil) from runPruneWithConfirm, and
// the function must not panic on that path.
func emitPruneOutput(opts *PruneOptions, result *core.PruneWorktreesResult, _ *core.Context) {
	if result == nil {
		return
	}

	outputPruneResults(opts.IO, result, opts.IsDryRun)

	if opts.IsAllProjects || opts.SpecificWorktree == "" {
		reporter := NewProgressReporter(opts.IO)
		reporter.Report("Prune complete")
	}

	if result.NavigationPath != "" {
		_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), result.NavigationPath)
	}
}

// runPruneWalk performs the prune walk and returns the aggregated
// result. It encapsulates the per-project iteration that previously
// lived in worktreeService.pruneProjectWorktrees + checkWorktreeSkip +
// deleteWorktreeAndBranch.
func runPruneWalk(ctx context.Context, client *git.Client, logger *slog.Logger, cfg *core.Config, req *core.PruneWorktreesRequest, currentCtx *core.Context) (*core.PruneWorktreesResult, error) {
	result := &core.PruneWorktreesResult{
		DeletedWorktrees: []*core.PruneWorktreeResult{},
		SkippedWorktrees: []*core.PruneWorktreeResult{},
		ProtectedSkipped: []*core.PruneWorktreeResult{},
		UnmergedSkipped:  []*core.PruneWorktreeResult{},
	}

	projects, err := resolvePruneProjects(ctx, client, cfg, req, currentCtx)
	if err != nil {
		return nil, err
	}

	singleTarget := ""
	if req.SpecificWorktree != "" {
		parts := strings.Split(req.SpecificWorktree, "/")
		if len(parts) == 2 {
			singleTarget = parts[1]
		}
	}

	cwd, _ := os.Getwd()

	for _, project := range projects {
		pruneProject(ctx, client, logger, cfg, req, project, result, singleTarget, cwd)
	}

	if len(result.DeletedWorktrees) == 1 && req.SpecificWorktree != "" {
		projectName, _, _ := strings.Cut(req.SpecificWorktree, "/")
		projectPath := filepath.Join(cfg.ProjectsDirectory, projectName)
		if _, statErr := os.Stat(projectPath); statErr == nil {
			result.NavigationPath = projectPath
		}
	}

	return result, nil
}

// resolvePruneProjects returns the list of projects to prune based on
// the request flags. Mirrors the project-resolution branch of
// worktreeService.PruneMergedWorktrees.
func resolvePruneProjects(ctx context.Context, client *git.Client, cfg *core.Config, req *core.PruneWorktreesRequest, currentCtx *core.Context) ([]*core.ProjectInfo, error) {
	if req.IsAllProjects {
		finder := git.NewRepoFinder(client)
		gitDirs, err := finder.FindGitRepositories(cfg.ProjectsDirectory)
		if err != nil {
			return nil, &core.OperationError{
				Op:      "prune.projects",
				Entity:  cfg.ProjectsDirectory,
				Message: "failed to list projects",
				Cause:   err,
			}
		}
		projects := make([]*core.ProjectInfo, len(gitDirs))
		for i, gitDir := range gitDirs {
			mainRepo := gitDir.Path
			if resolved := git.FindMainRepoByTraversal(gitDir.Path); resolved != "" {
				mainRepo = resolved
			}
			projects[i] = &core.ProjectInfo{
				Name:        filepath.Base(mainRepo),
				Path:        gitDir.Path,
				GitRepoPath: mainRepo,
			}
		}
		return projects, nil
	}

	if req.SpecificWorktree != "" {
		parts := strings.Split(req.SpecificWorktree, "/")
		if len(parts) != 2 {
			return nil, core.NewOpValidationError("PruneWorktreesRequest", "SpecificWorktree", req.SpecificWorktree, "must be in format project/branch")
		}
		project, err := discoverProject(ctx, client, cfg, parts[0], currentCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project: %w", err)
		}
		return []*core.ProjectInfo{project}, nil
	}

	projectName := req.ProjectName
	if projectName == "" && currentCtx != nil {
		projectName = currentCtx.ProjectName
	}
	if projectName == "" {
		project, err := discoverProject(ctx, client, cfg, "", currentCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project from context: %w", err)
		}
		return []*core.ProjectInfo{project}, nil
	}
	project, err := discoverProject(ctx, client, cfg, projectName, currentCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}
	return []*core.ProjectInfo{project}, nil
}

// pruneProject iterates a single project's worktrees and applies the
// prune logic.
func pruneProject(ctx context.Context, client *git.Client, logger *slog.Logger, cfg *core.Config, req *core.PruneWorktreesRequest, project *core.ProjectInfo, result *core.PruneWorktreesResult, singleTarget, cwd string) {
	worktrees, err := client.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		return
	}

	for _, wt := range worktrees {
		if wt.Path == project.GitRepoPath {
			continue
		}

		if singleTarget != "" && wt.Branch != singleTarget {
			continue
		}

		entry := &core.PruneWorktreeResult{
			ProjectName:  project.Name,
			WorktreePath: wt.Path,
			BranchName:   wt.Branch,
			WasDeleted:   false,
		}

		if skip := checkWorktreeSkip(ctx, client, cfg, wt, project, cwd, req); skip != nil {
			addSkippedPrune(result, entry, skip)
			continue
		}

		target := pruneTarget{Project: project, Worktree: wt, Request: req}
		if err := deleteWorktree(ctx, client, target); err != nil {
			entry.Error = err
			result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
			result.TotalSkipped++
			continue
		}

		entry.WasDeleted = true
		result.DeletedWorktrees = append(result.DeletedWorktrees, entry)
		result.TotalDeleted++

		if err := deleteBranchIfRequested(ctx, client, logger, target); err != nil {
			entry.Error = err
		} else if req.IsDeleteBranches {
			entry.WasBranchDeleted = true
			result.TotalBranchesDeleted++
		}
	}
}

type pruneSkipResult struct {
	reason   string
	err      error
	category string
}

// checkWorktreeSkip encodes the prune-skip decision tree.
func checkWorktreeSkip(ctx context.Context, client *git.Client, cfg *core.Config, wt core.Worktree, project *core.ProjectInfo, cwd string, req *core.PruneWorktreesRequest) *pruneSkipResult {
	if cwd != "" && (strings.HasPrefix(cwd, wt.Path+string(filepath.Separator)) || cwd == wt.Path) {
		return &pruneSkipResult{reason: "cannot prune current worktree", category: "current"}
	}
	if slices.Contains(cfg.Validation.ProtectedBranches, wt.Branch) {
		return &pruneSkipResult{reason: "protected branch", category: "protected"}
	}

	merged, err := client.IsBranchMerged(ctx, project.GitRepoPath, wt.Branch)
	if err != nil {
		return &pruneSkipResult{reason: "failed to check merge status", err: err, category: "skipped"}
	}
	if !merged {
		return &pruneSkipResult{reason: "branch not merged", category: "unmerged"}
	}

	if !req.IsForce && !req.IsDryRun {
		status, err := client.RepositoryStatus(ctx, wt.Path)
		if err == nil && !status.IsClean {
			return &pruneSkipResult{
				reason:   "uncommitted changes (use --force to override)",
				err:      fmt.Errorf("worktree %s: %w", wt.Path, core.ErrUncommittedChanges),
				category: "skipped",
			}
		}
	}

	if req.IsDryRun {
		return &pruneSkipResult{reason: "dry run", category: "skipped"}
	}

	return nil
}

// addSkippedPrune records the skip in the appropriate bucket.
func addSkippedPrune(result *core.PruneWorktreesResult, entry *core.PruneWorktreeResult, skip *pruneSkipResult) {
	entry.SkipReason = skip.reason
	entry.Error = skip.err

	switch skip.category {
	case "current":
		result.CurrentWorktreeSkipped = append(result.CurrentWorktreeSkipped, entry)
	case "protected":
		result.ProtectedSkipped = append(result.ProtectedSkipped, entry)
	case "unmerged":
		result.UnmergedSkipped = append(result.UnmergedSkipped, entry)
	default:
		result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
	}
	result.TotalSkipped++
}

// pruneTarget bundles the inputs the delete-branch step needs so
// deleteWorktree + deleteBranchIfRequested can share a single value
// without restating (project, worktree, request) at every call.
type pruneTarget struct {
	Project  *core.ProjectInfo
	Worktree core.Worktree
	Request  *core.PruneWorktreesRequest
}

// deleteWorktree removes the worktree at target.Worktree.Path from
// the repo at target.Project.GitRepoPath, honoring the request's
// IsForce flag. Returns the underlying git error so the caller can
// record it on the entry and bucket the worktree into
// SkippedWorktrees.
func deleteWorktree(ctx context.Context, client *git.Client, target pruneTarget) error {
	if err := client.DeleteWorktree(ctx, target.Project.GitRepoPath, target.Worktree.Path, target.Request.IsForce); err != nil {
		return fmt.Errorf("delete worktree %s: %w", target.Worktree.Path, err)
	}
	return nil
}

// deleteBranchIfRequested removes the worktree's git branch when
// the request opts in (--delete-branches). Prunes stale admin
// refs first to avoid "branch in use" failures. Returns the wrapped
// branch-deletion error when the worktree is gone but the branch
// could not be removed; the caller stores it on entry.Error so the
// summary still surfaces the worktree success + branch failure
// separately.
func deleteBranchIfRequested(ctx context.Context, client *git.Client, logger *slog.Logger, target pruneTarget) error {
	if !target.Request.IsDeleteBranches {
		return nil
	}
	if err := client.PruneWorktrees(ctx, target.Project.GitRepoPath); err != nil {
		logger.Error("prune worktrees failed", "error", err, "repo_path", target.Project.GitRepoPath)
	}
	if err := client.DeleteBranch(ctx, target.Project.GitRepoPath, target.Worktree.Branch); err != nil {
		return fmt.Errorf("worktree deleted but branch deletion failed: %w", err)
	}
	return nil
}

// confirmBulkPrune prompts on stderr for a y/n confirmation. Reads
// from ios.In so test helpers can supply canned input.
func confirmBulkPrune(ios *iostreams.IOStreams) (bool, error) {
	if ios == nil {
		return false, errors.New("confirmBulkPrune: nil IOStreams")
	}
	_, _ = fmt.Fprint(writeOrIgnore(ios.ErrOut), "This will prune merged worktrees across all projects. Continue? (y/n): ")
	reader := bufio.NewReader(ios.In)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// outputPruneResults formats the aggregated prune result for the user.
func outputPruneResults(ios *iostreams.IOStreams, result *core.PruneWorktreesResult, dryRun bool) {
	if ios == nil {
		return
	}
	errOut := writeOrIgnore(ios.ErrOut)

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
			if wt.WasBranchDeleted {
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
