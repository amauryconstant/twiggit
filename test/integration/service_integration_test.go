//go:build integration

package integration

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/internal/service"
	"twiggit/internal/services"
)

type ServiceIntegrationTestSuite struct {
	suite.Suite
	config            *domain.Config
	gitService        infrastructure.GitClient
	projectService    application.ProjectService
	worktreeService   application.WorktreeService
	navigationService application.NavigationService
	contextService    domain.ContextService
	tempDir           string
	projectsDir       string
	worktreesDir      string
	mainRepoPath      string
	mainBranchName    string
	worktree1Path     string
	worktree2Path     string
	branch1           string
	branch2           string
}

func (s *ServiceIntegrationTestSuite) SetupSuite() {
	if _, err := exec.LookPath("git"); err != nil {
		s.T().Skip("git not available for integration tests")
	}

	s.tempDir = s.T().TempDir()
	s.projectsDir = filepath.Join(s.tempDir, "Projects")
	s.worktreesDir = filepath.Join(s.tempDir, "Worktrees")
	s.branch1 = "feature-branch-1"
	s.branch2 = "feature-branch-2"

	s.config = &domain.Config{
		ProjectsDirectory:   s.projectsDir,
		WorktreesDirectory:  s.worktreesDir,
		DefaultSourceBranch: "main",
		ContextDetection: domain.ContextDetectionConfig{
			CacheTTL:            "5m",
			GitOperationTimeout: "30s",
			EnableGitValidation: true,
		},
		Git: domain.GitConfig{
			CLITimeout:   30,
			CacheEnabled: true,
		},
		Navigation: domain.NavigationConfig{
			EnableSuggestions: true,
			MaxSuggestions:    10,
		},
	}

	executor := infrastructure.NewDefaultCommandExecutor(30 * time.Second)
	goGitClient := infrastructure.NewGoGitClient(true)
	cliClient := infrastructure.NewCLIClient(executor, 30)
	s.gitService = infrastructure.NewCompositeGitClient(goGitClient, cliClient)

	detector := infrastructure.NewContextDetector(s.config)
	resolver := infrastructure.NewContextResolver(s.config, s.gitService)
	s.contextService = service.NewContextService(detector, resolver, s.config)

	s.projectService = services.NewProjectService(s.gitService, s.contextService, s.config)
	s.worktreeService = services.NewWorktreeService(s.gitService, s.projectService, s.config)
	s.navigationService = services.NewNavigationService(s.projectService, s.contextService, s.config)

	s.setupGitRepository()
}

func (s *ServiceIntegrationTestSuite) setupGitRepository() {
	projectName := "test-project"
	s.mainRepoPath = filepath.Join(s.projectsDir, projectName)
	require.NoError(s.T(), os.MkdirAll(s.mainRepoPath, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	testFile := filepath.Join(s.mainRepoPath, "README.md")
	require.NoError(s.T(), os.WriteFile(testFile, []byte("# Test Repository\n"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	// Get the actual branch name
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = s.mainRepoPath
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err)
	s.mainBranchName = strings.TrimSpace(string(output))

	s.worktree1Path = filepath.Join(s.worktreesDir, projectName, s.branch1)
	require.NoError(s.T(), os.MkdirAll(filepath.Dir(s.worktree1Path), 0755))

	cmd = exec.Command("git", "worktree", "add", s.worktree1Path, "-b", s.branch1)
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())

	s.worktree2Path = filepath.Join(s.worktreesDir, projectName, s.branch2)
	require.NoError(s.T(), os.MkdirAll(filepath.Dir(s.worktree2Path), 0755))

	cmd = exec.Command("git", "worktree", "add", s.worktree2Path, "-b", s.branch2)
	cmd.Dir = s.mainRepoPath
	require.NoError(s.T(), cmd.Run())
}

func (s *ServiceIntegrationTestSuite) TearDownSuite() {
	if s.worktree2Path != "" {
		os.RemoveAll(s.worktree2Path)
	}
	if s.worktree1Path != "" {
		os.RemoveAll(s.worktree1Path)
	}
	if s.mainRepoPath != "" {
		os.RemoveAll(s.mainRepoPath)
	}
}

func TestServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceIntegrationTestSuite))
}

func (s *ServiceIntegrationTestSuite) TestWorktreeDiscovery() {
	ctx := context.Background()

	projectName := "test-project"
	project, err := s.projectService.DiscoverProject(ctx, projectName, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), projectName, project.Name)
	assert.NotNil(s.T(), project.Worktrees)
	assert.GreaterOrEqual(s.T(), len(project.Worktrees), 1)

	foundWorktree := false
	for _, wt := range project.Worktrees {
		if wt.Path == s.worktree1Path {
			foundWorktree = true
			assert.Equal(s.T(), s.branch1, wt.Branch)
			break
		}
	}
	assert.True(s.T(), foundWorktree, "Worktree1 should be discovered by ProjectService")
}

func (s *ServiceIntegrationTestSuite) TestWorktreeListProjects() {
	ctx := context.Background()

	req := &domain.ListWorktreesRequest{
		ProjectName: "test-project",
		Context:     nil,
		IncludeMain: false,
	}

	worktrees, err := s.worktreeService.ListWorktrees(ctx, req)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(worktrees), 2)

	projectCount := 0
	for _, wt := range worktrees {
		if wt.Branch == s.branch1 || wt.Branch == s.branch2 {
			projectCount++
		}
	}
	assert.GreaterOrEqual(s.T(), projectCount, 2)
}

func (s *ServiceIntegrationTestSuite) TestCrossServiceErrorHandling() {
	ctx := context.Background()

	req := &domain.CreateWorktreeRequest{
		ProjectName:  "",
		BranchName:   "invalid-branch",
		SourceBranch: "",
		Context:      nil,
		Force:        false,
	}

	worktreeInfo, err := s.worktreeService.CreateWorktree(ctx, req)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), worktreeInfo)

	assert.Contains(s.T(), err.Error(), "validation failed")
}

func (s *ServiceIntegrationTestSuite) TestNavigationToWorktree() {
	ctx := context.Background()

	req := &domain.ResolvePathRequest{
		Target: s.branch1,
		Context: &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "test-project",
			Path:        s.mainRepoPath,
		},
		Search: false,
	}

	result, err := s.navigationService.ResolvePath(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.worktree1Path, result.ResolvedPath)
	assert.Equal(s.T(), domain.PathTypeWorktree, result.Type)
	assert.Equal(s.T(), "test-project", result.ProjectName)
	assert.Equal(s.T(), s.branch1, result.BranchName)

	projectInfo, err := s.projectService.GetProjectInfo(ctx, result.ResolvedPath)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-project", projectInfo.Name)
}

func (s *ServiceIntegrationTestSuite) TestRelativePathResolution() {
	ctx := context.Background()

	worktreeContext := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  s.branch1,
		Path:        s.worktree1Path,
	}

	req := &domain.ResolvePathRequest{
		Target:  "main",
		Context: worktreeContext,
		Search:  false,
	}

	result, err := s.navigationService.ResolvePath(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.PathTypeProject, result.Type)
	assert.Equal(s.T(), "test-project", result.ProjectName)
	assert.Equal(s.T(), s.mainRepoPath, result.ResolvedPath)

	projectInfo, err := s.projectService.GetProjectInfo(ctx, result.ResolvedPath)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-project", projectInfo.Name)
}

func (s *ServiceIntegrationTestSuite) TestFullWorkflow() {
	ctx := context.Background()
	projectName := "workflow-test-project"

	mainRepo := filepath.Join(s.projectsDir, projectName)
	require.NoError(s.T(), os.MkdirAll(mainRepo, 0755))

	cmd := exec.Command("git", "init")
	cmd.Dir = mainRepo
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "workflow@example.com")
	cmd.Dir = mainRepo
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Workflow User")
	cmd.Dir = mainRepo
	require.NoError(s.T(), cmd.Run())

	testFile := filepath.Join(mainRepo, "test.txt")
	require.NoError(s.T(), os.WriteFile(testFile, []byte("test"), 0644))

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = mainRepo
	require.NoError(s.T(), cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = mainRepo
	require.NoError(s.T(), cmd.Run())

	// Get the actual branch name
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = mainRepo
	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err)
	actualBranchName := strings.TrimSpace(string(output))

	discoveredProject, err := s.projectService.DiscoverProject(ctx, projectName, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), projectName, discoveredProject.Name)

	newBranch := "workflow-branch"
	createReq := &domain.CreateWorktreeRequest{
		ProjectName:  projectName,
		BranchName:   newBranch,
		SourceBranch: actualBranchName,
		Context: &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: projectName,
			Path:        mainRepo,
		},
		Force: false,
	}

	createdWorktree, err := s.worktreeService.CreateWorktree(ctx, createReq)
	if err != nil {
		s.T().Logf("Error creating worktree: %v", err)
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			s.T().Logf("Unwrapped error: %v", unwrapped)
			if deeperUnwrapped := errors.Unwrap(unwrapped); deeperUnwrapped != nil {
				s.T().Logf("Deeper unwrapped error: %v", deeperUnwrapped)
			}
		}
	}
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), createdWorktree)
	assert.Contains(s.T(), createdWorktree.Path, newBranch)
	assert.Equal(s.T(), newBranch, createdWorktree.Branch)

	navReq := &domain.ResolvePathRequest{
		Target: newBranch,
		Context: &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: projectName,
			Path:        mainRepo,
		},
		Search: false,
	}

	navResult, err := s.navigationService.ResolvePath(ctx, navReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), createdWorktree.Path, navResult.ResolvedPath)

	deleteReq := &domain.DeleteWorktreeRequest{
		WorktreePath: createdWorktree.Path,
		Force:        true,
		KeepBranch:   false,
		Context: &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: projectName,
			Path:        mainRepo,
		},
	}

	err = s.worktreeService.DeleteWorktree(ctx, deleteReq)
	require.NoError(s.T(), err)

	_, err = os.Stat(createdWorktree.Path)
	assert.True(s.T(), os.IsNotExist(err), "Worktree directory should be deleted")

	os.RemoveAll(mainRepo)
}

func (s *ServiceIntegrationTestSuite) TestBranchWorkflow() {
	ctx := context.Background()

	req := &domain.ListWorktreesRequest{
		ProjectName: "test-project",
		Context:     nil,
		IncludeMain: false,
	}

	worktrees, err := s.worktreeService.ListWorktrees(ctx, req)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(worktrees), 2)

	activeWorktree := worktrees[0]
	navReq := &domain.ResolvePathRequest{
		Target: activeWorktree.Branch,
		Context: &domain.Context{
			Type:        domain.ContextProject,
			ProjectName: "test-project",
			Path:        s.mainRepoPath,
		},
		Search: false,
	}

	navResult, err := s.navigationService.ResolvePath(ctx, navReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), activeWorktree.Path, navResult.ResolvedPath)

	projectInfo, err := s.projectService.GetProjectInfo(ctx, navResult.ResolvedPath)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-project", projectInfo.Name)

	found := false
	for _, wt := range projectInfo.Worktrees {
		if wt.Path == activeWorktree.Path {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Active worktree should be found in project info")
}

func (s *ServiceIntegrationTestSuite) TestContextAwareWorkflow() {
	ctx := context.Background()

	subDir := filepath.Join(s.worktree1Path, "subdir", "nested")
	require.NoError(s.T(), os.MkdirAll(subDir, 0755))

	testFile := filepath.Join(subDir, "test.txt")
	require.NoError(s.T(), os.WriteFile(testFile, []byte("test content"), 0644))

	detectedContext, err := s.contextService.DetectContextFromPath(subDir)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.ContextWorktree, detectedContext.Type)
	assert.Equal(s.T(), "test-project", detectedContext.ProjectName)
	assert.Equal(s.T(), s.branch1, detectedContext.BranchName)

	projectFromSubdir, err := s.projectService.DiscoverProject(ctx, "", detectedContext)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-project", projectFromSubdir.Name)

	suggestions, err := s.navigationService.GetNavigationSuggestions(ctx, detectedContext, "")
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), suggestions)

	os.RemoveAll(subDir)
}

func (s *ServiceIntegrationTestSuite) TestContextPropagation() {
	ctx := context.Background()

	initialContext := &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "test-project",
		Path:        s.mainRepoPath,
	}

	createReq := &domain.CreateWorktreeRequest{
		ProjectName:  "test-project",
		BranchName:   "context-test-branch",
		SourceBranch: s.mainBranchName,
		Context:      initialContext,
		Force:        false,
	}

	createdWorktree, err := s.worktreeService.CreateWorktree(ctx, createReq)
	require.NoError(s.T(), err)

	worktreeContext := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  "context-test-branch",
		Path:        createdWorktree.Path,
	}

	projectFromWorktree, err := s.projectService.DiscoverProject(ctx, "", worktreeContext)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-project", projectFromWorktree.Name)

	navReq := &domain.ResolvePathRequest{
		Target:  "main",
		Context: worktreeContext,
		Search:  false,
	}

	navResult, err := s.navigationService.ResolvePath(ctx, navReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.mainRepoPath, navResult.ResolvedPath)

	deleteReq := &domain.DeleteWorktreeRequest{
		WorktreePath: createdWorktree.Path,
		Force:        true,
		KeepBranch:   false,
		Context:      worktreeContext,
	}

	err = s.worktreeService.DeleteWorktree(ctx, deleteReq)
	require.NoError(s.T(), err)
}

func (s *ServiceIntegrationTestSuite) TestContextIsolation() {
	ctx := context.Background()

	context1 := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  s.branch1,
		Path:        s.worktree1Path,
	}

	context2 := &domain.Context{
		Type:        domain.ContextWorktree,
		ProjectName: "test-project",
		BranchName:  s.branch2,
		Path:        s.worktree2Path,
	}

	project1, err := s.projectService.DiscoverProject(ctx, "", context1)
	require.NoError(s.T(), err)

	project2, err := s.projectService.DiscoverProject(ctx, "", context2)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), project1.Name, project2.Name)
	assert.Equal(s.T(), project1.Path, project2.Path)

	navReq1 := &domain.ResolvePathRequest{
		Target:  "main",
		Context: context1,
		Search:  false,
	}

	navReq2 := &domain.ResolvePathRequest{
		Target:  "main",
		Context: context2,
		Search:  false,
	}

	result1, err := s.navigationService.ResolvePath(ctx, navReq1)
	require.NoError(s.T(), err)

	result2, err := s.navigationService.ResolvePath(ctx, navReq2)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), result1.ResolvedPath, result2.ResolvedPath)
	assert.Equal(s.T(), s.mainRepoPath, result1.ResolvedPath)
}

func (s *ServiceIntegrationTestSuite) TestCascadingErrors() {
	ctx := context.Background()

	req := &domain.ListWorktreesRequest{
		ProjectName: "",
		Context: &domain.Context{
			Type:        domain.ContextOutsideGit,
			ProjectName: "",
			Path:        "/nonexistent/path",
		},
		IncludeMain: false,
	}

	_, err := s.worktreeService.ListWorktrees(ctx, req)
	assert.Error(s.T(), err)

	validProject, err := s.projectService.DiscoverProject(ctx, "test-project", nil)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), validProject)

	validWorktrees, err := s.worktreeService.ListWorktrees(ctx, &domain.ListWorktreesRequest{
		ProjectName: "test-project",
		Context:     nil,
		IncludeMain: false,
	})
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(validWorktrees), 2)
}

func (s *ServiceIntegrationTestSuite) TestGracefulDegradation() {
	ctx := context.Background()

	projects, err := s.projectService.ListProjects(ctx)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(projects), 1)

	testProject := projects[0]
	assert.Equal(s.T(), "test-project", testProject.Name)
	assert.NotNil(s.T(), testProject.Branches)
	assert.NotNil(s.T(), testProject.Worktrees)

	if len(testProject.Branches) > 0 {
		req := &domain.ResolvePathRequest{
			Target: testProject.Branches[0].Name,
			Context: &domain.Context{
				Type:        domain.ContextProject,
				ProjectName: testProject.Name,
				Path:        testProject.Path,
			},
			Search: false,
		}

		result, err := s.navigationService.ResolvePath(ctx, req)
		if err == nil {
			assert.NotNil(s.T(), result)
		}
	}

	suggestions, err := s.navigationService.GetNavigationSuggestions(ctx, &domain.Context{
		Type:        domain.ContextProject,
		ProjectName: "test-project",
		Path:        s.mainRepoPath,
	}, "")
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), suggestions)
}
