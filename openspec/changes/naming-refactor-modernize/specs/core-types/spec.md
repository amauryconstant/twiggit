# Spec Delta: core-types

## ADDED Requirements

### Requirement: Data types in internal/core are named without Git* stutter or Info suffix

The `internal/core/` package SHALL expose data types whose names match
their semantic content. Types previously named `GitRepository`,
`GitDir`, `GitCommit`, `GitBranch`, `BranchInfo`, `WorktreeInfo`,
`RemoteInfo`, `CommitInfo` SHALL be renamed as follows:

| Old name        | New name      | File                |
|-----------------|---------------|---------------------|
| `GitRepository` | `Repository`  | `git_types.go`      |
| `GitDir`        | `RepoDir`     | `git_repo.go`       |
| `GitCommit`     | `Commit`      | `git_types.go`      |
| `GitBranch`     | `Branch`      | `git_types.go`      |
| `BranchInfo`    | `Branch`      | `git_types.go`      |
| `WorktreeInfo`  | `Worktree`    | `git_types.go`      |
| `RemoteInfo`    | `Remote`      | `git_types.go`      |
| `CommitInfo`    | `Commit`      | `git_types.go`      |
| `RepositoryStatus` | unchanged  | `git_types.go`      |

No `Git*` prefix SHALL appear in any data type declared in
`internal/core/` because the package name supplies the git context.
No `Info` suffix SHALL appear because the type itself IS the
information; a method returning `BranchInfo` (now `Branch`) becomes
`Branch(ctx) (Branch, error)` per `core-git` §"Role method names
follow the noun-only convention".

#### Scenario: Data types do not stutter Git

- **WHEN** `internal/core/git_types.go` and `internal/core/git_repo.go`
  are read
- **THEN** no exported type SHALL begin with `Git`
- **AND** `grep -E '^type Git' internal/core/*.go` returns zero matches

#### Scenario: Data types do not end with Info suffix

- **WHEN** `internal/core/git_types.go` is read
- **THEN** no exported type SHALL end with `Info`
- **AND** `grep -E 'Info\b' internal/core/git_types.go` returns zero
  type declarations

### Requirement: Data types and capability interfaces share a namespace without collision

`core.Repository`, `core.Branch`, `core.Worktree`, `core.Remote`,
`core.Commit`, `core.RepoDir` are **data types** (structs whose fields
describe git state). `core.RepositoryOpener`, `core.BranchReader`,
`core.WorktreeWriter`, `core.BranchWriter` are **capability
interfaces** (consumed via embedded promotion on `*git.Client`). The
two categories coexist in the same package; the spec uses "data type"
and "capability" prefixes in scenario names to disambiguate when the
context requires it.

#### Scenario: Capability interfaces stay separate from data types

- **WHEN** a consumer imports a `core.*` identifier
- **THEN** the identifier SHALL be either a struct (data type) or an
  interface (capability) — never both under the same name

#### Scenario: Sentinel compile-time checks distinguish both kinds

- **WHEN** `internal/git/client.go` is read
- **THEN** it SHALL contain one `var _ core.Role = (*git.Client)(nil)`
  declaration per capability interface
- **AND** `internal/core/` SHALL contain a separate compile-time
  fixture (function or struct literal) that exercises every data
  type's constructor or zero-value usage

### Requirement: Constructors and field accessors for renamed data types

Renamed data types SHALL keep their constructor-free, public-field
shape (the existing types are simple value objects). A type renamed
from `BranchInfo` to `Branch` keeps the same field set (`Name`,
`IsCurrent`, `Remote`, `Commit`, `Author`, `Date`) under identical
field names; only the type identifier changes.

#### Scenario: Renamed Branch preserves its field set

- **WHEN** `core.Branch` is read
- **THEN** the struct SHALL expose `Name`, `IsCurrent`, `Remote`,
  `Commit`, `Author`, `Date` — same as the prior `BranchInfo`
- **AND** no field SHALL be added or removed by the rename

#### Scenario: Renamed Worktree preserves its field set

- **WHEN** `core.Worktree` is read
- **THEN** the struct SHALL expose `Path`, `Branch`, `Commit`,
  `IsDetached`, `IsModified` — same as the prior `WorktreeInfo`
- **AND** no field SHALL be added or removed by the rename
