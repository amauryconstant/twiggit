package git

import (
	"errors"
	"os"
	"path/filepath"
)

// IsMainRepo reports whether path is the root of a git repository
// (a directory whose .git is a directory and not a worktree pointer
// file). The function performs I/O (os.Stat) and therefore lives in
// the git adapter rather than the pure core package.
func IsMainRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || !info.IsDir() {
		return false
	}

	gitdirPath := filepath.Join(gitPath, "gitdir")
	_, statErr := os.Stat(gitdirPath)
	return errors.Is(statErr, os.ErrNotExist)
}

// FindMainRepoByTraversal walks up from startPath until it finds a
// directory that IsMainRepo recognises, and returns that path.
// Returns the empty string when no main repository ancestor exists.
// Lives in the git adapter because IsMainRepo is I/O.
func FindMainRepoByTraversal(startPath string) string {
	currentPath := startPath
	for {
		if IsMainRepo(currentPath) {
			return currentPath
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	return ""
}
