package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktree_NewWorktree(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		branch       string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid worktree",
			path:        "/home/user/workspace/project/feature-branch",
			branch:      "feature-branch",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			branch:       "main",
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
		{
			name:         "empty branch",
			path:         "/valid/path",
			branch:       "",
			expectError:  true,
			errorMessage: "branch name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worktree, err := NewWorktree(tt.path, tt.branch)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, worktree)
			} else {
				require.NoError(t, err)
				require.NotNil(t, worktree)
				assert.Equal(t, tt.path, worktree.Path)
				assert.Equal(t, tt.branch, worktree.Branch)
				assert.Equal(t, StatusUnknown, worktree.Status)
				assert.False(t, worktree.LastUpdated.IsZero())
			}
		})
	}
}

func TestWorktree_UpdateStatus(t *testing.T) {
	worktree, err := NewWorktree("/test/path", "main")
	require.NoError(t, err)

	initialTime := worktree.LastUpdated

	// Wait a bit to ensure timestamp changes
	time.Sleep(time.Millisecond)

	err = worktree.UpdateStatus(StatusClean)
	require.NoError(t, err)

	assert.Equal(t, StatusClean, worktree.Status)
	assert.True(t, worktree.LastUpdated.After(initialTime))
}

func TestWorktree_IsClean(t *testing.T) {
	worktree, err := NewWorktree("/test/path", "main")
	require.NoError(t, err)

	// Initially unknown status
	assert.False(t, worktree.IsClean())

	// Clean status
	err = worktree.UpdateStatus(StatusClean)
	require.NoError(t, err)
	assert.True(t, worktree.IsClean())

	// Dirty status
	err = worktree.UpdateStatus(StatusDirty)
	require.NoError(t, err)
	assert.False(t, worktree.IsClean())
}

func TestWorktree_String(t *testing.T) {
	worktree, err := NewWorktree("/home/user/project/feature", "feature-branch")
	require.NoError(t, err)

	result := worktree.String()
	assert.Contains(t, result, "feature-branch")
	assert.Contains(t, result, "/home/user/project/feature")
	assert.Contains(t, result, "unknown")
}
