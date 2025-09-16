package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject_NewProject(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		gitRepo      string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project",
			projectName: "my-project",
			gitRepo:     "/path/to/repo",
			expectError: false,
		},
		{
			name:         "empty project name",
			projectName:  "",
			gitRepo:      "/path/to/repo",
			expectError:  true,
			errorMessage: "project name cannot be empty",
		},
		{
			name:         "empty git repo path",
			projectName:  "my-project",
			gitRepo:      "",
			expectError:  true,
			errorMessage: "git repository path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := NewProject(tt.projectName, tt.gitRepo)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, project)
			} else {
				require.NoError(t, err)
				require.NotNil(t, project)
				assert.Equal(t, tt.projectName, project.Name)
				assert.Equal(t, tt.gitRepo, project.GitRepo)
				assert.Empty(t, project.Worktrees)
			}
		})
	}
}

func TestProject_AddWorktree(t *testing.T) {
	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	worktree, err := NewWorktree("/worktree/path", "feature-branch")
	require.NoError(t, err)

	// Add first worktree
	err = project.AddWorktree(worktree)
	require.NoError(t, err)
	assert.Len(t, project.Worktrees, 1)
	assert.Equal(t, worktree, project.Worktrees[0])

	// Add second worktree
	worktree2, err := NewWorktree("/another/path", "main")
	require.NoError(t, err)

	err = project.AddWorktree(worktree2)
	require.NoError(t, err)
	assert.Len(t, project.Worktrees, 2)

	// Try to add duplicate worktree path
	duplicateWorktree, err := NewWorktree("/worktree/path", "different-branch")
	require.NoError(t, err)

	err = project.AddWorktree(duplicateWorktree)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree already exists at path")
	assert.Len(t, project.Worktrees, 2) // Should not be added
}

func TestProject_RemoveWorktree(t *testing.T) {
	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	worktree, err := NewWorktree("/worktree/path", "feature-branch")
	require.NoError(t, err)

	err = project.AddWorktree(worktree)
	require.NoError(t, err)

	// Remove existing worktree
	err = project.RemoveWorktree("/worktree/path")
	require.NoError(t, err)
	assert.Empty(t, project.Worktrees)

	// Try to remove non-existent worktree
	err = project.RemoveWorktree("/nonexistent/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found at path")
}

func TestProject_GetWorktree(t *testing.T) {
	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	worktree, err := NewWorktree("/worktree/path", "feature-branch")
	require.NoError(t, err)

	err = project.AddWorktree(worktree)
	require.NoError(t, err)

	// Get existing worktree
	found, err := project.GetWorktree("/worktree/path")
	require.NoError(t, err)
	assert.Equal(t, worktree, found)

	// Get non-existent worktree
	_, err = project.GetWorktree("/nonexistent/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found at path")
}

func TestProject_ListBranches(t *testing.T) {
	project, err := NewProject("test-project", "/repo/path")
	require.NoError(t, err)

	// Empty project
	branches := project.ListBranches()
	assert.Empty(t, branches)

	// Add worktrees
	worktree1, _ := NewWorktree("/path1", "main")
	worktree2, _ := NewWorktree("/path2", "feature-1")
	worktree3, _ := NewWorktree("/path3", "feature-2")
	worktree4, _ := NewWorktree("/path4", "main") // Duplicate branch

	require.NoError(t, project.AddWorktree(worktree1))
	require.NoError(t, project.AddWorktree(worktree2))
	require.NoError(t, project.AddWorktree(worktree3))
	require.NoError(t, project.AddWorktree(worktree4))

	branches = project.ListBranches()
	assert.Len(t, branches, 3) // Should deduplicate
	assert.Contains(t, branches, "main")
	assert.Contains(t, branches, "feature-1")
	assert.Contains(t, branches, "feature-2")
}
