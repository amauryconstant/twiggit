package helpers

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// RepoTestHelper provides functional repository management utilities
type RepoTestHelper struct {
	t           *testing.T
	baseDir     string
	repos       map[string]string
	commitCount int
	projectName string
	mu          sync.RWMutex
}

// NewRepoTestHelper creates a new RepoTestHelper instance
func NewRepoTestHelper(t *testing.T) *RepoTestHelper {
	t.Helper()
	return &RepoTestHelper{
		t:       t,
		baseDir: t.TempDir(),
		repos:   make(map[string]string),
	}
}

// WithProject sets the project name for functional composition
func (h *RepoTestHelper) WithProject(name string) *RepoTestHelper {
	h.projectName = name
	return h
}

// WithCommits sets the commit count for functional composition
func (h *RepoTestHelper) WithCommits(count int) *RepoTestHelper {
	h.commitCount = count
	return h
}

// SetupTestRepo creates a test repository with the given project name
func (h *RepoTestHelper) SetupTestRepo(projectName string) string {
	if projectName == "" {
		panic("project name cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if repo already exists
	if _, exists := h.repos[projectName]; exists {
		h.t.Fatalf("Repository %s already exists", projectName)
	}

	// Create repository directory
	repoPath := filepath.Join(h.baseDir, projectName)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		h.t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Use GitTestHelper to create the repository directly in the target location
	gitHelper := NewGitTestHelper(h.t)
	commitCount := h.commitCount
	if commitCount == 0 {
		commitCount = 1 // Default to 1 commit
	}

	// Create the repository directly in the target path
	createdRepoPath := gitHelper.CreateRepoWithCommits(commitCount)

	// Move contents of created repo to target location
	if err := moveDirContents(createdRepoPath, repoPath); err != nil {
		h.t.Fatalf("Failed to move repository contents: %v", err)
	}

	// Clean up the temporary repo directory
	if err := os.RemoveAll(createdRepoPath); err != nil {
		h.t.Logf("Warning: failed to clean up temporary repo: %v", err)
	}

	// Store the repository path
	h.repos[projectName] = repoPath

	return repoPath
}

// moveDirContents moves contents from src to dst directory
func moveDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// GetRepoPath returns the path for a stored repository
func (h *RepoTestHelper) GetRepoPath(projectName string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	path, exists := h.repos[projectName]
	if !exists {
		panic("repository " + projectName + " does not exist")
	}

	return path
}

// ListRepos returns a list of all stored repository names
func (h *RepoTestHelper) ListRepos() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var names []string
	for name := range h.repos {
		names = append(names, name)
	}

	return names
}

// Cleanup removes all created repositories
func (h *RepoTestHelper) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, path := range h.repos {
		if err := os.RemoveAll(path); err != nil {
			h.t.Logf("Warning: failed to remove repository %s: %v", path, err)
		}
	}

	h.repos = make(map[string]string)
}
