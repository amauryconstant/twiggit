package domain

import "context"

// ContextService defines domain service operations for context detection and resolution
type ContextService interface {
	// GetCurrentContext returns the current working context
	GetCurrentContext() (*Context, error)

	// DetectContextFromPath detects context from a file system path
	DetectContextFromPath(path string) (*Context, error)

	// ResolveIdentifier resolves an identifier to a resolution result
	ResolveIdentifier(identifier string) (*ResolutionResult, error)

	// ResolveIdentifierFromContext resolves an identifier within a specific context
	ResolveIdentifierFromContext(ctx *Context, identifier string) (*ResolutionResult, error)

	// GetCompletionSuggestions provides completion suggestions for partial identifiers
	GetCompletionSuggestions(partial string) ([]*ResolutionSuggestion, error)
}

// GitRepositoryInterface defines repository operations for git entities
type GitRepositoryInterface interface {
	// ValidateRepository checks if path contains valid git repository
	ValidateRepository(path string) error

	// GetRepositoryInfo returns comprehensive repository information
	GetRepositoryInfo(ctx context.Context, repoPath string) (*GitRepository, error)

	// ListWorktrees lists all worktrees in the repository
	ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
}

// ProjectRepository defines repository operations for project entities
type ProjectRepository interface {
	// FindByName finds a project by its name
	FindByName(name string) (*Project, error)

	// FindByPath finds a project by its path
	FindByPath(path string) (*Project, error)

	// ListAll returns all available projects
	ListAll() ([]*Project, error)

	// Save saves a project to the repository
	Save(project *Project) error
}
