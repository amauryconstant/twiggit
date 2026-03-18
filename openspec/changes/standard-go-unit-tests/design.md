## Context

Current unit tests use `github.com/stretchr/testify/suite` for test organization. This adds an abstraction layer that:
- Requires learning testify-specific patterns
- Obscures test setup/teardown flow
- Adds an unnecessary dependency for unit tests
- Makes debugging harder due to indirect execution

The standard Go `testing` package provides all needed functionality through `t.Run()`, `t.Cleanup()`, and `t.Helper()`.

## Goals / Non-Goals

**Goals:**
- Convert all 27 unit test files to standard Go testing patterns
- Use table-driven patterns for tests with 5+ cases
- Use `t.Cleanup()` for automatic mock assertion verification
- Remove testify/suite dependency from go.mod
- Update test documentation with new patterns

**Non-Goals:**
- Changing integration tests (testify/suite is acceptable there)
- Changing E2E tests (Ginkgo/Gomega is acceptable there)
- Changing test behavior or coverage
- Adding new test cases

## Decisions

### Decision 1: Use t.Run() for subtests

**Choice:** Use `t.Run()` for organizing subtests instead of suite methods.

**Rationale:** `t.Run()` is idiomatic Go, provides isolated subtest context, and integrates with `go test -run` filtering. Each subtest gets its own `*testing.T` for proper failure reporting.

**Alternatives:**
- Keep suite methods: Rejected - adds unnecessary abstraction
- Single test function per file: Rejected - loses test isolation

### Decision 2: Use t.Cleanup() for mock assertions

**Choice:** Register mock assertions via `t.Cleanup()` at test start.

**Rationale:** Ensures assertions run even if test panics or fails early. Cleaner than defer patterns and standard Go idiom.

**Pattern:**
```go
mock := mocks.NewMockService(t)
t.Cleanup(func() { mock.AssertExpectations(t) })
```

### Decision 3: Table-driven tests for 5+ cases

**Choice:** Use table-driven pattern when a function has 5+ test variations.

**Rationale:** Reduces boilerplate, makes test cases explicit, easier to add new cases. Below 5 cases, direct `t.Run()` calls are acceptable.

**Pattern:**
```go
tests := []struct {
    name      string
    input     string
    wantErr   bool
}{
    {"valid input", "valid", false},
    {"empty input", "", true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### Decision 4: Fresh dependencies per subtest

**Choice:** Each subtest creates its own mocks and dependencies.

**Rationale:** Ensures test isolation. No shared state between subtests. Use `t.TempDir()` for filesystem isolation or constructor injection for mocks.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Missing mock assertions | Require `t.Cleanup()` at mock creation, review each conversion |
| Shared state between tests | Enforce fresh dependencies per subtest, no suite-level fields |
| Large PR difficult to review | Split into phases by layer (domain → infrastructure → service → cmd) |
| Race conditions introduced | Run `mise run test:race` after each phase |
| testify/suite still imported | Add explicit cleanup task to remove import and verify |

## Migration Plan

1. **Phase 5a:** Convert domain layer (7 files)
2. **Phase 5b:** Convert infrastructure layer (11 files)
3. **Phase 5c:** Convert service layer (5 files)
4. **Phase 5d:** Convert command layer (4 files)
5. **Phase 5e:** Cleanup - remove import, run full test suite, update docs

Each phase runs `mise run test:full` and `mise run test:race` before proceeding.
