## Why

Testify/suite adds unnecessary abstraction for unit tests. Standard Go testing with `t.Run()`, `t.Cleanup()`, and `t.Helper()` provides all needed structure. Table-driven patterns are more explicit, easier to debug, and familiar to Go developers. Removing the suite dependency simplifies onboarding and reduces test complexity.

## What Changes

- Convert 27 unit test files from testify/suite to standard Go testing
- Replace `suite.Suite` embedding with plain test functions
- Use `t.Run()` for subtests instead of suite test methods
- Use `t.Cleanup()` for mock assertions instead of `suite.T().Cleanup()`
- Apply table-driven patterns for test cases with 5+ variations
- Each subtest creates fresh dependencies via `t.TempDir()` or constructor injection
- Remove `github.com/stretchr/testify/suite` import after conversion

## Capabilities

### New Capabilities

- `standard-go-unit-tests`: Standard Go testing patterns for unit tests using table-driven designs with `t.Run()` for subtests and `t.Cleanup()` for automatic mock verification

### Modified Capabilities

## Impact

**Files Modified**: 27 test files across all layers:
- Domain layer: 7 test files
- Infrastructure layer: 11 test files
- Service layer: 5 test files
- Command layer: 4 test files (contract tests: completion, error_handler, init, suggestions)

**Other Changes**:
- `go.mod`: Remove testify/suite import
- `test/AGENTS.md`: Update testing patterns documentation

**Dependencies**: Requires `test-helpers-cleanup` change (mock constructors with `t.Cleanup()`) to be complete first.
