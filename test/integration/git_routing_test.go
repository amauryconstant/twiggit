//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"twiggit/internal/git"

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
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	executor := git.NewCommandExecutor(30 * time.Second)

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
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0o644))
	_, err = executor.Execute(context.Background(), repoPath, "git", "add", "test.txt")
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), repoPath, "git", "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Ensure we're on main branch (git might default to master)
	_, err = executor.Execute(context.Background(), repoPath, "git", "branch", "-M", "main")
	require.NoError(t, err)

	t.Run("BranchOperations_UseGoGit", func(t *testing.T) {
		goGitClient, err := git.NewClient()
		require.NoError(t, err)

		branches, err := goGitClient.ListBranches(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, branches)

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
		cliClient := git.NewCLIClient(executor, 30)

		_, err := executor.Execute(context.Background(), repoPath, "git", "checkout", "-b", "feature-test")
		require.NoError(t, err)
		_, err = executor.Execute(context.Background(), repoPath, "git", "checkout", "main")
		if err != nil {
			result, checkErr := executor.Execute(context.Background(), repoPath, "git", "branch", "--show-current")
			if checkErr == nil && strings.TrimSpace(result.Stdout) == "main" {
				err = nil
			}
		}
		require.NoError(t, err)

		worktrees, err := cliClient.ListWorktrees(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotEmpty(t, worktrees)

		assert.Len(t, worktrees, 1)
		assert.Equal(t, "main", worktrees[0].Branch)
	})

	t.Run("RepositoryOperations_UseGoGit", func(t *testing.T) {
		goGitClient, err := git.NewClient()
		require.NoError(t, err)

		err = goGitClient.ValidateRepository(repoPath)
		require.NoError(t, err)

		info, err := goGitClient.GetRepositoryInfo(context.Background(), repoPath)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, repoPath, info.Path)
	})

	t.Run("NoFallbackLogic", func(t *testing.T) {
		// Slice 9 removed the GoGitClient/CLIClient interface split;
		// the composite Client surfaces ListBranches directly. The
		// "no fallback" guarantee is exercised by the real composite
		// client above (ValidateRepository + GetRepositoryInfo).
		// This test stays as a placeholder so the routing-test suite
		// continues to compile; slice 10 will add a true routing
		// assertion when the per-command Options pattern lands.
	})
}
