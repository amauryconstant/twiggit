# Testing Conventions

## Test Organization

| Type | Location | Framework | Build Tag |
|------|----------|-----------|-----------|
| Unit | `internal/**/*_test.go` | Testify | None |
| Integration | `test/integration/` | Testify | `//go:build integration` |
| E2E | `test/e2e/` | Ginkgo/Gomega | `//go:build e2e` |
| Mocks | `test/mocks/*.go` | testify/mock | - |

**Detailed patterns:** See test/integration/AGENTS.md (Testify suites, mocks, assertions)

## Quality Requirements

- `mise run test:full` - All tests must pass
- `mise run test:race` - Race detector must pass
- Coverage target: >80% for new code
- Tests must be deterministic, fast (<100ms each)

## Test Helpers

See test/helpers/AGENTS.md for:
- Repository helper (test repo creation)
- Git helper (git operations)
- Shell helper (shell utilities)
- Worktree helper (worktree utilities)

## E2E Testing

See test/e2e/AGENTS.md for Ginkgo patterns, CLI execution, verbose output testing.

## Build Tags

```go
//go:build integration  // Integration tests
//go:build e2e          // E2E tests

if testing.Short() { t.Skip() }  // Skip in short mode
```
