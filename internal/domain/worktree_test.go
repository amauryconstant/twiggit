package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktree_NewWorktree(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		branch       string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid worktree",
			path:        "/home/user/Workspaces/project/feature",
			branch:      "feature",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			branch:       "feature",
			expectError:  true,
			errorMessage: "new worktree: path cannot be empty",
		},
		{
			name:         "empty branch",
			path:         "/home/user/Workspaces/project/feature",
			branch:       "",
			expectError:  true,
			errorMessage: "new worktree: branch cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			worktree, err := NewWorktree(tc.path, tc.branch)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, worktree)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, worktree)
				assert.Equal(t, tc.path, worktree.Path())
				assert.Equal(t, tc.branch, worktree.Branch())
			}
		})
	}
}
