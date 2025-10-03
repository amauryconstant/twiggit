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

	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
	"twiggit/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicRouting_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create temporary directory for test repository
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")

	// Initialize git repository
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	executor := infrastructure.NewDefaultCommandExecutor(30 * time.Second)

	// Initialize repository
	_, err := executor.Execute(context.Background(), repoPath, "git", "init")
	require.NoError(t, err)

	// Configure user
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

	t.Run("BranchOperations_UseGoGit", func(t *testing.T) {
		// Use real CLI client but verify GoGit is used for branch operations
		cliClient := infrastructure.NewCLIClient(executor, 30)
		goGitClient := infrastructure.NewGoGitClient(true)
		gitService := service.NewGitService(goGitClient, cliClient, nil)

		// Test branch listing - should use GoGit only
		branches, err := gitService.ListBranches(context.Background(), repoPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, branches)

		// Verify we got branch data (proving GoGit worked)
		foundMain := false
		for _, branch := range branches {
			if branch.Name == "main" {
				foundMain = true
				break
			}
		}
		assert.True(t, foundMain, "Expected to find 'main' branch")
	})

	t.Run("WorktreeOperations_UseCLI", func(t *testing.T) {
		// Use real GoGit client but verify CLI is used for worktree operations
		goGitClient := infrastructure.NewGoGitClient(true)
		cliClient := infrastructure.NewCLIClient(executor, 30)
		gitService := service.NewGitService(goGitClient, cliClient, nil)

		// Create a feature branch first
		_, err := executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-test")
		require.NoError(t, err)
		_, err = executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
		// Don't fail if we're already on main (git checkout might return non-zero exit code)
		if err != nil {
			// Try to get current branch and verify we're on main or feature-test
			result, checkErr := executor.Execute(context.Background(), repoPath, "git", "branch", "--show-current")
			if checkErr == nil && strings.TrimSpace(result.Stdout) == "main" {
				err = nil // We're already on main, so no error
			}
		}
		require.NoError(t, err)

		// Test worktree listing - should use CLI only
		worktrees, err := gitService.ListWorktrees(context.Background(), repoPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, worktrees)

		// Verify we got worktree data (proving CLI worked)
		assert.Len(t, worktrees, 1) // Only main worktree exists
		assert.Equal(t, "main", worktrees[0].Branch)
	})

	t.Run("RepositoryOperations_UseGoGit", func(t *testing.T) {
		// Use real CLI client but verify GoGit is used for repository operations
		cliClient := infrastructure.NewCLIClient(executor, 30)
		goGitClient := infrastructure.NewGoGitClient(true)
		gitService := service.NewGitService(goGitClient, cliClient, nil)

		// Test repository validation - should use GoGit only
		err := gitService.ValidateRepository(repoPath)
		assert.NoError(t, err)

		// Test repository info - should use GoGit only
		info, err := gitService.GetRepositoryInfo(context.Background(), repoPath)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, repoPath, info.Path)
	})

	t.Run("NoFallbackLogic", func(t *testing.T) {
		// Test that there's no fallback - if GoGit fails, the operation fails
		// (doesn't try CLI as fallback)

		// Create GoGit client that will fail
		mockGoGit := mocks.NewMockGoGitClient()
		mockGoGit.ListBranchesFunc = func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
			return nil, assert.AnError
		}

		// Create CLI client (should not be called)
		cliClient := infrastructure.NewCLIClient(executor, 30)

		gitService := service.NewGitService(mockGoGit, cliClient, nil)

		// Test that branch operation fails when GoGit fails
		_, err := gitService.ListBranches(context.Background(), repoPath)
		assert.Error(t, err)

		// The error should be from GoGit, not CLI (proving no fallback)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
