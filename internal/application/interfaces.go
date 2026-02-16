package application

import (
	"context"

	"twiggit/internal/domain"
)

// ContextService provides context detection and resolution operations
type ContextService interface {
	// GetCurrentContext returns the current working context
	GetCurrentContext() (*domain.Context, error)

	// DetectContextFromPath detects context from a file system path
	DetectContextFromPath(path string) (*domain.Context, error)

	// ResolveIdentifier resolves an identifier to a resolution result
	ResolveIdentifier(identifier string) (*domain.ResolutionResult, error)

	// ResolveIdentifierFromContext resolves an identifier within a specific context
	ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error)

	// GetCompletionSuggestions provides completion suggestions for partial identifiers
	GetCompletionSuggestions(partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error)

	// GetCompletionSuggestionsFromContext provides completion suggestions for partial identifiers within a specific context
	GetCompletionSuggestionsFromContext(ctx *domain.Context, partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error)
}

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

	// PruneMergedWorktrees deletes merged worktrees with optional branch deletion
	PruneMergedWorktrees(ctx context.Context, req *domain.PruneWorktreesRequest) (*domain.PruneWorktreesResult, error)

	// BranchExists checks if a branch exists in the project
	BranchExists(ctx context.Context, projectPath, branchName string) (bool, error)

	// IsBranchMerged checks if a branch has been merged into its base branch
	IsBranchMerged(ctx context.Context, worktreePath, branchName string) (bool, error)

	// GetWorktreeByPath retrieves worktree info by its path
	GetWorktreeByPath(ctx context.Context, projectPath, worktreePath string) (*domain.WorktreeInfo, error)
}

// ProjectService provides project discovery and management operations
type ProjectService interface {
	// DiscoverProject discovers a project by name or from context
	DiscoverProject(ctx context.Context, projectName string, context *domain.Context) (*domain.ProjectInfo, error)

	// ValidateProject validates that a project is properly configured
	ValidateProject(ctx context.Context, projectPath string) error

	// ListProjects lists all available projects with full info (including worktrees)
	ListProjects(ctx context.Context) ([]*domain.ProjectInfo, error)

	// ListProjectSummaries lists project names/paths without loading worktrees
	ListProjectSummaries(ctx context.Context) ([]*domain.ProjectSummary, error)

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
