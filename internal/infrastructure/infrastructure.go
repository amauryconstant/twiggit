package infrastructure

// InfrastructureService provides external dependencies for domain operations
// This interface abstracts away filesystem and git repository operations
// to keep domain entities pure and testable
type InfrastructureService interface {
	// PathExists checks if a path exists on the filesystem
	PathExists(path string) bool

	// PathWritable checks if a path is writable
	PathWritable(path string) bool

	// IsGitRepository checks if a path is a valid git repository
	IsGitRepository(path string) bool
}
