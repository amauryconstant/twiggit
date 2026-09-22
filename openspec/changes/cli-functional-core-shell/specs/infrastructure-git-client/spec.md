# Spec Delta

## MODIFIED Requirements

### Requirement: Routing table

The composite `git.GitClient` SHALL route operations as follows:

| Operation | GoGitClient | CLIClient | Rationale |
|---|:---:|:---:|---|
| `OpenRepository` | Yes | No | Portable, deterministic |
| `ListBranches` | Yes | No | Portable, deterministic |
| `BranchExists` | Yes | No | Portable, deterministic |
| `GetRepositoryStatus` | Yes | No | Portable, deterministic |
| `ValidateRepository` | Yes | No | Portable, deterministic |
| `GetRepositoryInfo` | Yes | No | Portable, deterministic |
| `ListRemotes` | Yes | No | Portable, deterministic |
| `GetCommitInfo` | Yes | No | Portable, deterministic |
| `CreateWorktree` | No | Yes | go-git lacks support |
| `DeleteWorktree` | No | Yes | go-git lacks support |
| `ListWorktrees` | No | Yes | go-git lacks support |
| `PruneWorktrees` | No | Yes | go-git lacks support |
| `IsBranchMerged` | No | Yes | go-git limitations |
| `DeleteBranch` | No | Yes | Handles worktree-referenced branches |

The composite `git.GitClient` is the only injection point. Callers SHALL NOT reach into a single implementation (`git.GoGitClient` or `git.CLIClient`) directly. The previous rule that "No composite umbrella interface SHALL exist; the two role interfaces are the only injection points" is removed; the role interfaces are internal collaborators and are not exported as the public API.

#### Scenario: Service injects both role interfaces directly
- **WHEN** `WorktreeService`, `ProjectService`, or `NavigationService` is constructed (this scenario is REMOVED semantically: services now inject the composite; the previous two-role-interface pattern is consolidated)
- **THEN** the call site SHALL be migrated to inject `*git.GitClient` instead of `application.GoGitClient` + `application.CLIClient`

#### Scenario: Read operation uses the go-git role
- **WHEN** `client.BranchExists(ctx, projectPath, branchName)` is invoked
- **THEN** it SHALL dispatch through the embedded `GoGitClient`
- **AND** it SHALL NOT call into `CLIClient`

#### Scenario: Worktree mutation uses the CLI role
- **WHEN** `client.CreateWorktree(ctx, req)` is invoked
- **THEN** it SHALL dispatch through the embedded `CLIClient`
- **AND** it SHALL NOT call into `GoGitClient`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `OpenRepository`

`OpenRepository(path)` SHALL return a `*git.Repository` after normalizing the path to absolute. Results SHALL be cached in an LRU cache keyed by absolute path. The implementation lives in `internal/git/reader.go` (migrated from `internal/infrastructure/gogit_client.go`).

#### Scenario: Cache hit
- **WHEN** `OpenRepository` is called twice with the same path
- **THEN** the second call SHALL return the cached `*git.Repository`
- **AND** SHALL NOT re-open the repo

#### Scenario: Cache miss
- **WHEN** `OpenRepository` is called with a new path
- **THEN** system SHALL call `git.PlainOpen(absPath)` and cache the result

### Requirement: `ListBranches`

`ListBranches(ctx, repoPath)` SHALL return `[]BranchInfo` for every local branch, with `Name`, `IsCurrent`, `Commit`, `Author`, `Date`, and (when applicable) `Remote`.

#### Scenario: Current branch flag
- **WHEN** `ListBranches` iterates branches
- **THEN** the branch matching `repo.Head()` SHALL have `IsCurrent = true`

#### Scenario: Remote tracking
- **WHEN** a branch has a matching `refs/remotes/origin/<name>` reference
- **THEN** `BranchInfo.Remote` SHALL be set to `origin/<name>`

### Requirement: `BranchExists`

`BranchExists(ctx, repoPath, branchName)` SHALL return `(true, nil)` when `refs/heads/<branchName>` resolves, `(false, nil)` when it returns `plumbing.ErrReferenceNotFound`, and a `git.ExternalError` (walking to `*core.OperationError`) for any other failure.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `GetRepositoryStatus`

`GetRepositoryStatus(ctx, repoPath)` SHALL return `RepositoryStatus` with `IsClean`, `Branch`, `Commit`, and per-category file lists (`Modified`, `Added`, `Deleted`, `Untracked`).

#### Scenario: Workaround for staged-only files
- **WHEN** go-git reports a file as `Added` in staging but `Unmodified` in the worktree (a known go-git bug for worktrees)
- **THEN** system SHALL drop that file from `Added` and SHALL NOT mark the repo dirty

#### Scenario: Workaround for missing HEAD in worktree
- **WHEN** `repo.Head()` fails with "reference not found"
- **THEN** system SHALL still return a valid `RepositoryStatus`
- **AND** `Branch` SHALL be `"unknown"`, `Commit` SHALL be `""`

### Requirement: `ValidateRepository`

`ValidateRepository(path)` SHALL return `nil` when `git.PlainOpen(path)` succeeds, or a `git.ExternalError` (walking to `*core.OperationError`) with message "not a valid git repository" otherwise.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `GetRepositoryInfo`

`GetRepositoryInfo(ctx, repoPath)` SHALL return a `*GitRepository` combining branches, remotes, status, and the default branch (`main` or `master`, whichever exists).

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `ListRemotes`

`ListRemotes(ctx, repoPath)` SHALL return `[]RemoteInfo` with `Name`, `FetchURL`, and `PushURL` for every configured remote.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: `GetCommitInfo`

`GetCommitInfo(ctx, repoPath, commitHash)` SHALL return a `*CommitInfo` with `Hash`, `ShortHash` (7 chars), `Author`, `Email`, `Date`, and `Message`. The function SHALL return a `git.ExternalError` (walking to `*core.OperationError`) naming the commit hash when the commit does not exist.

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Cache configuration

`git.NewClient(opts ...ClientOption) (*git.GitClient, error)` SHALL create a composite client with cache enabled (default size 25). `git.NewClient(git.WithCacheSize(n))` SHALL allow custom sizes; `n <= 0` SHALL fall back to 25. `git.NewClient(git.WithCacheDisabled())` SHALL bypass the cache entirely. The constructor SHALL return a `*core.OperationError` wrapping the LRU cache initialization failure when cache init fails. The previous name `git.NewGoGitClient()` and the previous `application.GoGitClient` interface are removed.

#### Scenario: Successful construction
- **WHEN** `git.NewClient()` is called under normal conditions
- **THEN** it SHALL return `(client, nil)` where `client != nil`

#### Scenario: LRU allocation failure surfaces as error
- **WHEN** `git.NewClient(git.WithCacheSize(n))` is called
- **AND** the LRU cache allocator returns a non-nil error
- **THEN** the constructor SHALL return `(nil, error)`
- **AND** callers SHALL propagate the error
- **AND** the constructor SHALL NOT silently return a client with a nil cache

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: CLI worktree operations

`git.GitClient` SHALL implement worktree/branch mutations via the embedded `CLIClient` (lives in `internal/git/writer.go`):

- `CreateWorktree(ctx, repoPath, branch, source, path)`
- `DeleteWorktree(ctx, repoPath, path, force)`
- `ListWorktrees(ctx, repoPath)`
- `PruneWorktrees(ctx, repoPath)`
- `IsBranchMerged(ctx, repoPath, branch)`
- `DeleteBranch(ctx, repoPath, branch)` — handles branches currently checked out by other worktrees via `git branch -d`

#### Scenario: DeleteBranch handles checked-out branches
- **WHEN** `DeleteBranch` is called on a branch that is the `HEAD` of another worktree
- **THEN** system SHALL still delete the branch via `git branch -d` which handles the worktree reference correctly

### Requirement: Error wrapping

All failures SHALL be wrapped via `git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError` (defined in `internal/git/errors.go`). These constructors return `*git.ExternalError` whose embedded core type walks to `*core.OperationError`. The composite client SHALL NOT return raw `go-git` or `os/exec` errors. The previous `domain.NewGit*Error` constructors are removed.

#### Scenario: Failure walks to *core.OperationError
- **WHEN** `git.PlainOpen(path)` fails
- **THEN** `errors.As(err, &*core.OperationError{})` returns `true` with `Op = "git.open"`, `Message` naming the path, and `Cause` set via `%w`

#### Scenario: Definition holds
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
