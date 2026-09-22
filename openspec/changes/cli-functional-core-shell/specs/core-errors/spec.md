# Spec Delta

## Purpose

Defines the four `core.Error` subtypes that drive cmd-side error formatting and exit-code dispatch, replacing the 20-type taxonomy in `domain-typed-errors`. I/O-adapter-specific constructors live in their adapter package (`internal/git/errors.go`) as `git.ExternalError`-shaped wrappers; `output.FormatError` walks them via `errors.As` to a `*core.OperationError`.

## ADDED Requirements

### Requirement: ValidationError carries Field, Value, Message, Suggestions

`core.ValidationError` SHALL be a struct with `Field string`, `Value string`, `Message string`, and `Suggestions []string` fields. It SHALL implement `Unwrap() error` returning `nil`. The `Error()` string SHALL be lowercase, contain no trailing punctuation, and SHALL NOT embed an emoji glyph.

#### Scenario: ValidationError renders lowercase without emoji
- **WHEN** a validator returns `&core.ValidationError{Field: "BranchName", Value: "feat..bad", Message: "must not contain '..'"}`
- **THEN** `err.Error()` returns `"invalid branch name \"feat..bad\": must not contain '..'"` (lowercase, no emoji, no trailing punctuation)

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

`core.OperationError` SHALL carry `Op string`, `Message string`, `Cause error`, and `Suggestions []string`. It SHALL implement `Unwrap() error` returning the `Cause` field. `Suggestions` carries actionable hints rendered by `output.FormatError` after the user-facing message. The `Op` field identifies the operation and SHALL NOT be shown to the user unless `TWIGGIT_DEBUG=1` is set.

#### Scenario: errors.As walks through OperationError
- **WHEN** the error chain is `core.NewOperationError("git.open", msg, io.EOF)`
- **THEN** `errors.Is(err, io.EOF)` returns `true` via the `Unwrap` chain

#### Scenario: errors.As reaches *core.OperationError from a git.ExternalError
- **WHEN** the error is `git.NewRepoError(path, msg, io.EOF)` which returns `*git.ExternalError` with an embedded `*core.OperationError`
- **THEN** `errors.As(err, &*core.OperationError{})` returns `true` with `Op = "git.open"`

#### Scenario: Suggestions render after the message
- **WHEN** `OperationError.Suggestions = []string{"run 'twiggit init' first"}` and `FormatError` renders the error
- **THEN** the message appears on stderr followed by a hint line containing the suggestion

### Requirement: UsageError dispatches to ExitUsage (2); carries no Err field

`core.UsageError` SHALL carry `Message string` only (no `Err` field). It SHALL implement `Unwrap() error` returning `nil`. `cmdutil.ExitCodeFor` SHALL dispatch first via `errors.As(err, &*core.UsageError{})` and return `ExitUsage` (2) on match.

#### Scenario: UsageError exits 2
- **WHEN** the user runs `twiggit list -o xml` and the command returns `core.NewUsageError("unknown output format: xml")`
- **THEN** `main` propagates `ExitUsage` (2) to the OS

#### Scenario: UsageError Unwrap returns nil
- **WHEN** callers invoke `errors.Unwrap(uerr)`
- **THEN** the result SHALL be `nil`

### Requirement: I/O-adapter constructors live in internal/git/errors.go

I/O-adapter-specific constructors (`git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError`) live in `internal/git/errors.go` and return `*git.ExternalError` whose embedded core type is `*core.OperationError`. The `core` package SHALL NOT export `core.NewGit*Error` constructors; the previous `core.NewGitRepositoryError` form is removed.

#### Scenario: git.NewRepoError walks to *core.OperationError
- **WHEN** the error is `git.NewRepoError(path, msg, io.EOF)`
- **THEN** `errors.As(err, &*core.OperationError{})` returns `true` and `errors.Is(err, io.EOF)` returns `true` via the chain

#### Scenario: core does not export I/O constructors
- **WHEN** the `core` package's exported API is enumerated
- **THEN** it SHALL NOT contain `NewGitRepositoryError`, `NewGitWorktreeError`, or `NewGitCommandError` constructors
