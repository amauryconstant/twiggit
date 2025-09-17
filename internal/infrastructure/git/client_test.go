package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/pkg/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClient_NewClient(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
}

func TestGitClient_IsGitRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a non-git directory
	nonGitDir := filepath.Join(tempDir, "non-git")
	err := os.Mkdir(nonGitDir, 0755)
	require.NoError(t, err)

	client := NewClient()

	t.Run("should return true for valid git repository", func(t *testing.T) {
		// Create a git repository
		gitDir := filepath.Join(tempDir, "git-repo")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		isRepo, err := client.IsGitRepository(gitDir)
		assert.NoError(t, err)
		assert.True(t, isRepo)
	})

	t.Run("should return false for non-git directory", func(t *testing.T) {
		isRepo, err := client.IsGitRepository(nonGitDir)
		assert.NoError(t, err)
		assert.False(t, isRepo)
	})

	t.Run("should return error for non-existent path", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "does-not-exist")
		isRepo, err := client.IsGitRepository(nonExistentDir)
		assert.Error(t, err)
		assert.False(t, isRepo)
	})

	t.Run("should return error for empty path", func(t *testing.T) {
		isRepo, err := client.IsGitRepository("")
		assert.Error(t, err)
		assert.False(t, isRepo)
	})
}

func TestGitClient_IsMainRepository(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient()

	t.Run("should return true for main repository", func(t *testing.T) {
		// Create a git repository
		gitDir := filepath.Join(tempDir, "main-repo")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		isMain, err := client.IsMainRepository(gitDir)
		assert.NoError(t, err)
		assert.True(t, isMain)
	})

	t.Run("should return false for worktree", func(t *testing.T) {
		// Create main repository
		mainDir := filepath.Join(tempDir, "main-for-worktree")
		_, err := git.PlainInit(mainDir, false)
		require.NoError(t, err)

		// Create a worktree
		worktreeDir := filepath.Join(tempDir, "test-worktree")
		err = client.CreateWorktree(mainDir, "new-branch", worktreeDir)
		require.NoError(t, err)

		isMain, err := client.IsMainRepository(worktreeDir)
		assert.NoError(t, err)
		assert.False(t, isMain)
	})

	t.Run("should return false for non-git directory", func(t *testing.T) {
		nonGitDir := filepath.Join(tempDir, "not-git")
		err := os.MkdirAll(nonGitDir, 0755)
		require.NoError(t, err)

		isMain, err := client.IsMainRepository(nonGitDir)
		assert.NoError(t, err)
		assert.False(t, isMain)
	})

	t.Run("should return error for empty path", func(t *testing.T) {
		isMain, err := client.IsMainRepository("")
		assert.Error(t, err)
		assert.False(t, isMain)
	})
}

func TestGitClient_ListWorktrees(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient()

	t.Run("should return main repository for repository with no worktrees", func(t *testing.T) {
		// Create a git repository
		gitDir := filepath.Join(tempDir, "main-repo")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		worktrees, err := client.ListWorktrees(gitDir)
		assert.NoError(t, err)
		assert.Len(t, worktrees, 1)
		assert.Equal(t, gitDir, worktrees[0].Path)
		assert.NotEmpty(t, worktrees[0].Branch)
		assert.NotEmpty(t, worktrees[0].Commit)
	})

	t.Run("should return error for non-git directory", func(t *testing.T) {
		nonGitDir := filepath.Join(tempDir, "non-git")
		err := os.Mkdir(nonGitDir, 0755)
		require.NoError(t, err)

		worktrees, err := client.ListWorktrees(nonGitDir)
		assert.Error(t, err)
		assert.Nil(t, worktrees)
	})

	t.Run("should return error for empty repository path", func(t *testing.T) {
		worktrees, err := client.ListWorktrees("")
		assert.Error(t, err)
		assert.Nil(t, worktrees)
	})
}

func TestGitClient_CreateWorktree(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient()

	t.Run("should create worktree from existing branch", func(t *testing.T) {
		// Create a git repository with initial commit
		gitDir := filepath.Join(tempDir, "main-repo-create")
		repo, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		// Create initial commit
		wt, err := repo.Worktree()
		require.NoError(t, err)

		// Create a test file
		testFile := filepath.Join(gitDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = wt.Add("test.txt")
		require.NoError(t, err)

		_, err = wt.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
		})
		require.NoError(t, err)

		// Get current branch name (could be master or main depending on git config)
		head, err := repo.Head()
		require.NoError(t, err)
		branchName := head.Name().Short()

		// Create a new branch for the worktree
		branchRef := plumbing.ReferenceName("refs/heads/" + branchName + "-worktree")
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: true,
		})
		require.NoError(t, err)

		// Switch back to main branch
		mainRef := plumbing.ReferenceName("refs/heads/" + branchName)
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: mainRef,
		})
		require.NoError(t, err)

		// Create worktree from the new branch
		worktreePath := filepath.Join(tempDir, "worktree-1")
		err = client.CreateWorktree(gitDir, branchName+"-worktree", worktreePath)
		assert.NoError(t, err)

		// Verify worktree was created
		assert.DirExists(t, worktreePath)
	})

	t.Run("should return error for empty repository path", func(t *testing.T) {
		worktreePath := filepath.Join(tempDir, "worktree-1")
		err := client.CreateWorktree("", "main", worktreePath)
		assert.Error(t, err)
	})

	t.Run("should return error for empty branch name", func(t *testing.T) {
		gitDir := filepath.Join(tempDir, "main-repo-empty-branch")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		worktreePath := filepath.Join(tempDir, "worktree-empty-branch")
		err = client.CreateWorktree(gitDir, "", worktreePath)
		assert.Error(t, err)
	})

	t.Run("should return error for empty target path", func(t *testing.T) {
		gitDir := filepath.Join(tempDir, "main-repo-empty-target")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		err = client.CreateWorktree(gitDir, "main", "")
		assert.Error(t, err)
	})

	t.Run("should return error for existing target path", func(t *testing.T) {
		gitDir := filepath.Join(tempDir, "main-repo-existing-target")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		worktreePath := filepath.Join(tempDir, "existing-dir")
		err = os.Mkdir(worktreePath, 0755)
		require.NoError(t, err)

		err = client.CreateWorktree(gitDir, "main", worktreePath)
		assert.Error(t, err)
	})
}

func TestGitClient_GetWorktreeStatus(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient()

	t.Run("should return status for clean worktree", func(t *testing.T) {
		// Create a git repository
		gitDir := filepath.Join(tempDir, "main-repo")
		repo, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		// Create initial commit
		wt, err := repo.Worktree()
		require.NoError(t, err)

		testFile := filepath.Join(gitDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = wt.Add("test.txt")
		require.NoError(t, err)

		_, err = wt.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
		})
		require.NoError(t, err)

		status, err := client.GetWorktreeStatus(gitDir)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Clean)
		assert.Equal(t, gitDir, status.Path)
		assert.NotEmpty(t, status.Branch)
		assert.NotEmpty(t, status.Commit)
	})

	t.Run("should return error for non-existent worktree path", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "does-not-exist")
		status, err := client.GetWorktreeStatus(nonExistentPath)
		assert.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("should return error for empty worktree path", func(t *testing.T) {
		status, err := client.GetWorktreeStatus("")
		assert.Error(t, err)
		assert.Nil(t, status)
	})
}

func TestGitClient_RemoveWorktree(t *testing.T) {
	tempDir := t.TempDir()
	client := NewClient()

	t.Run("should remove existing worktree", func(t *testing.T) {
		// Create a git repository
		gitDir := filepath.Join(tempDir, "main-repo-remove")
		repo, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		// Create initial commit
		wt, err := repo.Worktree()
		require.NoError(t, err)

		testFile := filepath.Join(gitDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = wt.Add("test.txt")
		require.NoError(t, err)

		_, err = wt.Commit("Initial commit", &git.CommitOptions{
			Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
		})
		require.NoError(t, err)

		// Get current branch name
		head, err := repo.Head()
		require.NoError(t, err)
		branchName := head.Name().Short()

		// Create a new branch for the worktree
		branchRef := plumbing.ReferenceName("refs/heads/" + branchName + "-remove")
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: true,
		})
		require.NoError(t, err)

		// Switch back to main branch
		mainRef := plumbing.ReferenceName("refs/heads/" + branchName)
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: mainRef,
		})
		require.NoError(t, err)

		// Create worktree from the new branch
		worktreePath := filepath.Join(tempDir, "worktree-remove")
		err = client.CreateWorktree(gitDir, branchName+"-remove", worktreePath)
		require.NoError(t, err)

		// Remove worktree
		err = client.RemoveWorktree(gitDir, worktreePath, false)
		assert.NoError(t, err)

		// Verify worktree directory was removed (git worktree remove removes the directory by default)
		assert.NoDirExists(t, worktreePath)
	})

	t.Run("should return error for empty repository path", func(t *testing.T) {
		worktreePath := filepath.Join(tempDir, "worktree-1")
		err := client.RemoveWorktree("", worktreePath, false)
		assert.Error(t, err)
	})

	t.Run("should return error for empty worktree path", func(t *testing.T) {
		gitDir := filepath.Join(tempDir, "main-repo-empty-worktree")
		_, err := git.PlainInit(gitDir, false)
		require.NoError(t, err)

		err = client.RemoveWorktree(gitDir, "", false)
		assert.Error(t, err)
	})
}

func TestWorktreeInfo_Validation(t *testing.T) {
	tests := []struct {
		name        string
		worktree    types.WorktreeInfo
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid worktree info",
			worktree: types.WorktreeInfo{
				Path:   "/valid/path",
				Branch: "main",
				Commit: "abc123",
				Clean:  true,
			},
			expectError: false,
		},
		{
			name: "empty path",
			worktree: types.WorktreeInfo{
				Path:   "",
				Branch: "main",
				Commit: "abc123",
			},
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name: "empty branch",
			worktree: types.WorktreeInfo{
				Path:   "/valid/path",
				Branch: "",
				Commit: "abc123",
			},
			expectError: true,
			errorMsg:    "branch cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.worktree.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
