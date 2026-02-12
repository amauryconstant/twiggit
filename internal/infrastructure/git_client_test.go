package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

// GitClientTestSuite tests CompositeGitClient routing
type GitClientTestSuite struct {
	suite.Suite
	mockGoGitClient *mocks.MockGoGitClient
	mockCLIClient   *mocks.MockCLIClient
	compositeClient GitClient
}

func TestGitClientSuite(t *testing.T) {
	suite.Run(t, new(GitClientTestSuite))
}

func (s *GitClientTestSuite) SetupTest() {
	s.mockGoGitClient = mocks.NewMockGoGitClient()
	s.mockCLIClient = mocks.NewMockCLIClient()
	s.compositeClient = NewCompositeGitClient(s.mockGoGitClient, s.mockCLIClient)
}

// Test OpenRepository - routes to GoGitClient
func (s *GitClientTestSuite) TestOpenRepository_RoutesToGoGitClient() {
	repoPath := "/path/to/repo"
	expectedErr := errors.New("failed to open repository")

	s.mockGoGitClient.On("OpenRepository", repoPath).Return(nil, expectedErr)

	repo, err := s.compositeClient.OpenRepository(repoPath)

	s.Require().Error(err)
	s.Nil(repo)
	var gitRepoErr *domain.GitRepositoryError
	s.ErrorAs(err, &gitRepoErr)
	s.Contains(gitRepoErr.Error(), "failed to open repository")
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test ListBranches - routes to GoGitClient
func (s *GitClientTestSuite) TestListBranches_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedBranches := []domain.BranchInfo{
		{Name: "main", IsCurrent: true},
		{Name: "feature", IsCurrent: false},
	}

	s.mockGoGitClient.On("ListBranches", ctx, repoPath).Return(expectedBranches, nil)

	branches, err := s.compositeClient.ListBranches(ctx, repoPath)

	s.Require().NoError(err)
	s.Equal(expectedBranches, branches)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test BranchExists - routes to GoGitClient
func (s *GitClientTestSuite) TestBranchExists_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "main"

	s.mockGoGitClient.On("BranchExists", ctx, repoPath, branchName).Return(true, nil)

	exists, err := s.compositeClient.BranchExists(ctx, repoPath, branchName)

	s.Require().NoError(err)
	s.True(exists)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test GetRepositoryStatus - routes to GoGitClient
func (s *GitClientTestSuite) TestGetRepositoryStatus_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedStatus := domain.RepositoryStatus{
		IsClean: true,
		Branch:  "main",
	}

	s.mockGoGitClient.On("GetRepositoryStatus", ctx, repoPath).Return(expectedStatus, nil)

	status, err := s.compositeClient.GetRepositoryStatus(ctx, repoPath)

	s.Require().NoError(err)
	s.Equal(expectedStatus, status)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test ValidateRepository - routes to GoGitClient
func (s *GitClientTestSuite) TestValidateRepository_RoutesToGoGitClient() {
	repoPath := "/path/to/repo"

	s.mockGoGitClient.On("ValidateRepository", repoPath).Return(nil)

	err := s.compositeClient.ValidateRepository(repoPath)

	s.Require().NoError(err)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test GetRepositoryInfo - routes to GoGitClient
func (s *GitClientTestSuite) TestGetRepositoryInfo_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedInfo := &domain.GitRepository{
		Path:          repoPath,
		DefaultBranch: "main",
	}

	s.mockGoGitClient.On("GetRepositoryInfo", ctx, repoPath).Return(expectedInfo, nil)

	info, err := s.compositeClient.GetRepositoryInfo(ctx, repoPath)

	s.Require().NoError(err)
	s.Equal(expectedInfo, info)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test ListRemotes - routes to GoGitClient
func (s *GitClientTestSuite) TestListRemotes_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedRemotes := []domain.RemoteInfo{
		{Name: "origin", FetchURL: "git@github.com:user/repo.git"},
	}

	s.mockGoGitClient.On("ListRemotes", ctx, repoPath).Return(expectedRemotes, nil)

	remotes, err := s.compositeClient.ListRemotes(ctx, repoPath)

	s.Require().NoError(err)
	s.Equal(expectedRemotes, remotes)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test GetCommitInfo - routes to GoGitClient
func (s *GitClientTestSuite) TestGetCommitInfo_RoutesToGoGitClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	commitHash := "abc123"
	expectedInfo := &domain.CommitInfo{
		Hash:    commitHash,
		Message: "Test commit",
	}

	s.mockGoGitClient.On("GetCommitInfo", ctx, repoPath, commitHash).Return(expectedInfo, nil)

	info, err := s.compositeClient.GetCommitInfo(ctx, repoPath, commitHash)

	s.Require().NoError(err)
	s.Equal(expectedInfo, info)
	s.mockGoGitClient.AssertExpectations(s.T())
}

// Test CreateWorktree - routes to CLIClient
func (s *GitClientTestSuite) TestCreateWorktree_RoutesToCLIClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"
	sourceBranch := "main"
	worktreePath := "/path/to/worktree"

	s.mockCLIClient.On("CreateWorktree", ctx, repoPath, branchName, sourceBranch, worktreePath).Return(nil)

	err := s.compositeClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)

	s.Require().NoError(err)
	s.mockCLIClient.AssertExpectations(s.T())
	s.mockGoGitClient.AssertNotCalled(s.T(), "CreateWorktree")
}

func (s *GitClientTestSuite) TestCreateWorktree_ReturnsError() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"
	sourceBranch := "main"
	worktreePath := "/path/to/worktree"
	expectedErr := errors.New("failed to create worktree")

	s.mockCLIClient.On("CreateWorktree", ctx, repoPath, branchName, sourceBranch, worktreePath).Return(expectedErr)

	err := s.compositeClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath)

	s.Require().Error(err)
	var gitWorktreeErr *domain.GitWorktreeError
	s.ErrorAs(err, &gitWorktreeErr)
	s.Contains(gitWorktreeErr.Error(), "failed to create worktree")
}

// Test DeleteWorktree - routes to CLIClient
func (s *GitClientTestSuite) TestDeleteWorktree_RoutesToCLIClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	worktreePath := "/path/to/worktree"
	force := true

	s.mockCLIClient.On("DeleteWorktree", ctx, repoPath, worktreePath, force).Return(nil)

	err := s.compositeClient.DeleteWorktree(ctx, repoPath, worktreePath, force)

	s.Require().NoError(err)
	s.mockCLIClient.AssertExpectations(s.T())
}

// Test ListWorktrees - routes to CLIClient
func (s *GitClientTestSuite) TestListWorktrees_RoutesToCLIClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	expectedWorktrees := []domain.WorktreeInfo{
		{Path: "/path/to/worktree1", Branch: "feature1"},
		{Path: "/path/to/worktree2", Branch: "feature2"},
	}

	s.mockCLIClient.On("ListWorktrees", ctx, repoPath).Return(expectedWorktrees, nil)

	worktrees, err := s.compositeClient.ListWorktrees(ctx, repoPath)

	s.Require().NoError(err)
	s.Equal(expectedWorktrees, worktrees)
	s.mockCLIClient.AssertExpectations(s.T())
	s.mockGoGitClient.AssertNotCalled(s.T(), "ListWorktrees")
}

// Test PruneWorktrees - routes to CLIClient
func (s *GitClientTestSuite) TestPruneWorktrees_RoutesToCLIClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"

	s.mockCLIClient.On("PruneWorktrees", ctx, repoPath).Return(nil)

	err := s.compositeClient.PruneWorktrees(ctx, repoPath)

	s.Require().NoError(err)
	s.mockCLIClient.AssertExpectations(s.T())
}

// Test IsBranchMerged - routes to CLIClient
func (s *GitClientTestSuite) TestIsBranchMerged_RoutesToCLIClient() {
	ctx := context.Background()
	repoPath := "/path/to/repo"
	branchName := "feature"

	s.mockCLIClient.On("IsBranchMerged", ctx, repoPath, branchName).Return(true, nil)

	merged, err := s.compositeClient.IsBranchMerged(ctx, repoPath, branchName)

	s.Require().NoError(err)
	s.True(merged)
	s.mockCLIClient.AssertExpectations(s.T())
}
