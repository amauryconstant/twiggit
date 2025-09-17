//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/git"
	"github.com/amaury/twiggit/internal/testutil"
	"github.com/amaury/twiggit/internal/worktree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestRepo wraps testutil.GitRepo with integration-specific functionality
type IntegrationTestRepo struct {
	*testutil.GitRepo
	TempDir string
}

// NewTestGitRepo creates a new git repository for integration testing
func NewTestGitRepo(t *testing.T) *IntegrationTestRepo {
	tempDir, cleanup := testutil.TempDir(t, "twiggit-integration-*")

	repoDir := filepath.Join(tempDir, "test-repo")
	testutil.MustMkdirAll(t, repoDir, 0755)

	// Create git repo directly in the test-repo subdirectory
	repo := testutil.NewGitRepo(t, "twiggit-integration-*")

	// Move the repo content to our desired structure
	err := os.Rename(repo.Path, repoDir)
	require.NoError(t, err)
	repo.Path = repoDir

	integrationRepo := &IntegrationTestRepo{
		GitRepo: repo,
		TempDir: tempDir,
	}

	// Override cleanup to clean both temp dir and repo
	originalCleanup := repo.Cleanup
	integrationRepo.GitRepo.Cleanup = func() {
		if originalCleanup != nil {
			originalCleanup()
		}
		cleanup()
	}

	return integrationRepo
}

// RepoDir returns the repository directory (alias for Path for backward compatibility)
func (r *IntegrationTestRepo) RepoDir() string {
	return r.Path
}

// TestIntegration_FullWorktreeLifecycle tests the complete worktree lifecycle
func TestIntegration_FullWorktreeLifecycle(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create test git repository
	testRepo := NewTestGitRepo(t)
	defer testRepo.Cleanup()

	// Create some branches for testing
	testRepo.CreateBranch(t, "feature-1")
	testRepo.CreateBranch(t, "feature-2")
	testRepo.AddMiseConfig(t)

	// Initialize services
	gitClient := git.NewClient()
	discoveryService := worktree.NewDiscoveryService(gitClient)
	config := &config.Config{
		Workspace: testRepo.TempDir,
	}
	operationsService := worktree.NewOperationsService(gitClient, discoveryService, config)

	t.Run("should create worktree from existing branch", func(t *testing.T) {
		worktreePath := filepath.Join(testRepo.TempDir, "feature-1-worktree")

		// Verify the branch exists before creating worktree
		exists := gitClient.BranchExists(testRepo.RepoDir, "feature-1")
		assert.True(t, exists, "feature-1 branch should exist before creating worktree")

		err := operationsService.Create(testRepo.RepoDir, "feature-1", worktreePath)
		assert.NoError(t, err)

		// Verify worktree was created
		_, err = os.Stat(worktreePath)
		assert.NoError(t, err, "Worktree directory should exist")

		// Verify it's a valid git worktree
		isRepo, err := gitClient.IsGitRepository(worktreePath)
		assert.NoError(t, err)
		assert.True(t, isRepo, "Worktree should be a valid git repository")

		// Verify branch is checked out correctly
		status, err := gitClient.GetWorktreeStatus(worktreePath)
		assert.NoError(t, err)
		assert.Equal(t, "feature-1", status.Branch)

		// Verify mise config was copied
		miseFile := filepath.Join(worktreePath, ".mise.local.toml")
		_, err = os.Stat(miseFile)
		assert.NoError(t, err, "Mise configuration should be copied to worktree")
	})

	t.Run("should create worktree for new branch", func(t *testing.T) {
		worktreePath := filepath.Join(testRepo.TempDir, "new-feature-worktree")

		err := operationsService.Create(testRepo.RepoDir, "new-feature", worktreePath)
		assert.NoError(t, err)

		// Verify worktree was created
		_, err = os.Stat(worktreePath)
		assert.NoError(t, err, "Worktree directory should exist")

		// Verify branch exists now
		exists := gitClient.BranchExists(testRepo.RepoDir, "new-feature")
		assert.True(t, exists, "New branch should have been created")
	})

	t.Run("should list all worktrees", func(t *testing.T) {
		worktrees, err := gitClient.ListWorktrees(testRepo.RepoDir)
		assert.NoError(t, err)

		// Should have main repo + 2 worktrees
		assert.GreaterOrEqual(t, len(worktrees), 3, "Should have main repo and 2 worktrees")

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
		assert.True(t, foundFeature1, "Should find feature-1 worktree")
		assert.True(t, foundNewFeature, "Should find new-feature worktree")
	})

	t.Run("should remove worktree safely", func(t *testing.T) {
		worktreePath := filepath.Join(testRepo.TempDir, "feature-1-worktree")

		// Verify it exists first
		_, err := os.Stat(worktreePath)
		assert.NoError(t, err)

		// Remove the worktree (use force since mise config was copied)
		err = operationsService.Remove(worktreePath, true)
		assert.NoError(t, err)

		// Verify it was removed
		_, err = os.Stat(worktreePath)
		assert.True(t, os.IsNotExist(err), "Worktree directory should be removed")

		// Verify it's no longer in the worktree list
		worktrees, err := gitClient.ListWorktrees(testRepo.RepoDir)
		assert.NoError(t, err)

		for _, wt := range worktrees {
			assert.NotContains(t, wt.Path, "feature-1-worktree", "Removed worktree should not be in list")
		}
	})
}

// TestIntegration_DiscoveryService tests worktree discovery with real repositories
func TestIntegration_DiscoveryService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create workspace with multiple projects
	workspaceDir, err := os.MkdirTemp("", "twiggit-workspace-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceDir)

	// Create project 1
	project1 := NewTestGitRepo(t)
	project1Path := filepath.Join(workspaceDir, "project1")
	err = os.Rename(project1.RepoDir, project1Path)
	require.NoError(t, err)
	project1.RepoDir = project1Path

	// Create project 2
	project2 := NewTestGitRepo(t)
	project2Path := filepath.Join(workspaceDir, "project2")
	err = os.Rename(project2.RepoDir, project2Path)
	require.NoError(t, err)
	project2.RepoDir = project2Path

	// Create some worktrees
	gitClient := git.NewClient()

	worktree1Path := filepath.Join(workspaceDir, "project1-feature")
	err = gitClient.CreateWorktree(project1Path, "feature-branch-1", worktree1Path)
	require.NoError(t, err)

	worktree2Path := filepath.Join(workspaceDir, "project2-feature")
	err = gitClient.CreateWorktree(project2Path, "feature-branch-2", worktree2Path)
	require.NoError(t, err)

	// Test discovery
	discoveryService := worktree.NewDiscoveryService(gitClient)

	t.Run("should discover all projects", func(t *testing.T) {
		projects, err := discoveryService.DiscoverProjects(workspaceDir)
		assert.NoError(t, err)
		assert.Len(t, projects, 2, "Should discover 2 projects")

		projectNames := make([]string, len(projects))
		for i, p := range projects {
			projectNames[i] = p.Name
		}
		assert.Contains(t, projectNames, "project1")
		assert.Contains(t, projectNames, "project2")
	})

	t.Run("should discover all worktrees", func(t *testing.T) {
		worktrees, err := discoveryService.DiscoverWorktrees(workspaceDir)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(worktrees), 4, "Should discover at least 4 worktrees (2 main repos + 2 additional)")

		// Check that our specific worktrees are found
		worktreePaths := make([]string, len(worktrees))
		for i, wt := range worktrees {
			worktreePaths[i] = wt.Path
		}
		assert.Contains(t, worktreePaths, project1Path)
		assert.Contains(t, worktreePaths, project2Path)
		assert.Contains(t, worktreePaths, worktree1Path)
		assert.Contains(t, worktreePaths, worktree2Path)
	})
}

// TestIntegration_ErrorHandling tests error conditions with real repositories
func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testRepo := NewTestGitRepo(t)
	defer testRepo.Cleanup()

	gitClient := git.NewClient()
	discoveryService := worktree.NewDiscoveryService(gitClient)
	config := &config.Config{
		Workspace: testRepo.TempDir,
	}
	operationsService := worktree.NewOperationsService(gitClient, discoveryService, config)

	t.Run("should handle non-existent repository", func(t *testing.T) {
		err := operationsService.Create("/non/existent/repo", "feature", "/tmp/test-worktree")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
	})

	t.Run("should handle invalid target path", func(t *testing.T) {
		err := operationsService.Create(testRepo.RepoDir, "feature", "relative/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path must be absolute")
	})

	t.Run("should handle existing target directory", func(t *testing.T) {
		existingPath := filepath.Join(testRepo.TempDir, "existing")
		err := os.MkdirAll(existingPath, 0755)
		require.NoError(t, err)

		err = operationsService.Create(testRepo.RepoDir, "feature", existingPath)
		assert.Error(t, err)
	})

	t.Run("should handle removal of non-existent worktree", func(t *testing.T) {
		err := operationsService.Remove("/non/existent/worktree", false)
		assert.Error(t, err)
	})
}

// TestIntegration_Performance tests performance characteristics
func TestIntegration_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create workspace with multiple projects and worktrees
	workspaceDir, err := os.MkdirTemp("", "twiggit-perf-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceDir)

	projectCount := 5
	worktreesPerProject := 3

	gitClient := git.NewClient()

	// Create multiple projects
	for i := 0; i < projectCount; i++ {
		projectRepo := NewTestGitRepo(t)
		projectPath := filepath.Join(workspaceDir, fmt.Sprintf("project%d", i))
		err = os.Rename(projectRepo.RepoDir, projectPath)
		require.NoError(t, err)

		// Create branches and worktrees for each project
		for j := 0; j < worktreesPerProject; j++ {
			branchName := fmt.Sprintf("feature-%d", j)
			worktreePath := filepath.Join(workspaceDir, fmt.Sprintf("project%d-%s", i, branchName))

			err = gitClient.CreateWorktree(projectPath, branchName, worktreePath)
			require.NoError(t, err)
		}
	}

	discoveryService := worktree.NewDiscoveryService(gitClient)
	discoveryService.SetConcurrency(4) // Test with concurrent processing

	t.Run("should discover projects efficiently", func(t *testing.T) {
		projects, err := discoveryService.DiscoverProjects(workspaceDir)
		assert.NoError(t, err)
		assert.Len(t, projects, projectCount, "Should discover all projects")

		// Each project should have the expected number of worktrees
		for _, project := range projects {
			assert.GreaterOrEqual(t, len(project.Worktrees), worktreesPerProject,
				"Project %s should have at least %d worktrees", project.Name, worktreesPerProject)
		}
	})

	t.Run("should discover all worktrees efficiently", func(t *testing.T) {
		worktrees, err := discoveryService.DiscoverWorktrees(workspaceDir)
		assert.NoError(t, err)

		expectedCount := projectCount * (worktreesPerProject + 1) // +1 for main repo
		assert.GreaterOrEqual(t, len(worktrees), expectedCount,
			"Should discover at least %d worktrees", expectedCount)
	})
}
