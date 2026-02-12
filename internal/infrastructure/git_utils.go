package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
)

// GitDir represents a directory containing a git repository
type GitDir struct {
	Name string
	Path string
}

// FindGitRepositories finds all git repositories in the specified directory
// Returns a list of directories that contain valid git repositories
func FindGitRepositories(dir string, gitService GoGitClient) ([]GitDir, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return []GitDir{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	gitDirs := make([]GitDir, 0, 10)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		gitDirPath := filepath.Join(dir, entry.Name())

		if gitService != nil {
			if err := gitService.ValidateRepository(gitDirPath); err != nil {
				continue
			}
		}

		gitDirs = append(gitDirs, GitDir{
			Name: entry.Name(),
			Path: gitDirPath,
		})
	}

	return gitDirs, nil
}

// FindMainRepoByTraversal traverses up the directory tree from the given path
// to find the main git repository (not a worktree). Returns the path if found,
// empty string otherwise.
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

// FindGitDirByTraversal traverses up the directory tree from the given path
// to find any git repository (.git directory). Returns the path if found,
// nil otherwise.
func FindGitDirByTraversal(startPath string) *string {
	currentPath := startPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return &currentPath
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	return nil
}

// IsMainRepo checks if the given path is a main repository (not a worktree)
// by verifying that .git is a directory and does not contain a gitdir file
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
