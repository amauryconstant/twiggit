# domain-typed-errors Specification

## Purpose

Canonical taxonomy of domain-layer error types, per-resource NotFound
sentinels, and the unwrap/is contracts that make errors participate
in `errors.Is` and `errors.As` walks so the cmd-side formatter
(`cli-error-formatting`) can dispatch exit codes and resource-specific
hints without coupling to wrapper internals.

## Requirements

### Requirement: Error type taxonomy

The `core` package SHALL define exactly four error types: `ValidationError`, `NotFoundError`, `OperationError`, `UsageError`. The previous 20-type taxonomy (the seven shell subtypes, `NavigationServiceError`, `ResolutionError`, `ConfigError`, `ServiceError`, `WorktreeServiceError`, `ProjectServiceError`, `ContextDetectionError`, `GitRepositoryError`, `GitWorktreeError`, `GitCommandError`, `ConflictError`) is removed. I/O-adapter-specific constructors live in their adapter package (`internal/git/errors.go`) as `git.ExternalError`-shaped wrappers; `output.FormatError` walks them via `errors.As` to a `*core.OperationError`. Constructors uniformly use `core.NewXxxError` for the four core types only.

| Core type | Constructor | Carries | Unwrap |
|---|---|---|---|
| `ValidationError` | `core.NewValidationError(field, value, message)` | `Field`, `Value`, `Message`, `Suggestions []string` | `nil` |
| `NotFoundError` | `core.NewNotFoundError(entity, name)` | `Entity`, `Name` | `nil` |
| `OperationError` | `core.NewOperationError(op, message, cause)` | `Op`, `Message`, `Cause`, `Suggestions []string` | `Cause` |
| `UsageError` | `core.NewUsageError(message)` | `Message` | `nil` |

`OperationError` carries `Suggestions` (added in this change). `UsageError` carries no `Err` field. The previous builder methods `WithSuggestions` / `WithContext` and getters `Request()` / `Detail()` / `Context()` are removed.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: I/O error walks to *core.OperationError

- **WHEN** the error chain is `git.NewRepoError(path, msg, io.EOF)` which returns `*git.ExternalError` with an embedded `*core.OperationError`
- **THEN** `errors.As(err, &*core.OperationError{})` returns `true`

#### Scenario: ValidationError carries no Err field

- **WHEN** a validator returns `&core.ValidationError{Field: "BranchName", Value: "feat..bad", Message: "must not contain '..'"}`
- **THEN** the struct has no `Err` field and `errors.Unwrap(verr)` returns `nil`

#### Scenario: UsageError carries no Err field

- **WHEN** a usage-error site returns `&core.UsageError{Message: "unknown --output value"}`
- **THEN** the struct has no `Err` field and `errors.Unwrap(uerr)` returns `nil`

### Requirement: Is method participation

`core.NotFoundError` SHALL implement an `Is(target error) bool` method that returns true if and only if `target` equals one of the four NotFound sentinels. The method SHALL return false for any other target, including unrelated core sentinels. Implementations SHALL NOT rely on string equality or substring matching against `Error()`. The previous `WorktreeServiceError`, `ShellAlreadyInstalledError`, `UsageError` `Is()` methods (each matching its own per-type sentinel) are removed; only `NotFoundError` retains `Is()`.

#### Scenario: WorktreeServiceError matches its sentinel

- **WHEN** a `core.NotFoundError` with `Entity = "worktree"` and `Name = "feat/missing"` is constructed (the worktree case replaces the previous WorktreeServiceError.Is method)
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` SHALL return true
- **AND** `errors.Is(err, core.ErrProjectNotFound)` SHALL return false

#### Scenario: Shell subtype matches its sentinel

- **WHEN** shell-detect errors use `*core.OperationError` with `Op = "shell.detect"` or `Op = "shell.probe"` (the previous `ShellAlreadyInstalledError.Is` per-type sentinel is removed; all shell-detect errors walk through the generic OperationError)
- **THEN** `errors.Is(err, core.ErrGitRepoNotFound)` SHALL return false
- **AND** `errors.As(err, &*core.OperationError{})` SHALL return true with the appropriate `Op`

#### Scenario: UsageError matches its sentinel

- **WHEN** a `core.UsageError` is constructed (no `Err` field; replaces the previous UsageError.Is method)
- **THEN** `errors.As(err, &*core.UsageError{})` SHALL return true (dispatch is via `errors.As`, not via a sentinel `errors.Is`)
- **AND** `errors.As(err, &*core.NotFoundError{})` SHALL return false

#### Scenario: OperationError does not match NotFound sentinels

- **WHEN** the error is `core.NewOperationError("git.open", msg, io.EOF)`
- **THEN** `errors.Is(err, core.ErrGitRepoNotFound)` SHALL return false

### Requirement: Shell subtypes share a common base

The previous seven shell error subtypes (`ShellAlreadyInstalledError`, `ShellNotInstalledError`, `ShellInvalidTypeError`, `ShellInferenceError`, `ShellDetectionError`, `ShellWrapperError`, `ShellConfigError`) are removed. The `core.ShellType` type lives in `internal/core/shell_detect.go` (pure derivation). Filesystem probing of shell config files (formerly `os.Stat` calls in the shell-detect path) moves to `internal/git/shell_detect.go`. The `composeWrapper` renderer moves to `internal/output/wrapper.go` as `output.ComposeWrapper`. No `core.ShellError` interface SHALL be introduced.

#### Scenario: ShellAlreadyInstalledError format

- **WHEN** shell-detect errors render via their `Error()` method (the previous `ShellAlreadyInstalledError` type is removed; shell errors are now `*core.OperationError` with `Op = "shell.detect"` or `Op = "shell.probe"`)
- **THEN** the rendered text SHALL identify the shell type and the failure state
- **AND** SHALL NOT include any emoji or decoration

#### Scenario: No ShellError interface

- **WHEN** the `core` package is compiled
- **THEN** the package SHALL NOT contain any of the seven shell error subtypes
- **AND** SHALL NOT contain a `ShellError` interface

### Requirement: ValidationError terminal chain

`core.ValidationError` SHALL NOT carry an `Err` field. It SHALL implement `Unwrap() error` returning `nil`. The `Error()` string SHALL be lowercase, contain no emoji, and SHALL NOT embed a `💡` glyph. The previous immutable builder methods `WithSuggestions([]string) *ValidationError` and `WithContext(string) *ValidationError` are removed; `Suggestions` is a struct field set at construction. The previous getters `Request()`, `Detail()`, and the deprecated `Context()` are removed; callers use `Field()`, `Value()`, `Message()`, and `Suggestions()` only.

#### Scenario: Detail() returns the contextual explanation

- **WHEN** a caller invokes `e.Detail()` on a `*core.ValidationError`
- **THEN** the call SHALL fail to compile (Detail() is removed; callers migrate to `e.Message()` or `e.Suggestions()`)

#### Scenario: Context() getter no longer exists

- **WHEN** the `core` package is compiled
- **THEN** `(*ValidationError).Context()` SHALL NOT be defined
- **AND** any caller that referenced `e.Context()` SHALL fail to compile until migrated to `e.Suggestions()` (the previous `Context()` rename to `Detail()` is reversed: the getter is removed entirely)

#### Scenario: Unwrap returns nil

- **WHEN** `Unwrap()` is called on any `core.ValidationError` instance
- **THEN** the result SHALL be `nil`

#### Scenario: Plain text rendering

- **WHEN** `core.ValidationError.Error()` is called
- **THEN** the returned string SHALL contain no `💡` character
- **AND** SHALL contain no trailing punctuation
- **AND** SHALL be entirely lowercase

### Requirement: NotFound detection

Not-found detection SHALL be expressed as `errors.Is(err, core.ErrXNotFound)` against the per-resource sentinel associated with `*core.NotFoundError`. The `NotFoundError` type SHALL implement `Is(target error) bool` returning `true` if and only if `target` is one of: `ErrGitRepoNotFound`, `ErrWorktreeNotFound`, `ErrProjectNotFound`, `ErrResolutionNotFound`. The previous substring-based `IsNotFound() bool` method on `GitRepositoryError`, `GitWorktreeError`, and `WorktreeServiceError` is removed. The four NotFound sentinels all map to `ExitError` (1); per-resource distinction is exposed via the per-resource NotFound hints requirement in `cli-error-formatting`, not via per-resource exit codes.

#### Scenario: NotFound dispatch

- **WHEN** the cmd-side error formatter sees an error whose `errors.Is(err, core.ErrWorktreeNotFound)` returns true
- **THEN** the system SHALL exit with code `ExitError` (1)
- **AND** the formatter SHALL append the worktree-specific hint per the `cli-error-formatting` Actionable hints requirement

### Requirement: Cause-chain support

Every core error that wraps another error SHALL implement `Unwrap() error` returning the cause. `OperationError` wraps via the `Cause` field; `errors.Is` and `errors.As` walks SHALL reach the underlying cause through that chain. `ValidationError`, `NotFoundError`, and `UsageError` implement `Unwrap() error { return nil }` (terminal errors).

#### Scenario: errors.As across wrap

- **WHEN** a service wraps a `git.NewRepoError(...)` (returning `*git.ExternalError` with embedded `*core.OperationError`) inside a `core.OperationError`
- **THEN** `errors.As(err, &*core.OperationError{})` SHALL succeed

#### Scenario: errors.Is through wrap

- **WHEN** an error of type `core.NotFoundError` carries `Entity = "worktree"` and `Name = "feat/missing"`
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` SHALL return true

### Requirement: Exit-code mapping (canonical)

The canonical exit-code mapping SHALL be exactly:

| Code | Constant | Trigger |
|---|---|---|
| 0 | `ExitOK` | Clean exit |
| 1 | `ExitError` | Unclassified error, runtime failure, or recovered panic |
| 2 | `ExitUsage` | Cobra usage error (invalid syntax or args) typed via `errors.As` against `*core.UsageError` |

The cmd layer (`cli-error-formatting`) SHALL NOT define additional exit-code constants. `cmdutil.ExitCodeFor` SHALL dispatch first via `errors.As` against `*core.UsageError`, returning `ExitUsage`; otherwise returning `ExitError` (1) for any non-nil error and `ExitOK` (0) for nil. Per-resource discrimination happens at the formatter hint layer (`cli-error-formatting` Actionable hints requirement), not via per-resource exit codes. The previous `domain.ErrUsageFlag` sentinel is removed; the previous exit-code dispatch helper is replaced by `cmdutil.ExitCodeFor`.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Sentinel catalog

The `core` package SHALL export the following sentinel errors as package variables of type `error`:
| Sentinel | Message |
|---|---|
| `ErrGitRepoNotFound` | `"core: git repository not found"` |
| `ErrWorktreeNotFound` | `"core: worktree not found"` |
| `ErrProjectNotFound` | `"core: project not found"` |
| `ErrResolutionNotFound` | `"core: resolution target not found"` |
| `ErrShellAlreadyInstalled` | `"core: shell wrapper already installed"` |
| `ErrShellNotInstalled` | `"core: shell wrapper not installed"` |
| `ErrInvalidShellType` | `"core: invalid shell type"` |
| `ErrInferenceFailed` | `"core: could not infer shell type"` |
| `ErrDetectionFailed` | `"core: shell detection failed"` |

Each sentinel's `Error()` message SHALL be `"core: <resource> <state>"`. Callers SHALL identify these sentinels exclusively through `errors.Is`. Each shell sentinel SHALL be returned as the `Cause` (or wrapped via `*core.OperationError.Cause`) of a `*core.OperationError` whose `Op` field starts with the prefix `shell.` (specifically `shell.already_installed`, `shell.not_installed`, `shell.invalid_type`, `shell.inference`, `shell.detection`); callers SHALL detect the sentinel with `errors.Is(err, core.ErrShellAlreadyInstalled)` (and the corresponding sentinel for the other four cases) — string comparison against `OperationError.Op` SHALL NOT be used by callers.

#### Scenario: Sentinel catalog is exported and stable

- **WHEN** a caller imports the `core` package
- **THEN** the sentinel identifiers and their messages SHALL match the table above exactly
- **AND** no sentinel SHALL be unexported, renamed, or repurposed without a capability-level spec change

#### Scenario: shell sentinel match via errors.Is

- **WHEN** a `*core.OperationError` is returned with `Op = "shell.already_installed"` and `Cause = core.ErrShellAlreadyInstalled`
- **THEN** `errors.Is(err, core.ErrShellAlreadyInstalled)` returns `true` and `err.Error()` contains `"shell wrapper already installed"`

#### Scenario: shell sentinel match across wrap layers

- **WHEN** the shell-installation layer returns `fmt.Errorf("install wrapper for bash: %w", opErr)` where `opErr.Cause = core.ErrShellAlreadyInstalled`
- **THEN** `errors.Is(err, core.ErrShellAlreadyInstalled)` returns `true` at every layer of the chain

#### Scenario: not-found sentinels remain matched via errors.Is

- **WHEN** a caller encounters a `*core.NotFoundError{Entity: "worktree", Name: "feat/x"}` returned through any wrapper chain
- **THEN** `errors.Is(err, core.ErrWorktreeNotFound)` returns `true`

#### Scenario: not-found sentinels excluded from shell-sentinel matching

- **WHEN** a caller matches `errors.Is(err, core.ErrShellAlreadyInstalled)` against a non-shell error
- **THEN** the result is `false`
