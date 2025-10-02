package mocks

import (
	"context"

	"twiggit/internal/domain"

	"github.com/go-git/go-git/v5"
)

// MockGoGitClient implements service.GoGitClient for testing
type MockGoGitClient struct {
	// Mock functions
	OpenRepositoryFunc      func(path string) (*git.Repository, error)
	ListBranchesFunc        func(ctx context.Context, repoPath string) ([]domain.BranchInfo, error)
	BranchExistsFunc        func(ctx context.Context, repoPath, branchName string) (bool, error)
	GetRepositoryStatusFunc func(ctx context.Context, repoPath string) (domain.RepositoryStatus, error)
	ValidateRepositoryFunc  func(path string) error
	GetRepositoryInfoFunc   func(ctx context.Context, repoPath string) (*domain.GitRepository, error)
	ListRemotesFunc         func(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error)
	GetCommitInfoFunc       func(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error)
}

// NewMockGoGitClient creates a new mock GoGitClient for testing
func NewMockGoGitClient() *MockGoGitClient {
	return &MockGoGitClient{}
}

// OpenRepository mocks opening a git repository
func (m *MockGoGitClient) OpenRepository(path string) (*git.Repository, error) {
	if m.OpenRepositoryFunc != nil {
		return m.OpenRepositoryFunc(path)
	}
	return nil, nil
}

// ListBranches mocks listing branches in a repository
func (m *MockGoGitClient) ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
	if m.ListBranchesFunc != nil {
		return m.ListBranchesFunc(ctx, repoPath)
	}
	return []domain.BranchInfo{}, nil
}

// BranchExists mocks checking if a branch exists
func (m *MockGoGitClient) BranchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	if m.BranchExistsFunc != nil {
		return m.BranchExistsFunc(ctx, repoPath, branchName)
	}
	return false, nil
}

// GetRepositoryStatus mocks getting repository status
func (m *MockGoGitClient) GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error) {
	if m.GetRepositoryStatusFunc != nil {
		return m.GetRepositoryStatusFunc(ctx, repoPath)
	}
	return domain.RepositoryStatus{}, nil
}

// ValidateRepository mocks validating a repository
func (m *MockGoGitClient) ValidateRepository(path string) error {
	if m.ValidateRepositoryFunc != nil {
		return m.ValidateRepositoryFunc(path)
	}
	return nil
}

// GetRepositoryInfo mocks getting repository information
func (m *MockGoGitClient) GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
	if m.GetRepositoryInfoFunc != nil {
		return m.GetRepositoryInfoFunc(ctx, repoPath)
	}
	return &domain.GitRepository{}, nil
}

// ListRemotes mocks listing repository remotes
func (m *MockGoGitClient) ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error) {
	if m.ListRemotesFunc != nil {
		return m.ListRemotesFunc(ctx, repoPath)
	}
	return []domain.RemoteInfo{}, nil
}

// GetCommitInfo mocks getting commit information
func (m *MockGoGitClient) GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error) {
	if m.GetCommitInfoFunc != nil {
		return m.GetCommitInfoFunc(ctx, repoPath, commitHash)
	}
	return &domain.CommitInfo{}, nil
}

// MockCLIClient implements service.CLIClient for testing
type MockCLIClient struct {
	// Mock functions
	CreateWorktreeFunc func(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error
	DeleteWorktreeFunc func(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error
	ListWorktreesFunc  func(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error)
	PruneWorktreesFunc func(ctx context.Context, repoPath string) error
}

// NewMockCLIClient creates a new mock CLIClient for testing
func NewMockCLIClient() *MockCLIClient {
	return &MockCLIClient{}
}

// CreateWorktree mocks creating a new worktree
func (m *MockCLIClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(ctx, repoPath, branchName, sourceBranch, worktreePath)
	}
	return nil
}

// DeleteWorktree mocks deleting a worktree
func (m *MockCLIClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, keepBranch bool) error {
	if m.DeleteWorktreeFunc != nil {
		return m.DeleteWorktreeFunc(ctx, repoPath, worktreePath, keepBranch)
	}
	return nil
}

// ListWorktrees mocks listing worktrees in a repository
func (m *MockCLIClient) ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
	if m.ListWorktreesFunc != nil {
		return m.ListWorktreesFunc(ctx, repoPath)
	}
	return []domain.WorktreeInfo{}, nil
}

// PruneWorktrees mocks pruning worktrees in a repository
func (m *MockCLIClient) PruneWorktrees(ctx context.Context, repoPath string) error {
	if m.PruneWorktreesFunc != nil {
		return m.PruneWorktreesFunc(ctx, repoPath)
	}
	return nil
}

// MockGitService implements service.GitService for testing
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
