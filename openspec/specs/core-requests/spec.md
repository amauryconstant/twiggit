# Capability: Service Requests

## Purpose

Defines the request value objects for service-layer worktree and project operations. Types live in `internal/core/service_requests.go`. Constructors use the `New` prefix. No mutation methods exist (immutable post-construction).

## Requirements

### Requirement: CreateWorktreeRequest

`core.CreateWorktreeRequest` SHALL carry `ProjectName string`, `BranchName string`, `SourceBranch string` (the branch to create from; defaults to `Config.DefaultSourceBranch`), `WorktreePath string` (absolute path to the new worktree), and `RepoPath string` (absolute path to the main repository). The request SHALL be immutable; no `Set*` methods.

#### Scenario: Minimal request

- **WHEN** a caller constructs `CreateWorktreeRequest{ProjectName: "myapp", BranchName: "feat/x", SourceBranch: "main", WorktreePath: "...", RepoPath: "..."}`
- **THEN** every field SHALL be readable
- **AND** no `Set*` method SHALL exist on the type

### Requirement: DeleteWorktreeRequest

`core.DeleteWorktreeRequest` SHALL carry `ProjectName string`, `BranchName string`, `RepoPath string`, `Force bool` (bypass dirty check), and `KeepBranch bool` (do NOT delete the branch after worktree removal). The default value `KeepBranch == false` deletes the branch; `KeepBranch == true` preserves it.

#### Scenario: Default deletes branch

- **WHEN** a caller constructs `DeleteWorktreeRequest{Force: false, KeepBranch: false}`
- **THEN** the runner SHALL delete the worktree AND the corresponding branch

### Requirement: ListWorktreesRequest

`core.ListWorktreesRequest` SHALL carry `RepoPath string`, `ProjectName string` (optional; used for filtering), and `IncludeMainWorktree bool` (default `false`; the main worktree is excluded from listing). Returned slice fields SHALL be defensive copies of the loader's internal state.

#### Scenario: Main worktree excluded by default

- **WHEN** `ListWorktreesRequest{IncludeMainWorktree: false}` is dispatched
- **THEN** the returned worktree list SHALL NOT contain the main worktree entry

### Requirement: PruneWorktreesRequest

`core.PruneWorktreesRequest` SHALL carry `RepoPath string`, `Target string` (optional; empty means "all merged"), `DryRun bool`, `DeleteBranches bool`, `Force bool`, and `Yes bool` (skip confirmation). The runner SHALL honor these flags per `cli-prune`.

#### Scenario: Dry-run returns preview without deletion

- **WHEN** `PruneWorktreesRequest{DryRun: true, Target: ""}` is dispatched
- **THEN** the runner SHALL return a `PruneWorktreesResult` with `Deleted == 0` and `Preview []core.WorktreeInfo` populated

### Requirement: ResolvePathRequest

`core.ResolvePathRequest` SHALL carry `Identifier string` (the `branch` or `project/branch` form to resolve), `Cwd string` (current working directory for context detection), and `ForceRefresh bool` (bypass any context cache). Returns a `core.ResolutionResult` (owned by `core-context-types`).

#### Scenario: Resolve `feat/foo` from project context

- **WHEN** `ResolvePathRequest{Identifier: "feat/foo", Cwd: "$HOME/Projects/myapp"}` is dispatched
- **THEN** the runner SHALL return `ResolutionResult{Type: core.PathTypeWorktree, ProjectName: "myapp", BranchName: "feat/foo", ResolvedPath: "$HOME/Worktrees/myapp/feat/foo"}`