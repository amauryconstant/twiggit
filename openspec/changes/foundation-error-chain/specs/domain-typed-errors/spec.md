# Spec Delta: Typed Errors

## MODIFIED Requirements

### Requirement: Error type taxonomy

The domain layer SHALL define the following error types with the given
constructors. All types that wrap a cause SHALL implement both
`Unwrap() error` returning the `Err error` field and `Is(target error)
bool` participating in `errors.Is` walks against the appropriate
per-resource NotFound sentinel described in the Sentinel catalog
requirement. `ValidationError` SHALL implement `Unwrap() error`
returning `nil` (terminal). All field names for the wrapped cause SHALL
be `Err` (not `Cause`).

| Type | Constructor | Per-resource NotFound sentinel |
|---|---|---|
| `ValidationError` | `NewValidationError(request, field, value, message)` | — |
| `GitRepositoryError` | `NewGitRepositoryError(path, message, err)` | `ErrGitRepoNotFound` |
| `GitWorktreeError` | `NewGitWorktreeError(worktreePath, branchName, message, err)` | `ErrWorktreeNotFound` |
| `GitCommandError` | `NewGitCommandError(cmd, args, exitCode, stdout, stderr, msg, err)` | — |
| `ConfigError` | `NewConfigError(path, message, err)` | — |
| `ContextDetectionError` | `NewContextDetectionError(path, message, err)` | — |
| `ServiceError` | `NewServiceError(service, operation, message, err)` | — |
| `WorktreeServiceError` | `NewWorktreeServiceError(worktreePath, branchName, op, msg, err)` | `ErrWorktreeNotFound` |
| `ProjectServiceError` | `NewProjectServiceError(projectName, projectPath, op, msg, err)` | `ErrProjectNotFound` |
| `NavigationServiceError` | `NewNavigationServiceError(target, ctx, op, msg, err)` | `ErrResolutionNotFound` |
| `ResolutionError` | `NewResolutionError(target, ctx, msg, suggestions, err)` | `ErrResolutionNotFound` |
| `ConflictError` | `NewConflictError(resource, identifier, operation, message, err)` | — |
| `ShellAlreadyInstalledError` | `NewShellAlreadyInstalledError(shellType, context, err)` | — |
| `ShellNotInstalledError` | `NewShellNotInstalledError(shellType, context, err)` | — |
| `ShellInvalidTypeError` | `NewShellInvalidTypeError(shellType, context, err)` | — |
| `ShellInferenceError` | `NewShellInferenceError(shellType, context, err)` | — |
| `ShellDetectionError` | `NewShellDetectionError(context, err)` | — |
| `ShellWrapperError` | `NewShellWrapperError(shellType, op, context, err)` | — |
| `ShellConfigError` | `NewShellConfigError(path, context, err)` | — |

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Cause-chain support

Every domain error that wraps another error SHALL implement `Unwrap()`
returning its `Err` field. Errors SHALL participate in
`errors.Is` and `errors.As` walks through that chain.
Wrappers whose wrapping layer has a per-resource NotFound sentinel
SHALL additionally implement `Is(target error) bool` matching only that
sentinel so the sentinel is visible through wrapping.

#### Scenario: errors.As across wrap

- **WHEN** a service wraps a `GitRepositoryError` inside a
  `WorktreeServiceError`
- **THEN** `errors.As(err, &*domain.GitRepositoryError{})` SHALL succeed

#### Scenario: errors.Is through wrap

- **WHEN** an error of type `WorktreeServiceError` carries a populated
  `Err` field
- **AND** the `Err` chain reaches `ErrWorktreeNotFound`
- **THEN** `errors.Is(err, domain.ErrWorktreeNotFound)` SHALL return
  true regardless of whether `Err` is nil

### Requirement: Exit-code mapping (canonical)

The canonical exit-code mapping SHALL be exactly:

| Code | Constant | Trigger |
|---|---|---|
| 0 | `ExitCodeSuccess` | Clean exit |
| 1 | `ExitCodeError` | Unclassified error or recovered panic |
| 2 | `ExitCodeUsage` | Cobra usage error (invalid syntax/args) |
| 3 | `ExitCodeConfig` | `ConfigError` |
| 4 | `ExitCodeGit` | `GitRepositoryError`, `GitWorktreeError`, `GitCommandError` |
| 5 | `ExitCodeValidation` | `ValidationError` |
| 6 | `ExitCodeNotFound` | Any error matching a per-resource NotFound sentinel |

The cmd layer (`cli-error-formatting`) SHALL NOT redefine this table.
`GetExitCodeForError` SHALL dispatch by `errors.Is` walk against the
four NotFound sentinels first, then by typed `errors.As` walk, then to
`ExitCodeError` (1). The dispatch algorithm and hint discriminator
live in `cli-error-formatting`; this spec is the source of truth for
the codes and triggers.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

## ADDED Requirements

### Requirement: Sentinel catalog

The domain layer SHALL export the following sentinel errors as package
variables of type `error`. Each sentinel's `Error()` message SHALL be
`"domain: <resource> <state>"`. Callers SHALL identify these sentinels
exclusively through `errors.Is` and SHALL NOT compare sentinel values
via `==` or string equality.

| Sentinel | Message |
|---|---|
| `ErrGitRepoNotFound` | `"domain: git repository not found"` |
| `ErrWorktreeNotFound` | `"domain: worktree not found"` |
| `ErrProjectNotFound` | `"domain: project not found"` |
| `ErrResolutionNotFound` | `"domain: resolution target not found"` |
| `ErrShellAlreadyInstalled` | `"domain: shell wrapper already installed"` |
| `ErrShellNotInstalled` | `"domain: shell wrapper not installed"` |
| `ErrInvalidShellType` | `"domain: invalid shell type"` |
| `ErrShellInferenceFailed` | `"domain: could not infer shell type"` |
| `ErrShellDetectionFailed` | `"domain: shell detection failed"` |
| `ErrWrapperGeneration` | `"domain: wrapper generation failed"` |
| `ErrWrapperInstallation` | `"domain: wrapper installation failed"` |
| `ErrConfigFileNotFound` | `"domain: config file not found"` |

#### Scenario: Sentinel catalog is exported and stable

- **WHEN** a caller imports the domain package
- **THEN** the sentinel identifiers and their messages SHALL match the
  table above exactly
- **AND** no sentinel SHALL be unexported, renamed, or repurposed
  without a capability-level spec change

### Requirement: Is method participation

Each error type whose wrapping layer maps to a NotFound sentinel SHALL
implement an `Is(target error) bool` method that returns true if and
only if `target` equals the type's per-resource sentinel. The method
SHALL return false for any other target, including unrelated domain
sentinels. Implementations SHALL NOT rely on string equality or
substring matching against `Error()`.

#### Scenario: WorktreeServiceError matches its sentinel

- **WHEN** a `WorktreeServiceError` is constructed with a populated
  `Err` chain reaching `ErrWorktreeNotFound`
- **THEN** `errors.Is(worktreeErr, domain.ErrWorktreeNotFound)` SHALL
  return true
- **AND** `errors.Is(worktreeErr, domain.ErrProjectNotFound)` SHALL
  return false

#### Scenario: Shell subtype matches its sentinel

- **WHEN** a `ShellAlreadyInstalledError` is constructed with an
  `Err` chain that is nil
- **THEN** `errors.Is(shellErr, domain.ErrShellAlreadyInstalled)` SHALL
  return true
- **AND** `errors.Is(shellErr, domain.ErrShellNotInstalled)` SHALL
  return false

### Requirement: Shell subtypes share a common base

The domain layer SHALL define six shell error subtypes listed in the
Error type taxonomy requirement. Each subtype SHALL embed a private
`shellErrorBase` value carrying `ShellType`, `Context`, and `Err`
fields, SHALL implement `Unwrap() error` returning the `Err` field,
and SHALL implement `Is(target error) bool` returning true exactly
when the target is the subtype's own sentinel. No `domain.ShellError`
interface SHALL be introduced.

#### Scenario: ShellAlreadyInstalledError format

- **WHEN** a `ShellAlreadyInstalledError` is rendered via its
  `Error()` method
- **THEN** the result SHALL identify the shell type and the
  already-installed state
- **AND** SHALL NOT include the sentinel string code or any emoji or
  decoration

#### Scenario: No ShellError interface

- **WHEN** the domain package is compiled
- **THEN** the package SHALL NOT contain an exported interface named
  `ShellError`

### Requirement: ValidationError terminal chain

`ValidationError` SHALL NOT carry an `Err` field. It SHALL implement
`Unwrap() error` returning `nil`. The `Error()` string SHALL be
lowercase, contain no emoji, and SHALL NOT embed a `💡` glyph.

#### Scenario: Unwrap returns nil

- **WHEN** `Unwrap()` is called on any `ValidationError` instance
- **THEN** the result SHALL be `nil`

#### Scenario: Plain text rendering

- **WHEN** `ValidationError.Error()` is called
- **THEN** the returned string SHALL contain no `💡` character
- **AND** SHALL contain no trailing punctuation
