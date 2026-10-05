package cmd

import (
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsMainWorktree pins the predicate that powers filterNonMain:
// only the worktree whose path equals repoPath is treated as the
// main checkout.
func TestIsMainWorktree(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		wt       core.Worktree
		repoPath string
		want     bool
	}{
		{
			name:     "matching path is main",
			wt:       core.Worktree{Path: "/projects/foo"},
			repoPath: "/projects/foo",
			want:     true,
		},
		{
			name:     "non-matching path is non-main",
			wt:       core.Worktree{Path: "/projects/foo/feat-1"},
			repoPath: "/projects/foo",
			want:     false,
		},
		{
			name:     "empty path is non-main",
			wt:       core.Worktree{Path: ""},
			repoPath: "/projects/foo",
			want:     false,
		},
		{
			name:     "empty repoPath makes every worktree non-main",
			wt:       core.Worktree{Path: "/projects/foo"},
			repoPath: "",
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isMainWorktree(tc.wt, tc.repoPath)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestFilterNonMain_ExcludesMainCheckout covers the filterNonMain
// contract: every input worktree whose path matches repoPath is
// dropped from the output; the rest are returned in input order.
func TestFilterNonMain_ExcludesMainCheckout(t *testing.T) {
	t.Parallel()

	repoPath := "/projects/foo"
	worktrees := []core.Worktree{
		{Branch: "main", Path: repoPath},
		{Branch: "feat-1", Path: "/projects/foo/feat-1"},
		{Branch: "feat-2", Path: "/projects/foo/feat-2"},
		{Branch: "detached", Path: "/projects/foo/detached"},
	}

	got := filterNonMain(worktrees, repoPath)

	require.Len(t, got, 3, "main checkout must be excluded")
	for _, wt := range got {
		assert.NotEqual(t, repoPath, wt.Path,
			"returned worktree must not be the main checkout")
	}
	assert.Equal(t, "feat-1", got[0].Branch, "input order must be preserved")
	assert.Equal(t, "feat-2", got[1].Branch)
	assert.Equal(t, "detached", got[2].Branch)
}

// TestFilterNonMain_EmptyAndAllMain covers the edge cases: empty
// input returns empty output; input containing only the main
// checkout returns empty output.
func TestFilterNonMain_EmptyAndAllMain(t *testing.T) {
	t.Parallel()

	t.Run("empty input returns empty output", func(t *testing.T) {
		t.Parallel()

		got := filterNonMain(nil, "/projects/foo")
		assert.Empty(t, got, "nil input must yield empty output")
	})

	t.Run("only-main input returns empty output", func(t *testing.T) {
		t.Parallel()

		got := filterNonMain([]core.Worktree{{Path: "/projects/foo"}}, "/projects/foo")
		assert.Empty(t, got, "all-main input must yield empty output")
	})
}

// TestFilterNonMain_ReturnIsFreshSlice pins the clone-or-not rule
// from parseWorktreeList: filterNonMain must NOT return a slice
// that aliases the input. Mutating the returned slice must not
// affect the input.
func TestFilterNonMain_ReturnIsFreshSlice(t *testing.T) {
	t.Parallel()

	repoPath := "/projects/foo"
	worktrees := []core.Worktree{
		{Branch: "main", Path: repoPath},
		{Branch: "feat-1", Path: "/projects/foo/feat-1"},
	}

	got := filterNonMain(worktrees, repoPath)
	require.Len(t, got, 1)

	got[0].Branch = "mutated"

	assert.Equal(t, "feat-1", worktrees[1].Branch,
		"mutating the returned slice must not leak into the input")
}
