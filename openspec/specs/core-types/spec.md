# Capability: Types

## Purpose

Defines the value-object pattern and validation pipeline for the `internal/core/` package, making the rule that every domain type is always-valid post-construction. The pipeline's `ValidateAll` uses `errors.Join` to preserve the full error chain; `output.FormatError` walks each joined error independently.

## Requirements

### Requirement: Value objects expose always-valid construction

Every value object (`core.NewBranchName`, `core.NewProjectName`, `core.NewWorktreePath`) SHALL return `(T, error)` from its constructor. The returned value SHALL be safe to use without further validation; if the input fails the rules, the constructor SHALL return a `*core.ValidationError`. The validation pipeline exposes `core.ValidateBranchName(name) Result[bool]`, `core.ValidateProjectName(name) Result[bool]`, and `core.ValidateShellType(s) Result[bool]` for callers that prefer pipeline composition over constructor invocation.

#### Scenario: Valid branch name constructs successfully

- **WHEN** the caller invokes `r := core.ValidateBranchName("feat/foo")`
- **THEN** `r.IsSuccess()` returns `true` and `r.Value()` is `true`

#### Scenario: Invalid branch name returns ValidationError

- **WHEN** the caller invokes `r := core.ValidateBranchName("feat..bad")`
- **THEN** `r.IsError()` returns `true` and the wrapped error is a `*core.ValidationError` carrying `Field="BranchName"`, `Value="feat..bad"`, and a `Message` naming the rule violated

### Requirement: Pipeline[T] composes Validator[T] with Validate (fail-fast) or ValidateAll (collect-all)

The `core.Pipeline[T]` type SHALL compose zero or more `core.Validator[T]` functions. `Pipeline.Validate(t)` SHALL run each validator in order and return the first error (fail-fast); `Pipeline.ValidateAll(t)` SHALL run all validators and MUST use `errors.Join(...)` to compose every failure into one error that preserves the full chain. `FormatError` SHALL walk the joined chain via `errors.As` and render each failure as a separate validation hint. `ValidateAll` SHALL NOT aggregate failures into a single `*core.ValidationError` with a `Suggestions` slice; each failure SHALL remain an independent error in the join so `errors.Is` and `errors.As` can inspect them individually.

#### Scenario: Validate stops at first failure

- **WHEN** a pipeline of two validators runs `pipeline.Validate(input)` and the first fails
- **THEN** the second validator SHALL NOT run and the returned error SHALL be the first validator's error

#### Scenario: ValidateAll collects every failure via errors.Join

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and validators 1 and 3 fail
- **THEN** `errors.Is(err, validator1Err)` returns `true` and `errors.Is(err, validator3Err)` returns `true` (both preserved in the chain)
- **AND** `errors.Is(err, validator2Err)` returns `false` (the passing validator is not in the chain)

#### Scenario: ValidateAll wraps each failure independently

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and validators 1 and 3 fail
- **THEN** the returned error SHALL be the `errors.Join` composition and SHALL NOT be a single aggregated `*core.ValidationError`
- **AND** `errors.AsType[*core.ValidationError](err)` SHALL return a non-nil result twice (once per independent failure), confirming each failure is its own error in the chain

#### Scenario: ValidateAll with no failures returns nil

- **WHEN** a pipeline of three validators runs `pipeline.ValidateAll(input)` and all validators pass
- **THEN** the returned error SHALL be `nil` (an empty `errors.Join` result is normalized to nil)

#### Scenario: ValidateAll with one failure preserves errors.Is

- **WHEN** a pipeline of one validator runs `pipeline.ValidateAll(input)` and the validator returns `validatorErr`
- **THEN** `errors.Is(err, validatorErr)` returns `true` via the join chain

### Requirement: Core types are immutable post-construction

Once constructed, a value object SHALL expose read-only fields (or accessor methods). Mutation methods SHALL NOT exist on core types.

#### Scenario: No mutation methods on BranchName

- **WHEN** the package user searches `core.BranchName` for setter or mutation methods
- **THEN** no method SHALL match `Set`, `Mutate`, `Assign`, `Replace`, or `Update` patterns

### Requirement: Result[T] wraps a value or an error

The `core.Result[T]` type SHALL carry either a `Value T` and a `Success bool`, or a non-nil `Err error`. Constructors `core.NewResult(value T)` and `core.NewErrResult(err error)` SHALL produce success and failure variants respectively. `Result.IsError()` and `Result.IsSuccess()` SHALL dispatch on the `Success` flag without unwrapping. `Result.Value()` SHALL return the stored value when `IsSuccess()` is true; `Result.Err()` SHALL return the stored error when `IsError()` is true.

#### Scenario: Success result exposes Value

- **WHEN** the caller invokes `r := core.NewResult("feat/foo")`
- **THEN** `r.IsSuccess()` returns `true` and `r.Value()` returns `"feat/foo"`

#### Scenario: Error result exposes Err

- **WHEN** the caller invokes `r := core.NewErrResult(core.NewValidationError(...))`
- **THEN** `r.IsError()` returns `true` and `errors.Is(r.Err(), &core.ValidationError{})` matches

### Requirement: Value objects use New prefix per branch

Exported constructors for value objects SHALL use the `New` prefix (`core.NewBranchName`, `core.NewProjectName`, `core.NewWorktreePath`). Sentinels SHALL use the `Err` prefix; error types SHALL use the `Error` suffix. Enum-typed value objects (`ShellType`, `ContextType`, `PathType`) SHALL place an explicit `Unknown`/`Invalid` variant at `iota` position 0.

#### Scenario: Constructors use New prefix

- **WHEN** the package user enumerates exported constructors in `internal/core/`
- **THEN** every constructor SHALL match `^New[A-Z][A-Za-z0-9]*$`

#### Scenario: Enum types declare zero-value sentinel

- **WHEN** the package user reads `core.ShellType`, `core.ContextType`, or `core.PathType`
- **THEN** each type SHALL declare an `Unknown` or `Invalid` constant at `iota` position 0

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
