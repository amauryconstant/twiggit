package domain

// PathValidator defines infrastructure-specific path validation operations
// This interface belongs to domain because it's used by domain entities,
// but implementations are in infrastructure layer
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
