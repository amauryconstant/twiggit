package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
)

// PruneOptions captures every input to runPrune.
type PruneOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (*git.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions

	// Per-command fields.
	Force            bool
	Yes              bool
	DeleteBranches   bool
	AllProjects      bool
	DryRun           bool
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
			if len(args) > 0 {
				opts.SpecificWorktree = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runPrune(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force deletion even with uncommitted changes")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Auto-confirm prompts (keeps safety checks)")
	cmd.Flags().BoolVarP(&opts.DeleteBranches, "delete-branches", "d", false, "Delete branches after worktree removal")
	cmd.Flags().BoolVarP(&opts.AllProjects, "all", "a", false, "Prune across all projects")
	cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "Preview only, no actual deletion")

	carapace.Gen(cmd).PositionalCompletion(
		actionWorktreeTarget(f, git.WithExistingOnly()),
	)

	return cmd
}

// runPrune performs the prune walk. After slice 9 the worktree
// iteration, skip logic, branch deletion, and result aggregation all
// live in cmd/.
func runPrune(opts *PruneOptions) error {
	ctx := opts.Ctx

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	gitClient, err := opts.GitClient()
	if err != nil {
		return fmt.Errorf("git client init failed: %w", err)
	}

	detector, err := git.NewContextDetector(cfg)
	if err != nil {
		return fmt.Errorf("context detector init failed: %w", err)
	}

	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	currentCtx, err := detector.DetectContext(wd)
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	reporter := NewProgressReporter(opts.IO)

	if opts.AllProjects && !opts.Force && !opts.Yes && !opts.DryRun {
		previewReq := &core.PruneWorktreesRequest{
			Context:          currentCtx,
			Force:            opts.Force,
			DeleteBranches:   opts.DeleteBranches,
			AllProjects:      opts.AllProjects,
			DryRun:           true,
			SpecificWorktree: opts.SpecificWorktree,
		}
		reporter.Report("Previewing prune operation...")
		previewResult, err := runPruneWalk(ctx, gitClient, cfg, previewReq, currentCtx)
		if err != nil {
			return fmt.Errorf("prune preview failed: %w", err)
		}
		outputPruneResults(opts.IO, previewResult, true)

		confirmed, err := confirmBulkPrune(opts.IO)
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(opts.IO.ErrOut, "Prune cancelled.")
			return nil
		}
	}

	if opts.AllProjects || opts.SpecificWorktree == "" {
		reporter.Report("Pruning merged worktrees...")
	}

	req := &core.PruneWorktreesRequest{
		Context:          currentCtx,
		Force:            opts.Force,
		DeleteBranches:   opts.DeleteBranches,
		AllProjects:      opts.AllProjects,
		DryRun:           opts.DryRun,
		SpecificWorktree: opts.SpecificWorktree,
	}
	result, err := runPruneWalk(ctx, gitClient, cfg, req, currentCtx)
	if err != nil {
		return fmt.Errorf("prune failed: %w", err)
	}

	outputPruneResults(opts.IO, result, opts.DryRun)

	if opts.AllProjects || opts.SpecificWorktree == "" {
		reporter.Report("Prune complete")
	}

	if result.NavigationPath != "" {
		_, _ = fmt.Fprintln(opts.IO.Out, result.NavigationPath)
	}

	return nil
}

// runPruneWalk performs the prune walk and returns the aggregated
// result. It encapsulates the per-project iteration that previously
// lived in worktreeService.pruneProjectWorktrees + checkWorktreeSkip +
// deleteWorktreeAndBranch.
func runPruneWalk(ctx context.Context, client *git.Client, cfg *core.Config, req *core.PruneWorktreesRequest, currentCtx *core.Context) (*core.PruneWorktreesResult, error) {
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
		pruneProject(ctx, client, cfg, req, project, result, singleTarget, cwd)
	}

	if len(result.DeletedWorktrees) == 1 && req.SpecificWorktree != "" {
		projectName := strings.Split(req.SpecificWorktree, "/")[0]
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
	if req.AllProjects {
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
			if resolved := core.FindMainRepoByTraversal(gitDir.Path); resolved != "" {
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
func pruneProject(ctx context.Context, client *git.Client, cfg *core.Config, req *core.PruneWorktreesRequest, project *core.ProjectInfo, result *core.PruneWorktreesResult, singleTarget, cwd string) {
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
			Deleted:      false,
		}

		if skip := checkWorktreeSkip(ctx, client, cfg, wt, project, cwd, req); skip != nil {
			addSkippedPrune(result, entry, skip)
			continue
		}

		deleteWorktreeAndBranch(ctx, client, project, wt, req, entry, result)
	}
}

type pruneSkipResult struct {
	reason   string
	err      error
	category string
}

// checkWorktreeSkip encodes the prune-skip decision tree.
func checkWorktreeSkip(ctx context.Context, client *git.Client, cfg *core.Config, wt core.WorktreeInfo, project *core.ProjectInfo, cwd string, req *core.PruneWorktreesRequest) *pruneSkipResult {
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

	if !req.Force && !req.DryRun {
		status, err := client.GetRepositoryStatus(ctx, wt.Path)
		if err == nil && !status.IsClean {
			return &pruneSkipResult{reason: "uncommitted changes (use --force to override)", category: "skipped"}
		}
	}

	if req.DryRun {
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

// deleteWorktreeAndBranch performs the actual delete + optional branch removal.
func deleteWorktreeAndBranch(ctx context.Context, client *git.Client, project *core.ProjectInfo, wt core.WorktreeInfo, req *core.PruneWorktreesRequest, entry *core.PruneWorktreeResult, result *core.PruneWorktreesResult) {
	if err := client.DeleteWorktree(ctx, project.GitRepoPath, wt.Path, req.Force); err != nil {
		entry.Error = err
		result.SkippedWorktrees = append(result.SkippedWorktrees, entry)
		result.TotalSkipped++
		return
	}

	entry.Deleted = true
	result.DeletedWorktrees = append(result.DeletedWorktrees, entry)
	result.TotalDeleted++

	if req.DeleteBranches {
		if err := client.PruneWorktrees(ctx, project.GitRepoPath); err != nil {
			slog.Default().Error("prune worktrees failed", "error", err, "repo_path", project.GitRepoPath)
		}
		if err := client.DeleteBranch(ctx, project.GitRepoPath, wt.Branch); err != nil {
			entry.Error = fmt.Errorf("worktree deleted but branch deletion failed: %w", err)
		} else {
			entry.BranchDeleted = true
			result.TotalBranchesDeleted++
		}
	}
}

// confirmBulkPrune prompts on stderr for a y/n confirmation. Reads
// from ios.In so test helpers can supply canned input.
func confirmBulkPrune(ios *iostreams.IOStreams) (bool, error) {
	if ios == nil {
		return false, fmt.Errorf("confirmBulkPrune: nil IOStreams")
	}
	_, _ = fmt.Fprint(ios.ErrOut, "This will prune merged worktrees across all projects. Continue? (y/n): ")
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
	errOut := ios.ErrOut

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
