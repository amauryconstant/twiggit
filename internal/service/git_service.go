package service

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"twiggit/internal/domain"
)

// GoGitClient defines go-git operations (deterministic routing - no CLI fallback)
// All methods SHALL be idempotent and thread-safe
type GoGitClient interface {
	// OpenRepository opens git repository (pure function, idempotent)
	OpenRepository(path string) (*git.Repository, error)

	// ListBranches lists all branches in repository (idempotent)
	ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error)

	// BranchExists checks if branch exists (idempotent)
	BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)

	// GetRepositoryStatus returns repository status (idempotent)
	GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error)

	// ValidateRepository checks if path contains valid git repository (pure function)
	ValidateRepository(path string) error

	// GetRepositoryInfo returns comprehensive repository information
	GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error)

	// ListRemotes lists all remotes in repository
	ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error)

	// GetCommitInfo returns information about a specific commit
	GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error)
}

// CLIClient defines CLI operations for worktree management ONLY
// All methods SHALL be idempotent and thread-safe
type CLIClient interface {
	// CreateWorktree creates new worktree using git CLI (idempotent)
	CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error

	// DeleteWorktree removes worktree using git CLI (idempotent, no-op if already deleted)
	DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error

	// ListWorktrees lists all worktrees using git CLI (idempotent)
	ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error)

	// PruneWorktrees removes stale worktree references
	PruneWorktrees(ctx context.Context, repoPath string) error

	// IsBranchMerged checks if a branch is merged into the current branch
	IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error)
}

// GitService provides unified git operations with deterministic routing
// No fallback logic - operations use predetermined implementation
type GitService interface {
	GoGitClient
	CLIClient
}

// GitServiceConfig represents configuration for git operations
type GitServiceConfig struct {
	// CLITimeout is the timeout for CLI operations
	CLITimeout int `toml:"cli_timeout"`

	// CacheEnabled enables caching of git operations
	CacheEnabled bool `toml:"cache_enabled"`

	// OperationTimeout is the default timeout for git operations
	OperationTimeout int `toml:"operation_timeout"`
}

// DefaultGitServiceConfig returns default configuration for git operations
func DefaultGitServiceConfig() *GitServiceConfig {
	return &GitServiceConfig{
		CLITimeout:       30, // 30 seconds
		CacheEnabled:     true,
		OperationTimeout: 30, // 30 seconds
	}
}

// gitService implements GitService with deterministic routing
type gitService struct {
	goGitClient GoGitClient
	cliClient   CLIClient
	config      *GitServiceConfig
}

// NewGitService creates a new GitService with deterministic routing
func NewGitService(goGitClient GoGitClient, cliClient CLIClient, config *GitServiceConfig) GitService {
	if config == nil {
		config = DefaultGitServiceConfig()
	}

	return &gitService{
		goGitClient: goGitClient,
		cliClient:   cliClient,
		config:      config,
	}
}

// GoGit operations - delegate to goGitClient

func (gs *gitService) OpenRepository(path string) (*git.Repository, error) {
	repo, err := gs.goGitClient.OpenRepository(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return repo, nil
}

func (gs *gitService) ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
	branches, err := gs.goGitClient.ListBranches(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	return branches, nil
}

func (gs *gitService) BranchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	exists, err := gs.goGitClient.BranchExists(ctx, repoPath, branchName)
	if err != nil {
		return false, fmt.Errorf("failed to check if branch exists: %w", err)
	}
	return exists, nil
}

func (gs *gitService) GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error) {
	status, err := gs.goGitClient.GetRepositoryStatus(ctx, repoPath)
	if err != nil {
		return domain.RepositoryStatus{}, fmt.Errorf("failed to get repository status: %w", err)
	}
	return status, nil
}

func (gs *gitService) ValidateRepository(path string) error {
	if err := gs.goGitClient.ValidateRepository(path); err != nil {
		return fmt.Errorf("failed to validate repository: %w", err)
	}
	return nil
}

func (gs *gitService) GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
	info, err := gs.goGitClient.GetRepositoryInfo(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}
	return info, nil
}

func (gs *gitService) ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error) {
	remotes, err := gs.goGitClient.ListRemotes(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}
	return remotes, nil
}

func (gs *gitService) GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error) {
	info, err := gs.goGitClient.GetCommitInfo(ctx, repoPath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	}
	return info, nil
}

// CLI operations - delegate to cliClient

func (gs *gitService) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	if err := gs.cliClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	return nil
}

func (gs *gitService) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	if err := gs.cliClient.DeleteWorktree(ctx, repoPath, worktreePath, force); err != nil {
		return fmt.Errorf("failed to delete worktree: %w", err)
	}
	return nil
}

func (gs *gitService) ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
	worktrees, err := gs.cliClient.ListWorktrees(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	return worktrees, nil
}

func (gs *gitService) PruneWorktrees(ctx context.Context, repoPath string) error {
	if err := gs.cliClient.PruneWorktrees(ctx, repoPath); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}
	return nil
}

func (gs *gitService) IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	merged, err := gs.cliClient.IsBranchMerged(ctx, repoPath, branchName)
	if err != nil {
		return false, fmt.Errorf("failed to check if branch is merged: %w", err)
	}
	return merged, nil
}
