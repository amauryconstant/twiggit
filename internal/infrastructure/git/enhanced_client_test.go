package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test git repository
func createTestGitRepo(t *testing.T) string {
	repo := testutil.NewGitRepo(t, "twiggit-git-test-*")
	t.Cleanup(repo.Cleanup)
	return repo.Path
}

// Helper function to create test branches
func createTestBranches(t *testing.T, repoPath string, branches []string) {
	for _, branch := range branches {
		cmd := exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = repoPath
		require.NoError(t, cmd.Run(), "Failed to create branch %s", branch)

		// Make a small change and commit
		testFile := filepath.Join(repoPath, branch+".txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "Add "+branch+".txt")
		cmd.Dir = repoPath
		require.NoError(t, cmd.Run())
	}

	// Switch back to main branch
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		require.NoError(t, cmd.Run())
	}
}

func TestEnhancedGitClient_GetRepositoryRoot(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name        string
		setupPath   func() string
		expectError bool
	}{
		{
			name: "should return root for repository root",
			setupPath: func() string {
				return createTestGitRepo(t)
			},
			expectError: false,
		},
		{
			name: "should return root for subdirectory in repository",
			setupPath: func() string {
				repoPath := createTestGitRepo(t)
				subDir := filepath.Join(repoPath, "subdir", "nested")
				require.NoError(t, os.MkdirAll(subDir, 0755))
				return subDir
			},
			expectError: false,
		},
		{
			name: "should return error for non-repository path",
			setupPath: func() string {
				tempDir, _ := os.MkdirTemp("", "non-repo-*")
				return tempDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath()
			defer func() { _ = os.RemoveAll(path) }()

			root, err := client.GetRepositoryRoot(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, root)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, root)
				// Verify the root is actually a git repository
				isRepo, _ := client.IsGitRepository(root)
				assert.True(t, isRepo)
			}
		})
	}
}

func TestEnhancedGitClient_GetCurrentBranch(t *testing.T) {
	client := NewClient()

	t.Run("should return current branch name", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		branch, err := client.GetCurrentBranch(repoPath)

		assert.NoError(t, err)
		// Should be either "main" or "master" depending on git version
		assert.Contains(t, []string{"main", "master"}, branch)
	})

	t.Run("should return error for non-repository", func(t *testing.T) {
		tempDir, cleanup := testutil.TempDir(t, "non-repo-*")
		defer cleanup()

		_, err := client.GetCurrentBranch(tempDir)
		assert.Error(t, err)
	})

	t.Run("should return error for empty path", func(t *testing.T) {
		_, err := client.GetCurrentBranch("")
		assert.Error(t, err)
	})
}

func TestEnhancedGitClient_GetAllBranches(t *testing.T) {
	client := NewClient()

	t.Run("should return all local branches", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branches
		testBranches := []string{"feature-1", "feature-2", "bugfix"}
		createTestBranches(t, repoPath, testBranches)

		branches, err := client.GetAllBranches(repoPath)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(branches), 4) // main/master + 3 test branches

		// Check that our test branches are included
		branchMap := make(map[string]bool)
		for _, branch := range branches {
			branchMap[branch] = true
		}

		for _, testBranch := range testBranches {
			assert.True(t, branchMap[testBranch], "Branch %s should be in the list", testBranch)
		}
	})

	t.Run("should return error for non-repository", func(t *testing.T) {
		tempDir, cleanup := testutil.TempDir(t, "non-repo-*")
		defer cleanup()

		_, err := client.GetAllBranches(tempDir)
		assert.Error(t, err)
	})
}

func TestEnhancedGitClient_GetRemoteBranches(t *testing.T) {
	client := NewClient()

	t.Run("should return empty list for repository with no remotes", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		branches, err := client.GetRemoteBranches(repoPath)

		assert.NoError(t, err)
		assert.Empty(t, branches)
	})

	t.Run("should return error for non-repository", func(t *testing.T) {
		tempDir, cleanup := testutil.TempDir(t, "non-repo-*")
		defer cleanup()

		_, err := client.GetRemoteBranches(tempDir)
		assert.Error(t, err)
	})
}

func TestEnhancedGitClient_BranchExists(t *testing.T) {
	client := NewClient()

	t.Run("should return true for existing branch", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branch
		createTestBranches(t, repoPath, []string{"test-branch"})

		exists := client.BranchExists(repoPath, "test-branch")
		assert.True(t, exists)
	})

	t.Run("should return false for non-existing branch", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		exists := client.BranchExists(repoPath, "non-existing-branch")
		assert.False(t, exists)
	})

	t.Run("should return false for non-repository", func(t *testing.T) {
		tempDir, cleanup := testutil.TempDir(t, "non-repo-*")
		defer cleanup()

		exists := client.BranchExists(tempDir, "any-branch")
		assert.False(t, exists)
	})
}

func TestEnhancedGitClient_HasUncommittedChanges(t *testing.T) {
	client := NewClient()

	t.Run("should return false for clean repository", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		hasChanges := client.HasUncommittedChanges(repoPath)
		assert.False(t, hasChanges)
	})

	t.Run("should return true for repository with uncommitted changes", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Make uncommitted changes
		testFile := filepath.Join(repoPath, "new-file.txt")
		err := os.WriteFile(testFile, []byte("new content"), 0644)
		require.NoError(t, err)

		hasChanges := client.HasUncommittedChanges(repoPath)
		assert.True(t, hasChanges)
	})

	t.Run("should return true for repository with modified files", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Modify existing file
		testFile := filepath.Join(repoPath, "README.md")
		err := os.WriteFile(testFile, []byte("# Modified README\n"), 0644)
		require.NoError(t, err)

		hasChanges := client.HasUncommittedChanges(repoPath)
		assert.True(t, hasChanges)
	})

	t.Run("should return false for non-repository", func(t *testing.T) {
		tempDir, cleanup := testutil.TempDir(t, "non-repo-*")
		defer cleanup()

		hasChanges := client.HasUncommittedChanges(tempDir)
		assert.False(t, hasChanges)
	})
}

func TestEnhancedGitClient_Integration(t *testing.T) {
	client := NewClient()

	t.Run("should work together for complete repository analysis", func(t *testing.T) {
		repoPath := createTestGitRepo(t)
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branches
		testBranches := []string{"feature-a", "feature-b"}
		createTestBranches(t, repoPath, testBranches)

		// Test repository detection
		isRepo, err := client.IsGitRepository(repoPath)
		assert.NoError(t, err)
		assert.True(t, isRepo)

		// Test repository root
		root, err := client.GetRepositoryRoot(repoPath)
		assert.NoError(t, err)
		assert.Equal(t, repoPath, root)

		// Test current branch
		currentBranch, err := client.GetCurrentBranch(repoPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, currentBranch)

		// Test all branches
		allBranches, err := client.GetAllBranches(repoPath)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(allBranches), 3) // main/master + 2 test branches

		// Test branch existence
		for _, branch := range testBranches {
			exists := client.BranchExists(repoPath, branch)
			assert.True(t, exists, "Branch %s should exist", branch)
		}

		// Test uncommitted changes (should be clean)
		hasChanges := client.HasUncommittedChanges(repoPath)
		assert.False(t, hasChanges)

		// Create worktree and test it
		worktreePath := filepath.Join(repoPath, "worktrees", "feature-a-wt")
		err = client.CreateWorktree(repoPath, "feature-a", worktreePath)
		assert.NoError(t, err)

		// Verify worktree was created
		worktrees, err := client.ListWorktrees(repoPath)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(worktrees), 2) // main repo + new worktree

		// Clean up
		err = client.RemoveWorktree(repoPath, worktreePath, false)
		assert.NoError(t, err)
	})
}
