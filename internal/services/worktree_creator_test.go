package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, projectDir, "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, projectDir, "feature-branch", targetPath).
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", projectDir, targetPath).
			Return(nil)

		// Execute
		err := creator.Create(ctx, projectDir, "feature-branch", targetPath)

		// Verify
		require.NoError(t, err)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})

	t.Run("should create worktree successfully with new branch", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, projectDir, "new-branch").
			Return(false)
		mockGitClient.On("CreateWorktree", ctx, projectDir, "new-branch", targetPath).
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", projectDir, targetPath).
			Return(nil)

		// Execute
		err := creator.Create(ctx, projectDir, "new-branch", targetPath)

		// Verify
		require.NoError(t, err)
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})

	t.Run("should return error when project path is empty", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		validationService := NewValidationService(os.DirFS("/tmp"))
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Execute
		err := creator.Create(ctx, "", "feature-branch", "/target/path")

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project path cannot be empty")

		// Verify no other calls were made
		mockGitClient.AssertNotCalled(t, "IsGitRepository")
	})

	t.Run("should return error when validation fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation - simulate non-writable parent
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create project directory but not target parent directory to simulate validation failure
		projectDir := filepath.Join(tempDir, "project")
		require.NoError(t, os.MkdirAll(projectDir, 0755))
		// Note: we don't create the target directory to simulate validation failure

		// Execute - this should fail validation because target parent doesn't exist
		err := creator.Create(ctx, projectDir, "invalid-branch", filepath.Join(tempDir, "nonexistent", "path"))

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parent directory does not exist")

		mockGitClient.AssertNotCalled(t, "IsGitRepository")
	})

	t.Run("should return error when project is not a git repository", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(false, errors.New("not a git repository"))

		// Execute
		err := creator.Create(ctx, projectDir, "feature-branch", targetPath)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Contains(t, err.Error(), "failed to validate project repository")

		mockGitClient.AssertExpectations(t)
		mockGitClient.AssertNotCalled(t, "BranchExists")
	})

	t.Run("should return error when git repository check fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(false, errors.New("permission denied"))

		// Execute
		err := creator.Create(ctx, projectDir, "feature-branch", targetPath)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
		assert.Contains(t, err.Error(), "failed to validate project repository")

		mockGitClient.AssertExpectations(t)
	})

	t.Run("should return error when worktree creation fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, projectDir, "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, projectDir, "feature-branch", targetPath).
			Return(errors.New("git worktree add failed"))

		// Execute
		err := creator.Create(ctx, projectDir, "feature-branch", targetPath)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "git command failed")
		assert.Contains(t, err.Error(), "failed to create worktree")

		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertNotCalled(t, "SetupWorktree")
	})

	t.Run("should continue when mise setup fails", func(t *testing.T) {
		// Setup mocks
		mockGitClient := new(mocks.GitClientMock)
		mockMiseIntegration := new(mocks.MiseIntegrationMock)

		// Create test directory structure for validation
		tempDir := t.TempDir()
		testFileSystem := os.DirFS(tempDir)
		validationService := NewValidationService(testFileSystem)
		creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

		// Create test directories for validation
		targetDir := filepath.Join(tempDir, "target")
		projectDir := filepath.Join(tempDir, "project")
		targetPath := filepath.Join(targetDir, "path")
		require.NoError(t, os.MkdirAll(filepath.Dir(targetPath), 0755))
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		// Setup expectations
		mockGitClient.On("IsGitRepository", ctx, projectDir).
			Return(true, nil)
		mockGitClient.On("BranchExists", ctx, projectDir, "feature-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", ctx, projectDir, "feature-branch", targetPath).
			Return(nil)
		mockMiseIntegration.On("SetupWorktree", projectDir, targetPath).
			Return(errors.New("mise not available"))

		// Execute
		err := creator.Create(ctx, projectDir, "feature-branch", targetPath)

		// Verify
		require.NoError(t, err, "Should not fail when mise setup fails")
		mockGitClient.AssertExpectations(t)
		mockMiseIntegration.AssertExpectations(t)
	})
}

func TestNewWorktreeCreator(t *testing.T) {
	mockGitClient := new(mocks.GitClientMock)
	mockMiseIntegration := new(mocks.MiseIntegrationMock)

	tempDir := t.TempDir()
	testFileSystem := os.DirFS(tempDir)
	validationService := NewValidationService(testFileSystem)
	creator := NewWorktreeCreator(mockGitClient, validationService, mockMiseIntegration)

	require.NotNil(t, creator, "WorktreeCreator should not be nil")
	assert.Equal(t, mockGitClient, creator.gitClient, "gitClient should be set correctly")
	assert.Equal(t, validationService, creator.validation, "validation should be set correctly")
	assert.Equal(t, mockMiseIntegration, creator.mise, "mise should be set correctly")
}
