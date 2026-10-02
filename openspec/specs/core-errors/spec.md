# Capability: Errors

## Purpose

Defines the four `core.Error` subtypes that drive cmd-side error formatting and exit-code dispatch, replacing the 20-type taxonomy in `domain-typed-errors`. I/O-adapter-specific constructors live in their adapter package (`internal/git/errors.go`) as `git.ExternalError`-shaped wrappers; `output.FormatError` walks them via `errors.AsType[*core.OperationError]` to a `*core.OperationError`.

## Requirements

### Requirement: ValidationError carries Field, Value, Message, Suggestions

`core.ValidationError` SHALL be a struct with `Op string`, `Entity string`, `Field string`, `Value string`, `Message string`, and `Suggestions []string` fields. It SHALL implement `Unwrap() error` returning `nil`. The `Error()` string SHALL be lowercase, contain no trailing punctuation, and SHALL NOT embed an emoji glyph.

#### Scenario: ValidationError renders lowercase without emoji

- **WHEN** a validator returns `&core.ValidationError{Field: "BranchName", Value: "feat..bad", Message: "must not contain '..'"}`
- **THEN** `err.Error()` returns a string matching the pattern `^validation failed .* lowercase$` with no emoji glyph and no trailing punctuation

#### Scenario: ValidationError Unwrap returns nil

- **WHEN** callers invoke `errors.Unwrap(verr)`
- **THEN** the result SHALL be `nil`; `errors.Is` SHALL NOT find any sentinel in the chain

### Requirement: NotFoundError carries Entity, Name; participates in errors.Is via per-entity sentinel

`core.NotFoundError` SHALL carry `Entity string` and `Name string`. It SHALL implement `Unwrap() error` returning `nil` and `Is(target error) bool` returning `true` when `target` is one of the per-entity sentinels (`core.ErrGitRepoNotFound`, `core.ErrWorktreeNotFound`, `core.ErrProjectNotFound`, `core.ErrResolutionNotFound`).

#### Scenario: errors.Is finds ErrWorktreeNotFound in the chain

- **WHEN** the error is `&core.NotFoundError{Entity: "worktree", Name: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` returns `true`

#### Scenario: errors.Is does not match unrelated sentinels

- **WHEN** the error is `&core.NotFoundError{Entity: "worktree", Name: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrProjectNotFound)` returns `false`

### Requirement: OperationError carries Op, Message, Cause, Suggestions; wraps via Unwrap

`core.OperationError` SHALL carry `Op string`, `Entity string`, `Field string`, `Message string`, `Cause error`, and `Suggestions []string`. It SHALL implement `Unwrap() error` returning `e.Cause`, satisfying `errors.Is`/`errors.AsType` walks across the wrapped chain. `Suggestions` carries actionable hints rendered by `output.FormatError` after the user-facing message. The `Op` field identifies the operation and SHALL NOT be shown to the user unless `TWIGGIT_DEBUG=1` is set.

#### Scenario: errors.As walks through OperationError

- **WHEN** the error chain is `core.NewOperationError("git.open", msg, io.EOF)`
- **THEN** `errors.Is(err, io.EOF)` returns `true` via the `Unwrap` chain

#### Scenario: errors.As reaches *core.OperationError from a git.ExternalError

- **WHEN** the error is `git.NewRepoError(path, msg, io.EOF)` which returns `*git.ExternalError` with an embedded `*core.OperationError`
- **THEN** `errors.AsType[*core.OperationError](err)` returns a non-nil result with `Op = "git.open"`

#### Scenario: Suggestions render after the message

- **WHEN** `OperationError.Suggestions = []string{"run 'twiggit init' first"}` and `FormatError` renders the error
- **THEN** the message appears on stderr followed by a hint line containing the suggestion

### Requirement: UsageError dispatches to ExitUsage (2); carries no Err field

`core.UsageError` SHALL carry `Message string` only (no `Err` field). It SHALL implement `Unwrap() error` returning `nil`. `cmdutil.ExitCodeFor` SHALL dispatch first via `errors.AsType[*core.UsageError](err)` and return `ExitUsage` (2) on match.

#### Scenario: UsageError exits 2

- **WHEN** the user runs `twiggit list -o xml` and the command returns `core.NewUsageError("unknown output format: xml")`
- **THEN** `main` propagates `ExitUsage` (2) to the OS

#### Scenario: UsageError Unwrap returns nil

- **WHEN** callers invoke `errors.Unwrap(uerr)`
- **THEN** the result SHALL be `nil`

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

Sentinel variables SHALL use the `Err` prefix (`ErrGitRepoNotFound`, `ErrShellAlreadyInstalled`); error-typed constructors SHALL use the `New` prefix (`NewValidationError`, `NewOpValidationError`, `NewUsageError`); error types SHALL use the `Error` suffix. Enum-style sentinels (none currently; reserved for future discriminator fields) SHALL place an explicit `Unknown` variant at `iota` position 0.

#### Scenario: Sentinel identifiers use Err prefix

- **WHEN** the package user enumerates `core.Err*` variables in `internal/core/sentinels.go` and `internal/core/errors.go`
- **THEN** every sentinel SHALL match the `^Err[A-Z][A-Za-z0-9]*$` pattern

#### Scenario: Error constructors use New prefix

- **WHEN** the package user enumerates exported error-constructor functions
- **THEN** every constructor SHALL match `^New[A-Z][A-Za-z0-9]*Error?$`
