package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/application"
	"twiggit/internal/core"
	"twiggit/test/mocks"
)

type stubRepoLocator struct {
	dirs []core.GitDir
	err  error
}

func (s *stubRepoLocator) FindGitRepositories(dir string) ([]core.GitDir, error) {
	return s.dirs, s.err
}

func configureGitMock(goGit *mocks.MockGoGitClient, cli *mocks.MockCLIClient) {
	goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)
	goGit.On("GetRepositoryInfo", mock.Anything, mock.AnythingOfType("string")).Return(&core.GitRepository{
		Path:          "/path/to/project",
		IsBare:        false,
		DefaultBranch: "main",
		Remotes:       []core.RemoteInfo{},
		Branches:      []core.BranchInfo{},
		Worktrees:     []core.WorktreeInfo{},
		Status:        core.RepositoryStatus{},
	}, nil).Maybe()
	cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{}, nil)
}

func TestProjectService_DiscoverProject(t *testing.T) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	configureGitMock(goGit, cli)
	repoLocator := &stubRepoLocator{}
	service := NewProjectService(goGit, cli, repoLocator, config)

	tests := []struct {
		name         string
		projectName  string
		context      *core.Context
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project discovery",
			projectName: "test-project",
			context: &core.Context{
				Type: core.ContextOutsideGit,
			},
			expectError: false,
		},
		{
			name:        "empty project name outside context",
			projectName: "",
			context: &core.Context{
				Type: core.ContextOutsideGit,
			},
			expectError:  true,
			errorMessage: "project name required when outside git context",
		},
		{
			name:        "project discovery from project context",
			projectName: "",
			context: &core.Context{
				Type:        core.ContextProject,
				ProjectName: "project",
				Path:        "/path/to/project",
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.DiscoverProject(context.Background(), tc.projectName, tc.context)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				expectedName := tc.projectName
				if expectedName == "" && tc.context != nil {
					expectedName = tc.context.ProjectName
				}
				assert.Equal(t, expectedName, result.Name)
			}
		})
	}
}

func TestProjectService_ValidateProject(t *testing.T) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	configureGitMock(goGit, cli)
	repoLocator := &stubRepoLocator{}
	service := NewProjectService(goGit, cli, repoLocator, config)

	tests := []struct {
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestProjectService_ListProjects(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "demo")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	config := &core.Config{ProjectsDirectory: tempDir}
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	configureGitMock(goGit, cli)
	repoLocator := &stubRepoLocator{
		dirs: []core.GitDir{{Name: "demo", Path: projectDir}},
	}
	service := NewProjectService(goGit, cli, repoLocator, config)

	result, err := service.ListProjects(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "demo", result[0].Name)
}

func TestProjectService_GetProjectInfo(t *testing.T) {
	config := core.DefaultConfig()
	goGit := mocks.NewMockGoGitClient()
	cli := mocks.NewMockCLIClient()
	configureGitMock(goGit, cli)
	repoLocator := &stubRepoLocator{}
	service := NewProjectService(goGit, cli, repoLocator, config)

	tests := []struct {
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

func TestProjectService_SearchProjectByName(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*testing.T) (*projectService, string)
		projectName    string
		expectError    bool
		errorMessage   string
		validateResult func(*testing.T, *core.ProjectInfo)
	}{
		{
			name: "directory not found error",
			setupFunc: func(t *testing.T) (*projectService, string) {
				t.Helper()
				tempDir := t.TempDir()
				projectsDir := filepath.Join(tempDir, "nonexistent")
				config := &core.Config{
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

				config := &core.Config{
					ProjectsDirectory: projectsDir,
				}
				goGit := mocks.NewMockGoGitClient()
				cli := mocks.NewMockCLIClient()
				goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)
				goGit.On("GetRepositoryInfo", mock.Anything, mock.AnythingOfType("string")).Return(&core.GitRepository{
					Path:          "/path/to/project",
					IsBare:        false,
					DefaultBranch: "main",
					Remotes:       []core.RemoteInfo{},
					Branches:      []core.BranchInfo{},
					Worktrees:     []core.WorktreeInfo{},
					Status:        core.RepositoryStatus{},
				}, nil)
				cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{}, nil)

				service := &projectService{
					goGit:  goGit,
					cli:    cli,
					config: config,
				}
				return service, projectsDir
			},
			projectName: "TestProject",
			expectError: false,
			validateResult: func(t *testing.T, result *core.ProjectInfo) {
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

				config := &core.Config{
					ProjectsDirectory: projectsDir,
				}
				goGit := mocks.NewMockGoGitClient()
				cli := mocks.NewMockCLIClient()
				goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)
				goGit.On("GetRepositoryInfo", mock.Anything, mock.AnythingOfType("string")).Return(&core.GitRepository{
					Path:          "/path/to/project",
					IsBare:        false,
					DefaultBranch: "main",
					Remotes:       []core.RemoteInfo{},
					Branches:      []core.BranchInfo{},
					Worktrees:     []core.WorktreeInfo{},
					Status:        core.RepositoryStatus{},
				}, nil)
				cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{}, nil)

				service := &projectService{
					goGit:  goGit,
					cli:    cli,
					config: config,
				}
				return service, projectsDir
			},
			projectName: "myproject",
			expectError: false,
			validateResult: func(t *testing.T, result *core.ProjectInfo) {
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

				config := &core.Config{
					ProjectsDirectory: projectsDir,
				}
				goGit := mocks.NewMockGoGitClient()
				cli := mocks.NewMockCLIClient()
				goGit.On("ValidateRepository", mock.AnythingOfType("string")).Return(nil)
				goGit.On("GetRepositoryInfo", mock.Anything, mock.AnythingOfType("string")).Return(&core.GitRepository{
					Path:          "/path/to/project",
					IsBare:        false,
					DefaultBranch: "main",
					Remotes:       []core.RemoteInfo{},
					Branches:      []core.BranchInfo{},
					Worktrees:     []core.WorktreeInfo{},
					Status:        core.RepositoryStatus{},
				}, nil)
				cli.On("ListWorktrees", mock.Anything, mock.AnythingOfType("string")).Return([]core.WorktreeInfo{}, nil)

				service := &projectService{
					goGit:  goGit,
					cli:    cli,
					config: config,
				}
				return service, projectsDir
			},
			projectName: "aproject",
			expectError: false,
			validateResult: func(t *testing.T, result *core.ProjectInfo) {
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

				config := &core.Config{
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

				config := &core.Config{
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

	for _, tc := range tests {
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
				assert.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}
		})
	}
}

func TestProjectService_FindMainRepoFromWorktree(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*testing.T) (worktreePath string, expectedPath string)
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			worktreePath, expectedPath := tc.setupFunc(t)

			service := &projectService{}
			result := service.findMainRepoFromWorktree(worktreePath)

			assert.Equal(t, expectedPath, result)
		})
	}
}

var _ application.RepoLocator = (*stubRepoLocator)(nil)
