# Capability: Factory

## Purpose

Defines the Factory composition pattern with lazy initialization for CLI command dependencies, giving each command a single injection point for config, git client, IOStreams, and output formatters without an interface explosion.

## Requirements

### Requirement: Factory exposes lazy function fields

The `cmdutil.Factory` type SHALL expose the lazy `func() (T, error)` fields: `Config`, `GitClient`, `Logger`. The `IOStreams` SHALL be wired eagerly via the factory constructor (not lazy). The Factory SHALL NOT expose a `Formatter` or `FormatError` lazy field; formatters live in `internal/output/` and are constructed per call by the cmd layer. The fields SHALL be assigned at construction time and invoked at first use. Role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) SHALL NOT appear as Factory fields; callers obtain role-narrowed access via type assertion on the value returned by `Factory.GitClient()`.

#### Scenario: Lazy field is nil before first call

- **WHEN** a command reads `f.Config` before invoking it
- **THEN** the field returns the zero value and the command MUST call the function form (`f.Config()`) to initialize it

#### Scenario: Lazy field caches its result across calls

- **WHEN** a command calls `f.Config()` twice during one binary invocation
- **THEN** the second call SHALL return the same `*core.Config` pointer as the first call

#### Scenario: Factory does not expose per-role fields

- **WHEN** `internal/cmdutil/factory.go` is read
- **THEN** no field SHALL be named `RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, or `BranchWriter`

### Requirement: Config is cached via sync.OnceValues

`Factory.Config` SHALL use `sync.OnceValues(func() (*core.Config, error))` to guarantee the config file is read and parsed exactly once per binary invocation, even when multiple commands access it. `Factory.GitClient` and `Factory.Logger` SHALL also use `sync.OnceValues` for the same caching guarantee.

#### Scenario: Config file is read at most once

- **WHEN** three commands each call `f.Config()` in sequence
- **THEN** the underlying koanf loader SHALL read `$XDG_CONFIG_HOME/twiggit/config.toml` exactly once

#### Scenario: Config failure surfaces on first access

- **WHEN** the config file is invalid TOML
- **THEN** the first call to `f.Config()` SHALL return a `*core.OperationError` wrapping the parse error; subsequent calls SHALL return the same error without re-reading the file

### Requirement: Factory is the only place where infrastructure is instantiated

Every `os/exec`, `git.PlainOpen`, file read, and XDG path lookup SHALL happen inside a Factory lazy function. `cmd/*.go` files SHALL NOT construct infrastructure types directly.

#### Scenario: cmd/ files do not import os/exec directly

- **WHEN** `golangci-lint run` enforces the Tier 2 depguard
- **THEN** `cmd/*.go` SHALL NOT import `os/exec` directly; the `CommandExecutor` is the only path

#### Scenario: Factory is the wiring seam for tests

- **WHEN** a test invokes `NewCmdList(testFactory, runF)` with a Factory whose fields are pre-set to test doubles
- **THEN** the command runs against the test doubles without spawning real git subprocesses

### Requirement: Role interfaces obtained via type assertion on GitClient

When a command needs a role-narrowed view of the git client, the caller SHALL obtain it by calling `f.GitClient()` and narrowing the returned value to `*git.Client` (the composite role carrier). `*git.Client` satisfies every role declared in `core-git` through embedded promotion, so no separate Factory field is required.

The Factory's `GitClient` field type is `func() (cmdutil.Client, error)` where `cmdutil.Client` is a package-level alias for `any`. This is a deliberate depguard-driven indirection: it lets `internal/cmdutil` declare the Factory surface without importing `internal/git`, while preserving the embedded-promotion contract that makes role methods reachable through `*git.Client`. The narrow is performed at the cmd boundary (in `cmd/setup.go`) where the depguard rule no longer applies. Any future change that lets `internal/cmdutil` import `internal/git` may collapse `cmdutil.Client` to `*git.Client` directly.

#### Scenario: Caller narrows to *git.Client via type assertion

- **WHEN** a command requires any role interface on the git client
- **THEN** the call site SHALL reach the role by first calling `f.GitClient()` to obtain the cached client and then asserting the result to `*git.Client` (the composite carrier)
- **AND** the command SHALL NOT call any per-role Factory accessor (no `f.BranchReader()`, no `f.RepoOpener()`, etc.)
- **AND** the assertion SHALL fail loudly (return an error, never a silent zero-value) on any type mismatch

#### Scenario: Underlying composite is unchanged

- **WHEN** `f.GitClient()` is called twice from the same command
- **THEN** both calls SHALL return the same `*git.Client` pointer (after the cmd-side assertion)
- **AND** any role methods invoked on that pointer SHALL target the same composite instance

### Requirement: Factory Init touches every lazy field exactly once

`Factory.Init()` SHALL call each non-role lazy field (`Config`, `GitClient`, `Logger`) exactly once and SHALL join any errors with `errors.Join`. `Init` SHALL NOT iterate role interfaces because no role fields exist, SHALL NOT touch `IOStreams` (wired eagerly in the constructor per the MODIFIED requirement above), and SHALL NOT call `Formatter` or `FormatError` (those live in `internal/output/` and are constructed per call by the cmd layer). Errors SHALL surface during `Init` rather than at first lazy access.

#### Scenario: Init surfaces composite construction failures

- **WHEN** `Factory.Init()` is invoked and `f.GitClient()` returns a `*core.OperationError`
- **THEN** the joined error SHALL contain that `OperationError`
- **AND** `main.go` SHALL receive the joined error before `Execute` runs

#### Scenario: Init joins multiple lazy-field failures

- **WHEN** `Factory.Init()` is invoked and both `f.Config()` and `f.GitClient()` fail
- **THEN** the returned error SHALL match both underlying errors via `errors.Is`
- **AND** the joined chain SHALL preserve their order
