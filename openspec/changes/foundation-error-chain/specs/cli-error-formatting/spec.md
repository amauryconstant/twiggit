# Spec Delta: Error Formatting

## MODIFIED Requirements

### Requirement: Exit code mapping

The system SHALL map error categories to exit codes via the
`GetExitCodeForError` helper. The mapping SHALL be exactly:

| Exit code | Constant | Meaning |
|---|---|---|
| 0 | `ExitCodeSuccess` | Success |
| 1 | `ExitCodeError` | General unclassified error or panic |
| 2 | `ExitCodeUsage` | Usage error (invalid command syntax) |
| 3 | `ExitCodeConfig` | Configuration error |
| 4 | `ExitCodeGit` | Git operation error |
| 5 | `ExitCodeValidation` | Validation error |
| 6 | `ExitCodeNotFound` | Any error matching a per-resource NotFound sentinel |

The cmd layer SHALL NOT redefine this mapping. See `domain-typed-errors`
for the canonical definitions. `GetExitCodeForError` SHALL dispatch
first via `errors.Is` against the four NotFound sentinels, then via
typed `errors.As` for the remaining categories.

#### Scenario: Validation error → exit 5

- **WHEN** a `domain.ValidationError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeValidation` (5)

#### Scenario: Git error → exit 4

- **WHEN** a `domain.GitRepositoryError`, `domain.GitWorktreeError`, or
  `domain.GitCommandError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeGit` (4)

#### Scenario: NotFound → exit 6

- **WHEN** an error matching `domain.ErrGitRepoNotFound`,
  `domain.ErrWorktreeNotFound`, `domain.ErrProjectNotFound`, or
  `domain.ErrResolutionNotFound` via `errors.Is` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeNotFound` (6)

### Requirement: Type-matched dispatch

The system SHALL dispatch formatters via an explicit matcher strategy
using `errors.As()` rather than reflection. Registration order SHALL
place more specific matchers before generic ones. Formatters SHALL
extract typed errors via a single `errors.As` check, plus an
`asType[T error]` generic helper that returns the typed value and a
boolean indicating whether the chain reached the type. The registry
SHALL register formatters for the six shell subtypes
(`ShellAlreadyInstalledError`, `ShellNotInstalledError`,
`ShellInvalidTypeError`, `ShellInferenceError`,
`ShellDetectionError`, `ShellWrapperError`, `ShellConfigError`),
`NavigationServiceError`, `ResolutionError`, and the remaining typed
domain errors listed in `domain-typed-errors`.

#### Scenario: Specific matcher wins

- **WHEN** `errors.As(err, &*domain.WorktreeServiceError{})` is true
- **AND** `errors.As(err, &*domain.ServiceError{})` is also true
- **THEN** the WorktreeServiceError formatter SHALL win because it is
  registered first

#### Scenario: asType helper guards nil dereference

- **WHEN** a formatter receives an error that does not contain a
  matching typed error
- **THEN** the formatter SHALL fall back to the generic message via a
  returned `false` from `asType[T]` rather than dereferencing a nil
  pointer

### Requirement: Actionable hints

The system SHALL attach hints to common error cases, pointing the user
at remediation. For errors matching a per-resource NotFound sentinel
via `errors.Is`, the system SHALL attach a resource-specific hint as
documented in the table below.

| Sentinel | Hint |
|---|---|
| `domain.ErrProjectNotFound` | "Use 'twiggit list --all' to see available projects" |
| `domain.ErrWorktreeNotFound` | "Use 'twiggit list' to see available worktrees" |
| `domain.ErrResolutionNotFound` | "Use 'twiggit list' to see available navigation targets" |
| `domain.ErrGitRepoNotFound` | "Verify the repository path" |

For non-NotFound errors the system SHALL retain a generic hint
("Check your configuration and try again" or equivalent).

#### Scenario: Project not found

- **WHEN** an error matching `domain.ErrProjectNotFound` via `errors.Is`
  is rendered
- **THEN** system SHALL append the project-specific hint from the hint
  table

#### Scenario: Worktree not found

- **WHEN** an error matching `domain.ErrWorktreeNotFound` via
  `errors.Is` is rendered
- **THEN** system SHALL append the worktree-specific hint from the hint
  table

#### Scenario: Quiei mode strips hints

- **WHEN** the formatter is configured for quiet mode
- **THEN** the hint lines SHALL be omitted from the rendered output
