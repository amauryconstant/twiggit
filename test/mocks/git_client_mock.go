// Package mocks contains test mocks for twiggit
package mocks

import (
	"github.com/amaury/twiggit/internal/domain"
	"github.com/stretchr/testify/mock"
)

// GitClientMock is a centralized mock for GitClient interface
type GitClientMock struct {
	mock.Mock
}

// IsGitRepository checks if a path is a git repository
func (m *GitClientMock) IsGitRepository(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

// IsMainRepository checks if a path is the main git repository
func (m *GitClientMock) IsMainRepository(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

// BranchExists checks if a branch exists in the repository
func (m *GitClientMock) BranchExists(repoPath, branch string) bool {
	args := m.Called(repoPath, branch)
	return args.Bool(0)
}

// CreateWorktree creates a new git worktree
func (m *GitClientMock) CreateWorktree(repoPath, branch, targetPath string) error {
	args := m.Called(repoPath, branch, targetPath)
	return args.Error(0)
}

// RemoveWorktree removes a git worktree
func (m *GitClientMock) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	args := m.Called(repoPath, worktreePath, force)
	return args.Error(0)
}

// ListWorktrees lists all worktrees in a repository
func (m *GitClientMock) ListWorktrees(repoPath string) ([]*domain.WorktreeInfo, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]*domain.WorktreeInfo), args.Error(1)
}

// GetWorktreeStatus gets the status of a worktree
func (m *GitClientMock) GetWorktreeStatus(worktreePath string) (*domain.WorktreeInfo, error) {
	args := m.Called(worktreePath)
	return args.Get(0).(*domain.WorktreeInfo), args.Error(1)
}

// HasUncommittedChanges checks if a worktree has uncommitted changes
func (m *GitClientMock) HasUncommittedChanges(worktreePath string) bool {
	args := m.Called(worktreePath)
	return args.Bool(0)
}

// GetCurrentBranch gets the current branch of a repository
func (m *GitClientMock) GetCurrentBranch(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

// GetRepositoryRoot gets the repository root path
func (m *GitClientMock) GetRepositoryRoot(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

// GetAllBranches gets all branches in a repository
func (m *GitClientMock) GetAllBranches(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

// GetRemoteBranches gets all remote branches in a repository
func (m *GitClientMock) GetRemoteBranches(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

// SetupValidRepo sets up a valid repository mock
func (m *GitClientMock) SetupValidRepo(path string) *GitClientMock {
	m.On("IsGitRepository", path).Return(true, nil)
	m.On("IsMainRepository", path).Return(true, nil)
	return m
}

// SetupBranchExists sets up branch existence mock
func (m *GitClientMock) SetupBranchExists(repoPath, branch string, exists bool) *GitClientMock {
	m.On("BranchExists", repoPath, branch).Return(exists)
	return m
}

// SetupWorktreeCreation sets up worktree creation mock
func (m *GitClientMock) SetupWorktreeCreation(repoPath, branch, targetPath string) *GitClientMock {
	m.On("CreateWorktree", repoPath, branch, targetPath).Return(nil)
	return m
}
