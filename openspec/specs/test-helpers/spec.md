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

### Requirement: Automatic resource cleanup
Test helpers SHALL use appropriate cleanup mechanisms for automatic resource cleanup when tests complete.

#### Scenario: RepoTestHelper constructor registers cleanup
- **GIVEN** a test requiring RepoTestHelper
- **WHEN** NewRepoTestHelper is called
- **THEN** a cleanup function is registered via t.Cleanup() to call the helper's Cleanup() method
- **AND** the cleanup function removes all created repositories
- **AND** cleanup runs even if the test fails or panics

#### Scenario: GitTestHelper uses t.TempDir for automatic cleanup
- **GIVEN** a test requiring GitTestHelper
- **WHEN** NewGitTestHelper is called
- **THEN** the helper uses t.TempDir() for the base directory
- **AND** the testing package automatically cleans up the temp directory when the test completes
- **AND** cleanup runs even if the test fails or panics

#### Scenario: Multiple cleanup functions execute in LIFO order
- **GIVEN** a test with multiple resources that register cleanup
- **WHEN** multiple resources are created in sequence
- **THEN** cleanup functions execute in reverse order of registration (LIFO)
- **AND** the last registered cleanup runs first

### Requirement: Helper function error line reporting
Test helper functions SHALL call t.Helper() to improve error line reporting.

#### Scenario: Helper function marks itself
- **GIVEN** a test that calls a helper function
- **WHEN** a helper function calls t.Helper()
- **THEN** error reports point to the calling test code
- **AND** not to the helper function internals

#### Scenario: Nested helper functions both call t.Helper()
- **GIVEN** a test that calls a helper function
- **WHEN** a helper function calls another helper function
- **THEN** both functions call t.Helper()
- **AND** error reports point to the original test code
