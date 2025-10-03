package services

import (
	"context"
	"testing"

	"twiggit/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockWorktreeService_FunctionalBehavior(t *testing.T) {
	testCases := []struct {
		name         string
		setupFunc    func(*MockWorktreeService)
		request      *domain.CreateWorktreeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful worktree creation with functional setup",
			setupFunc: func(m *MockWorktreeService) {
				m.SetupCreateWorktreeSuccess("test-project", "feature-branch").
					WithPath("/test/worktrees/test-project/feature-branch").
					WithBranch("feature-branch")
			},
			request: &domain.CreateWorktreeRequest{
				ProjectName: "test-project",
				BranchName:  "feature-branch",
			},
			expectError: false,
		},
		{
			name: "failed worktree creation with functional setup",
			setupFunc: func(m *MockWorktreeService) {
				m.SetupCreateWorktreeError("test-project", "feature-branch", "worktree already exists")
			},
			request: &domain.CreateWorktreeRequest{
				ProjectName: "test-project",
				BranchName:  "feature-branch",
			},
			expectError:  true,
			errorMessage: "worktree already exists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockWorktreeService()
			tc.setupFunc(mock)

			result, err := mock.CreateWorktree(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}

			mock.AssertExpectations(t)
		})
	}
}

func TestMockProjectService_FunctionalBehavior(t *testing.T) {
	testCases := []struct {
		name         string
		setupFunc    func(*MockProjectService)
		projectName  string
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful project discovery with functional setup",
			setupFunc: func(m *MockProjectService) {
				m.SetupDiscoverProjectSuccess("test-project").
					WithPath("/test/projects/test-project").
					WithBranches([]string{"main", "feature-a", "feature-b"})
			},
			projectName: "test-project",
			expectError: false,
		},
		{
			name: "failed project discovery with functional setup",
			setupFunc: func(m *MockProjectService) {
				m.SetupDiscoverProjectError("test-project", "project not found")
			},
			projectName:  "test-project",
			expectError:  true,
			errorMessage: "project not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockProjectService()
			tc.setupFunc(mock)

			result, err := mock.DiscoverProject(context.Background(), tc.projectName, &domain.Context{})

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}

			mock.AssertExpectations(t)
		})
	}
}

func TestMockNavigationService_FunctionalBehavior(t *testing.T) {
	testCases := []struct {
		name         string
		setupFunc    func(*MockNavigationService)
		request      *domain.ResolvePathRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful path resolution with functional setup",
			setupFunc: func(m *MockNavigationService) {
				m.SetupResolvePathSuccess("feature-branch").
					WithPath("/test/worktrees/test-project/feature-branch").
					WithType(domain.PathTypeWorktree)
			},
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
			name: "failed path resolution with functional setup",
			setupFunc: func(m *MockNavigationService) {
				m.SetupResolvePathError("feature-branch", "target not found")
			},
			request: &domain.ResolvePathRequest{
				Target: "feature-branch",
				Context: &domain.Context{
					Type:        domain.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError:  true,
			errorMessage: "target not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := NewMockNavigationService()
			tc.setupFunc(mock)

			result, err := mock.ResolvePath(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}

			mock.AssertExpectations(t)
		})
	}
}
