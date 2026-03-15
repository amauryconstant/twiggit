# Suggestion Tracking

This file tracks optional improvements identified during verification.

## 2026-03-15 - PHASE2 Verification

- [ ] **[test]** Add test coverage for exit code 3 (config errors)
  - Location: `test/e2e/error_clarity_test.go`
  - Impact: Medium
  - Notes: Tests verify exit codes 0, 1, 2, and 5, but not exit code 3 for configuration errors

- [ ] **[test]** Add test coverage for exit code 4 (git errors)
  - Location: `test/e2e/error_clarity_test.go`
  - Impact: Medium
  - Notes: Tests verify exit codes 0, 1, 2, and 5, but not exit code 4 for git operation errors

- [ ] **[test]** Add test coverage for exit code 6 (not-found errors)
  - Location: `test/e2e/error_clarity_test.go`
  - Impact: Medium
  - Notes: Tests verify exit codes 0, 1, 2, and 5, but not exit code 6 for resource not found errors

- [ ] **[reliability]** Improve IsNotFound() reliability beyond message content
  - Location: `internal/domain/service_errors.go:132-136`
  - Impact: Medium
  - Notes: `WorktreeServiceError.IsNotFound()` checks message content for "not found" or "does not exist". This string matching is fragile and may miss some not-found cases. Consider adding explicit error types or using error wrapping more reliably.

- [ ] **[consistency]** Add IsNotFound() to ProjectServiceError
  - Location: `internal/domain/service_errors.go`
  - Impact: Low
  - Notes: Only `WorktreeServiceError` has `IsNotFound()` implemented. `ProjectServiceError` and `NavigationServiceError` don't have this method. Not critical as git-level errors (`GitRepositoryError`, `GitWorktreeError`) have it, but would improve consistency.

## 2026-03-15 - PHASE2 Verification

- [ ] **[test]** Fix E2E test expectations for cd command
  - Location: test/e2e/error_clarity_test.go:86
  - Impact: Low - test failure doesn't affect correctness
  - Notes: Test expects exit code 2 (Cobra usage error) for `twiggit cd` with no arguments, but cd command uses `Args: cobra.MaximumNArgs(1)` which allows 0 arguments. Current behavior returns exit code 1 for "project 'main' not found". Either update test expectation to exit code 1, or change command to require exactly 1 argument.

- [ ] **[test]** Fix E2E test for prune --all with specific worktree
  - Location: test/e2e/prune_test.go:77
  - Impact: Low - test failure doesn't affect correctness
  - Notes: Test expects exit code 5 (validation error) for `prune --all test/feature-1`, but without --force flag, command prompts for confirmation and exits with code 1 for "failed to read confirmation: EOF". Add --force flag to test to bypass confirmation and trigger validation error.
