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

func TestWorktree_EnhancedFeatures(t *testing.T) {
	t.Run("should support commit hash tracking", func(t *testing.T) {
		worktree, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		// This should fail initially - we need to add Commit field
		err = worktree.SetCommit("abc123def456")
		assert.NoError(t, err)
		assert.Equal(t, "abc123def456", worktree.GetCommit())
	})

	t.Run("should validate path existence", func(t *testing.T) {
		// Test with non-existent path
		worktree, err := NewWorktree("/non/existent/path", "main")
		require.NoError(t, err)

		// This should fail initially - we need to add path validation
		isValid, err := worktree.ValidatePathExists()
		assert.Error(t, err)
		assert.False(t, isValid)
	})

	t.Run("should support status aging", func(t *testing.T) {
		worktree, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		// Set initial status
		err = worktree.UpdateStatus(StatusClean)
		require.NoError(t, err)

		// This should fail initially - we need to add status aging
		isStale := worktree.IsStatusStale()
		assert.False(t, isStale) // Should not be stale immediately

		// This should fail initially - we need to add stale threshold configuration
		isStale = worktree.IsStatusStaleWithThreshold(time.Hour)
		assert.False(t, isStale)
	})

	t.Run("should support equality comparison", func(t *testing.T) {
		worktree1, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		worktree2, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		worktree3, err := NewWorktree("/different/path", "main")
		require.NoError(t, err)

		// This should fail initially - we need to add equality methods
		assert.True(t, worktree1.Equals(worktree2))
		assert.False(t, worktree1.Equals(worktree3))
		assert.True(t, worktree1.SameLocationAs(worktree2))
		assert.False(t, worktree1.SameLocationAs(worktree3))
	})

	t.Run("should support worktree metadata", func(t *testing.T) {
		worktree, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		// This should fail initially - we need to add metadata support
		worktree.SetMetadata("last-checked-by", "user1")
		worktree.SetMetadata("priority", "high")

		value, exists := worktree.GetMetadata("last-checked-by")
		assert.True(t, exists)
		assert.Equal(t, "user1", value)

		value, exists = worktree.GetMetadata("priority")
		assert.True(t, exists)
		assert.Equal(t, "high", value)

		_, exists = worktree.GetMetadata("non-existent")
		assert.False(t, exists)
	})

	t.Run("should support worktree health check", func(t *testing.T) {
		worktree, err := NewWorktree("/test/path", "main")
		require.NoError(t, err)

		// This should fail initially - we need to add health check
		health := worktree.GetHealth()
		assert.NotNil(t, health)
		assert.Equal(t, "unhealthy", health.Status)
		assert.Contains(t, health.Issues, "path not validated")
	})
}
