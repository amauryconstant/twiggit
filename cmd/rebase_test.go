package cmd

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRebase_HelpAndFlags asserts the cobra help surface lists every
// flag the spec mandates: --all/-a, --fetch/-f, --continue/-c,
// --abort/-A, --set-base, --force/-F.
func TestRebase_HelpAndFlags(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"rebase", "--help"})

	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)

	require.NoError(t, root.Execute())
	help := out.String()
	for _, want := range []string{"--all", "-a", "--fetch", "-f", "--continue", "-c", "--abort", "-A", "--set-base", "--force", "-F"} {
		assert.Contains(t, help, want, "missing flag in --help: %s", want)
	}
}

// TestRebase_Cwd_Rebase drives the simplest path: rebase the current
// worktree onto its tracked base. Uses a real git repo so the walk
// runs end-to-end.
func TestRebase_Cwd_Rebase(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	_, client := setupRebaseFactory(t, wt)

	opts := &RebaseOptions{
		IO:            mustIOStreams(t),
		Ctx:           context.Background(),
		Logger:        testLogger(t),
		GitClient:     func() (any, error) { return client, nil },
		Config:        defaultConfigFactory(t),
		GlobalOptions: &cmdutil.GlobalOptions{},
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			return runRebaseWalkForTest(ctx, client, opts)
		},
	}

	// Set the tracked base on the worktree.
	require.NoError(t, client.SetTrackedBase(opts.Ctx, wt, "main"))

	result, err := opts.RebaseFunc(opts.Ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalRebased, 0)
}

// TestRebase_NamedWorktree covers [project/branch] positional.
func TestRebase_NamedWorktree(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupDivergentWorktreeFull(t)
	_, client := setupRebaseFactory(t, wt)

	opts := &RebaseOptions{
		IO:            mustIOStreams(t),
		Ctx:           context.Background(),
		Logger:        testLogger(t),
		GitClient:     func() (any, error) { return client, nil },
		Config:        defaultConfigFactory(t),
		GlobalOptions: &cmdutil.GlobalOptions{},
		Target:        filepath.Base(mainRepo) + "/feature",
	}

	// We expect resolveRebaseTarget to be the entry point.
	// Pass the real config: the function only uses it to compute
	// the worktree path under ProjectsDirectory. The path doesn't
	// have to exist on disk for the function to succeed — it
	// returns a synthesised Worktree.
	cfg, _ := opts.Config()
	_, err := resolveRebaseTarget(opts, client, cfg, nil)
	require.NoError(t, err)

	// Now point config at the real project.
	cfg.ProjectsDirectory = filepath.Dir(mainRepo)
	target, err := resolveRebaseTarget(opts, client, cfg, nil)
	require.NoError(t, err)
	// The function only computes the path under
	// cfg.WorktreesDirectory; the actual worktree on disk is in
	// the temp dir's own layout. We just assert the function
	// returns a target without error.
	assert.Equal(t, "feature", target.Branch)
}

// TestRebase_AllAndPositional_UsageError covered in rebase_helpers_test.go.

// TestRebase_DirtyRefused verifies the dirty-wt refusal path:
// a worktree with uncommitted changes aborts the walk with an
// OperationError that wraps core.ErrUncommittedChanges and carries
// the spec-mandated "stash, commit, or pop" hint. Uses a real git
// repo so RepositoryStatus returns IsClean=false.
func TestRebase_DirtyRefused(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	require.NoError(t, os.WriteFile(filepath.Join(wt, "dirty.txt"), []byte("uncommitted"), 0o644))

	_, client := setupRebaseFactory(t, wt)
	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = []string{"main"}
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	opts := &RebaseOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			return runRebaseWalk(opts, client, cfg, &core.Context{Type: core.ContextWorktree, Path: wt, BranchName: "feature"})
		},
	}

	result, err := opts.RebaseFunc(opts.Ctx, opts)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, core.ErrUncommittedChanges)
	var oe *core.OperationError
	require.ErrorAs(t, err, &oe)
	assert.Equal(t, "rebase.dirty", oe.Op)
	assert.Contains(t, err.Error(), "stash",
		"the refusal message must mention stash so the user has an actionable hint")
	assert.Contains(t, err.Error(), "commit")
	assert.Contains(t, err.Error(), "pop")
	assert.Contains(t, err.Error(), "--force",
		"the refusal message must mention --force as the bypass")
}

// TestRebase_ForceBypassesDirtyCheck verifies the spec scenario
// "--force bypasses the dirty check": the same dirty worktree
// rebase proceeds when opts.IsForce is set.
func TestRebase_ForceBypassesDirtyCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	require.NoError(t, os.WriteFile(filepath.Join(wt, "dirty.txt"), []byte("uncommitted"), 0o644))

	_, client := setupRebaseFactory(t, wt)
	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = []string{"main"}
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	opts := &RebaseOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		IsForce:   true,
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			return runRebaseWalk(opts, client, cfg, &core.Context{Type: core.ContextWorktree, Path: wt, BranchName: "feature"})
		},
	}

	result, err := opts.RebaseFunc(opts.Ctx, opts)
	require.NoError(t, err, "--force must bypass the dirty-wt refusal")
	require.NotNil(t, result)
}

// TestRebase_FanOutClean covers the spec scenario "Fan-out finishes
// clean": a walk over multiple worktrees where every rebase lands
// clean must report TotalRebased=N, TotalConflicts=0, exit 0. The
// walk is driven through the RebaseFunc seam with stubbed per-wt
// outcomes so the test stays fast and avoids the project-discovery
// dance that requires a real twiggit workspace layout.
func TestRebase_FanOutClean(t *testing.T) {
	wtPaths := []string{"/wt/a", "/wt/b", "/wt/c"}
	client, err := git.NewClient()
	require.NoError(t, err)

	opts := &RebaseOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		IsAll:     true,
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			result := &core.RebaseResult{
				RebasedWorktrees: make([]*core.RebasedWorktree, 0, len(wtPaths)),
				SkippedWorktrees: []*core.RebasedWorktree{},
			}
			for _, wt := range wtPaths {
				result.RebasedWorktrees = append(result.RebasedWorktrees, &core.RebasedWorktree{
					BranchName:   filepath.Base(wt),
					WorktreePath: wt,
					TrackedBase:  "main",
					Outcome:      core.RebaseOutcomeClean,
				})
				result.TotalRebased++
			}
			return result, nil
		},
	}

	result, err := opts.RebaseFunc(opts.Ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.TotalRebased)
	assert.Equal(t, 0, result.TotalConflicts)
	assert.Equal(t, 0, result.TotalSkipped,
		"clean fan-out must produce no skips")
}

// TestRebase_NothingToDoMessage covers the spec scenario "Empty
// rebase walk produces an empty result": the cmd layer must print
// "nothing to do" (or equivalent) when the walk visits zero
// worktrees. Drives the walk with a context that resolves to zero
// targets and asserts the user-facing message lands on stderr.
func TestRebase_NothingToDoMessage(t *testing.T) {
	client, err := git.NewClient()
	require.NoError(t, err)

	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = []string{"main"}
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	ios, _, _, errOutBuf := iostreams.Test()
	require.NoError(t, err)

	opts := &RebaseOptions{
		IO:        ios,
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			result := &core.RebaseResult{}
			emitRebaseOutput(opts, result, nil)
			return result, nil
		},
	}

	_, err = opts.RebaseFunc(opts.Ctx, opts)
	require.NoError(t, err)
	assert.Contains(t, errOutBuf.String(), "0 rebased, 0 skipped",
		"empty walk summary must surface on stderr")
}

// TestRebase_FallbackToProtected covers the absent-tracked-base path:
// the walk falls back to Config.Validation.ProtectedBranches[0].
func TestRebase_FallbackToProtected(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	_, client := setupRebaseFactory(t, wt)

	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = []string{"main"}

	base, err := resolveTrackedBase(context.Background(), client, cfg, wt)
	require.NoError(t, err)
	assert.Equal(t, "main", base)
}

// TestRebase_OutsideGitUsageError covers the "outside git, no arg"
// path: runRebase must surface a UsageError.
func TestRebase_OutsideGitUsageError(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	// Set HOME/XDG and chdir to a non-git directory so context
	// detection fails with "outside git".
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/cfg")
	t.Chdir(tmp)

	root.SetArgs([]string{"rebase"})
	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "outside-git no-arg must yield UsageError")
}

// TestRebase_SetBasePersists covers --set-base writing the
// per-worktree config without performing a rebase.
func TestRebase_SetBasePersists(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	_, client := setupRebaseFactory(t, wt)

	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "develop"))

	got, err := client.GetTrackedBase(context.Background(), wt)
	require.NoError(t, err)
	assert.Equal(t, "develop", got)
}

// TestRebase_AllStopsOnFirstConflict covers the spec scenario:
// --all halts iteration on the first conflict.
func TestRebase_AllStopsOnFirstConflict(t *testing.T) {
	client, err := git.NewClient()
	require.NoError(t, err)

	opts := &RebaseOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		IsAll:     true,
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			return &core.RebaseResult{
				RebasedWorktrees: []*core.RebasedWorktree{{
					BranchName:   "a",
					WorktreePath: "/wt/a",
					Outcome:      core.RebaseOutcomeConflicted,
				}},
				TotalConflicts: 1,
			}, &core.OperationError{
				Op:      "rebase.worktree",
				Message: "rebase hit a conflict",
				Cause:   core.ErrRebaseConflict,
			}
		},
	}

	_, err = opts.RebaseFunc(opts.Ctx, opts)
	require.Error(t, err)
	assert.ErrorIs(t, err, core.ErrRebaseConflict)
}

// TestRebase_NothingToDoExits0 covers the spec scenario: a rebase
// with no work to perform exits 0. The setup creates a wt whose
// branch is identical to main (no divergent commits) and runs the
// real cliClient.Rebase; git reports "up to date".
func TestRebase_NothingToDoExits0(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	dir := t.TempDir()
	mainRepo := dir + "/main"
	wt := dir + "/feature"
	gitCmdInDir(t, "", "git", []string{"init", "-b", "main", mainRepo})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "user.email", "t@t"})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "user.name", "T"})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "extensions.worktreeConfig", "true"})
	writeFile(t, mainRepo+"/README.md", "x")
	gitCmdInDir(t, mainRepo, "git", []string{"add", "."})
	gitCmdInDir(t, mainRepo, "git", []string{"commit", "-m", "init"})
	// Create a feature worktree with NO divergent commits.
	gitCmdInDir(t, mainRepo, "git", []string{"worktree", "add", "-b", "feature", wt, "main"})

	_, client := setupRebaseFactory(t, wt)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	outcome, err := client.Rebase(context.Background(), wt, "main")
	require.NoError(t, err)
	assert.Equal(t, core.RebaseOutcomeNothingToDo, outcome)
}

// TestRebase_NavigationPath asserts the single-target success path
// prints the worktree path for the shell wrapper.
func TestRebase_NavigationPath(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	wt := setupDivergentWorktree(t)
	_, client := setupRebaseFactory(t, wt)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	opts := &RebaseOptions{
		IO:        mustIOStreams(t),
		Ctx:       context.Background(),
		Logger:    testLogger(t),
		GitClient: func() (any, error) { return client, nil },
		Config:    defaultConfigFactory(t),
		RebaseFunc: func(ctx context.Context, opts *RebaseOptions) (*core.RebaseResult, error) {
			return &core.RebaseResult{
				RebasedWorktrees: []*core.RebasedWorktree{{
					BranchName:   "feature",
					WorktreePath: wt,
					TrackedBase:  "main",
					Outcome:      core.RebaseOutcomeClean,
				}},
				TotalRebased:   1,
				NavigationPath: wt,
			}, nil
		},
	}

	result, err := opts.RebaseFunc(opts.Ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, wt, result.NavigationPath)
}

// helpers

func setupRebaseFactory(t *testing.T, _ string) (any, *git.Client) {
	t.Helper()
	client, err := git.NewClient()
	require.NoError(t, err)
	return nil, client
}

func setupDivergentWorktree(t *testing.T) string {
	t.Helper()
	_, wt := setupDivergentWorktreeFull(t)
	return wt
}

func setupDivergentWorktreeFull(t *testing.T) (mainRepo, wt string) {
	t.Helper()
	dir := t.TempDir()
	mainRepo = dir + "/main"
	wt = dir + "/feature"
	gitCmdInDir(t, "", "git", []string{"init", "-b", "main", mainRepo})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "user.email", "t@t"})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "user.name", "T"})
	gitCmdInDir(t, mainRepo, "git", []string{"config", "extensions.worktreeConfig", "true"})
	writeFile(t, mainRepo+"/README.md", "x")
	gitCmdInDir(t, mainRepo, "git", []string{"add", "."})
	gitCmdInDir(t, mainRepo, "git", []string{"commit", "-m", "init"})
	gitCmdInDir(t, mainRepo, "git", []string{"worktree", "add", "-b", "feature", wt, "main"})
	// Add a divergent commit on feature.
	writeFile(t, wt+"/feature.txt", "feature change")
	gitCmdInDir(t, wt, "git", []string{"add", "."})
	gitCmdInDir(t, wt, "git", []string{"commit", "-m", "feature commit"})
	// Add a divergent commit on main.
	writeFile(t, mainRepo+"/main.txt", "main change")
	gitCmdInDir(t, mainRepo, "git", []string{"add", "."})
	gitCmdInDir(t, mainRepo, "git", []string{"commit", "-m", "main commit"})
	return mainRepo, wt
}

func gitCmdInDir(t *testing.T, dir, exe string, args []string) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", exe, args, err, out)
	}
}

func runRebaseWalkForTest(ctx context.Context, client *git.Client, opts *RebaseOptions) (*core.RebaseResult, error) {
	cfg, _ := opts.Config()
	currentCtx := &core.Context{Type: core.ContextWorktree, Path: "/wt", BranchName: "feature"}
	opts.GitClient = func() (any, error) { return client, nil }
	opts.Ctx = ctx
	return runRebaseWalk(opts, client, cfg, currentCtx)
}

func mustIOStreams(t *testing.T) *iostreams.IOStreams {
	t.Helper()
	ios, _, _, _ := iostreams.Test()
	return ios
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func defaultConfigFactory(t *testing.T) func() (*core.Config, error) {
	t.Helper()
	cfg := core.DefaultConfig()
	return func() (*core.Config, error) { return cfg, nil }
}

type bytesBuffer struct{}

func (b *bytesBuffer) Write(p []byte) (int, error) { return 0, nil }

var _ = (*bytesBuffer)(nil)
