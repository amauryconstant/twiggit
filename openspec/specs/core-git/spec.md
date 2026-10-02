# core-git Specification

## Purpose
Defines the role-interface segregation contract for git operations.
Establishes consumer-side role interfaces that mirror the method sets on
the composite `*git.Client` (which embeds `*reader` and `*cliClient`),
gives compile-time drift detection on the concrete, and exposes
additive per-role lazy fields on `cmdutil.Factory` so future read-only
(or write-only) commands can opt into narrow role dependencies without
disturbing the composite injection point.

## Requirements

### Requirement: RepositoryOpener exposes ValidateRepository

The system SHALL expose a `RepositoryOpener` role interface with exactly
one method: `ValidateRepository(path string) error`. The role SHALL be
defined consumer-side and SHALL NOT include `OpenRepository`. The
godoc SHALL explain that `OpenRepository` is intentionally excluded: no
command consumes it directly, and `*go-git.Repository` is not an
exportable return type in `internal/core/` (`core-isolation` depguard
permits only `$gostd`).

#### Scenario: RepositoryOpener declares one method

- **WHEN** the role is read from `internal/core/git.go`
- **THEN** the role SHALL declare exactly `ValidateRepository(path string) error`
- **AND** the role SHALL NOT declare `OpenRepository` or any other method

#### Scenario: Composite *git.Client satisfies RepositoryOpener

- **WHEN** the composite `*git.Client` is type-asserted against
  `RepositoryOpener`
- **THEN** the assertion SHALL succeed at compile time via a
  `var _ core.RepositoryOpener = (*git.Client)(nil)` declaration in
  `internal/git/client.go`

### Requirement: BranchReader exposes ListBranches and BranchExists

The system SHALL expose a `BranchReader` role interface with exactly two
context-aware methods: `ListBranches` and `BranchExists`. The role SHALL
be defined consumer-side.

#### Scenario: BranchReader is the only role with both list-and-check methods

- **WHEN** a caller requires both `ListBranches` and `BranchExists`
- **THEN** the role SHALL be `BranchReader` and SHALL NOT include any
  other method

#### Scenario: Composite *git.Client satisfies BranchReader

- **WHEN** the composite `*git.Client` is type-asserted against
  `BranchReader`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.BranchReader = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: RepositoryReader exposes GetRepositoryStatus, GetRepositoryInfo, and GetCommitInfo

The system SHALL expose a `RepositoryReader` role interface with exactly
three context-aware methods: `GetRepositoryStatus`,
`GetRepositoryInfo`, and `GetCommitInfo`. The role SHALL be defined
consumer-side.

#### Scenario: RepositoryReader groups repository inspection methods

- **WHEN** a caller requires `GetRepositoryStatus`,
  `GetRepositoryInfo`, and `GetCommitInfo`
- **THEN** the role SHALL be `RepositoryReader` and SHALL NOT include
  branch listing, remote listing, or repository opening

#### Scenario: Composite *git.Client satisfies RepositoryReader

- **WHEN** the composite `*git.Client` is type-asserted against
  `RepositoryReader`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.RepositoryReader = (*git.Client)(nil)` in
  `internal/git/client.go`

### Requirement: RemoteReader exposes ListRemotes

The system SHALL expose a `RemoteReader` role interface with exactly one
context-aware method: `ListRemotes`. The role SHALL be defined
consumer-side.

#### Scenario: RemoteReader is the single-method role for remote listing

- **WHEN** a caller requires `ListRemotes` only
- **THEN** the role SHALL be `RemoteReader` and SHALL NOT include any
  other method

#### Scenario: Composite *git.Client satisfies RemoteReader

- **WHEN** the composite `*git.Client` is type-asserted against
  `RemoteReader`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.RemoteReader = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: WorktreeWriter exposes the four worktree-lifecycle methods

The system SHALL expose a `WorktreeWriter` role interface with exactly
four context-aware methods: `CreateWorktree`, `DeleteWorktree`,
`ListWorktrees`, and `PruneWorktrees`. The role SHALL be defined
consumer-side.

> **Note:** `WorktreeWriter` sits at 4 methods, exceeding the 1-3
> method rule recommended by the
> `golang-structs-interfaces` skill. The methods form a tightly coupled
> worktree-lifecycle set; every code path that creates a worktree also
> lists, deletes, or prunes them. Further splitting is deferred as a
> follow-up.

#### Scenario: WorktreeWriter covers the worktree lifecycle

- **WHEN** a caller requires any combination of `CreateWorktree`,
  `DeleteWorktree`, `ListWorktrees`, or `PruneWorktrees`
- **THEN** the role SHALL be `WorktreeWriter` and SHALL NOT include
  branch operations

#### Scenario: Composite *git.Client satisfies WorktreeWriter

- **WHEN** the composite `*git.Client` is type-asserted against
  `WorktreeWriter`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.WorktreeWriter = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: BranchWriter exposes DeleteBranch and IsBranchMerged

The system SHALL expose a `BranchWriter` role interface with exactly two
context-aware methods: `DeleteBranch` and `IsBranchMerged`. The role
SHALL be defined consumer-side.

#### Scenario: BranchWriter is the only role with both branch-mutation methods

- **WHEN** a caller requires both `DeleteBranch` and `IsBranchMerged`
- **THEN** the role SHALL be `BranchWriter` and SHALL NOT include any
  other method

#### Scenario: Composite *git.Client satisfies BranchWriter

- **WHEN** the composite `*git.Client` is type-asserted against
  `BranchWriter`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.BranchWriter = (*git.Client)(nil)` in `internal/git/client.go`

### Requirement: Role interfaces are defined consumer-side

All six role interfaces SHALL be declared in `internal/core/git.go`. No
role interface SHALL be declared in `internal/git/`, `internal/cmdutil/`,
or any `cmd/` command file. The roles SHALL be importable from
`twiggit/internal/core` without importing any I/O package.

#### Scenario: Role interfaces importable from internal/core only

- **WHEN** a consumer in `internal/core/` requires a git role
- **THEN** the import path SHALL be `twiggit/internal/core` and the
  role name SHALL be one of the six declared in `git.go`

#### Scenario: No role declared outside internal/core

- **WHEN** `internal/git`, `internal/cmdutil`, or `cmd/` is searched
  for interface declarations
- **THEN** no role interface SHALL be declared there

### Requirement: Composite *git.Client satisfies all six roles via embedded promotion

The composite `git.Client` SHALL satisfy all six role interfaces
through embedded promotion of `*reader` (read-side methods) and
`*cliClient` (write-side methods). Each satisfaction SHALL be asserted
at compile time via a `var _ core.Role = (*git.Client)(nil)`
declaration placed in `internal/git/client.go`. The six declarations
SHALL be co-located at the bottom of that file so a single edit covers
all 14 role methods.

#### Scenario: Compile-time checks are placed once on the composite

- **WHEN** `internal/git/client.go` is read
- **THEN** it SHALL contain exactly six `var _ core.Role = (*git.Client)(nil)`
  lines covering all six roles, in the same source file

#### Scenario: Method-set drift on the composite fails the build

- **WHEN** any of the 14 role methods on `*reader` or `*cliClient` is
  renamed or removed
- **THEN** `go build ./...` SHALL fail with a type-mismatch error
  referencing the affected role

### Requirement: Role interfaces respect the Tier 2 depguard

Each role interface declared in `internal/core/git.go` SHALL be
importable from `twiggit/internal/core` without importing any package
other than the Go standard library. No role interface SHALL introduce
a dependency on `internal/git/`, `internal/cmdutil/`, `internal/config/`,
`internal/output/`, `internal/iostreams/`, or `cmd/`. The Tier 2
depguard rule from `openspec/config.yaml`
(`core-isolation: $gostd`) governs this contract.

#### Scenario: Role interfaces import only stdlib

- **WHEN** the import graph of `internal/core/git.go` is computed
- **THEN** every import SHALL be a `$gostd` package
- **AND** no role interface declared in that file SHALL reference
  `*go-git.Repository` or any other type from `internal/git/`

#### Scenario: depguard permits the file

- **WHEN** `golangci-lint run` runs with the project's `.golangci.yml`
  depguard config
- **THEN** `internal/core/git.go` SHALL be admitted under the
  `core-isolation` rule

### Requirement: Drift sentinel enforces method-set invariants

A single sentinel struct SHALL exist in
`internal/git/client_test.go`, embedding all six role fields against
the composite `*git.Client`. Renaming or removing any role method on
the composite SHALL cause `go build ./...` to fail with a type-mismatch
error referencing the affected role. The sentinel SHALL NOT be skipped
or `//go:build`-excluded; it protects the role surface for every
consumer.

#### Scenario: Drift sentinel compiles against the composite

- **WHEN** `go build ./internal/git/...` runs
- **THEN** the `_DriftCheck` sentinel SHALL embed every role field
  that the composite `*git.Client` satisfies

#### Scenario: Method-set drift fails the drift sentinel

- **WHEN** any of the 14 role methods on `*reader` or `*cliClient` is
  renamed or removed
- **THEN** `go build ./...` SHALL fail with a type-mismatch error
  naming the affected role

### Requirement: Factory per-role fields are additive to composite

The `cmdutil.Factory` SHALL expose one lazy `func() (core.Role,
error)` field per role defined in this capability: `RepoOpener`,
`BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`,
`BranchWriter`. Each per-role field SHALL return the same
`*git.Client` instance as `f.GitClient()` (the composite field),
because the per-role field's body routes through `f.GitClient()` and
reuses its existing `sync.OnceValues` cache.

The composite `f.GitClient() (*git.Client, error)` field SHALL remain
unchanged and SHALL be the canonical injection point for any command
that needs both read and write methods. Per-role fields are documented
narrowing for future read-only (or write-only) commands; existing
commands that use `f.GitClient()` SHALL keep working without
modification.

#### Scenario: Per-role call reuses the composite's cached concrete

- **WHEN** a command calls `f.BranchReader()` and `f.WorktreeWriter()`
  in sequence
- **THEN** both calls SHALL return the same `*git.Client` instance
  typed as the requested role
- **AND** the underlying LRU cache SHALL be initialized exactly once
  (the existing `sync.OnceValues` on `f.GitClient` is the only cache)

#### Scenario: Composite field is retained

- **WHEN** `internal/cmdutil/factory.go` is read
- **THEN** the `GitClient func() (*git.Client, error)` field SHALL be
  present with the same signature as before
- **AND** every existing `f.GitClient()` call site in `cmd/*.go` SHALL
  continue to compile unchanged

#### Scenario: Init() touches every per-role field

- **WHEN** `Factory.Init()` is invoked
- **THEN** it SHALL call each of the six per-role fields once, joining
  any errors with `errors.Join`, so a failed concrete construction
  surfaces during `Init` rather than at first lazy access

### Requirement: Factory lazy-field errors are propagated, not swallowed

When a per-role lazy field's underlying composite construction returns
an error, the field SHALL return that error to the caller wrapped via
the canonical `git.New*Error` family (`git.NewRepoError`,
`git.NewWorktreeError`, `git.NewCommandError`) per `git-client`'s
"Routing table and error contract" requirement. The Factory SHALL NOT
log, format, mutate, or swallow the error; the command layer is the
single handling point that maps errors to exit codes and stderr. The
single-handling rule (logged OR returned, never both) SHALL apply.

#### Scenario: Per-role construction error reaches the command layer once

- **WHEN** `f.BranchReader()` is called and the underlying
  `git.NewClient()` returns a `*git.ExternalError`
- **THEN** `f.BranchReader()` SHALL return the wrapped error to the
  caller
- **AND** no log line SHALL be emitted on the Factory path
- **AND** the underlying `sync.OnceValues` cache SHALL retain its
  failure-or-success semantics for subsequent calls (per the
  `errors`-wrapped behavior of `sync.OnceValues`)

### Requirement: Role interface naming follows no-stuttering

Consumer-side role interfaces SHALL NOT repeat the `git` package qualifier in their identifiers (`BranchReader`, not `git.GitBranchReader`); sentinel errors returned by role methods SHALL use the `Err` prefix (`ErrBranchNotFound`, not `BranchNotFoundError`). The six role names declared in `internal/core/git.go` SHALL each match `^[A-Z][A-Za-z0-9]*Reader$` or `^[A-Z][A-Za-z0-9]*Writer$` (no `Git`/`Repository`/`Branch` prefix stutter).

#### Scenario: Role-interface naming follows conventions

- **WHEN** the package user enumerates the six role interfaces in `internal/core/git.go`
- **THEN** no role SHALL begin with `Git` followed by the role's noun (e.g., `GitBranchReader` is forbidden)
- **AND** no role SHALL carry the package qualifier inside its name
- **AND** any sentinel error returned by a role method SHALL match `^Err[A-Z][A-Za-z0-9]*$` (per `err-prefix-suffix`)
