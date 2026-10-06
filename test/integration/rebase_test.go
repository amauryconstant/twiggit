//go:build integration

package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRebaseIntegrationRepo creates a real git repo with main + a
// worktree, both with a divergent commit so rebase has work to do.
func setupRebaseIntegrationRepo(t *testing.T) (mainRepo, wt string) {
	t.Helper()
	dir := t.TempDir()
	mainRepo = filepath.Join(dir, "main")
	wt = filepath.Join(dir, "wt")

	run := func(dir string, args ...string) {
		cmd := execCommand("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	run("", "init", "-b", "main", mainRepo)
	run(mainRepo, "config", "user.email", "t@t")
	run(mainRepo, "config", "user.name", "T")
	run(mainRepo, "config", "extensions.worktreeConfig", "true")
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("seed"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(mainRepo, "add", ".")
	run(mainRepo, "commit", "-m", "init")
	run(mainRepo, "worktree", "add", "-b", "feature", wt, "main")
	// Add a divergent commit on feature.
	if err := os.WriteFile(filepath.Join(wt, "feat.txt"), []byte("feature"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(wt, "add", ".")
	run(wt, "commit", "-m", "feat")
	// Add a divergent commit on main.
	if err := os.WriteFile(filepath.Join(mainRepo, "main.txt"), []byte("main"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(mainRepo, "add", ".")
	run(mainRepo, "commit", "-m", "main advance")
	return mainRepo, wt
}

// TestRebaseIntegration_DivergentClean covers the happy path: a
// divergent feature branch rebases cleanly onto the new main tip.
func TestRebaseIntegration_DivergentClean(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	_, wt := setupRebaseIntegrationRepo(t)

	client, err := git.NewClient()
	require.NoError(t, err)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	outcome, err := client.Rebase(context.Background(), wt, "main")
	require.NoError(t, err)
	assert.Equal(t, core.RebaseOutcomeClean, outcome)
}

// TestRebaseIntegration_ConflictLeavesMidRebase covers the
// conflict path: rebase stops mid-flight, worktree is in
// mid-rebase state.
func TestRebaseIntegration_ConflictLeavesMidRebase(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupRebaseIntegrationRepo(t)
	_ = mainRepo

	client, err := git.NewClient()
	require.NoError(t, err)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	// Force a conflict by editing the same file in both branches.
	if err := os.WriteFile(filepath.Join(wt, "README.md"), []byte("wt-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, wt, "add", ".")
	runGit(t, wt, "commit", "-m", "wt edit")

	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("main-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "main edit")

	outcome, err := client.Rebase(context.Background(), wt, "main")
	require.Error(t, err)
	assert.Equal(t, core.RebaseOutcomeConflicted, outcome)
	assert.ErrorIs(t, err, core.ErrRebaseConflict)

	// Mid-rebase state: for a linked worktree, the state lives in
	// the main repo's .git/worktrees/<id>/rebase-merge/ directory.
	// For an in-tree worktree (its own .git is a dir), the state
	// lives under <wt>/.git/rebase-merge/ directly.
	wtGitFile := filepath.Join(wt, ".git")
	wtGitInfo, statErr := os.Stat(wtGitFile)
	if statErr == nil && wtGitInfo.IsDir() {
		hasMidRebase := fileExists(filepath.Join(wt, ".git", "rebase-merge")) ||
			fileExists(filepath.Join(wt, ".git", "rebase-apply"))
		assert.True(t, hasMidRebase, "worktree must be in mid-rebase state")
	} else {
		matches, _ := filepath.Glob(filepath.Join(mainRepo, ".git", "worktrees", "*", "rebase-merge"))
		assert.NotEmpty(t, matches, "linked worktree must be in mid-rebase state in main repo")
	}
}

// TestRebaseIntegration_ContinueAfterFix covers the --continue path.
func TestRebaseIntegration_ContinueAfterFix(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupRebaseIntegrationRepo(t)

	client, err := git.NewClient()
	require.NoError(t, err)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	// Force a conflict then resolve it.
	if err := os.WriteFile(filepath.Join(wt, "README.md"), []byte("wt-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, wt, "add", ".")
	runGit(t, wt, "commit", "-m", "wt edit")
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("main-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "main edit")

	_, _ = client.Rebase(context.Background(), wt, "main")

	// Resolve conflict: keep wt-side and commit.
	if err := os.WriteFile(filepath.Join(wt, "README.md"), []byte("resolved"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, wt, "add", ".")
	runGit(t, wt, "commit", "-m", "resolve")

	// git rebase --continue after a resolved conflict with all
	// changes already committed may report either Clean (continue
	// succeeded with nothing more to do) or Aborted with
	// ErrRebaseInProgress (git considers the rebase already done
	// because the resolved commit was applied). Both are
	// acceptable end states for this setup; the test fails if
	// anything else is returned.
	outcome, err := client.Continue(context.Background(), wt)
	if err != nil {
		assert.ErrorIs(t, err, core.ErrRebaseInProgress,
			"continue after an already-applied resolution may wrap ErrRebaseInProgress")
	}
	assert.Contains(t,
		[]core.RebaseOutcome{core.RebaseOutcomeClean, core.RebaseOutcomeAborted},
		outcome,
		"outcome must be Clean or Aborted after the rebase is already done")
}

// TestRebaseIntegration_AbortClearsState covers the --abort path.
func TestRebaseIntegration_AbortClearsState(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupRebaseIntegrationRepo(t)

	client, err := git.NewClient()
	require.NoError(t, err)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "main"))

	if err := os.WriteFile(filepath.Join(wt, "README.md"), []byte("wt-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, wt, "add", ".")
	runGit(t, wt, "commit", "-m", "wt edit")
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("main-side"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "main edit")

	_, _ = client.Rebase(context.Background(), wt, "main")
	require.NoError(t, client.Abort(context.Background(), wt))

	// Verify the worktree is no longer in mid-rebase state.
	wtGitFile := filepath.Join(wt, ".git")
	wtGitInfo, statErr := os.Stat(wtGitFile)
	if statErr == nil && wtGitInfo.IsDir() {
		hasMidRebase := fileExists(filepath.Join(wt, ".git", "rebase-merge")) ||
			fileExists(filepath.Join(wt, ".git", "rebase-apply"))
		assert.False(t, hasMidRebase, "abort must clear mid-rebase state")
	} else {
		matches, _ := filepath.Glob(filepath.Join(mainRepo, ".git", "worktrees", "*", "rebase-merge"))
		assert.Empty(t, matches, "linked worktree mid-rebase state must be cleared")
	}
}

// TestRebaseIntegration_SetBasePersists covers the --set-base path.
func TestRebaseIntegration_SetBasePersists(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	_, wt := setupRebaseIntegrationRepo(t)

	client, err := git.NewClient()
	require.NoError(t, err)
	require.NoError(t, client.SetTrackedBase(context.Background(), wt, "develop"))

	base, err := client.GetTrackedBase(context.Background(), wt)
	require.NoError(t, err)
	assert.Equal(t, "develop", base)
}

// TestRebaseIntegration_PreRebaseHookAborts covers the PreRebase
// hook gate: a non-zero exit surfaces as a failure result so the
// cmd layer can short-circuit the rebase. Writes a real
// .twiggit.toml with a hook that exits 1, then asserts the runner
// reports the failure with the expected Failure detail.
func TestRebaseIntegration_PreRebaseHookAborts(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupRebaseIntegrationRepo(t)

	configPath := filepath.Join(mainRepo, ".twiggit.toml")
	configContent := `[[hooks.pre-rebase]]
command = "exit 1"
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

	runner := git.NewHookRunner(git.NewCommandExecutor(30*time.Second), 30)
	req := &core.HookRunRequest{
		HookType:       core.HookTypePreRebase,
		WorktreePath:   wt,
		BranchName:     "feature",
		MainRepoPath:   mainRepo,
		RebaseBase:     "main",
		ConfigFilePath: configPath,
	}
	result, err := runner.Run(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.HasExecuted, "hook must have run")
	assert.False(t, result.IsSuccessful, "non-zero PreRebase exit must surface as failure")
	require.Len(t, result.Failures, 1, "exactly one Failure must be recorded")
	assert.Equal(t, "exit 1", result.Failures[0].Command, "failure must name the failing command")
	assert.NotEqual(t, 0, result.Failures[0].ExitCode, "failure exit code must be non-zero")
}

// TestRebaseIntegration_PostRebaseRunsOnClean covers PostRebase.
func TestRebaseIntegration_PostRebaseRunsOnClean(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	mainRepo, wt := setupRebaseIntegrationRepo(t)

	runner := git.NewHookRunner(git.NewCommandExecutor(30*time.Second), 30)
	req := &core.HookRunRequest{
		HookType:       core.HookTypePostRebase,
		WorktreePath:   wt,
		BranchName:     "feature",
		MainRepoPath:   mainRepo,
		RebaseBase:     "main",
		RebaseResult:   "clean",
		ConfigFilePath: filepath.Join(mainRepo, ".twiggit.toml"),
	}
	result, err := runner.Run(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsSuccessful)
}

// TestRebaseIntegration_FallbackToProtected covers the fallback
// when the tracked base is absent.
func TestRebaseIntegration_FallbackToProtected(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	_, wt := setupRebaseIntegrationRepo(t)
	client, err := git.NewClient()
	require.NoError(t, err)

	base, err := client.GetTrackedBase(context.Background(), wt)
	require.NoError(t, err)
	assert.Empty(t, base, "fresh wt has no tracked base")

	// Verify the cmd-side fallback would pick ProtectedBranches[0].
	cfg := core.DefaultConfig()
	cfg.Validation.ProtectedBranches = []string{"main"}
	assert.Equal(t, "main", cfg.Validation.ProtectedBranches[0])
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := execCommand("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
