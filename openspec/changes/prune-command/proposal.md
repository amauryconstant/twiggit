## Why

Users need efficient post-merge cleanup for merged feature branches. After merging a PR, users must manually delete worktrees and branches, which becomes tedious across multiple projects. A bulk `prune` command streamlines this workflow by automatically deleting merged worktrees with optional branch deletion.

## What Changes

- Add new `prune` command for context-aware bulk deletion of merged worktrees
- Add `DeleteBranch()` method to `GoGitClient` interface (go-git implementation)
- Add protected branch configuration (default: main, master, develop, staging, production)
- Add `PruneMergedWorktrees()` method to `WorktreeService` interface
- Add domain types: `PruneWorktreesRequest`, `PruneWorktreesResult`, `PruneWorktreeResult`
- Add `ProtectedBranchError` error type for protected branch deletion attempts
- Support flags: `--force`, `--delete-branches`, `--all`, `--dry-run`
- Require user confirmation for `--all` bulk mode
- Navigate to project directory after single-worktree prune (context-aware behavior like `delete -C`)

## Capabilities

### New Capabilities
- `worktree-pruning`: Context-aware bulk deletion of merged worktrees for post-merge cleanup with protected branch safety
- `branch-deletion`: Delete branch references using go-git library

### Modified Capabilities
- `worktree-management`: Adding prune command as new operation in worktree management capability set

## Impact

**Affected code:**
- `cmd/`: New `prune.go` command file
- `cmd/root.go`: Register prune command
- `internal/infrastructure/interfaces.go`: Add `DeleteBranch()` to GoGitClient
- `internal/infrastructure/gogit_client.go`: Implement `DeleteBranch()` with domain error wrapping
- `internal/application/interfaces.go`: Add `PruneMergedWorktrees()` to WorktreeService
- `internal/service/worktree_service.go`: Implement prune logic with protected branch safety
- `internal/domain/config.go`: Add `ProtectedBranches` to ValidationConfig
- `internal/domain/service_requests.go`: Add prune request/result types
- `internal/domain/service_errors.go`: Add ProtectedBranchError
- Test files: Unit, integration, E2E tests using Testify patterns with centralized mocks

**New dependencies:** None (uses existing go-git library)

**Breaking changes:** None (new command only)
