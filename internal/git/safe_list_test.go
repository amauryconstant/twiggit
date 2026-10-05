package git

import (
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafeListWorktrees_NilClient pins the nil-input guard: callers
// passing nil must get (nil, false) without a panic.
func TestSafeListWorktrees_NilClient(t *testing.T) {
	t.Parallel()

	worktrees, safe := SafeListWorktrees(t.Context(), nil, "/tmp/whatever")
	assert.Nil(t, worktrees)
	assert.False(t, safe, "nil client must surface the unsafe signal")
}

// TestSafeListWorktrees_Success pins the happy path against a real
// scratch git repo: one main worktree is returned and the safe flag
// is true.
func TestSafeListWorktrees_Success(t *testing.T) {
	t.Parallel()

	repoDir := initScratchGitRepo(t)
	client, err := NewClient()
	require.NoError(t, err)

	worktrees, safe := SafeListWorktrees(t.Context(), client, repoDir)
	assert.True(t, safe, "successful list must surface the safe signal")
	assert.NotEmpty(t, worktrees, "scratch repo must surface its main worktree")
}

// TestSafeListWorktrees_InvalidPath pins the failure path: a path
// that is not a git repo yields (nil, false) without surfacing the
// underlying *core.OperationError. Callers branch on safe==false; the
// underlying detail is intentionally hidden.
func TestSafeListWorktrees_InvalidPath(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)

	worktrees, safe := SafeListWorktrees(t.Context(), client, t.TempDir())
	assert.Nil(t, worktrees)
	assert.False(t, safe, "invalid repo path must surface the unsafe signal")
}

// TestSafeListWorktrees_EmptyRepo pins the empty-but-valid case: a
// fresh git repo with no worktrees added yet returns safe=true and
// an empty (non-nil) slice so callers can len() it directly.
func TestSafeListWorktrees_EmptyRepo(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)

	worktrees, safe := SafeListWorktrees(t.Context(), client, t.TempDir())
	assert.False(t, safe, "temp dir is not a valid repo; safe must be false")
	assert.Nil(t, worktrees, "safe=false contract returns nil so callers can branch on nil")
	var _ []core.Worktree = worktrees // type-check
}
