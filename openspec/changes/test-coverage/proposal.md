## Why

The main package has 0% test coverage and test helpers are only at 43.5%, creating reliability gaps before public release. Additionally, there are no concurrent operation tests or edge case fixtures to validate graceful error handling in unusual repository states.

## What Changes

- Create main package tests for entry point and initialization paths
- Add concurrent operation tests with race detector validation
- Create edge case test fixtures (corrupted repo, bare repo, submodules, detached HEAD)
- Improve test/helpers coverage from 43.5% to >70%

## Capabilities

### New Capabilities

- `concurrent-operations`: Test suite for concurrent worktree operations with race detector validation
- `edge-case-fixtures`: Test fixtures for unusual repository states (corrupted, bare, submodules, detached HEAD)
- `main-entry-point`: Tests for main package initialization and error handling paths

### Modified Capabilities

- `test-helpers`: Extend coverage for worktree and shell helpers to >70%

## Impact

**New Files:**
- `main_test.go` - Entry point tests
- `test/concurrent/` - Concurrent operation tests with `//go:build concurrent` tag
- `test/e2e/fixtures/corrupted/` - Corrupted repository fixture
- `test/e2e/fixtures/bare/` - Bare repository fixture
- `test/e2e/fixtures/submodule/` - Repository with submodules fixture
- `test/e2e/fixtures/detached/` - Detached HEAD state fixture
- `test/helpers/worktree_test.go` - Worktree helper tests
- `test/helpers/shell_test.go` - Shell helper tests

**Dependencies:**
- This change SHOULD be implemented AFTER all other pre-release changes are complete
- Validates that the entire system works correctly under test coverage

**Coverage Targets:**
| Package | Current | Target |
|---------|---------|--------|
| main | 0.0% | >50% |
| test/helpers | 43.5% | >70% |
