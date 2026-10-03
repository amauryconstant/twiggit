// Package git is the git I/O adapter for twiggit.
//
// It owns the git client lifecycle (open, cache, close), the read-side
// surface (OpenRepository, ListBranches, BranchExists, RepositoryStatus,
// ListRemotes, Commit, Repository, ValidateRepository), the
// write-side surface (worktree mutations, branch mutations), the
// context-detection resolver, the hook runner, and the I/O-shelled portion
// of shell-detection (os.Stat probing).
//
// All public types and functions return core.* values where they cross back
// into command wiring; this package may import internal/core but not the
// reverse.
package git
