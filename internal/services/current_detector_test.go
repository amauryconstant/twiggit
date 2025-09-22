package services

import (
	"context"
	"errors"
	"testing"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentDirectoryDetector_Detect(t *testing.T) {
	ctx := context.Background()

	t.Run("should detect worktree when current directory is a worktree", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("/project/path", nil)
		mockGitClient.On("ListWorktrees", ctx, "/project/path").
			Return([]*domain.WorktreeInfo{
				{Path: "/project/path", Branch: "main"},
				{Path: "/worktree/path", Branch: "feature-branch"},
			}, nil)

		// Execute
		worktree, err := detector.Detect(ctx, "/worktree/path")

		// Verify
		require.NoError(t, err)
		require.NotNil(t, worktree)
		assert.Equal(t, "/worktree/path", worktree.Path)
		assert.Equal(t, "feature-branch", worktree.Branch)

		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return nil when current directory is main worktree", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/project/path").
			Return("/project/path", nil)
		mockGitClient.On("ListWorktrees", ctx, "/project/path").
			Return([]*domain.WorktreeInfo{
				{Path: "/project/path", Branch: "main"},
				{Path: "/worktree/path", Branch: "feature-branch"},
			}, nil)

		// Execute
		worktree, err := detector.Detect(ctx, "/project/path")

		// Verify
		require.NoError(t, err)
		require.Nil(t, worktree, "Should return nil for main worktree")

		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return error when current directory is not a git repository", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/non/git/dir").
			Return("", errors.New("not a git repository"))

		// Execute
		worktree, err := detector.Detect(ctx, "/non/git/dir")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Nil(t, worktree)

		mockGitClient.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "ListWorktrees")
	})

	t.Run("should return error when listing worktrees fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/current/dir").
			Return("/project/path", nil)
		mockGitClient.On("ListWorktrees", ctx, "/project/path").
			Return([]*domain.WorktreeInfo{}, errors.New("failed to list worktrees"))

		// Execute
		worktree, err := detector.Detect(ctx, "/current/dir")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list worktrees")
		assert.Nil(t, worktree)

		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return nil when current directory is not a worktree", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/other/dir").
			Return("/project/path", nil)
		mockGitClient.On("ListWorktrees", ctx, "/project/path").
			Return([]*domain.WorktreeInfo{
				{Path: "/project/path", Branch: "main"},
				{Path: "/worktree/path", Branch: "feature-branch"},
			}, nil)

		// Execute
		worktree, err := detector.Detect(ctx, "/other/dir")

		// Verify
		require.NoError(t, err)
		require.Nil(t, worktree, "Should return nil when directory is not a worktree")

		mockGitClient.AssertExpectations(t)
	})

	t.Run("should handle empty current directory path", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		detector := NewCurrentDirectoryDetector(mockGitClient)

		// Execute
		worktree, err := detector.Detect(ctx, "")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "current directory path cannot be empty")
		assert.Nil(t, worktree)

		// Verify no other calls were made
		mockGitClient.AssertNotCalled(t, "GetRepositoryRoot")
	})
}

func TestNewCurrentDirectoryDetector(t *testing.T) {
	mockGitClient := new(mocks.GitClientMock)

	detector := NewCurrentDirectoryDetector(mockGitClient)

	require.NotNil(t, detector, "CurrentDirectoryDetector should not be nil")
	assert.Equal(t, mockGitClient, detector.gitClient, "gitClient should be set correctly")
}
