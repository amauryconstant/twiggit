package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_NewWorkspace(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid workspace",
			path:        "/home/user/workspace",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorMessage: "workspace path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, err := NewWorkspace(tt.path)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, workspace)
			} else {
				require.NoError(t, err)
				require.NotNil(t, workspace)
				assert.Equal(t, tt.path, workspace.Path)
				assert.Empty(t, workspace.Projects)
			}
		})
	}
}

func TestWorkspace_AddProject(t *testing.T) {
	workspace, err := NewWorkspace("/workspace")
	require.NoError(t, err)

	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	// Add first project
	err = workspace.AddProject(project)
	require.NoError(t, err)
	assert.Len(t, workspace.Projects, 1)
	assert.Equal(t, project, workspace.Projects[0])

	// Try to add duplicate project
	duplicateProject, err := NewProject("test-project", "/different/repo")
	require.NoError(t, err)

	err = workspace.AddProject(duplicateProject)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project already exists")
	assert.Len(t, workspace.Projects, 1) // Should not be added
}

func TestWorkspace_RemoveProject(t *testing.T) {
	workspace, err := NewWorkspace("/workspace")
	require.NoError(t, err)

	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	err = workspace.AddProject(project)
	require.NoError(t, err)

	// Remove existing project
	err = workspace.RemoveProject("test-project")
	require.NoError(t, err)
	assert.Empty(t, workspace.Projects)

	// Try to remove non-existent project
	err = workspace.RemoveProject("nonexistent-project")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestWorkspace_GetProject(t *testing.T) {
	workspace, err := NewWorkspace("/workspace")
	require.NoError(t, err)

	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	err = workspace.AddProject(project)
	require.NoError(t, err)

	// Get existing project
	found, err := workspace.GetProject("test-project")
	require.NoError(t, err)
	assert.Equal(t, project, found)

	// Get non-existent project
	_, err = workspace.GetProject("nonexistent-project")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

func TestWorkspace_ListAllWorktrees(t *testing.T) {
	workspace, err := NewWorkspace("/workspace")
	require.NoError(t, err)

	// Empty workspace
	worktrees := workspace.ListAllWorktrees()
	assert.Empty(t, worktrees)

	// Add projects with worktrees
	project1, _ := NewProject("project1", "/repo1")
	project2, _ := NewProject("project2", "/repo2")

	worktree1, _ := NewWorktree("/path1", "main")
	worktree2, _ := NewWorktree("/path2", "feature")
	worktree3, _ := NewWorktree("/path3", "develop")

	require.NoError(t, project1.AddWorktree(worktree1))
	require.NoError(t, project1.AddWorktree(worktree2))
	require.NoError(t, project2.AddWorktree(worktree3))

	require.NoError(t, workspace.AddProject(project1))
	require.NoError(t, workspace.AddProject(project2))

	worktrees = workspace.ListAllWorktrees()
	assert.Len(t, worktrees, 3)
	assert.Contains(t, worktrees, worktree1)
	assert.Contains(t, worktrees, worktree2)
	assert.Contains(t, worktrees, worktree3)
}

func TestWorkspace_GetWorktreeByPath(t *testing.T) {
	workspace, err := NewWorkspace("/workspace")
	require.NoError(t, err)

	project, _ := NewProject("project1", "/repo1")
	worktree, _ := NewWorktree("/worktree/path", "main")
	require.NoError(t, project.AddWorktree(worktree))
	require.NoError(t, workspace.AddProject(project))

	// Find existing worktree
	found, err := workspace.GetWorktreeByPath("/worktree/path")
	require.NoError(t, err)
	assert.Equal(t, worktree, found)

	// Try to find non-existent worktree
	_, err = workspace.GetWorktreeByPath("/nonexistent/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestWorkspace_EnhancedFeatures(t *testing.T) {
	t.Run("should validate workspace path existence", func(t *testing.T) {
		workspace, err := NewWorkspace("/non/existent/workspace")
		require.NoError(t, err)

		// This should fail initially - we need to add path validation
		isValid, err := workspace.ValidatePathExists()
		assert.Error(t, err)
		assert.False(t, isValid)
	})

	t.Run("should provide workspace statistics", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// Add projects with worktrees
		project1, _ := NewProject("project1", "/repo1")
		project2, _ := NewProject("project2", "/repo2")

		worktree1, _ := NewWorktree("/path1", "main")
		worktree2, _ := NewWorktree("/path2", "feature")
		worktree3, _ := NewWorktree("/path3", "develop")

		require.NoError(t, project1.AddWorktree(worktree1))
		require.NoError(t, project1.AddWorktree(worktree2))
		require.NoError(t, project2.AddWorktree(worktree3))

		require.NoError(t, workspace.AddProject(project1))
		require.NoError(t, workspace.AddProject(project2))

		// This should fail initially - we need to add statistics
		stats := workspace.GetStatistics()
		assert.NotNil(t, stats)
		assert.Equal(t, 2, stats.ProjectCount)
		assert.Equal(t, 3, stats.TotalWorktreeCount)
		assert.Equal(t, 3, stats.UnknownWorktreeCount)
		assert.Equal(t, 0, stats.CleanWorktreeCount)
		assert.Equal(t, 0, stats.DirtyWorktreeCount)
		assert.Equal(t, 3, len(stats.AllBranches))
	})

	t.Run("should support workspace configuration", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// This should fail initially - we need to add configuration support
		workspace.SetConfig("scan-depth", 3)
		workspace.SetConfig("exclude-patterns", []string{".git", "node_modules"})
		workspace.SetConfig("auto-discover", true)

		value, exists := workspace.GetConfig("scan-depth")
		assert.True(t, exists)
		assert.Equal(t, 3, value)

		value, exists = workspace.GetConfig("auto-discover")
		assert.True(t, exists)
		assert.Equal(t, true, value)

		patterns, exists := workspace.GetConfig("exclude-patterns")
		assert.True(t, exists)
		assert.Equal(t, []string{".git", "node_modules"}, patterns)

		_, exists = workspace.GetConfig("non-existent")
		assert.False(t, exists)
	})

	t.Run("should provide workspace health check", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// This should fail initially - we need to add health check
		health := workspace.GetHealth()
		assert.NotNil(t, health)
		assert.Equal(t, "unhealthy", health.Status)
		assert.Contains(t, health.Issues, "workspace path not validated")
		assert.Equal(t, 0, health.ProjectCount)
		assert.Equal(t, 0, health.WorktreeCount)
	})

	t.Run("should support project discovery", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// This should fail initially - we need to add project discovery
		discovered, err := workspace.DiscoverProjects()
		assert.NoError(t, err) // Should not fail for minimal implementation
		assert.Empty(t, discovered)
	})

	t.Run("should support workspace metadata", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// This should fail initially - we need to add metadata support
		workspace.SetMetadata("created-at", "2023-01-01")
		workspace.SetMetadata("last-scanned", "2023-01-02")
		workspace.SetMetadata("version", "1.0.0")

		value, exists := workspace.GetMetadata("created-at")
		assert.True(t, exists)
		assert.Equal(t, "2023-01-01", value)

		value, exists = workspace.GetMetadata("version")
		assert.True(t, exists)
		assert.Equal(t, "1.0.0", value)

		_, exists = workspace.GetMetadata("non-existent")
		assert.False(t, exists)
	})

	t.Run("should support worktree search and filtering", func(t *testing.T) {
		workspace, err := NewWorkspace("/workspace")
		require.NoError(t, err)

		// Add projects with worktrees
		project1, _ := NewProject("project1", "/repo1")
		project2, _ := NewProject("project2", "/repo2")

		worktree1, _ := NewWorktree("/path1", "main")
		worktree2, _ := NewWorktree("/path2", "feature-1")
		worktree3, _ := NewWorktree("/path3", "feature-2")

		require.NoError(t, project1.AddWorktree(worktree1))
		require.NoError(t, project1.AddWorktree(worktree2))
		require.NoError(t, project2.AddWorktree(worktree3))

		require.NoError(t, workspace.AddProject(project1))
		require.NoError(t, workspace.AddProject(project2))

		// This should fail initially - we need to add search functionality
		mainWorktrees := workspace.FindWorktreesByBranch("main")
		assert.Len(t, mainWorktrees, 1)

		featureWorktrees := workspace.FindWorktreesByBranchPattern("feature-*")
		assert.Len(t, featureWorktrees, 2)

		project1Worktrees := workspace.FindWorktreesByProject("project1")
		assert.Len(t, project1Worktrees, 2)

		cleanWorktrees := workspace.FindWorktreesByStatus(StatusClean)
		assert.Len(t, cleanWorktrees, 0) // None are clean yet
	})
}
