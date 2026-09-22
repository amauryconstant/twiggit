# Proposal: Interface Segregation

## Why

After `cli-functional-core-shell` lands, service constructors in `internal/core/` take `core.GoGitClient` (8 methods) and `core.CLIClient` (6 methods) as role interfaces. The `golang-structs-interfaces` skill explicitly identifies 1-3 method interfaces as the optimal shape; both `GoGitClient` and `CLIClient` violate this rule with concrete downsides:

- Tests must implement 8 or 6 methods even when exercising a single operation. A test for `BranchExists` alone still implements every other method on the bundle.
- New role requirements (for example adding `WithExistingOnly` to a read path) bloat the existing role rather than spawning a new focused interface.
- Service constructors take two large role interfaces where consumers of just read operations should not depend on write methods, and vice versa.

This change segregates `GoGitClient` and `CLIClient` into 1-3 method role interfaces, placed consumer-side per the `golang-structs-interfaces` skill (in `internal/core/git.go`), so each service receives only the roles it consumes.

## What Changes

### Role interfaces (consumer-side, in `internal/core/git.go`)

Split `GoGitClient` (8 methods) into:

- `core.RepositoryOpener` (2 methods): `OpenRepository(path)`, `ValidateRepository(path)`.
- `core.BranchReader` (2 methods): `ListBranches(ctx, repoPath)`, `BranchExists(ctx, repoPath, branchName)`.
- `core.RepositoryReader` (3 methods): `GetRepositoryStatus(ctx, repoPath)`, `GetRepositoryInfo(ctx, repoPath)`, `GetCommitInfo(ctx, repoPath, commitHash)`.
- `core.RemoteReader` (1 method): `ListRemotes(ctx, repoPath)`.

Split `CLIClient` (6 methods) into:

- `core.WorktreeWriter` (4 methods, documented skill-rule violation): `CreateWorktree(ctx, repoPath, branch, source, path)`, `DeleteWorktree(ctx, repoPath, path, force)`, `ListWorktrees(ctx, repoPath)`, `PruneWorktrees(ctx, repoPath)`.
- `core.BranchWriter` (2 methods): `DeleteBranch(ctx, repoPath, branch)`, `IsBranchMerged(ctx, repoPath, branch)`.

### Service constructor signatures

Service constructors take the new role interfaces as separate fields. The exact signatures:

- `core.NewWorktreeService(branchReader core.BranchReader, repoReader core.RepositoryReader, worktreeWriter core.WorktreeWriter, branchWriter core.BranchWriter, projectService core.ProjectService, config *core.Config, hookRunner core.HookRunner) core.WorktreeService`.
- `core.NewProjectService(repoOpener core.RepositoryOpener, repoReader core.RepositoryReader, remoteReader core.RemoteReader, repoLocator core.RepoLocator, config *core.Config) core.ProjectService`.
- `core.NewNavigationService(repoLocator core.RepoLocator, contextResolver core.ContextResolver, config *core.Config) core.NavigationService`.
- `core.NewContextResolver(repoOpener core.RepositoryOpener, repoReader core.RepositoryReader, config *core.Config) core.ContextResolver`.

### Factory wiring

`internal/cmdutil/factory.go` exposes per-role lazy function fields:

```go
type Factory struct {
    IOStreams       *iostreams.IOStreams
    Config          func() (*config.AppConfig, error)
    RepoOpener      func() (core.RepositoryOpener, error)
    BranchReader    func() (core.BranchReader, error)
    RepoReader      func() (core.RepositoryReader, error)
    RemoteReader    func() (core.RemoteReader, error)
    WorktreeWriter  func() (core.WorktreeWriter, error)
    BranchWriter    func() (core.BranchWriter, error)
    RepoLocator     func() (core.RepoLocator, error)
    HookRunner      func() (core.HookRunner, error)
    ContextResolver func() (core.ContextResolver, error)
    Logger          func() *slog.Logger
}
```

Each lazy function returns the concrete `internal/git.GoGitClient` or `internal/git.CLIClient` typed as the requested role (Go interface satisfaction). A `sync.Once` cache inside the factory prevents re-initialization within one command execution.

### Mock rewrites (testify, hand-written)

- Delete `MockGitClientBundle` from `test/mocks/git_service_mock.go`.
- Split into per-role mock structs:
  - `MockRepositoryOpener` (2 methods).
  - `MockBranchReader` (2 methods).
  - `MockRepositoryReader` (3 methods).
  - `MockRemoteReader` (1 method).
  - `MockWorktreeWriter` (4 methods).
  - `MockBranchWriter` (2 methods).
- Mechanical rename across `test/integration/`, `test/concurrent/`, `test/e2e/fixtures/`: each test that constructed one bundle now constructs the subset of mocks matching its command's consumed roles.

### Concrete client satisfaction

The concrete `internal/git.GoGitClient` (in `internal/git/gogit_client.go`) and `internal/git.CLIClient` (in `internal/git/cli_client.go`) satisfy all roles via their existing method sets. No implementation changes needed; only the interface declarations move.

### RepoLocator unchanged

`core.RepoLocator` is already a single-method interface (added in `architecture-layer-inversion`); no segregation work needed.

## Capabilities

### Modified Capabilities

- `git-client` (modifies `infrastructure-git-client`) — replace the composite `GoGitClient` (8 methods) and `CLIClient` (6 methods) surface descriptions with the role segregation documented in `core-git`. The "Routing table" requirement references the role interfaces by name.

### New Capabilities

- `core-git` (new) — documents the role-interface segregation contract: method sets per role, consumer-side placement in `internal/core/git.go`, factory wiring contract, test mock structure.

## Impact

| Layer | Files | Lines (est.) |
|---|---|---|
| `internal/core/git.go` | new file: role interface definitions | ~80 |
| `internal/core/worktree.go`, `project.go`, `navigation.go`, `context_resolver.go` | constructor signatures update | ~50 |
| `internal/cmdutil/factory.go` | per-role lazy fields | ~40 |
| `internal/git/gogit_client.go`, `cli_client.go` | add compile-time interface checks | ~10 |
| `test/mocks/` | split `MockGitClientBundle` into 6 mock files | ~200 |
| `test/integration/`, `test/concurrent/`, `test/e2e/fixtures/` | mechanical mock rename | ~600 |
| Specs | 1 new (`core-git`), 1 modified (`git-client`) | — |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged |
| Public API of `internal/core/` | New role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`); constructor signatures change |
| Public API of `internal/cmdutil/` | Factory exposes per-role lazy fields instead of two composite fields |
| Tests | Mechanical: split `MockGitClientBundle` into per-role mocks |
| Mock tooling | testify retained; `moq` deferred |

## Non-goals

- Splitting `WorktreeWriter` (4 methods) into 1-3 method sub-roles — deferred. The 1-3 rule violation is documented; further splitting complicates constructor signatures without proportional benefit at this stage.
- `moq` generation tooling switch — deferred.
- Per-cmd consumer-side interface placement (for example `cmd/<command>.go`) — kept in `internal/core/git.go` per the locked decision.
- Migration of legacy spec prefixes (`domain-*`, `infrastructure-*`, `application-*`) to Tier 2 prefixes (`core-*`, `git-*`, `cli-*`) — deferred.
- `infrastructure-shell-detect` does not consume `GoGitClient` or `CLIClient` (only shell-type probing logic); no delta for that capability.
