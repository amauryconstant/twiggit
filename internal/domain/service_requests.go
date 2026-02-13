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

// PruneWorktreesRequest represents a request to prune merged worktrees
type PruneWorktreesRequest struct {
	ProjectName      string   // Name of the project (optional, uses context if empty)
	Context          *Context // Current context for project resolution
	Force            bool     // Force pruning even with uncommitted changes
	DeleteBranches   bool     // Delete branches after worktree removal
	DryRun           bool     // Preview only, no actual deletion
	AllProjects      bool     // Prune across all projects
	SpecificWorktree string   // Specific worktree to prune (project/branch format)
}

// PruneWorktreesResult represents the result of a prune operation
type PruneWorktreesResult struct {
	DeletedWorktrees       []*PruneWorktreeResult // Worktrees that were deleted
	SkippedWorktrees       []*PruneWorktreeResult // Worktrees that were skipped
	ProtectedSkipped       []*PruneWorktreeResult // Worktrees skipped due to protected branch
	UnmergedSkipped        []*PruneWorktreeResult // Worktrees skipped due to unmerged status
	CurrentWorktreeSkipped []*PruneWorktreeResult // Worktrees skipped because they are the current worktree
	NavigationPath         string                 // Path to navigate to after single worktree prune
	TotalDeleted           int                    // Total number of worktrees deleted
	TotalSkipped           int                    // Total number of worktrees skipped
	TotalBranchesDeleted   int                    // Total number of branches deleted
}

// PruneWorktreeResult represents the result of pruning a single worktree
type PruneWorktreeResult struct {
	ProjectName   string // Name of the project
	WorktreePath  string // Path to the worktree
	BranchName    string // Branch name
	Deleted       bool   // Whether the worktree was deleted
	BranchDeleted bool   // Whether the branch was deleted
	SkipReason    string // Reason for skipping (if applicable)
	Error         error  // Error that occurred during pruning (if any)
}
