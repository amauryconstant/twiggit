package helpers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WorktreeTestHelper coverage tests

func TestWorktreeTestHelper_New(t *testing.T) {
	helper := NewWorktreeTestHelper()
	assert.NotNil(t, helper)
	assert.Equal(t, 30*time.Second, helper.timeout)
}

func TestWorktreeTestHelper_WithTimeout(t *testing.T) {
	helper := NewWorktreeTestHelper()
	result := helper.WithTimeout(10 * time.Second)

	assert.Equal(t, helper, result) // Returns same instance for chaining
	assert.Equal(t, 10*time.Second, helper.timeout)
}

func TestWorktreeTestHelper_CreateWorktree(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	helper := NewWorktreeTestHelper()
	worktreePath := filepath.Join(t.TempDir(), "feature-branch")

	err := helper.CreateWorktree(repoPath, worktreePath, "feature-branch")
	require.NoError(t, err)

	assert.DirExists(t, worktreePath)
	assert.FileExists(t, filepath.Join(worktreePath, ".git"))
}

func TestWorktreeTestHelper_CreateWorktreeFromSource(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	cmd := exec.Command("git", "checkout", "-b", "source-branch")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	helper := NewWorktreeTestHelper()
	worktreePath := filepath.Join(t.TempDir(), "feature-from-source")

	err := helper.CreateWorktreeFromSource(repoPath, worktreePath, "feature-from-source", "source-branch")
	require.NoError(t, err)

	assert.DirExists(t, worktreePath)
}

func TestWorktreeTestHelper_CheckoutExistingBranch(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	cmd := exec.Command("git", "branch", "existing-branch")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	helper := NewWorktreeTestHelper()
	worktreePath := filepath.Join(t.TempDir(), "existing-branch")

	err := helper.CheckoutExistingBranch(repoPath, worktreePath, "existing-branch")
	require.NoError(t, err)

	assert.DirExists(t, worktreePath)
}

func TestWorktreeTestHelper_RemoveWorktree(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	helper := NewWorktreeTestHelper()
	worktreePath := filepath.Join(t.TempDir(), "to-remove")

	err := helper.CreateWorktree(repoPath, worktreePath, "to-remove")
	require.NoError(t, err)

	err = helper.RemoveWorktree(worktreePath, false)
	require.NoError(t, err)

	assert.NoDirExists(t, worktreePath)
}

func TestWorktreeTestHelper_RemoveWorktree_Force(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	helper := NewWorktreeTestHelper()
	worktreePath := filepath.Join(t.TempDir(), "to-force-remove")

	err := helper.CreateWorktree(repoPath, worktreePath, "to-force-remove")
	require.NoError(t, err)

	err = helper.RemoveWorktree(worktreePath, true)
	require.NoError(t, err)

	assert.NoDirExists(t, worktreePath)
}

func TestWorktreeTestHelper_RemoveWorktree_NonExistent(t *testing.T) {
	helper := NewWorktreeTestHelper()

	err := helper.RemoveWorktree("/nonexistent/path", false)
	// This should succeed because git worktree remove handles non-existent gracefully
	// when the output contains "not found" or "does not exist"
	// But for truly non-existent, it will fail on findRepoPath
	assert.Error(t, err)
}

func TestWorktreeTestHelper_ListWorktrees(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	helper := NewWorktreeTestHelper()

	for i := 0; i < 3; i++ {
		branchName := "list-test-" + strings.ToLower(time.Now().Format("20060102150405")) + string(rune('a'+i))
		worktreePath := filepath.Join(t.TempDir(), branchName)
		err := helper.CreateWorktree(repoPath, worktreePath, branchName)
		require.NoError(t, err)
	}

	worktrees, err := helper.ListWorktrees(repoPath)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(worktrees), 1)
}

func TestWorktreeTestHelper_ParseWorktreeList(t *testing.T) {
	helper := NewWorktreeTestHelper()

	output := `worktree /path/to/main
HEAD abc123def456
branch refs/heads/main

worktree /path/to/feature
HEAD def456abc123
branch refs/heads/feature-branch

worktree /path/to/detached
HEAD 123456abcdef
detached
`

	worktrees, err := helper.parseWorktreeList(output)
	require.NoError(t, err)

	assert.Len(t, worktrees, 3)
	assert.Equal(t, "/path/to/main", worktrees[0].Path)
	assert.Equal(t, "abc123def456", worktrees[0].Commit)
	assert.Equal(t, "main", worktrees[0].Branch)
	assert.False(t, worktrees[0].IsDetached)
	assert.True(t, worktrees[2].IsDetached)
}

func TestWorktreeTestHelper_FindRepoPath(t *testing.T) {
	repoPath := createTestRepoForWorktree(t)

	result := findRepoPath(repoPath)
	assert.Equal(t, repoPath, result)
}

// createTestRepoForWorktree creates a test git repository
func createTestRepoForWorktree(t *testing.T) string {
	t.Helper()

	repoPath := t.TempDir()

	cmd := exec.Command("git", "init", repoPath)
	require.NoError(t, cmd.Run(), "Failed to init repo")

	cmd = exec.Command("git", "config", "user.email", "test@twiggit.dev")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit", "--allow-empty")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_EMAIL=test@twiggit.dev",
		"GIT_AUTHOR_NAME=Test User",
	)
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = repoPath
	require.NoError(t, cmd.Run())

	return repoPath
}
