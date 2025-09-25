//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/services"
	"github.com/amaury/twiggit/test/helpers"
	"github.com/stretchr/testify/suite"
)

// ArchitectureIntegrationTestSuite tests the correct Projects/Workspaces architecture
type ArchitectureIntegrationTestSuite struct {
	suite.Suite
	tempDir          string
	projectsDir      string
	workspacesDir    string
	gitClient        infrastructure.GitClient
	discoveryService *services.DiscoveryService
	config           *config.Config
}

func TestArchitectureIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ArchitectureIntegrationTestSuite))
}

func (s *ArchitectureIntegrationTestSuite) SetupSuite() {
	// Skip if not in integration test mode
	if testing.Short() {
		s.T().Skip("Skipping integration test")
	}

	// Create temporary directory structure
	var err error
	s.tempDir, err = os.MkdirTemp("", "twiggit-arch-test-*")
	s.Require().NoError(err)

	// Create the expected directory structure
	s.projectsDir = filepath.Join(s.tempDir, "Projects")
	s.workspacesDir = filepath.Join(s.tempDir, "Workspaces")

	err = os.MkdirAll(s.projectsDir, 0755)
	s.Require().NoError(err)

	err = os.MkdirAll(s.workspacesDir, 0755)
	s.Require().NoError(err)

	// Initialize services
	s.gitClient = git.NewClient()

	// Create config with both paths
	s.config = &config.Config{
		ProjectsPath:   s.projectsDir,
		WorkspacesPath: s.workspacesDir,
	}

	// Create filesystem
	fileSystem := os.DirFS(s.tempDir) // Use temp directory as filesystem root for integration tests

	s.discoveryService = services.NewDiscoveryService(s.gitClient, s.config, fileSystem)
}

func (s *ArchitectureIntegrationTestSuite) TearDownSuite() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *ArchitectureIntegrationTestSuite) SetupTest() {
	// Clean up any existing files in temp directories before each test
	s.cleanupDirectories()
}

func (s *ArchitectureIntegrationTestSuite) cleanupDirectories() {
	// Remove all contents but keep directories
	entries, _ := os.ReadDir(s.projectsDir)
	for _, entry := range entries {
		os.RemoveAll(filepath.Join(s.projectsDir, entry.Name()))
	}

	entries, _ = os.ReadDir(s.workspacesDir)
	for _, entry := range entries {
		os.RemoveAll(filepath.Join(s.workspacesDir, entry.Name()))
	}
}

func (s *ArchitectureIntegrationTestSuite) TestCorrectArchitecture_ProjectsInProjectsDir() {
	// Create project repositories in Projects directory
	_ = s.createProject("project1")
	_ = s.createProject("project2")

	// Verify projects are found in Projects directory
	// Convert absolute path to relative path for FileSystem
	relativeProjectsPath := "Projects"
	projects, err := s.discoveryService.DiscoverProjects(context.Background(), relativeProjectsPath)
	s.Require().NoError(err)

	// Should find both projects
	s.Assert().Len(projects, 2, "Should discover exactly 2 projects")

	projectNames := make([]string, len(projects))
	for i, p := range projects {
		projectNames[i] = p.Name
	}

	s.Assert().Contains(projectNames, "project1")
	s.Assert().Contains(projectNames, "project2")

	// Verify project paths are correct
	for _, p := range projects {
		s.Assert().Equal(filepath.Join(s.projectsDir, p.Name), p.GitRepo)
	}
}

func (s *ArchitectureIntegrationTestSuite) TestCorrectArchitecture_WorktreesInWorkspacesDir() {
	// Create project in Projects directory
	project1 := s.createProject("project1")

	// Create worktrees in Workspaces directory
	worktree1Path := filepath.Join(s.workspacesDir, "project1", "feature1")
	worktree2Path := filepath.Join(s.workspacesDir, "project1", "feature2")

	s.createWorktree(project1, "feature1", worktree1Path)
	s.createWorktree(project1, "feature2", worktree2Path)

	// Discover worktrees in Workspaces directory
	// Convert absolute path to relative path for FileSystem
	relativeWorkspacesPath := "Workspaces"
	worktrees, err := s.discoveryService.DiscoverWorktrees(context.Background(), relativeWorkspacesPath)
	s.Require().NoError(err)

	// Should find both worktrees
	s.Assert().GreaterOrEqual(len(worktrees), 2, "Should discover at least 2 worktrees")

	// Verify worktree paths and branches
	var foundFeature1, foundFeature2 bool
	for _, wt := range worktrees {
		if wt.Path == worktree1Path {
			foundFeature1 = true
			s.Assert().Equal("feature1", wt.Branch)
		}
		if wt.Path == worktree2Path {
			foundFeature2 = true
			s.Assert().Equal("feature2", wt.Branch)
		}
	}

	s.Assert().True(foundFeature1, "Should find feature1 worktree")
	s.Assert().True(foundFeature2, "Should find feature2 worktree")
}

func (s *ArchitectureIntegrationTestSuite) TestCorrectArchitecture_CrossReferenceProjectsAndWorktrees() {
	// Create multiple projects
	project1 := s.createProject("project1")
	project2 := s.createProject("project2")

	// Create worktrees for different projects
	worktree1Path := filepath.Join(s.workspacesDir, "project1", "feature1")
	worktree2Path := filepath.Join(s.workspacesDir, "project2", "develop")

	s.createWorktree(project1, "feature1", worktree1Path)
	s.createWorktree(project2, "develop", worktree2Path)

	// Discover projects
	relativeProjectsPath := "Projects"
	projects, err := s.discoveryService.DiscoverProjects(context.Background(), relativeProjectsPath)
	s.Require().NoError(err)
	s.Assert().Len(projects, 2, "Should discover 2 projects")

	// Discover worktrees
	relativeWorkspacesPath := "Workspaces"
	worktrees, err := s.discoveryService.DiscoverWorktrees(context.Background(), relativeWorkspacesPath)
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(worktrees), 2, "Should discover at least 2 worktrees")

	// Verify worktrees are correctly associated with their projects
	// This test will fail with current implementation because discovery logic is wrong
	s.T().Log("This test should fail with current implementation")
	s.T().Logf("Found %d worktrees: %v", len(worktrees), worktreePaths(worktrees))

	// The current implementation treats projects as worktrees, so it will find
	// the project directories themselves as worktrees, which is incorrect
	for _, wt := range worktrees {
		// Worktrees should NOT be in the Projects directory
		s.Assert().False(
			isSubpath(wt.Path, s.projectsDir),
			"Worktree %s should not be in Projects directory", wt.Path,
		)

		// Worktrees should be in the Workspaces directory
		s.Assert().True(
			isSubpath(wt.Path, s.workspacesDir),
			"Worktree %s should be in Workspaces directory", wt.Path,
		)
	}
}

func (s *ArchitectureIntegrationTestSuite) TestCorrectArchitecture_CommitTimestamps() {
	// Create project with a commit that has a known timestamp
	project1 := s.createProject("project1")

	// Add a commit with a specific timestamp
	commitTime := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	s.addCommitWithTimestamp(project1, "Initial commit", commitTime)

	// Get the current commit hash
	cmd2 := exec.Command("git", "rev-parse", "HEAD")
	cmd2.Dir = project1.Path
	output2, err2 := cmd2.CombinedOutput()
	s.Require().NoError(err2, "Failed to get HEAD hash: %s", string(output2))
	headHash := strings.TrimSpace(string(output2))
	s.T().Logf("Current commit hash: %s", headHash)

	// Create worktree from the current commit
	worktreePath := filepath.Join(s.workspacesDir, "project1", "timestamp-test")
	s.createWorktreeFromCommit(project1, headHash, worktreePath)

	// Discover worktrees
	relativeWorkspacesPath := "Workspaces"
	worktrees, err := s.discoveryService.DiscoverWorktrees(context.Background(), relativeWorkspacesPath)
	s.Require().NoError(err)

	// Find our worktree
	var targetWorktree *domain.Worktree
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			targetWorktree = wt
			break
		}
	}

	s.Require().NotNil(targetWorktree, "Should find the created worktree")

	// Check the worktree's commit directly
	cmd := exec.Command("git", "log", "-1", "--format=%ad", "--date=iso")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get worktree commit timestamp: %s", string(output))
	worktreeTimestamp := strings.TrimSpace(string(output))
	s.T().Logf("Worktree commit timestamp from git log: %s", worktreeTimestamp)

	// Check worktree HEAD commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = worktreePath
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get worktree HEAD hash: %s", string(output))
	worktreeHeadHash := strings.TrimSpace(string(output))
	s.T().Logf("Worktree HEAD commit hash: %s", worktreeHeadHash)

	// Verify that the implementation now uses commit timestamp instead of time.Now()
	s.T().Log("Verifying that worktree LastUpdated uses commit timestamp")
	s.T().Logf("Worktree LastUpdated: %v", targetWorktree.LastUpdated)
	s.T().Logf("Expected commit time: %v", commitTime)

	// The implementation should now use the actual commit timestamp instead of time.Now()
	// Compare the time values directly, ignoring timezone differences
	s.Assert().Equal(
		commitTime.Truncate(time.Second).Unix(),
		targetWorktree.LastUpdated.Truncate(time.Second).Unix(),
		"Worktree LastUpdated should use commit timestamp, not current time",
	)
}

func (s *ArchitectureIntegrationTestSuite) TestCorrectArchitecture_ListCommandShowsOnlyWorktrees() {
	// Create project in Projects directory
	project1 := s.createProject("project1")

	// Create worktree in Workspaces directory
	worktreePath := filepath.Join(s.workspacesDir, "project1", "feature1")
	s.createWorktree(project1, "feature1", worktreePath)

	// Test discovery from Workspaces directory (what the CLI command should do)
	relativeWorkspacesPath := "Workspaces"
	worktrees, err := s.discoveryService.DiscoverWorktrees(context.Background(), relativeWorkspacesPath)
	s.Require().NoError(err)

	// Should find the worktree but NOT the project directory
	s.Assert().GreaterOrEqual(len(worktrees), 1, "Should discover at least the worktree")

	// Verify we found the worktree
	var foundWorktree bool
	for _, wt := range worktrees {
		if wt.Path == worktreePath {
			foundWorktree = true
			break
		}
	}
	s.Assert().True(foundWorktree, "Should find the worktree")

	// Should NOT find the project directory as a worktree
	projectPath := filepath.Join(s.projectsDir, "project1")
	for _, wt := range worktrees {
		s.Assert().NotEqual(
			projectPath,
			wt.Path,
			"Project directory should not be listed as a worktree",
		)
	}

	s.T().Log("This test should fail with current implementation")
	s.T().Logf("Current implementation will incorrectly list project %s as a worktree", projectPath)
}

// Helper methods

func (s *ArchitectureIntegrationTestSuite) createProject(name string) *helpers.GitRepo {
	projectPath := filepath.Join(s.projectsDir, name)

	// Create git repository
	repo := helpers.NewGitRepo(s.T(), name)

	// Move it to the correct location
	err := os.Rename(repo.Path, projectPath)
	s.Require().NoError(err)
	repo.Path = projectPath

	// Add initial commit (already done by NewGitRepo)

	return repo
}

func (s *ArchitectureIntegrationTestSuite) createWorktree(project *helpers.GitRepo, branch, worktreePath string) {
	// Create branch in project
	project.CreateBranch(s.T(), branch)

	// Create parent directory for worktree
	parentDir := filepath.Dir(worktreePath)
	err := os.MkdirAll(parentDir, 0755)
	s.Require().NoError(err)

	// Create worktree using git client
	err = s.gitClient.CreateWorktree(context.Background(), project.Path, branch, worktreePath)
	s.Require().NoError(err)
}

func (s *ArchitectureIntegrationTestSuite) createWorktreeFromCommit(project *helpers.GitRepo, commitHash, worktreePath string) {
	// Create parent directory for worktree
	parentDir := filepath.Dir(worktreePath)
	err := os.MkdirAll(parentDir, 0755)
	s.Require().NoError(err)

	// Create a temporary branch from the commit hash
	branchName := "timestamp-test-branch"
	cmd := exec.Command("git", "branch", branchName, commitHash)
	cmd.Dir = project.Path
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to create branch from commit: %s", string(output))

	// Create worktree using git client (which uses go-git)
	err = s.gitClient.CreateWorktree(context.Background(), project.Path, branchName, worktreePath)
	s.Require().NoError(err, "Failed to create worktree from branch: %v", err)

	// Verify worktree was created and has correct commit
	cmd = exec.Command("git", "log", "-1", "--format=%H")
	cmd.Dir = worktreePath
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get worktree HEAD: %s", string(output))
	worktreeHead := strings.TrimSpace(string(output))
	s.T().Logf("Worktree HEAD commit: %s", worktreeHead)
	s.Assert().Equal(commitHash, worktreeHead, "Worktree should point to correct commit")
}

func (s *ArchitectureIntegrationTestSuite) addCommitWithTimestamp(repo *helpers.GitRepo, message string, timestamp time.Time) {
	// Create a file and commit it with specific timestamp
	testFile := filepath.Join(repo.Path, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	s.Require().NoError(err)

	// Use git command to add and commit with specific timestamp
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = repo.Path
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to add file: %s", string(output))

	// Set environment variables for commit timestamp
	timestampStr := timestamp.Format(time.RFC3339)
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repo.Path
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+timestampStr,
		"GIT_COMMITTER_DATE="+timestampStr,
	)
	s.T().Logf("Setting GIT_AUTHOR_DATE=%s", timestampStr)
	s.T().Logf("Setting GIT_COMMITTER_DATE=%s", timestampStr)
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to commit: %s", string(output))

	// Verify the commit timestamp was set correctly
	cmd = exec.Command("git", "log", "-1", "--format=%ad", "--date=iso")
	cmd.Dir = repo.Path
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get commit timestamp: %s", string(output))
	actualTimestamp := strings.TrimSpace(string(output))
	s.T().Logf("Actual commit timestamp from git log: %s", actualTimestamp)

	// Show all commits
	cmd = exec.Command("git", "log", "--oneline")
	cmd.Dir = repo.Path
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get commit log: %s", string(output))
	s.T().Logf("All commits:\n%s", string(output))

	// Show HEAD commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repo.Path
	output, err = cmd.CombinedOutput()
	s.Require().NoError(err, "Failed to get HEAD hash: %s", string(output))
	headHash := strings.TrimSpace(string(output))
	s.T().Logf("HEAD commit hash: %s", headHash)

	s.T().Logf("Added commit '%s' with timestamp %s", message, timestampStr)
}

// Helper functions

func worktreePaths(worktrees []*domain.Worktree) []string {
	paths := make([]string, len(worktrees))
	for i, wt := range worktrees {
		paths[i] = wt.Path
	}
	return paths
}

func isSubpath(path, parent string) bool {
	relPath, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(relPath) && !strings.HasPrefix(relPath, "..")
}
