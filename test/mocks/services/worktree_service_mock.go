package services

import (
	"context"
	"fmt"

	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockWorktreeService provides functional setup builders for worktree service testing
type MockWorktreeService struct {
	mock.Mock
}

// NewMockWorktreeService creates a new MockWorktreeService instance
func NewMockWorktreeService() *MockWorktreeService {
	return &MockWorktreeService{}
}

// CreateWorktree mocks creating a worktree
func (m *MockWorktreeService) CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorktreeInfo), args.Error(1)
}

// DeleteWorktree mocks deleting a worktree
func (m *MockWorktreeService) DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// ListWorktrees mocks listing worktrees
func (m *MockWorktreeService) ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WorktreeInfo), args.Error(1)
}

// GetWorktreeStatus mocks getting worktree status
func (m *MockWorktreeService) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error) {
	args := m.Called(ctx, worktreePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorktreeStatus), args.Error(1)
}

// ValidateWorktree mocks validating a worktree
func (m *MockWorktreeService) ValidateWorktree(ctx context.Context, worktreePath string) error {
	args := m.Called(ctx, worktreePath)
	return args.Error(0)
}

// Functional setup methods for fluent interface
func (m *MockWorktreeService) SetupCreateWorktreeSuccess(projectName, branchName string) *MockWorktreeSetup {
	setup := &MockWorktreeSetup{mock: m}
	return setup.ForCreateWorktree(projectName, branchName).WillSucceed()
}

func (m *MockWorktreeService) SetupCreateWorktreeError(projectName, branchName, errorMsg string) *MockWorktreeSetup {
	setup := &MockWorktreeSetup{mock: m}
	return setup.ForCreateWorktree(projectName, branchName).WillFail(errorMsg)
}

// MockWorktreeSetup provides functional setup builder for worktree operations
type MockWorktreeSetup struct {
	mock      *MockWorktreeService
	operation string
	params    map[string]interface{}
}

func (s *MockWorktreeSetup) ForCreateWorktree(projectName, branchName string) *MockWorktreeSetup {
	s.operation = "CreateWorktree"
	s.params = map[string]interface{}{
		"projectName": projectName,
		"branchName":  branchName,
	}
	return s
}

func (s *MockWorktreeSetup) WithPath(path string) *MockWorktreeSetup {
	s.params["path"] = path
	return s
}

func (s *MockWorktreeSetup) WithBranch(branch string) *MockWorktreeSetup {
	s.params["branch"] = branch
	return s
}

func (s *MockWorktreeSetup) WillSucceed() *MockWorktreeSetup {
	// Set default values if not provided
	path, hasPath := s.params["path"]
	if !hasPath {
		path = "/default/path"
	}
	branch, hasBranch := s.params["branch"]
	if !hasBranch {
		branch = s.params["branchName"].(string)
	}

	worktree := &domain.WorktreeInfo{
		Path:   path.(string),
		Branch: branch.(string),
	}

	s.mock.On("CreateWorktree", mock.Anything, mock.MatchedBy(func(req *domain.CreateWorktreeRequest) bool {
		return req.ProjectName == s.params["projectName"] && req.BranchName == s.params["branchName"]
	})).Return(worktree, nil)

	return s
}

func (s *MockWorktreeSetup) WillFail(errorMsg string) *MockWorktreeSetup {
	s.mock.On("CreateWorktree", mock.Anything, mock.MatchedBy(func(req *domain.CreateWorktreeRequest) bool {
		return req.ProjectName == s.params["projectName"] && req.BranchName == s.params["branchName"]
	})).Return(nil, fmt.Errorf("%s", errorMsg))

	return s
}
