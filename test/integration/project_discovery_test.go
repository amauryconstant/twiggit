//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
	"twiggit/test/mocks"
)

func TestProjectDiscovery_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create temporary projects directory
	tempDir := t.TempDir()
	projectsDir := filepath.Join(tempDir, "Projects")
	worktreesDir := filepath.Join(tempDir, "Worktrees")

	// Create project directories
	project1Path := filepath.Join(projectsDir, "project1")
	project2Path := filepath.Join(projectsDir, "test-project")
	nonRepoPath := filepath.Join(projectsDir, "not-a-repo")

	err := os.MkdirAll(project1Path, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(project2Path, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(nonRepoPath, 0755)
	require.NoError(t, err)

	// Initialize git repositories in project1 and project2
	setupTestGitRepo(t, project1Path)
	setupTestGitRepo(t, project2Path)
	// nonRepoPath remains a non-git directory

	// Create configuration
	config := &domain.Config{
		ProjectsDirectory:  projectsDir,
		WorktreesDirectory: worktreesDir,
	}

	// Create mock git service that validates only the git repositories
	mockGitService := mocks.NewMockGitService()
	mockGitService.ValidateRepositoryFunc = func(path string) error {
		switch path {
		case project1Path, project2Path:
			return nil // Valid git repositories
		case nonRepoPath:
			return assert.AnError // Not a git repository
		default:
			return assert.AnError
		}
	}

	// Create context resolver
	resolver := infrastructure.NewContextResolver(config, mockGitService)

	// Test context detection from outside git
	ctx := &domain.Context{
		Type: domain.ContextOutsideGit,
		Path: tempDir,
	}

	// Test suggestions with partial "proj"
	suggestions, err := resolver.GetResolutionSuggestions(ctx, "proj")
	require.NoError(t, err)
	assert.Len(t, suggestions, 1, "Should find 1 project matching 'proj'")
	assert.Equal(t, "project1", suggestions[0].Text)
	assert.Equal(t, domain.PathTypeProject, suggestions[0].Type)
	assert.Equal(t, "project1", suggestions[0].ProjectName)

	// Test suggestions with partial "test"
	suggestions, err = resolver.GetResolutionSuggestions(ctx, "test")
	require.NoError(t, err)
	assert.Len(t, suggestions, 1, "Should find 1 project matching 'test'")
	assert.Equal(t, "test-project", suggestions[0].Text)

	// Test suggestions with empty partial (should return all)
	suggestions, err = resolver.GetResolutionSuggestions(ctx, "")
	require.NoError(t, err)
	assert.Len(t, suggestions, 2, "Should find all 2 projects")

	// Verify suggestion texts
	projectNames := make(map[string]bool)
	for _, suggestion := range suggestions {
		projectNames[suggestion.Text] = true
		assert.Equal(t, domain.PathTypeProject, suggestion.Type)
		assert.Equal(t, "Project directory", suggestion.Description)
	}
	assert.True(t, projectNames["project1"], "Should include project1")
	assert.True(t, projectNames["test-project"], "Should include test-project")
	assert.False(t, projectNames["not-a-repo"], "Should not include non-git directory")

	// Test resolution of project identifier
	result, err := resolver.ResolveIdentifier(ctx, "project1")
	require.NoError(t, err)
	assert.Equal(t, domain.PathTypeProject, result.Type)
	assert.Equal(t, "project1", result.ProjectName)
	assert.Equal(t, project1Path, result.ResolvedPath)
	assert.Contains(t, result.Explanation, "project1")
}

// setupTestGitRepo creates a test git repository with initial commit
func setupTestGitRepo(t *testing.T, repoPath string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(repoPath, 0755))

	// Use command executor to initialize git repo
	executor := infrastructure.NewDefaultCommandExecutor(30 * time.Second)

	// Initialize repository
	_, err := executor.Execute(context.Background(), repoPath, "git", "init")
	require.NoError(t, err)

	// Configure user (required for commits)
	_, err = executor.Execute(context.Background(), repoPath, "git", "config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), repoPath, "git", "config", "user.email", "test@example.com")
	require.NoError(t, err)

	// Create initial commit
	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))
	_, err = executor.Execute(context.Background(), repoPath, "git", "add", "test.txt")
	require.NoError(t, err)
	_, err = executor.Execute(context.Background(), repoPath, "git", "commit", "-m", "Initial commit")
	require.NoError(t, err)

	// Ensure we're on main branch
	_, err = executor.Execute(context.Background(), repoPath, "git", "branch", "-M", "main")
	require.NoError(t, err)
}
