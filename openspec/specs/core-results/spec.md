# Capability: Service Results

## Purpose

Defines the result value objects returned by service-layer worktree and project operations. Types live in `internal/core/service_results.go`. Constructors use the `New` prefix where applicable; result objects are immutable post-construction.

## Requirements

### Requirement: CreateWorktreeResult

`core.CreateWorktreeResult` SHALL carry `Worktree core.Worktree` (the new worktree info), `HookResult *core.HookResult` (nil when no hooks ran), and `CreatedAt time.Time`. The hook result SHALL be `nil` when no hooks ran (not an empty struct) so callers can use `result.HookResult == nil` as the no-hooks check.

#### Scenario: Successful create, no hooks

- **WHEN** `CreateWorktree` returns and no `.twiggit.toml` exists
- **THEN** `result.Worktree.Path` SHALL equal the new worktree path
- **AND** `result.HookResult == nil`

### Requirement: PruneWorktreeResult (per worktree)

`core.PruneWorktreeResult` SHALL carry `ProjectName string`, `BranchName string`, `Path string`, `WasDeleted bool`, `WasBranchDeleted bool`, `Skipped bool`, and `SkipReason string` (empty when not skipped; one of `"protected branch"`, `"unmerged"`, `"current worktree"`, `"dirty"` when skipped).

#### Scenario: Protected branch skipped

- **WHEN** `prune myproject/main` runs and `main` is in `Config.Validation.ProtectedBranches`
- **THEN** the per-worktree result SHALL have `Skipped == true` and `SkipReason == "protected branch"`

### Requirement: PruneWorktreesResult (bulk)

`core.PruneWorktreesResult` SHALL carry `DeletedCount int`, `BranchDeletedCount int`, `SkippedCount int`, `Preview []core.Worktree` (populated on `--dry-run`), and `Skipped []core.PruneWorktreeResult` (per-worktree skip details).

#### Scenario: Bulk prune summary

- **WHEN** `twiggit prune --all` deletes three worktrees, deletes one branch, and skips two protected branches
- **THEN** `result.DeletedCount == 3`, `result.BranchDeletedCount == 1`, `result.SkippedCount == 2`
- **AND** `result.Preview == nil` (dry-run was not requested)

### Requirement: ProjectInfo and ProjectSummary

`core.ProjectInfo` SHALL carry `Name string`, `Path string`, `DefaultBranch string`, `WorktreeCount int`, and `LastModified time.Time`. `core.ProjectSummary` SHALL be the lightweight projection for `--output json` output over multiple projects.

#### Scenario: Project summary projection

- **WHEN** `ListProjects(...)` enumerates `$HOME/Projects/`
- **THEN** the result SHALL be a `[]core.ProjectSummary` with one entry per project
- **AND** each entry SHALL carry `Name`, `Path`, `DefaultBranch`, and `WorktreeCount`

### Requirement: WorktreeStatus

`core.WorktreeStatus` SHALL carry `Path string`, `BranchName string`, `IsClean bool`, `IsDetached bool`, `IsCurrent bool`, and `CommitHash string`. The `IsClean` flag is `true` when no uncommitted changes are present (including untracked files outside `.gitignore`); `IsDetached` is `true` when HEAD is not attached to a branch.

#### Scenario: Clean attached worktree

- **WHEN** `GetWorktreeStatus(path)` runs against a worktree with no uncommitted changes and HEAD attached to `feat/foo`
- **THEN** `WorktreeStatus.IsClean == true`, `IsDetached == false`, `BranchName == "feat/foo"`