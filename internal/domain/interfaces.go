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
