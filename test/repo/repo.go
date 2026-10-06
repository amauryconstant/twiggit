// Package repo provides RepoTestHelper utilities for managing multiple git repository fixtures in tests.
package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"twiggit/test/git"
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

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// CreateDivergentBranch adds a new commit on top of an existing
// branch, diverging it from the base. Returns the new commit hash.
func (h *RepoTestHelper) CreateDivergentBranch(projectName, branchName, file, content string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	repoPath, ok := h.repos[projectName]
	if !ok {
		h.t.Fatalf("CreateDivergentBranch: project %q not found", projectName)
	}
	wtPath := filepath.Join(h.baseDir, "wt-"+branchName)
	runGitCmd(h.t, repoPath, "worktree", "add", "-b", branchName, wtPath, "main")
	if err := os.WriteFile(filepath.Join(wtPath, file), []byte(content), 0o644); err != nil {
		h.t.Fatalf("write %s: %v", file, err)
	}
	runGitCmd(h.t, wtPath, "add", file)
	runGitCmd(h.t, wtPath, "commit", "-m", "add "+file)
	out, err := exec.Command("git", "-C", wtPath, "rev-parse", "HEAD").Output()
	if err != nil {
		h.t.Fatalf("rev-parse: %v", err)
	}
	return string(out[:len(out)-1])
}

// CreateDirtyWorktree writes an uncommitted file inside the named
// worktree. Used to test the dirty-wt refusal path.
func (h *RepoTestHelper) CreateDirtyWorktree(projectName, branchName, file, content string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	wtPath := filepath.Join(h.baseDir, "wt-"+branchName)
	if err := os.WriteFile(filepath.Join(wtPath, file), []byte(content), 0o644); err != nil {
		h.t.Fatalf("write %s: %v", file, err)
	}
}

// CreateMidRebase sets the worktree into a mid-rebase state by
// creating a divergent branch, checking it out as the named
// branch, then running `git rebase main` from inside it. The
// conflict-free path leaves the worktree mid-rebase.
func (h *RepoTestHelper) CreateMidRebase(projectName, branchName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	wtPath := filepath.Join(h.baseDir, "wt-"+branchName)
	// The caller is expected to have already created the worktree
	// via CreateDivergentBranch. Just kick off a rebase that will
	// conflict; the rebase pauses mid-flight.
	runGitCmd(h.t, wtPath, "rebase", "main")
}

// AdvanceBase adds a new commit to the base branch (main) so the
// named worktree is now behind the base. Used to set up
// "needs-rebase" states.
func (h *RepoTestHelper) AdvanceBase(projectName, file, content string) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	repoPath, ok := h.repos[projectName]
	if !ok {
		h.t.Fatalf("AdvanceBase: project %q not found", projectName)
	}
	if err := os.WriteFile(filepath.Join(repoPath, file), []byte(content), 0o644); err != nil {
		h.t.Fatalf("write %s: %v", file, err)
	}
	runGitCmd(h.t, repoPath, "add", file)
	runGitCmd(h.t, repoPath, "commit", "-m", "advance "+file)
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD").Output()
	if err != nil {
		h.t.Fatalf("rev-parse: %v", err)
	}
	return string(out[:len(out)-1])
}

// NewRepoTestHelper creates a new RepoTestHelper instance
func NewRepoTestHelper(t *testing.T) *RepoTestHelper {
	t.Helper()
	helper := &RepoTestHelper{
		t:       t,
		baseDir: t.TempDir(),
		repos:   make(map[string]string),
	}
	t.Cleanup(helper.Cleanup)
	return helper
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
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		h.t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Use GitTestHelper to create the repository directly in the target location
	gitHelper := git.NewGitTestHelper(h.t)
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
