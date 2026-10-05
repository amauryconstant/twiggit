package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectInfoFromGitDir_Success pins the happy path: a real
// scratch git repo with one commit yields a ProjectInfo whose Name
// matches the repo basename, Path/GitRepoPath match the supplied
// directory, and at least one worktree entry surfaces.
func TestProjectInfoFromGitDir_Success(t *testing.T) {
	t.Parallel()

	repoDir := initScratchGitRepo(t)
	client, err := NewClient()
	require.NoError(t, err)

	info, err := ProjectInfoFromGitDir(t.Context(), client, repoDir)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, filepath.Base(repoDir), info.Name)
	assert.Equal(t, repoDir, info.Path)
	assert.Equal(t, repoDir, info.GitRepoPath, "single-commit repo IS the main repo")
	assert.False(t, info.IsBare)
	assert.Empty(t, info.DefaultBranch, "scratch repo has no main/master marker branch")
	assert.Empty(t, info.Remotes, "scratch repo has no remotes")
	assert.NotEmpty(t, info.Worktrees, "scratch repo must surface its main worktree")
}

// TestProjectInfoFromGitDir_NilClient pins the nil-input guard: a nil
// client yields an error rather than panicking on the first method
// call. Defensive contract per the cli-iostreams pattern of safe
// error returns over nil deref panics.
func TestProjectInfoFromGitDir_NilClient(t *testing.T) {
	t.Parallel()

	info, err := ProjectInfoFromGitDir(t.Context(), nil, "/tmp/whatever")
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "nil client")
}

// TestProjectInfoFromGitDir_EmptyGitDir pins the empty-path guard:
// callers that pass "" get an OperationError rather than a panic on
// filepath.Base or the subsequent stat.
func TestProjectInfoFromGitDir_EmptyGitDir(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)

	info, err := ProjectInfoFromGitDir(t.Context(), client, "")
	require.Error(t, err)
	assert.Nil(t, info)
	var oe *core.OperationError
	require.ErrorAs(t, err, &oe)
	assert.Equal(t, "project.info", oe.Op)
}

// TestProjectInfoFromGitDir_InvalidRepo pins the not-a-git-repo
// branch: the path exists but does not contain a .git directory, so
// ValidateRepository returns an error and the helper wraps it as
// *core.OperationError with Cause=core.ErrGitRepoNotFound via the
// embedded Op-prefix walk.
func TestProjectInfoFromGitDir_InvalidRepo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	client, err := NewClient()
	require.NoError(t, err)

	info, err := ProjectInfoFromGitDir(t.Context(), client, dir)
	require.Error(t, err)
	assert.Nil(t, info)
	assert.ErrorIs(t, err, core.ErrGitRepoNotFound,
		"invalid repo must walk to ErrGitRepoNotFound via OperationError.Op prefix")
}

// TestProjectInfoFromGitDir_MissingDefaultBranch pins the contract
// that DefaultBranch is empty (not panic, not zero-value-but-mistaken)
// when the underlying Repository() returns no main/master marker.
// The slice fields must remain non-nil defensive copies.
func TestProjectInfoFromGitDir_MissingDefaultBranch(t *testing.T) {
	t.Parallel()

	repoDir := initScratchGitRepo(t)
	client, err := NewClient()
	require.NoError(t, err)

	info, err := ProjectInfoFromGitDir(t.Context(), client, repoDir)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Empty(t, info.DefaultBranch)
	assert.NotNil(t, info.Branches, "Branches slice must be non-nil defensive copy")
	assert.NotNil(t, info.Worktrees, "Worktrees slice must be non-nil defensive copy")
	assert.NotNil(t, info.Remotes, "Remotes slice must be non-nil defensive copy (even if empty)")
}

// TestProjectInfoFromGitDir_MissingRemote pins the no-remotes branch:
// the returned Remotes slice is non-nil but empty, so cmd callers
// can branch on len(remotes) without a nil-check.
func TestProjectInfoFromGitDir_MissingRemote(t *testing.T) {
	t.Parallel()

	repoDir := initScratchGitRepo(t)
	client, err := NewClient()
	require.NoError(t, err)

	info, err := ProjectInfoFromGitDir(t.Context(), client, repoDir)
	require.NoError(t, err)
	require.NotNil(t, info)

	require.NotNil(t, info.Remotes)
	assert.Empty(t, info.Remotes, "scratch repo has no remotes")
}

// TestProjectInfoFromGitDir_DefensiveCopies pins the slice-isolation
// contract: mutating the Worktrees slice returned by the helper must
// not affect a subsequent call's slice. This is the safety net for
// any future refactor that drops the make+copy loop.
func TestProjectInfoFromGitDir_DefensiveCopies(t *testing.T) {
	t.Parallel()

	repoDir := initScratchGitRepo(t)
	client, err := NewClient()
	require.NoError(t, err)

	first, err := ProjectInfoFromGitDir(t.Context(), client, repoDir)
	require.NoError(t, err)
	require.NotEmpty(t, first.Worktrees)

	originalLen := len(first.Worktrees)
	first.Worktrees = append(first.Worktrees, &core.Worktree{Path: "/sentinel"})

	second, err := ProjectInfoFromGitDir(t.Context(), client, repoDir)
	require.NoError(t, err)
	assert.Len(t, second.Worktrees, originalLen,
		"mutating the first slice must not change the second slice")
}

// initScratchGitRepo builds a minimal git repo in a temp dir: one
// commit on a branch other than main/master so DefaultBranch
// resolution returns empty. Skips when git is unavailable so the test
// still runs in slim CI images.
func initScratchGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary unavailable")
	}

	repoDir := t.TempDir()
	cmds := [][]string{
		{"init", "--quiet", "--initial-branch=scratch", repoDir},
	}
	for _, args := range cmds {
		cmd := exec.Command("git", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Skipf("git %v failed: %v: %s", args, err, out)
		}
	}

	// Configure identity so the commit succeeds in CI.
	for _, args := range [][]string{
		{"-C", repoDir, "config", "user.email", "test@example.com"},
		{"-C", repoDir, "config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		_ = cmd.Run() // best-effort; CI may block
	}

	cmd := exec.Command("git", "-C", repoDir, "commit", "--allow-empty", "-m", "init")
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("git commit failed: %v: %s", err, out)
	}

	// Drop any default-branch naming hint the scratch repo inherits
	// so the test stays portable across git versions that auto-create
	// main on init.
	_ = os.RemoveAll(filepath.Join(repoDir, ".git", "refs", "heads", "main"))
	_ = os.RemoveAll(filepath.Join(repoDir, ".git", "refs", "heads", "master"))

	return repoDir
}

// TestMustProjectInfo_PanicsOnError pins the panic wrapper contract.
// Useful for init-time wiring where a missing repo is a programmer
// bug rather than a runtime condition.
func TestMustProjectInfo_PanicsOnError(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		_ = MustProjectInfo(nil, errors.New("boom"))
	}, "MustProjectInfo must panic on error to surface init-time bugs")
}

func TestMustProjectInfo_ReturnsOnSuccess(t *testing.T) {
	t.Parallel()

	want := &core.ProjectInfo{Name: "demo"}
	got := MustProjectInfo(want, nil)
	assert.Same(t, want, got)
}
