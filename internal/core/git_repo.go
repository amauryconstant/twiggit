package core

import (
	"os"
	"path/filepath"
)

// RepoDir is a minimal git-repository descriptor: its base name and
// absolute filesystem path.
type RepoDir struct {
	Name string
	Path string
}

// FindGitDirByTraversal walks up from startPath looking for the
// nearest directory containing a .git entry (file or directory).
// Returns the directory path and true on success, or "" and false
// when no such ancestor exists.
func FindGitDirByTraversal(startPath string) (string, bool) {
	currentPath := startPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath, true
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	return "", false
}
