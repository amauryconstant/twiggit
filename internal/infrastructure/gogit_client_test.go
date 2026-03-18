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
	tempDir := t.TempDir()

	repo, err := client.OpenRepository("/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, repo)

	repo, err = client.OpenRepository(tempDir)
	require.Error(t, err)
	assert.Nil(t, repo)
}

func TestGoGitClient_ValidateRepository(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	err := client.ValidateRepository("/non/existent/path")
	require.Error(t, err)

	err = client.ValidateRepository(tempDir)
	require.Error(t, err)

	repoPath := setupTestRepo(t, tempDir)
	err = client.ValidateRepository(repoPath)
	require.NoError(t, err)
}

func TestGoGitClient_ListBranches(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	branches, err := client.ListBranches(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, branches)

	repoPath := setupTestRepo(t, tempDir)
	branches, err = client.ListBranches(context.Background(), repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, branches)

	mainBranch := findBranch(branches, "main")
	assert.NotNil(t, mainBranch)
	assert.Equal(t, "main", mainBranch.Name)
}

func TestGoGitClient_BranchExists(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	exists, err := client.BranchExists(context.Background(), "/non/existent/path", "main")
	require.Error(t, err)
	assert.False(t, exists)

	repoPath := setupTestRepo(t, tempDir)

	exists, err = client.BranchExists(context.Background(), repoPath, "main")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = client.BranchExists(context.Background(), repoPath, "non-existent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGoGitClient_GetRepositoryStatus(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	status, err := client.GetRepositoryStatus(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Equal(t, domain.RepositoryStatus{}, status)

	repoPath := setupTestRepo(t, tempDir)
	status, err = client.GetRepositoryStatus(context.Background(), repoPath)
	require.NoError(t, err)
	assert.True(t, status.IsClean)
	assert.Equal(t, "main", status.Branch)
}

func TestGoGitClient_GetRepositoryInfo(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	info, err := client.GetRepositoryInfo(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, info)

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
	tempDir := t.TempDir()

	remotes, err := client.ListRemotes(context.Background(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, remotes)

	repoPath := setupTestRepo(t, tempDir)
	remotes, err = client.ListRemotes(context.Background(), repoPath)
	require.NoError(t, err)
	assert.Empty(t, remotes)
}

func TestGoGitClient_GetCommitInfo(t *testing.T) {
	client := NewGoGitClient()
	tempDir := t.TempDir()

	commit, err := client.GetCommitInfo(context.Background(), "/non/existent/path", "HEAD")
	require.Error(t, err)
	assert.Nil(t, commit)

	repoPath := setupTestRepo(t, tempDir)
	commit, err = client.GetCommitInfo(context.Background(), repoPath, "HEAD")
	require.Error(t, err)
	assert.Nil(t, commit)
}

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

	client := NewGoGitClient()
	repoPath := filepath.Join(tempDir, "test-repo")

	err := os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	_, err = client.OpenRepository(repoPath)
	if err == nil {
		return repoPath
	}

	gitDir := filepath.Join(repoPath, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	require.NoError(t, err)

	refsDir := filepath.Join(gitDir, "refs", "heads")
	err = os.MkdirAll(refsDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(refsDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0644)
	require.NoError(t, err)

	return repoPath
}
