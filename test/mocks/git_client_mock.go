// Package mocks contains test mocks for twiggit
package mocks

import (
	"context"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/stretchr/testify/mock"
)

// GitClientMock is a centralized mock for GitClient interface
type GitClientMock struct {
	mock.Mock
}

// IsGitRepository checks if a path is a git repository
func (m *GitClientMock) IsGitRepository(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

// IsMainRepository checks if a path is the main git repository
func (m *GitClientMock) IsMainRepository(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

// BranchExists checks if a branch exists in the repository
func (m *GitClientMock) BranchExists(ctx context.Context, repoPath, branch string) bool {
	args := m.Called(ctx, repoPath, branch)
	return args.Bool(0)
}

// CreateWorktree creates a new git worktree
func (m *GitClientMock) CreateWorktree(ctx context.Context, repoPath, branch, targetPath string) error {
	args := m.Called(ctx, repoPath, branch, targetPath)
	return args.Error(0)
}

// RemoveWorktree removes a git worktree
func (m *GitClientMock) RemoveWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	args := m.Called(ctx, repoPath, worktreePath, force)
	return args.Error(0)
}

// ListWorktrees lists all worktrees in a repository
func (m *GitClientMock) ListWorktrees(ctx context.Context, repoPath string) ([]*domain.WorktreeInfo, error) {
	args := m.Called(ctx, repoPath)
	return args.Get(0).([]*domain.WorktreeInfo), args.Error(1)
}

// GetWorktreeStatus gets the status of a worktree
func (m *GitClientMock) GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeInfo, error) {
	args := m.Called(ctx, worktreePath)
	return args.Get(0).(*domain.WorktreeInfo), args.Error(1)
}

// HasUncommittedChanges checks if a worktree has uncommitted changes
func (m *GitClientMock) HasUncommittedChanges(ctx context.Context, worktreePath string) bool {
	args := m.Called(ctx, worktreePath)
	return args.Bool(0)
}

// GetCurrentBranch gets the current branch of a repository
func (m *GitClientMock) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	args := m.Called(ctx, repoPath)
	return args.String(0), args.Error(1)
}

// GetRepositoryRoot gets the repository root path
func (m *GitClientMock) GetRepositoryRoot(ctx context.Context, path string) (string, error) {
	args := m.Called(ctx, path)
	return args.String(0), args.Error(1)
}

// GetAllBranches gets all branches in a repository
func (m *GitClientMock) GetAllBranches(ctx context.Context, repoPath string) ([]string, error) {
	args := m.Called(ctx, repoPath)
	return args.Get(0).([]string), args.Error(1)
}

// GetRemoteBranches gets all remote branches in a repository
func (m *GitClientMock) GetRemoteBranches(ctx context.Context, repoPath string) ([]string, error) {
	args := m.Called(ctx, repoPath)
	return args.Get(0).([]string), args.Error(1)
}

// SetupValidRepo sets up a valid repository mock
func (m *GitClientMock) SetupValidRepo(ctx context.Context, path string) *GitClientMock {
	m.On("IsGitRepository", ctx, path).Return(true, nil)
	m.On("IsMainRepository", ctx, path).Return(true, nil)
	return m
}

// SetupBranchExists sets up branch existence mock
func (m *GitClientMock) SetupBranchExists(ctx context.Context, repoPath, branch string, exists bool) *GitClientMock {
	m.On("BranchExists", ctx, repoPath, branch).Return(exists)
	return m
}

// SetupWorktreeCreation sets up worktree creation mock
func (m *GitClientMock) SetupWorktreeCreation(ctx context.Context, repoPath, branch, targetPath string) *GitClientMock {
	m.On("CreateWorktree", ctx, repoPath, branch, targetPath).Return(nil)
	return m
}
