package core

import (
	"errors"
	"os"
	"path/filepath"
)

// RepoDir is a minimal git-repository descriptor: its base name and
// absolute filesystem path.
type RepoDir struct {
	Name string
	Path string
}

// IsMainRepo reports whether path is the root of a git worktree (a
// repository whose .git is a directory and not a worktree pointer).
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
// directory that IsMainRepo recognises, and returns that path. Returns
// the empty string when no main repository ancestor exists.
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
