package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeInfo_Structure(t *testing.T) {
	t.Run("zero value has no modification flag set", func(t *testing.T) {
		is := assert.New(t)
		wt := WorktreeInfo{}
		is.False(wt.IsModified)
		is.False(wt.IsDetached)
		is.Empty(wt.Path)
		is.Empty(wt.Branch)
		is.Empty(wt.Commit)
	})

	t.Run("populated fields round-trip", func(t *testing.T) {
		is := assert.New(t)
		wt := WorktreeInfo{
			Path:       "/abs/path",
			Branch:     "feature/x",
			Commit:     "abcdef0123456789",
			IsDetached: true,
			IsModified: true,
		}
		is.Equal("/abs/path", wt.Path)
		is.Equal("feature/x", wt.Branch)
		is.Equal("abcdef0123456789", wt.Commit)
		is.True(wt.IsDetached)
		is.True(wt.IsModified)
	})
}

func TestWorktreeInfo_IsModifiedBooleanSemantics(t *testing.T) {
	tests := []struct {
		name     string
		modified bool
	}{
		{"unmodified worktree", false},
		{"modified worktree", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			wt := WorktreeInfo{IsModified: tt.modified}
			is.Equal(tt.modified, wt.IsModified)
		})
	}
}

func TestWorktreeInfo_UsedInGitRepositoryList(t *testing.T) {
	is := assert.New(t)
	require := require.New(t)

	repo := GitRepository{
		Path:   "/repos/example",
		Worktrees: []WorktreeInfo{
			{Path: "/repos/example/main", Branch: "main", Commit: "abc1234", IsModified: false},
			{Path: "/repos/example/feat", Branch: "feature/x", Commit: "def5678", IsModified: true},
		},
	}
	require.Len(repo.Worktrees, 2)
	is.False(repo.Worktrees[0].IsModified)
	is.True(repo.Worktrees[1].IsModified)
}
