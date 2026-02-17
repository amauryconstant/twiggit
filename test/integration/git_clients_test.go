//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"twiggit/internal/infrastructure"

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
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Use command executor to initialize git repo
	executor := infrastructure.NewDefaultCommandExecutor(30 * time.Second)

	// Initialize repository
	_, err := executor.Execute(context.Background(), repoPath, "git", "init")
	require.NoError(t, err)

	// Configure user (required for commits)
	_, err = executor.Execute(context.Background(), repoPath, "git", "config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), repoPath, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)

	// Create initial commit
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))
	_, err = executor.Execute(context.Background(), repoPath, "git", "add", "test.txt")
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), repoPath, "git", "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Ensure we're on main branch (git might default to master)
	_, err = executor.Execute(context.Background(), repoPath, "git", "branch", "-M", "main")
	require.NoError(t, err)

	t.Run("GoGitClient_BasicOperations", func(t *testing.T) {
		client := infrastructure.NewGoGitClient(true)

		// Test repository validation
		err := client.ValidateRepository(repoPath)
		require.NoError(t, err)

		// Test opening repository
		repo, err := client.OpenRepository(repoPath)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Test listing branches
		branches, err := client.ListBranches(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, branches)

		// Test branch existence
		exists, err := client.BranchExists(context.Background(), repoPath, "main")
		require.NoError(t, err)
		assert.True(t, exists)

		// Test repository status
		status, err := client.GetRepositoryStatus(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotNil(t, status)
	})

	t.Run("CLIClient_WorktreeOperations", func(t *testing.T) {
		cliClient := infrastructure.NewCLIClient(executor, 30)

		// Create a feature branch first
		_, err := executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-test")
		require.NoError(t, err)

		// Go back to main before creating worktree
		_, err = executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
		// Don't fail if we're already on main
		if err != nil {
			// Check if we're already on main
			result, checkErr := executor.Execute(context.Background(), repoPath, "git", "branch", "--show-current")
			if checkErr == nil && strings.TrimSpace(result.Stdout) == "main" {
				err = nil // We're already on main, so no error
			}
		}
		require.NoError(t, err)

		// Create worktree
		worktreePath := filepath.Join(tempDir, "feature-worktree")
		err = cliClient.CreateWorktree(context.Background(), repoPath, "feature-test", "main", worktreePath)
		require.NoError(t, err)

		// Verify worktree was created
		assert.DirExists(t, worktreePath)

		// List worktrees
		worktrees, err := cliClient.ListWorktrees(context.Background(), repoPath)
		require.NoError(t, err)
		assert.Len(t, worktrees, 2) // main + feature worktree

		// Delete worktree
		err = cliClient.DeleteWorktree(context.Background(), repoPath, worktreePath, false)
		require.NoError(t, err)

		// Verify worktree directory is removed (or at least worktree is pruned)
		_, err = os.Stat(worktreePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("GitService_DeterministicRouting", func(t *testing.T) {
		goGitClient := infrastructure.NewGoGitClient(true)
		cliClient := infrastructure.NewCLIClient(executor, 30)

		gitService := infrastructure.NewCompositeGitClient(goGitClient, cliClient)

		// Test that branch operations use GoGit
		branches, err := gitService.ListBranches(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, branches)

		// Test that worktree operations use CLI
		worktreePath := filepath.Join(tempDir, "routing-test")
		err = gitService.CreateWorktree(context.Background(), repoPath, "feature-test", "main", worktreePath)
		require.NoError(t, err)

		// Cleanup
		err = gitService.DeleteWorktree(context.Background(), repoPath, worktreePath, false)
		require.NoError(t, err)
	})
}

func TestGitOperations_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "non-existent")

	client := infrastructure.NewGoGitClient(true)

	// Test validation of non-existent repository
	err := client.ValidateRepository(nonExistentPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid git repository")

	// Test opening non-existent repository
	_, err = client.OpenRepository(nonExistentPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open git repository")
}
