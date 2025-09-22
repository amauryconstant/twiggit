package services

import (
	"context"
	"errors"
	"testing"

	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeCreator_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create worktree successfully with existing branch", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, "/project/path", "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, "/project/path", "feature-branch", "/target/path").
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", "/project/path", "/target/path").
			Return(nil)

		// Execute
		err := creator.Create(ctx, "/project/path", "feature-branch", "/target/path")

		// Verify
		require.NoError(t, err)
		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})

	t.Run("should create worktree successfully with new branch", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, "/project/path", "new-branch").
			Return(false)
		mockGitClient.On("CreateWorktree", ctx, "/project/path", "new-branch", "/target/path").
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", "/project/path", "/target/path").
			Return(nil)

		// Execute
		err := creator.Create(ctx, "/project/path", "new-branch", "/target/path")

		// Verify
		require.NoError(t, err)
		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})

	t.Run("should return error when project path is empty", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Execute
		err := creator.Create(ctx, "", "feature-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project path cannot be empty")

		// Verify no other calls were made
		mockInfraService.AssertNotCalled(t, "PathWritable")
		mockGitClient.AssertNotCalled(t, "IsGitRepository")
	})

	t.Run("should return error when validation fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations - simulate path not writable
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(false)

		// Execute
		err := creator.Create(ctx, "/project/path", "invalid-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")

		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "IsGitRepository")
	})

	t.Run("should return error when project is not a git repository", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(false, errors.New("not a git repository"))

		// Execute
		err := creator.Create(ctx, "/project/path", "feature-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Contains(t, err.Error(), "failed to validate project repository")

		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "BranchExists")
	})

	t.Run("should return error when git repository check fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(false, errors.New("permission denied"))

		// Execute
		err := creator.Create(ctx, "/project/path", "feature-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Contains(t, err.Error(), "failed to validate project repository")

		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return error when worktree creation fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, "/project/path", "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, "/project/path", "feature-branch", "/target/path").
			Return(errors.New("git worktree add failed"))

		// Execute
		err := creator.Create(ctx, "/project/path", "feature-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "git command failed")
		assert.Contains(t, err.Error(), "failed to create worktree")

		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertNotCalled(t, "SetupWorktree")
	})

	t.Run("should continue when mise setup fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockInfraService := new(mocks.InfrastructureServiceMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(mockInfraService)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Setup expectations
		mockInfraService.On("PathExists", "/target/path").
			Return(false)
		mockInfraService.On("PathExists", "/target").
			Return(true)
		mockInfraService.On("PathWritable", "/target/path").
			Return(true)
		mockGitClient.On("IsGitRepository", ctx, "/project/path").
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, "/project/path", "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, "/project/path", "feature-branch", "/target/path").
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", "/project/path", "/target/path").
			Return(errors.New("mise not available"))

		// Execute
		err := creator.Create(ctx, "/project/path", "feature-branch", "/target/path")

		// Verify
		require.NoError(t, err, "Should not fail when mise setup fails")
		mockInfraService.AssertExpectations(t)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})
}

func TestNewWorktreeCreator(t *testing.T) {
	mockGitClient := new(mocks.GitClientMock)
	mockInfraService := new(mocks.InfrastructureServiceMock)
	mockMiseIntegration := new(mocks.MiseIntegrationMock)

	validationService := NewValidationService(mockInfraService)
	creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

	require.NotNil(t, creator, "WorktreeCreator should not be nil")
	assert.Equal(t, mockGitClient, creator.gitClient, "gitClient should be set correctly")
	assert.Equal(t, validationService, creator.validation, "validation should be set correctly")
	assert.Equal(t, mockMiseIntegration, creator.mise, "mise should be set correctly")
}
