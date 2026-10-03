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
three context-aware methods: `RepositoryStatus(ctx context.Context, repoPath string) (core.RepositoryStatus, error)`, `Repository(ctx context.Context, repoPath string) (*core.Repository, error)`, and `Commit(ctx context.Context, repoPath, commitHash string) (*core.Commit, error)`. The role SHALL be defined consumer-side. Method names SHALL NOT carry a `Get` prefix; return types SHALL be the renamed data types from `core-types` (`Repository` not `GitRepository`, `Commit` not `CommitInfo`).

#### Scenario: RepositoryReader groups repository inspection methods

- **WHEN** a caller requires `RepositoryStatus`, `Repository`, and `Commit`
- **THEN** the role SHALL be `RepositoryReader` and SHALL NOT include
  branch listing, remote listing, or repository opening
- **AND** none of the three methods SHALL begin with `Get`

#### Scenario: Composite *git.Client satisfies RepositoryReader

- **WHEN** the composite `*git.Client` is type-asserted against
  `RepositoryReader`
- **THEN** the assertion SHALL succeed at compile time via
  `var _ core.RepositoryReader = (*git.Client)(nil)`

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

### Requirement: Role interface naming follows no-stuttering

Consumer-side role interfaces SHALL NOT repeat the `git` package qualifier in their identifiers (`BranchReader`, not `git.GitBranchReader`); sentinel errors returned by role methods SHALL use the `Err` prefix (`ErrBranchNotFound`, not `BranchNotFoundError`). The six role names declared in `internal/core/git.go` SHALL each match `^[A-Z][A-Za-z0-9]*Reader$` or `^[A-Z][A-Za-z0-9]*Writer$` (no `Git`/`Repository`/`Branch` prefix stutter).

#### Scenario: Role-interface naming follows conventions

- **WHEN** the package user enumerates the six role interfaces in `internal/core/git.go`
- **THEN** no role SHALL begin with `Git` followed by the role's noun (e.g., `GitBranchReader` is forbidden)
- **AND** no role SHALL carry the package qualifier inside its name
- **AND** any sentinel error returned by a role method SHALL match `^Err[A-Z][A-Za-z0-9]*$` (per `err-prefix-suffix`)

### Requirement: Role method names follow the noun-only convention

Every method on the six role interfaces SHALL be named after its return
type, minus the `Get` prefix. Read methods (`RepositoryStatus`,
`Repository`, `Commit`, `ListBranches`, `BranchExists`, `ListRemotes`,
`ValidateRepository`) SHALL be callable without a verb prefix. Mutating
methods (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`,
`PruneWorktrees`, `DeleteBranch`, `IsBranchMerged`) MAY keep imperative
verbs because their return is `error`, not a noun.

#### Scenario: No Get-prefixed methods on role interfaces

- **WHEN** the role interface declarations in the core package are
  read for method names
- **THEN** no method SHALL begin with `Get`
- **AND** the declaration surface SHALL NOT carry a `//nolint`
  directive suppressing the rule
