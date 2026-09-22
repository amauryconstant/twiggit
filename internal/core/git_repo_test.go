package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMainRepo(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name: "directory with .git as directory and no gitdir file is main repo",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
			},
			expected: true,
		},
		{
			name: "directory with .git and gitdir file is worktree, not main repo",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				gitDir := filepath.Join(dir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(gitDir, "gitdir"), []byte("gitdir: /tmp/somewhere"), 0o644))
			},
			expected: false,
		},
		{
			name: "directory without .git is not a main repo",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				_ = dir
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			dir := t.TempDir()
			tt.setup(t, dir)
			is.Equal(tt.expected, IsMainRepo(dir))
		})
	}
}

func TestFindMainRepoByTraversal(t *testing.T) {
	t.Run("returns empty when no main repo in tree", func(t *testing.T) {
		is := assert.New(t)
		dir := t.TempDir()
		got := FindMainRepoByTraversal(dir)
		is.Empty(got)
	})

	t.Run("returns starting path when start is a main repo", func(t *testing.T) {
		is := assert.New(t)
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
		got := FindMainRepoByTraversal(dir)
		is.Equal(dir, got)
	})

	t.Run("traverses upward to find main repo", func(t *testing.T) {
		is := assert.New(t)
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
		nested := filepath.Join(root, "a", "b", "c")
		require.NoError(t, os.MkdirAll(nested, 0o755))
		got := FindMainRepoByTraversal(nested)
		is.Equal(root, got)
	})
}

func TestFindGitDirByTraversal(t *testing.T) {
	t.Run("returns false when no .git in tree", func(t *testing.T) {
		is := assert.New(t)
		dir := t.TempDir()
		path, ok := FindGitDirByTraversal(dir)
		is.False(ok)
		is.Empty(path)
	})

	t.Run("returns true and starting path when start has .git", func(t *testing.T) {
		is := assert.New(t)
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0o755))
		path, ok := FindGitDirByTraversal(dir)
		is.True(ok)
		is.Equal(dir, path)
	})

	t.Run("traverses upward to find .git directory", func(t *testing.T) {
		is := assert.New(t)
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
		nested := filepath.Join(root, "x", "y")
		require.NoError(t, os.MkdirAll(nested, 0o755))
		path, ok := FindGitDirByTraversal(nested)
		is.True(ok)
		is.Equal(root, path)
	})

	t.Run("does not return false-positive for filesystem root", func(t *testing.T) {
		is := assert.New(t)
		path, ok := FindGitDirByTraversal(os.TempDir())
		is.False(ok)
		is.Empty(path)
	})
}

func TestGitDir_Structure(t *testing.T) {
	is := assert.New(t)
	g := GitDir{Name: "myrepo", Path: "/path/to/myrepo"}
	is.Equal("myrepo", g.Name)
	is.Equal("/path/to/myrepo", g.Path)
}
