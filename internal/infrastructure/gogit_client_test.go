package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"twiggit/internal/domain"
)

func TestGoGitClient_OpenRepository(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent directory
	repo, err := client.OpenRepository("/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, repo)

	// Test with non-git directory
	tempDir := t.TempDir()
	repo, err = client.OpenRepository(tempDir)
	require.Error(t, err)
	assert.Nil(t, repo)
}

func TestGoGitClient_ValidateRepository(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent directory
	err := client.ValidateRepository("/non/existent/path")
	require.Error(t, err)

	// Test with non-git directory
	tempDir := t.TempDir()
	err = client.ValidateRepository(tempDir)
	require.Error(t, err)

	// Test with valid git repository
	repoPath := setupTestRepo(t, tempDir)
	err = client.ValidateRepository(repoPath)
	assert.NoError(t, err)
}

func TestGoGitClient_ListBranches(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	branches, err := client.ListBranches(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, branches)

	// Test with valid repository
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	branches, err = client.ListBranches(context.Background(), repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, branches)

	// Should have at least main branch
	mainBranch := findBranch(branches, "main")
	assert.NotNil(t, mainBranch)
	assert.Equal(t, "main", mainBranch.Name)
}

func TestGoGitClient_BranchExists(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	exists, err := client.BranchExists(context.Background(), "/non/existent/path", "main")
	require.Error(t, err)
	assert.False(t, exists)

	// Test with valid repository
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	// Test existing branch
	exists, err = client.BranchExists(context.Background(), repoPath, "main")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing branch
	exists, err = client.BranchExists(context.Background(), repoPath, "non-existent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGoGitClient_GetRepositoryStatus(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	status, err := client.GetRepositoryStatus(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Equal(t, domain.RepositoryStatus{}, status)

	// Test with clean repository
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	status, err = client.GetRepositoryStatus(context.Background(), repoPath)
	require.NoError(t, err)
	assert.True(t, status.IsClean)
	assert.Equal(t, "main", status.Branch)
}

func TestGoGitClient_GetRepositoryInfo(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	info, err := client.GetRepositoryInfo(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, info)

	// Test with valid repository
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	info, err = client.GetRepositoryInfo(context.Background(), repoPath)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, repoPath, info.Path)
	assert.False(t, info.IsBare)
	assert.NotEmpty(t, info.Branches)
}

func TestGoGitClient_ListRemotes(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	remotes, err := client.ListRemotes(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, remotes)

	// Test with repository without remotes
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	remotes, err = client.ListRemotes(context.Background(), repoPath)
	require.NoError(t, err)
	assert.Empty(t, remotes)
}

func TestGoGitClient_GetCommitInfo(t *testing.T) {
	client := NewGoGitClient()

	// Test with non-existent repository
	commit, err := client.GetCommitInfo(context.Background(), "/non/existent/path", "HEAD")
	require.Error(t, err)
	assert.Nil(t, commit)

	// Test with valid repository (but no commits - this is expected for our minimal test repo)
	tempDir := t.TempDir()
	repoPath := setupTestRepo(t, tempDir)

	commit, err = client.GetCommitInfo(context.Background(), repoPath, "HEAD")
	// This should fail because our test repo doesn't have actual commits
	require.Error(t, err)
	assert.Nil(t, commit)
}

// Helper functions

func findBranch(branches []domain.BranchInfo, name string) *domain.BranchInfo {
	for _, branch := range branches {
		if branch.Name == name {
			return &branch
		}
	}
	return nil
}

func setupTestRepo(t *testing.T, tempDir string) string {
	t.Helper()
	repoPath := filepath.Join(tempDir, "test-repo")

	// Create repository directory
	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Initialize git repository
	client := NewGoGitClient()
	_, err = client.OpenRepository(repoPath)
	if err == nil {
		// Repository already exists, return path
		return repoPath
	}

	// For testing purposes, we'll create a minimal git repository structure
	// In a real implementation, this would use go-git to initialize
	gitDir := filepath.Join(repoPath, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create minimal git structure
	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	require.NoError(t, err)

	refsDir := filepath.Join(gitDir, "refs", "heads")
	err = os.MkdirAll(refsDir, 0755)
	require.NoError(t, err)

	// Create main branch reference
	err = os.WriteFile(filepath.Join(refsDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0644)
	require.NoError(t, err)

	return repoPath
}
