package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/core"
	"twiggit/test/mocks"
)

func configureWorktreeServiceMocks(goGit *mocks.MockGoGitClient, cli *mocks.MockCLIClient, projectService *mocks.MockProjectService, testProject *core.ProjectInfo) {
	projectService.On("ListProjects", mock.Anything).Return([]*core.ProjectInfo{testProject}, nil).Maybe()

	testSummary := &core.ProjectSummary{
		Name:        testProject.Name,
		Path:        testProject.Path,
		GitRepoPath: testProject.GitRepoPath,
	}
	projectService.On("ListProjectSummaries", mock.Anything).Return([]*core.ProjectSummary{testSummary}, nil).Maybe()

	projectService.On("DiscoverProject", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*core.Context")).Return(testProject, nil).Maybe()

	projectService.On("GetProjectInfo", mock.Anything, "/path/to/project").Return(testProject, nil).Maybe()
	projectService.On("GetProjectInfo", mock.Anything, "").Return(testProject, nil).Maybe()
	projectService.On("GetProjectInfo", mock.Anything, mock.AnythingOfType("string")).Return((*core.ProjectInfo)(nil), nil).Maybe()

	projectService.On("ValidateProject", mock.Anything, mock.AnythingOfType("string")).Return(nil).Maybe()

	worktrees := []core.WorktreeInfo{
		{
			Path:   "/path/to/worktree",
			Branch: "feature-branch",
			Commit: "abc123",
		},
	}
	cli.On("ListWorktrees", mock.Anything, "/path/to/project/.git").Return(worktrees, nil).Maybe()
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{}, nil).Maybe()
	cli.On("CreateWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Return(nil).Maybe()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil).Maybe()
	cli.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Maybe()
	cli.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
	goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil).Maybe()
	goGit.On("BranchExists", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil).Maybe()

	goGit.On("GetRepositoryStatus", mock.Anything, mock.AnythingOfType("string")).Return(core.RepositoryStatus{
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

func setupWorktreeService() (application.WorktreeService, *mocks.MockGoGitClient, *mocks.MockCLIClient, *mocks.MockProjectService, *core.Config) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	projectService := mocks.NewMockProjectService()
	testProject := &core.ProjectInfo{
		Name:        "test-project",
		Path:        "/path/to/project",
		GitRepoPath: "/path/to/project/.git",
		Worktrees: []*core.WorktreeInfo{
			{Path: "/path/to/worktree", Branch: "feature-branch", Commit: "abc123"},
		},
		Branches: []*core.BranchInfo{
			{Name: "main", IsCurrent: true},
			{Name: "feature-branch", IsCurrent: false},
		},
	}
	configureWorktreeServiceMocks(goGit, cli, projectService, testProject)
	service := NewWorktreeService(goGit, cli, projectService, config, nil)

	return service, goGit, cli, projectService, config
}

func TestWorktreeService_CreateWorktree(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

	tests := []struct {
		name         string
		request      *core.CreateWorktreeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid worktree creation",
			request: &core.CreateWorktreeRequest{
				ProjectName:  "test-project",
				BranchName:   "feature-branch",
				SourceBranch: "main",
				Context: &core.Context{
					Type:        core.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError: false,
		},
		{
			name: "empty branch name",
			request: &core.CreateWorktreeRequest{
				ProjectName:  "test-project",
				BranchName:   "",
				SourceBranch: "main",
				Context: &core.Context{
					Type:        core.ContextProject,
					ProjectName: "test-project",
				},
			},
			expectError:  true,
			errorMessage: "branch name is required",
		},
		{
			name: "empty project name",
			request: &core.CreateWorktreeRequest{
				ProjectName:  "",
				BranchName:   "feature-branch",
				SourceBranch: "main",
				Context: &core.Context{
					Type: core.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "project name required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CreateWorktree(context.Background(), tc.request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.BranchName, result.Worktree.Branch)
			}
		})
	}
}

func TestWorktreeService_DeleteWorktree(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

	tests := []struct {
		name         string
		request      *core.DeleteWorktreeRequest
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid worktree deletion",
			request: &core.DeleteWorktreeRequest{
				WorktreePath: "/path/to/worktree",
				Force:        false,
				Context: &core.Context{
					Type: core.ContextWorktree,
				},
			},
			expectError: false,
		},
		{
			name: "empty worktree path",
			request: &core.DeleteWorktreeRequest{
				WorktreePath: "",
				Force:        false,
				Context: &core.Context{
					Type: core.ContextOutsideGit,
				},
			},
			expectError:  true,
			errorMessage: "worktree path cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestWorktreeService_DeleteWorktree_Idempotent(t *testing.T) {
	t.Run("non-existent worktree should succeed", func(t *testing.T) {
		service, _, _, _, _ := setupWorktreeService()
		request := &core.DeleteWorktreeRequest{
			WorktreePath: "/non/existent/worktree",
			Force:        false,
			Context: &core.Context{
				Type: core.ContextOutsideGit,
			},
		}

		err := service.DeleteWorktree(context.Background(), request)
		assert.NoError(t, err, "DeleteWorktree should be idempotent and succeed for non-existent worktrees")
	})
}

func TestWorktreeService_ListWorktrees(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

	tests := []struct {
		name        string
		request     *core.ListWorktreesRequest
		expectError bool
	}{
		{
			name: "valid worktree listing",
			request: &core.ListWorktreesRequest{
				ProjectName: "test-project",
				Context: &core.Context{
					Type:        core.ContextProject,
					ProjectName: "test-project",
				},
				IncludeMain: true,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestWorktreeService_GetWorktreeStatus(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

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
		t.Run(tc.name, func(t *testing.T) {
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

func TestWorktreeService_ValidateWorktree(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

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
		t.Run(tc.name, func(t *testing.T) {
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

func TestWorktreeService_PruneMergedWorktrees_DryRun(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         true,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalDeleted)
}

func TestWorktreeService_PruneMergedWorktrees_ProtectedBranch(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	cli.ExpectedCalls = nil
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-main", Branch: "main", Commit: "abc123"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil).Maybe()

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: true,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalDeleted)
	assert.Len(t, result.ProtectedSkipped, 1)
}

func TestWorktreeService_PruneMergedWorktrees_UnmergedBranch(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	cli.ExpectedCalls = nil
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-unmerged", Commit: "abc123"},
	}, nil)
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-unmerged").Return(false, nil)

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalDeleted)
	assert.Len(t, result.UnmergedSkipped, 1)
}

func TestWorktreeService_PruneMergedWorktrees_InvalidRequest(t *testing.T) {
	service, _, _, _, _ := setupWorktreeService()

	req := &core.PruneWorktreesRequest{
		AllProjects:      true,
		SpecificWorktree: "project/branch",
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --all with specific worktree")
	assert.Nil(t, result)
}

func TestWorktreeService_PruneMergedWorktrees_ForceFlag(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil).Once()

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalDeleted)
}

func TestWorktreeService_PruneMergedWorktrees_SingleWorktree(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

	req := &core.PruneWorktreesRequest{
		Context:          &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:           false,
		Force:            false,
		DeleteBranches:   false,
		SpecificWorktree: "test-project/feature-branch",
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalDeleted)
}

func TestWorktreeService_PruneMergedWorktrees_NavigationPath(t *testing.T) {
	t.Run("sets navigation path when single worktree deleted with specific worktree and project exists", func(t *testing.T) {
		service, _, cli, _, config := setupWorktreeService()

		cli.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		tempDir := t.TempDir()
		projectDir := tempDir + "/test-project"
		require.NoError(t, os.MkdirAll(projectDir, 0755))

		originalProjectsDir := config.ProjectsDirectory
		config.ProjectsDirectory = tempDir
		defer func() { config.ProjectsDirectory = originalProjectsDir }()

		req := &core.PruneWorktreesRequest{
			Context:          &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:           false,
			Force:            false,
			DeleteBranches:   false,
			SpecificWorktree: "test-project/feature-branch",
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalDeleted)
		assert.NotEmpty(t, result.NavigationPath)
		assert.Contains(t, result.NavigationPath, "test-project")
	})

	t.Run("does not set navigation path when project path does not exist", func(t *testing.T) {
		service, _, cli, _, config := setupWorktreeService()

		cli.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		config.ProjectsDirectory = "/nonexistent/path"

		req := &core.PruneWorktreesRequest{
			Context:          &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:           false,
			Force:            false,
			DeleteBranches:   false,
			SpecificWorktree: "test-project/feature-branch",
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalDeleted)
		assert.Empty(t, result.NavigationPath)
	})

	t.Run("does not set navigation path without specific worktree", func(t *testing.T) {
		service, _, cli, _, _ := setupWorktreeService()

		cli.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()

		req := &core.PruneWorktreesRequest{
			Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: false,
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalDeleted)
		assert.Empty(t, result.NavigationPath)
	})

	t.Run("does not set navigation path with multiple deletions", func(t *testing.T) {
		service, _, cli, _, _ := setupWorktreeService()

		cli.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature1", Branch: "feature-1", Commit: "abc123"},
			{Path: "/path/to/worktree-feature2", Branch: "feature-2", Commit: "def456"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil)

		req := &core.PruneWorktreesRequest{
			Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: false,
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalDeleted)
		assert.Empty(t, result.NavigationPath)
	})
}

func TestWorktreeService_PruneMergedWorktrees_DeleteBranches(t *testing.T) {
	t.Run("deletes branch when DeleteBranches is true", func(t *testing.T) {
		service, _, cli, _, _ := setupWorktreeService()

		cli.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()
		cli.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
		cli.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(nil).Once()

		req := &core.PruneWorktreesRequest{
			Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: true,
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalDeleted)
		assert.Equal(t, 1, result.TotalBranchesDeleted)
		assert.Len(t, result.DeletedWorktrees, 1)
		assert.True(t, result.DeletedWorktrees[0].BranchDeleted)
	})

	t.Run("continues when branch deletion fails", func(t *testing.T) {
		service, goGit, cli, _, _ := setupWorktreeService()

		cli.ExpectedCalls = nil
		goGit.ExpectedCalls = nil
		cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
			{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
		}, nil).Once()
		cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
		cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil).Once()
		goGit.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(core.RepositoryStatus{IsClean: true}, nil).Once()
		cli.On("PruneWorktrees", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
		cli.On("DeleteBranch", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(errors.New("branch in use")).Once()

		req := &core.PruneWorktreesRequest{
			Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
			DryRun:         false,
			Force:          false,
			DeleteBranches: true,
		}

		result, err := service.PruneMergedWorktrees(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalDeleted)
		assert.Equal(t, 0, result.TotalBranchesDeleted)
		assert.Len(t, result.DeletedWorktrees, 1)
		assert.False(t, result.DeletedWorktrees[0].BranchDeleted)
		require.Error(t, result.DeletedWorktrees[0].Error)
		assert.Contains(t, result.DeletedWorktrees[0].Error.Error(), "branch deletion failed")
	})
}

func TestWorktreeService_PruneMergedWorktrees_ForceBypassesUncommittedCheck(t *testing.T) {
	service, goGit, cli, _, _ := setupWorktreeService()

	cli.ExpectedCalls = nil
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), true).Return(nil).Once()
	goGit.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(core.RepositoryStatus{
		IsClean:  false,
		Modified: []string{"file.txt"},
	}, nil).Maybe()

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalDeleted)
}

func TestWorktreeService_PruneMergedWorktrees_SkipsUncommittedWithoutForce(t *testing.T) {
	service, goGit, cli, _, _ := setupWorktreeService()

	cli.ExpectedCalls = nil
	goGit.ExpectedCalls = nil
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: "/path/to/worktree-feature", Branch: "feature-branch", Commit: "abc123"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()
	goGit.On("GetRepositoryStatus", mock.Anything, "/path/to/worktree-feature").Return(core.RepositoryStatus{
		IsClean:  false,
		Modified: []string{"file.txt"},
	}, nil).Once()

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          false,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalDeleted)
	assert.Equal(t, 1, result.TotalSkipped)
	assert.Len(t, result.SkippedWorktrees, 1)
	assert.Contains(t, result.SkippedWorktrees[0].SkipReason, "uncommitted changes")
}

func TestWorktreeService_PruneMergedWorktrees_AllProjects(t *testing.T) {
	service, _, cli, projectService, _ := setupWorktreeService()

	cli.ExpectedCalls = nil
	projectService.ExpectedCalls = nil

	project1 := &core.ProjectInfo{
		Name:        "project1",
		Path:        "/path/to/project1",
		GitRepoPath: "/path/to/project1/.git",
	}
	project2 := &core.ProjectInfo{
		Name:        "project2",
		Path:        "/path/to/project2",
		GitRepoPath: "/path/to/project2/.git",
	}

	summaries := []*core.ProjectSummary{
		{Name: project1.Name, Path: project1.Path, GitRepoPath: project1.GitRepoPath},
		{Name: project2.Name, Path: project2.Path, GitRepoPath: project2.GitRepoPath},
	}
	projectService.On("ListProjectSummaries", mock.Anything).Return(summaries, nil).Once()
	cli.On("ListWorktrees", mock.Anything, "/path/to/project1/.git").Return([]core.WorktreeInfo{
		{Path: "/path/to/wt1", Branch: "feature-1", Commit: "abc123"},
	}, nil).Once()
	cli.On("ListWorktrees", mock.Anything, "/path/to/project2/.git").Return([]core.WorktreeInfo{
		{Path: "/path/to/wt2", Branch: "feature-2", Commit: "def456"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), false).Return(nil)

	req := &core.PruneWorktreesRequest{
		Context:     &core.Context{Type: core.ContextOutsideGit},
		AllProjects: true,
		DryRun:      false,
		Force:       false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalDeleted)
}

func TestWorktreeService_PruneMergedWorktrees_CurrentWorktreeSkipped(t *testing.T) {
	service, _, cli, _, _ := setupWorktreeService()

	tempDir := t.TempDir()
	worktreeCurrentPath := tempDir + "/worktree-current"
	worktreeOtherPath := tempDir + "/worktree-other"

	cli.ExpectedCalls = nil
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{
		{Path: worktreeCurrentPath, Branch: "feature-current", Commit: "abc123"},
		{Path: worktreeOtherPath, Branch: "feature-other", Commit: "def456"},
	}, nil).Once()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-current").Return(true, nil).Maybe()
	cli.On("IsBranchMerged", mock.Anything, mock.AnythingOfType("string"), "feature-other").Return(true, nil).Maybe()
	cli.On("DeleteWorktree", mock.Anything, mock.AnythingOfType("string"), worktreeOtherPath, mock.AnythingOfType("bool")).Return(nil).Maybe()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	err = os.MkdirAll(worktreeCurrentPath, 0755)
	require.NoError(t, err)
	err = os.Chdir(worktreeCurrentPath)
	require.NoError(t, err)

	req := &core.PruneWorktreesRequest{
		Context:        &core.Context{Type: core.ContextProject, ProjectName: "test-project", Path: "/path/to/project"},
		DryRun:         false,
		Force:          true,
		DeleteBranches: false,
	}

	result, err := service.PruneMergedWorktrees(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalDeleted, "should delete non-current worktree")
	assert.Equal(t, 1, result.TotalSkipped, "should skip current worktree")
	assert.Len(t, result.CurrentWorktreeSkipped, 1)
	assert.Equal(t, "feature-current", result.CurrentWorktreeSkipped[0].BranchName)
	assert.Contains(t, result.CurrentWorktreeSkipped[0].SkipReason, "cannot prune current worktree")
}

func TestWorktreeService_Routing_BranchExistsUsesGoGitOnly(t *testing.T) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	projectService := mocks.NewMockProjectService()
	testProject := &core.ProjectInfo{
		Name:        "test-project",
		Path:        "/path/to/project",
		GitRepoPath: "/path/to/project/.git",
	}
	configureWorktreeServiceMocks(goGit, cli, projectService, testProject)
	service := NewWorktreeService(goGit, cli, projectService, config, nil)

	goGit.ExpectedCalls = nil
	cli.ExpectedCalls = nil
	goGit.Calls = nil
	cli.Calls = nil

	projectService.On("ListProjects", mock.Anything).Return([]*core.ProjectInfo{testProject}, nil).Maybe()
	projectService.On("ListProjectSummaries", mock.Anything).Return([]*core.ProjectSummary{{
		Name: testProject.Name, Path: testProject.Path, GitRepoPath: testProject.GitRepoPath,
	}}, nil).Maybe()
	projectService.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*core.Context")).Return(testProject, nil).Maybe()
	projectService.On("ValidateProject", mock.Anything, mock.AnythingOfType("string")).Return(nil).Maybe()

	goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil).Once()
	goGit.On("BranchExists", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(true, nil).Once()

	_, err := service.BranchExists(context.Background(), testProject.Path, "feature-branch")
	require.NoError(t, err)

	goGit.AssertCalled(t, "BranchExists", mock.Anything, mock.AnythingOfType("string"), "feature-branch")
	cli.AssertNotCalled(t, "BranchExists", mock.Anything, mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "DeleteWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "ListWorktrees", mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "PruneWorktrees", mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "IsBranchMerged", mock.Anything, mock.Anything, mock.Anything)
	cli.AssertNotCalled(t, "DeleteBranch", mock.Anything, mock.Anything, mock.Anything)
}

func TestWorktreeService_Routing_CreateWorktreeUsesCLIOnly(t *testing.T) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	projectService := mocks.NewMockProjectService()
	testProject := &core.ProjectInfo{
		Name:        "test-project",
		Path:        "/path/to/project",
		GitRepoPath: "/path/to/project/.git",
	}
	configureWorktreeServiceMocks(goGit, cli, projectService, testProject)
	service := NewWorktreeService(goGit, cli, projectService, config, nil)

	goGit.ExpectedCalls = nil
	cli.ExpectedCalls = nil
	goGit.Calls = nil
	cli.Calls = nil

	projectService.On("DiscoverProject", mock.Anything, "test-project", mock.AnythingOfType("*core.Context")).Return(testProject, nil).Once()
	projectService.On("ValidateProject", mock.Anything, testProject.Path).Return(nil).Once()
	goGit.On("BranchExists", mock.Anything, mock.AnythingOfType("string"), "feature-branch").Return(false, nil).Once()
	cli.On("CreateWorktree", mock.Anything, testProject.GitRepoPath, "feature-branch", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()

	req := &core.CreateWorktreeRequest{
		ProjectName:  "test-project",
		BranchName:   "feature-branch",
		SourceBranch: "main",
		Context: &core.Context{
			Type:        core.ContextProject,
			ProjectName: "test-project",
			Path:        testProject.Path,
		},
	}

	_, err := service.CreateWorktree(context.Background(), req)
	require.NoError(t, err)

	cli.AssertCalled(t, "CreateWorktree", mock.Anything, testProject.GitRepoPath, "feature-branch", mock.AnythingOfType("string"), mock.AnythingOfType("string"))
	goGit.AssertNotCalled(t, "CreateWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	goGit.AssertNotCalled(t, "DeleteWorktree", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	goGit.AssertNotCalled(t, "ListWorktrees", mock.Anything, mock.Anything)
	goGit.AssertNotCalled(t, "PruneWorktrees", mock.Anything, mock.Anything)
	goGit.AssertNotCalled(t, "IsBranchMerged", mock.Anything, mock.Anything, mock.Anything)
	goGit.AssertNotCalled(t, "DeleteBranch", mock.Anything, mock.Anything, mock.Anything)
}
