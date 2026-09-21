# Spec Delta: Error Formatting

## REMOVED Requirements

### Requirement: Exit code mapping

The system SHALL map error categories to exit codes via the
`GetExitCodeForError` helper. The mapping SHALL be exactly:

| Exit code | Constant | Meaning |
|---|---|---|
| 0 | `ExitCodeSuccess` | Success |
| 1 | `ExitCodeError` | General unclassified error, runtime failure, or recovered panic |
| 2 | `ExitCodeUsage` | Usage error (invalid command syntax) |
| 3 | `ExitCodeConfig` | Configuration error |
| 4 | `ExitCodeGit` | Git operation error |
| 5 | `ExitCodeValidation` | Validation error |
| 6 | `ExitCodeNotFound` | Not-found error |

The cmd layer SHALL NOT define additional exit-code constants. See
`domain-typed-errors` for the canonical definitions; per-resource
NotFound categories are distinguished in user-facing output by the
Actionable hints requirement, not by exit code. `GetExitCodeForError`
SHALL dispatch first via `errors.As` against typed usage-error sentinels
(`*cobra.FlagError`, `cobra.ErrSubCommandRequired`,
`*pflag.Error{Code: pflag.ErrRequired}`), returning `ExitCodeUsage`;
otherwise returning `ExitCodeError` for any non-nil error and
`ExitCodeSuccess` for nil.

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
- **AND** the formatter SHALL append the resource-specific hint per
  the Actionable hints requirement

#### Scenario: Success → exit 0

- **WHEN** the command returns no error
- **THEN** system SHALL exit with code `ExitCodeSuccess` (0)

#### Scenario: Cobra usage error → exit 2

- **WHEN** an error matching `*cobra.FlagError`,
  `cobra.ErrSubCommandRequired`, or
  `*pflag.Error{Code: pflag.ErrRequired}` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeUsage` (2)

## MODIFIED Requirements

### Requirement: Type-matched dispatch

The system SHALL dispatch formatters via an explicit matcher strategy
using `errors.As()` rather than reflection. Registration order SHALL
place more specific matchers before generic ones. Formatters SHALL
extract typed errors via a single `errors.As` check, plus an
`asType[T error]` generic helper that returns the typed value and a
boolean indicating whether the chain reached the type. The registry
SHALL register formatters for the seven shell subtypes
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

#### Scenario: Quiet mode strips hints

- **WHEN** the formatter is configured for quiet mode
- **THEN** the hint lines SHALL be omitted from the rendered output
- **AND** quiet-mode behavior SHALL match `cli-quiet-mode`

## ADDED Requirements

### Requirement: Exit code mapping (3-code canonical)

The system SHALL map error categories to exit codes via the
`GetExitCodeForError` helper. The mapping SHALL be exactly:

| Exit code | Constant | Meaning |
|---|---|---|
| 0 | `ExitCodeSuccess` | Success |
| 1 | `ExitCodeError` | General unclassified error, runtime failure, or recovered panic |
| 2 | `ExitCodeUsage` | Usage error (invalid command syntax) |

The cmd layer SHALL NOT define additional exit-code constants. See
`domain-typed-errors` for the canonical definitions; per-resource
NotFound categories are distinguished in user-facing output by the
Actionable hints requirement, not by exit code. `GetExitCodeForError`
SHALL dispatch first via `errors.As(err, &*domain.UsageError{})` (or
`errors.Is(err, domain.ErrUsageFlag)` against typed usage errors),
returning `ExitCodeUsage`; otherwise returning `ExitCodeError` for any
non-nil error and `ExitCodeSuccess` for nil.

#### Scenario: Validation error → ExitCodeError

- **WHEN** a `domain.ValidationError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeError`

#### Scenario: Git error → ExitCodeError

- **WHEN** a `domain.GitRepositoryError`, `domain.GitWorktreeError`, or
  `domain.GitCommandError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeError`

#### Scenario: NotFound → ExitCodeError

- **WHEN** an error matching `domain.ErrGitRepoNotFound`,
  `domain.ErrWorktreeNotFound`, `domain.ErrProjectNotFound`, or
  `domain.ErrResolutionNotFound` via `errors.Is` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeError`
- **AND** the formatter SHALL append the resource-specific hint per
  the Actionable hints requirement

#### Scenario: Success → ExitCodeSuccess

- **WHEN** the command returns no error
- **THEN** system SHALL exit with code `ExitCodeSuccess`

#### Scenario: Cobra usage error → ExitCodeUsage

- **WHEN** an error matching `*domain.UsageError` via `errors.As`
  (or `domain.ErrUsageFlag` via `errors.Is`) reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeUsage`
