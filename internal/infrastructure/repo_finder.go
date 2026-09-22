package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

type RepoFinder struct {
	goGit application.GoGitClient
}

func NewRepoFinder(goGit application.GoGitClient) *RepoFinder {
	return &RepoFinder{goGit: goGit}
}

func (f *RepoFinder) FindGitRepositories(dir string) ([]core.GitDir, error) {
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []core.GitDir{}, nil
		}
		return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	repos := make([]core.GitDir, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(dir, entry.Name())
		if f.goGit != nil {
			if err := f.goGit.ValidateRepository(candidate); err != nil {
				continue
			}
		}
		repos = append(repos, core.GitDir{
			Name: entry.Name(),
			Path: candidate,
		})
	}

	return slices.Clone(repos), nil
}

var _ application.RepoLocator = (*RepoFinder)(nil)
