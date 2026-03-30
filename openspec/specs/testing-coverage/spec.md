# Testing Coverage

## Purpose

Improve test coverage for cmd layer and prune --delete-branches functionality.

## Requirements

### Requirement: Cmd layer SHALL have unit tests

The cmd layer SHALL have unit tests for error_formatter.go and util.go using testify.

#### Scenario: ErrorFormatter formats ValidationError correctly
- **WHEN** ValidationError is passed to FormatError
- **THEN** formatted output includes error message and context

#### Scenario: ErrorFormatter formats ServiceError correctly
- **WHEN** ServiceError is passed to FormatError
- **THEN** formatted output includes service context

#### Scenario: Utility functions work correctly
- **WHEN** utility functions like logv, isQuiet, ProgressReporter are called
- **THEN** they return expected values for various inputs

### Requirement: Prune --delete-branches SHALL be e2e tested

The prune command with --delete-branches flag SHALL have end-to-end test coverage.

#### Scenario: Prune deletes worktree and branch together
- **WHEN** user runs twiggit prune --delete-branches --worktree=project/branch
- **THEN** worktree is deleted
- **AND** branch is deleted from repository

#### Scenario: Prune --delete-branches fails gracefully
- **WHEN** user runs twiggit prune --delete-branches with non-existent worktree
- **THEN** appropriate error is displayed
- **AND** no partial cleanup occurs

### Requirement: Shell wrapper block SHALL have integration test

The shell wrapper block removal and append logic SHALL have integration test coverage.

#### Scenario: Shell wrapper is correctly installed
- **WHEN** shell integration is installed
- **THEN** wrapper block is appended to shell config
- **AND** wrapper block contains twiggit cd function
