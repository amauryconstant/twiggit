## Context

Test helpers in `test/helpers/*.go` provide testing utilities for the test suite. Currently:
- `NewGitTestHelper` constructor already calls `t.Helper()` and uses `t.TempDir()` for automatic cleanup
- `NewRepoTestHelper` constructor already calls `t.Helper()` and uses `t.TempDir()` for base directory, and has a `Cleanup()` method that should be registered with `t.Cleanup()`
- `NewShellTestHelper` constructor already calls `t.Helper()` and uses temp directories for config files
- Constructor methods do not currently call `t.Helper()` (methods excluded from scope)
- Some other helper constructors may lack `t.Helper()` calls, causing error line reports to point to helper internals

## Goals / Non-Goals

**Goals:**
- Verify `t.Helper()` is present in all helper constructors for accurate error line reporting (methods excluded from scope)
- Register `t.Cleanup()` in RepoTestHelper constructor to call Cleanup() method automatically
- Verify GitTestHelper uses `t.TempDir()` for automatic cleanup (cleanup already handled by testing package)
- Document cleanup patterns in test/helpers/AGENTS.md

**Non-Goals:**
- Refactoring helper function signatures
- Adding new helper functions
- Changing existing test assertions
- Modifying WorktreeTestHelper (constructor does not accept `*testing.T` parameter)
- Adding Cleanup methods to helpers that use t.TempDir() (cleanup already automatic)

## Decisions

### Decision 1: Register t.Cleanup() for RepoTestHelper

**Choice:** Register Cleanup() method in RepoTestHelper constructor via `t.Cleanup()` to ensure automatic cleanup of created repositories.

**Rationale:** While t.TempDir() handles cleanup of the base directory, the Cleanup() method removes subdirectories (created repos) before the base directory is cleaned up. Registering it ensures cleanup happens even if tests fail or panic.

**Alternatives:**
- Relying on t.TempDir() only - rejected: Cleanup() removes individual repos which is more explicit and cleaner
- Requiring explicit `defer h.Cleanup()` in tests - rejected: easy to forget, leads to leaks

### Decision 2: Verify t.TempDir() usage for other helpers

**Choice:** Verify that GitTestHelper and other helpers use t.TempDir() which provides automatic cleanup via the testing package.

**Rationale:** t.TempDir() is the idiomatic Go testing approach for temporary directory management. The testing package guarantees cleanup runs even on test failure or panic.

**Alternatives:**
- Adding manual Cleanup methods to all helpers - rejected: t.TempDir() is simpler and idiomatic
- Using os.RemoveAll in deferred calls - rejected: redundant with t.TempDir()

### Decision 3: t.Helper() on helper constructors only

**Choice:** Verify helper constructors call `t.Helper()` at the start of the function. Methods are excluded from scope.

**Rationale:** Error messages will report the line number of the test code calling the helper constructor, not the constructor internals. Methods are excluded because they already have access to `*testing.T` through the helper struct and adding `t.Helper()` to all methods would require passing `*testing.T` as a method parameter.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Cleanup ordering issues if dependencies exist | LIFO order from t.Cleanup() naturally handles dependencies |
| Existing tests rely on manual cleanup | No change to behavior, cleanup becomes automatic |
| Performance impact of cleanup registration | Negligible - just function registration |
| Verification reveals missing t.Helper() calls | Add t.Helper() to constructors that lack it |

## Implementation Notes

- GitTestHelper uses t.TempDir() so no additional cleanup needed
- RepoTestHelper needs t.Cleanup() registration for its Cleanup() method
- ShellTestHelper uses temp directories for config files, verification only needed
- All helpers should be verified for t.Helper() presence
