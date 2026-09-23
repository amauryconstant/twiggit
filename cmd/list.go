package cmd

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"twiggit/internal/core"
	"twiggit/internal/git"
)

// NewListCommand creates a new list command
func NewListCommand(f *CommandConfig) *cobra.Command {
	var all bool
	var output string

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if output != "" && output != "text" && output != "json" {
				return fmt.Errorf("invalid output format '%s': must be 'text' or 'json'", output)
			}
			return executeList(cmd, f, all, output)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "List worktrees from all projects")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "Output format (text or json)")

	return cmd
}

// executeList implements the orchestration that previously lived in
// worktreeService.ListWorktrees + projectService.ListProjectSummaries.
// It composes the git context detector, RepoFinder, and composite
// Client (which embeds the read- and write-side halves) directly so
// no service-layer indirection is required.
func executeList(cmd *cobra.Command, f *CommandConfig, all bool, output string) error {
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

	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	currentCtx, err := detector.DetectContext(wd)
	if err != nil {
		return fmt.Errorf("context detection failed: %w", err)
	}

	worktrees, err := listWorktrees(ctx, gitClient, cfg, currentCtx, all)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	logv(cmd, 1, "Listing worktrees")
	if all {
		logv(cmd, 2, "  repository: all projects")
	} else if currentCtx.ProjectName != "" {
		logv(cmd, 2, "  project: %s", currentCtx.ProjectName)
	}

	var formatter OutputFormatter
	if output == "json" {
		formatter = &JSONFormatter{}
	} else {
		formatter = &TextFormatter{}
	}

	if err := displayWorktrees(cmd.OutOrStdout(), worktrees, formatter); err != nil {
		return err
	}

	return nil
}

// listWorktrees returns the slice of *core.WorktreeInfo matching the
// list request. When listAll is set every discovered project is
// queried; otherwise the current context's project is used.
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
