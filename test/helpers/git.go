package helpers

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
)

// GitRepo represents a test git repository with cleanup
type GitRepo struct {
	Path    string
	cleanup func()
}

// Cleanup removes the test repository
func (g *GitRepo) Cleanup() {
	if g.cleanup != nil {
		g.cleanup()
	}
}

// NewGitRepo creates a new git repository for testing.
// It initializes the repository with an initial commit.
func NewGitRepo(t *testing.T, pattern string) *GitRepo {
	t.Helper()

	tempDir, cleanup := TempDir(t, pattern)

	// Initialize git repository
	_, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Set up git config for commits
	gitConfig(t, tempDir, "user.email", "test@example.com")
	gitConfig(t, tempDir, "user.name", "Test User")

	// Create initial commit
	readmeFile := filepath.Join(tempDir, "README.md")
	require.NoError(t, os.WriteFile(readmeFile, []byte("# Test Repository\n"), 0644))

	gitCmd(t, tempDir, "add", ".")
	gitCmd(t, tempDir, "commit", "-m", "Initial commit")

	return &GitRepo{
		Path:    tempDir,
		cleanup: cleanup,
	}
}

// NewGitRepoWithBranches creates a git repository with additional branches
func NewGitRepoWithBranches(t *testing.T, pattern string, branches []string) *GitRepo {
	t.Helper()

	repo := NewGitRepo(t, pattern)

	for _, branch := range branches {
		repo.CreateBranch(t, branch)
	}

	// Switch back to main/master
	repo.SwitchToDefaultBranch(t)

	return repo
}

// CreateBranch creates a new branch with some content
func (g *GitRepo) CreateBranch(t *testing.T, branchName string) {
	t.Helper()

	gitCmd(t, g.Path, "checkout", "-b", branchName)

	// Add some content to make the branch different
	branchFile := filepath.Join(g.Path, branchName+".txt")
	content := "Content for " + branchName + "\n"
	require.NoError(t, os.WriteFile(branchFile, []byte(content), 0644))

	gitCmd(t, g.Path, "add", ".")
	gitCmd(t, g.Path, "commit", "-m", "Add content for "+branchName)
}

// SwitchToDefaultBranch switches to the default branch (main or master)
func (g *GitRepo) SwitchToDefaultBranch(t *testing.T) {
	t.Helper()

	// Try main first, then master
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = g.Path
	if err := cmd.Run(); err != nil {
		// Fallback to master
		gitCmd(t, g.Path, "checkout", "master")
	}
}

// AddMiseConfig adds a mise configuration to the repository
func (g *GitRepo) AddMiseConfig(t *testing.T) {
	t.Helper()

	miseFile := filepath.Join(g.Path, ".mise.local.toml")
	miseContent := `[tools]
node = "20.0.0"
python = "3.11"

[env]
NODE_ENV = "development"
`
	require.NoError(t, os.WriteFile(miseFile, []byte(miseContent), 0644))

	gitCmd(t, g.Path, "add", ".mise.local.toml")
	gitCmd(t, g.Path, "commit", "-m", "Add mise configuration")
}

// gitCmd runs a git command in the repository directory
func gitCmd(t *testing.T, repoPath string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Git command failed: %s\nOutput: %s", cmd.String(), string(output))
}

// gitConfig sets a git configuration value
func gitConfig(t *testing.T, repoPath, key, value string) {
	t.Helper()
	gitCmd(t, repoPath, "config", key, value)
}
