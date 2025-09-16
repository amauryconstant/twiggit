package git

import (
	"testing"

	"github.com/amaury/twiggit/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClient_NewClient(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
}

func TestGitClient_IsGitRepository(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		path     string
		expected bool
		setupFn  func(t *testing.T) string // Returns the actual path to test
	}{
		{
			name:     "non-existent path",
			path:     "/non/existent/path",
			expected: false,
			setupFn:  func(t *testing.T) string { return "/non/existent/path" },
		},
		{
			name:     "regular directory (not git repo)",
			expected: false,
			setupFn: func(t *testing.T) string {
				return t.TempDir() // Creates a temp directory that's not a git repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFn(t)
			result, err := client.IsGitRepository(path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitClient_ListWorktrees(t *testing.T) {
	client := NewClient()

	// Test with non-git directory
	tempDir := t.TempDir()
	worktrees, err := client.ListWorktrees(tempDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
	assert.Nil(t, worktrees)
}

func TestGitClient_CreateWorktree(t *testing.T) {
	client := NewClient()

	// Test with invalid parameters
	err := client.CreateWorktree("", "main", "/target")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository path cannot be empty")

	err = client.CreateWorktree("/repo", "", "/target")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch name cannot be empty")

	err = client.CreateWorktree("/repo", "main", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target path cannot be empty")
}

func TestGitClient_RemoveWorktree(t *testing.T) {
	client := NewClient()

	// Test with invalid parameters
	err := client.RemoveWorktree("", "/target")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository path cannot be empty")

	err = client.RemoveWorktree("/repo", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree path cannot be empty")
}

func TestGitClient_GetWorktreeStatus(t *testing.T) {
	client := NewClient()

	// Test with invalid parameters
	_, err := client.GetWorktreeStatus("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree path cannot be empty")
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
