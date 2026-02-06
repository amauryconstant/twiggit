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

// GetWorktreeStatusRequest represents a request to get worktree status
type GetWorktreeStatusRequest struct {
	WorktreePath string   // Path to the worktree
	Context      *Context // Current context for validation
}

// ValidateWorktreeRequest represents a request to validate a worktree
type ValidateWorktreeRequest struct {
	WorktreePath string   // Path to the worktree to validate
	Context      *Context // Current context for validation
}

// DiscoverProjectRequest represents a request to discover a project
type DiscoverProjectRequest struct {
	ProjectName string   // Name of the project to discover
	Context     *Context // Current context for resolution
	SearchPath  string   // Path to search in (optional)
}

// ValidateProjectRequest represents a request to validate a project
type ValidateProjectRequest struct {
	ProjectPath string   // Path to the project to validate
	Context     *Context // Current context for validation
}

// ListProjectsRequest represents a request to list projects
type ListProjectsRequest struct {
	SearchPath string   // Path to search in (optional)
	Context    *Context // Current context for resolution
}

// GetProjectInfoRequest represents a request to get project information
type GetProjectInfoRequest struct {
	ProjectPath string   // Path to the project
	Context     *Context // Current context for validation
}

// ResolvePathRequest represents a request to resolve a path identifier
type ResolvePathRequest struct {
	Target  string   // Target identifier to resolve
	Context *Context // Current context for resolution
	Search  bool     // Enable search if exact match not found
}

// ValidatePathRequest represents a request to validate a path
type ValidatePathRequest struct {
	Path    string   // Path to validate
	Context *Context // Current context for validation
}

// GetNavigationSuggestionsRequest represents a request for navigation suggestions
type GetNavigationSuggestionsRequest struct {
	Context  *Context // Current context for suggestions
	Partial  string   // Partial input for completion
	MaxCount int      // Maximum number of suggestions to return
	Search   bool     // Enable search-based suggestions
}
