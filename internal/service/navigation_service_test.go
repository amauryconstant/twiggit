package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
	"twiggit/test/mocks"
)

func TestNavigationService(t *testing.T) {
	tests := []struct {
		name         string
		request      *core.ResolvePathRequest
		expectError  bool
		errorMessage string
		setupMocks   func(*mocks.MockProjectService, *mocks.MockContextService)
	}{
		{
			name: "valid branch resolution from project context",
			request: &core.ResolvePathRequest{
				Target: "feature-branch",
				Context: &core.Context{
					Type:        core.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError: false,
			setupMocks:  func(ps *mocks.MockProjectService, cs *mocks.MockContextService) {},
		},
		{
			name: "empty target",
			request: &core.ResolvePathRequest{
				Target: "",
				Context: &core.Context{
					Type: core.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "validation failed for ResolvePathRequest.target: cannot be empty",
			setupMocks:   func(ps *mocks.MockProjectService, cs *mocks.MockContextService) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := core.DefaultConfig()

			projectService := mocks.NewMockProjectService()
			contextService := mocks.NewMockContextService()

			tc.setupMocks(projectService, contextService)

			// Only setup the default expectation if not expecting an error
			if !tc.expectError {
				contextService.On("ResolveIdentifierFromContext", mock.AnythingOfType("*core.Context"), mock.AnythingOfType("string")).Return(&core.ResolutionResult{
					Type:         core.PathTypeWorktree,
					ResolvedPath: "/path/to/worktree",
				}, nil)
			}

			t.Cleanup(func() {
				projectService.AssertExpectations(t)
				contextService.AssertExpectations(t)
			})

			service := NewNavigationService(projectService, contextService, config)
			result, err := service.ResolvePath(context.Background(), tc.request)

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

func TestNavigationService_ValidatePath(t *testing.T) {
	config := core.DefaultConfig()
	projectService := mocks.NewMockProjectService()
	contextService := mocks.NewMockContextService()
	service := NewNavigationService(projectService, contextService, config)

	tests := []struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid path validation",
			path:        ".",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorMessage: "validation failed for ValidatePath.path: cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidatePath(context.Background(), tc.path)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNavigationService_GetNavigationSuggestions(t *testing.T) {
	config := core.DefaultConfig()

	tests := []struct {
		name        string
		context     *core.Context
		partial     string
		expectError bool
	}{
		{
			name: "valid suggestions from project context",
			context: &core.Context{
				Type:        core.ContextProject,
				ProjectName: "test-project",
			},
			partial:     "feat",
			expectError: false,
		},
		{
			name: "suggestions from outside context",
			context: &core.Context{
				Type: core.ContextOutsideGit,
			},
			partial:     "test",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projectService := mocks.NewMockProjectService()
			contextService := mocks.NewMockContextService()

			// Setup both mock expectations since the service may call either
			contextService.On("GetCompletionSuggestionsFromContext", mock.AnythingOfType("*core.Context"), mock.AnythingOfType("string")).Return([]*core.ResolutionSuggestion{
				{
					Text:        "feature-branch",
					Description: "Feature branch",
					Type:        core.PathTypeWorktree,
					ProjectName: "test-project",
					BranchName:  "feature-branch",
				},
			}, nil).Maybe()

			contextService.On("GetCompletionSuggestions", mock.AnythingOfType("string")).Return([]*core.ResolutionSuggestion{
				{
					Text:        "test-project",
					Description: "Test project",
					Type:        core.PathTypeProject,
					ProjectName: "test-project",
				},
			}, nil).Maybe()

			service := NewNavigationService(projectService, contextService, config)
			result, err := service.GetNavigationSuggestions(context.Background(), tc.context, tc.partial)

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
