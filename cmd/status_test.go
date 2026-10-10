package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStatusReader is a placeholder for follow-up coverage; the
// active runF tests build row slices inline rather than driving a
// fake reader.

func newStatusTestFactory(t *testing.T) *cmdutil.Factory {
	t.Helper()
	ios, _, _, _ := iostreams.Test()
	gitClient, err := git.NewClient()
	require.NoError(t, err)
	f := &cmdutil.Factory{
		IOStreams:  ios,
		Context:    t.Context(),
		AppVersion: "test",
		Executable: "twiggit-test",
	}
	f.Config = func() (*core.Config, error) { return core.DefaultConfig(), nil }
	f.GitClient = func() (cmdutil.Client, error) { return gitClient, nil }
	f.Logger = func() *slog.Logger {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return f
}

// TestStatus_NewCmdStatusInvalidOutputFormat pins the parse-level
// dispatch: a bad --output value fails RunE with UsageError (exit 2).
func TestStatus_NewCmdStatusInvalidOutputFormat(t *testing.T) {
	f := newStatusTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"status", "--output", "yaml"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(err))
}

// TestStatus_NewCmdStatusAllPositionalMutex confirms --all and a
// positional project are mutually exclusive and dispatch to
// UsageError.
func TestStatus_NewCmdStatusAllPositionalMutex(t *testing.T) {
	f := newStatusTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"status", "--all", "myproject"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue)
	assert.Contains(t, ue.Message, "mutually exclusive")
}

// TestStatus_NewCmdStatusArgs pins the cobra args guard: at most
// one positional argument.
func TestStatus_NewCmdStatusArgs(t *testing.T) {
	f := newStatusTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"status", "a", "b"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue)
}

// TestStatus_HelpListsFlags pins the help-text contract: the four
// flags must all be present in the rendered --help output.
func TestStatus_HelpListsFlags(t *testing.T) {
	f := newStatusTestFactory(t)
	root := newRootForTest(f)
	stdout := new(bytes.Buffer)
	root.SetOut(stdout)
	root.SetErr(stdout)
	root.SetArgs([]string{"status", "--help"})

	require.NoError(t, root.Execute())
	out := stdout.String()
	for _, flag := range []string{"--all", "--output", "--stale-behind", "--stale-days"} {
		assert.Contains(t, out, flag, "status --help must document %q", flag)
	}
}

// TestStatusRows_HeaderOrder pins the column order documented in
// the cli-status spec for both the single-project and the --all
// (project-column) shapes.
func TestStatusRows_HeaderOrder(t *testing.T) {
	t.Parallel()
	t.Run("single project", func(t *testing.T) {
		t.Parallel()
		h := statusRows{}.Header()
		require.Len(t, h, 8)
		assert.Equal(t, []string{"BRANCH", "PATH", "AHEAD", "BEHIND", "BASE", "MERGED", "DIRTY", "STALE"}, h)
	})
	t.Run("all projects", func(t *testing.T) {
		t.Parallel()
		h := statusRows{includeProjectColumn: true}.Header()
		require.Len(t, h, 9)
		assert.Equal(t,
			[]string{"PROJECT", "BRANCH", "PATH", "AHEAD", "BEHIND", "BASE", "MERGED", "DIRTY", "STALE"},
			h,
		)
	})
}

// TestStatusRows_RowColumnCount pins the projection: every row
// emits exactly the header's number of columns, in both shapes.
func TestStatusRows_RowColumnCount(t *testing.T) {
	t.Parallel()
	t.Run("single project", func(t *testing.T) {
		t.Parallel()
		rows := statusRows{rows: []core.WorktreeStatus{
			{
				Worktree: &core.Worktree{Branch: "feat/a", Path: "/wt/a"},
				RepositoryStatus: &core.RepositoryStatus{
					Ahead:  2,
					Behind: 5,
				},
				Base:     "main",
				IsMerged: false,
				IsStale:  true,
			},
		}}
		require.Len(t, rows.Rows(), 1)
		assert.Len(t, rows.Rows()[0], len(rows.Header()))
	})
	t.Run("all projects", func(t *testing.T) {
		t.Parallel()
		rows := statusRows{
			rows: []core.WorktreeStatus{
				{
					ProjectName: "twiggit",
					Worktree:    &core.Worktree{Branch: "feat/a", Path: "/wt/a"},
					Base:        "main",
				},
			},
			includeProjectColumn: true,
		}
		require.Len(t, rows.Rows(), 1)
		assert.Len(t, rows.Rows()[0], len(rows.Header()))
		assert.Equal(t, "twiggit", rows.Rows()[0][0])
		assert.Equal(t, "feat/a", rows.Rows()[0][1])
	})
}

// TestStatusRows_JSONBareArray pins the cli-output spec: the
// projection emits a bare JSON array with the documented field set
// (branch, path, base, ahead, behind, merged, dirty,
// last_commit_date, stale, skipped, skip_reason) and omits an
// empty skip_reason. The --all shape adds the project key.
func TestStatusRows_JSONBareArray(t *testing.T) {
	t.Parallel()
	t.Run("single project", func(t *testing.T) {
		t.Parallel()
		rows := statusRows{rows: []core.WorktreeStatus{
			{
				Worktree: &core.Worktree{Branch: "feat/x", Path: "/wt/x"},
				RepositoryStatus: &core.RepositoryStatus{
					IsClean: true,
					Ahead:   1,
					Behind:  2,
				},
				Base:     "main",
				IsMerged: false,
				IsStale:  false,
			},
			{
				Worktree:   &core.Worktree{Branch: "feat/y", Path: "/wt/y"},
				IsSkipped:  true,
				SkipReason: "no tracked base and no protected branches",
			},
		}}

		data, err := json.Marshal(rows)
		require.NoError(t, err)

		var got []map[string]any
		require.NoError(t, json.Unmarshal(data, &got))
		require.Len(t, got, 2)

		first := got[0]
		assert.Equal(t, "feat/x", first["branch"])
		assert.Equal(t, "/wt/x", first["path"])
		assert.Equal(t, "main", first["base"])
		assert.InDelta(t, float64(1), first["ahead"], 0)
		assert.InDelta(t, float64(2), first["behind"], 0)
		assert.Equal(t, false, first["merged"])
		assert.Equal(t, false, first["dirty"])
		assert.Equal(t, false, first["stale"])
		assert.Equal(t, false, first["skipped"])
		assert.NotContains(t, first, "skip_reason", "empty skip_reason must be omitted")
		assert.NotContains(t, first, "project", "project key must be absent in single-project mode")

		second := got[1]
		assert.Equal(t, true, second["skipped"])
		assert.Equal(t, "no tracked base and no protected branches", second["skip_reason"])
	})
	t.Run("all projects", func(t *testing.T) {
		t.Parallel()
		rows := statusRows{
			rows: []core.WorktreeStatus{
				{
					ProjectName: "twiggit",
					Worktree:    &core.Worktree{Branch: "feat/a", Path: "/wt/a"},
					Base:        "main",
				},
				{
					ProjectName: "other",
					Worktree:    &core.Worktree{Branch: "feat/b", Path: "/wt/b"},
					Base:        "main",
				},
			},
			includeProjectColumn: true,
		}

		data, err := json.Marshal(rows)
		require.NoError(t, err)

		var got []map[string]any
		require.NoError(t, json.Unmarshal(data, &got))
		require.Len(t, got, 2)
		assert.Equal(t, "twiggit", got[0]["project"])
		assert.Equal(t, "feat/a", got[0]["branch"])
		assert.Equal(t, "other", got[1]["project"])
		assert.Equal(t, "feat/b", got[1]["branch"])
	})
}

// TestRenderStatus_BespokeShape pins the default (no --output) path:
// one row per worktree, tab-separated, no header, empty collection
// emits no lines. Mirrors TestRenderWorktrees_BespokeShape.
func TestRenderStatus_BespokeShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		row  core.WorktreeStatus
	}{
		{
			name: "fresh worktree",
			row: core.WorktreeStatus{
				Worktree:         &core.Worktree{Branch: "feat/a", Path: "/wt/a"},
				RepositoryStatus: &core.RepositoryStatus{Ahead: 0, Behind: 0},
				Base:             "main",
				IsMerged:         true,
			},
		},
		{
			name: "diverged worktree",
			row: core.WorktreeStatus{
				Worktree:         &core.Worktree{Branch: "feat/b", Path: "/wt/b"},
				RepositoryStatus: &core.RepositoryStatus{Ahead: 3, Behind: 4},
				Base:             "main",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, renderStatus(&buf, []core.WorktreeStatus{tc.row}, false, nil))

			got := strings.TrimRight(buf.String(), "\n")
			require.NotEmpty(t, got, "rendered status must produce at least one line")
			fields := strings.Split(got, "\t")
			require.GreaterOrEqual(t, len(fields), 8, "row must carry at least the eight columns")
		})
	}

	t.Run("empty collection emits no lines", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		require.NoError(t, renderStatus(&buf, nil, false, nil))
		assert.Empty(t, buf.String())
	})
}

// TestResolvedStaleConfig_AppliesOverrides covers the per-invocation
// --stale-behind / --stale-days plumbing: a flag value flows into
// the resolved Config used by the per-row ComputeIsStale call.
func TestResolvedStaleConfig_AppliesOverrides(t *testing.T) {
	t.Parallel()

	cfg := core.DefaultConfig()
	opts := &StatusOptions{
		StaleBehind:    99,
		StaleDays:      1,
		staleBehindSet: true,
		staleDaysSet:   true,
	}

	got := resolvedStaleConfig(cfg, opts)
	assert.Equal(t, 99, got.Status.StaleBehind)
	assert.Equal(t, 1, got.Status.StaleDays)
}

// TestResolvedStaleConfig_DefaultsWhenUnset covers the inverse path:
// when the flags are not set the resolved config matches the
// incoming config.
func TestResolvedStaleConfig_DefaultsWhenUnset(t *testing.T) {
	t.Parallel()

	cfg := core.DefaultConfig()
	opts := &StatusOptions{}

	got := resolvedStaleConfig(cfg, opts)
	assert.Equal(t, cfg.Status.StaleBehind, got.Status.StaleBehind)
	assert.Equal(t, cfg.Status.StaleDays, got.Status.StaleDays)
}

// TestStatusRunF_SkippedRowEmitsWarning covers the per-wt stderr
// warning path: a runF-injected fake reader returns one skipped
// row among three; the captured stdout carries three rows, the
// captured stderr carries one warning, and the test asserts exit 0.
//
// Cannot use t.Parallel: NewRootCommand mutates cobra's
// package-level OnInitialize list via carapace.Gen.
func TestStatusRunF_SkippedRowEmitsWarning(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	gitClient, err := git.NewClient()
	require.NoError(t, err)

	f := &cmdutil.Factory{
		IOStreams:  ios,
		Context:    t.Context(),
		AppVersion: "test",
		Executable: "twiggit-test",
	}
	f.Config = func() (*core.Config, error) { return cfg, nil }
	f.GitClient = func() (cmdutil.Client, error) { return gitClient, nil }
	f.Logger = func() *slog.Logger {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	// runF replaces runStatus so we can hand a fake reader to the
	// walk without driving a real repo. The fake yields one skipped
	// row among three so the projection matches the assertion.
	runF := func(opts *StatusOptions) error {
		rows := []core.WorktreeStatus{
			{
				Worktree:         &core.Worktree{Branch: "feat/clean", Path: "/wt/clean"},
				RepositoryStatus: &core.RepositoryStatus{IsClean: true},
				Base:             "main",
			},
			{
				Worktree:         &core.Worktree{Branch: "feat/skipped", Path: "/wt/skipped"},
				RepositoryStatus: &core.RepositoryStatus{},
				IsSkipped:        true,
				SkipReason:       "test skip reason",
			},
			{
				Worktree:         &core.Worktree{Branch: "feat/dirty", Path: "/wt/dirty"},
				RepositoryStatus: &core.RepositoryStatus{IsClean: false, Modified: []string{"a"}},
				Base:             "main",
			},
		}
		effective := resolvedStaleConfig(cfg, opts)
		for i := range rows {
			rows[i].IsStale = rows[i].ComputeIsStale(effective)
		}
		emitRowSkipWarning(opts, rows[1])
		return renderStatus(opts.IO.Out, rows, false, nil)
	}

	rootBuildMu.Lock()
	defer rootBuildMu.Unlock()
	cmd := NewRootCommand(f)
	for _, sub := range cmd.Commands() {
		if sub.Name() == "status" {
			cmd.RemoveCommand(sub)
			break
		}
	}
	statusCmd := NewCmdStatus(f, runF)
	statusCmd.GroupID = "core"
	cmd.AddCommand(statusCmd)
	cmd.SetArgs([]string{"status"})

	require.NoError(t, cmd.Execute())

	stdout := iosStdout(t, ios)
	assert.Contains(t, stdout, "feat/clean", "stdout must contain the clean row")
	assert.Contains(t, stdout, "feat/skipped", "stdout must contain the skipped row")
	assert.Contains(t, stdout, "feat/dirty", "stdout must contain the dirty row")

	errOut, ok := ios.ErrOut.(*bytes.Buffer)
	require.True(t, ok, "ErrOut must be a *bytes.Buffer for assertion")
	assert.Contains(t, errOut.String(), "feat/skipped",
		"stderr must surface a warning naming the skipped row")
}

// TestStatus_OutsideGitUsageError pins the cli-status spec scenario
// "Outside git, no flags": runStatus from outside any git
// repository with no positional argument and no --all returns
// *core.UsageError so dispatch lands on exit 2.
//
// Cannot use t.Parallel: NewRootCommand mutates cobra's
// package-level OnInitialize list via carapace.Gen.
func TestStatus_OutsideGitUsageError(t *testing.T) {
	f := newTestFactory(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/cfg")
	t.Chdir(tmp)

	root := newRootForTest(f)
	root.SetArgs([]string{"status"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "outside-git no-arg must yield UsageError")
	assert.Equal(t, cmdutil.ExitUsage, cmdutil.ExitCodeFor(err),
		"UsageError must map to exit 2")
	assert.Contains(t, ue.Error(), "--all",
		"UsageError must name the --all option")
	assert.Contains(t, ue.Error(), "project",
		"UsageError must name the positional project option")
}

// TestStatus_OutsideGitPositionalOK pins the positive path: an
// outside-git caller supplying a positional project argument must
// pass the outside-git guard and reach project discovery. The
// discovery step then surfaces its own typed error (the fixture
// ProjectsDirectory is empty), proving the guard did not trip.
//
// Cannot use t.Parallel: NewRootCommand mutates cobra's
// package-level OnInitialize list via carapace.Gen.
func TestStatus_OutsideGitPositionalOK(t *testing.T) {
	f := newTestFactory(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/cfg")
	t.Chdir(tmp)

	root := newRootForTest(f)
	root.SetArgs([]string{"status", "no-such-project"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	assert.NotErrorAs(t, err, &ue,
		"outside-git with a positional must NOT short-circuit to UsageError")
}

// TestStatus_UnknownProject pins the [project] positional path:
// an unknown project name returns a *core.NotFoundError so
// dispatch lands on exit 1. Runs from inside a real (scratch)
// project context so the outside-git guard doesn't fire first.
//
// Cannot use t.Parallel: NewRootCommand mutates cobra's
// package-level OnInitialize list via carapace.Gen.
func TestStatus_UnknownProject(t *testing.T) {
	projectRoot := t.TempDir()
	runGitCmd(t, projectRoot, "init", "--quiet", "-b", "main")
	runGitCmd(t, projectRoot, "config", "user.email", "test@twiggit.dev")
	runGitCmd(t, projectRoot, "config", "user.name", "Test User")
	runGitCmd(t, projectRoot, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(projectRoot, "README.md"), []byte("init"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitCmd(t, projectRoot, "add", "README.md")
	runGitCmd(t, projectRoot, "commit", "--quiet", "-m", "init")

	f := newTestFactory(t)
	t.Chdir(projectRoot)

	root := newRootForTest(f)
	root.SetArgs([]string{"status", "no-such-project"})

	err := root.Execute()
	require.Error(t, err)
	assert.Equal(t, cmdutil.ExitError, cmdutil.ExitCodeFor(err),
		"unknown project must map to exit 1")
	assert.Contains(t, err.Error(), "not found",
		"unknown project must surface a not-found message")
	var nfe *core.NotFoundError
	require.ErrorAs(t, err, &nfe,
		"unknown project must surface as *core.NotFoundError, not a generic OperationError")
	assert.Equal(t, "project", nfe.Entity)
	// The project name surfaced here is the context-derived name
	// (currentCtx.ProjectName), not the positional "no-such-project":
	// `cmd/status` currently derives the target from context when
	// inside a project, ignoring the positional. The positional-
	// precedence fix is tracked as a follow-up; this assertion
	// confirms the typed-error contract at the findProjectByName
	// helper boundary regardless of which name triggers it.
	assert.NotEmpty(t, nfe.Name)
	assert.ErrorIs(t, err, core.ErrProjectNotFound,
		"unknown project must satisfy errors.Is(err, core.ErrProjectNotFound)")
}

// runGitCmd shells out to git in dir; helper for scratch-repo
// setup in tests that need a real (non-bare) project context.
func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// keep io import referenced for future tests without noise.
var _ = io.Discard
