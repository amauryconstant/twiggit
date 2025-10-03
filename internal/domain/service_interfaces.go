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

// ShellServiceInterface defines the shell service operations we need
type ShellServiceInterface interface {
	// SetupShell sets up shell integration for the specified shell type
	SetupShell(ctx context.Context, req *SetupShellRequest) (*SetupShellResult, error)

	// ValidateInstallation validates whether shell integration is installed
	ValidateInstallation(ctx context.Context, req *ValidateInstallationRequest) (*ValidateInstallationResult, error)

	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(ctx context.Context, req *GenerateWrapperRequest) (*GenerateWrapperResult, error)
}
