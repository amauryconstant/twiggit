# Spec Delta: cli-factory

## MODIFIED Requirements

### Requirement: Factory exposes lazy function fields

The `cmdutil.Factory` type SHALL expose the lazy `func() (T, error)`
fields: `Config`, `GitClient`, `Logger`. The `IOStreams` SHALL be
wired eagerly via the factory constructor (not lazy). The Factory
SHALL NOT expose a `Formatter` or `FormatError` lazy field;
formatters live in `internal/output/` and are constructed per call by
the cmd layer. The fields SHALL be assigned at construction time and
invoked at first use. Role interfaces (`RepositoryOpener`,
`BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`,
`BranchWriter`) SHALL NOT appear as Factory fields; callers obtain
role-narrowed access via type assertion on the value returned by
`Factory.GitClient()`.

#### Scenario: Lazy field is nil before first call

- **WHEN** a command reads `f.Config` before invoking it
- **THEN** the field returns the zero value and the command MUST call
  the function form (`f.Config()`) to initialize it

#### Scenario: Lazy field caches its result across calls

- **WHEN** a command calls `f.Config()` twice during one binary
  invocation
- **THEN** the second call SHALL return the same `*core.Config` pointer
  as the first call

#### Scenario: Factory does not expose per-role fields

- **WHEN** `internal/cmdutil/factory.go` is read
- **THEN** no field SHALL be named `RepoOpener`, `BranchReader`,
  `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, or
  `BranchWriter`

## ADDED Requirements

### Requirement: Role interfaces obtained via type assertion on GitClient

When a command needs a role-narrowed view of the git client, the
caller SHALL obtain it by calling `f.GitClient()` and assigning the
result to a local role-typed variable. `*git.Client` satisfies every
role declared in `core-git` through embedded promotion, so no
separate Factory field is required. The type assertion introduces one
new identifier per call site and zero new lazy-cache surface.

#### Scenario: Caller narrows to BranchReader via assignment

- **WHEN** a command requires `core.BranchReader`
- **THEN** the call site SHALL read `client, err := f.GitClient(); if
  err != nil { ... }; br := core.BranchReader(client)` or equivalent
- **AND** the command SHALL NOT call `f.BranchReader()`

#### Scenario: Underlying composite is unchanged

- **WHEN** `f.GitClient()` is called twice from the same command
- **THEN** both calls SHALL return the same `*git.Client` pointer
- **AND** any role-typed local variables derived from those pointers
  SHALL target the same composite instance

### Requirement: Factory Init touches every lazy field exactly once

`Factory.Init()` SHALL call each non-role lazy field (`Config`,
`GitClient`, `Logger`) exactly once and SHALL join any errors with
`errors.Join`. `Init` SHALL NOT iterate role interfaces because no role
fields exist, SHALL NOT touch `IOStreams` (wired eagerly in the
constructor per the MODIFIED requirement above), and SHALL NOT call
`Formatter` or `FormatError` (those live in `internal/output/` and are
constructed per call by the cmd layer). Errors SHALL surface during
`Init` rather than at first lazy access.

#### Scenario: Init surfaces composite construction failures

- **WHEN** `Factory.Init()` is invoked and `f.GitClient()` returns a
  `*core.OperationError`
- **THEN** the joined error SHALL contain that `OperationError`
- **AND** `main.go` SHALL receive the joined error before `Execute`
  runs

#### Scenario: Init joins multiple lazy-field failures

- **WHEN** `Factory.Init()` is invoked and both `f.Config()` and
  `f.GitClient()` fail
- **THEN** the returned error SHALL match both underlying errors via
  `errors.Is`
- **AND** the joined chain SHALL preserve their order
