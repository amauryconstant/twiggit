# Spec Delta

## MODIFIED Requirements

### Requirement: Routing table

The composite git client exposes operations to read-side roles or
write-side roles as follows:

| Operation | Concrete | Role | Rationale |
|---|---|---|---|
| `ValidateRepository` | go-git | `RepositoryOpener` | Portable, deterministic |
| `ListBranches` | go-git | `BranchReader` | Portable, deterministic |
| `BranchExists` | go-git | `BranchReader` | Portable, deterministic |
| `GetRepositoryStatus` | go-git | `RepositoryReader` | Portable, deterministic |
| `GetRepositoryInfo` | go-git | `RepositoryReader` | Portable, deterministic |
| `GetCommitInfo` | go-git | `RepositoryReader` | Portable, deterministic |
| `ListRemotes` | go-git | `RemoteReader` | Portable, deterministic |
| `CreateWorktree` | git CLI | `WorktreeWriter` | go-git lacks support |
| `DeleteWorktree` | git CLI | `WorktreeWriter` | go-git lacks support |
| `ListWorktrees` | git CLI | `WorktreeWriter` | go-git lacks support |
| `PruneWorktrees` | git CLI | `WorktreeWriter` | go-git lacks support |
| `DeleteBranch` | git CLI | `BranchWriter` | Handles worktree-referenced branches |
| `IsBranchMerged` | git CLI | `BranchWriter` | go-git limitations |

`OpenRepository` (exposed by the read-side concrete but not in any role
interface per `core-git`) is omitted from the table intentionally; it
is a concrete-only method consumed by other read-side methods and by
integration tests, not by any command.

This routing table is canonical for which concrete client implements
each operation; the per-method signatures and consumer-side placement
live in `core-git`. Callers SHALL NOT reach into a single
implementation; each role interface is the injection point for the
operations listed against it. The composite `git.Client` remains the
canonical injection point for commands that need both read and write
operations; see the ADDED requirement "Per-role Factory fields are
additive to composite" for the additive per-role narrowing.

#### Scenario: Routing table matches the documented shape

- **WHEN** the role surface described above is exercised
- **THEN** each operation in the table SHALL be matched by exactly one
  role
- **AND** every role's full method set SHALL be defined in `core-git`
- **AND** the composite `git.Client` SHALL satisfy every role via
  embedded promotion of `*reader` and `*cliClient`

## ADDED Requirements

### Requirement: Per-role Factory fields are additive to composite

The `cmdutil.Factory` SHALL expose one lazy `func() (core.Role,
error)` field per role defined in `core-git`: `RepoOpener`,
`BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`,
`BranchWriter`. Each field SHALL return the same `*git.Client`
instance as the composite `f.GitClient()` field, cached via
`sync.OnceValues` on the composite (not a separate `sync.Once` per
role).

The composite `f.GitClient() (*git.Client, error)` field SHALL be
retained unchanged and SHALL remain the canonical injection point for
commands that need both read and write operations. Per-role fields
are an opt-in narrowing; commands SHALL NOT be required to migrate
to per-role fields unless they actually need a narrow role
dependency.

The composite `git.NewClient(opts ...)` constructor SHALL be retained
unchanged. No prior name (`git.NewGoGitClient`, `git.GitClient`) is
deprecated by this change because the prior names were never present
in the Tier 2 layout.

#### Scenario: Composite field remains canonical

- **WHEN** `internal/cmdutil/factory.go` is read
- **THEN** the `GitClient func() (*git.Client, error)` field SHALL
  exist with the same signature and the same `sync.OnceValues` cache
  it had before this change
- **AND** every existing `f.GitClient()` call site in `cmd/*.go`
  SHALL continue to compile unchanged

#### Scenario: Per-role call returns the same composite instance

- **WHEN** a command calls `f.BranchReader()` and `f.RepositoryReader()`
  in sequence
- **THEN** both calls SHALL return the same `*git.Client` instance
  typed as the requested role
- **AND** the underlying LRU cache SHALL be initialized exactly once

#### Scenario: Per-role fields are additive, not a replacement

- **WHEN** this change is applied
- **THEN** the `git.NewClient` constructor SHALL continue to exist
  with the same signature
- **AND** the composite `git.Client` type SHALL continue to be the
  caller-side name; no type is renamed, hidden, or replaced

### Requirement: NewCLIClient returns the write-side concrete client

`git.NewCLIClient(executor)` SHALL return `(*git.CLIClient, error)`
where `CLIClient` is the exported type alias for `cliClient` (the
unexported write-side concrete). The returned concrete SHALL satisfy
both write-side roles (`WorktreeWriter`, `BranchWriter`) via the
compile-time role-satisfaction declarations on the composite
`*git.Client` (per `core-git` "Composite satisfies all six roles via
embedded promotion" and `core-git` "Drift sentinel enforces
method-set invariants").

#### Scenario: NewCLIClient returns success

- **WHEN** the caller invokes `c, err := git.NewCLIClient(executor)`
- **THEN** `err` SHALL be nil and `c` SHALL be a non-nil `*git.CLIClient`
- **AND** the composite `*git.Client` constructed via `git.NewClient()`
  SHALL satisfy `core.WorktreeWriter` and `core.BranchWriter` at
  compile time

### Requirement: All failures wrapped via git.New*Error family

All failures from read-side and write-side clients SHALL be wrapped
via the canonical error constructors: `git.NewRepoError`,
`git.NewWorktreeError`, and `git.NewCommandError` (defined in
`internal/git/errors.go`). These constructors return
`*git.ExternalError` whose embedded core type walks to
`*core.OperationError` with `Op` set to the operation name, `Message`
naming the resource, and `Cause` set via `%w`. No concrete client
SHALL return raw `go-git` or `os/exec` errors.

#### Scenario: Read-side failure surfaces as *core.OperationError

- **WHEN** `git.PlainOpen(path)` fails inside the read-side concrete
  client
- **THEN** the returned error's `errors.As(err, &*core.OperationError{})`
  SHALL be true with `Op = "git.repository"`, `Message` naming the
  path, and `Cause` set via `%w`

#### Scenario: Write-side failure surfaces as *core.OperationError

- **WHEN** the git CLI subprocess returns a non-zero exit code
  inside the write-side concrete client
- **THEN** the returned error's `errors.As(err, &*core.OperationError{})`
  SHALL be true with `Op = "git.worktree.<command>"` (or `Op =
  "git.branch.<command>"`) and a `Cause` wrapping the underlying
  exit-code error

### Requirement: Role interface segregation contract references core-git

The role-interface segregation contract — the method-set boundaries of
`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`,
`WorktreeWriter`, `BranchWriter`, the consumer-side placement in
`internal/core/git.go`, the composite satisfaction via embedded
promotion, the additive Factory per-role fields, and the single
drift sentinel in `internal/git/client_test.go` — SHALL be owned by
`core-git`. Readers of this capability SHALL consult `core-git` for
the per-role method sets and the placement / drift rules rather than
re-deriving them from concrete-client signatures.

#### Scenario: core-git owns the role method sets

- **WHEN** a reader of `git-client` requires the precise method
  signatures of any role
- **THEN** the source of truth SHALL be `core-git`
- **AND** this capability SHALL NOT redefine them
