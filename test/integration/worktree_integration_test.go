//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/infrastructure/mise"
	"github.com/amaury/twiggit/internal/infrastructure/validation"
	"github.com/amaury/twiggit/internal/services"
	"github.com/amaury/twiggit/test/helpers"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestRepo wraps helpers.GitRepo with integration-specific functionality
type IntegrationTestRepo struct {
	*helpers.GitRepo
	TempDir string
	cleanup func()
}

// NewTestGitRepo creates a new git repository for integration testing
func NewTestGitRepo(t *testing.T) *IntegrationTestRepo {
	tempDir, cleanup := helpers.TempDir(t, "twiggit-integration-*")

	// Create git repo directly in the temp directory
	repo := helpers.NewGitRepo(t, "twiggit-integration-*")

	integrationRepo := &IntegrationTestRepo{
		GitRepo: repo,
		TempDir: tempDir,
		cleanup: func() {
			repo.Cleanup()
			cleanup()
		},
	}

	return integrationRepo
}

// Cleanup removes the test repository and temp directory
func (r *IntegrationTestRepo) Cleanup() {
	if r.cleanup != nil {
		r.cleanup()
	}
}

// RepoDir returns the repository directory (alias for Path for backward compatibility)
func (r *IntegrationTestRepo) RepoDir() string {
	return r.Path
}

type WorktreeIntegrationTestSuite struct {
	suite.Suite
	testRepo                 *IntegrationTestRepo
	gitClient                domain.GitClient
	discoveryService         *services.DiscoveryService
	config                   *config.Config
	worktreeCreator          *services.WorktreeCreator
	worktreeRemover          *services.WorktreeRemover
	currentDirectoryDetector *services.CurrentDirectoryDetector
}

func (s *WorktreeIntegrationTestSuite) SetupSuite() {
	// Skip if not in integration test mode
	if testing.Short() {
		s.T().Skip("Skipping integration test")
	}

	// Create test git repository
	s.testRepo = NewTestGitRepo(s.T())

	// Create some branches for testing
	s.testRepo.CreateBranch(s.T(), "feature-1")
	s.testRepo.CreateBranch(s.T(), "feature-2")
	s.testRepo.AddMiseConfig(s.T())

	// Initialize services
	client := git.NewClient()
	s.gitClient = client
	s.config = &config.Config{
		Workspace: s.testRepo.TempDir,
	}

	// Create path validator and filesystem
	pathValidator := validation.NewPathValidator()
	fileSystem := os.DirFS("/") // Use real filesystem for integration tests

	s.discoveryService = services.NewDiscoveryService(s.gitClient, s.config, fileSystem, pathValidator)
	infraService := infrastructure.NewInfrastructureService(s.gitClient, fileSystem, pathValidator)
	validationService := services.NewValidationService(infraService)
	miseService := mise.NewMiseIntegration()
	s.worktreeCreator = services.NewWorktreeCreator(s.gitClient, validationService, miseService)
	s.worktreeRemover = services.NewWorktreeRemover(s.gitClient)
	s.currentDirectoryDetector = services.NewCurrentDirectoryDetector(s.gitClient)
}

func (s *WorktreeIntegrationTestSuite) TearDownSuite() {
	if s.testRepo != nil {
		s.testRepo.Cleanup()
	}
}

func TestWorktreeIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WorktreeIntegrationTestSuite))
}

func (s *WorktreeIntegrationTestSuite) TestFullWorktreeLifecycle() {
	// Generate unique suffix for this test run to avoid conflicts
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	s.Run("should create worktree from existing branch", func() {
		worktreePath := filepath.Join(filepath.Dir(s.testRepo.RepoDir()), "feature-1-worktree-"+uniqueSuffix)

		// Verify the branch exists before creating worktree
		exists := s.gitClient.BranchExists(context.Background(), s.testRepo.RepoDir(), "feature-1")
		s.Assert().True(exists, "feature-1 branch should exist before creating worktree")

		err := s.worktreeCreator.Create(context.Background(), s.testRepo.RepoDir(), "feature-1", worktreePath)
		s.Assert().NoError(err)

		// Verify worktree was created
		_, err = os.Stat(worktreePath)
		s.Assert().NoError(err, "Worktree directory should exist")

		// Verify it's a valid git worktree
		isRepo, err := s.gitClient.IsGitRepository(context.Background(), worktreePath)
		s.Assert().NoError(err)
		s.Assert().True(isRepo, "Worktree should be a valid git repository")

		// Verify branch is checked out correctly
		status, err := s.gitClient.GetWorktreeStatus(context.Background(), worktreePath)
		s.Assert().NoError(err)
		s.Assert().Equal("feature-1", status.Branch)

		// Verify mise config was copied
		miseFile := filepath.Join(worktreePath, ".mise.local.toml")
		_, err = os.Stat(miseFile)
		s.Assert().NoError(err, "Mise configuration should be copied to worktree")
	})

	s.Run("should create worktree for new branch", func() {
		worktreePath := filepath.Join(filepath.Dir(s.testRepo.RepoDir()), "new-feature-worktree-"+uniqueSuffix)

		err := s.worktreeCreator.Create(context.Background(), s.testRepo.RepoDir(), "new-feature", worktreePath)
		s.Assert().NoError(err)

		// Verify worktree was created
		_, err = os.Stat(worktreePath)
		s.Assert().NoError(err, "Worktree directory should exist")

		// Verify branch exists now
		exists := s.gitClient.BranchExists(context.Background(), s.testRepo.RepoDir(), "new-feature")
		s.Assert().True(exists, "New branch should have been created")
	})

	s.Run("should list all worktrees", func() {
		worktrees, err := s.gitClient.ListWorktrees(context.Background(), s.testRepo.RepoDir())
		s.Assert().NoError(err)

		// Debug: print what we found
		for i, wt := range worktrees {
			s.T().Logf("Worktree %d: Path=%s, Branch=%s", i, wt.Path, wt.Branch)
		}

		// Should have main repo + 2 worktrees
		s.Assert().GreaterOrEqual(len(worktrees), 3, "Should have main repo and 2 worktrees")

		// Find our specific worktrees
		var foundFeature1, foundNewFeature bool
		for _, wt := range worktrees {
			if wt.Branch == "feature-1" {
				foundFeature1 = true
			}
			if wt.Branch == "new-feature" {
				foundNewFeature = true
			}
		}
		s.Assert().True(foundFeature1, "Should find feature-1 worktree")
		s.Assert().True(foundNewFeature, "Should find new-feature worktree")
	})

	s.Run("should remove worktree safely", func() {
		worktreePath := filepath.Join(filepath.Dir(s.testRepo.RepoDir()), "feature-1-worktree-"+uniqueSuffix)

		// Verify it exists first
		_, err := os.Stat(worktreePath)
		s.Assert().NoError(err)

		// Remove the worktree (use force since mise config was copied)
		err = s.worktreeRemover.Remove(context.Background(), worktreePath, true)
		s.Assert().NoError(err)

		// Verify it was removed
		_, err = os.Stat(worktreePath)
		s.Assert().True(os.IsNotExist(err), "Worktree directory should be removed")

		// Verify it's no longer in the worktree list
		worktrees, err := s.gitClient.ListWorktrees(context.Background(), s.testRepo.RepoDir())
		s.Assert().NoError(err)

		for _, wt := range worktrees {
			s.Assert().NotContains(wt.Path, "feature-1-worktree", "Removed worktree should not be in list")
		}
	})
}

func (s *WorktreeIntegrationTestSuite) TestDiscoveryService() {
	// Create workspace with multiple projects
	workspaceDir, err := os.MkdirTemp("", "twiggit-workspace-*")
	s.Require().NoError(err)
	defer os.RemoveAll(workspaceDir)

	// Create project 1
	project1 := NewTestGitRepo(s.T())
	project1Path := filepath.Join(workspaceDir, "project1")
	err = os.Rename(project1.Path, project1Path)
	s.Require().NoError(err)
	project1.Path = project1Path

	// Create project 2
	project2 := NewTestGitRepo(s.T())
	project2Path := filepath.Join(workspaceDir, "project2")
	err = os.Rename(project2.Path, project2Path)
	s.Require().NoError(err)
	project2.Path = project2Path

	// Create some worktrees in the proper nested structure
	gitClient := git.NewClient()

	// Create worktrees for project1
	worktree1Path := filepath.Join(workspaceDir, "project1", "feature1")
	err = gitClient.CreateWorktree(context.Background(), project1Path, "feature-branch-1", worktree1Path)
	s.Require().NoError(err)

	worktree2Path := filepath.Join(workspaceDir, "project1", "feature2")
	err = gitClient.CreateWorktree(context.Background(), project1Path, "feature-branch-2", worktree2Path)
	s.Require().NoError(err)

	// Create worktrees for project2
	worktree3Path := filepath.Join(workspaceDir, "project2", "develop")
	err = gitClient.CreateWorktree(context.Background(), project2Path, "develop", worktree3Path)
	s.Require().NoError(err)

	// Test discovery
	config := &config.Config{Workspace: workspaceDir}
	fileSystem := os.DirFS(workspaceDir)
	pathValidator := validation.NewPathValidator()
	discoveryService := services.NewDiscoveryService(gitClient, config, fileSystem, pathValidator)

	s.Run("should discover all projects", func() {
		// FileSystem is rooted at workspaceDir, so use "." to scan current directory
		projects, err := discoveryService.DiscoverProjects(context.Background(), ".")
		s.Assert().NoError(err)

		// Debug: print what we found
		for i, p := range projects {
			s.T().Logf("Project %d: Name=%s, GitRepo=%s, Worktrees=%d", i, p.Name, p.GitRepo, len(p.Worktrees))
		}

		s.Assert().Len(projects, 2, "Should discover 2 projects")

		projectNames := make([]string, len(projects))
		for i, p := range projects {
			projectNames[i] = p.Name
		}
		s.Assert().Contains(projectNames, "project1")
		s.Assert().Contains(projectNames, "project2")
	})

	s.Run("should discover all worktrees", func() {
		// FileSystem is rooted at workspaceDir, so use "." to scan current directory
		worktrees, err := discoveryService.DiscoverWorktrees(context.Background(), ".")
		s.Assert().NoError(err)
		s.Assert().GreaterOrEqual(len(worktrees), 3, "Should discover at least 3 worktrees (feature1, feature2, develop)")

		// Check that our specific worktrees are found
		worktreePaths := make([]string, len(worktrees))
		for i, wt := range worktrees {
			worktreePaths[i] = wt.Path
		}
		s.Assert().Contains(worktreePaths, worktree1Path)
		s.Assert().Contains(worktreePaths, worktree2Path)
		s.Assert().Contains(worktreePaths, worktree3Path)
	})
}

func (s *WorktreeIntegrationTestSuite) TestErrorHandling() {
	testRepo := NewTestGitRepo(s.T())
	defer testRepo.Cleanup()

	gitClient := git.NewClient()
	config := &config.Config{
		Workspace: testRepo.TempDir,
	}
	fileSystem := os.DirFS("/") // Use real filesystem for integration tests
	pathValidator := validation.NewPathValidator()
	_ = services.NewDiscoveryService(gitClient, config, fileSystem, pathValidator)
	infraService := infrastructure.NewInfrastructureService(gitClient, fileSystem, pathValidator)
	validationService := services.NewValidationService(infraService)
	miseService := mise.NewMiseIntegration()
	worktreeCreator := services.NewWorktreeCreator(gitClient, validationService, miseService)
	worktreeRemover := services.NewWorktreeRemover(gitClient)

	s.Run("should handle non-existent repository", func() {
		err := worktreeCreator.Create(context.Background(), "/non/existent/repo", "feature", "/tmp/test-worktree")
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "not a git repository")
	})

	s.Run("should handle invalid target path", func() {
		err := worktreeCreator.Create(context.Background(), testRepo.RepoDir(), "feature", "relative/path")
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), "path must be absolute")
	})

	s.Run("should handle existing target directory", func() {
		existingPath := filepath.Join(testRepo.TempDir, "existing")
		err := os.MkdirAll(existingPath, 0755)
		s.Require().NoError(err)

		err = worktreeCreator.Create(context.Background(), testRepo.RepoDir(), "feature", existingPath)
		s.Assert().Error(err)
	})

	s.Run("should handle removal of non-existent worktree", func() {
		err := worktreeRemover.Remove(context.Background(), "/non/existent/worktree", false)
		s.Assert().Error(err)
	})
}

func (s *WorktreeIntegrationTestSuite) TestPerformance() {
	// Create workspace with multiple projects and worktrees
	workspaceDir, err := os.MkdirTemp("", "twiggit-perf-*")
	s.Require().NoError(err)
	defer os.RemoveAll(workspaceDir)

	projectCount := 5
	worktreesPerProject := 3

	gitClient := git.NewClient()

	// Create multiple projects
	for i := 0; i < projectCount; i++ {
		projectRepo := NewTestGitRepo(s.T())
		projectPath := filepath.Join(workspaceDir, fmt.Sprintf("project%d", i))
		err = os.Rename(projectRepo.Path, projectPath)
		s.Require().NoError(err)

		// Create branches and worktrees for each project
		for j := 0; j < worktreesPerProject; j++ {
			branchName := fmt.Sprintf("feature-%d", j)
			worktreePath := filepath.Join(workspaceDir, fmt.Sprintf("project%d", i), branchName)

			err = gitClient.CreateWorktree(context.Background(), projectPath, branchName, worktreePath)
			s.Require().NoError(err)
		}
	}

	config := &config.Config{Workspace: workspaceDir}
	fileSystem := os.DirFS(workspaceDir)
	pathValidator := validation.NewPathValidator()
	discoveryService := services.NewDiscoveryService(gitClient, config, fileSystem, pathValidator)
	discoveryService.SetConcurrency(4) // Test with concurrent processing

	s.Run("should discover projects efficiently", func() {
		// FileSystem is rooted at workspaceDir, so use "." to scan current directory
		projects, err := discoveryService.DiscoverProjects(context.Background(), ".")
		s.Assert().NoError(err)
		s.Assert().Len(projects, projectCount, "Should discover all projects")

		// Note: DiscoverProjects doesn't populate worktrees - that's DiscoverWorktrees' job
		// We just verify that projects are discovered correctly
		projectNames := make([]string, len(projects))
		for i, project := range projects {
			projectNames[i] = project.Name
		}
		for i := 0; i < projectCount; i++ {
			s.Assert().Contains(projectNames, fmt.Sprintf("project%d", i))
		}
	})

	s.Run("should discover all worktrees efficiently", func() {
		// FileSystem is rooted at workspaceDir, so use "." to scan current directory
		worktrees, err := discoveryService.DiscoverWorktrees(context.Background(), ".")
		s.Assert().NoError(err)

		expectedCount := projectCount * (worktreesPerProject + 1) // +1 for main repo
		s.Assert().GreaterOrEqual(len(worktrees), expectedCount,
			"Should discover at least %d worktrees", expectedCount)
	})
}
