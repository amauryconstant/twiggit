# Capability: Git Types

## Purpose

Defines the read-side value objects produced by the git client (`*git.reader`) and the path-traversal helpers used by the context resolver. Types live in `internal/core/git_types.go` and `internal/core/git_repo.go`. Enum-typed fields on these structs SHALL declare explicit `Unknown` sentinels at the start of their respective type declarations.

## Requirements

### Requirement: BranchInfo

`core.BranchInfo` SHALL carry `Name string`, `Hash string`, `IsCurrent bool`, `IsRemote bool`, and `Upstream string` (empty when no upstream is configured). The type is immutable post-construction.

#### Scenario: BranchInfo carries upstream

- **WHEN** `ListBranches(ctx, repoPath)` returns a local branch with upstream `origin/main`
- **THEN** `BranchInfo{Name: "feat/foo", Hash: "abc123", IsCurrent: true, IsRemote: false, Upstream: "origin/main"}` SHALL be returned

### Requirement: CommitInfo

`core.CommitInfo` SHALL carry `Hash string`, `ShortHash string` (first 7 chars), `AuthorName string`, `AuthorEmail string`, `CommitTime time.Time`, and `Message string` (first line of the commit body).

#### Scenario: CommitInfo short hash

- **WHEN** `GetCommitInfo(ctx, repoPath, "abc123def456789...")` runs
- **THEN** `CommitInfo.Hash == "abc123def456789..."` and `ShortHash == "abc123d"`

### Requirement: RemoteInfo

`core.RemoteInfo` SHALL carry `Name string` (e.g., `"origin"`) and `URLs []string` (one entry per fetch/push URL). `URLs` SHALL be a defensive copy (`slices.Clone`) of the read-side concrete's internal state.

#### Scenario: RemoteInfo URLs defensive copy

- **WHEN** `ListRemotes(ctx, repoPath)` returns a `core.RemoteInfo` with two URLs
- **AND** the caller appends a third URL to the returned slice
- **THEN** a second call to `ListRemotes` SHALL still return exactly the original two URLs (the underlying state was not mutated)

### Requirement: WorktreeInfo

`core.WorktreeInfo` SHALL carry `Path string`, `BranchName string` (empty for detached worktrees), `CommitHash string`, `IsMain bool` (true for the main worktree), `IsBare bool`, `IsDetached bool`, and `IsCurrent bool`. The `IsMain` flag is the inverse of the `cli-list` "exclude main worktree" rule when listing.

#### Scenario: WorktreeInfo for main worktree

- **WHEN** `ListWorktrees(...)` enumerates a project
- **THEN** the main worktree entry SHALL carry `IsMain == true`
- **AND** the cmd layer SHALL exclude entries where `IsMain == true` per `cli-list`

### Requirement: GitRepository and GitDir

`core.GitRepository` SHALL be an opaque value object carrying `Path string` (absolute path to the main repo) and `IsBare bool`. `core.GitDir` SHALL carry `Path string` (absolute path to `.git/` or `.git` file pointer). The `FindGitDirByTraversal(dir) (GitDir, error)` and `FindMainRepoByTraversal(dir) (GitRepository, error)` helpers SHALL walk up the filesystem until they locate the `.git` marker. `IsMainRepo(gitDir) bool` SHALL return `true` when the `.git` is a directory (not a file pointer) — the canonical main-repo marker.

#### Scenario: FindGitDirByTraversal walks upward

- **WHEN** `FindGitDirByTraversal("/home/u/Worktrees/myapp/feat/foo")` runs
- **THEN** it SHALL return `GitDir{Path: "/home/u/Worktrees/myapp/feat/foo/.git"}` (the worktree pointer file)
- **AND** `FindMainRepoByTraversal(...)` SHALL return `GitRepository{Path: "/home/u/Projects/myapp"}` (the main repo)
- **AND** `IsMainRepo(worktreeGitDir)` SHALL return `false` (pointer file, not directory)

#### Scenario: IsMainRepo true for main repo

- **WHEN** `gitDir = GitDir{Path: "/home/u/Projects/myapp/.git"}` and the path is a directory
- **THEN** `IsMainRepo(gitDir)` SHALL return `true`

### Requirement: RepositoryStatus

`core.RepositoryStatus` SHALL carry `IsClean bool`, `HasUntracked bool` (untracked files outside `.gitignore`), `ModifiedFiles []string` (path list of modified-but-not-staged files), and `StagedFiles []string` (path list of staged files). Per the `defensive-copy-exports` rule, both `ModifiedFiles` and `StagedFiles` SHALL be defensive copies (`slices.Clone`).

#### Scenario: Clean repo

- **WHEN** `GetRepositoryStatus(ctx, repoPath)` runs against a repo with no uncommitted or staged changes
- **THEN** `RepositoryStatus.IsClean == true`, `HasUntracked == false`, `len(ModifiedFiles) == 0`, `len(StagedFiles) == 0`

#### Scenario: Modified files defensive copy

- **WHEN** `GetRepositoryStatus(ctx, repoPath)` returns `RepositoryStatus.ModifiedFiles` with N entries
- **THEN** mutating the returned slice SHALL NOT affect the underlying git-client cache
- **AND** a second call to `GetRepositoryStatus` SHALL return a slice with the original contents (modulo real-time filesystem changes)