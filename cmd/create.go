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
	Spec     string
	Source   string
	IsCdFlag bool
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
	cmd.Flags().BoolVarP(&opts.IsCdFlag, "cd", "C", false, "Output worktree path to stdout (for shell wrapper)")

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

	if err := validateBranchNameForCreate(opts); err != nil {
		return err
	}

	req, gitClient, err := prepareCreateRequest(ctx, opts)
	if err != nil {
		return err
	}

	worktree, err := materialiseWorktree(ctx, gitClient, opts, req)
	if err != nil {
		return err
	}

	result := &core.CreateWorktreeResult{Worktree: worktree}
	result.HookResult = dispatchCreateHooks(ctx, opts, req)

	renderCreateResult(opts, result)
	return nil
}

// createRequest is the cmd-level resolved request after
// prepareCreateRequest finishes. Project carries the full
// *core.ProjectInfo (so dispatchCreateHooks can resolve the
// .twiggit.toml config), WorktreePath is computed from cfg +
// project.Name + branchName, and SourceBranch mirrors opts.Source so
// the helper signatures don't need the full options struct.
type createRequest struct {
	BranchName   string
	SourceBranch string
	Project      *core.ProjectInfo
	WorktreePath string
}

// validateBranchNameForCreate extracts the branch-name fragment from
// the spec and runs the canonical core validator. Returns the error
// directly so runCreate can short-circuit before any I/O.
func validateBranchNameForCreate(opts *CreateOptions) error {
	branchName := extractBranchNameForValidation(opts.Spec)
	branchValidation := core.ValidateBranchName(branchName)
	if branchValidation.IsError() {
		return branchValidation.Error
	}
	return nil
}

// prepareCreateRequest resolves the spec into a *createRequest with
// the project metadata, worktree path, and source branch all checked.
// It performs every read-side step: context detection, project
// discovery, source-branch existence check, worktree-path conflict
// check. Returns the resolved request AND the typed *git.Client so
// materialiseWorktree can perform the write-side step without a
// second type-assertion round-trip through cmdutil.Client = any.
func prepareCreateRequest(ctx context.Context, opts *CreateOptions) (*createRequest, *git.Client, error) {
	currentCtx, cfg, gitClient, err := loadCreateContext(opts)
	if err != nil {
		return nil, nil, err
	}

	projectName, branchName, err := parseProjectBranch(opts.Spec, currentCtx)
	if err != nil {
		return nil, nil, err
	}

	project, err := discoverProject(ctx, gitClient, cfg, projectName, currentCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover project %s: %w", projectName, err)
	}

	sourceBranchExists, err := gitClient.BranchExists(ctx, project.GitRepoPath, opts.Source)
	if err != nil {
		return nil, nil, core.NewOpValidationError("CreateWorktreeRequest", "source", opts.Source, "failed to check if source branch exists: "+err.Error())
	}
	if !sourceBranchExists {
		return nil, nil, core.NewOpValidationError("CreateWorktreeRequest", "source", opts.Source, fmt.Sprintf("source branch '%s' does not exist", opts.Source))
	}

	worktreePath := calculateWorktreePath(cfg, project.Name, branchName)
	if _, err := os.Stat(worktreePath); err == nil {
		return nil, nil, core.NewConflictError("worktree", branchName, "CreateWorktree", "worktree already exists at "+worktreePath, nil)
	}

	return &createRequest{
		BranchName:   branchName,
		SourceBranch: opts.Source,
		Project:      project,
		WorktreePath: worktreePath,
	}, gitClient, nil
}

// loadCreateContext runs the three cmd-side pre-flight steps: detect
// the current git context, load config, and resolve the composite
// git client. Centralising them keeps prepareCreateRequest focused
// on the spec-to-request resolution.
func loadCreateContext(opts *CreateOptions) (*core.Context, *core.Config, *git.Client, error) {
	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return nil, nil, nil, err
	}
	cfg, err := opts.Config()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("config load failed: %w", err)
	}
	return currentCtx, cfg, gitClient, nil
}

// materialiseWorktree creates the worktree directory's parent and
// dispatches the git worktree add command. Returns the new
// *core.Worktree on success so the caller can attach it to the
// CreateWorktreeResult. The split from prepareCreateRequest keeps the
// I/O-causing step clearly separated from the read-side validation.
func materialiseWorktree(ctx context.Context, gitClient *git.Client, opts *CreateOptions, req *createRequest) (*core.Worktree, error) {
	parentDir := filepath.Dir(req.WorktreePath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil { // #nosec G301 -- standard directory perms (rwxr-xr-x)
		return nil, fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	if err := gitClient.CreateWorktree(ctx, req.Project.GitRepoPath, req.BranchName, req.SourceBranch, req.WorktreePath); err != nil {
		return nil, &core.OperationError{
			Op:      "create.worktree",
			Entity:  req.WorktreePath,
			Message: "failed to create worktree",
			Cause:   err,
		}
	}

	verbosef(opts.IO, "Creating worktree for %s/%s", req.Project.Name, req.BranchName)
	verbosef(opts.IO, "from branch: %s", req.SourceBranch)
	verbosef(opts.IO, "to path: %s", req.Project.Name+"/"+req.BranchName)

	// Persist the tracked base so the subsequent `twiggit rebase`
	// knows which branch to rebase onto. Failure is non-fatal: the
	// worktree already exists, the user can rebase manually or run
	// `twiggit rebase --set-base` later.
	if err := gitClient.SetTrackedBase(ctx, req.WorktreePath, req.SourceBranch); err != nil {
		opts.IO.Logger.Warn("set tracked base failed",
			"worktree_path", req.WorktreePath,
			"tracked_base", req.SourceBranch,
			"err", err,
		)
	}

	return &core.Worktree{Path: req.WorktreePath, Branch: req.BranchName}, nil
}

// dispatchCreateHooks runs the post-create hook runner against the
// freshly created worktree. Errors are logged but do not abort the
// create — the worktree already exists. Returns the *HookResult so
// the formatter can show per-command failures if any.
func dispatchCreateHooks(ctx context.Context, opts *CreateOptions, req *createRequest) *core.HookResult {
	cfg, err := opts.Config()
	if err != nil {
		opts.IO.Logger.Error("hook config load failed", "error", err)
		return nil
	}
	hookResult, err := runPostCreateHooks(ctx, nil, cfg, req.Project, req.BranchName, req.SourceBranch, req.WorktreePath)
	if err != nil {
		opts.IO.Logger.Error("hook execution failed",
			"error", err,
			"worktree_path", req.WorktreePath,
			"hook_type", core.HookTypePostCreate.String(),
		)
	}
	return hookResult
}

// renderCreateResult emits the post-create output: the worktree path
// for `-C` callers, the success message, and the hook-failure summary
// when any command failed. Quiet mode suppresses the success line.
func renderCreateResult(opts *CreateOptions, result *core.CreateWorktreeResult) {
	if opts.IsCdFlag {
		_, _ = fmt.Fprintln(writeOrIgnore(opts.IO.Out), result.Worktree.Path)
		return
	}
	if !opts.IO.IsQuiet {
		_ = displayCreateSuccess(opts.IO.Out, result.Worktree)
	}
	if result.HookResult != nil && !result.HookResult.IsSuccessful {
		displayHookFailures(opts.IO.ErrOut, result.HookResult)
	}
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
		HookType:       core.HookTypePostCreate,
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
				mainRepo := git.FindMainRepoByTraversal(projectPath)
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

// buildProjectInfo delegates to the shared git adapter helper.
func buildProjectInfo(ctx context.Context, client *git.Client, _ *core.Config, projectPath string) (*core.ProjectInfo, error) {
	info, err := git.ProjectInfoFromGitDir(ctx, client, projectPath)
	if err != nil {
		return nil, fmt.Errorf("discover project %s: %w", projectPath, err)
	}
	return info, nil
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
