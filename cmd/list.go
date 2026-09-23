package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
)

// ListOptions captures every input to the runList entry point.
// RunE populates it from the cobra flag system + Factory so the
// runList body reads a single value-type struct instead of juggling
// *cobra.Command, args, and Factory references.
type ListOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (*git.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions

	// Per-command flag fields.
	All bool
}

// NewCmdList creates a new list command.
//
// runF is the optional override used by tests; pass nil to install
// the default runList body. The persistent --output / --quiet /
// --verbose flags are inherited from the root cobra command via
// cmdutil.AddPersistentFlags; runList reads them off
// opts.GlobalOptions.
func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List worktrees",
		Long: `List worktrees for the current project or all projects.
By default, lists worktrees for the detected project context.

Examples:
  twiggit list              List worktrees for current project
  twiggit list -a           List worktrees from all projects
  twiggit list --output json  Output in JSON format for scripts`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.NoArgs),
		RunE: func(_ *cobra.Command, _ []string) error {
			out := ""
			if opts.GlobalOptions != nil {
				out = opts.GlobalOptions.Output
			}
			if out != "" && out != "text" && out != "json" {
				return fmt.Errorf("invalid output format '%s': must be 'text' or 'json'", out)
			}
			if runF != nil {
				return runF(opts)
			}
			return runList(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "List worktrees from all projects")

	return cmd
}

// runList implements the orchestration that previously lived in
// worktreeService.ListWorktrees + projectService.ListProjectSummaries.
// It composes the git context detector, RepoFinder, and composite
// Client (which embeds the read- and write-side halves) directly so
// no service-layer indirection is required.
func runList(opts *ListOptions) error {
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

	worktrees, err := listWorktrees(ctx, gitClient, cfg, currentCtx, opts.All)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	verbosef(opts.IO, 1, "Listing worktrees")
	if opts.All {
		verbosef(opts.IO, 2, "  repository: all projects")
		verbosef(opts.IO, 2, "  including main worktree: false")
	} else if currentCtx.ProjectName != "" {
		verbosef(opts.IO, 2, "  project: %s", currentCtx.ProjectName)
		verbosef(opts.IO, 2, "  including main worktree: false")
	}

	format := ""
	if opts.GlobalOptions != nil {
		format = opts.GlobalOptions.Output
	}
	var formatter OutputFormatter
	if format == "json" {
		formatter = &JSONFormatter{}
	} else {
		formatter = &TextFormatter{}
	}

	if err := displayWorktrees(opts.IO.Out, worktrees, formatter); err != nil {
		return err
	}

	return nil
}

// listWorktrees returns the slice of *core.WorktreeInfo matching the
// list request. When listAll is set every discovered project is
// queried; otherwise the current context's project is used. When the
// caller has no project context, plain `list` falls through to the
// all-projects branch so an empty projects directory produces the
// friendly "No worktrees found" line (matches `list --all`).
func listWorktrees(ctx context.Context, client *git.Client, cfg *core.Config, currentCtx *core.Context, listAll bool) ([]*core.WorktreeInfo, error) {
	if listAll {
		return listAllProjectsWorktrees(ctx, client, cfg)
	}

	projectName := currentCtx.ProjectName
	if projectName == "" {
		return listAllProjectsWorktrees(ctx, client, cfg)
	}

	repoPath := filepath.Join(cfg.ProjectsDirectory, projectName)
	if err := client.ValidateRepository(repoPath); err != nil {
		return nil, &core.OperationError{
			Op:      "list.worktrees",
			Entity:  repoPath,
			Message: "project path is not a valid git repository",
		}
	}

	worktrees, err := client.ListWorktrees(ctx, repoPath)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "list.worktrees",
			Entity:  repoPath,
			Message: "failed to list worktrees",
			Cause:   err,
		}
	}

	filtered := filterNonMain(worktrees, repoPath)
	out := make([]*core.WorktreeInfo, len(filtered))
	for i := range filtered {
		out[i] = &filtered[i]
	}
	return out, nil
}

// listAllProjectsWorktrees discovers every project under
// cfg.ProjectsDirectory, lists worktrees for each, and returns the
// flattened set (main worktrees are excluded to match the legacy
// behavior).
func listAllProjectsWorktrees(ctx context.Context, client *git.Client, cfg *core.Config) ([]*core.WorktreeInfo, error) {
	finder := git.NewRepoFinder(client)
	gitDirs, err := finder.FindGitRepositories(cfg.ProjectsDirectory)
	if err != nil {
		return nil, &core.OperationError{
			Op:      "list.worktrees",
			Entity:  cfg.ProjectsDirectory,
			Message: "failed to scan for git repositories",
			Cause:   err,
		}
	}

	var out []*core.WorktreeInfo
	for _, gitDir := range gitDirs {
		worktrees, err := client.ListWorktrees(ctx, gitDir.Path)
		if err != nil {
			continue
		}
		filtered := filterNonMain(worktrees, gitDir.Path)
		for i := range filtered {
			out = append(out, &filtered[i])
		}
	}
	return out, nil
}

// filterNonMain returns the worktrees that are not the repo's main
// checkout. Mirrors the legacy "IncludeMain: false" default.
func filterNonMain(worktrees []core.WorktreeInfo, repoPath string) []core.WorktreeInfo {
	out := make([]core.WorktreeInfo, 0, len(worktrees))
	for _, wt := range worktrees {
		if wt.Path != repoPath {
			out = append(out, wt)
		}
	}
	return slices.Clone(out)
}

// displayWorktrees displays the worktrees using the specified formatter
func displayWorktrees(out io.Writer, worktrees []*core.WorktreeInfo, formatter OutputFormatter) error {
	formatted := formatter.FormatWorktrees(worktrees)
	if _, err := fmt.Fprint(out, formatted); err != nil {
		return fmt.Errorf("failed to display worktrees: %w", err)
	}
	return nil
}
