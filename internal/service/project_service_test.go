package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	contextService := mocks.NewMockContextService()
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

func TestProjectService_searchProjectByName(t *testing.T) {
	testCases := []struct {
		name           string
		setupFunc      func(t *testing.T) (*projectService, string)
		projectName    string
		expectError    bool
		errorMessage   string
		validateResult func(t *testing.T, result *domain.ProjectInfo)
	}{
		{
			name: "directory not found error",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "nonexistent")
				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}
				service := &projectService{
					config: config,
				}
				return service, projectsDir
			},
			projectName:  "testproject",
			expectError:  true,
			errorMessage: "project not found",
		},
		{
			name: "case-insensitive matching",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				require.NoError(t, os.Mkdir(projectsDir, 0755))

				projectPath := filepath.Join(projectsDir, "testproject")
				require.NoError(t, os.Mkdir(projectPath, 0755))

				gitDir := filepath.Join(projectPath, ".git")
				require.NoError(t, os.Mkdir(gitDir, 0755))

				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}
				gitService := mocks.NewMockGitService()
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

				service := &projectService{
					gitService: gitService,
					config:     config,
				}
				return service, projectsDir
			},
			projectName: "TestProject",
			expectError: false,
			validateResult: func(t *testing.T, result *domain.ProjectInfo) {
				t.Helper()
				assert.Equal(t, "testproject", result.Name)
			},
		},
		{
			name: "exact match works",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				require.NoError(t, os.Mkdir(projectsDir, 0755))

				projectPath := filepath.Join(projectsDir, "myproject")
				require.NoError(t, os.Mkdir(projectPath, 0755))

				gitDir := filepath.Join(projectPath, ".git")
				require.NoError(t, os.Mkdir(gitDir, 0755))

				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}
				gitService := mocks.NewMockGitService()
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

				service := &projectService{
					gitService: gitService,
					config:     config,
				}
				return service, projectsDir
			},
			projectName: "myproject",
			expectError: false,
			validateResult: func(t *testing.T, result *domain.ProjectInfo) {
				t.Helper()
				assert.Equal(t, "myproject", result.Name)
			},
		},
		{
			name: "multiple matches returns first one",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				require.NoError(t, os.Mkdir(projectsDir, 0755))

				for _, name := range []string{"aproject", "bproject", "cproject"} {
					projectPath := filepath.Join(projectsDir, name)
					require.NoError(t, os.Mkdir(projectPath, 0755))

					gitDir := filepath.Join(projectPath, ".git")
					require.NoError(t, os.Mkdir(gitDir, 0755))
				}

				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}
				gitService := mocks.NewMockGitService()
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

				service := &projectService{
					gitService: gitService,
					config:     config,
				}
				return service, projectsDir
			},
			projectName: "aproject",
			expectError: false,
			validateResult: func(t *testing.T, result *domain.ProjectInfo) {
				t.Helper()
				assert.Equal(t, "aproject", result.Name)
			},
		},
		{
			name: "empty directory returns not found",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				require.NoError(t, os.Mkdir(projectsDir, 0755))

				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}

				service := &projectService{
					config: config,
				}
				return service, projectsDir
			},
			projectName:  "nonexistent",
			expectError:  true,
			errorMessage: "project not found",
		},
		{
			name: "read directory error",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "projects")

				require.NoError(t, os.WriteFile(projectsDir, []byte("not a directory"), 0644))

				config := &domain.Config{
					ProjectsDirectory: projectsDir,
				}

				service := &projectService{
					config: config,
				}
				return service, projectsDir
			},
			projectName:  "testproject",
			expectError:  true,
			errorMessage: "failed to search projects",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := tc.setupFunc(t)

			result, err := service.searchProjectByName(context.Background(), tc.projectName)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}
		})
	}
}

func TestProjectService_findMainRepoFromWorktree(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) (worktreePath string, expectedPath string)
	}{
		{
			name: "main repo found at parent directory",
			setupFunc: func(t *testing.T) (string, string) {
				t.Helper()
				tempDir := t.TempDir()

				mainRepoPath := tempDir
				gitDir := filepath.Join(mainRepoPath, ".git")
				require.NoError(t, os.Mkdir(gitDir, 0755))

				headsDir := filepath.Join(gitDir, "heads")
				require.NoError(t, os.Mkdir(headsDir, 0755))

				worktreePath := filepath.Join(tempDir, "worktree")
				require.NoError(t, os.Mkdir(worktreePath, 0755))

				gitFileContent := fmt.Sprintf("gitdir: %s\n", filepath.ToSlash(gitDir))
				gitFilePath := filepath.Join(worktreePath, ".git")
				require.NoError(t, os.WriteFile(gitFilePath, []byte(gitFileContent), 0644))

				return worktreePath, tempDir
			},
		},
		{
			name: "root directory stops at root",
			setupFunc: func(t *testing.T) (string, string) {
				t.Helper()
				tempDir := t.TempDir()
				return tempDir, tempDir
			},
		},
		{
			name: "no main repo found returns input path",
			setupFunc: func(t *testing.T) (string, string) {
				t.Helper()
				tempDir := t.TempDir()

				subdirPath := filepath.Join(tempDir, "subdir1", "subdir2")
				require.NoError(t, os.MkdirAll(subdirPath, 0755))

				return subdirPath, subdirPath
			},
		},
		{
			name: "main repo found at intermediate directory",
			setupFunc: func(t *testing.T) (string, string) {
				t.Helper()
				tempDir := t.TempDir()

				mainRepoPath := filepath.Join(tempDir, "main")
				require.NoError(t, os.Mkdir(mainRepoPath, 0755))

				gitDir := filepath.Join(mainRepoPath, ".git")
				require.NoError(t, os.Mkdir(gitDir, 0755))

				subdirPath := filepath.Join(mainRepoPath, "subdir1", "subdir2")
				require.NoError(t, os.MkdirAll(subdirPath, 0755))

				return subdirPath, mainRepoPath
			},
		},
		{
			name: "worktree directory returns worktree path if no main repo",
			setupFunc: func(t *testing.T) (string, string) {
				t.Helper()
				tempDir := t.TempDir()

				worktreePath := filepath.Join(tempDir, "worktree")
				require.NoError(t, os.Mkdir(worktreePath, 0755))

				return worktreePath, worktreePath
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			worktreePath, expectedPath := tc.setupFunc(t)

			service := &projectService{}

			result := service.findMainRepoFromWorktree(worktreePath)

			assert.Equal(t, expectedPath, result)
		})
	}
}
