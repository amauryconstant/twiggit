package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
	"twiggit/internal/output"

	"github.com/spf13/cobra"
)

// StatusOptions captures every input to runStatus.
type StatusOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions
	Logger        *slog.Logger

	All         bool
	StaleBehind int
	StaleDays   int
	Args        []string

	// staleBehindSet / staleDaysSet mark whether the corresponding
	// flag was supplied on the command line. The values are kept on
	// the Options so the runF seam can override them in tests.
	staleBehindSet bool
	staleDaysSet   bool
}

// NewCmdStatus creates a new status command.
//
// runF is the optional override used by tests; pass nil to install
// the default runStatus body.
func NewCmdStatus(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

	cmd := &cobra.Command{
		Use:   "status [project]",
		Short: "Show per-worktree diagnostic status",
		Long: `Show the diagnostic view of every worktree: ahead/behind counts against
the tracked base, merge readiness, dirty state, last-commit date, and a stale
flag. Output is human-readable by default and pipeable via --output.

Examples:
  twiggit status                 Show worktree status for the current project
  twiggit status myproject       Show worktree status for a named project
  twiggit status --all           Show worktree status across every project
  twiggit status --output json   Emit a bare JSON array for scripts`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = BoundaryLogger(f, cmd)
			opts.Args = args
			opts.staleBehindSet = cmd.Flags().Changed("stale-behind")
			opts.staleDaysSet = cmd.Flags().Changed("stale-days")
			opts.All, _ = cmd.Flags().GetBool("all")
			opts.StaleBehind, _ = cmd.Flags().GetInt("stale-behind")
			opts.StaleDays, _ = cmd.Flags().GetInt("stale-days")
			format := ""
			if opts.GlobalOptions != nil {
				format = opts.GlobalOptions.Output
			}
			if _, err := output.NewFormatter(format); err != nil {
				//nolint:wrapcheck // core.UsageError is returned unwrapped per the [errors-wrap-context] project rule; NewFormatter returns a UsageError for unknown formats.
				return err
			}
			if opts.All && len(args) > 0 {
				return core.NewUsageError(
					"--all and a positional project argument are mutually exclusive",
					nil,
				)
			}
			if runF != nil {
				return runF(opts)
			}
			return runStatus(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Show worktrees from all projects")
	cmd.Flags().IntVar(&opts.StaleBehind, "stale-behind", 0, "Override StaleBehind threshold (0 disables)")
	cmd.Flags().IntVar(&opts.StaleDays, "stale-days", 0, "Override StaleDays threshold (0 disables)")

	return cmd
}

// runStatus composes the per-project walk that fills the diagnostic
// rows, derives per-row IsStale from the resolved thresholds, and
// emits the projection through the chosen formatter. Mirrors
// runList's shape so the per-command output machinery stays uniform.
func runStatus(opts *StatusOptions) error {
	ctx := opts.Ctx

	currentCtx, gitClient, err := detectContext(opts.Config, opts.GitClient)
	if err != nil {
		return err
	}
	if !opts.All && (currentCtx == nil || currentCtx.Type == core.ContextOutsideGit) && len(opts.Args) == 0 {
		return core.NewUsageError(
			"status: not inside a git context; provide a [project] argument or use --all",
			nil,
		)
	}
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	format := ""
	if opts.GlobalOptions != nil {
		format = opts.GlobalOptions.Output
	}
	formatter, err := output.NewFormatter(format)
	if err != nil {
		//nolint:wrapcheck // core.UsageError is returned unwrapped per the [errors-wrap-context] project rule; NewFormatter returns a UsageError for unknown formats.
		return err
	}

	effectiveCfg := resolvedStaleConfig(cfg, opts)
	rows, err := runStatusWalk(ctx, opts, gitClient, currentCtx, effectiveCfg)
	if err != nil {
		return fmt.Errorf("failed to gather status: %w", err)
	}

	includeProjectColumn := opts.All
	if err := renderStatus(opts.IO.Out, rows, includeProjectColumn, formatter); err != nil {
		return err
	}
	return nil
}

// resolvedStaleConfig returns a copy of cfg with the per-invocation
// --stale-behind / --stale-days overrides applied. The copy is
// shallow because StatusConfig contains only int fields; the
// underlying slice fields (ProtectedBranches) are not mutated.
func resolvedStaleConfig(cfg *core.Config, opts *StatusOptions) *core.Config {
	out := *cfg
	out.Status = cfg.Status
	if opts.staleBehindSet {
		out.Status.StaleBehind = opts.StaleBehind
	}
	if opts.staleDaysSet {
		out.Status.StaleDays = opts.StaleDays
	}
	return &out
}

// runStatusWalk returns the per-worktree diagnostic rows for the
// requested scope (current project, --all, or positional project).
// Each worktree read goes through the WorktreeStatusReader role
// installed on the composite *git.Client; per-row errors populate
// the skip fields on the row, they never abort the walk.
func runStatusWalk(
	ctx context.Context,
	opts *StatusOptions,
	gitClient *git.Client,
	currentCtx *core.Context,
	cfg *core.Config,
) ([]core.WorktreeStatus, error) {
	projects, err := resolveStatusProjects(ctx, opts, gitClient, currentCtx, cfg)
	if err != nil {
		return nil, err
	}

	var rows []core.WorktreeStatus
	for _, project := range projects {
		rows = append(rows, walkProjectStatus(ctx, opts, gitClient, project, cfg)...)
	}
	return rows, nil
}

// resolveStatusProjects returns the set of projects to walk for the
// invocation. Mirrors the resolvePruneProjects shape from cmd/prune
// so the per-command worktree-mutation walk stays uniform.
func resolveStatusProjects(
	_ context.Context,
	opts *StatusOptions,
	client *git.Client,
	currentCtx *core.Context,
	_ *core.Config,
) ([]*core.ProjectInfo, error) {
	if opts.All {
		finder := git.NewRepoFinder(client)
		gitDirs, err := finder.FindGitRepositories(cfgStatusProjectsDir(opts))
		if err != nil {
			return nil, &core.OperationError{
				Op:      "status.projects",
				Entity:  cfgStatusProjectsDir(opts),
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

	// Single project: either the current context's project or a
	// positional argument. Falls back to project discovery from
	// ProjectsDirectory when the caller is outside a git context.
	projectName := ""
	if currentCtx != nil {
		projectName = currentCtx.ProjectName
	}
	if projectName == "" {
		// Outside git without --all. runStatus short-circuits before
		// this path when no positional is supplied, so reaching here
		// means a positional was given.
		project, err := discoverProjectFromOpts(opts, client, "")
		if err != nil {
			return nil, err
		}
		return []*core.ProjectInfo{project}, nil
	}
	project, err := discoverProjectFromOpts(opts, client, projectName)
	if err != nil {
		return nil, err
	}
	return []*core.ProjectInfo{project}, nil
}

// cfgStatusProjectsDir returns the configured projects directory by
// invoking opts.Config; centralises the error path so the
// --all branch reports a single OperationError.
func cfgStatusProjectsDir(opts *StatusOptions) string {
	if opts.Config == nil {
		return ""
	}
	cfg, err := opts.Config()
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.ProjectsDirectory
}

// discoverProjectFromOpts reuses the discoverProject helper from
// cmd/create.go. Kept thin so tests can pass an isolated Factory.
func discoverProjectFromOpts(
	opts *StatusOptions,
	client *git.Client,
	name string,
) (*core.ProjectInfo, error) {
	if opts.Config == nil {
		return nil, core.NewOpValidationError("discoverProject", "config", "", "config loader not wired")
	}
	cfg, err := opts.Config()
	if err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}
	return discoverProject(opts.Ctx, client, cfg, name, nil)
}

// walkProjectStatus emits one core.WorktreeStatus per non-main
// worktree in the project. Per-worktree read failures populate the
// skip fields on the row instead of aborting the walk. The
// contributing project name is stamped onto every row so the
// renderer can include the leading column under --all.
func walkProjectStatus(
	ctx context.Context,
	opts *StatusOptions,
	gitClient *git.Client,
	project *core.ProjectInfo,
	cfg *core.Config,
) []core.WorktreeStatus {
	worktrees, err := gitClient.ListWorktrees(ctx, project.GitRepoPath)
	if err != nil {
		// One skipped row representing the project-level failure so
		// the user sees something rather than a silent absence.
		row := core.WorktreeStatus{
			ProjectName: project.Name,
			Base:        "",
			IsSkipped:   true,
			SkipReason:  "failed to list worktrees: " + err.Error(),
		}
		return []core.WorktreeStatus{row}
	}

	rows := make([]core.WorktreeStatus, 0, len(worktrees))
	for _, wt := range worktrees {
		if wt.Path == project.GitRepoPath {
			continue
		}
		row, err := gitClient.ReadWorktreeStatus(ctx, cfg, project.GitRepoPath, wt.Path)
		if err != nil {
			opts.Logger.Warn("worktree status precondition failed",
				"wt", wt.Path, "err", err)
		}
		row.ProjectName = project.Name
		row.IsStale = row.ComputeIsStale(cfg)
		rows = append(rows, row)
		emitRowSkipWarning(opts, row)
	}
	return rows
}

// emitRowSkipWarning writes a per-row warning to stderr when the
// row was skipped and --quiet is not set. Skipped rows are still
// present in the data projection so the warning is purely a UX
// affordance.
func emitRowSkipWarning(opts *StatusOptions, row core.WorktreeStatus) {
	if !row.IsSkipped {
		return
	}
	if opts.IO == nil || opts.IO.IsQuiet {
		return
	}
	branch := ""
	if row.Worktree != nil {
		branch = row.Worktree.Branch
	}
	path := ""
	if row.Worktree != nil {
		path = row.Worktree.Path
	}
	_, _ = fmt.Fprintf(writeOrIgnore(opts.IO.ErrOut),
		"warning: %s (%s): %s\n", branch, path, row.SkipReason)
}

// statusRows is the per-command projection for table and plain
// output. It implements output.Tabular (header + row) and the
// json.Marshaler contract for the bare-array wire shape. When
// includeProjectColumn is set (--all invocation) the leading
// PROJECT column is added to the header, every row, and the
// JSON object key.
type statusRows struct {
	rows                 []core.WorktreeStatus
	includeProjectColumn bool
}

// Header is the canonical column order for status output. The order
// matches the spec (BRANCH, PATH, AHEAD, BEHIND, BASE, MERGED, DIRTY, STALE),
// with PROJECT prepended when includeProjectColumn is set.
func (r statusRows) Header() []string {
	if r.includeProjectColumn {
		return []string{"PROJECT", "BRANCH", "PATH", "AHEAD", "BEHIND", "BASE", "MERGED", "DIRTY", "STALE"}
	}
	return []string{"BRANCH", "PATH", "AHEAD", "BEHIND", "BASE", "MERGED", "DIRTY", "STALE"}
}

// Rows projects each core.WorktreeStatus to a string row aligned
// with Header.
func (r statusRows) Rows() [][]string {
	out := make([][]string, 0, len(r.rows))
	for _, row := range r.rows {
		rowOut := []string{
			statusBranch(row),
			statusPath(row),
			statusAhead(row),
			statusBehind(row),
			row.Base,
			statusMerged(row),
			statusDirty(row),
			statusStale(row),
		}
		if r.includeProjectColumn {
			rowOut = append([]string{row.ProjectName}, rowOut...)
		}
		out = append(out, rowOut)
	}
	return out
}

// MarshalJSON emits the bare-array shape pinned by the cli-status
// spec. Object keys: branch, path, base, ahead, behind, merged,
// dirty, last_commit_date, stale, skipped, skip_reason. The
// skip_reason key is omitted when empty (per the omitempty tag on
// core.WorktreeStatus.SkipReason).
func (r statusRows) MarshalJSON() ([]byte, error) {
	type entry struct {
		Project        string `json:"project,omitempty"`
		Branch         string `json:"branch"`
		Path           string `json:"path"`
		Base           string `json:"base"`
		Ahead          int    `json:"ahead"`
		Behind         int    `json:"behind"`
		Merged         bool   `json:"merged"`
		Dirty          bool   `json:"dirty"`
		LastCommitDate string `json:"last_commit_date"`
		Stale          bool   `json:"stale"`
		Skipped        bool   `json:"skipped"`
		SkipReason     string `json:"skip_reason,omitempty"`
	}
	entries := make([]entry, 0, len(r.rows))
	for _, row := range r.rows {
		e := entry{
			Branch:         statusBranch(row),
			Path:           statusPath(row),
			Base:           row.Base,
			Ahead:          statusAheadInt(row),
			Behind:         statusBehindInt(row),
			Merged:         row.IsMerged,
			Dirty:          row.Dirty(),
			LastCommitDate: formatLastCommitDate(row.LastCommitDate),
			Stale:          row.IsStale,
			Skipped:        row.IsSkipped,
			SkipReason:     row.SkipReason,
		}
		if r.includeProjectColumn {
			e.Project = row.ProjectName
		}
		entries = append(entries, e)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshal status rows: %w", err)
	}
	return data, nil
}

func statusBranch(row core.WorktreeStatus) string {
	if row.Worktree != nil {
		return row.Worktree.Branch
	}
	return ""
}

func statusPath(row core.WorktreeStatus) string {
	if row.Worktree != nil {
		return row.Worktree.Path
	}
	return ""
}

func statusAhead(row core.WorktreeStatus) string {
	return statusCount(row, true)
}

func statusBehind(row core.WorktreeStatus) string {
	return statusCount(row, false)
}

func statusAheadInt(row core.WorktreeStatus) int {
	if row.RepositoryStatus == nil {
		return 0
	}
	return row.RepositoryStatus.Ahead
}

func statusBehindInt(row core.WorktreeStatus) int {
	if row.RepositoryStatus == nil {
		return 0
	}
	return row.RepositoryStatus.Behind
}

func statusCount(row core.WorktreeStatus, ahead bool) string {
	if row.RepositoryStatus == nil {
		return "0"
	}
	if ahead {
		return strconv.Itoa(row.RepositoryStatus.Ahead)
	}
	return strconv.Itoa(row.RepositoryStatus.Behind)
}

func statusMerged(row core.WorktreeStatus) string {
	if row.IsSkipped {
		return "-"
	}
	if row.IsMerged {
		return "true"
	}
	return "false"
}

func statusDirty(row core.WorktreeStatus) string {
	if row.IsSkipped {
		return "-"
	}
	if row.Dirty() {
		return "true"
	}
	return "false"
}

func statusStale(row core.WorktreeStatus) string {
	if row.IsSkipped {
		return "-"
	}
	if row.IsStale {
		return "true"
	}
	return "false"
}

func formatLastCommitDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

// renderStatus dispatches to the formatter when one is resolved, and
// falls through to a default projection otherwise. Mirrors
// renderWorktrees in cmd/list.go. includeProjectColumn gates the
// leading PROJECT column / JSON key (set by --all).
func renderStatus(
	out io.Writer,
	rows []core.WorktreeStatus,
	includeProjectColumn bool,
	formatter output.Formatter,
) error {
	projection := statusRows{rows: rows, includeProjectColumn: includeProjectColumn}
	if formatter == nil {
		// Default projection: a flat row per worktree, no header,
		// columns separated by tab. Empty collection emits nothing.
		for _, row := range projection.Rows() {
			if _, err := fmt.Fprintln(out, strings.Join(row, "\t")); err != nil {
				return fmt.Errorf("failed to display status: %w", err)
			}
		}
		return nil
	}
	if err := formatter.Write(out, projection); err != nil {
		return fmt.Errorf("render status: %w", err)
	}
	return nil
}
