# Spec Delta

## Purpose

Defines the four `core.Error` subtypes that drive cmd-side error formatting and exit-code dispatch, replacing the 20-type taxonomy in `domain-typed-errors`.

## ADDED Requirements

### Requirement: ValidationError carries Field, Value, Message, Suggestions

`core.ValidationError` SHALL be a struct with `Field string`, `Value string`, `Message string`, and `Suggestions []string` fields. It SHALL implement `Unwrap() error` returning `nil`. The `Error()` string SHALL be lowercase, contain no emoji, and SHALL NOT embed a `💡` glyph.

#### Scenario: ValidationError renders lowercase without emoji
- **WHEN** a validator returns `&core.ValidationError{Field: "BranchName", Value: "feat..bad", Message: "must not contain '..'"}`
- **THEN** `err.Error()` returns `"branch name: must not contain '..' (got \"feat..bad\")"` (or equivalent lowercase message) with no emoji

#### Scenario: ValidationError Unwrap returns nil
- **WHEN** callers invoke `errors.Unwrap(verr)`
- **THEN** the result SHALL be `nil`; `errors.Is` SHALL NOT find any sentinel in the chain

### Requirement: NotFoundError carries Resource, Identifier; participates in errors.Is via per-resource sentinel

`core.NotFoundError` SHALL carry `Resource string` and `Identifier string`. It SHALL implement `Is(target error) bool` returning `true` when `target` is one of the per-resource sentinels (`ErrGitRepoNotFound`, `ErrWorktreeNotFound`, `ErrProjectNotFound`, `ErrResolutionNotFound`).

#### Scenario: errors.Is finds ErrWorktreeNotFound in the chain
- **WHEN** the error is `&core.NotFoundError{Resource: "worktree", Identifier: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` returns `true`

#### Scenario: errors.Is does not match unrelated sentinels
- **WHEN** the error is `&core.NotFoundError{Resource: "worktree", Identifier: "feat/foo"}`
- **THEN** `errors.Is(err, core.ErrProjectNotFound)` returns `false`

### Requirement: OperationError carries Op, Message, Cause; wraps via %w

`core.OperationError` SHALL carry `Op string`, `Message string`, and `Cause error`. It SHALL implement `Unwrap() error` returning the `Cause` field. Constructors (`core.NewGitRepositoryError`, `core.NewGitWorktreeError`, etc.) SHALL wrap the cause via `%w`.

#### Scenario: errors.As walks through OperationError
- **WHEN** the error chain is `core.NewGitRepositoryError(path, msg, io.EOF)`
- **THEN** `errors.Is(err, io.EOF)` returns `true` via the `Unwrap` chain

#### Scenario: errors.As reaches *core.OperationError
- **WHEN** the error is `core.NewGitRepositoryError(...)`
- **THEN** `errors.As(err, &*core.OperationError{})` returns `true`

### Requirement: UsageError dispatches to ExitUsage (2)

`core.UsageError` SHALL carry `Message string` and `Err error`. `cmdutil.ExitCodeFor` SHALL dispatch first via `errors.As(err, &*core.UsageError{})` and return `ExitUsage` (2) on match.

#### Scenario: UsageError exits 2
- **WHEN** the user runs `twiggit list -o xml` and the command returns `core.NewUsageError("unknown output format: xml")`
- **THEN** `main` propagates `ExitUsage` (2) to the OS
