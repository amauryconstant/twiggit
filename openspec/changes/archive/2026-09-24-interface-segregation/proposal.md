# Proposal: Interface Segregation

## Supersession note (added at revise time)

This change was originally drafted against an architectural state where
`internal/core/git.go` did not yet exist, `internal/git/{gogit_client,cli_client}.go`
were the concrete-type sources, a service layer (`internal/application/`,
`internal/service/`) sat between commands and the git adapter, and
`test/mocks/git_service_mock.go` carried a `MockGitClientBundle` consumed
across `test/integration/`, `test/concurrent/`, and `test/e2e/fixtures/`.

Subsequent work — primarily the archived `cli-functional-core-shell` change —
already:

- Collapsed the service layer entirely (`internal/application/`,
  `internal/service/`, `internal/infrastructure/` deleted).
- Renamed concrete files (`gogit_client.go` → `client.go` + `reader.go`,
  `cli_client.go` → `writer.go`; concrete types are unexported `*reader` and
  `*cliClient` with `CLIClient = cliClient` exported as a type alias).
- Deleted `test/mocks/`. Tests now use a real `*git.Client` constructed via
  `git.NewClient()`, plus the `cmdutil.Factory` `runF` test seam.

The original service-constructor refactor (§4 of `tasks.md`) and mock-split
(§§6-9 of `tasks.md`) are therefore no-ops against the current codebase.

The remaining motivation is real but smaller: declare the role interfaces in
`internal/core/git.go` (consumer-side, per `golang-structs-interfaces`), add
compile-time satisfaction checks on the concrete `*git.Client`, and expose
per-role lazy fields on `cmdutil.Factory` so commands that read (or only
write) can request a narrow role. The composite `*git.Client` remains the
canonical injection point; per-role fields are additive.

The original tasks 4.1-4.6 (service-constructor refactor), 6.1-6.7 (mock
rewrites), and 7.1-9.2 (test updates against mocks) are dropped; the rest is
re-shaped into 25 actionable items in `tasks.md`.

## Why

Every method on the read-side `*reader` and write-side `*cliClient` is
exposed via the composite `*git.Client` (which embeds both halves). The
`golang-structs-interfaces` skill identifies 1-3 method interfaces as the
optimal shape; declaring named role interfaces on the consumer side
documents which method subset each role covers, gives compile-time drift
detection on the concrete, and lets commands opt into narrow role fields on
the Factory without forcing all of them through a narrow role.

The win is documentation and future flexibility (a future read-only command
can declare `f.BranchReader()` and the compiler enforces that it cannot
accidentally invoke `client.DeleteBranch`). It is not a call-site narrowing
today: every current git-consuming command needs both read and write
methods, so the composite remains the right answer for them.

## What Changes

### Role interfaces (consumer-side, in `internal/core/git.go`)

Six role interfaces, each at 1-3 methods except `WorktreeWriter` (4 methods,
documented violation):

| Role | Methods | Count | Source |
|---|---|---|---|
| `RepositoryOpener` | `ValidateRepository` | 1 | read-side (no `OpenRepository`; see Non-goals) |
| `BranchReader` | `ListBranches`, `BranchExists` | 2 | read-side |
| `RepositoryReader` | `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo` | 3 | read-side |
| `RemoteReader` | `ListRemotes` | 1 | read-side |
| `WorktreeWriter` | `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees` | 4 | write-side (1-3 violation, documented) |
| `BranchWriter` | `DeleteBranch`, `IsBranchMerged` | 2 | write-side |

`OpenRepository(path) (*go-git.Repository, error)` is **not** in any role
interface: no command consumes it directly, and `*go-git.Repository` is not
an exportable type in `internal/core/` (the `core-isolation` depguard allows
only stdlib + `samber/lo`). `OpenRepository` remains a concrete method on
`*reader` and is consumed only by other `*reader` methods.

### Factory wiring (additive, not a replacement)

`internal/cmdutil/factory.go` gains six per-role lazy function fields:

```go
type Factory struct {
    // ... existing fields unchanged ...

    GitClient func() (*git.Client, error) // composite, retained

    // Per-role lazy fields. Each returns the cached *git.Client
    // typed as the requested role (Go interface satisfaction).
    RepoOpener     func() (core.RepositoryOpener, error)
    BranchReader   func() (core.BranchReader, error)
    RepositoryReader func() (core.RepositoryReader, error)
    RemoteReader   func() (core.RemoteReader, error)
    WorktreeWriter func() (core.WorktreeWriter, error)
    BranchWriter   func() (core.BranchWriter, error)
}
```

Each per-role field reaches the same `*git.Client` as `f.GitClient()`; the
concrete is built exactly once per command execution because `f.GitClient`
already uses `sync.OnceValues`. Per-role fields are additive: commands that
already use `f.GitClient()` keep working unchanged; commands that want to
type-narrow can opt into one or more role fields.

### Compile-time role satisfaction

`internal/git/client.go` carries `var _ core.Role = (*git.Client)(nil)`
declarations for each of the six roles (the `*git.Client` composite
satisfies them via embedded promotion of `*reader` and `*cliClient`). A
drift metatest in `internal/git/client_test.go` declares a `_DriftCheck`
sentinel embedding all six role fields; renaming or removing any role
method fails the build.

### Service constructor signatures

**No changes.** The service layer was collapsed in `cli-functional-core-shell`.
Commands call `*git.Client` (or per-role fields) directly. There is no
`core.NewWorktreeService` / `core.NewProjectService` /
`core.NewNavigationService` / `core.NewContextResolver` to refactor; if a
future change reintroduces services, the role interfaces here will be its
constructor argument shapes.

### Mock rewrites

**No changes.** `test/mocks/` does not exist. Tests use a real `*git.Client`
constructed via `git.NewClient()` (integration tests against real git
repos) and the `cmdutil.Factory` `runF` test seam (command unit tests). No
`MockGitClientBundle`, `MockGoGitClient`, or `MockCLIClient` exists to split.

## Capabilities

### Modified Capabilities

- `git-client` (modifies the Tier 2 spec at `openspec/specs/git-client/spec.md`) — MODIFY the "Routing table" requirement so it names the six role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) and clarifies that the composite `git.Client` remains the canonical injection point. REPLACE the "Per-role factory wiring supersedes composite constructor" requirement with "Per-role Factory fields are additive to composite" (the composite was kept; per-role fields are opt-in narrowing). Drop references to `application.GoGitClient` (the package was deleted in `cli-functional-core-shell`). The full role method sets, consumer-side placement, and compile-time drift discipline live in `core-git`; `git-client` cross-references rather than re-derives them.

### New Capabilities

- `core-git` (new) — documents the role-interface segregation contract: method sets per role, consumer-side placement in `internal/core/git.go`, composite satisfaction via embedded promotion, Factory per-role fields as additive narrowing, and the single-sentinel drift metatest on `*git.Client`.

## Impact

| Layer | Files | Lines (est.) |
|---|---|---|
| `internal/core/git.go` | new file: role interface declarations | ~80 |
| `internal/git/client.go` | append compile-time `var _` checks for 6 roles | ~10 |
| `internal/git/client_test.go` | append `_DriftCheck` sentinel | ~10 |
| `internal/cmdutil/factory.go` | add 6 per-role lazy fields | ~30 |
| Specs | 1 new (`core-git`), 1 modified (`git-client`) | — |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged |
| Public API of `internal/core/` | New role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) |
| Public API of `internal/cmdutil/` | Factory exposes 6 additional per-role lazy fields alongside the composite `GitClient` |
| Tests | No changes (no mocks to split; integration tests already use real clients) |
| Mock tooling | No change (project uses real clients + `runF` seam) |

## Non-goals

- Splitting `WorktreeWriter` (4 methods) into 1-3 method sub-roles — deferred. The 1-3 rule violation is documented; further splitting complicates the surface without proportional benefit at this stage.
- `moq` generation tooling switch — not applicable; no mocks exist.
- Per-cmd consumer-side interface placement (for example `cmd/<command>.go`) — kept in `internal/core/git.go` per the locked decision.
- Migration of legacy spec prefixes (`domain-*`, `infrastructure-*`, `application-*`) to Tier 2 prefixes (`core-*`, `git-*`, `cli-*`) — deferred.
- Service-layer reintroduction and constructor refactor — collapsed in `cli-functional-core-shell`; not part of this change.
- Mock-bundle split and integration/concurrent/e2e fixture rename — no mocks to split (directory deleted).
- Including `OpenRepository` in `RepositoryOpener` — no command consumes it; `*go-git.Repository` would violate `core-isolation` depguard.
