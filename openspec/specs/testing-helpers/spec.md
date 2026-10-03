# Capability: Test Helpers

## Purpose

The `test/helpers` package provides utility functions for creating,
validating, and cleaning up test worktrees, plus shell-related test
utilities. Helpers enforce consistent setup/teardown across the test
suite and improve error-line reporting via `t.Helper()`.

## Requirements

### Requirement: Automatic resource cleanup

Helpers SHALL register cleanup functions via `t.Cleanup()` or use
`t.TempDir()` for automatic teardown. Cleanup SHALL run even when the
test fails or panics. Multiple cleanups SHALL execute in LIFO order.
This contract applies to every helper in `test/worktree`, `test/shell`,
`test/git`, `test/repo`, `test/golden`, and `test/perf` — no helper
package is exempt.

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

All helper functions across every `test/<domain>/` package SHALL call
`t.Helper()` so that test failures report the calling line, not the
helper internals. Nested helpers SHALL each call `t.Helper()`.

#### Scenario: Helper marks itself

- **WHEN** a helper function calls `t.Helper()`
- **THEN** error reports SHALL point to the calling test code

#### Scenario: Nested helpers

- **WHEN** a helper calls another helper that calls `t.Helper()`
- **THEN** both layers SHALL call `t.Helper()`
- **AND** errors SHALL point to the original test code

### Requirement: Test helper packages are content-named

The repository SHALL organise test helpers under
`test/<domain>/` directories, one per concern. The package name SHALL
describe the concern, not the role (no `test/helpers`,
`test/util`, `test/common`, `test/misc`). The current split
establishes:

- `test/worktree/` — worktree creation, validation, cleanup
- `test/shell/` — shell command execution, temp shell configs,
  shell-type detection
- `test/git/` — low-level git fixture helpers
- `test/repo/` — git repository fixtures
- `test/golden/` — golden-file compare and update helpers
- `test/perf/` — performance measurement helpers

#### Scenario: No helpers or util package exists

- **WHEN** `test/` is searched for packages named `helpers`, `util`,
  `common`, `misc`, or `support`
- **THEN** no directory under `test/` SHALL match those names

#### Scenario: Worktree helpers live in test/worktree

- **WHEN** a test imports `CreateTestWorktree`, `ValidateWorktree`,
  or `CleanupWorktree`
- **THEN** the import path SHALL be `twiggit/test/worktree`

#### Scenario: Shell helpers live in test/shell

- **WHEN** a test imports `ExecuteShellCommand`, `CreateTempShellConfig`,
  or `GetShellPath`
- **THEN** the import path SHALL be `twiggit/test/shell`

#### Scenario: Golden and perf helpers live in their own packages

- **WHEN** a test imports a golden-file compare helper
- **THEN** the import path SHALL be `twiggit/test/golden`
- **WHEN** a test imports a perf-measurement helper
- **THEN** the import path SHALL be `twiggit/test/perf`
