package services

import (
	"context"
	"errors"
	"testing"

	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeRemover_Remove(t *testing.T) {
	ctx := context.Background()

	t.Run("should remove worktree successfully without force", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("/project/path", nil)
		mockGitClient.On("HasUncommittedChanges", ctx, "/worktree/path").
			Return(false)
		mockGitClient.On("RemoveWorktree", ctx, "/project/path", "/worktree/path", false).
			Return(nil)

		// Execute
		err := remover.Remove(ctx, "/worktree/path", false)

		// Verify
		require.NoError(t, err)
		mockGitClient.AssertExpectations(t)
	})

	t.Run("should remove worktree successfully with force", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("/project/path", nil)
		mockGitClient.On("RemoveWorktree", ctx, "/project/path", "/worktree/path", true).
			Return(nil)

		// Execute
		err := remover.Remove(ctx, "/worktree/path", true)

		// Verify
		require.NoError(t, err)
		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return error when worktree path is empty", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Execute
		err := remover.Remove(ctx, "", false)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "worktree path cannot be empty")

		// Verify no other calls were made
		mockGitClient.AssertNotCalled(t, "GetRepositoryRoot")
	})

	t.Run("should return error when getting repository root fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("", errors.New("not a git repository"))

		// Execute
		err := remover.Remove(ctx, "/worktree/path", false)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Contains(t, err.Error(), "failed to get repository root")

		mockGitClient.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "HasUncommittedChanges")
	})

	t.Run("should return error when worktree has uncommitted changes and force is false", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("/project/path", nil)
		mockGitClient.On("HasUncommittedChanges", ctx, "/worktree/path").
			Return(true)

		// Execute
		err := remover.Remove(ctx, "/worktree/path", false)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove worktree with uncommitted changes")
		assert.Contains(t, err.Error(), "uncommitted changes detected")

		mockGitClient.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "RemoveWorktree")
	})

	t.Run("should return error when remove worktree fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		remover := NewWorktreeRemover(mockGitClient)

		// Setup expectations
		mockGitClient.On("GetRepositoryRoot", ctx, "/worktree/path").
			Return("/project/path", nil)
		mockGitClient.On("HasUncommittedChanges", ctx, "/worktree/path").
			Return(false)
		mockGitClient.On("RemoveWorktree", ctx, "/project/path", "/worktree/path", false).
			Return(errors.New("git worktree remove failed"))

		// Execute
		err := remover.Remove(ctx, "/worktree/path", false)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "git command failed")
		assert.Contains(t, err.Error(), "failed to remove worktree")

		mockGitClient.AssertExpectations(t)
	})
}

func TestNewWorktreeRemover(t *testing.T) {
	mockGitClient := new(mocks.GitClientMock)

	remover := NewWorktreeRemover(mockGitClient)

	require.NotNil(t, remover, "WorktreeRemover should not be nil")
	assert.Equal(t, mockGitClient, remover.gitClient, "gitClient should be set correctly")
}
