# Spec Delta: Infrastructure Git Client

## MODIFIED Requirements

### Requirement: Operation routing table

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

This routing table is canonical. Callers SHALL select the role
interface that owns the operation (`GoGitClient` for read paths,
`CLIClient` for write paths). No composite umbrella interface
SHALL exist; the two role interfaces are the only injection points.

#### Scenario: Service injects both role interfaces directly

- **WHEN** `WorktreeService`, `ProjectService`, or `NavigationService`
  is constructed
- **THEN** its constructor SHALL accept `application.GoGitClient` and
  `application.CLIClient` as two distinct arguments
- **AND** callers SHALL NOT depend on any composite `GitClient`
  interface

#### Scenario: Read operation uses the go-git role

- **WHEN** `service.BranchExists(ctx, projectPath, branchName)` is
  invoked
- **THEN** it SHALL dispatch through the `GoGitClient` field
- **AND** it SHALL NOT call into `CLIClient`

#### Scenario: Worktree mutation uses the CLI role

- **WHEN** `service.CreateWorktree(ctx, req)` is invoked
- **THEN** it SHALL dispatch through the `CLIClient` field
- **AND** it SHALL NOT call into `GoGitClient`

### Requirement: GoGitClient constructor returns error

`NewGoGitClient()` SHALL create a client with cache enabled
(default size 25). `NewGoGitClientWithSize(n)` SHALL allow custom
sizes; `n <= 0` SHALL fall back to 25. `cacheEnabled=false` SHALL
bypass the cache entirely. Both constructors SHALL return
`(*GoGitClient, error)`; the error SHALL be non-nil when the
underlying LRU cache cannot be allocated.

#### Scenario: Successful construction

- **WHEN** `NewGoGitClient()` is called under normal conditions
- **THEN** it SHALL return `(client, nil)` where `client != nil`

#### Scenario: LRU allocation failure surfaces as error

- **WHEN** `NewGoGitClientWithSize(n)` is called
- **AND** the LRU cache allocator returns a non-nil error
- **THEN** the constructor SHALL return `(nil, error)`
- **AND** callers SHALL propagate the error
- **AND** the constructor SHALL NOT silently return a client with a
  nil cache
