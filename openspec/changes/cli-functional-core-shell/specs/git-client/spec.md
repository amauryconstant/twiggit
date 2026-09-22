# Spec Delta

## MODIFIED Requirements

### Requirement: Routing table

The composite client SHALL route operations as follows:
| Operation | GoGitClient | CLIClient | Rationale |
|---|:---:|:---:|---|
| `OpenRepository` | ✅ | ❌ | Portable, deterministic |
| `ListBranches` | ✅ | ❌ | Portable, deterministic |
| `BranchExists` | ✅ | ❌ | Portable, deterministic |
| `GetRepositoryStatus` | ✅ | ❌ | Portable, deterministic |
| `ValidateRepository` | ✅ | ❌ | Portable, deterministic |
| `GetRepositoryInfo` | ✅ | ❌ | Portable, deterministic |
| `ListRemotes` | ✅ | ❌ | Portable, deterministic |
| `GetCommitInfo` | ✅ | ❌ | Portable, deterministic |
| `CreateWorktree` | ❌ | ✅ | go-git lacks support |
| `DeleteWorktree` | ❌ | ✅ | go-git lacks support |
| `ListWorktrees` | ❌ | ✅ | go-git lacks support |
| `PruneWorktrees` | ❌ | ✅ | go-git lacks support |
| `IsBranchMerged` | ❌ | ✅ | go-git limitations |
| `DeleteBranch` | ❌ | ✅ | Handles worktree-referenced branches |

This routing table is canonical. Callers SHALL NOT reach into a single implementation; the composite `git.GitClient` is the only injection point.

#### Scenario: Routing table matches the documented shape
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: GoGitClient constructor returns error

`git.NewGoGitClient()` SHALL return `(*git.GoGitClient, error)` and the client SHALL be created with cache enabled (default size 25). `git.NewGoGitClientWithSize(n)` SHALL accept a custom size; `n <= 0` SHALL fall back to 25. `cacheEnabled=false` SHALL bypass the cache entirely. The constructor SHALL return a `*core.OperationError` wrapping the LRU cache initialization failure when cache init fails.

#### Scenario: Constructor returns success
- **WHEN** the caller invokes `c, err := git.NewGoGitClient()`
- **THEN** `err` SHALL be nil and `c` SHALL be a non-nil `*git.GoGitClient` with the default cache size

#### Scenario: Custom size overrides default
- **WHEN** the caller invokes `c, err := git.NewGoGitClientWithSize(100)`
- **THEN** `c` SHALL use cache size 100

#### Scenario: Non-positive size falls back to 25
- **WHEN** the caller invokes `c, err := git.NewGoGitClientWithSize(0)`
- **THEN** `c` SHALL use the default cache size 25

### Requirement: All failures wrapped via core.NewGit*Error

All failures SHALL be wrapped via `core.NewGitRepositoryError` / `core.NewGitWorktreeError` / `core.NewGitCommandError`. The composite client SHALL NOT return raw `go-git` or `os/exec` errors.

#### Scenario: Failure surfaces as core.OperationError
- **WHEN** `git.PlainOpen(path)` fails
- **THEN** the returned error SHALL be a `*core.OperationError` carrying `Op = "git.open"`, `Message` naming the path, and `Cause` set via `%w`
