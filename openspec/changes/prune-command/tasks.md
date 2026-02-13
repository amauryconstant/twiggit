## 1. Domain Layer

- [x] 1.1 Add `PruneWorktreesRequest` type to `internal/domain/service_requests.go`
- [x] 1.2 Add `PruneWorktreesResult` type to `internal/domain/service_requests.go`
- [x] 1.3 Add `PruneWorktreeResult` type to `internal/domain/service_requests.go`
- [x] 1.4 Add `ProtectedBranchError` type to `internal/domain/service_errors.go`
- [x] 1.5 Add `ProtectedBranches` field to `ValidationConfig` in `internal/domain/config.go`

## 2. Infrastructure Layer

- [x] 2.1 Add `DeleteBranch()` method signature to `GoGitClient` interface in `internal/infrastructure/interfaces.go`
- [x] 2.2 Implement `DeleteBranch()` method in `internal/infrastructure/gogit_client.go` using go-git's `DeleteReference()` with structured error handling
- [x] 2.3 Add validation logic to check if branch is current HEAD before deletion
- [x] 2.4 Add validation logic to check if branch exists before deletion
- [x] 2.5 Return appropriate error for empty repository path or empty branch name

## 3. Application Layer

- [x] 3.1 Add `PruneMergedWorktrees()` method signature to `WorktreeService` interface in `internal/application/interfaces.go`
- [x] 3.2 Implement `PruneMergedWorktrees()` method in `internal/service/worktree_service.go` with structured error handling
- [x] 3.3 Add logic to identify merged worktrees using existing `IsBranchMerged()`
- [x] 3.4 Add logic to filter out main worktree and protected branches
- [x] 3.5 Add logic to handle `--dry-run` flag (preview only, no deletion)
- [x] 3.6 Add logic to handle `--force` flag (bypass uncommitted changes check and confirmation)
- [x] 3.7 Add logic to handle `--delete-branches` flag (call `DeleteBranch()`)
- [x] 3.8 Add logic to handle `--all` flag (prune across all projects)
- [x] 3.9 Add user confirmation prompt for bulk mode (`--all` without `--force`)
- [x] 3.10 Add navigation output to stdout for single-worktree prune (project directory)
- [x] 3.11 Add protected branch validation before branch deletion
- [x] 3.12 Add comprehensive error handling with operation summary
- [x] 3.13 Implement non-blocking branch deletion failures (continue with worktree cleanup)
- [x] 3.14 Add current worktree validation (cannot prune worktree we're currently in)

## 4. CLI Layer

- [x] 4.1 Create `cmd/prune.go` command file with Cobra command definition
- [x] 4.2 Add `--force` flag for bypassing safety checks and confirmation
- [x] 4.3 Add `--delete-branches` flag for optional branch deletion
- [x] 4.4 Add `--all` flag for bulk pruning across all projects
- [x] 4.5 Add `--dry-run` flag for preview mode
- [x] 4.6 Add context-aware project inference (worktree > project > outside git)
- [x] 4.7 Register prune command in `cmd/root.go`
- [x] 4.8 Add argument support for specific worktree specification (project/branch format)
- [x] 4.9 Add shell wrapper path output for single-worktree prune

## 5. Unit Tests

- [x] 5.1 Add unit tests for `ProtectedBranchError` type
- [x] 5.2 Add unit tests for `DeleteBranch()` method in `gogit_client.go` using Testify patterns
- [x] 5.3 Add unit tests for branch validation (current HEAD, existence, empty inputs)
- [x] 5.4 Add unit tests for protected branch filtering logic
- [x] 5.5 Add unit tests for merged worktree identification
- [x] 5.6 Add unit tests for dry-run mode
- [x] 5.7 Add unit tests for force flag behavior
- [x] 5.8 Add unit tests for delete-branches flag behavior
- [x] 5.9 Add unit tests for bulk mode (--all) with confirmation
- [x] 5.10 Add unit tests for navigation output logic
- [x] 5.11 Add unit tests for current worktree validation (3.14)

## 6. Integration Tests

- [x] 6.1 Convert existing tests to Testify suite pattern per AGENTS.md
- [x] 6.2 Add integration test for uncommitted changes detection (skipped without --force)
- [x] 6.3 Add integration test for --force bypassing uncommitted changes check
- [x] 6.4 Add integration test for current worktree validation (cannot prune active worktree)
- [x] 6.5 Add integration test for navigation output with proper config setup
- [x] 6.6 Add integration test for operation summary verification

## 7. E2E Tests

- [x] 7.1 Add E2E test for `twiggit prune --dry-run` in project context
- [x] 7.2 Add E2E test for `twiggit prune --dry-run --all` across projects
- [x] 7.3 Add E2E test for `twiggit prune myproject/feature-branch` (single worktree)
- [x] 7.4 Add E2E test for `twiggit prune --delete-branches myproject/feature-branch`
- [x] 7.5 Add E2E test for `twiggit prune --force` with uncommitted changes
- [x] 7.6 Add E2E test for protected branch protection (attempt to delete protected branch)
- [x] 7.7 Add E2E test for navigation output (stdout path for shell wrapper)
- [x] 7.8 Add E2E test for error handling (invalid worktree, unmerged branch)
- [x] 7.9 Add E2E test for context-aware behavior (run from worktree vs project vs outside git)
- [x] 7.10 Add E2E test for operation summary display
- [x] 7.11 Add stdin support methods to CLI helper (`RunWithStdin`, `RunWithStdinAndDir`)
- [x] 7.12 Add stdin context method to fixture helper (`FromOutsideGitWithStdin`)
- [x] 7.13 Add merged worktree fixture method (`CreateMergedWorktreeSetup`)
- [x] 7.14 Add E2E test for confirmation prompt with `--all --delete-branches` (stdin "y" acceptance)
- [x] 7.15 Add E2E test for confirmation cancellation ("n" and empty response)
- [x] 7.16 Add E2E test for `--force` bypassing confirmation prompt
