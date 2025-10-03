package domain

import "context"

// ContextServiceInterface defines the context service operations we need
type ContextServiceInterface interface {
	GetCurrentContext() (*Context, error)
	DetectContextFromPath(path string) (*Context, error)
	ResolveIdentifier(identifier string) (*ResolutionResult, error)
	ResolveIdentifierFromContext(ctx *Context, identifier string) (*ResolutionResult, error)
	GetCompletionSuggestions(partial string) ([]*ResolutionSuggestion, error)
}

// GitServiceInterface defines the git service operations we need
type GitServiceInterface interface {
	ValidateRepository(path string) error
	GetRepositoryInfo(ctx context.Context, repoPath string) (*GitRepository, error)
	ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
}
