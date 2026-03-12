## ADDED Requirements

### Requirement: User-friendly error messages

The system SHALL display error messages that are actionable for users without exposing internal implementation details.

Service error messages SHALL NOT include internal operation names like "WorktreeService.CreateWorktree".

Service error messages SHALL provide context about what operation failed and which resource was involved.

#### Scenario: Worktree creation failure
- **WHEN** a worktree creation fails due to a project not existing
- **THEN** the error message SHALL display "could not create worktree for '&lt;project&gt;'"
- **AND** the message SHALL NOT include "WorktreeService" or "CreateWorktree"

#### Scenario: Project discovery failure
- **WHEN** a project discovery fails
- **THEN** the error message SHALL display "could not find project '&lt;project&gt;'"
- **AND** the message SHALL NOT include internal service or operation names

#### Scenario: Navigation failure
- **WHEN** a navigation operation fails
- **THEN** the error message SHALL describe the navigation target and failure reason
- **AND** the message SHALL NOT include internal operation names

### Requirement: Granular exit codes

The system SHALL return specific exit codes for different error categories to enable reliable scripting.

Exit codes SHALL be defined as follows:
- 0: Success
- 1: General error (unclassified)
- 2: Usage error (incorrect command syntax)
- 3: Configuration error
- 4: Git operation error
- 5: Validation error (input validation, not usage)
- 6: Resource not found

#### Scenario: Configuration error exit code
- **WHEN** a configuration error occurs (e.g., invalid config file)
- **THEN** the process SHALL exit with code 3

#### Scenario: Git operation error exit code
- **WHEN** a git operation fails (e.g., worktree add fails)
- **THEN** the process SHALL exit with code 4

#### Scenario: Validation error exit code
- **WHEN** input validation fails (e.g., invalid branch name format)
- **THEN** the process SHALL exit with code 5

#### Scenario: Resource not found exit code
- **WHEN** a requested resource does not exist (e.g., project or worktree not found)
- **THEN** the process SHALL exit with code 6

#### Scenario: Usage error exit code unchanged
- **WHEN** command syntax is incorrect (e.g., missing required argument)
- **THEN** the process SHALL exit with code 2

### Requirement: Panic recovery

The system SHALL catch unexpected panics and display a user-friendly error message.

When a panic occurs, the system SHALL display "Internal error: &lt;panic value&gt;" to stderr.

When `TWIGGIT_DEBUG` environment variable is set, the system SHALL display the full stack trace after the panic message.

The system SHALL exit with code 1 when a panic is recovered.

#### Scenario: Panic with debug disabled
- **WHEN** an unexpected panic occurs
- **AND** `TWIGGIT_DEBUG` is not set
- **THEN** stderr SHALL contain "Internal error:"
- **AND** stderr SHALL NOT contain a Go stack trace
- **AND** the process SHALL exit with code 1

#### Scenario: Panic with debug enabled
- **WHEN** an unexpected panic occurs
- **AND** `TWIGGIT_DEBUG` is set to any non-empty value
- **THEN** stderr SHALL contain "Internal error:"
- **AND** stderr SHALL contain a Go stack trace
- **AND** the process SHALL exit with code 1

### Requirement: Debug mode preserves internal details

When `TWIGGIT_DEBUG` environment variable is set, error messages MAY include additional internal details for debugging purposes.

This requirement is OPTIONAL and applies only when explicitly enabled by the user.

#### Scenario: Debug mode shows additional context
- **WHEN** an error occurs
- **AND** `TWIGGIT_DEBUG` is set
- **THEN** the error output MAY include internal operation names and stack traces
