# Design: Interface Segregation

See `proposal.md` for motivation (the `GoGitClient` + `CLIClient` 1-3 method rule violation, the test-mock bloat, the constructor over-broad dependencies). This document covers the technical decisions, the constraints that shape them, and the execution order that gets the segregation landed without forcing a wholesale rewrite.

## Context

After `cli-functional-core-shell` lands, the codebase sits in a Tier 2 layout (`internal/core/`, `internal/git/`, `internal/output/`, `internal/iostreams/`, `internal/cmdutil/`, `cmd/`). The `core` package owns the role interfaces that services depend on: `core.GoGitClient` (8 methods) and `core.CLIClient` (6 methods). Both interfaces are still composite facets, inherited from the legacy `application.GitClient` umbrella that `architecture-layer-inversion` deleted without splitting the underlying methods.

Concrete downsides at this state:

- `core.NewWorktreeService(goGit core.GoGitClient, cli core.CLIClient, …)` takes a write-capable `CLIClient` even for read-only code paths inside the service, and an 8-method `GoGitClient` even for a method that only needs `BranchExists`.
- Tests in `test/mocks/git_service_mock.go` carry `MockGitClientBundle { MockGoGitClient, MockCLIClient }` because every service test must stub every method on both bundles, even when exercising one operation.
- The `golang-structs-interfaces` skill recommends 1-3 methods per interface and "define interfaces where consumed". Both recommendations were deferred by `architecture-layer-inversion` (Decision 15) and `cli-functional-core-shell` (Decision 1, alternative) for stability.
- `RepoLocator` was already split to a single-method interface in `architecture-layer-inversion` (Decision 5); it is the reference point for this change.

This change applies the skill's recommendations without touching the package layout (already Tier 2) or the Factory composition root (already lazy). It moves the role interface declarations into `internal/core/git.go` (consumer-side), splits each composite into 1-3 method roles, and updates service constructors + Factory wiring + test mocks atomically.

## Goals / Non-Goals

**Goals:**

- Replace the 8-method `core.GoGitClient` with four role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`), each at 1-3 methods.
- Replace the 6-method `core.CLIClient` with two role interfaces (`WorktreeWriter`, `BranchWriter`), one of which sits at 4 methods with the violation documented.
- Place every role interface declaration in `internal/core/git.go` (consumer-side), per the `golang-structs-interfaces` skill.
- Update each service constructor to take the narrow set of roles it consumes.
- Update the `Factory` to expose one lazy `func() Role` field per role, so commands receive the smallest possible dependency set.
- Add compile-time interface satisfaction checks (`var _ core.BranchReader = (*GoGitClient)(nil)`) on each concrete client to prevent method-set drift.
- Delete `MockGitClientBundle` and split the mock surface into one mock per role; update every test in `test/integration/`, `test/concurrent/`, and `test/e2e/fixtures/` to construct only the mocks it consumes.
- Preserve end-user CLI behavior (commands, flags, output, exit codes) exactly.

**Non-Goals:**

- Wholesale Tier 2 collapse (already done by `cli-functional-core-shell`).
- Splitting `WorktreeWriter` further — the 4-method violation is documented and deferred.
- `moq` generation tooling switch — testify hand-written mocks retained.
- Consumer-side interface placement in `cmd/<command>.go` — locked decision keeps interfaces in `internal/core/git.go`.
- Per-method role interfaces (`WorktreeCreator`, `WorktreeDeleter`, `BranchDeleter`, etc.) — rejected for ceremony; the 1-3 method split already reduces the surface enough.
- Migration of legacy spec prefixes (`domain-*`, `infrastructure-*`, `application-*`) to Tier 2 prefixes (`core-*`, `git-*`, `cli-*`) — deferred.
- Shell-detect role updates — `infrastructure-shell-detect` does not depend on `GoGitClient` or `CLIClient`; no delta required.
- Touching `internal/git/` implementations beyond adding compile-time checks.

## Decisions

### 1. Role interfaces placed consumer-side in `internal/core/git.go`

**Choice.** All six role interfaces declared in a single file `internal/core/git.go` under the `core` package (where services consume them).

**Rationale.** The `golang-structs-interfaces` skill explicitly recommends "define interfaces where consumed; accept interfaces, return structs". `core` is the consumer for `WorktreeService`, `ProjectService`, `NavigationService`, `ContextResolver` business logic. The locked decision (consumer-side in `internal/core/git.go`) rules out per-cmd placement.

**Alternatives considered:**

- Place roles in `internal/git/` (where the concrete clients live) — rejected; the skill's "consumer-side" rule plus the Tier 2 depguard (`core` = stdlib + samber/lo only) make `internal/git/` the wrong home.
- Place each role in its consumer service file (`worktree.go`, `project.go`, …) — rejected by the locked decision; keeps the public surface co-located in one file for discoverability.
- Keep `GoGitClient`/`CLIClient` composites and add the role interfaces alongside — rejected; two parallel surfaces forever.

### 2. Split `GoGitClient` into four roles (2+2+3+1 methods)

**Choice.** Four roles:

| Role | Methods | Count |
|---|---|---|
| `RepositoryOpener` | `OpenRepository`, `ValidateRepository` | 2 |
| `BranchReader` | `ListBranches`, `BranchExists` | 2 |
| `RepositoryReader` | `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo` | 3 |
| `RemoteReader` | `ListRemotes` | 1 |

**Rationale.** The split groups by semantic intent (open/validate, list branches, inspect repo state, list remotes) rather than by accidental method name. `RepositoryOpener` is the only role that needs no `ctx` argument (path-only helpers); the other three are `ctx`-aware. The 1-3 method rule is satisfied (4 roles, none exceed 3). `RemoteReader` sits at 1 method because `ListRemotes` has no natural co-mate at the same granularity.

**Alternatives considered:**

- Three roles (merge `BranchReader` into `RepositoryReader`) — rejected; branch reads have a distinct semantic from repo status and are often used without the other repo reads.
- Five roles (split `GetRepositoryInfo` out of `RepositoryReader`) — rejected; `GetRepositoryInfo` composes the other read operations and consumers that need it usually need the others too.
- Single-method roles (`BranchExister`, `BranchLister`, …) — rejected; constructor bloat without proportional testability win.

### 3. Split `CLIClient` into two roles (4+2 methods; one violation documented)

**Choice.** Two roles:

| Role | Methods | Count |
|---|---|---|
| `WorktreeWriter` | `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees` | 4 |
| `BranchWriter` | `DeleteBranch`, `IsBranchMerged` | 2 |

**Rationale.** `BranchWriter` sits cleanly at 2 methods. `WorktreeWriter` sits at 4 methods — a 1-3 rule violation. The methods form a tightly coupled "worktree lifecycle" set: every code path that creates a worktree also needs to list / delete / prune them, so further splitting adds constructor parameters without reducing real coupling. The violation is documented in the role's spec requirement and as a deferred follow-up.

**Alternatives considered:**

- Three roles (`WorktreeCreator`, `WorktreeDeleter`, `WorktreeInspector`) — rejected; constructor takes 3 roles that are always used together.
- Five roles (split `ListWorktrees` out of `WorktreeWriter`) — rejected; `ListWorktrees` is on every worktree-mutating path.

### 4. Service constructors take role interfaces as separate fields

**Choice.** Each service constructor signature changes from `(goGit GoGitClient, cli CLIClient, …)` to a flat field list of roles consumed. Example:

```go
func NewWorktreeService(
    branchReader BranchReader,
    repoReader RepositoryReader,
    worktreeWriter WorktreeWriter,
    branchWriter BranchWriter,
    projectService ProjectService,
    config *Config,
    hookRunner HookRunner,
) WorktreeService
```

**Rationale.** Mirrors `golang-structs-interfaces` "accept interfaces, return structs"; each role is the dependency the service actually consumes. The constructor parameter list grows (7 fields for `WorktreeService` vs. the previous 5), but each field has a clear name and a single test stub. A bundle struct (`GitOperations interface { BranchReader; RepositoryReader; … }`) reintroduces the very umbrella the skill's role split removes.

**Alternatives considered:**

- Bundle struct `GitOperations interface { BranchReader; RepositoryReader; WorktreeWriter; BranchWriter }` — rejected; reintroduces a composite.
- Pass a single struct holding role-typed fields (`type WorktreeDeps struct { BranchReader …; RepositoryReader … }`) — rejected for now; revisit if constructor parameter count exceeds 8 in a future change.

### 5. Factory exposes per-role lazy function fields

**Choice.** `cmdutil.Factory` gains one lazy `func() (Role, error)` field per role. A `sync.Once` cache inside the factory prevents re-initialization within one command execution.

**Rationale.** Same lazy pattern as `cli-factory`. Each command invokes only the roles it needs (via `f.BranchReader()` etc.), and the same `*GoGitClient` / `*CLIClient` value is returned across calls because the factory caches the concrete instance once and types it as the requested role on each access. The factory's `Init()` step (added in `cli-functional-core-shell`) remains the single composition root.

**Alternatives considered:**

- One `f.Git()` returning `*GoGitClient` and let the consumer type-assert — rejected; defeats the role segregation at the boundary.
- Per-command helper (`f.NewWorktreeServiceDeps()` returning a `WorktreeDeps` struct) — rejected; the Factory's job is dependency injection, not service construction.
- Keep `f.GitClient func() (core.GoGitClient, error)` and `f.CLIClient func() (core.CLIClient, error)` as composites — rejected; contradicts the segregation goal.

### 6. Concrete client satisfaction via existing method sets (no implementation changes)

**Choice.** The concrete `internal/git.GoGitClient` and `internal/git.CLIClient` satisfy the role interfaces automatically because the role method sets are subsets of the concrete method sets. Add a compile-time check on each concrete type:

```go
var (
    _ core.RepositoryOpener = (*GoGitClient)(nil)
    _ core.BranchReader     = (*GoGitClient)(nil)
    _ core.RepositoryReader = (*GoGitClient)(nil)
    _ core.RemoteReader     = (*GoGitClient)(nil)
    _ core.WorktreeWriter   = (*CLIClient)(nil)
    _ core.BranchWriter     = (*CLIClient)(nil)
)
```

**Rationale.** The skill rule "`var _ Iface = (*Concrete)(nil)` near the concrete type" gives a compile-time guarantee that the role method set has not drifted away from the concrete method set. No implementation work needed; only the interface declarations move.

**Alternatives considered:**

- Add per-role helper methods that delegate to existing ones — rejected; redundant.
- Type-assert at the boundary to validate the concrete satisfies all roles — rejected; compile-time check is the build-time version of the same guarantee.

### 7. `RepoLocator` stays single-method

**Choice.** No changes to `core.RepoLocator`. It is already a single-method interface (added in `architecture-layer-inversion`).

**Rationale.** Already satisfies the 1-3 rule. Re-touching it would risk a regression for zero benefit.

**Alternatives considered:** None — the decision is settled.

### 8. Mock split strategy: testify hand-written per-role mocks

**Choice.** Delete `MockGitClientBundle` from `test/mocks/git_service_mock.go`. Create six mock files, one per role (`MockRepositoryOpener`, `MockBranchReader`, `MockRepositoryReader`, `MockRemoteReader`, `MockWorktreeWriter`, `MockBranchWriter`). Each test constructs only the mocks matching its command's consumed roles.

**Rationale.** Mirrors the role split 1:1. Test ergonomics: a test for `BranchExists` no longer stubs `CreateWorktree`, `ListRemotes`, etc. — the mock surface shrinks to exactly the role under test. testify/mock retained per Q3 (`moq` deferred).

**Alternatives considered:**

- Keep `MockGitClientBundle` as a backwards-compatible shim that embeds all per-role mocks — rejected; two parallel surfaces for tests.
- Generate mocks via `moq` — deferred per locked decision.
- Build a `MockAllGit struct { MockBranchReader; MockRepositoryReader; … }` — rejected; the per-role split is the goal; aggregating at the test layer contradicts it.

### 9. Test fixture mechanical update path

**Choice.** Per test, identify the constructor's role set, construct only those mocks, and pass each role mock as the matching constructor parameter. Mechanical rename via `gopls rename` for the type names; parameter order updates by hand at each constructor call site.

**Rationale.** The change is mechanical at the test layer; `gopls rename` propagates the mock type names, and the parameter order at each call site is a small per-test edit. Estimated ~600 lines of test churn across `test/integration/`, `test/concurrent/`, and `test/e2e/fixtures/`.

**Alternatives considered:**

- One-shot sed across the test tree — rejected; cannot reorder constructor parameters safely.
- Per-slice test rewrites — chosen (see Migration Plan).

### 10. Compile-time interface satisfaction check placement

**Choice.** Place each `var _ core.Role = (*Concrete)(nil)` declaration in the concrete client's source file (next to the type declaration), not in the consumer's file.

**Rationale.** The check belongs with the type that has the obligation to satisfy the interface. A consumer-side check would force every service to import every role, defeating the consumer-side placement of the role interfaces.

**Alternatives considered:**

- Place all checks in `internal/core/git.go` — rejected; the consumer must not import the concrete type (Tier 2 depguard).
- Place checks in a separate `internal/git/interfaces_check.go` — rejected; adds a file for what fits naturally next to each type.

## Risks / Trade-offs

- **Constructor parameter count inflation** — `NewWorktreeService` grows from 5 to 7 parameters. → Mitigation: each parameter is named and has a single role; if the count exceeds 8 in a future change, introduce a per-service `Deps` struct.
- **Test ergonomics: more mock files** — six mock files vs. one bundle. → Mitigation: each test that exercises one role mocks only one type; the per-test line count drops.
- **Factory field count growth** — Factory grows from ~6 lazy fields to ~9. → Mitigation: the Factory is the single composition root; commands that don't need a role never call its lazy field.
- **Backward compatibility for external consumers** — none. → Mitigation: the `core` package is internal to the repo (`internal/`); `gopls rename` propagates the rename across the test tree in one pass.
- **`WorktreeWriter` 4-method violation** — sits above the 1-3 rule. → Mitigation: documented in the spec requirement as a deferred follow-up; further split would split methods that are always used together.
- **Mock coverage drift** — a test that previously exercised `CreateWorktree` via `MockGitClientBundle.MockCLIClient` now needs `MockWorktreeWriter`. → Mitigation: `gopls rename` catches type references; grep for `MockCLIClient` / `MockGoGitClient` flags any unconverted call sites.
- **Lockstep with `cli-functional-core-shell`** — `interface-segregation` depends on `cli-functional-core-shell`'s package layout (`internal/core/git.go` exists). → Mitigation: the change's Migration Plan assumes cli-functional-core-shell applies first; `openspec status` for `interface-segregation` lists it as a precondition.

## Migration Plan

Phased slice order (~14 slices). Each slice ends with the build green and Tier 2 depguard satisfied.

1. **Setup** — confirm `cli-functional-core-shell` is applied; verify `internal/core/git.go` exists as the planned target file; verify `internal/git/gogit_client.go` and `internal/git/cli_client.go` carry the methods listed above.
2. **Role interface declarations** — create `internal/core/git.go` with the six role interfaces (`RepositoryOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`); each role has a godoc comment naming its skill-rule compliance (1-3 method compliance documented for five roles; `WorktreeWriter` documents the 4-method violation).
3. **Compile-time checks** — add `var _ core.Role = (*Concrete)(nil)` blocks at the bottom of `internal/git/gogit_client.go` and `internal/git/cli_client.go`; verify by `go build ./internal/git/...` clean.
4. **Service constructor refactor (per service)** — for each of `NewWorktreeService`, `NewProjectService`, `NewNavigationService`, `NewContextResolver`: update the signature to take role interfaces, update the call sites in `cmd/` and in tests. Each slice ends with `go build ./...` clean.
5. **Factory wiring** — update `internal/cmdutil/factory.go` to expose per-role lazy fields (`RepoOpener`, `BranchReader`, `RepoReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) returning the cached concrete client typed as the role; remove the composite `GitClient` / `CLIClient` lazy fields. Update `main.go` to assign the per-role lazy fields.
6. **Mock rewrites (per role)** — for each role, create `test/mocks/<role>_mock.go` with the testify/mock struct; delete `MockGitClientBundle`.
7. **Integration test updates** — update each test in `test/integration/` that constructs `MockGitClientBundle` to construct only the mocks matching its command's consumed roles.
8. **Concurrent test updates** — same mechanical update for `test/concurrent/`.
9. **E2E fixture updates** — same mechanical update for `test/e2e/fixtures/`.
10. **Add `core-git` spec** — write `openspec/changes/interface-segregation/specs/core-git/spec.md` with the role interface method sets, consumer-side placement, factory wiring contract, and mock structure.
11. **Modify `git-client` spec** — write `openspec/changes/interface-segregation/specs/git-client/spec.md` with the updated "Routing table" requirement that references the new role interfaces by name.
12. **Final lint pass** — run `golangci-lint run`; ensure no new findings.
13. **Build + test verification** — `go build ./... && go test ./...` exits 0.
14. **OpenSpec verification** — `openspec validate interface-segregation --json` returns `valid: true`; `openspec status --change interface-segregation --json` returns `isPlanningComplete: true`.

**Rollback:** every slice is a discrete commit. `git revert` from the merge commit restores the prior state. The compile-time interface checks are the only point where a half-applied state breaks the build; staging them in slice 3 keeps every earlier slice self-consistent.

## Open Questions

None. The 10 decisions above are settled; the only deferred items (per-role method-count tightening for `WorktreeWriter`, `moq` migration, per-cmd consumer-side placement, legacy spec prefix migration) are explicitly out of scope and do not change the segregation contract.
