package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestProjectService_DiscoverProject_Success(t *testing.T) {
	testCases := []struct {
		name         string
		projectName  string
		context      *domain.Context
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project discovery",
			projectName: "test-project",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
			},
			expectError: false,
		},
		{
			name:        "empty project name outside context",
			projectName: "",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
			},
			expectError:  true,
			errorMessage: "project name required when outside git context",
		},
		{
			name:        "project discovery from project context",
			projectName: "",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "project",
				Path:        "/path/to/project",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestProjectService()
			result, err := service.DiscoverProject(context.Background(), tc.projectName, tc.context)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				// If projectName is empty, we should get the name from context
				expectedName := tc.projectName
				if expectedName == "" && tc.context != nil {
					expectedName = tc.context.ProjectName
				}
				assert.Equal(t, expectedName, result.Name)
			}
		})
	}
}

func TestProjectService_ValidateProject_Success(t *testing.T) {
	testCases := []struct {
		name         string
		projectPath  string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project validation",
			projectPath: "/path/to/project",
			expectError: false,
		},
		{
			name:         "empty project path",
			projectPath:  "",
			expectError:  true,
			errorMessage: "project path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestProjectService()
			err := service.ValidateProject(context.Background(), tc.projectPath)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectService_ListProjects_Success(t *testing.T) {
	testCases := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "valid projects listing",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestProjectService()
			result, err := service.ListProjects(context.Background())

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

func TestProjectService_GetProjectInfo_Success(t *testing.T) {
	testCases := []struct {
		name         string
		projectPath  string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project info",
			projectPath: "/path/to/project",
			expectError: false,
		},
		{
			name:         "empty project path",
			projectPath:  "",
			expectError:  true,
			errorMessage: "project path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestProjectService()
			result, err := service.GetProjectInfo(context.Background(), tc.projectPath)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.projectPath, result.Path)
			}
		})
	}
}

// setupTestProjectService creates a test instance of ProjectService
func setupTestProjectService() application.ProjectService {
	gitService := mocks.NewMockGitService()
	contextService := &mockContextService{}
	config := domain.DefaultConfig()

	// Configure git service mock
	gitService.MockGoGitClient.ValidateRepositoryFunc = func(path string) error {
		return nil
	}

	gitService.MockGoGitClient.GetRepositoryInfoFunc = func(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
		return &domain.GitRepository{
			Path:          repoPath,
			IsBare:        false,
			DefaultBranch: "main",
			Remotes:       []domain.RemoteInfo{},
			Branches:      []domain.BranchInfo{},
			Worktrees:     []domain.WorktreeInfo{},
			Status:        domain.RepositoryStatus{},
		}, nil
	}

	gitService.MockCLIClient.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
		return []domain.WorktreeInfo{}, nil
	}

	return NewProjectService(gitService, contextService, config)
}

// mockContextService implements service.ContextService for testing
type mockContextService struct{}

func (m *mockContextService) GetCurrentContext() (*domain.Context, error) {
	return &domain.Context{
		Type: domain.ContextOutsideGit,
	}, nil
}

func (m *mockContextService) DetectContextFromPath(path string) (*domain.Context, error) {
	return &domain.Context{
		Type: domain.ContextOutsideGit,
	}, nil
}

func (m *mockContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	return &domain.ResolutionResult{
		ResolvedPath: "/path/to/project",
		Type:         domain.PathTypeProject,
		ProjectName:  "test-project",
		Explanation:  "mock resolution",
	}, nil
}

func (m *mockContextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	return &domain.ResolutionResult{
		ResolvedPath: "/path/to/project",
		Type:         domain.PathTypeProject,
		ProjectName:  "test-project",
		Explanation:  "mock resolution",
	}, nil
}

func (m *mockContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	return []*domain.ResolutionSuggestion{
		{
			Text:        "feature-branch",
			Description: "Feature branch",
			Type:        domain.PathTypeWorktree,
			ProjectName: "test-project",
			BranchName:  "feature-branch",
		},
	}, nil
}
