package domain

import (
	"os"
	"path/filepath"
)

type GitDir struct {
	Name string
	Path string
}

func IsMainRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || !info.IsDir() {
		return false
	}

	gitdirPath := filepath.Join(gitPath, "gitdir")
	_, statErr := os.Stat(gitdirPath)
	return os.IsNotExist(statErr)
}

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
