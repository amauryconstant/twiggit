package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestNavigationService_ResolvePath_Success(t *testing.T) {
	testCases := []struct {
		name         string
		request      *domain.ResolvePathRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid branch resolution from project context",
			request: &domain.ResolvePathRequest{
				Target: "feature-branch",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError: false,
		},
		{
			name: "empty target",
			request: &domain.ResolvePathRequest{
				Target: "",
				Context: &domain.Context{
					Type: domain.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "target cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestNavigationService()
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

func TestNavigationService_ValidatePath_Success(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid path validation",
			path:        ".", // Current directory should always exist
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorMessage: "path cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestNavigationService()
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

func TestNavigationService_GetNavigationSuggestions_Success(t *testing.T) {
	testCases := []struct {
		name        string
		context     *domain.Context
		partial     string
		expectError bool
	}{
		{
			name: "valid suggestions from project context",
			context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: "test-project",
			},
			partial:     "feat",
			expectError: false,
		},
		{
			name: "suggestions from outside context",
			context: &domain.Context{
				Type: domain.ContextOutsideGit,
			},
			partial:     "test",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := setupTestNavigationService()
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

// setupTestNavigationService creates a test instance of NavigationService
func setupTestNavigationService() NavigationService {
	projectService := &mockProjectService{}
	contextService := &mockContextService{}
	config := domain.DefaultConfig()

	return NewNavigationService(projectService, contextService, config)
}
