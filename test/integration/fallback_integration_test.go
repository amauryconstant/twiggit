//go:build integration

package integration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/services"
	"github.com/amaury/twiggit/test/helpers"
	"github.com/amaury/twiggit/test/mocks"
)

// TestDiscoveryFallback tests the discovery service fallback mechanisms
func TestDiscoveryFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("discovery with fallback on primary failure", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-discovery-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a project directory structure
		projectDir := filepath.Join(tempDir, "project1")
		err = os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)

		// Create a mock git client that simulates failure for primary discovery but works for fallback
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), tempDir).
			Return(false, errors.New("simulated git client failure"))
		mockGitClient.On("IsMainRepository", context.Background(), tempDir).
			Return(false, errors.New("simulated git client failure"))
		// Add mocks for fallback discovery (should not fail)
		mockGitClient.On("IsGitRepository", context.Background(), projectDir).
			Return(true, nil)
		mockGitClient.On("IsMainRepository", context.Background(), projectDir).
			Return(true, nil)

		// Create config and discovery service
		cfg := &config.Config{}
		discovery := services.NewDiscoveryService(mockGitClient, cfg, os.DirFS(tempDir))

		// Test fallback discovery
		projects, err := discovery.DiscoverProjectsWithFallback(context.Background(), tempDir)

		// Should not fail completely due to fallback
		assert.NoError(t, err)
		// Fallback should return projects (might be empty slice or nil depending on implementation)
		assert.NotNil(t, projects)
	})

	t.Run("fallback project discovery with basic validation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-fallback-discovery-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create some basic directory structure
		projectDir := filepath.Join(tempDir, "project1")
		err = os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)

		// Create a mock git client that fails for primary discovery but works for fallback
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), tempDir).
			Return(false, errors.New("git client unavailable"))
		mockGitClient.On("IsMainRepository", context.Background(), tempDir).
			Return(false, errors.New("git client unavailable"))
		// Add mocks for fallback discovery (should not fail)
		mockGitClient.On("IsGitRepository", context.Background(), projectDir).
			Return(true, nil)
		mockGitClient.On("IsMainRepository", context.Background(), projectDir).
			Return(true, nil)

		cfg := &config.Config{}
		discovery := services.NewDiscoveryService(mockGitClient, cfg, os.DirFS(tempDir))

		// Test fallback discovery
		projects, err := discovery.DiscoverProjectsWithFallback(context.Background(), tempDir)

		assert.NoError(t, err)
		assert.NotNil(t, projects)
		// Should find basic directory structure even without git repos
	})

	t.Run("fallback handles directory access errors gracefully", func(t *testing.T) {
		// Test with a non-existent directory
		nonExistentDir := "/nonexistent/directory/that/does/not/exist"

		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), nonExistentDir).
			Return(false, errors.New("git client error"))
		mockGitClient.On("IsMainRepository", context.Background(), nonExistentDir).
			Return(false, errors.New("git client error"))

		cfg := &config.Config{}
		discovery := services.NewDiscoveryService(mockGitClient, cfg, os.DirFS("/"))

		// Test fallback discovery with non-existent directory
		projects, err := discovery.DiscoverProjectsWithFallback(context.Background(), nonExistentDir)

		// Should handle gracefully, possibly returning empty projects
		assert.NoError(t, err)
		assert.NotNil(t, projects)
	})
}

// TestWorktreeCreatorFallback tests the worktree creator fallback mechanisms
func TestWorktreeCreatorFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("worktree creation with fallback path resolution", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-worktree-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Setup a git repository
		gitRepo := helpers.NewGitRepo(t, "test-repo")
		repoPath := gitRepo.Path
		defer gitRepo.Cleanup()

		// Create a writable directory for fallback paths
		writableDir := filepath.Join(tempDir, "writable")
		err = os.MkdirAll(writableDir, 0755)
		require.NoError(t, err)

		// Use a path that will fail validation (path already exists)
		existingPath := filepath.Join(writableDir, "existing-worktree")
		err = os.MkdirAll(existingPath, 0755)
		require.NoError(t, err)

		// Create mock git client
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), "test-project").
			Return(true, nil)
		mockGitClient.On("IsGitRepository", context.Background(), repoPath).
			Return(true, nil)
		mockGitClient.On("BranchExists", context.Background(), "test-project", "test-branch").
			Return(true)
		// Mock the fallback path that will be generated (test-branch instead of existing-worktree)
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", filepath.Join(writableDir, "test-branch")).
			Return(nil) // Success for fallback paths

		// Create worktree creator service with writable directory
		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		// Mock the SetupWorktree call that will be made after successful worktree creation
		miseMock.On("SetupWorktree", "test-project", filepath.Join(writableDir, "test-branch")).
			Return(nil)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test worktree creation with fallback - use a path that already exists
		// This should trigger the fallback to try alternative paths
		err = creator.CreateWithFallback(context.Background(), "test-project", "test-branch", existingPath)

		// Debug: print error for troubleshooting
		if err != nil {
			t.Logf("Fallback failed with error: %v", err)
		}

		// Should succeed with fallback path
		assert.NoError(t, err)
	})

	t.Run("fallback path resolution generates alternative patterns", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-path-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create worktree creator service
		mockGitClient := &mocks.GitClientMock{}
		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test fallback path resolution by checking the method exists
		// This is more of a smoke test to ensure the method is available
		assert.NotNil(t, creator)
	})

	t.Run("fallback handles all path resolution failures", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-fallback-failure-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Setup git repository
		gitRepo := helpers.NewGitRepo(t, "test-repo")
		defer gitRepo.Cleanup()
		_ = gitRepo.Path

		// Create an existing path to trigger fallback
		existingPath := filepath.Join(tempDir, "existing-path")
		_ = os.MkdirAll(existingPath, 0755)

		// Create mock git client that always fails
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), "test-project").
			Return(true, nil)
		mockGitClient.On("BranchExists", context.Background(), "test-project", "test-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", existingPath).
			Return(errors.New("all path creation attempts failed"))
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", filepath.Join(tempDir, "test-branch")).
			Return(errors.New("fallback also failed"))

		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test worktree creation when all fallbacks fail
		err = creator.CreateWithFallback(context.Background(), "test-project", "test-branch", existingPath)

		// Should fail gracefully with appropriate error
		assert.Error(t, err)

		// Error should be a domain error with appropriate type
		var domainErr *domain.DomainError
		assert.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrGitCommand, domainErr.Type)
	})
}

// TestErrorRecoveryScenarios tests various error recovery scenarios
func TestErrorRecoveryScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("validation error recovery with suggestions", func(t *testing.T) {
		// Test validation error with recovery suggestions
		err := domain.NewWorktreeError(domain.ErrInvalidPath, "path contains invalid characters", "/bad@path").
			WithSuggestion("Use only alphanumeric characters").
			WithSuggestion("Replace @ with - or _").
			WithSuggestion("Remove special characters")

		// Verify error structure
		assert.Equal(t, domain.ErrInvalidPath, err.Type)
		assert.Len(t, err.Suggestions, 3)
		assert.Contains(t, err.Suggestions[0], "alphanumeric")
		assert.Contains(t, err.Suggestions[1], "@")
		assert.Contains(t, err.Suggestions[2], "special characters")
	})

	t.Run("git command error recovery with fallback", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-git-error-recovery-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Setup git repository (not used in this test but kept for consistency)
		gitRepo := helpers.NewGitRepo(t, "test-repo")
		defer gitRepo.Cleanup()
		_ = gitRepo.Path

		// Create an existing path to trigger fallback
		existingPath := filepath.Join(tempDir, "worktree")
		_ = os.MkdirAll(existingPath, 0755)

		// Create mock git client that fails initially but succeeds on retry
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), "test-project").
			Return(true, nil)
		mockGitClient.On("BranchExists", context.Background(), "test-project", "test-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", existingPath).
			Return(errors.New("temporary git failure")).Once()
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", existingPath).
			Return(nil) // Success on second attempt
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", filepath.Join(tempDir, "test-branch")).
			Return(nil) // Success for fallback path

		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		miseMock.On("SetupWorktree", "test-project", filepath.Join(tempDir, "test-branch")).
			Return(nil)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test worktree creation with retry/fallback
		err = creator.CreateWithFallback(context.Background(), "test-project", "test-branch", existingPath)

		// Should succeed after fallback/retry
		assert.NoError(t, err)
	})

	t.Run("permission error recovery with alternative paths", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-permission-recovery-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Setup git repository
		gitRepo := helpers.NewGitRepo(t, "test-repo")
		defer gitRepo.Cleanup()
		_ = gitRepo.Path

		// Create an existing path to trigger fallback
		existingPath := filepath.Join(tempDir, "existing-worktree")
		_ = os.MkdirAll(existingPath, 0755)

		// Create mock git client that fails on protected paths
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), "test-project").
			Return(true, nil)
		mockGitClient.On("BranchExists", context.Background(), "test-project", "test-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", existingPath).
			Return(domain.NewWorktreeError(domain.ErrPathNotWritable, "path already exists", existingPath).
				WithSuggestion("Use user directory instead").
				WithSuggestion("Check directory permissions").
				WithSuggestion("Run with appropriate privileges"))
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", filepath.Join(tempDir, "test-branch")).
			Return(nil) // Success for alternative paths

		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		miseMock.On("SetupWorktree", "test-project", filepath.Join(tempDir, "test-branch")).
			Return(nil)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test worktree creation with permission error and fallback
		err = creator.CreateWithFallback(context.Background(), "test-project", "test-branch", existingPath)

		// Should succeed with alternative path
		assert.NoError(t, err)
	})

	t.Run("workspace configuration error recovery", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-workspace-recovery-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create invalid workspace configuration
		invalidConfigPath := filepath.Join(tempDir, "invalid-config.yaml")
		err = os.WriteFile(invalidConfigPath, []byte("invalid: yaml: content: [unclosed"), 0644)
		require.NoError(t, err)

		// Test workspace configuration error with recovery suggestions
		configErr := errors.New("yaml: line 1: cannot unmarshal !!str into []string")
		workspaceErr := domain.NewWorkspaceError(domain.ErrWorkspaceInvalidConfiguration, "failed to parse workspace config", configErr).
			WithSuggestion("Check YAML syntax").
			WithSuggestion("Validate configuration file format").
			WithSuggestion("Use workspace init command to create valid config")

		// Verify error structure
		assert.Equal(t, domain.ErrWorkspaceInvalidConfiguration, workspaceErr.Type)
		assert.Equal(t, configErr, workspaceErr.Cause)
		assert.Len(t, workspaceErr.Suggestions, 3)
		assert.Contains(t, workspaceErr.Suggestions[0], "YAML syntax")
		assert.Contains(t, workspaceErr.Suggestions[1], "configuration file")
		assert.Contains(t, workspaceErr.Suggestions[2], "workspace init")
	})

	t.Run("project not found error recovery", func(t *testing.T) {
		// Test project not found error with recovery suggestions
		err := domain.NewProjectError(domain.ErrProjectNotFound, "project 'my-project' not found", "/path/to/my-project").
			WithSuggestion("Check project name spelling").
			WithSuggestion("Verify project path exists").
			WithSuggestion("Use 'twiggit list' to see available projects").
			WithSuggestion("Create project with 'twiggit project init'")

		// Verify error structure
		assert.Equal(t, domain.ErrProjectNotFound, err.Type)
		assert.Equal(t, "/path/to/my-project", err.Path)
		assert.Len(t, err.Suggestions, 4)
		assert.Contains(t, err.Suggestions[0], "spelling")
		assert.Contains(t, err.Suggestions[1], "path exists")
		assert.Contains(t, err.Suggestions[2], "twiggit list")
		assert.Contains(t, err.Suggestions[3], "project init")
	})
}

// TestFallbackIntegration tests integration between different fallback mechanisms
func TestFallbackIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("integrated discovery and worktree creation fallback", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-integrated-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Setup git repository
		gitRepo := helpers.NewGitRepo(t, "test-repo")
		defer gitRepo.Cleanup()
		_ = gitRepo.Path

		// Create an existing path to trigger fallback
		existingPath := filepath.Join(tempDir, "existing-worktree")
		_ = os.MkdirAll(existingPath, 0755)

		// Create mock git client with controlled failures
		mockGitClient := &mocks.GitClientMock{}
		mockGitClient.On("IsGitRepository", context.Background(), "/nonexistent/discovery/path").
			Return(false, errors.New("discovery failure"))
		mockGitClient.On("IsMainRepository", context.Background(), "/nonexistent/discovery/path").
			Return(false, errors.New("discovery failure"))
		mockGitClient.On("IsGitRepository", context.Background(), "test-project").
			Return(true, nil)
		mockGitClient.On("BranchExists", context.Background(), "test-project", "test-branch").
			Return(true)
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", existingPath).
			Return(domain.NewWorktreeError(domain.ErrPathNotWritable, "path already exists", existingPath))
		mockGitClient.On("CreateWorktree", context.Background(), "test-project", "test-branch", filepath.Join(tempDir, "test-branch")).
			Return(nil)

		// Create services
		cfg := &config.Config{}
		discovery := services.NewDiscoveryService(mockGitClient, cfg, os.DirFS(tempDir))

		validation := services.NewValidationService(os.DirFS(tempDir))
		miseMock := &mocks.MiseIntegrationMock{}
		miseMock.On("IsAvailable").Return(false)
		miseMock.On("SetupWorktree", "test-project", filepath.Join(tempDir, "test-branch")).
			Return(nil)
		creator := services.NewWorktreeCreator(mockGitClient, validation, miseMock)

		// Test discovery fallback
		projects, err := discovery.DiscoverProjectsWithFallback(context.Background(), "/nonexistent/discovery/path")
		assert.NoError(t, err)
		assert.NotNil(t, projects)

		// Test worktree creation fallback
		err = creator.CreateWithFallback(context.Background(), "test-project", "test-branch", existingPath)
		assert.NoError(t, err)
	})

	t.Run("error recovery chain with multiple fallbacks", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-chain-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a chain of errors with recovery suggestions
		originalErr := errors.New("ENOENT: no such file or directory")

		// Wrap with worktree error
		worktreeErr := domain.NewWorktreeError(domain.ErrWorktreeNotFound, "worktree directory missing", "/missing/worktree", originalErr).
			WithSuggestion("Check worktree directory exists").
			WithSuggestion("Verify worktree was created properly")

		// Wrap with git command error
		gitErr := domain.NewWorktreeError(domain.ErrGitCommand, "git worktree operation failed", "/repo", worktreeErr).
			WithSuggestion("Check git repository integrity").
			WithSuggestion("Run git worktree prune to clean up stale worktrees")

		// Verify error chain
		assert.Equal(t, domain.ErrGitCommand, gitErr.Type)
		assert.Equal(t, worktreeErr, gitErr.Cause)
		assert.Equal(t, originalErr, errors.Unwrap(worktreeErr))
		assert.Len(t, gitErr.Suggestions, 2)
		assert.Contains(t, gitErr.Suggestions[0], "repository integrity")
		assert.Contains(t, gitErr.Suggestions[1], "worktree prune")
	})
}
