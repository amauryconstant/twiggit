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
