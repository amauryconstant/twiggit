## Why

Test helpers lack automatic resource cleanup, leading to potential resource leaks when tests fail or panic. Additionally, helper functions don't mark themselves with t.Helper(), causing error reports to point to the wrong line. This change establishes consistent patterns for cleanup and error reporting across all test helpers.

## What Changes

- Add t.Helper() calls to helper constructors in test/helpers/*.go (methods excluded)
- Add t.Cleanup() patterns to GitTestHelper and RepoTestHelper constructors for automatic resource cleanup
- Update test/helpers/AGENTS.md with cleanup documentation

## Capabilities

### New Capabilities
- `test-helpers`: Test helpers with automatic cleanup and accurate error line reporting

### Modified Capabilities
None

## Impact

- Files Modified: test/helpers/*.go, test/helpers/AGENTS.md
- No API changes - internal test infrastructure only
- All existing tests continue to pass with enhanced cleanup guarantees
