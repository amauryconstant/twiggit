package mocks

import (
	"context"

	"twiggit/internal/domain"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/mock"
)

// MockGoGitClient implements infrastructure.GoGitClient for testing
type MockGoGitClient struct {
	mock.Mock
}

// NewMockGoGitClient creates a new mock GoGitClient for testing
func NewMockGoGitClient() *MockGoGitClient {
	return &MockGoGitClient{}
}

// OpenRepository mocks opening a git repository
func (m *MockGoGitClient) OpenRepository(path string) (*git.Repository, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*git.Repository), args.Error(1)
}

// ListBranches mocks listing branches in a repository
func (m *MockGoGitClient) ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
	args := m.Called(ctx, repoPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BranchInfo), args.Error(1)
}

// BranchExists mocks checking if a branch exists
func (m *MockGoGitClient) BranchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	args := m.Called(ctx, repoPath, branchName)
	return args.Bool(0), args.Error(1)
}

// GetRepositoryStatus mocks getting repository status
func (m *MockGoGitClient) GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error) {
	args := m.Called(ctx, repoPath)
	return args.Get(0).(domain.RepositoryStatus), args.Error(1)
}

// ValidateRepository mocks validating a repository
func (m *MockGoGitClient) ValidateRepository(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// GetRepositoryInfo mocks getting repository information
func (m *MockGoGitClient) GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
	args := m.Called(ctx, repoPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitRepository), args.Error(1)
}

// ListRemotes mocks listing repository remotes
func (m *MockGoGitClient) ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error) {
	args := m.Called(ctx, repoPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RemoteInfo), args.Error(1)
}

// GetCommitInfo mocks getting commit information
func (m *MockGoGitClient) GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error) {
	args := m.Called(ctx, repoPath, commitHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CommitInfo), args.Error(1)
}

// MockCLIClient implements infrastructure.CLIClient for testing
type MockCLIClient struct {
	mock.Mock
}

// NewMockCLIClient creates a new mock CLIClient for testing
func NewMockCLIClient() *MockCLIClient {
	return &MockCLIClient{}
}

// CreateWorktree mocks creating a new worktree
func (m *MockCLIClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	args := m.Called(ctx, repoPath, branchName, sourceBranch, worktreePath)
	return args.Error(0)
}

// DeleteWorktree mocks deleting a worktree
func (m *MockCLIClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	args := m.Called(ctx, repoPath, worktreePath, force)
	return args.Error(0)
}

// ListWorktrees mocks listing worktrees in a repository
func (m *MockCLIClient) ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
	args := m.Called(ctx, repoPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.WorktreeInfo), args.Error(1)
}

// PruneWorktrees mocks pruning worktrees in a repository
func (m *MockCLIClient) PruneWorktrees(ctx context.Context, repoPath string) error {
	args := m.Called(ctx, repoPath)
	return args.Error(0)
}

// IsBranchMerged mocks checking if a branch is merged
func (m *MockCLIClient) IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	args := m.Called(ctx, repoPath, branchName)
	return args.Bool(0), args.Error(1)
}

// MockGitService implements infrastructure.GitClient for testing
type MockGitService struct {
	*MockGoGitClient
	*MockCLIClient
}

// NewMockGitService creates a new mock GitService for testing
func NewMockGitService() *MockGitService {
	return &MockGitService{
		MockGoGitClient: NewMockGoGitClient(),
		MockCLIClient:   NewMockCLIClient(),
	}
}
