package services

import (
	"context"
	"testing"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

func TestServiceInterfaces_ContractCompliance(t *testing.T) {
	testCases := []struct {
		name        string
		serviceType interface{}
		expectError bool
	}{
		{
			name:        "WorktreeService interface compliance",
			serviceType: (*application.WorktreeService)(nil),
			expectError: false,
		},
		{
			name:        "ProjectService interface compliance",
			serviceType: (*application.ProjectService)(nil),
			expectError: false,
		},
		{
			name:        "NavigationService interface compliance",
			serviceType: (*application.NavigationService)(nil),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Interface compliance tests - verify that interfaces are properly defined
			// by checking that the type is not nil when cast to interface{}
			if tc.serviceType == nil {
				t.Errorf("Service interface should be defined")
			}
		})
	}
}

func TestWorktreeService_InterfaceMethods(t *testing.T) {
	// This test will fail until WorktreeService interface is properly defined
	var _ application.WorktreeService = (*mockWorktreeService)(nil)
}

func TestProjectService_InterfaceMethods(t *testing.T) {
	// This test will fail until ProjectService interface is properly defined
	var _ application.ProjectService = (*mockProjectService)(nil)
}

func TestNavigationService_InterfaceMethods(t *testing.T) {
	// This test will fail until NavigationService interface is properly defined
	var _ application.NavigationService = (*mockNavigationService)(nil)
}

// Mock implementations that will fail to compile until interfaces are defined
type mockWorktreeService struct{}

func (m *mockWorktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	return nil
}

func (m *mockWorktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	return nil, nil
}

func (m *mockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	return nil, nil
}

func (m *mockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	return nil
}

type mockProjectService struct{}

func (m *mockProjectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	return &domain.ProjectInfo{
		Name:          projectName,
		Path:          "/path/to/project",
		GitRepoPath:   "/path/to/project/.git",
		Worktrees:     []*domain.WorktreeInfo{},
		Branches:      []*domain.BranchInfo{},
		Remotes:       []*domain.RemoteInfo{},
		DefaultBranch: "main",
		IsBare:        false,
	}, nil
}

func (m *mockProjectService) ValidateProject(ctx context.Context, projectPath string) error {
	return nil
}

func (m *mockProjectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	return []*domain.ProjectInfo{
		{
			Name:          "test-project",
			Path:          "/path/to/project",
			GitRepoPath:   "/path/to/project/.git",
			Worktrees:     []*domain.WorktreeInfo{},
			Branches:      []*domain.BranchInfo{},
			Remotes:       []*domain.RemoteInfo{},
			DefaultBranch: "main",
			IsBare:        false,
		},
	}, nil
}

func (m *mockProjectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	return &domain.ProjectInfo{
		Name:          "test-project",
		Path:          projectPath,
		GitRepoPath:   projectPath + "/.git",
		Worktrees:     []*domain.WorktreeInfo{},
		Branches:      []*domain.BranchInfo{},
		Remotes:       []*domain.RemoteInfo{},
		DefaultBranch: "main",
		IsBare:        false,
	}, nil
}

type mockNavigationService struct{}

func (m *mockNavigationService) ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	return nil, nil
}

func (m *mockNavigationService) ValidatePath(ctx context.Context, path string) error {
	return nil
}

func (m *mockNavigationService) GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	return nil, nil
}
