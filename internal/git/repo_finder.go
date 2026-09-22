package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

// RepoFinder discovers git repositories within a directory tree.
type RepoFinder struct {
	goGit application.GoGitClient
}

// NewRepoFinder constructs a RepoFinder. goGit may be nil; when nil
// the finder accepts every subdirectory as a candidate without
// validating it against go-git.
func NewRepoFinder(goGit application.GoGitClient) *RepoFinder {
	return &RepoFinder{goGit: goGit}
}

// FindGitRepositories returns the immediate-subdirectory git repos under
// dir. If goGit is set, candidates that fail ValidateRepository are
// filtered out. The returned slice is defensively copied so callers may
// mutate it without affecting subsequent calls.
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

// Compile-time assertion: RepoFinder satisfies application.RepoLocator.
var _ application.RepoLocator = (*RepoFinder)(nil)
