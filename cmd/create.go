package cmd

import (
	"context"
	"fmt"
	"io"
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

// CreateOptions captures every input to runCreate.
// RunE populates it from cobra flags + Factory; runCreate body reads
// a single value struct rather than juggling *cobra.Command, args,
// and Factory references.
type CreateOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	// Per-command fields.
	Spec   string
	Source string
	CdFlag bool
}

// NewCmdCreate creates a new create command.
//
// runF is the optional override used by tests; pass nil to install
// the default runCreate body.
func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "create <project>/<branch> | <branch>",
		Short: "Create a new worktree",
		Long: `Create a new worktree for the specified project and branch.
If only a branch name is provided, the project is inferred from the current context.

Examples:
  twiggit create feature/my-feature              Create from current project
  twiggit create myproject/feature/my-feature    Create for specific project
  twiggit create feature --source develop       Create from specific source branch
  twiggit create feature -C                     Create and output path for shell`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			opts.Spec = args[0]
			// Use config default source branch if available, otherwise fallback to "main"
			if opts.Source == "" {
				if cfg, cerr := opts.Config(); cerr == nil && cfg.DefaultSourceBranch != "" {
					opts.Source = cfg.DefaultSourceBranch
				} else {
					opts.Source = "main"
				}
			}
			if runF != nil {
				return runF(opts)
			}
			return runCreate(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Source, "source", "", "Source branch to create from")
	cmd.Flags().BoolVarP(&opts.CdFlag, "cd", "C", false, "Output worktree path to stdout (for shell wrapper)")

	carapace.Gen(cmd).PositionalCompletion(actionWorktreeTarget(f))
	carapace.Gen(cmd).FlagCompletion(map[string]carapace.Action{
		"source": actionBranches(f),
	})

	return cmd
}

// runCreate composes the git context detector, resolver, and Client
// to materialise a new worktree. After slice 10 the orchestration
// lives in cmd/ directly; the concrete git and core types own the
// per-step work.
func runCreate(opts *CreateOptions) error {
	ctx := opts.Ctx

	branchName := extractBranchNameForValidation(opts.Spec)
	branchValidation := core.ValidateBranchName(branchName)
	if branchValidation.IsError() {
		return branchValidation.Error
	}

	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	projectName, branchName, err := parseProjectBranch(opts.Spec, currentCtx)
	if err != nil {
		return err
	}

	project, err := discoverProject(ctx, gitClient, cfg, projectName, currentCtx)
	if err != nil {
		return fmt.Errorf("failed to discover project %s: %w", projectName, err)
	}

	// Validate source branch exists before creating worktree
	sourceBranchExists, err := gitClient.BranchExists(ctx, project.GitRepoPath, opts.Source)
	if err != nil {
		return core.NewOpValidationError("CreateWorktreeRequest", "source", opts.Source, "failed to check if source branch exists: "+err.Error())
	}
	if !sourceBranchExists {
		return core.NewOpValidationError("CreateWorktreeRequest", "source", opts.Source, fmt.Sprintf("source branch '%s' does not exist", opts.Source))
	}

	worktreePath := calculateWorktreePath(cfg, project.Name, branchName)
	if _, err := os.Stat(worktreePath); err == nil {
		return core.NewConflictError("worktree", branchName, "CreateWorktree", "worktree already exists at "+worktreePath, nil)
	}

	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil { // #nosec G301 -- standard directory perms (rwxr-xr-x)
		return fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	if err := gitClient.CreateWorktree(ctx, project.GitRepoPath, branchName, opts.Source, worktreePath); err != nil {
		return &core.OperationError{
			Op:      "create.worktree",
			Entity:  worktreePath,
			Message: "failed to create worktree",
			Cause:   err,
		}
	}

	result := &core.CreateWorktreeResult{
		Worktree: &core.Worktree{Path: worktreePath, Branch: branchName},
	}

	verbosef(opts.IO, "Creating worktree for %s/%s", project.Name, branchName)
	verbosef(opts.IO, "from branch: %s", opts.Source)
	verbosef(opts.IO, "to path: %s", project.Name+"/"+branchName)

	// Run post-create hooks if the project has a .twiggit.toml config.
	hookResult, err := runPostCreateHooks(ctx, gitClient, cfg, project, branchName, opts.Source, worktreePath)
	if err != nil {
		opts.IO.Logger.Error("hook execution failed",
			"error", err,
			"worktree_path", worktreePath,
			"hook_type", string(core.HookPostCreate),
		)
	}
	result.HookResult = hookResult

	if opts.CdFlag {
		_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), result.Worktree.Path)
	} else if !opts.IO.Quiet {
		if err := displayCreateSuccess(opts.IO.Out, result.Worktree); err != nil {
			return err
		}
	}

	if result.HookResult != nil && !result.HookResult.IsSuccessful {
		displayHookFailures(opts.IO.ErrOut, result.HookResult)
	}

	return nil
}

// runPostCreateHooks constructs a HookRunner and dispatches the
// post-create hook with the worktree context. Errors from the runner
// are logged but do not abort the create — the worktree already
// exists. The runner returns a *HookResult that captures per-command
// success/failure for the caller to surface.
func runPostCreateHooks(ctx context.Context, _ *git.Client, cfg *core.Config, project *core.ProjectInfo, branchName, sourceBranch, worktreePath string) (*core.HookResult, error) {
	timeout := time.Duration(cfg.Shell.HookTimeout) * time.Second
	executor := git.NewCommandExecutor(timeout)
	runner := git.NewHookRunner(executor, cfg.Shell.HookTimeout)

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   worktreePath,
		ProjectName:    project.Name,
		BranchName:     branchName,
		SourceBranch:   sourceBranch,
		MainRepoPath:   project.GitRepoPath,
		ConfigFilePath: filepath.Join(project.GitRepoPath, ".twiggit.toml"),
	}
	hookResult, err := runner.Run(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("run post-create hooks: %w", err)
	}
	return hookResult, nil
}

// discoverProject resolves the project by name with the current
// context, mirroring projectService.DiscoverProject without its
// service wrapper.
func discoverProject(ctx context.Context, client *git.Client, cfg *core.Config, projectName string, currentCtx *core.Context) (*core.ProjectInfo, error) {
	if projectName != "" {
		projectPath := filepath.Join(cfg.ProjectsDirectory, projectName)
		if err := client.ValidateRepository(projectPath); err == nil {
			return buildProjectInfo(ctx, client, cfg, projectPath)
		}
		// Fall back to a case-insensitive search of the projects directory.
		return findProjectByName(ctx, client, cfg, projectName)
	}

	if currentCtx != nil {
		switch currentCtx.Type {
		case core.ContextProject, core.ContextWorktree:
			projectPath := currentCtx.Path
			if currentCtx.Type == core.ContextWorktree {
				mainRepo := core.FindMainRepoByTraversal(projectPath)
				if mainRepo != "" {
					projectPath = mainRepo
				}
			}
			return buildProjectInfo(ctx, client, cfg, projectPath)
		}
		return nil, core.NewOpValidationError("DiscoverProject", "projectName", "", "project name required when outside git context")
	}

	return nil, core.NewOpValidationError("DiscoverProject", "projectName", "", "project name required when outside git context")
}

// buildProjectInfo returns the full *core.ProjectInfo for the path.
func buildProjectInfo(ctx context.Context, client *git.Client, _ *core.Config, projectPath string) (*core.ProjectInfo, error) {
	if err := client.ValidateRepository(projectPath); err != nil {
		return nil, &core.OperationError{
			Op:      "discover.project",
			Entity:  projectPath,
			Message: "invalid git repository",
			Cause:   err,
		}
	}

	mainRepoPath := projectPath
	if resolved := core.FindMainRepoByTraversal(projectPath); resolved != "" {
		mainRepoPath = resolved
	}

	repoInfo, err := client.Repository(ctx, mainRepoPath)
	if err != nil {
		repoInfo = &core.Repository{Path: mainRepoPath}
	}

	worktrees, err := client.ListWorktrees(ctx, mainRepoPath)
	if err != nil {
		worktrees = nil
	}

	worktreePtrs := make([]*core.Worktree, len(worktrees))
	for i := range worktrees {
		worktreePtrs[i] = &worktrees[i]
	}

	branchPtrs := make([]*core.Branch, len(repoInfo.Branches))
	for i := range repoInfo.Branches {
		branchPtrs[i] = &repoInfo.Branches[i]
	}

	remotePtrs := make([]*core.Remote, len(repoInfo.Remotes))
	for i := range repoInfo.Remotes {
		remotePtrs[i] = &repoInfo.Remotes[i]
	}

	return &core.ProjectInfo{
		Name:          filepath.Base(mainRepoPath),
		Path:          projectPath,
		GitRepoPath:   mainRepoPath,
		Worktrees:     worktreePtrs,
		Branches:      branchPtrs,
		Remotes:       remotePtrs,
		DefaultBranch: repoInfo.DefaultBranch,
		IsBare:        repoInfo.IsBare,
	}, nil
}

// findProjectByName does a case-insensitive directory scan of
// cfg.ProjectsDirectory.
func findProjectByName(ctx context.Context, client *git.Client, cfg *core.Config, projectName string) (*core.ProjectInfo, error) {
	entries, err := os.ReadDir(cfg.ProjectsDirectory)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "discover.project",
			Entity:  projectName,
			Message: "failed to search projects",
			Cause:   err,
		}
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.EqualFold(entry.Name(), projectName) {
			projectPath := filepath.Join(cfg.ProjectsDirectory, entry.Name())
			return buildProjectInfo(ctx, client, cfg, projectPath)
		}
	}
	return nil, &core.OperationError{
		Op:      "discover.project",
		Entity:  projectName,
		Message: "project not found",
	}
}

// calculateWorktreePath builds the worktree path under
// cfg.WorktreesDirectory for the project/branch pair.
func calculateWorktreePath(cfg *core.Config, projectName, branchName string) string {
	return filepath.Join(cfg.WorktreesDirectory, filepath.Base(projectName), filepath.Base(branchName))
}

// parseProjectBranch parses the project/branch specification
func parseProjectBranch(spec string, ctx *core.Context) (string, string, error) {
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", core.NewOpValidationError("parseProjectBranch", "spec", spec, "invalid format: expected <project>/<branch>")
		}
		projectName := parts[0]

		if validation := core.ValidateProjectName(projectName); validation.IsError() {
			return "", "", validation.Error
		}

		return projectName, parts[1], nil
	}

	if ctx != nil && ctx.ProjectName != "" {
		return ctx.ProjectName, spec, nil
	}

	return "", "", core.NewOpValidationError("parseProjectBranch", "spec", spec, "cannot infer project: not in a project context and no project specified")
}

// extractBranchNameForValidation extracts branch name from spec for validation
func extractBranchNameForValidation(spec string) string {
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		if len(parts) == 2 && parts[1] != "" {
			return parts[1]
		}
	}
	return spec
}

// displayCreateSuccess displays the success message for worktree creation
func displayCreateSuccess(out io.Writer, worktree *core.Worktree) error {
	_, err := fmt.Fprintf(out, "Created worktree: %s -> %s\n", worktree.Branch, worktree.Path)
	if err != nil {
		return fmt.Errorf("failed to display success message: %w", err)
	}
	return nil
}

// displayHookFailures displays hook failure warnings to stderr
func displayHookFailures(out io.Writer, result *core.HookResult) {
	out = writeOrIgnore(out)
	_, _ = fmt.Fprintf(out, "\nWarning: %d post-create hook(s) failed. Worktree created but setup may be incomplete.\n", len(result.Failures))
	for _, failure := range result.Failures {
		_, _ = fmt.Fprintf(out, "\n  Command: %s\n", failure.Command)
		_, _ = fmt.Fprintf(out, "  Exit code: %d\n", failure.ExitCode)
		if failure.Output != "" {
			_, _ = fmt.Fprintf(out, "  Output:\n")
			for line := range strings.SplitSeq(failure.Output, "\n") {
				_, _ = fmt.Fprintf(out, "    %s\n", line)
			}
		}
	}
}
