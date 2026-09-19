# Capability: Git Client

## Purpose

Composite git client that routes read-only and branch-existence
operations to `GoGitClient` (pure go-git, deterministic, thread-safe)
and worktree/branch-mutation operations to `CLIClient` (git CLI
subprocess, required for operations go-git does not support).
All implementations SHALL be idempotent and thread-safe.

## Requirements

### Requirement: Routing table

The composite client SHALL route operations as follows:

| Operation | GoGitClient | CLIClient | Rationale |
|---|:---:|:---:|---|
| `OpenRepository` | ✅ | ❌ | Portable, deterministic |
| `ListBranches` | ✅ | ❌ | Portable, deterministic |
| `BranchExists` | ✅ | ❌ | Portable, deterministic |
| `GetRepositoryStatus` | ✅ | ❌ | Portable, deterministic |
| `ValidateRepository` | ✅ | ❌ | Portable, deterministic |
| `GetRepositoryInfo` | ✅ | ❌ | Portable, deterministic |
| `ListRemotes` | ✅ | ❌ | Portable, deterministic |
| `GetCommitInfo` | ✅ | ❌ | Portable, deterministic |
| `CreateWorktree` | ❌ | ✅ | go-git lacks support |
| `DeleteWorktree` | ❌ | ✅ | go-git lacks support |
| `ListWorktrees` | ❌ | ✅ | go-git lacks support |
| `PruneWorktrees` | ❌ | ✅ | go-git lacks support |
| `IsBranchMerged` | ❌ | ✅ | go-git limitations |
| `DeleteBranch` | ❌ | ✅ | Handles worktree-referenced branches |

This routing table is canonical. Callers SHALL NOT reach into a single
implementation; the composite `GitClient` is the only injection point.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `OpenRepository`

`OpenRepository(path)` SHALL return a `*git.Repository` after
normalizing the path to absolute. Results SHALL be cached in an LRU
cache keyed by absolute path.

#### Scenario: Cache hit

- **WHEN** `OpenRepository` is called twice with the same path
- **THEN** the second call SHALL return the cached `*git.Repository`
- **AND** SHALL NOT re-open the repo

#### Scenario: Cache miss

- **WHEN** `OpenRepository` is called with a new path
- **THEN** system SHALL call `git.PlainOpen(absPath)` and cache the result

### Requirement: `ListBranches`

`ListBranches(ctx, repoPath)` SHALL return `[]BranchInfo` for every
local branch, with `Name`, `IsCurrent`, `Commit`, `Author`, `Date`,
and (when applicable) `Remote`.

#### Scenario: Current branch flag

- **WHEN** `ListBranches` iterates branches
- **THEN** the branch matching `repo.Head()` SHALL have `IsCurrent = true`

#### Scenario: Remote tracking

- **WHEN** a branch has a matching `refs/remotes/origin/<name>` reference
- **THEN** `BranchInfo.Remote` SHALL be set to `origin/<name>`

### Requirement: `BranchExists`

`BranchExists(ctx, repoPath, branchName)` SHALL return `(true, nil)`
when `refs/heads/<branchName>` resolves, `(false, nil)` when it
returns `plumbing.ErrReferenceNotFound`, and a `GitRepositoryError`
for any other failure.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `GetRepositoryStatus`

`GetRepositoryStatus(ctx, repoPath)` SHALL return `RepositoryStatus`
with `IsClean`, `Branch`, `Commit`, and per-category file lists
(`Modified`, `Added`, `Deleted`, `Untracked`).

#### Scenario: Workaround for staged-only files

- **WHEN** go-git reports a file as `Added` in staging but `Unmodified`
  in the worktree (a known go-git bug for worktrees)
- **THEN** system SHALL drop that file from `Added` and SHALL NOT mark
  the repo dirty

#### Scenario: Workaround for missing HEAD in worktree

- **WHEN** `repo.Head()` fails with "reference not found"
- **THEN** system SHALL still return a valid `RepositoryStatus`
- **AND** `Branch` SHALL be `"unknown"`, `Commit` SHALL be `""`

### Requirement: `ValidateRepository`

`ValidateRepository(path)` SHALL return `nil` when `git.PlainOpen(path)`
succeeds, or a `GitRepositoryError` with message "not a valid git
repository" otherwise.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `GetRepositoryInfo`

`GetRepositoryInfo(ctx, repoPath)` SHALL return a `*GitRepository`
combining branches, remotes, status, and the default branch
(`main` or `master`, whichever exists).



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `ListRemotes`

`ListRemotes(ctx, repoPath)` SHALL return `[]RemoteInfo` with `Name`,
`FetchURL`, and `PushURL` for every configured remote.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `GetCommitInfo`

`GetCommitInfo(ctx, repoPath, commitHash)` SHALL return a
`*CommitInfo` with `Hash`, `ShortHash` (7 chars), `Author`, `Email`,
`Date`, and `Message`. The function SHALL return a `GitRepositoryError`
naming the commit hash when the commit does not exist.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Cache configuration

`NewGoGitClient()` SHALL create a client with cache enabled (default
size 25). `NewGoGitClientWithSize(n)` SHALL allow custom sizes;
`n <= 0` SHALL fall back to 25. `cacheEnabled=false` SHALL bypass the
cache entirely.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: CLI worktree operations

`CLIClient` SHALL implement worktree/branch mutations via the git CLI:

- `CreateWorktree(ctx, repoPath, branch, source, path)`
- `DeleteWorktree(ctx, repoPath, path, force)`
- `ListWorktrees(ctx, repoPath)`
- `PruneWorktrees(ctx, repoPath)`
- `IsBranchMerged(ctx, repoPath, branch)`
- `DeleteBranch(ctx, repoPath, branch)` — handles branches currently
  checked out by other worktrees via `git branch -d`

#### Scenario: DeleteBranch handles checked-out branches

- **WHEN** `DeleteBranch` is called on a branch that is the
  `HEAD` of another worktree
- **THEN** system SHALL still delete the branch via `git branch -d`
  which handles the worktree reference correctly

### Requirement: Error wrapping

All failures SHALL be wrapped via `domain.NewGitRepositoryError` /
`domain.NewGitWorktreeError` / `domain.NewGitCommandError` per the
infrastructure AGENTS error-wrapping rules. The composite client
SHALL NOT return raw `go-git` or `os/exec` errors.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
