# Spec Delta: Domain Typed Errors

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
| `UsageError` | `NewUsageError(message, err)` or `UsageWrap(err)` | `ErrUsageFlag` |

Constructors and sentinels retain their existing names. Constructors
uniformly use the `NewXxxError` style and sentinels uniformly use the
`ErrXxx` style per the `golang-naming` skill rule that reserves the
`Error` suffix for error types and the `Err` prefix for sentinel
error variables.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: ValidationError terminal chain

`ValidationError` SHALL support immutable builder methods:

- `WithSuggestions([]string) *ValidationError`
- `WithContext(string) *ValidationError`

Getters SHALL be: `Field()`, `Value()`, `Message()`, `Request()`,
`Suggestions()`, and `Detail()` (the prior `Context()` getter is
renamed to `Detail()` to avoid collision with the `domain.Context`
type at call sites that pass both).

#### Scenario: Detail() returns the contextual explanation

- **WHEN** a caller invokes `e.Detail()` on a `*ValidationError`
- **THEN** it SHALL return the same string that the prior
  `e.Context()` getter returned
- **AND** it SHALL NOT collide with the `domain.Context` type when
  the caller writes `e.Detail()` next to a `*domain.Context`
  argument in the same scope

#### Scenario: Context() getter no longer exists

- **WHEN** the domain package is compiled
- **THEN** `(*ValidationError).Context()` SHALL NOT be defined
- **AND** any caller that referenced `e.Context()` SHALL fail to
  compile until migrated to `e.Detail()`

#### Scenario: Unwrap returns nil

- **WHEN** `Unwrap()` is called on any `ValidationError` instance
- **THEN** the result SHALL be `nil`

#### Scenario: Plain text rendering

- **WHEN** `ValidationError.Error()` is called
- **THEN** the returned string SHALL contain no `💡` character
- **AND** SHALL contain no trailing punctuation
