package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

type WorktreeServiceTestSuite struct {
	suite.Suite
	service        application.WorktreeService
	gitService     *mocks.MockGitService
	projectService *mocks.MockProjectService
	config         *domain.Config
	testProject    *domain.ProjectInfo
}

func (s *WorktreeServiceTestSuite) SetupTest() {
	s.config = domain.DefaultConfig()
	s.gitService = mocks.NewMockGitService()
	s.projectService = mocks.NewMockProjectService()
	s.testProject = &domain.ProjectInfo{
		Name:        "test-project",
		Path:        "/path/to/project",
		GitRepoPath: "/path/to/project/.git",
		Branches: []*domain.BranchInfo{
			{Name: "main", IsCurrent: true},
			{Name: "feature-branch", IsCurrent: false},
		},
	}
	s.configureMocks()
	s.service = NewWorktreeService(s.gitService, s.projectService, s.config)
}

func (s *WorktreeServiceTestSuite) configureMocks() {
	s.projectService.On("ListProjects", mock.Anything).Return([]*domain.ProjectInfo{s.testProject}, nil)

	s.projectService.On("DiscoverProject", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.Context")).Return(s.testProject, nil)

	s.projectService.On("GetProjectInfo", mock.Anything, "/path/to/project").Return(s.testProject, nil)
	s.projectService.On("GetProjectInfo", mock.Anything, "").Return(s.testProject, nil)
	s.projectService.On("GetProjectInfo", mock.Anything, mock.AnythingOfType("string")).Return((*domain.ProjectInfo)(nil), nil).Maybe()

	s.projectService.On("ValidateProject", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	worktrees := []domain.WorktreeInfo{
		{
			Path:   "/path/to/worktree",
			Branch: "feature-branch",
			Commit: "abc123",
			IsBare: false,
		},
	}
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/path/to/project/.git").Return(worktrees, nil)
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{}, nil).Maybe()
	s.gitService.MockCLIClient.On("CreateWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Return(nil).Maybe()
	s.gitService.MockGoGitClient.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)
	s.gitService.MockGoGitClient.On("BranchExists", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil).Maybe()

	s.gitService.MockGoGitClient.On("GetRepositoryStatus", mock.Anything, mock.AnythingOfType("string")).Return(domain.RepositoryStatus{
		IsClean:   true,
		Branch:    "feature-branch",
		Commit:    "abc123",
		Modified:  []string{},
		Added:     []string{},
		Deleted:   []string{},
		Untracked: []string{},
		Ahead:     0,
		Behind:    0,
	}, nil).Maybe()
}

func TestWorktreeService(t *testing.T) {
	suite.Run(t, new(WorktreeServiceTestSuite))
}

func (s *WorktreeServiceTestSuite) TestCreateWorktree() {
	tests := []struct {
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
			errorMessage: "branch name is required",
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.CreateWorktree(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
				s.Equal(tc.request.BranchName, result.Branch)
			}
		})
	}
}

func (s *WorktreeServiceTestSuite) TestDeleteWorktree() {
	tests := []struct {
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.service.DeleteWorktree(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *WorktreeServiceTestSuite) TestDeleteWorktree_Idempotent() {
	s.Run("non-existent worktree should succeed", func() {
		request := &domain.DeleteWorktreeRequest{
			WorktreePath: "/non/existent/worktree",
			Force:        false,
			Context: &domain.Context{
				Type: domain.ContextOutsideGit,
			},
		}

		err := s.service.DeleteWorktree(context.Background(), request)
		s.Require().NoError(err, "DeleteWorktree should be idempotent and succeed for non-existent worktrees")
	})
}

func (s *WorktreeServiceTestSuite) TestListWorktrees() {
	tests := []struct {
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
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.ListWorktrees(context.Background(), tc.request)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *WorktreeServiceTestSuite) TestGetWorktreeStatus() {
	tests := []struct {
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.GetWorktreeStatus(context.Background(), tc.worktreePath)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *WorktreeServiceTestSuite) TestValidateWorktree() {
	tests := []struct {
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.service.ValidateWorktree(context.Background(), tc.worktreePath)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
