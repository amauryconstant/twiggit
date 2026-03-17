## Context

E2E tests for CLI commands produce complex multi-line output that is cumbersome to verify with hard-coded expectations. Golden file testing provides a maintainable approach: store expected output in files, compare actual output, and easily update when intentional changes occur.

## Goals / Non-Goals

**Goals:**
- Create CompareGolden helper function for snapshot testing
- Support UPDATE_GOLDEN environment variable to update golden files
- Create golden file tests for list command output (text and JSON)
- Create golden file tests for error formatting
- Add mise tasks for running and updating golden tests

**Non-Goals:**
- Replacing all existing tests with golden tests
- Golden tests for unit or integration tests (E2E only)

## Decisions

### Decision 1: Golden file location
**Choice:** test/golden/<category>/<name>.golden
**Rationale:** Mirrors test structure, easy to find related tests
**Alternatives:**
- Inline golden files: Harder to review diffs
- test/e2e/testdata/: Mixes with other fixtures

### Decision 2: UPDATE_GOLDEN via environment variable
**Choice:** Environment variable UPDATE_GOLDEN=true
**Rationale:** Standard Go pattern (used by testing frameworks), easy CI integration
**Alternatives:**
- Command-line flag: Requires test binary modification
- Config file: Over-engineering for this use case

### Decision 3: CompareGolden signature
**Choice:** CompareGolden(t *testing.T, goldenFile string, actual string)
**Rationale:** Simple, testifies to testing.T, flexible for any output
**Alternatives:**
- []byte instead of string: Adds conversion noise for most CLI output
- io.Reader: Over-complicates for string output

## Risks / Trade-offs

**Risk:** Golden files drift from actual behavior → Mitigation: Run in CI without UPDATE_GOLDEN
**Risk:** Large golden files are hard to review → Mitigation: Keep tests focused, split by scenario
