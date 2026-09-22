# Spec Delta

## MODIFIED Requirements

### Requirement: Role interface segregation routing table

The composite client SHALL route operations to read-side roles (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`) implemented by the go-git-backed concrete client, or to write-side roles (`WorktreeWriter`, `BranchWriter`) implemented by the git-CLI-backed concrete client, as follows:

| Operation | Satisfied by | Role | Rationale |
|---|---|---|---|
| `OpenRepository` | go-git concrete client | `RepositoryOpener` | Portable, deterministic |
| `ValidateRepository` | go-git concrete client | `RepositoryOpener` | Portable, deterministic |
| `ListBranches` | go-git concrete client | `BranchReader` | Portable, deterministic |
| `BranchExists` | go-git concrete client | `BranchReader` | Portable, deterministic |
| `GetRepositoryStatus` | go-git concrete client | `RepositoryReader` | Portable, deterministic |
| `GetRepositoryInfo` | go-git concrete client | `RepositoryReader` | Portable, deterministic |
| `GetCommitInfo` | go-git concrete client | `RepositoryReader` | Portable, deterministic |
| `ListRemotes` | go-git concrete client | `RemoteReader` | Portable, deterministic |
| `CreateWorktree` | git CLI concrete client | `WorktreeWriter` | go-git lacks support |
| `DeleteWorktree` | git CLI concrete client | `WorktreeWriter` | go-git lacks support |
| `ListWorktrees` | git CLI concrete client | `WorktreeWriter` | go-git lacks support |
| `PruneWorktrees` | git CLI concrete client | `WorktreeWriter` | go-git lacks support |
| `DeleteBranch` | git CLI concrete client | `BranchWriter` | Handles worktree-referenced branches |
| `IsBranchMerged` | git CLI concrete client | `BranchWriter` | go-git limitations |

This routing table is canonical. Callers SHALL NOT reach into a single implementation; each role interface is the injection point for the operations listed against it. The method-set drift is enforced at compile time by `var _ core.Role = (*Concrete)(nil)` declarations placed next to each concrete type.

#### Scenario: Routing table matches the documented shape
- **WHEN** the role surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the concrete clients SHALL compile against every role they satisfy

### Requirement: NewGoGitClient constructor returns error

`git.NewGoGitClient()` SHALL return `(*git.GoGitClient, error)` and the client SHALL be created with cache enabled (default size 25). `git.NewGoGitClientWithSize(n)` SHALL accept a custom size; `n <= 0` SHALL fall back to 25. `cacheEnabled=false` SHALL bypass the cache entirely. The constructor SHALL return a `*core.OperationError` wrapping the LRU cache initialization failure when cache init fails.

The concrete `*git.GoGitClient` SHALL satisfy all four read-side roles (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`) via compile-time `var _ core.Role = (*GoGitClient)(nil)` declarations placed next to the type.

#### Scenario: Constructor returns success
- **WHEN** the caller invokes `c, err := git.NewGoGitClient()`
- **THEN** `err` SHALL be nil and `c` SHALL be a non-nil `*git.GoGitClient` with the default cache size

#### Scenario: Custom size overrides default
- **WHEN** the caller invokes `c, err := git.NewGoGitClientWithSize(100)`
- **THEN** `c` SHALL use cache size 100

#### Scenario: Non-positive size falls back to 25
- **WHEN** the caller invokes `c, err := git.NewGoGitClientWithSize(0)`
- **THEN** `c` SHALL use the default cache size 25

### Requirement: NewCLIClient returns the write-side concrete client

`git.NewCLIClient(executor)` SHALL return `(*git.CLIClient, error)` and the client SHALL wrap the supplied executor. The concrete `*git.CLIClient` SHALL satisfy both write-side roles (`WorktreeWriter`, `BranchWriter`) via compile-time `var _ core.Role = (*CLIClient)(nil)` declarations placed next to the type.

#### Scenario: NewCLIClient returns success
- **WHEN** the caller invokes `c, err := git.NewCLIClient(executor)`
- **THEN** `err` SHALL be nil and `c` SHALL be a non-nil `*git.CLIClient` typed as both `core.WorktreeWriter` and `core.BranchWriter` via compile-time assertions

### Requirement: All failures wrapped via core.NewGit*Error

All failures SHALL be wrapped via `core.NewGitRepositoryError` / `core.NewGitWorktreeError` / `core.NewGitCommandError`. Neither concrete client SHALL return raw `go-git` or `os/exec` errors.

#### Scenario: Failure surfaces as core.OperationError
- **WHEN** `git.PlainOpen(path)` fails inside the go-git concrete client
- **THEN** the returned error SHALL be a `*core.OperationError` carrying `Op = "git.open"`, `Message` naming the path, and `Cause` set via `%w`

#### Scenario: Worktree failure surfaces as core.OperationError
- **WHEN** the git CLI subprocess returns a non-zero exit code inside the CLI concrete client
- **THEN** the returned error SHALL be a `*core.OperationError` carrying `Op = "git.worktree.<command>"` and a `Cause` that wraps the underlying exit-code error

## ADDED Requirements

### Requirement: Role interface segregation contract references core-git

The role-interface segregation contract — the method-set boundaries of `RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`, the consumer-side placement, the factory wiring contract, and the test mock structure — SHALL be owned by `core-git`. Readers of this capability SHALL consult `core-git` for the per-role method sets rather than re-deriving them from concrete-client signatures.

#### Scenario: core-git owns the role method sets
- **WHEN** a reader of `git-client` requires the precise method signatures of any role
- **THEN** the source of truth SHALL be `core-git`
- **AND** this capability SHALL NOT redefine them
