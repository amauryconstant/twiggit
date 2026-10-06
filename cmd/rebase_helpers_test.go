package cmd

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRebaseHelpers_ResolveTrackedBase covers the fallback chain:
// per-worktree config → ProtectedBranches[0] → ErrBaseNotSet.
// Uses a real *git.Client against a real temp git repo (the
// per-worktree config CLI requires a valid worktree path).
func TestRebaseHelpers_ResolveTrackedBase(t *testing.T) {
	if testing.Short() {
		t.Skip("requires real git binary")
	}

	ctx := context.Background()
	cli, err := newGitClientForTest(t)
	require.NoError(t, err)

	wt := setupRealGitWorktree(t)

	t.Run("no tracked base falls back to protected branches", func(t *testing.T) {
		cfg := core.DefaultConfig()
		cfg.Validation.ProtectedBranches = []string{"main", "develop"}

		base, err := resolveTrackedBase(ctx, cli, cfg, wt)
		require.NoError(t, err)
		assert.Equal(t, "main", base, "must fall back to ProtectedBranches[0]")
	})

	t.Run("empty fallback list returns ErrBaseNotSet", func(t *testing.T) {
		cfg := core.DefaultConfig()
		cfg.Validation.ProtectedBranches = nil

		_, err := resolveTrackedBase(ctx, cli, cfg, wt)
		require.Error(t, err)
		var oe *core.OperationError
		require.ErrorAs(t, err, &oe)
		assert.Equal(t, "rebase.tracked-base", oe.Op)
	})
}

// TestRebaseHelpers_ResolveRebaseTargets covers [project/branch]
// parsing and the outside-git UsageError path.
func TestRebaseHelpers_ResolveRebaseTargets(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	t.Run("malformed positional returns UsageError", func(t *testing.T) {
		cli, err := newGitClientForTest(t)
		require.NoError(t, err)
		opts := &RebaseOptions{Target: "no-slash"}

		_, err = resolveRebaseTargets(context.Background(), cli, cfg, opts, nil)
		require.Error(t, err)
		var ue *core.UsageError
		require.ErrorAs(t, err, &ue)
	})

	t.Run("outside git with no argument returns UsageError", func(t *testing.T) {
		cli, err := newGitClientForTest(t)
		require.NoError(t, err)
		opts := &RebaseOptions{}

		_, err = resolveRebaseTargets(context.Background(), cli, cfg, opts, nil)
		require.Error(t, err)
		var ue *core.UsageError
		require.ErrorAs(t, err, &ue)
	})
}

// TestRebase_AllAndPositional covers the args validator that
// rejects --all combined with a positional argument.
func TestRebase_AllAndPositional(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"rebase", "--all", "myproject/feature"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "--all + positional must surface UsageError")
	assert.Contains(t, ue.Error(), "mutually exclusive")
}

// TestRebaseHelpers_DispatchRebaseHooks_NoConfig covers the no-op
// path: a worktree with no .twiggit.toml runs no hooks and reports
// abort=false. The harder abort=true case is covered by the runner
// tests in internal/git (which own the runner's hard-failure
// semantics; this test just pins the cmd-side wiring).
func TestRebaseHelpers_DispatchRebaseHooks_NoConfig(t *testing.T) {
	cfg := core.DefaultConfig()
	target := rebaseTarget{
		Project:      &core.ProjectInfo{Name: "p", Path: t.TempDir(), GitRepoPath: t.TempDir()},
		BranchName:   "feature",
		WorktreePath: t.TempDir(),
	}

	result, abort := dispatchRebaseHooks(context.Background(), nil, cfg, target, "main")
	assert.False(t, abort, "no config file means no hooks run, no abort")
	assert.NotNil(t, result)
}

// newGitClientForTest constructs a *git.Client using the default
// constructor. The internal executor is acceptable for the helper
// tests because the per-worktree config calls target throwaway
// temp dirs that git will report as absent.
func newGitClientForTest(t *testing.T) (*git.Client, error) {
	t.Helper()
	return git.NewClient()
}

// setupRealGitWorktree creates a real git repo with a main branch
// and a worktree, returning the worktree path. The repo is cleaned
// up via t.TempDir. Enables the worktreeConfig extension so the
// per-worktree `git config --worktree` calls succeed.
func setupRealGitWorktree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mainRepo := dir + "/main"
	worktree := dir + "/feature"
	import_exec(t, "", "git", []string{"init", "-b", "main", mainRepo})
	import_exec(t, mainRepo, "git", []string{"config", "user.email", "t@t"})
	import_exec(t, mainRepo, "git", []string{"config", "user.name", "T"})
	import_exec(t, mainRepo, "git", []string{"config", "extensions.worktreeConfig", "true"})
	writeFile(t, mainRepo+"/README.md", "x")
	import_exec(t, mainRepo, "git", []string{"add", "."})
	import_exec(t, mainRepo, "git", []string{"commit", "-m", "init"})
	import_exec(t, mainRepo, "git", []string{"worktree", "add", "-b", "feature", worktree, "main"})
	return worktree
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func import_exec(t *testing.T, dir, exe string, args []string) {
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
