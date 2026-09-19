# Capability: Error Formatting

## Purpose

User-friendly error rendering on stderr with actionable hints, exit-code
dispatch, and panic recovery. This spec is the cmd-side surface;
the error-type taxonomy itself is owned by `domain-typed-errors`.

## Requirements

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
| 6 | `ExitCodeNotFound` | Resource not found |

The cmd layer SHALL NOT redefine this mapping. See `domain-typed-errors`
for the canonical definitions.

#### Scenario: Validation error → exit 5

- **WHEN** a `domain.ValidationError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeValidation` (5)

#### Scenario: Git error → exit 4

- **WHEN** a `domain.GitRepositoryError`, `domain.GitWorktreeError`, or
  `domain.GitCommandError` reaches the formatter
- **THEN** system SHALL exit with code `ExitCodeGit` (4)

#### Scenario: NotFound → exit 6

- **WHEN** an error whose `IsNotFound() bool` returns true reaches the
  formatter
- **THEN** system SHALL exit with code `ExitCodeNotFound` (6)

### Requirement: User-friendly messages

The system SHALL format error messages for users, omitting internal
operation names and including actionable hints.

#### Scenario: Format ValidationError

- **WHEN** `domain.ValidationError` is rendered
- **THEN** output SHALL be "Error: <message>"
- **AND** SHALL include any `Suggestions` lines

#### Scenario: Format WorktreeServiceError

- **WHEN** `domain.WorktreeServiceError` is rendered
- **THEN** output SHALL be "Error: <message>" with the worktree path
  and operation removed from the user-visible text

### Requirement: Type-matched dispatch

The system SHALL dispatch formatters via an explicit matcher strategy
using `errors.As()` rather than reflection. Registration order SHALL
place more specific matchers before generic ones.

#### Scenario: Specific matcher wins

- **WHEN** `errors.As(err, &*domain.WorktreeServiceError{})` is true
- **AND** `errors.As(err, &*domain.ServiceError{})` is also true
- **THEN** the WorktreeServiceError formatter SHALL win because it is
  registered first

### Requirement: Actionable hints

The system SHALL attach hints to common error cases, pointing the user
at remediation.

#### Scenario: Project not found

- **WHEN** a project-not-found error is rendered
- **THEN** system SHALL append a hint: "Use 'twiggit list' to see
  available projects"

### Requirement: Panic recovery

The system SHALL recover from panics in `main.go` via `defer`/`recover`,
display "Internal error: <panic value>" to stderr, and exit with
code 1.

#### Scenario: Recover from panic

- **WHEN** a panic occurs anywhere in the command tree
- **THEN** system SHALL write "Internal error: ..." to stderr
- **AND** SHALL exit with code 1
- **AND** SHALL NOT print a stack trace unless `TWIGGIT_DEBUG=1`

#### Scenario: Debug mode shows stack

- **WHEN** `TWIGGIT_DEBUG=1` is set and a panic occurs
- **THEN** system SHALL print the full stack trace in addition to the
  panic value

### Requirement: Debug mode details

The system SHALL honor `TWIGGIT_DEBUG=1` to surface internal operation
names and error context normally hidden from users.

#### Scenario: Debug shows internals

- **WHEN** `TWIGGIT_DEBUG=1` and a `ServiceError` is rendered
- **THEN** the rendered output SHALL include the service name and
  operation name (normally hidden)
