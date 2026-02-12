package domain

import (
	"time"
)

// BranchInfo represents information about a git branch
type BranchInfo struct {
	Name      string    // Branch name
	IsCurrent bool      // Whether this is the current branch
	Remote    string    // Remote tracking branch (if any)
	Commit    string    // Latest commit hash
	Author    string    // Author of latest commit
	Date      time.Time // Date of latest commit
}

// WorktreeInfo represents information about a git worktree
type WorktreeInfo struct {
	Path       string // Absolute path to worktree
	Branch     string // Branch name
	Commit     string // Commit hash
	IsBare     bool   // Whether this is a bare worktree
	IsDetached bool   // Whether worktree is in detached HEAD state
	Modified   bool   // Whether worktree has uncommitted changes
}

// RepositoryStatus represents the status of a git repository
type RepositoryStatus struct {
	IsClean   bool     // Whether working directory is clean
	Branch    string   // Current branch name
	Commit    string   // Current commit hash
	Modified  []string // List of modified files
	Added     []string // List of added files
	Deleted   []string // List of deleted files
	Untracked []string // List of untracked files
	Ahead     int      // Commits ahead of remote
	Behind    int      // Commits behind remote
}

// RemoteInfo represents information about a git remote
type RemoteInfo struct {
	Name     string // Remote name
	FetchURL string // Fetch URL
	PushURL  string // Push URL
}

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash      string    // Commit hash
	Author    string    // Author name
	Email     string    // Author email
	Date      time.Time // Commit date
	Message   string    // Commit message
	ShortHash string    // Short commit hash (7 characters)
}

// GitRepository represents a git repository with metadata
type GitRepository struct {
	Path          string           // Repository path
	IsBare        bool             // Whether repository is bare
	DefaultBranch string           // Default branch name
	Remotes       []RemoteInfo     // List of remotes
	Branches      []BranchInfo     // List of branches
	Worktrees     []WorktreeInfo   // List of worktrees
	Status        RepositoryStatus // Current status
}
