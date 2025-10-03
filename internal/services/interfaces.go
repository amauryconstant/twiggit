package services

import (
	"context"

	"twiggit/internal/domain"
)

// WorktreeService provides high-level worktree management operations
type WorktreeService interface {
	// CreateWorktree creates a new worktree for the specified project and branch
	CreateWorktree(ctx context.Context, req *domain.CreateWorktreeRequest) (*domain.WorktreeInfo, error)

	// DeleteWorktree deletes an existing worktree
	DeleteWorktree(ctx context.Context, req *domain.DeleteWorktreeRequest) error

	// ListWorktrees lists all worktrees for a project
	ListWorktrees(ctx context.Context, req *domain.ListWorktreesRequest) ([]*domain.WorktreeInfo, error)

	// GetWorktreeStatus retrieves the status of a specific worktree
	GetWorktreeStatus(ctx context.Context, worktreePath string) (*domain.WorktreeStatus, error)

	// ValidateWorktree validates that a worktree is properly configured
	ValidateWorktree(ctx context.Context, worktreePath string) error
}

// ProjectService provides project discovery and management operations
type ProjectService interface {
	// DiscoverProject discovers a project by name or from context
	DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error)

	// ValidateProject validates that a project is properly configured
	ValidateProject(ctx context.Context, projectPath string) error

	// ListProjects lists all available projects
	ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error)

	// GetProjectInfo retrieves detailed information about a project
	GetProjectInfo(ctx context.Context, projectPath string) (*domain.ProjectInfo, error)
}

// NavigationService provides path resolution and navigation operations
type NavigationService interface {
	// ResolvePath resolves a target identifier to a concrete path
	ResolvePath(ctx context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error)

	// ValidatePath validates that a path is accessible and valid
	ValidatePath(ctx context.Context, path string) error

	// GetNavigationSuggestions provides completion suggestions for navigation
	GetNavigationSuggestions(ctx context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error)
}

// ShellService provides shell integration and wrapper management operations
type ShellService interface {
	// SetupShell sets up shell integration for the specified shell type
	SetupShell(ctx context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error)

	// ValidateInstallation validates whether shell integration is installed
	ValidateInstallation(ctx context.Context, req *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error)

	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(ctx context.Context, req *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error)
}
