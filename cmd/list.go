package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
	"twiggit/internal/output"

	"github.com/spf13/cobra"
)

// ListOptions captures every input to the runList entry point.
// RunE populates it from the cobra flag system + Factory so the
// runList body reads a single value-type struct instead of juggling
// *cobra.Command, args, and Factory references.
type ListOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			out := ""
			if opts.GlobalOptions != nil {
				out = opts.GlobalOptions.Output
			}
			if _, err := output.NewFormatter(out); err != nil {
				return fmt.Errorf("list: %w", err)
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

	client, err := opts.GitClient()
	if err != nil {
		return fmt.Errorf("git client init failed: %w", err)
	}
	gitClient, _ := client.(*git.Client)

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
		opts.Logger.Debug("detect failed", "err", err)
		return fmt.Errorf("context detection failed: %w", err)
	}

	worktrees, err := listWorktrees(ctx, gitClient, cfg, currentCtx, opts.All)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	verbosef(opts.IO, "Listing worktrees")
	if opts.All {
		verbosef(opts.IO, "repository: all projects")
		verbosef(opts.IO, "including main worktree: false")
	} else if currentCtx.ProjectName != "" {
		verbosef(opts.IO, "project: %s", currentCtx.ProjectName)
		verbosef(opts.IO, "including main worktree: false")
	}

	format := ""
	if opts.GlobalOptions != nil {
		format = opts.GlobalOptions.Output
	}
	formatter, err := output.NewFormatter(format)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	if err := renderWorktrees(opts.IO.Out, worktrees, formatter); err != nil {
		return err
	}

	return nil
}

// worktreeRows projects a slice of *core.Worktree into the
// Tabular interface consumed by TableFormatter and PlainFormatter
// and the JSON shape consumed by JSONFormatter via MarshalJSON.
// The Header is the canonical column order for list-style output;
// Rows projects each worktree to the matching string triplet, with
// status ∈ {clean, modified, detached} per the spec.
type worktreeRows struct {
	worktrees []*core.Worktree
}

func (w worktreeRows) Header() []string {
	return []string{"BRANCH", "PATH", "STATUS"}
}

func (w worktreeRows) Rows() [][]string {
	rows := make([][]string, 0, len(w.worktrees))
	for _, wt := range w.worktrees {
		rows = append(rows, []string{wt.Branch, wt.Path, worktreeStatus(wt)})
	}
	return rows
}

// MarshalJSON emits the projection as a bare JSON array of
// `{branch,path,status}` objects per the cli-output spec JSON
// shape table. This lets JSONFormatter.Write accept worktreeRows
// without an envelope wrapper; the bare-array shape is required
// for small collections per the spec.
func (w worktreeRows) MarshalJSON() ([]byte, error) {
	type entry struct {
		Branch string `json:"branch"`
		Path   string `json:"path"`
		Status string `json:"status"`
	}
	entries := make([]entry, 0, len(w.worktrees))
	for _, wt := range w.worktrees {
		entries = append(entries, entry{
			Branch: wt.Branch,
			Path:   wt.Path,
			Status: worktreeStatus(wt),
		})
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshal worktree rows: %w", err)
	}
	return data, nil
}

func worktreeStatus(wt *core.Worktree) string {
	switch {
	case wt.IsDetached:
		return "detached"
	case wt.IsModified:
		return "modified"
	default:
		return "clean"
	}
}

// renderWorktrees dispatches to the formatter if one was resolved
// for the requested format, otherwise falls back to the bespoke
// human-readable rendering defined by the cli-list spec for the
// empty --output default. The bespoke shape is `branch -> path`
// with optional `(modified)` / `(detached)` suffixes. An empty
// collection emits no lines per the cli-list spec
// (`Status indicators / Empty project renders no lines`).
func renderWorktrees(out io.Writer, worktrees []*core.Worktree, formatter output.Formatter) error {
	if formatter == nil {
		for _, wt := range worktrees {
			suffix := ""
			if wt.IsDetached {
				suffix = " (detached)"
			} else if wt.IsModified {
				suffix = " (modified)"
			}
			if _, err := fmt.Fprintf(out, "%s -> %s%s\n", wt.Branch, wt.Path, suffix); err != nil {
				return fmt.Errorf("failed to display worktrees: %w", err)
			}
		}
		return nil
	}
	if err := formatter.Write(out, worktreeRows{worktrees: worktrees}); err != nil {
		return fmt.Errorf("render worktrees: %w", err)
	}
	return nil
}

// listWorktrees returns the slice of *core.Worktree matching the
// list request. When listAll is set every discovered project is
// queried; otherwise the current context's project is used. When the
// caller has no project context, plain `list` falls through to the
// all-projects branch so an empty projects directory produces the
// friendly "No worktrees found" line (matches `list --all`).
func listWorktrees(ctx context.Context, client *git.Client, cfg *core.Config, currentCtx *core.Context, listAll bool) ([]*core.Worktree, error) {
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
	out := make([]*core.Worktree, len(filtered))
	for i := range filtered {
		out[i] = &filtered[i]
	}
	return out, nil
}

// listAllProjectsWorktrees discovers every project under
// cfg.ProjectsDirectory, lists worktrees for each, and returns the
// flattened set (main worktrees are excluded to match the legacy
// behavior).
func listAllProjectsWorktrees(ctx context.Context, client *git.Client, cfg *core.Config) ([]*core.Worktree, error) {
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

	var out []*core.Worktree
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
func filterNonMain(worktrees []core.Worktree, repoPath string) []core.Worktree {
	out := make([]core.Worktree, 0, len(worktrees))
	for _, wt := range worktrees {
		if wt.Path != repoPath {
			out = append(out, wt)
		}
	}
	return slices.Clone(out)
}
