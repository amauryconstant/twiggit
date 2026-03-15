# Test Helpers

## Purpose

The test/helpers package provides utility functions for creating, validating, and cleaning up test worktrees, as well as shell-related test utilities. These helpers ensure consistent test setup and teardown across the test suite.

## Requirements

### Requirement: Worktree helper coverage
The test/helpers package SHALL provide comprehensive worktree utilities with >70% test coverage.

#### Scenario: Create test worktree
- **WHEN** CreateTestWorktree is called with valid repo and branch
- **THEN** worktree is created at expected path
- **AND** worktree points to correct branch

#### Scenario: Validate worktree
- **WHEN** ValidateWorktree is called with valid worktree path
- **THEN** validation succeeds for correct branch
- **AND** validation fails for incorrect branch

#### Scenario: Cleanup worktree
- **WHEN** CleanupWorktree is called
- **THEN** worktree is removed from filesystem
- **AND** no residual files remain

#### Scenario: Worktree helper error handling
- **WHEN** worktree helper operations fail
- **THEN** descriptive error is returned
- **AND** test fails with meaningful message

### Requirement: Shell helper coverage
The test/helpers package SHALL provide comprehensive shell utilities with >70% test coverage.

#### Scenario: Execute shell command
- **WHEN** ExecuteShellCommand is called with valid command
- **THEN** command output is returned
- **AND** error is nil for successful commands

#### Scenario: Get shell path
- **WHEN** GetShellPath is called
- **THEN** valid shell path is returned
- **AND** shell exists on system

#### Scenario: Create temp shell config
- **WHEN** CreateTempShellConfig is called with shell type
- **THEN** temporary config file is created
- **AND** config file contains valid shell configuration
- **AND** file is cleaned up after test

#### Scenario: Shell helper error handling
- **WHEN** shell helper operations fail
- **THEN** descriptive error is returned
- **AND** test fails with meaningful message
