package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type GoGitClientTestSuite struct {
	suite.Suite
	client  GoGitClient
	tempDir string
}

func (s *GoGitClientTestSuite) SetupTest() {
	s.client = NewGoGitClient()
	s.tempDir = s.T().TempDir()
}

func TestGoGitClient(t *testing.T) {
	suite.Run(t, new(GoGitClientTestSuite))
}

func (s *GoGitClientTestSuite) TestOpenRepository() {
	repo, err := s.client.OpenRepository("/non/existent/path")
	s.Require().Error(err)
	s.Nil(repo)

	repo, err = s.client.OpenRepository(s.tempDir)
	s.Require().Error(err)
	s.Nil(repo)
}

func (s *GoGitClientTestSuite) TestValidateRepository() {
	err := s.client.ValidateRepository("/non/existent/path")
	s.Require().Error(err)

	err = s.client.ValidateRepository(s.tempDir)
	s.Require().Error(err)

	repoPath := s.setupTestRepo()
	err = s.client.ValidateRepository(repoPath)
	s.Require().NoError(err)
}

func (s *GoGitClientTestSuite) TestListBranches() {
	branches, err := s.client.ListBranches(context.Background(), "/non/existent/path")
	s.Require().Error(err)
	s.Nil(branches)

	repoPath := s.setupTestRepo()
	branches, err = s.client.ListBranches(context.Background(), repoPath)
	s.Require().NoError(err)
	s.NotEmpty(branches)

	mainBranch := s.findBranch(branches, "main")
	s.NotNil(mainBranch)
	s.Equal("main", mainBranch.Name)
}

func (s *GoGitClientTestSuite) TestBranchExists() {
	exists, err := s.client.BranchExists(context.Background(), "/non/existent/path", "main")
	s.Require().Error(err)
	s.False(exists)

	repoPath := s.setupTestRepo()

	exists, err = s.client.BranchExists(context.Background(), repoPath, "main")
	s.Require().NoError(err)
	s.True(exists)

	exists, err = s.client.BranchExists(context.Background(), repoPath, "non-existent")
	s.Require().NoError(err)
	s.False(exists)
}

func (s *GoGitClientTestSuite) TestGetRepositoryStatus() {
	status, err := s.client.GetRepositoryStatus(context.Background(), "/non/existent/path")
	s.Require().Error(err)
	s.Equal(domain.RepositoryStatus{}, status)

	repoPath := s.setupTestRepo()
	status, err = s.client.GetRepositoryStatus(context.Background(), repoPath)
	s.Require().NoError(err)
	s.True(status.IsClean)
	s.Equal("main", status.Branch)
}

func (s *GoGitClientTestSuite) TestGetRepositoryInfo() {
	info, err := s.client.GetRepositoryInfo(context.Background(), "/non/existent/path")
	s.Require().Error(err)
	s.Nil(info)

	repoPath := s.setupTestRepo()
	info, err = s.client.GetRepositoryInfo(context.Background(), repoPath)
	s.Require().NoError(err)
	s.NotNil(info)
	s.Equal(repoPath, info.Path)
	s.False(info.IsBare)
	s.NotEmpty(info.Branches)
}

func (s *GoGitClientTestSuite) TestListRemotes() {
	remotes, err := s.client.ListRemotes(context.Background(), "/non/existent/path")
	s.Require().Error(err)
	s.Nil(remotes)

	repoPath := s.setupTestRepo()
	remotes, err = s.client.ListRemotes(context.Background(), repoPath)
	s.Require().NoError(err)
	s.Empty(remotes)
}

func (s *GoGitClientTestSuite) TestGetCommitInfo() {
	commit, err := s.client.GetCommitInfo(context.Background(), "/non/existent/path", "HEAD")
	s.Require().Error(err)
	s.Nil(commit)

	repoPath := s.setupTestRepo()
	commit, err = s.client.GetCommitInfo(context.Background(), repoPath, "HEAD")
	s.Require().Error(err)
	s.Nil(commit)
}

func (s *GoGitClientTestSuite) findBranch(branches []domain.BranchInfo, name string) *domain.BranchInfo {
	for _, branch := range branches {
		if branch.Name == name {
			return &branch
		}
	}
	return nil
}

func (s *GoGitClientTestSuite) setupTestRepo() string {
	s.T().Helper()

	repoPath := filepath.Join(s.tempDir, "test-repo")

	err := os.MkdirAll(repoPath, 0755)
	s.Require().NoError(err)

	_, err = s.client.OpenRepository(repoPath)
	if err == nil {
		return repoPath
	}

	gitDir := filepath.Join(repoPath, ".git")
	err = os.MkdirAll(gitDir, 0755)
	s.Require().NoError(err)

	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644)
	s.Require().NoError(err)

	refsDir := filepath.Join(gitDir, "refs", "heads")
	err = os.MkdirAll(refsDir, 0755)
	s.Require().NoError(err)

	err = os.WriteFile(filepath.Join(refsDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0644)
	s.Require().NoError(err)

	return repoPath
}
