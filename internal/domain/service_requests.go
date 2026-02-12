package domain

// CreateWorktreeRequest represents a request to create a new worktree
type CreateWorktreeRequest struct {
	ProjectName  string   // Name of the project
	BranchName   string   // Name of the branch to create
	SourceBranch string   // Source branch to create from
	Context      *Context // Current context for resolution
	Force        bool     // Force creation even if branch exists
}

// DeleteWorktreeRequest represents a request to delete a worktree
type DeleteWorktreeRequest struct {
	WorktreePath string   // Path to the worktree to delete
	Force        bool     // Force deletion even if there are uncommitted changes
	Context      *Context // Current context for validation
}

// ListWorktreesRequest represents a request to list worktrees
type ListWorktreesRequest struct {
	ProjectName     string   // Name of the project (optional, uses context if empty)
	Context         *Context // Current context for project resolution
	IncludeMain     bool     // Include main worktree in results
	ListAllProjects bool     // List worktrees from all discovered projects (overrides ProjectName)
}

// ResolvePathRequest represents a request to resolve a path identifier
type ResolvePathRequest struct {
	Target  string   // Target identifier to resolve
	Context *Context // Current context for resolution
	Search  bool     // Enable search if exact match not found
}
