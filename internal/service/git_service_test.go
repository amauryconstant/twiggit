package service

import (
	"context"
	"testing"

	"twiggit/internal/domain"
	"twiggit/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitService_InterfaceDefinitions(t *testing.T) {
	// Test that our service can be created and satisfies the interface
	mockGoGit := mocks.NewMockGoGitClient()
	mockCLI := mocks.NewMockCLIClient()

	service := NewGitService(mockGoGit, mockCLI, nil)
	assert.NotNil(t, service)

	// Test that the service implements GitService
	var _ GitService = service
}

func TestGitService_DeterministicRouting(t *testing.T) {
	// Create mock clients
	mockGoGit := mocks.NewMockGoGitClient()
	mockCLI := mocks.NewMockCLIClient()

	// Create service
	service := NewGitService(mockGoGit, mockCLI, nil)

	// Test that branch operations use GoGit only
	mockGoGit.ListBranchesFunc = func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
		return []domain.BranchInfo{{Name: "main"}}, nil
	}

	branches, err := service.ListBranches(context.Background(), "/test/repo")
	require.NoError(t, err)
	assert.Len(t, branches, 1)
	assert.Equal(t, "main", branches[0].Name)

	// Test that worktree operations use CLI only
	mockCLI.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
		return []domain.WorktreeInfo{{Path: "/test/worktree", Branch: "feature"}}, nil
	}

	worktrees, err := service.ListWorktrees(context.Background(), "/test/repo")
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, "/test/worktree", worktrees[0].Path)
}

func TestGitService_Configuration(t *testing.T) {
	// Test default configuration
	service := NewGitService(nil, nil, nil)
	assert.NotNil(t, service)

	// Test custom configuration
	config := &GitServiceConfig{
		CLITimeout:       60,
		CacheEnabled:     false,
		OperationTimeout: 45,
	}

	service2 := NewGitService(nil, nil, config)
	assert.NotNil(t, service2)
}

func TestDefaultGitServiceConfig(t *testing.T) {
	config := DefaultGitServiceConfig()

	assert.Equal(t, 30, config.CLITimeout)
	assert.True(t, config.CacheEnabled)
	assert.Equal(t, 30, config.OperationTimeout)
}

func TestGitService_Integration(t *testing.T) {
	// This test demonstrates the integration between GoGit and CLI clients
	// In a real scenario, you would use real implementations

	// Create mock GoGit client
	mockGoGit := mocks.NewMockGoGitClient()
	mockGoGit.ValidateRepositoryFunc = func(path string) error {
		if path == "/valid/repo" {
			return nil
		}
		return domain.NewGitRepositoryError(path, "not a git repository", nil)
	}

	mockGoGit.ListBranchesFunc = func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
		return []domain.BranchInfo{
			{Name: "main", IsCurrent: true},
			{Name: "feature", IsCurrent: false},
		}, nil
	}

	// Create mock CLI client
	mockCLI := mocks.NewMockCLIClient()
	mockCLI.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
		return []domain.WorktreeInfo{
			{Path: "/valid/repo", Branch: "main"},
			{Path: "/valid/repo-feature", Branch: "feature"},
		}, nil
	}

	// Create service
	service := NewGitService(mockGoGit, mockCLI, nil)

	// Test repository validation
	err := service.ValidateRepository("/valid/repo")
	require.NoError(t, err)

	err = service.ValidateRepository("/invalid/repo")
	require.Error(t, err)

	// Test branch listing
	branches, err := service.ListBranches(context.Background(), "/valid/repo")
	require.NoError(t, err)
	assert.Len(t, branches, 2)

	// Test worktree listing
	worktrees, err := service.ListWorktrees(context.Background(), "/valid/repo")
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)
}

func TestGitService_ErrorHandling(t *testing.T) {
	// Create mock clients that return errors
	mockGoGit := mocks.NewMockGoGitClient()
	mockGoGit.ListBranchesFunc = func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to list branches", nil)
	}

	mockCLI := mocks.NewMockCLIClient()
	mockCLI.CreateWorktreeFunc = func(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
		return domain.NewGitWorktreeError(worktreePath, branchName, "failed to create worktree", nil)
	}

	// Create service
	service := NewGitService(mockGoGit, mockCLI, nil)

	// Test error propagation
	branches, err := service.ListBranches(context.Background(), "/test/repo")
	require.Error(t, err)
	assert.Nil(t, branches)

	err = service.CreateWorktree(context.Background(), "/test/repo", "feature", "main", "/path/to/worktree")
	require.Error(t, err)
	var worktreeErr *domain.GitWorktreeError
	require.ErrorAs(t, err, &worktreeErr)
}

func TestGitService_ContextCancellation(t *testing.T) {
	// Create mock clients that respect context cancellation
	mockGoGit := mocks.NewMockGoGitClient()
	mockGoGit.ListBranchesFunc = func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return []domain.BranchInfo{{Name: "main"}}, nil
		}
	}

	// Create service
	service := NewGitService(mockGoGit, nil, nil)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	branches, err := service.ListBranches(ctx, "/test/repo")
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, branches)
}
