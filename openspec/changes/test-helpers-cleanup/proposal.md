## Why

Test helpers require verification of t.Helper() usage and consistent cleanup patterns. While some helpers (NewGitTestHelper, NewRepoTestHelper) already call t.Helper(), they should be verified for consistency. RepoTestHelper has a Cleanup() method that should be registered with t.Cleanup() for automatic cleanup. This change ensures all test helpers follow consistent patterns for cleanup and accurate error line reporting.

## What Changes

- Verify t.Helper() calls are present in helper constructors in test/helpers/*.go (methods excluded)
- Register t.Cleanup() for RepoTestHelper constructor to call Cleanup() method automatically
- Verify GitTestHelper uses t.TempDir() for automatic cleanup (already handled)
- Update test/helpers/AGENTS.md with cleanup documentation

## Capabilities

### New Capabilities
- `test-helpers`: Test helpers with automatic cleanup and accurate error line reporting

### Modified Capabilities
None

## Implementation Status
- NewGitTestHelper: t.Helper() already present, uses t.TempDir() for cleanup
- NewRepoTestHelper: t.Helper() already present, needs t.Cleanup() registration
- NewShellTestHelper: t.Helper() already present
- RepoTestHelper: Has Cleanup() method, needs t.Cleanup() registration

## Impact

- Files Modified: test/helpers/repo.go, test/helpers/AGENTS.md
- Verification only: test/helpers/git.go, test/helpers/shell.go
- No API changes - internal test infrastructure only
- All existing tests continue to pass with enhanced cleanup guarantees
