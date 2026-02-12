package infrastructure

import (
	"context"

	"github.com/go-git/go-git/v5"
	"twiggit/internal/domain"
)

// CompositeGitClient implements GitClient by combining GoGit and CLI functionality
type CompositeGitClient struct {
	goGitClient GoGitClient
	cliClient   CLIClient
}

// NewCompositeGitClient creates a new composite GitClient
func NewCompositeGitClient(goGitClient GoGitClient, cliClient CLIClient) GitClient {
	return &CompositeGitClient{
		goGitClient: goGitClient,
		cliClient:   cliClient,
	}
}

// OpenRepository opens a git repository using the GoGit client
func (c *CompositeGitClient) OpenRepository(path string) (*git.Repository, error) {
	repo, err := c.goGitClient.OpenRepository(path)
	if err != nil {
		return nil, domain.NewGitRepositoryError(path, "failed to open repository", err)
	}
	return repo, nil
}

// ListBranches lists all branches in the repository using the GoGit client
func (c *CompositeGitClient) ListBranches(ctx context.Context, repoPath string) ([]domain.BranchInfo, error) {
	branches, err := c.goGitClient.ListBranches(ctx, repoPath)
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to list branches", err)
	}
	return branches, nil
}

// BranchExists checks if a branch exists using the GoGit client
func (c *CompositeGitClient) BranchExists(ctx context.Context, repoPath, branchName string) (bool, error) {
	exists, err := c.goGitClient.BranchExists(ctx, repoPath, branchName)
	if err != nil {
		return false, domain.NewGitRepositoryError(repoPath, "failed to check branch existence", err)
	}
	return exists, nil
}

// GetRepositoryStatus gets the repository status using the GoGit client
func (c *CompositeGitClient) GetRepositoryStatus(ctx context.Context, repoPath string) (domain.RepositoryStatus, error) {
	status, err := c.goGitClient.GetRepositoryStatus(ctx, repoPath)
	if err != nil {
		return domain.RepositoryStatus{}, domain.NewGitRepositoryError(repoPath, "failed to get repository status", err)
	}
	return status, nil
}

// ValidateRepository validates a repository using the GoGit client
func (c *CompositeGitClient) ValidateRepository(path string) error {
	if err := c.goGitClient.ValidateRepository(path); err != nil {
		return domain.NewGitRepositoryError(path, "failed to validate repository", err)
	}
	return nil
}

// GetRepositoryInfo gets repository information using the GoGit client
func (c *CompositeGitClient) GetRepositoryInfo(ctx context.Context, repoPath string) (*domain.GitRepository, error) {
	info, err := c.goGitClient.GetRepositoryInfo(ctx, repoPath)
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to get repository info", err)
	}
	return info, nil
}

// ListRemotes lists all remotes using the GoGit client
func (c *CompositeGitClient) ListRemotes(ctx context.Context, repoPath string) ([]domain.RemoteInfo, error) {
	remotes, err := c.goGitClient.ListRemotes(ctx, repoPath)
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to list remotes", err)
	}
	return remotes, nil
}

// GetCommitInfo gets commit information using the GoGit client
func (c *CompositeGitClient) GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*domain.CommitInfo, error) {
	info, err := c.goGitClient.GetCommitInfo(ctx, repoPath, commitHash)
	if err != nil {
		return nil, domain.NewGitRepositoryError(repoPath, "failed to get commit info", err)
	}
	return info, nil
}

// CreateWorktree creates a worktree using the CLI client
func (c *CompositeGitClient) CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error {
	if err := c.cliClient.CreateWorktree(ctx, repoPath, branchName, sourceBranch, worktreePath); err != nil {
		return domain.NewGitWorktreeError(worktreePath, branchName, "failed to create worktree", err)
	}
	return nil
}

// DeleteWorktree deletes a worktree using the CLI client
func (c *CompositeGitClient) DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error {
	if err := c.cliClient.DeleteWorktree(ctx, repoPath, worktreePath, force); err != nil {
		return domain.NewGitWorktreeError(worktreePath, "", "failed to delete worktree", err)
	}
	return nil
}

// ListWorktrees lists worktrees using the CLI client
func (c *CompositeGitClient) ListWorktrees(ctx context.Context, repoPath string) ([]domain.WorktreeInfo, error) {
	worktrees, err := c.cliClient.ListWorktrees(ctx, repoPath)
	if err != nil {
		return nil, domain.NewGitWorktreeError(repoPath, "", "failed to list worktrees", err)
	}
	return worktrees, nil
}

// PruneWorktrees prunes stale worktree references using the CLI client
func (c *CompositeGitClient) PruneWorktrees(ctx context.Context, repoPath string) error {
	if err := c.cliClient.PruneWorktrees(ctx, repoPath); err != nil {
		return domain.NewGitWorktreeError(repoPath, "", "failed to prune worktrees", err)
	}
	return nil
}

// IsBranchMerged checks if a branch is merged into the current branch using the CLI client
func (c *CompositeGitClient) IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error) {
	merged, err := c.cliClient.IsBranchMerged(ctx, repoPath, branchName)
	if err != nil {
		return false, domain.NewGitWorktreeError(repoPath, branchName, "failed to check if branch is merged", err)
	}
	return merged, nil
}
