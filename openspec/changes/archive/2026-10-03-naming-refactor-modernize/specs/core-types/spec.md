# Spec Delta: core-types

## ADDED Requirements

### Requirement: Data types in internal/core are named without Git* stutter or Info suffix

The `internal/core/` package SHALL expose data types whose names match
their semantic content. Types previously named `GitRepository`,
`GitDir`, `GitCommit`, `GitBranch`, `BranchInfo`, `WorktreeInfo`,
`RemoteInfo`, `CommitInfo` SHALL be renamed as follows:

| Old name           | New name         |
|--------------------|------------------|
| `GitRepository`    | `Repository`     |
| `GitDir`           | `RepoDir`        |
| `GitCommit`        | `Commit`         |
| `GitBranch`        | `Branch`         |
| `BranchInfo`       | `Branch`         |
| `WorktreeInfo`     | `Worktree`       |
| `RemoteInfo`       | `Remote`         |
| `CommitInfo`       | `Commit`         |
| `RepositoryStatus` | (unchanged)      |

No `Git*` prefix SHALL appear in any data type declared in
`internal/core/` because the package name supplies the git context.
No `Info` suffix SHALL appear because the type itself IS the
information; a method returning `BranchInfo` (now `Branch`) becomes
`Branch(ctx) (Branch, error)` per `core-git`.

#### Scenario: Data types do not stutter Git

- **WHEN** the data-type declarations in the core package are read
  for `Git`-prefixed identifiers
- **THEN** no exported type SHALL begin with `Git`

#### Scenario: Data types do not end with Info suffix

- **WHEN** the data-type declarations in the core package are read
  for `Info`-suffixed identifiers
- **THEN** no exported type SHALL end with `Info`

### Requirement: Data types and capability interfaces share a namespace without collision

`core.Repository`, `core.Branch`, `core.Worktree`, `core.Remote`,
`core.Commit`, `core.RepoDir` are **data types** (structs whose fields
describe git state). `core.RepositoryOpener`, `core.BranchReader`,
`core.WorktreeWriter`, `core.BranchWriter` are **capability
interfaces** (consumed via embedded promotion on `*git.Client`). The
two categories SHALL coexist in the same package; the spec SHALL use
"data type" and "capability" prefixes in scenario names to
disambiguate when the context requires it.

#### Scenario: Capability interfaces stay separate from data types

- **WHEN** a consumer imports a `core.*` identifier
- **THEN** the identifier SHALL be either a struct (data type) or an
  interface (capability) — never both under the same name

#### Scenario: Sentinel compile-time checks distinguish both kinds

- **WHEN** the role-interface satisfaction surface on `*git.Client`
  is read
- **THEN** it SHALL contain one `var _ core.Role = (*git.Client)(nil)`
  declaration per capability interface
- **AND** the core package SHALL contain a separate compile-time
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

### Requirement: Compile-time fixture exercises every renamed data type

The core package SHALL contain a `_test.go` file
(`data_type_fixture_test.go`) that exercises the zero value or
constructor of every renamed data type. The fixture exists to
prove at compile time that every type in the rename table
(Requirement: Data types in internal/core are named without Git*
stutter or Info suffix) is reachable from production code; if a
type is removed or its constructor is broken, the test
compilation fails.

#### Scenario: Fixture compiles against every data type

- **WHEN** the `TestDataTypeFixture` test is invoked
- **THEN** the test SHALL reference every renamed data type by
  its zero value or constructor (`core.Repository{}`,
  `core.Branch{}`, `core.Worktree{}`, `core.Commit{}`,
  `core.Remote{}`, `core.RepoDir{}`)
- **AND** compilation SHALL succeed

#### Scenario: Fixture failure surfaces a missing constructor

- **WHEN** a renamed data type's constructor is removed or
  renamed without updating the fixture
- **THEN** the test build SHALL fail with a compile error
  identifying the missing symbol
