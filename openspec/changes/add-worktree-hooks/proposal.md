## Why

After creating a new worktree, developers often need to run setup commands (e.g., `mise trust` to enable tool version management). Currently this requires manual execution after each `twiggit create`. A hook system allows project-specific setup automation, reducing friction and ensuring consistent development environment configuration across worktrees.

## What Changes

- Add opt-in hook system triggered after worktree creation
- Support `.twiggit.toml` configuration file at repository root
- Execute configured commands in the new worktree directory
- Provide environment variables with context (branch name, paths, etc.)
- Report hook failures as warnings without rolling back the worktree

## Capabilities

### New Capabilities

- `worktree-hooks`: Hook system for executing commands after worktree lifecycle events, starting with `post-create`

### Modified Capabilities

- `worktree-management`: `CreateWorktree` operation now returns hook execution results alongside worktree info

## Impact

**New Files:**
- `internal/domain/hook_types.go` - Hook domain types
- `internal/infrastructure/hook_runner.go` - Hook execution implementation

**Modified Files:**
- `internal/domain/service_results.go` - Add `CreateWorktreeResult` type
- `internal/infrastructure/interfaces.go` - Add `HookRunner` interface
- `internal/application/interfaces.go` - Update `WorktreeService.CreateWorktree` signature
- `internal/service/worktree_service.go` - Inject and call `HookRunner`
- `cmd/create.go` - Display hook results/warnings

**Dependencies:**
- Uses existing `CommandExecutor` for command execution
- Uses koanf/toml parsing (already in use by ConfigManager)
