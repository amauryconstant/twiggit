# Capability: Test Helpers

## Purpose

The `test/helpers` package provides utility functions for creating,
validating, and cleaning up test worktrees, plus shell-related test
utilities. Helpers enforce consistent setup/teardown across the test
suite and improve error-line reporting via `t.Helper()`.

## Requirements

### Requirement: Worktree helpers

The package SHALL provide worktree utilities: create, validate,
cleanup, plus error handling. Coverage SHALL exceed 70%.

#### Scenario: Create test worktree

- **WHEN** `CreateTestWorktree` is called with a valid repo and branch
- **THEN** a worktree SHALL be created at the expected path
- **AND** the worktree SHALL point to the correct branch

#### Scenario: Validate worktree

- **WHEN** `ValidateWorktree` is called with a valid worktree path
- **THEN** validation SHALL succeed when the branch matches
- **AND** SHALL fail with a descriptive error otherwise

#### Scenario: Cleanup worktree

- **WHEN** `CleanupWorktree` is called
- **THEN** the worktree SHALL be removed from the filesystem
- **AND** no residual files SHALL remain

### Requirement: Shell helpers

The package SHALL provide shell utilities: execute command, get shell
path, create temp shell config, plus error handling. Coverage SHALL
exceed 70%.

#### Scenario: Execute shell command

- **WHEN** `ExecuteShellCommand` is called with a valid command
- **THEN** the command output SHALL be returned
- **AND** error SHALL be `nil` for successful commands

#### Scenario: Create temp shell config

- **WHEN** `CreateTempShellConfig` is called with a shell type
- **THEN** a temporary config file SHALL be created
- **AND** the file SHALL contain valid shell configuration
- **AND** the file SHALL be cleaned up after the test

### Requirement: Automatic resource cleanup

Helpers SHALL register cleanup functions via `t.Cleanup()` or use
`t.TempDir()` for automatic teardown. Cleanup SHALL run even when the
test fails or panics. Multiple cleanups SHALL execute in LIFO order.

#### Scenario: RepoTestHelper cleanup

- **WHEN** `NewRepoTestHelper` is called
- **THEN** the helper SHALL register a `t.Cleanup` callback to remove
  all created repositories
- **AND** cleanup SHALL run on test failure or panic

#### Scenario: GitTestHelper temp dir

- **WHEN** `NewGitTestHelper` is called
- **THEN** the helper SHALL use `t.TempDir()` for the base directory
- **AND** the temp directory SHALL be removed automatically when the
  test completes

### Requirement: `t.Helper()` for error reporting

All helper functions SHALL call `t.Helper()` so that test failures
report the calling line, not the helper internals. Nested helpers
SHALL each call `t.Helper()`.

#### Scenario: Helper marks itself

- **WHEN** a helper function calls `t.Helper()`
- **THEN** error reports SHALL point to the calling test code

#### Scenario: Nested helpers

- **WHEN** a helper calls another helper that calls `t.Helper()`
- **THEN** both layers SHALL call `t.Helper()`
- **AND** errors SHALL point to the original test code
