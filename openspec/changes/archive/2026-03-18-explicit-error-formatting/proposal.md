## Why

The current error formatter uses reflection-based type dispatch, which introduces runtime overhead, makes the matcher-to-formatter mapping unclear, and complicates debugging and extension. This change refactors to an explicit strategy pattern using idiomatic Go `errors.As()` matching.

## What Changes

- Define `matcherFunc` and `formatterFunc` types in `cmd/error_formatter.go`
- Create explicit matcher functions (`isValidationError`, `isWorktreeError`, `isProjectError`, `isServiceError`) using `errors.As()`
- Refactor `ErrorFormatter.register` to accept matcher+formatter pair
- Refactor `ErrorFormatter.Format` to iterate through matchers instead of reflection map lookup
- Update formatter functions to accept error directly (remove `ErrorFormatter` receiver parameter)
- Update tests to verify behavior unchanged

## Capabilities

### New Capabilities
- `explicit-error-formatting`: Error formatter using explicit strategy pattern with errors.As() matching (no reflection)

### Modified Capabilities
*(none - this is an internal refactoring with no spec-level behavior changes)*

## Impact

**Files Modified:**
- `cmd/error_formatter.go` - Core refactoring to explicit pattern
- `cmd/error_formatter_test.go` - Update tests for new structure
- `cmd/AGENTS.md` - Document explicit error formatter pattern

**API Stability:** No changes to public API or error output format
**Dependencies:** Removes reliance on `reflect` package for type detection
