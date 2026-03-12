## 1. Main Package Tests

- [ ] 1.1 Create `main_test.go` in root directory with `//go:build integration` tag
- [ ] 1.2 Test config load failure handling (invalid YAML, missing directory)
- [ ] 1.3 Test successful execution path with valid config
- [ ] 1.4 Test help command execution
- [ ] 1.5 Test command execution failure with appropriate exit codes
- [ ] 1.6 Verify main package coverage >50%

## 2. Concurrent Operation Tests

- [ ] 2.1 Create `test/concurrent/` directory structure
- [ ] 2.2 Create `test/concurrent/concurrent_test.go` with `//go:build concurrent` tag
- [ ] 2.3 Test concurrent list operations on same project
- [ ] 2.4 Test concurrent create operations on different worktrees
- [ ] 2.5 Test concurrent delete operations on different worktrees
- [ ] 2.6 Test create and delete different worktrees concurrently
- [ ] 2.7 Test prune while listing operations
- [ ] 2.8 Run all tests with race detector: `mise run test:race` (or equivalent with concurrent tag)

## 3. Edge Case Fixtures

- [ ] 3.1 Create corrupted repository fixture in `test/e2e/fixtures/corrupted/`
  - Create tar.gz of repo with corrupted `.git/objects`
- [ ] 3.2 Create bare repository fixture in `test/e2e/fixtures/bare/`
  - Create tar.gz of bare git repo (no working tree)
- [ ] 3.3 Create submodule repository fixture in `test/e2e/fixtures/submodule/`
  - Create tar.gz of repo containing git submodules
- [ ] 3.4 Create detached HEAD fixture in `test/e2e/fixtures/detached/`
  - Create tar.gz of repo in detached HEAD state
- [ ] 3.5 Add fixture loading functions to `test/e2e/fixtures/e2e_fixtures.go`
- [ ] 3.6 Create E2E tests for corrupted repository handling
- [ ] 3.7 Create E2E tests for bare repository handling
- [ ] 3.8 Create E2E tests for submodule repository handling
- [ ] 3.9 Create E2E tests for detached HEAD handling

## 4. Test Helpers Coverage

- [ ] 4.1 Create `test/helpers/worktree_test.go`
- [ ] 4.2 Test CreateTestWorktree function
- [ ] 4.3 Test ValidateWorktree function
- [ ] 4.4 Test CleanupWorktree function
- [ ] 4.5 Test worktree helper error handling
- [ ] 4.6 Create `test/helpers/shell_test.go`
- [ ] 4.7 Test ExecuteShellCommand function
- [ ] 4.8 Test GetShellPath function
- [ ] 4.9 Test CreateTempShellConfig function
- [ ] 4.10 Test shell helper error handling
- [ ] 4.11 Verify test/helpers coverage >70%

## 5. Validation

- [ ] 5.1 Run `mise run test:full` - all tests must pass
- [ ] 5.2 Run `mise run test:race` - race detector must pass
- [ ] 5.3 Generate coverage report and verify targets met
- [ ] 5.4 Run `mise run check` - all validation must pass

## Dependencies

**This change MUST be implemented AFTER all other pre-release changes are complete.**

This ensures:
- All features are in place before validation testing
- Coverage targets reflect final codebase state
- Edge case tests cover all implemented behaviors
