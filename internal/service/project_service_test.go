package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

type ProjectServiceTestSuite struct {
	suite.Suite
	service    application.ProjectService
	gitService *mocks.MockGitService
	config     *domain.Config
}

func (s *ProjectServiceTestSuite) SetupTest() {
	s.config = domain.DefaultConfig()
	s.gitService = mocks.NewMockGitService()
	s.configureGitMock()
	s.service = NewProjectService(s.gitService, mocks.NewMockContextService(), s.config)
}

func (s *ProjectServiceTestSuite) configureGitMock() {
	s.gitService.MockGoGitClient.ValidateRepositoryFunc = func(path string) error {
		return nil
	}

	s.gitService.MockGoGitClient.GetRepositoryInfoFunc = func(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
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

	s.gitService.MockCLIClient.ListWorktreesFunc = func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
		return []domain.WorktreeInfo{}, nil
	}
}

func TestProjectService(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

func (s *ProjectServiceTestSuite) TestDiscoverProject() {
	tests := []struct {
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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.DiscoverProject(context.Background(), tc.projectName, tc.context)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
				expectedName := tc.projectName
				if expectedName == "" && tc.context != nil {
					expectedName = tc.context.ProjectName
				}
				s.Equal(expectedName, result.Name)
			}
		})
	}
}

func (s *ProjectServiceTestSuite) TestValidateProject() {
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
		s.Run(tc.name, func() {
			err := s.service.ValidateProject(context.Background(), tc.projectPath)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ProjectServiceTestSuite) TestListProjects() {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "valid projects listing",
			expectError: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result, err := s.service.ListProjects(context.Background())

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

func (s *ProjectServiceTestSuite) TestGetProjectInfo() {
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
		s.Run(tc.name, func() {
			result, err := s.service.GetProjectInfo(context.Background(), tc.projectPath)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(result)
			} else {
				s.Require().NoError(err)
				s.NotNil(result)
				s.Equal(tc.projectPath, result.Path)
			}
		})
	}
}

func (s *ProjectServiceTestSuite) TestSearchProjectByName() {
	tests := []struct {
		name           string
		setupFunc      func() (*projectService, string)
		projectName    string
		expectError    bool
		errorMessage   string
		validateResult func(*domain.ProjectInfo)
	}{
		{
			name: "directory not found error",
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
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
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				s.Require().NoError(os.Mkdir(projectsDir, 0755))

				projectPath := filepath.Join(projectsDir, "testproject")
				s.Require().NoError(os.Mkdir(projectPath, 0755))

				gitDir := filepath.Join(projectPath, ".git")
				s.Require().NoError(os.Mkdir(gitDir, 0755))

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
			validateResult: func(result *domain.ProjectInfo) {
				s.Equal("testproject", result.Name)
			},
		},
		{
			name: "exact match works",
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				s.Require().NoError(os.Mkdir(projectsDir, 0755))

				projectPath := filepath.Join(projectsDir, "myproject")
				s.Require().NoError(os.Mkdir(projectPath, 0755))

				gitDir := filepath.Join(projectPath, ".git")
				s.Require().NoError(os.Mkdir(gitDir, 0755))

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
			validateResult: func(result *domain.ProjectInfo) {
				s.Equal("myproject", result.Name)
			},
		},
		{
			name: "multiple matches returns first one",
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				s.Require().NoError(os.Mkdir(projectsDir, 0755))

				for _, name := range []string{"aproject", "bproject", "cproject"} {
					projectPath := filepath.Join(projectsDir, name)
					s.Require().NoError(os.Mkdir(projectPath, 0755))

					gitDir := filepath.Join(projectPath, ".git")
					s.Require().NoError(os.Mkdir(gitDir, 0755))
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
			validateResult: func(result *domain.ProjectInfo) {
				s.Equal("aproject", result.Name)
			},
		},
		{
			name: "empty directory returns not found",
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
				projectsDir := filepath.Join(tempDir, "projects")
				s.Require().NoError(os.Mkdir(projectsDir, 0755))

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
			setupFunc: func() (*projectService, string) {
				tempDir := s.T().TempDir()
				projectsDir := filepath.Join(tempDir, "projects")

				s.Require().NoError(os.WriteFile(projectsDir, []byte("not a directory"), 0644))

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

	for _, tc := range tests {
		s.Run(tc.name, func() {
			service, _ := tc.setupFunc()

			result, err := service.searchProjectByName(context.Background(), tc.projectName)

			if tc.expectError {
				s.Require().Error(err)
				s.Nil(result)
				if tc.errorMessage != "" {
					s.Contains(err.Error(), tc.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				if tc.validateResult != nil {
					tc.validateResult(result)
				}
			}
		})
	}
}

func (s *ProjectServiceTestSuite) TestFindMainRepoFromWorktree() {
	tests := []struct {
		name      string
		setupFunc func() (worktreePath string, expectedPath string)
	}{
		{
			name: "main repo found at parent directory",
			setupFunc: func() (string, string) {
				tempDir := s.T().TempDir()

				mainRepoPath := tempDir
				gitDir := filepath.Join(mainRepoPath, ".git")
				s.Require().NoError(os.Mkdir(gitDir, 0755))

				headsDir := filepath.Join(gitDir, "heads")
				s.Require().NoError(os.Mkdir(headsDir, 0755))

				worktreePath := filepath.Join(tempDir, "worktree")
				s.Require().NoError(os.Mkdir(worktreePath, 0755))

				gitFileContent := fmt.Sprintf("gitdir: %s\n", filepath.ToSlash(gitDir))
				gitFilePath := filepath.Join(worktreePath, ".git")
				s.Require().NoError(os.WriteFile(gitFilePath, []byte(gitFileContent), 0644))

				return worktreePath, tempDir
			},
		},
		{
			name: "root directory stops at root",
			setupFunc: func() (string, string) {
				tempDir := s.T().TempDir()
				return tempDir, tempDir
			},
		},
		{
			name: "no main repo found returns input path",
			setupFunc: func() (string, string) {
				tempDir := s.T().TempDir()

				subdirPath := filepath.Join(tempDir, "subdir1", "subdir2")
				s.Require().NoError(os.MkdirAll(subdirPath, 0755))

				return subdirPath, subdirPath
			},
		},
		{
			name: "main repo found at intermediate directory",
			setupFunc: func() (string, string) {
				tempDir := s.T().TempDir()

				mainRepoPath := filepath.Join(tempDir, "main")
				s.Require().NoError(os.Mkdir(mainRepoPath, 0755))

				gitDir := filepath.Join(mainRepoPath, ".git")
				s.Require().NoError(os.Mkdir(gitDir, 0755))

				subdirPath := filepath.Join(mainRepoPath, "subdir1", "subdir2")
				s.Require().NoError(os.MkdirAll(subdirPath, 0755))

				return subdirPath, mainRepoPath
			},
		},
		{
			name: "worktree directory returns worktree path if no main repo",
			setupFunc: func() (string, string) {
				tempDir := s.T().TempDir()

				worktreePath := filepath.Join(tempDir, "worktree")
				s.Require().NoError(os.Mkdir(worktreePath, 0755))

				return worktreePath, worktreePath
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			worktreePath, expectedPath := tc.setupFunc()

			service := &projectService{}
			result := service.findMainRepoFromWorktree(worktreePath)

			s.Equal(expectedPath, result)
		})
	}
}
