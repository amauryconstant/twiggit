package service

import (
	"context"
	"errors"
	"os"
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
		Worktrees: []*domain.WorktreeInfo{
			{Path: "/path/to/worktree", Branch: "feature-branch", Commit: "abc123", IsBare: false},
		},
		Branches: []*domain.BranchInfo{
			{Name: "main", IsCurrent: true},
			{Name: "feature-branch", IsCurrent: false},
		},
	}
	s.configureMocks()
	s.service = NewWorktreeService(s.gitService, s.projectService, s.config)
}

func (s *WorktreeServiceTestSuite) configureMocks() {
	s.projectService.On("ListProjects", mock.Anything).Return([]*domain.ProjectInfo{s.testProject}, nil).Maybe()

	testSummary := &domain.ProjectSummary{
		Name:        s.testProject.Name,
		Path:        s.testProject.Path,
		GitRepoPath: s.testProject.GitRepoPath,
	}
	s.projectService.On("ListProjectSummaries", mock.Anything).Return([]*domain.ProjectSummary{testSummary}, nil).Maybe()

	s.projectService.On("DiscoverProject", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*domain.Context")).Return(s.testProject, nil).Maybe()

	s.projectService.On("GetProjectInfo", mock.Anything, "/path/to/project").Return(s.testProject, nil).Maybe()
	s.projectService.On("GetProjectInfo", mock.Anything, "").Return(s.testProject, nil).Maybe()
	s.projectService.On("GetProjectInfo", mock.Anything, mock.AnythingOfType("string")).Return((*domain.ProjectInfo)(nil), nil).Maybe()

	s.projectService.On("ValidateProject", mock.Anything, mock.AnythingOfType("string")).Return(nil).Maybe()

	worktrees := []domain.WorktreeInfo{
		{
			Path:   "/path/to/worktree",
			Branch: "feature-branch",
			Commit: "abc123",
			IsBare: false,
		},
	}
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/path/to/project/.git").Return(worktrees, nil).Maybe()
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{}, nil).Maybe()
	s.gitService.MockCLIClient.On("CreateWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Return(nil).Maybe()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil).Maybe()
	s.gitService.MockCLIClient.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Maybe()
	s.gitService.MockCLIClient.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
	s.gitService.MockGoGitClient.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil).Maybe()
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

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_DryRun() {
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         true,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_ProtectedBranch() {
	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-main", Branch: "main", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil).Maybe()

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: true,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted)
	s.Len(result.ProtectedSkipped, 1)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_UnmergedBranch() {
	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-unmerged", Commit: "abc123"},
	}, nil)
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-unmerged").Return(false, nil)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted)
	s.Len(result.UnmergedSkipped, 1)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_InvalidRequest() {
	req := &domain.PruneWorktreesRequest{
		AllProjects:      true,
		SpecificWorktree: "project/branch",
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot use --all with specific worktree")
	s.Nil(result)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_ForceFlag() {
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil).Once()

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_SingleWorktree() {
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

	req := &domain.PruneWorktreesRequest{
		Context:          &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:           false,
		Force:            false,
		DeleteBranches:   false,
		SpecificWorktree: "test-project/feature-branch",
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_NavigationPath() {
	s.Run("sets navigation path when single worktree deleted with specific worktree and project exists", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		tempDir := s.T().TempDir()
		projectDir := tempDir + "/test-project"
		s.Require().NoError(os.MkdirAll(projectDir, 0755))

		originalProjectsDir := s.config.ProjectsDirectory
		s.config.ProjectsDirectory = tempDir
		defer func() { s.config.ProjectsDirectory = originalProjectsDir }()

		req := &domain.PruneWorktreesRequest{
			Context:          &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:           false,
			Force:            false,
			DeleteBranches:   false,
			SpecificWorktree: "test-project/feature-branch",
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(1, result.TotalDeleted)
		s.NotEmpty(result.NavigationPath)
		s.Contains(result.NavigationPath, "test-project")
	})

	s.Run("does not set navigation path when project path does not exist", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		s.config.ProjectsDirectory = "/nonexistent/path"

		req := &domain.PruneWorktreesRequest{
			Context:          &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:           false,
			Force:            false,
			DeleteBranches:   false,
			SpecificWorktree: "test-project/feature-branch",
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(1, result.TotalDeleted)
		s.Empty(result.NavigationPath)
	})

	s.Run("does not set navigation path without specific worktree", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		req := &domain.PruneWorktreesRequest{
			Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: false,
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(1, result.TotalDeleted)
		s.Empty(result.NavigationPath)
	})

	s.Run("does not set navigation path with multiple deletions", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature1", Branch: "feature-1", Commit: "abc123"},
			{Path: "/path/to/worktree-feature2", Branch: "feature-2", Commit: "def456"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil)

		req := &domain.PruneWorktreesRequest{
			Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: false,
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(2, result.TotalDeleted)
		s.Empty(result.NavigationPath)
	})
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_DeleteBranches() {
	s.Run("deletes branch when DeleteBranches is true", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()
		s.gitService.MockCLIClient.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
		s.gitService.MockCLIClient.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(nil).Once()

		req := &domain.PruneWorktreesRequest{
			Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: true,
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(1, result.TotalDeleted)
		s.Equal(1, result.TotalBranchesDeleted)
		s.Len(result.DeletedWorktrees, 1)
		s.True(result.DeletedWorktrees[0].BranchDeleted)
	})

	s.Run("continues when branch deletion fails", func() {
		s.gitService.MockCLIClient.ExpectedCalls = nil
		s.gitService.MockGoGitClient.ExpectedCalls = nil
		s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()
		s.gitService.MockGoGitClient.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(domain.RepositoryStatus{IsClean: true}, nil).Once()
		s.gitService.MockCLIClient.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
		s.gitService.MockCLIClient.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(errors.New("branch in use")).Once()

		req := &domain.PruneWorktreesRequest{
			Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: true,
		}

		result, err := s.service.PruneMergedWorktrees(context.Background(), req)
		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(1, result.TotalDeleted)
		s.Equal(0, result.TotalBranchesDeleted)
		s.Len(result.DeletedWorktrees, 1)
		s.False(result.DeletedWorktrees[0].BranchDeleted)
		s.Require().Error(result.DeletedWorktrees[0].Error)
		s.Contains(result.DeletedWorktrees[0].Error.Error(), "branch deletion failed")
	})
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_ForceBypassesUncommittedCheck() {
	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil).Once()
	s.gitService.MockGoGitClient.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(domain.RepositoryStatus{
		IsClean:  false,
		Modified: []string{"file.txt"},
	}, nil).Maybe()

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_SkipsUncommittedWithoutForce() {
	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.gitService.MockGoGitClient.ExpectedCalls = nil
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	s.gitService.MockGoGitClient.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(domain.RepositoryStatus{
		IsClean:  false,
		Modified: []string{"file.txt"},
	}, nil).Once()

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TotalDeleted)
	s.Equal(1, result.TotalSkipped)
	s.Len(result.SkippedWorktrees, 1)
	s.Contains(result.SkippedWorktrees[0].SkipReason, "uncommitted changes")
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_AllProjects() {
	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.projectService.ExpectedCalls = nil

	project1 := &domain.ProjectInfo{
		Name:        "project1",
		Path:        "/path/to/project1",
		GitRepoPath: "/path/to/project1/.git",
	}
	project2 := &domain.ProjectInfo{
		Name:        "project2",
		Path:        "/path/to/project2",
		GitRepoPath: "/path/to/project2/.git",
	}

	summaries := []*domain.ProjectSummary{
		{Name: project1.Name, Path: project1.Path, GitRepoPath: project1.GitRepoPath},
		{Name: project2.Name, Path: project2.Path, GitRepoPath: project2.GitRepoPath},
	}
	s.projectService.On("ListProjectSummaries", mock.Anything).Return(summaries, nil).Once()
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/path/to/project1/.git").Return([]domain.WorktreeInfo{
		{Path: "/path/to/wt1", Branch: "feature-1", Commit: "abc123"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, "/path/to/project2/.git").Return([]domain.WorktreeInfo{
		{Path: "/path/to/wt2", Branch: "feature-2", Commit: "def456"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil)

	req := &domain.PruneWorktreesRequest{
		Context:     &domain.Context{Type: domain.ContextOutsideGit},
		AllProjects: true,
		DryRun:      false,
		Force:       false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(2, result.TotalDeleted)
}

func (s *WorktreeServiceTestSuite) TestPruneMergedWorktrees_CurrentWorktreeSkipped() {
	tempDir := s.T().TempDir()
	worktreeCurrentPath := tempDir + "/worktree-current"
	worktreeOtherPath := tempDir + "/worktree-other"

	s.gitService.MockCLIClient.ExpectedCalls = nil
	s.gitService.MockCLIClient.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]domain.WorktreeInfo{
		{Path: worktreeCurrentPath, Branch: "feature-current", Commit: "abc123"},
		{Path: worktreeOtherPath, Branch: "feature-other", Commit: "def456"},
	}, nil).Once()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-current").Return(true, nil).Maybe()
	s.gitService.MockCLIClient.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-other").Return(true, nil).Maybe()
	s.gitService.MockCLIClient.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), worktreeOtherPath, mock.AnythingOfType("bool")).Return(nil).Maybe()

	originalWd, err := os.Getwd()
	s.Require().NoError(err)
	defer os.Chdir(originalWd)

	err = os.MkdirAll(worktreeCurrentPath, 0755)
	s.Require().NoError(err)
	err = os.Chdir(worktreeCurrentPath)
	s.Require().NoError(err)

	req := &domain.PruneWorktreesRequest{
		Context:        &domain.Context{Type: domain.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := s.service.PruneMergedWorktrees(context.Background(), req)
	s.Require().NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TotalDeleted, "should delete non-current worktree")
	s.Equal(1, result.TotalSkipped, "should skip current worktree")
	s.Len(result.CurrentWorktreeSkipped, 1)
	s.Equal("feature-current", result.CurrentWorktreeSkipped[0].BranchName)
	s.Contains(result.CurrentWorktreeSkipped[0].SkipReason, "cannot prune current worktree")
}
