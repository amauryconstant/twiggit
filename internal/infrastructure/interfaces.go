// Package infrastructure provides concrete implementations of external dependencies
// including Git clients, configuration management, and validation services.
package infrastructure

import (
	"context"
	"github.com/amaury/twiggit/internal/domain"
)

// GitClient interface defines Git operations for worktree management
type GitClient interface {
	// Repository operations
	IsGitRepository(ctx context.Context, path string) (bool, error)
	IsBareRepository(ctx context.Context, path string) (bool, error)
	IsMainRepository(ctx context.Context, path string) (bool, error)
	GetRepositoryRoot(ctx context.Context, path string) (string, error)

	// Worktree operations
	ListWorktrees(ctx context.Context, repoPath string) ([]*domain.WorktreeInfo, error)
	CreateWorktree(ctx context.Context, repoPath, branch, targetPath string) error
	RemoveWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error
	GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeInfo, error)

	// Branch operations
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)
	GetAllBranches(ctx context.Context, repoPath string) ([]string, error)
	GetRemoteBranches(ctx context.Context, repoPath string) ([]string, error)
	BranchExists(ctx context.Context, repoPath, branch string) bool

	// Status operations
	HasUncommittedChanges(ctx context.Context, repoPath string) bool
}

// PathValidator defines infrastructure-specific path validation operations
type PathValidator interface {
	// IsValidGitRepoPath validates if a path is suitable for a git repository
	IsValidGitRepoPath(path string) bool

	// IsValidWorkspacePath validates if a path is suitable for a workspace
	IsValidWorkspacePath(path string) bool
}

// MiseIntegration defines the interface for mise development environment integration
type MiseIntegration interface {
	// SetupWorktree sets up mise configuration for a new worktree
	SetupWorktree(sourceRepoPath, worktreePath string) error

	// IsAvailable checks if mise is available on system
	IsAvailable() bool

	// DetectConfigFiles finds mise configuration files in the given repository path
	DetectConfigFiles(repoPath string) []string

	// CopyConfigFiles copies mise configuration files from source to target
	CopyConfigFiles(sourceDir, targetDir string, configFiles []string) error

	// TrustDirectory runs 'mise trust' on the specified directory if mise is available
	TrustDirectory(dirPath string) error

	// Disable disables mise integration
	Disable()

	// Enable enables mise integration if mise is available
	Enable()

	// IsEnabled returns whether mise integration is currently enabled
	IsEnabled() bool

	// SetExecPath allows customizing the mise executable path (useful for testing)
	SetExecPath(path string)
}
