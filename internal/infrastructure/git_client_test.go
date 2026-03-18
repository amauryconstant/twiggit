package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestGitClient_OpenRepository_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	repoPath := "/path/to/repo"
	expectedErr := errors.New("failed to open repository")

	mockGoGitClient.On("OpenRepository", repoPath).Return(nil, expectedErr)

	repo, err := compositeClient.OpenRepository(repoPath)

	require.Error(t, err)
	assert.Nil(t, repo)
	var gitRepoErr *domain.GitRepositoryError
	assert.ErrorAs(t, err, &gitRepoErr)
	assert.Contains(t, gitRepoErr.Error(), "failed to open repository")
}

func TestGitClient_ListBranches_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedBranches := []domain.BranchInfo{
		{Name: "main", IsCurrent: true},
		{Name: "feature", IsCurrent: false},
	}

	mockGoGitClient.On("ListBranches", ctx, repoPath).Return(expectedBranches, nil)

	branches, err := compositeClient.ListBranches(ctx, repoPath)

	require.NoError(t, err)
	assert.Equal(t, expectedBranches, branches)
}

func TestGitClient_BranchExists_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "main"

	mockGoGitClient.On("BranchExists", ctx, repoPath, branchName).Return(true, nil)

	exists, err := compositeClient.BranchExists(ctx, repoPath, branchName)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestGitClient_GetRepositoryStatus_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedStatus := domain.RepositoryStatus{
		IsClean: true,
		Branch:  "main",
	}

	mockGoGitClient.On("GetRepositoryStatus", ctx, repoPath).Return(expectedStatus, nil)

	status, err := compositeClient.GetRepositoryStatus(ctx, repoPath)

	require.NoError(t, err)
	assert.Equal(t, expectedStatus, status)
}

func TestGitClient_ValidateRepository_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	repoPath := "/path/to/repo"

	mockGoGitClient.On("ValidateRepository", repoPath).Return(nil)

	err := compositeClient.ValidateRepository(repoPath)

	require.NoError(t, err)
}

func TestGitClient_GetRepositoryInfo_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedInfo := &domain.GitRepository{
		Path:          repoPath,
		DefaultBranch: "main",
	}

	mockGoGitClient.On("GetRepositoryInfo", ctx, repoPath).Return(expectedInfo, nil)

	info, err := compositeClient.GetRepositoryInfo(ctx, repoPath)

	require.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
}

func TestGitClient_ListRemotes_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedRemotes := []domain.RemoteInfo{
		{Name: "origin", FetchURL: "git@github.com:user/repo.git"},
	}

	mockGoGitClient.On("ListRemotes", ctx, repoPath).Return(expectedRemotes, nil)

	remotes, err := compositeClient.ListRemotes(ctx, repoPath)

	require.NoError(t, err)
	assert.Equal(t, expectedRemotes, remotes)
}

func TestGitClient_GetCommitInfo_RoutesToGoGitClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockGoGitClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	commitHash := "abc123"
	expectedInfo := &domain.CommitInfo{
		Hash:    commitHash,
		Message: "Test commit",
	}

	mockGoGitClient.On("GetCommitInfo", ctx, repoPath, commitHash).Return(expectedInfo, nil)

	info, err := compositeClient.GetCommitInfo(ctx, repoPath, commitHash)

	require.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
}

func TestGitClient_CreateWorktree_RoutesToCLIClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockCLIClient.AssertExpectations(t)
		mockGoGitClient.AssertNotCalled(t, "CreateWorktree")
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"
	sourceBranch := "main"
	worktreePath := "/path/to/worktree"

	mockCLIClient.On("CreateWorktree", ctx, repoPath, branchName, sourceBranch, worktreePath).Return(nil)

	err := compositeClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)

	require.NoError(t, err)
}

func TestGitClient_CreateWorktree_ReturnsError(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)

	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"
	sourceBranch := "main"
	worktreePath := "/path/to/worktree"
	expectedErr := errors.New("failed to create worktree")

	mockCLIClient.On("CreateWorktree", ctx, repoPath, branchName, sourceBranch, worktreePath).Return(expectedErr)

	err := compositeClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)

	require.Error(t, err)
	var gitWorktreeErr *domain.GitWorktreeError
	assert.ErrorAs(t, err, &gitWorktreeErr)
	assert.Contains(t, gitWorktreeErr.Error(), "failed to create worktree")
}

func TestGitClient_DeleteWorktree_RoutesToCLIClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockCLIClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	worktreePath := "/path/to/worktree"
	force := true

	mockCLIClient.On("DeleteWorktree", ctx, repoPath, worktreePath, force).Return(nil)

	err := compositeClient.DeleteWorktree(ctx, repoPath, worktreePath, force)

	require.NoError(t, err)
}

func TestGitClient_ListWorktrees_RoutesToCLIClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockCLIClient.AssertExpectations(t)
		mockGoGitClient.AssertNotCalled(t, "ListWorktrees")
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedWorktrees := []domain.WorktreeInfo{
		{Path: "/path/to/worktree1", Branch: "feature1"},
		{Path: "/path/to/worktree2", Branch: "feature2"},
	}

	mockCLIClient.On("ListWorktrees", ctx, repoPath).Return(expectedWorktrees, nil)

	worktrees, err := compositeClient.ListWorktrees(ctx, repoPath)

	require.NoError(t, err)
	assert.Equal(t, expectedWorktrees, worktrees)
}

func TestGitClient_PruneWorktrees_RoutesToCLIClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockCLIClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"

	mockCLIClient.On("PruneWorktrees", ctx, repoPath).Return(nil)

	err := compositeClient.PruneWorktrees(ctx, repoPath)

	require.NoError(t, err)
}

func TestGitClient_IsBranchMerged_RoutesToCLIClient(t *testing.T) {
	mockGoGitClient := mocks.NewMockGoGitClient()
	mockCLIClient := mocks.NewMockCLIClient()
	compositeClient := NewCompositeGitClient(mockGoGitClient, mockCLIClient)
	t.Cleanup(func() {
		mockCLIClient.AssertExpectations(t)
	})

	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"

	mockCLIClient.On("IsBranchMerged", ctx, repoPath, branchName).Return(true, nil)

	merged, err := compositeClient.IsBranchMerged(ctx, repoPath, branchName)

	require.NoError(t, err)
	assert.True(t, merged)
}
