## Why

Twiggit has accumulated technical debt across multiple areas: inconsistent error handling (several places return `fmt.Errorf` instead of proper domain types per project conventions), type safety gaps (nil pointer risks), duplicate code patterns, concurrency issues in prune operations, hardcoded configuration values, and CLI ergonomics. Addressing these improves code correctness, maintainability, and user experience without changing any user-facing behavior.

## What Changes

1. **Error Handling**: Fix 10 instances of improper error wrapping/returning across `cmd/` and `service/` layers; add missing `IsNotFound()` methods to `ProjectServiceError`, `NavigationServiceError`, and `ResolutionError`
2. **Type Safety**: Add nil checks for `req.Context` and `resolution.ResolvedPath` to prevent panics
3. **Duplicate Code**: Consolidate shell auto-detection logic, navigation target resolution, and path validation into shared helpers
4. **Concurrency**: Add mutex protection to `PruneMergedWorktrees` result modifications to eliminate race conditions
5. **CLI Improvements**: Use config defaults for source branch, add missing short flags, add prune preview
6. **Hardcoded Values**: Route CLI timeout and hook timeout through config instead of hardcoding
7. **Testing**: Add unit tests for `cmd/` layer, e2e test for `prune --delete-branches`

**BREAKING**: None - all changes preserve existing behavior.

## Capabilities

### New Capabilities
- `error-handling`: Consistent error types with IsNotFound() methods across all service errors
- `type-safety`: Nil-safe operations preventing panics in edge cases
- `code-deduplication`: Shared helper functions eliminating duplicate logic
- `cli-improvements`: Flag consistency and preview before destructive operations
- `config-driven-timeouts`: Configuration-based timeouts instead of hardcoded values
- `testing-coverage`: Unit tests for cmd layer and e2e coverage for prune --delete-branches

### Modified Capabilities
- `concurrent-operations`: Add mutex protection to PruneMergedWorktrees (implementation detail, no requirement change)
- `error-clarity`: Implement proper domain error types (implementation detail, no requirement change)

## Impact

| Layer | Files Affected |
|-------|---------------|
| `cmd/` | `create.go`, `delete.go`, `cd.go`, `prune.go`, `util.go` (new) |
| `service/` | `worktree_service.go`, `shell_service.go` |
| `domain/` | `service_errors.go` |
| `infrastructure/` | `cli_client.go`, `hook_runner.go`, `context_resolver.go` |
| `test/` | New tests in `cmd/`, `test/e2e/`, `test/integration/` |
