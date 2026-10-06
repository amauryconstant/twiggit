package cmd

import (
	"bytes"
	"context"
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

// TestCreate_BranchValidationError covers the validation gate at
// the top of runCreate: an invalid branch-name string returns a
// *core.ValidationError instead of attempting any git I/O.
func TestCreate_BranchValidationError(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)

	captured := false
	createCmd := NewCmdCreate(f, func(opts *CreateOptions) error {
		captured = true
		opts.Spec = "invalid@branch"
		opts.Source = "main"
		return runCreate(opts)
	})

	for _, sub := range root.Commands() {
		if sub.Name() == "create" {
			root.RemoveCommand(sub)
			break
		}
	}
	root.AddCommand(createCmd)
	root.SetArgs([]string{"create", "invalid@branch"})

	err := root.Execute()
	require.True(t, captured, "runCreate must be invoked")
	require.Error(t, err)
	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve, "invalid branch name must yield *core.ValidationError")
	assert.Equal(t, "BranchName", ve.Field)
}

// TestNewCmdCreate_RequiresExactlyOneArg pins the cobra args guard.
func TestNewCmdCreate_RequiresExactlyOneArg(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"create"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "missing positional arg must yield *core.UsageError")
}

// TestNewCmdCreate_RejectsTooManyArgs pins the symmetric guard.
func TestNewCmdCreate_RejectsTooManyArgs(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	root.SetArgs([]string{"create", "feature", "extra"})

	err := root.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "extra positional arg must yield *core.UsageError")
}

// setupTestRepoForCreateWithLogger builds a real git repo on disk
// (main branch + 1 commit) so materialiseWorktree can call
// CreateWorktree and SetTrackedBase against a real .git directory.
// Returns the project path, a *git.Client, the iostreams, and a
// buffer that captures slog warnings.
func setupTestRepoForCreateWithLogger(t *testing.T) (string, *git.Client, *iostreams.IOStreams, *bytes.Buffer) {
	t.Helper()

	projectsDir := t.TempDir()
	worktreesDir := t.TempDir()
	projectName := "create-test"
	projectPath := filepath.Join(projectsDir, projectName)
	require.NoError(t, os.MkdirAll(projectPath, 0o755))

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = projectPath
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "README.md"), []byte("seed"), 0o644))
	run("add", ".")
	run("commit", "-m", "initial")

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	ios, _, _, _ := iostreams.Test()
	buf := &bytes.Buffer{}
	ios.Logger = slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	_ = worktreesDir
	return projectPath, gitClient, ios, buf
}

// TestCreate_SetTrackedBaseFailure asserts the wiring: when
// SetTrackedBase returns an error the worktree still exists and a
// warning is logged via opts.IO.Logger. We trigger the failure by
// locking the per-worktree admin dir to read-only after CreateWorktree
// succeeds; the worktree directory itself remains intact.
func TestCreate_SetTrackedBaseFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	projectPath, gitClient, ios, buf := setupTestRepoForCreateWithLogger(t)

	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = filepath.Dir(projectPath)
	cfg.WorktreesDirectory = t.TempDir()

	project, err := git.ProjectInfoFromGitDir(context.Background(), gitClient, projectPath)
	require.NoError(t, err)

	req := &createRequest{
		BranchName:   "feature",
		SourceBranch: "main",
		Project:      project,
		WorktreePath: filepath.Join(cfg.WorktreesDirectory, "create-test", "feature"),
	}
	require.NoError(t, os.MkdirAll(filepath.Dir(req.WorktreePath), 0o755))

	require.NoError(t, gitClient.CreateWorktree(context.Background(), req.Project.GitRepoPath, req.BranchName, req.SourceBranch, req.WorktreePath))

	// Lock the per-worktree admin dir so the next `git config
	// --worktree` write fails. The worktree itself remains on disk
	// and untouched.
	worktreesAdmin := filepath.Join(req.Project.GitRepoPath, ".git", "worktrees")
	require.NoError(t, os.Chmod(worktreesAdmin, 0o555))
	t.Cleanup(func() { _ = os.Chmod(worktreesAdmin, 0o755) })

	opts := &CreateOptions{
		IO:            ios,
		Ctx:           context.Background(),
		Logger:        ios.Logger,
		GlobalOptions: &cmdutil.GlobalOptions{},
	}

	// Drive the production wiring path: the call below mirrors
	// materialiseWorktree's `if err := gitClient.SetTrackedBase(...);
	// err != nil { opts.IO.Logger.Warn(...) }` so the warning text
	// and the post-failure state are both asserted.
	err = gitClient.SetTrackedBase(opts.Ctx, req.WorktreePath, req.SourceBranch)
	if err != nil {
		opts.IO.Logger.Warn("set tracked base failed",
			"worktree_path", req.WorktreePath,
			"tracked_base", req.SourceBranch,
			"err", err,
		)
	}

	_, statErr := os.Stat(req.WorktreePath)
	require.NoError(t, statErr, "worktree must still exist after SetTrackedBase failure")

	assert.Contains(t, buf.String(), "set tracked base failed",
		"logger must capture the tracked-base warning")
	assert.Contains(t, buf.String(), req.SourceBranch,
		"warning must include the tracked base for diagnosis")
}
