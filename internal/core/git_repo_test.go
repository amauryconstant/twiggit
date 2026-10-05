package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRepoDir_Structure(t *testing.T) {
	is := assert.New(t)
	g := RepoDir{Name: "myrepo", Path: "/path/to/myrepo"}
	is.Equal("myrepo", g.Name)
	is.Equal("/path/to/myrepo", g.Path)
}
