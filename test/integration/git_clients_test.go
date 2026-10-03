//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitOperations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create temporary directory for test repository
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")

	// Initialize git repository
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	// Use command executor to initialize git repo
	executor := git.NewCommandExecutor(30 * time.Second)

	// Initialize repository
	_, err := executor.Execute(t.Context(), repoPath, "git", "init")
	require.NoError(t, err)

	// Configure user (required for commits)
	_, err = executor.Execute(t.Context(), repoPath, "git", "config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = executor.Execute(t.Context(), repoPath, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)

	// Create initial commit
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0o644))
	_, err = executor.Execute(t.Context(), repoPath, "git", "add", "test.txt")
	require.NoError(t, err)
	_, err = executor.Execute(t.Context(), repoPath, "git", "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Ensure we're on main branch (git might default to master)
	_, err = executor.Execute(t.Context(), repoPath, "git", "branch", "-M", "main")
	require.NoError(t, err)

	t.Run("GoGitClient_BasicOperations", func(t *testing.T) {
		client, err := git.NewClient()
		require.NoError(t, err)

		// Test repository validation
		err = client.ValidateRepository(repoPath)
		require.NoError(t, err)

		// Test opening repository
		repo, err := client.OpenRepository(repoPath)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Test listing branches
		branches, err := client.ListBranches(t.Context(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, branches)

		// Test branch existence
		exists, err := client.BranchExists(t.Context(), repoPath, "main")
		require.NoError(t, err)
		assert.True(t, exists)

		// Test repository status
		status, err := client.RepositoryStatus(t.Context(), repoPath)
		require.NoError(t, err)
		assert.NotNil(t, status)
	})

	t.Run("CLIClient_WorktreeOperations", func(t *testing.T) {
		cliClient := git.NewCLIClient(executor, 30)

		// Create a feature branch first
		_, err := executor.Execute(t.Context(), repoPath, "git", "checkout", "-b", "feature-test")
		require.NoError(t, err)

		// Go back to main before creating worktree
		_, err = executor.Execute(t.Context(), repoPath, "git", "checkout", "main")
		// Don't fail if we're already on main
		if err != nil {
			// Check if we're already on main
			result, checkErr := executor.Execute(t.Context(), repoPath, "git", "branch", "--show-current")
			if checkErr == nil && strings.TrimSpace(result.Stdout) == "main" {
				err = nil // We're already on main, so no error
			}
		}
		require.NoError(t, err)

		// Create worktree
		worktreePath := filepath.Join(tempDir, "feature-worktree")
		err = cliClient.CreateWorktree(t.Context(), repoPath, "feature-test", "main", worktreePath)
		require.NoError(t, err)

		// Verify worktree was created
		assert.DirExists(t, worktreePath)

		// List worktrees
		worktrees, err := cliClient.ListWorktrees(t.Context(), repoPath)
		require.NoError(t, err)
		assert.Len(t, worktrees, 2) // main + feature worktree

		// Delete worktree
		err = cliClient.DeleteWorktree(t.Context(), repoPath, worktreePath, false)
		require.NoError(t, err)

		// Verify worktree directory is removed (or at least worktree is pruned)
		_, err = os.Stat(worktreePath)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("GitService_DeterministicRouting", func(t *testing.T) {
		goGitClient, err := git.NewClient()
		require.NoError(t, err)
		cliClient := git.NewCLIClient(executor, 30)

		branches, err := goGitClient.ListBranches(t.Context(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, branches)

		worktreePath := filepath.Join(tempDir, "routing-test")
		err = cliClient.CreateWorktree(t.Context(), repoPath, "feature-test", "main", worktreePath)
		require.NoError(t, err)

		err = cliClient.DeleteWorktree(t.Context(), repoPath, worktreePath, false)
		require.NoError(t, err)
	})
}

func TestGitOperations_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "non-existent")

	client, err := git.NewClient()
	require.NoError(t, err)

	// Test validation of non-existent repository
	err = client.ValidateRepository(nonExistentPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid git repository")

	// Test opening non-existent repository
	_, err = client.OpenRepository(nonExistentPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open git repository")
}
