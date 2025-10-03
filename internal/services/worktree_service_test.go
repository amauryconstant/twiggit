package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestWorktreeService_CreateWorktree_Success(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.CreateWorktreeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid worktree creation",
			request: &domain.CreateWorktreeRequest{
				ProjectName:  "test-project",
				BranchName:   "feature-branch",
				SourceBranch: "main",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError: false,
		},
		{
			name: "empty branch name",
			request: &domain.CreateWorktreeRequest{
				ProjectName:  "test-project",
				BranchName:   "",
				SourceBranch: "main",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError:  true,
			errorMessage: "branch name cannot be empty",
		},
		{
			name: "empty project name",
			request: &domain.CreateWorktreeRequest{
				ProjectName:  "",
				BranchName:   "feature-branch",
				SourceBranch: "main",
				Context: &domain.Context{
					Type: domain.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "project name required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestWorktreeService()
			result, err := service.CreateWorktree(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.BranchName, result.Branch)
			}
		})
	}
}

func TestWorktreeService_DeleteWorktree_Success(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.DeleteWorktreeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid worktree deletion",
			request: &domain.DeleteWorktreeRequest{
				WorktreePath: "/path/to/worktree",
				Force:        false,
				Context: &domain.Context{
					Type: domain.ContextWorktree,
				},
			},
			expectError: false,
		},
		{
			name: "empty worktree path",
			request: &domain.DeleteWorktreeRequest{
				WorktreePath: "",
				Force:        false,
				Context: &domain.Context{
					Type: domain.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestWorktreeService()
			err := service.DeleteWorktree(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWorktreeService_ListWorktrees_Success(t *testing.T) {
	testCases := []struct {
		name        string
		request     *domain.ListWorktreesRequest
		expectError bool
	}{
		{
			name: "valid worktree listing",
			request: &domain.ListWorktreesRequest{
				ProjectName: "test-project",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
				IncludeMain: true,
			},
			expectError: false,
		},
		{
			name: "list with context only",
			request: &domain.ListWorktreesRequest{
				ProjectName: "",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
				IncludeMain: false,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestWorktreeService()
			result, err := service.ListWorktrees(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestWorktreeService_GetWorktreeStatus_Success(t *testing.T) {
	testCases := []struct {
		name         string
		worktreePath string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "valid worktree status",
			worktreePath: "/path/to/worktree",
			expectError:  false,
		},
		{
			name:         "empty worktree path",
			worktreePath: "",
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestWorktreeService()
			result, err := service.GetWorktreeStatus(context.Background(), tc.worktreePath)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestWorktreeService_ValidateWorktree_Success(t *testing.T) {
	testCases := []struct {
		name         string
		worktreePath string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "valid worktree validation",
			worktreePath: "/path/to/worktree",
			expectError:  false,
		},
		{
			name:         "empty worktree path",
			worktreePath: "",
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestWorktreeService()
			err := service.ValidateWorktree(context.Background(), tc.worktreePath)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// setupTestWorktreeService creates a test instance of WorktreeService
func setupTestWorktreeService() WorktreeService {
	gitService := mocks.NewMockGitService()
	projectService := &mockProjectService{}
	config := domain.DefaultConfig()

	// Configure mocks for basic operations
	gitService.MockCLIClient.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
		if repoPath == "/path/to/project/.git" {
			return []domain.WorktreeInfo{
				{
					Path:   "/path/to/worktree",
					Branch: "feature-branch",
					Commit: "abc123",
					IsBare: false,
				},
			}, nil
		}
		// When called from ValidateWorktree with the worktree path itself
		if repoPath == "/path/to/worktree" {
			return []domain.WorktreeInfo{
				{
					Path:   "/path/to/worktree",
					Branch: "feature-branch",
					Commit: "abc123",
					IsBare: false,
				},
			}, nil
		}
		return []domain.WorktreeInfo{}, nil
	}

	gitService.MockGoGitClient.ValidateRepositoryFunc = func(path string) error {
		return nil
	}

	gitService.MockGoGitClient.GetRepositoryStatusFunc = func(ctx context.Context, repoPath string) (domain.RepositoryStatus, error) {
		return domain.RepositoryStatus{
			IsClean:   true,
			Branch:    "feature-branch",
			Commit:    "abc123",
			Modified:  []string{},
			Added:     []string{},
			Deleted:   []string{},
			Untracked: []string{},
			Ahead:     0,
			Behind:    0,
		}, nil
	}

	return NewWorktreeService(gitService, projectService, config)
}
