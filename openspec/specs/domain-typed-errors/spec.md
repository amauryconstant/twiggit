# Capability: Typed Errors

## Purpose

The canonical owner of every `domain.*Error` struct, constructor,
and exit-code mapping. Other specs reference error types by name; this
spec defines their fields, behaviour, and `Unwrap()` chains.

## Requirements

### Requirement: Error type taxonomy

The domain layer SHALL define the following error types with the given
constructors. All types SHALL implement `Unwrap() error` for
`errors.As()` chain support.

| Type | Constructor | `IsNotFound()` |
|---|---|---|
| `ValidationError` | `NewValidationError(request, field, value, message)` | — |
| `GitRepositoryError` | `NewGitRepositoryError(path, message, cause)` | yes |
| `GitWorktreeError` | `NewGitWorktreeError(worktreePath, branchName, message, cause)` | yes |
| `GitCommandError` | `NewGitCommandError(cmd, args, exitCode, stdout, stderr, msg, cause)` | — |
| `ConfigError` | `NewConfigError(path, message, cause)` | — |
| `ContextDetectionError` | `NewContextDetectionError(path, message, cause)` | — |
| `ServiceError` | `NewServiceError(service, operation, message, cause)` | — |
| `WorktreeServiceError` | `NewWorktreeServiceError(worktreePath, branchName, op, msg, cause)` | yes |
| `ProjectServiceError` | `NewProjectServiceError(projectName, projectPath, op, msg, cause)` | — |
| `NavigationServiceError` | `NewNavigationServiceError(target, ctx, op, msg, cause)` | — |
| `ShellError` | `NewShellError(code, shellType, context)` / `NewShellErrorWithCause(..., cause)` | — |
| `ResolutionError` | `NewResolutionError(target, ctx, msg, suggestions, cause)` | — |
| `ConflictError` | `NewConflictError(resource, identifier, operation, message, cause)` | — |



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: ValidationError contract

`ValidationError` SHALL support immutable builder methods:

- `WithSuggestions([]string) *ValidationError`
- `WithContext(string) *ValidationError`

Getters SHALL be: `Field()`, `Value()`, `Message()`, `Request()`,
`Suggestions()`, `Context()`.

#### Scenario: ValidationError returned directly

- **WHEN** a validation failure occurs at any layer
- **THEN** the layer SHALL return the `ValidationError` directly
- **AND** SHALL NOT wrap it via `fmt.Errorf("...: %w", err)`

### Requirement: NotFound detection

`GitRepositoryError`, `GitWorktreeError`, and `WorktreeServiceError`
SHALL implement an `IsNotFound() bool` method that returns true when
the message contains a "not found", "does not exist", or "no such file
or directory" substring (case-insensitive).

#### Scenario: NotFound dispatch

- **WHEN** the cmd-side error formatter sees an error whose
  `IsNotFound()` returns true (via `errors.As` walk)
- **THEN** the system SHALL exit with code 6 (`ExitCodeNotFound`)
- **AND** the formatter SHALL emit a not-found-style message

### Requirement: Cause-chain support

Every domain error SHALL implement `Unwrap()` returning its `Cause`
field. `errors.As()` and `errors.Is()` SHALL walk the chain to find
matching types.

#### Scenario: errors.As across wrap

- **WHEN** a service wraps a `GitRepositoryError` inside a
  `WorktreeServiceError`
- **THEN** `errors.As(err, &*domain.GitRepositoryError{})` SHALL succeed

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
| 6 | `ExitCodeNotFound` | Any error whose `IsNotFound()` returns true |

The cmd layer (`cli-error-formatting`) SHALL NOT redefine this table.
`GetExitCodeForError` SHALL dispatch by `errors.As` walk: try
specific types first (`ValidationError`, `*ServiceError`,
`GitRepositoryError`, etc.), then fall back to `IsNotFound`, then to
`ExitCodeError` (1). The dispatch algorithm and formatter-registration
order live in `cli-error-formatting`; this spec is the source of truth
for the codes and triggers.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: No comments-on-error wrapping

Validation failures SHALL be returned directly (already self-describing).
All other errors SHALL be wrapped using the appropriate `domain.New*`
constructor at the originating layer (service / infrastructure). Bare
`fmt.Errorf("...: %w", err)` SHALL only appear when no domain error
type applies.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Nil-context safety

Service methods that accept a `*domain.Context` SHALL validate that
the pointer is non-nil before dereferencing, returning a
`domain.ValidationError` on field `context` with message "context must
not be nil" otherwise.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Empty-path safety

Service methods that resolve a path SHALL validate that the input is
non-empty before invoking git operations, returning a
`domain.ValidationError` on the relevant field otherwise.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
