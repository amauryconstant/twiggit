package services

import (
	"context"
	"fmt"

	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockProjectService provides functional setup builders for project service testing
type MockProjectService struct {
	mock.Mock
}

// NewMockProjectService creates a new MockProjectService instance
func NewMockProjectService() *MockProjectService {
	return &MockProjectService{}
}

// DiscoverProject mocks discovering a project
func (m *MockProjectService) DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error) {
	args := m.Called(ctx, projectName, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectInfo), args.Error(1)
}

// ValidateProject mocks validating a project
func (m *MockProjectService) ValidateProject(ctx context.Context, projectPath string) error {
	args := m.Called(ctx, projectPath)
	return args.Error(0)
}

// ListProjects mocks listing projects
func (m *MockProjectService) ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ProjectInfo), args.Error(1)
}

// GetProjectInfo mocks getting project info
func (m *MockProjectService) GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error) {
	args := m.Called(ctx, projectPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectInfo), args.Error(1)
}

// Functional setup methods for fluent interface
func (m *MockProjectService) SetupDiscoverProjectSuccess(projectName string) *MockProjectSetup {
	setup := &MockProjectSetup{mock: m}
	return setup.ForDiscoverProject(projectName).WillSucceed()
}

func (m *MockProjectService) SetupDiscoverProjectError(projectName, errorMsg string) *MockProjectSetup {
	setup := &MockProjectSetup{mock: m}
	return setup.ForDiscoverProject(projectName).WillFail(errorMsg)
}

// MockProjectSetup provides functional setup builder for project operations
type MockProjectSetup struct {
	mock      *MockProjectService
	operation string
	params    map[string]interface{}
}

func (s *MockProjectSetup) ForDiscoverProject(projectName string) *MockProjectSetup {
	s.operation = "DiscoverProject"
	s.params = map[string]interface{}{
		"projectName": projectName,
	}
	return s
}

func (s *MockProjectSetup) WithPath(path string) *MockProjectSetup {
	s.params["path"] = path
	return s
}

func (s *MockProjectSetup) WithBranches(branches []string) *MockProjectSetup {
	s.params["branches"] = branches
	return s
}

func (s *MockProjectSetup) WillSucceed() *MockProjectSetup {
	// Set default values if not provided
	path, hasPath := s.params["path"]
	if !hasPath {
		path = "/default/path"
	}

	// Convert string branches to BranchInfo
	var branches []*domain.BranchInfo
	if branchStrings, hasBranches := s.params["branches"]; hasBranches {
		branchStringsList := branchStrings.([]string)
		branches = make([]*domain.BranchInfo, len(branchStringsList))
		for i, branchName := range branchStringsList {
			branches[i] = &domain.BranchInfo{
				Name:      branchName,
				IsCurrent: branchName == "main" || branchName == "master",
			}
		}
	}

	project := &domain.ProjectInfo{
		Name:     s.params["projectName"].(string),
		Path:     path.(string),
		Branches: branches,
	}

	s.mock.On("DiscoverProject", mock.Anything, s.params["projectName"], mock.Anything).Return(project, nil)

	return s
}

func (s *MockProjectSetup) WillFail(errorMsg string) *MockProjectSetup {
	s.mock.On("DiscoverProject", mock.Anything, s.params["projectName"], mock.Anything).Return(nil, fmt.Errorf("%s", errorMsg))

	return s
}
