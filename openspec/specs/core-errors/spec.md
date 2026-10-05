# Capability: Errors

## Purpose

Defines the four `core.Error` subtypes that drive cmd-side error formatting and exit-code dispatch, replacing the 20-type taxonomy in `domain-typed-errors`. I/O-adapter-specific constructors live in their adapter package (`internal/git/errors.go`) as `git.ExternalError`-shaped wrappers; `output.FormatError` walks them via `errors.AsType[*core.OperationError]` to a `*core.OperationError`.

## Requirements

### Requirement: ValidationError carries Field, Value, Message, Suggestions

`core.ValidationError` SHALL be a struct with `Op string`, `Entity string`, `Field string`, `Value string`, `Message string`, and `Suggestions []string` fields. It SHALL implement `Unwrap() error` returning a hidden cause when constructed with `NewValidationErrorWithCause` (so error chains from wrapping with `%w` walk via `errors.Is`/`errors.AsType`); otherwise the chain terminates at the `ValidationError`. The `Error()` string SHALL be lowercase, contain no trailing punctuation, and SHALL NOT embed an emoji glyph.

#### Scenario: ValidationError renders lowercase without emoji

- **WHEN** a validator returns `&core.ValidationError{Field: "BranchName", Value: "feat..bad", Message: "must not contain '..'"}`
- **THEN** `err.Error()` returns a string matching the pattern `^validation failed .* lowercase$` with no emoji glyph and no trailing punctuation

#### Scenario: ValidationError Unwrap returns hidden cause when constructed with cause

- **WHEN** a caller constructs `verr := core.NewValidationErrorWithCause(op, entity, field, value, message, cause)`
- **THEN** `errors.Unwrap(verr)` SHALL return `cause`
- **AND** `errors.Is(verr, cause)` SHALL return `true` via the chain

### Requirement: NotFoundError carries Entity, Name; participates in errors.Is via per-entity sentinel

`core.NotFoundError` SHALL carry `Entity string` and `Name string`. It SHALL implement `Unwrap() error` returning `nil` and `Is(target error) bool` returning `true` only when `target` matches the sentinel that maps to `Entity` (entity-to-sentinel mapping per requirement: "NotFoundError.Is matches per-entity sentinel", not a blanket true for all 4 sentinels).

#### Scenario: errors.Is finds ErrWorktreeNotFound in the chain

- **WHEN** the error is `&core.NotFoundError{Entity: "worktree", Name: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` returns `true`

#### Scenario: errors.Is does not match unrelated sentinels

- **WHEN** the error is `&core.NotFoundError{Entity: "worktree", Name: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrProjectNotFound)` returns `false`

### Requirement: NotFoundError.Is matches per-entity sentinel

`core.NotFoundError.Is(target)` SHALL return `true` only when `target` is the sentinel whose entity matches `e.Entity`:

| Entity value | Matching sentinel |
|---|---|
| `"git.repo"` or `"repo"` | `core.ErrGitRepoNotFound` |
| `"worktree"` | `core.ErrWorktreeNotFound` |
| `"project"` | `core.ErrProjectNotFound` |
| `"resolution"` | `core.ErrResolutionNotFound` |

The dispatch SHALL NOT return `true` for any other sentinel; `errors.Is(err, core.ErrWorktreeNotFound)` succeeds only when `Entity == "worktree"`.

#### Scenario: Entity "worktree" matches only ErrWorktreeNotFound

- **WHEN** `err = &core.NotFoundError{Entity: "worktree"}`
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` is `true`
- **AND** `errors.Is(err, core.ErrProjectNotFound)` is `false`
- **AND** `errors.Is(err, core.ErrGitRepoNotFound)` is `false`
- **AND** `errors.Is(err, core.ErrResolutionNotFound)` is `false`

### Requirement: OperationError carries Op, Message, Cause, Suggestions; wraps via Unwrap

`core.OperationError` SHALL carry `Op string`, `Entity string`, `Field string`, `Message string`, `Cause error`, and `Suggestions []string`. It SHALL implement `Unwrap() error` returning `e.Cause`, satisfying `errors.Is`/`errors.AsType` walks across the wrapped chain. `Suggestions` carries actionable hints rendered by `output.FormatError` after the user-facing message. The `Op` field identifies the operation and SHALL NOT be shown to the user unless `TWIGGIT_DEBUG=1` is set. `OperationError.Is` SHALL match any of the four `not-found` sentinels when `Op` carries the corresponding prefix (`git.repository*`, `git.worktree*`, `project*`, `resolution*`), so `errors.Is(err, Sentinel)` succeeds regardless of which constructor produced the error.

#### Scenario: errors.As walks through OperationError

- **WHEN** the error chain is `core.NewOperationError("git.open", msg, io.EOF)`
- **THEN** `errors.Is(err, io.EOF)` returns `true` via the `Unwrap` chain

#### Scenario: errors.As reaches *core.OperationError from a git.ExternalError

- **WHEN** the error is `git.NewRepoError(path, msg, io.EOF)` which returns `*git.ExternalError` with an embedded `*core.OperationError`
- **THEN** `errors.AsType[*core.OperationError](err)` returns a non-nil result with `Op = "git.open"`

#### Scenario: Suggestions render after the message

- **WHEN** `OperationError.Suggestions = []string{"run 'twiggit init' first"}` and `FormatError` renders the error
- **THEN** the message appears on stderr followed by a hint line containing the suggestion

#### Scenario: OperationError.Is matches ErrProjectNotFound on project.* Op prefix

- **WHEN** `err = core.NewOperationError("project.create", msg, cause)`
- **THEN** `errors.Is(err, core.ErrProjectNotFound)` is `true`
- **AND** the chain still walks to `cause` via `Unwrap`

### Requirement: UsageError dispatches to ExitUsage (2); carries hidden cause field

`core.UsageError` SHALL carry `Message string` plus a hidden `cause error` field (no exported `Err` field). It SHALL implement `Unwrap() error` returning the stored cause. `cmdutil.ExitCodeFor` SHALL dispatch first via `errors.AsType[*core.UsageError](err)` and return `ExitUsage` (2) on match.

#### Scenario: UsageError exits 2

- **WHEN** the user runs `twiggit list -o xml` and the command returns `core.NewUsageError("unknown output format: xml")`
- **THEN** `main` propagates `ExitUsage` (2) to the OS

#### Scenario: UsageError Unwrap walks cause

- **WHEN** a caller invokes `core.NewUsageError(message, cause)`
- **THEN** `errors.Unwrap(uerr)` SHALL return `cause`
- **AND** `errors.Is(uerr, cause)` SHALL return `true` via the chain

### Requirement: I/O-adapter constructors live in internal/git/errors.go

I/O-adapter-specific constructors (`git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError`) live in `internal/git/errors.go` and return `*git.ExternalError` whose embedded core type is `*core.OperationError`. Callers SHALL construct I/O-adapter failures via the `git.NewRepoError` / `git.NewWorktreeError` / `git.NewCommandError` constructors; the `core` package exports no `core.NewGit*Error` form.

#### Scenario: git.NewRepoError walks to *core.OperationError

- **WHEN** the error is `git.NewRepoError(path, msg, io.EOF)`
- **THEN** `errors.AsType[*core.OperationError](err)` returns a non-nil result and `errors.Is(err, io.EOF)` returns `true` via the chain

#### Scenario: git.NewWorktreeError walks to *core.OperationError

- **WHEN** the error is `git.NewWorktreeError(name, msg, io.EOF)`
- **THEN** `errors.AsType[*core.OperationError](err)` returns a non-nil result whose `Op` identifies the worktree operation

#### Scenario: git.NewCommandError walks to *core.OperationError

- **WHEN** the error is `git.NewCommandError(args, msg, exitErr)`
- **THEN** `errors.AsType[*core.OperationError](err)` returns a non-nil result whose `Op` identifies the command and `Cause` is set to `exitErr`

#### Scenario: core does not export I/O constructors

- **WHEN** the `core` package's exported API is enumerated
- **THEN** it SHALL NOT contain any `core.NewGit*Error` constructor; callers route I/O failures through `git.NewRepoError`, `git.NewWorktreeError`, or `git.NewCommandError`

### Requirement: Error types follow naming and sentinel-zero-value rules

Sentinel variables SHALL use the `Err` prefix (`ErrGitRepoNotFound`, `ErrShellAlreadyInstalled`); error-typed constructors SHALL use the `New` prefix (`NewValidationError`, `NewOpValidationError`, `NewUsageError`); error types SHALL use the `Error` suffix. Enum-style sentinels (none currently; reserved for future discriminator fields) SHALL place an explicit `Unknown` variant at `iota` position 0. All `core.Err*` sentinels live in `internal/core/sentinels.go` (single canonical home); `internal/core/shell_errors.go` keeps only constructors, not sentinel declarations.

#### Scenario: Sentinel identifiers use Err prefix

- **WHEN** the package user enumerates `core.Err*` variables in `internal/core/sentinels.go`
- **THEN** every sentinel SHALL match the `^Err[A-Z][A-Za-z0-9]*$` pattern

#### Scenario: Error constructors use New prefix

- **WHEN** the package user enumerates exported error-constructor functions
- **THEN** every constructor SHALL match the `^New[A-Z][A-Za-z0-9]*Error?$` pattern

#### Scenario: Sentinels live in sentinels.go

- **WHEN** `internal/core/shell_errors.go` is read for sentinel declarations
- **THEN** it SHALL NOT redeclare any `core.Err*` variable; sentinel declarations live exclusively in `internal/core/sentinels.go`

### Requirement: ErrUncommittedChanges sentinel

`core.ErrUncommittedChanges` SHALL be a package-level sentinel that the uncommitted-changes guard surfaces via `errors.Is`. `core.NewUncommittedChangesError(worktreePath string)` SHALL return a `*core.OperationError` whose `Cause` is `ErrUncommittedChanges` so callers can detect the case via `errors.Is(err, core.ErrUncommittedChanges)`.

#### Scenario: errors.Is matches uncommitted-changes failure

- **WHEN** the runner detects uncommitted changes in a worktree and returns `core.NewUncommittedChangesError(path)`
- **THEN** `errors.Is(err, core.ErrUncommittedChanges)` SHALL return `true`
- **AND** `errors.AsType[*core.OperationError](err)` SHALL return a non-nil result with `Entity` identifying the worktree path