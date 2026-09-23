# Spec Delta

## Purpose

Documents the composite `git.GitClient` for the Tier 2 layout: a single client type that routes read operations to `GoGitClient` and mutation operations to `CLIClient`. The legacy `infrastructure-git-client` spec retains the two-role-interface design per the deferred-migration non-goal in the proposal; this new spec is the Tier 2 path going forward and the canonical surface for new code.

## ADDED Requirements

### Requirement: Routing table

The composite client SHALL route operations as follows:

| Operation | GoGitClient | CLIClient | Rationale |
|---|:---:|:---:|---|
| `OpenRepository` | Yes | No | Portable, deterministic |
| `ListBranches` | Yes | No | Portable, deterministic |
| `BranchExists` | Yes | No | Portable, deterministic |
| `GetRepositoryStatus` | Yes | No | Portable, deterministic |
| `ValidateRepository` | Yes | No | Portable, deterministic |
| `GetRepositoryInfo` | Yes | No | Portable, deterministic |
| `ListRemotes` | Yes | No | Portable, deterministic |
| `GetCommitInfo` | Yes | No | Portable, deterministic |
| `CreateWorktree` | No | Yes | go-git lacks support |
| `DeleteWorktree` | No | Yes | go-git lacks support |
| `ListWorktrees` | No | Yes | go-git lacks support |
| `PruneWorktrees` | No | Yes | go-git lacks support |
| `IsBranchMerged` | No | Yes | go-git limitations |
| `DeleteBranch` | No | Yes | Handles worktree-referenced branches |

This routing table is canonical. Callers SHALL NOT reach into a single implementation; the composite `git.GitClient` is the only injection point.

#### Scenario: Routing table matches the documented shape
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: git.NewClient constructor returns the composite

`git.NewClient(opts ...ClientOption) (*git.GitClient, error)` SHALL return a composite client with cache enabled (default size 25). `git.NewClient(git.WithCacheSize(n))` SHALL accept a custom size; `n <= 0` SHALL fall back to 25. `git.NewClient(git.WithCacheDisabled())` SHALL bypass the cache entirely. The constructor SHALL return a `*core.OperationError` wrapping the LRU cache initialization failure when cache init fails. The previous name `git.NewGoGitClient()` is removed.

#### Scenario: Constructor returns success
- **WHEN** the caller invokes `c, err := git.NewClient()`
- **THEN** `err` SHALL be nil and `c` SHALL be a non-nil `*git.GitClient` with the default cache size

#### Scenario: Custom size overrides default
- **WHEN** the caller invokes `c, err := git.NewClient(git.WithCacheSize(100))`
- **THEN** `c` SHALL use cache size 100

#### Scenario: Non-positive size falls back to 25
- **WHEN** the caller invokes `c, err := git.NewClient(git.WithCacheSize(0))`
- **THEN** `c` SHALL use the default cache size 25

#### Scenario: No stutter at call sites
- **WHEN** the test reads `c, err := git.NewClient()`
- **THEN** the package qualifier `git` is the only occurrence of the package name in the qualified identifier (no `git.GitClient` stutter in the constructor name)

### Requirement: All failures wrapped via git.New*Error

All failures SHALL be wrapped via `git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError` (defined in `internal/git/errors.go`). These constructors return `*git.ExternalError` whose embedded core type is `*core.OperationError` with `Op` set to the operation name, `Message` naming the resource, and `Cause` set via `%w`. The composite client SHALL NOT return raw `go-git` or `os/exec` errors.

#### Scenario: Failure surfaces as *core.OperationError
- **WHEN** `git.PlainOpen(path)` fails
- **THEN** the returned error's `errors.As(err, &*core.OperationError{})` returns `true` with `Op = "git.open"`, `Message` naming the path, and `Cause` set via `%w`
