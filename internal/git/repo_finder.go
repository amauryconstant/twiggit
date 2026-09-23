package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"twiggit/internal/core"
)

// RepoValidator is the minimal contract RepoFinder needs: a single
// ValidateRepository method. *Client satisfies it implicitly through
// its embedded *reader. Tests inject a mock to drive the validation
// branch without standing up a real go-git repo.
type RepoValidator interface {
	ValidateRepository(path string) error
}

// RepoFinder discovers git repositories within a directory tree.
type RepoFinder struct {
	goGit RepoValidator
}

// NewRepoFinder constructs a RepoFinder. goGit may be nil; when nil
// the finder accepts every subdirectory as a candidate without
// validating it against go-git.
func NewRepoFinder(goGit RepoValidator) *RepoFinder {
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
