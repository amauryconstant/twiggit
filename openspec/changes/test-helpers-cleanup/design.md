## Context

Test helpers in `test/helpers/*.go` provide testing utilities for the test suite. Currently:
- `NewGitTestHelper` and `NewRepoTestHelper` constructors already call `t.Helper()` but lack automatic cleanup registration
- Other helper constructors lack `t.Helper()` calls, causing error line reports to point to helper internals instead of test code
- Constructor methods do not currently call `t.Helper()` (methods excluded from scope)

## Goals / Non-Goals

**Goals:**
- Add `t.Helper()` to helper constructors for accurate error line reporting (methods excluded from scope)
- Add `t.Cleanup()` patterns to constructors for automatic resource management
- Ensure cleanup runs even on test failure or panic
- Document cleanup patterns in test/helpers/AGENTS.md

**Non-Goals:**
- Refactoring helper function signatures
- Adding new helper functions
- Changing existing test assertions
- Modifying WorktreeTestHelper (constructor does not accept `*testing.T` parameter)

## Decisions

### Decision 1: Use t.Cleanup() in constructors

**Choice:** Register cleanup in constructor via `t.Cleanup()` rather than requiring explicit cleanup calls.

**Rationale:** Automatic cleanup prevents resource leaks when tests fail or panic. The testing package guarantees cleanup runs even on failure.

**Alternatives:**
- Requiring explicit `defer h.Cleanup()` in tests - rejected: easy to forget, leads to leaks
- Using `t.TempDir()` only - rejected: doesn't handle all resources (mocks, open files, etc.)

### Decision 2: LIFO cleanup order

**Choice:** Multiple cleanup functions execute in LIFO (last-in, first-out) order per Go's testing package behavior.

**Rationale:** Natural stack-like behavior ensures dependencies are cleaned up in reverse order of creation.

### Decision 3: t.Helper() on helper constructors only

**Choice:** Mark helper constructors with `t.Helper()` at the start of the function. Methods are excluded from scope.

**Rationale:** Error messages will report the line number of the test code calling the helper constructor, not the constructor internals. Methods are excluded because they already have access to `*testing.T` through the helper struct and adding `t.Helper()` to all methods would require passing `*testing.T` as a method parameter.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Cleanup ordering issues if dependencies exist | LIFO order naturally handles dependencies |
| Existing tests rely on manual cleanup | No change to behavior, cleanup becomes automatic |
| Performance impact of cleanup registration | Negligible - just function registration |
