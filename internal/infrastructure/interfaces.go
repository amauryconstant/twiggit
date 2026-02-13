package infrastructure

import (
	"context"

	"github.com/go-git/go-git/v5"
	"twiggit/internal/domain"
)

// GitClient provides unified git operations with deterministic routing
type GitClient interface {
	GoGitClient
	CLIClient
}

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

	// DeleteBranch deletes a branch using git CLI (handles worktree-referenced branches)
	DeleteBranch(ctx context.Context, repoPath, branchName string) error
}

// ShellInfrastructure defines low-level shell infrastructure operations
type ShellInfrastructure interface {
	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(shellType domain.ShellType) (string, error)

	// DetectConfigFile detects the appropriate config file for the shell type
	DetectConfigFile(shellType domain.ShellType) (string, error)

	// InstallWrapper installs the wrapper to the shell config file
	InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error

	// ValidateInstallation validates whether the wrapper is installed
	ValidateInstallation(shellType domain.ShellType, configFile string) error
}
