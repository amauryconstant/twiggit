package application

import (
	"context"

	"github.com/go-git/go-git/v5"
	"twiggit/internal/core"
)

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from defaults and config file
	Load() (*core.Config, error)

	// GetConfig returns the loaded configuration (immutable after Load)
	GetConfig() *core.Config
}

// ContextDetector detects the current git context
type ContextDetector interface {
	// DetectContext detects the context from the given directory
	DetectContext(dir string) (*core.Context, error)
}

// ContextResolver resolves target identifiers based on current context
type ContextResolver interface {
	// ResolveIdentifier resolves target identifier based on context
	ResolveIdentifier(ctx *core.Context, identifier string) (*core.ResolutionResult, error)

	// GetResolutionSuggestions provides completion suggestions
	GetResolutionSuggestions(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error)
}

// HookRunRequest contains the context needed to execute hooks
type HookRunRequest struct {
	HookType       core.HookType
	WorktreePath   string
	ProjectName    string
	BranchName     string
	SourceBranch   string
	MainRepoPath   string
	ConfigFilePath string
}

// HookRunner defines the interface for executing post-create hooks
type HookRunner interface {
	// Run executes hooks of the specified type with the given request context
	Run(ctx context.Context, req *HookRunRequest) (*core.HookResult, error)
}

// GoGitClient defines go-git operations (deterministic routing - no CLI fallback)
// All methods SHALL be idempotent and thread-safe
type GoGitClient interface {
	// OpenRepository opens git repository (pure function, idempotent)
	OpenRepository(path string) (*git.Repository, error)

	// ListBranches lists all branches in repository (idempotent)
	ListBranches(ctx context.Context, repoPath string) ([]core.BranchInfo, error)

	// BranchExists checks if branch exists (idempotent)
	BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)

	// GetRepositoryStatus returns repository status (idempotent)
	GetRepositoryStatus(ctx context.Context, repoPath string) (core.RepositoryStatus, error)

	// ValidateRepository checks if path contains valid git repository (pure function)
	ValidateRepository(path string) error

	// GetRepositoryInfo returns comprehensive repository information
	GetRepositoryInfo(ctx context.Context, repoPath string) (*core.GitRepository, error)

	// ListRemotes lists all remotes in repository
	ListRemotes(ctx context.Context, repoPath string) ([]core.RemoteInfo, error)

	// GetCommitInfo returns information about a specific commit
	GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*core.CommitInfo, error)
}

// CLIClient defines CLI operations for worktree management ONLY
// All methods SHALL be idempotent and thread-safe
type CLIClient interface {
	// CreateWorktree creates new worktree using git CLI (idempotent)
	CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch string, worktreePath string) error

	// DeleteWorktree removes worktree using git CLI (idempotent, no-op if already deleted)
	DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error

	// ListWorktrees lists all worktrees using git CLI (idempotent)
	ListWorktrees(ctx context.Context, repoPath string) ([]core.WorktreeInfo, error)

	// PruneWorktrees removes stale worktree references
	PruneWorktrees(ctx context.Context, repoPath string) error

	// IsBranchMerged checks if a branch is merged into the current branch
	IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error)

	// DeleteBranch deletes a branch using git CLI (handles worktree-referenced branches)
	DeleteBranch(ctx context.Context, repoPath, branchName string) error
}

// RepoLocator discovers git repositories within a directory tree
type RepoLocator interface {
	FindGitRepositories(dir string) ([]core.GitDir, error)
}

// ShellInfrastructure defines low-level shell infrastructure operations
type ShellInfrastructure interface {
	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(shellType core.ShellType) (string, error)

	// ComposeWrapper composes a custom template with placeholder replacements
	ComposeWrapper(template string, shellType core.ShellType) string

	// DetectConfigFile detects the appropriate config file for the shell type
	DetectConfigFile(shellType core.ShellType) (string, error)

	// InstallWrapper installs the wrapper to the shell config file
	InstallWrapper(shellType core.ShellType, wrapper, configFile string, force bool) error

	// ValidateInstallation validates whether the wrapper is installed
	ValidateInstallation(shellType core.ShellType, configFile string) error
}

// ContextService provides context detection and resolution operations
type ContextService interface {
	// GetCurrentContext returns the current working context
	GetCurrentContext() (*core.Context, error)

	// DetectContextFromPath detects context from a file system path
	DetectContextFromPath(path string) (*core.Context, error)

	// ResolveIdentifier resolves an identifier to a resolution result
	ResolveIdentifier(identifier string) (*core.ResolutionResult, error)

	// ResolveIdentifierFromContext resolves an identifier within a specific context
	ResolveIdentifierFromContext(ctx *core.Context, identifier string) (*core.ResolutionResult, error)

	// GetCompletionSuggestions provides completion suggestions for partial identifiers
	GetCompletionSuggestions(partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error)

	// GetCompletionSuggestionsFromContext provides completion suggestions for partial identifiers within a specific context
	GetCompletionSuggestionsFromContext(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error)
}

// WorktreeService provides high-level worktree management operations
type WorktreeService interface {
	// CreateWorktree creates a new worktree for the specified project and branch
	CreateWorktree(ctx context.Context, req *core.CreateWorktreeRequest) (*core.CreateWorktreeResult, error)

	// DeleteWorktree deletes an existing worktree
	DeleteWorktree(ctx context.Context, req *core.DeleteWorktreeRequest) error

	// ListWorktrees lists all worktrees for a project
	ListWorktrees(ctx context.Context, req *core.ListWorktreesRequest) ([]*core.WorktreeInfo, error)

	// GetWorktreeStatus retrieves the status of a specific worktree
	GetWorktreeStatus(ctx context.Context, worktreePath string) (*core.WorktreeStatus, error)

	// ValidateWorktree validates that a worktree is properly configured
	ValidateWorktree(ctx context.Context, worktreePath string) error

	// PruneMergedWorktrees deletes merged worktrees with optional branch deletion
	PruneMergedWorktrees(ctx context.Context, req *core.PruneWorktreesRequest) (*core.PruneWorktreesResult, error)

	// BranchExists checks if a branch exists in the project
	BranchExists(ctx context.Context, projectPath, branchName string) (bool, error)

	// IsBranchMerged checks if a branch has been merged into its base branch
	IsBranchMerged(ctx context.Context, worktreePath, branchName string) (bool, error)

	// GetWorktreeByPath retrieves worktree info by its path
	GetWorktreeByPath(ctx context.Context, projectPath, worktreePath string) (*core.WorktreeInfo, error)
}

// ProjectService provides project discovery and management operations
type ProjectService interface {
	// DiscoverProject discovers a project by name or from context
	DiscoverProject(ctx context.Context, projectName string, context *core.Context) (*core.ProjectInfo, error)

	// ValidateProject validates that a project is properly configured
	ValidateProject(ctx context.Context, projectPath string) error

	// ListProjects lists all available projects with full info (including worktrees)
	ListProjects(ctx context.Context) ([]*core.ProjectInfo, error)

	// ListProjectSummaries lists project names/paths without loading worktrees
	ListProjectSummaries(ctx context.Context) ([]*core.ProjectSummary, error)

	// GetProjectInfo retrieves detailed information about a project
	GetProjectInfo(ctx context.Context, projectPath string) (*core.ProjectInfo, error)
}

// NavigationService provides path resolution and navigation operations
type NavigationService interface {
	// ResolvePath resolves a target identifier to a concrete path
	ResolvePath(ctx context.Context, req *core.ResolvePathRequest) (*core.ResolutionResult, error)

	// ValidatePath validates that a path is accessible and valid
	ValidatePath(ctx context.Context, path string) error

	// GetNavigationSuggestions provides completion suggestions for navigation
	GetNavigationSuggestions(ctx context.Context, context *core.Context, partial string) ([]*core.ResolutionSuggestion, error)
}

// ShellService provides shell integration and wrapper management operations
type ShellService interface {
	// SetupShell sets up shell integration for the specified shell type
	SetupShell(ctx context.Context, req *core.SetupShellRequest) (*core.SetupShellResult, error)

	// ValidateInstallation validates whether shell integration is installed
	ValidateInstallation(ctx context.Context, req *core.ValidateInstallationRequest) (*core.ValidateInstallationResult, error)

	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(ctx context.Context, req *core.GenerateWrapperRequest) (*core.GenerateWrapperResult, error)
}
