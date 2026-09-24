package core

import "context"

// Role interfaces for git I/O consumers.
//
// The composite *git.Client (defined in internal/git/client.go) embeds *reader
// (read-side go-git operations) and *cliClient (write-side git CLI operations),
// so it satisfies every role below through embedded promotion. The role
// interfaces are declared consumer-side per the golang-structs-interfaces skill:
// "define interfaces where consumed; accept interfaces, return structs."
//
// Each role is sized to 1-3 methods, the range recommended by the
// golang-structs-interfaces skill ("The bigger the interface, the weaker the
// abstraction"). WorktreeWriter is the documented exception at 4 methods —
// see its docstring for the rationale and the deferred split.
//
// The roles are additive: cmd/ continues to consume *git.Client directly
// through cmdutil.Factory.GitClient; per-role cmdutil.Factory fields expose
// type-narrowed access for future read-only or write-only commands.

// RepositoryOpener validates a path as a usable git repository.
//
// OpenRepository (which returns *go-git.Repository) is intentionally excluded
// from this role: no command consumes it directly, and *go-git.Repository is
// not an exportable return type in internal/core/ (the core-isolation
// depguard permits only stdlib + samber/lo). OpenRepository remains a
// concrete method on *reader and is consumed only by other *reader methods
// and by integration tests.
type RepositoryOpener interface {
	ValidateRepository(path string) error
}

// BranchReader exposes the branch-listing and branch-existence methods.
type BranchReader interface {
	ListBranches(ctx context.Context, repoPath string) ([]BranchInfo, error)
	BranchExists(ctx context.Context, repoPath, branchName string) (bool, error)
}

// RepositoryReader exposes the repository-inspection methods.
type RepositoryReader interface {
	GetRepositoryStatus(ctx context.Context, repoPath string) (RepositoryStatus, error)
	GetRepositoryInfo(ctx context.Context, repoPath string) (*GitRepository, error)
	GetCommitInfo(ctx context.Context, repoPath, commitHash string) (*CommitInfo, error)
}

// RemoteReader exposes the remote-listing method.
type RemoteReader interface {
	ListRemotes(ctx context.Context, repoPath string) ([]RemoteInfo, error)
}

// WorktreeWriter exposes the worktree-lifecycle methods.
//
// WorktreeWriter sits at 4 methods, exceeding the 1-3 method rule recommended
// by golang-structs-interfaces. The methods form a tightly coupled
// worktree-lifecycle set: every code path that creates a worktree also
// lists, deletes, or prunes them. Further splitting is deferred as a
// follow-up because it would split methods that are always used together.
type WorktreeWriter interface {
	CreateWorktree(ctx context.Context, repoPath, branchName, sourceBranch, worktreePath string) error
	DeleteWorktree(ctx context.Context, repoPath, worktreePath string, force bool) error
	ListWorktrees(ctx context.Context, repoPath string) ([]WorktreeInfo, error)
	PruneWorktrees(ctx context.Context, repoPath string) error
}

// BranchWriter exposes the branch-mutation methods.
type BranchWriter interface {
	DeleteBranch(ctx context.Context, repoPath, branchName string) error
	IsBranchMerged(ctx context.Context, repoPath, branchName string) (bool, error)
}
