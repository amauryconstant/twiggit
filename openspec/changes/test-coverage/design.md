## Context

The twiggit project is approaching public release with inadequate test coverage in critical areas:
- Main package: 0% coverage (entry point untested)
- Test helpers: 43.5% coverage (utilities that support all other tests)
- No concurrent operation tests (race conditions undetected)
- No edge case fixtures (unusual repository states untested)

This change fills these gaps to ensure reliability for public release.

**Constraints:**
- Tests written AFTER implementation per project convention
- Use existing test frameworks (Testify for helpers, Ginkgo for E2E)
- Follow existing build tag patterns (`//go:build integration`, `//go:build e2e`)
- New concurrent tag: `//go:build concurrent`

## Goals / Non-Goals

**Goals:**
- Achieve >50% coverage for main package
- Achieve >70% coverage for test/helpers
- Create concurrent operation tests that pass race detector
- Create edge case fixtures for graceful error handling validation

**Non-Goals:**
- Refactoring existing code for testability
- Adding new production features
- Changing existing test patterns

## Decisions

### Decision 1: Main Package Test Strategy

**Choice:** Create `main_test.go` with integration-style tests using build tag `//go:build integration`

**Rationale:** The main package orchestrates initialization and error handling. Unit testing would require extensive mocking of infrastructure and service layers. Integration-style tests with real git repos provide meaningful coverage without brittle mocks.

**Alternatives Considered:**
- Unit tests with mocks: Rejected due to excessive setup for initialization paths
- E2E tests: Rejected because E2E tests the CLI via built binary, not main.go directly

### Decision 2: Concurrent Test Build Tag

**Choice:** Use `//go:build concurrent` tag for concurrent operation tests

**Rationale:** Concurrent tests are inherently slower and more resource-intensive. A dedicated build tag allows selective execution and prevents CI timeouts during regular test runs.

**Alternatives Considered:**
- No tag (always run): Rejected due to CI time impact
- Integration tag: Rejected to maintain separation of concerns

### Decision 3: Edge Case Fixture Location

**Choice:** Create fixtures in `test/e2e/fixtures/` alongside existing fixtures

**Rationale:** Existing fixture infrastructure (tar.gz extraction, cleanup) can be extended. E2E tests already use fixtures for repository states.

**Alternatives Considered:**
- `test/fixtures/`: Rejected to avoid creating new fixture infrastructure
- Inline test creation: Rejected because corrupted/bare repos are complex to create programmatically

### Decision 4: Test Helpers Coverage Approach

**Choice:** Create dedicated test files (`worktree_test.go`, `shell_test.go`) using Testify

**Rationale:** Test helpers are pure utility functions with no external dependencies. Standard unit tests provide fast, reliable coverage.

**Alternatives Considered:**
- Integration tests: Rejected because helpers don't need real git repos
- E2E tests: Rejected for same reason

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Concurrent tests may be flaky | Use deterministic test data, add retry logic, ensure proper cleanup |
| Edge case fixtures may become stale | Document creation process, add fixture validation tests |
| Main package tests may be slow | Use short tests where possible, parallelize where safe |
| Coverage targets may not be achievable | Prioritize critical paths, document uncovered code |
