# Testing Conventions

## Test Organization

| Type | Location | Framework | Build Tag |
|------|----------|-----------|-----------|
| Unit | `internal/**/*_test.go` | Testify | None |
| Integration | `test/integration/` | Testify | `//go:build integration` |
| E2E | `test/e2e/` | Ginkgo/Gomega | `//go:build e2e` |
| Concurrent | `test/concurrent/` | Testify | `//go:build concurrent` |
| Main | `main_test.go` | Testify | `//go:build integration` |
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
- Golden helper (snapshot testing)

## E2E Testing

See test/e2e/AGENTS.md for Ginkgo patterns, CLI execution, verbose output testing.

## Golden File Testing

Golden file testing provides snapshot testing for CLI output verification.

**Purpose:** Verify complex multi-line output with easy updates and clear diffs on failure.

**Location:** test/golden/<category>/<name>.golden

**Usage:**
```go
// In E2E tests (requires //go:build e2e build tag)
actualOutput := executeCLICommand("list", "--all")
helpers.CompareGolden(t, "list/basic_output.golden", actualOutput)
```

**Update Golden Files:**
```bash
UPDATE_GOLDEN=true mise run test:golden
# or
UPDATE_GOLDEN=true mise run test:golden:update
```

**mise tasks:**
- `mise run test:golden` - Run golden file tests
- `mise run test:golden:update` - Update golden files

**Benefits:**
- Easy to review output changes via git diff
- Clear diffs when tests fail
- No hard-coded expectations in test code
- Simple update mechanism for intentional changes

## Build Tags

```go
//go:build integration  // Integration tests
//go:build e2e          // E2E tests
//go:build concurrent   // Concurrent operation tests (race detector)

if testing.Short() { t.Skip() }  // Skip in short mode
```

## Concurrent Tests

See test/concurrent/AGENTS.md for race detector validation patterns.
