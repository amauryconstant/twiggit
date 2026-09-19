# Capability: Worktree Service

## Purpose

Service-layer orchestration for worktree lifecycle operations
(create, list, delete, prune), helper lookups (status, validation,
branch existence, merged check, by-path), and concurrency safety.

## Requirements

### Requirement: Service contract surface

The `WorktreeService` interface SHALL expose the following methods:

- `CreateWorktree(ctx, *CreateWorktreeRequest) (*CreateWorktreeResult, error)`
- `DeleteWorktree(ctx, *DeleteWorktreeRequest) error`
- `ListWorktrees(ctx, *ListWorktreesRequest) ([]*WorktreeInfo, error)`
- `GetWorktreeStatus(ctx, worktreePath) (*WorktreeStatus, error)`
- `ValidateWorktree(ctx, worktreePath) error`
- `PruneMergedWorktrees(ctx, *PruneWorktreesRequest) (*PruneWorktreesResult, error)`
- `BranchExists(ctx, projectPath, branchName) (bool, error)`
- `IsBranchMerged(ctx, worktreePath, branchName) (bool, error)`
- `GetWorktreeByPath(ctx, projectPath, worktreePath) (*WorktreeInfo, error)`

Implementation lives in `internal/service/` and SHALL satisfy the
`application.WorktreeService` interface.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Concurrency safety

Per-project worktree mutations SHALL be protected by a per-project
mutex so that concurrent `delete`, `prune`, or `create` operations on
the same project SHALL NOT race. Operations on different projects
SHALL proceed concurrently.

#### Scenario: Same-project concurrent mutations

- **WHEN** two goroutines simultaneously invoke
  `WorktreeService.DeleteWorktree` on the same project
- **THEN** operations SHALL serialize on a per-project mutex
- **AND** no data race SHALL occur on the per-project worktree set

#### Scenario: Cross-project concurrency

- **WHEN** two goroutines mutate worktrees on different projects
- **THEN** the operations SHALL proceed concurrently
- **AND** no global lock SHALL be held

### Requirement: Status inspection

`GetWorktreeStatus(ctx, worktreePath)` SHALL return the worktree's
clean/dirty state, current branch, and HEAD commit, reading from the
git client. See `infrastructure-git-client.GetRepositoryStatus`.

#### Scenario: Clean worktree

- **WHEN** worktree has no uncommitted changes
- **THEN** `WorktreeStatus.IsClean` SHALL be `true`
- **AND** `WorktreeStatus.Branch` SHALL equal the current branch name

#### Scenario: Missing worktree

- **WHEN** `worktreePath` does not exist
- **THEN** system SHALL return a not-found error (exit code 6)

### Requirement: Validation

`ValidateWorktree(ctx, worktreePath)` SHALL return nil if the worktree
is properly configured, or a `domain.ValidationError` if it is not.

#### Scenario: Valid worktree

- **WHEN** `worktreePath` points to a real worktree
- **THEN** `ValidateWorktree` SHALL return nil

#### Scenario: Invalid path

- **WHEN** `worktreePath` is empty or not a git worktree
- **THEN** `ValidateWorktree` SHALL return a validation error

### Requirement: Branch existence and merged check

`BranchExists` SHALL use the git client's branch-existence check;
`IsBranchMerged` SHALL use the CLI-backed merged check
(see `infrastructure-git-client` routing).

#### Scenario: Branch exists

- **WHEN** `BranchExists(ctx, project, "main")` is called
- **AND** `main` exists
- **THEN** it SHALL return `(true, nil)`

#### Scenario: Branch merged

- **WHEN** `IsBranchMerged(ctx, worktree, "feature")` is called
- **AND** `feature` is merged into the base branch
- **THEN** it SHALL return `(true, nil)`

### Requirement: Path-based lookup

`GetWorktreeByPath(ctx, projectPath, worktreePath)` SHALL return the
`WorktreeInfo` for the worktree at `worktreePath` within the project,
or a not-found error if no such worktree exists.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Service receives git client and config via constructor

The `WorktreeService` constructor SHALL accept the `GitClient`
(composite), `ProjectService`, and `*domain.Config` as injected
dependencies. No globals, no `init()`. See `application-service-interfaces`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
