# Design: Interface Segregation

See `proposal.md` for the motivation and the supersession note that scopes
this change (the service-constructor refactor and mock split originally
envisioned were already delivered or made redundant by the archived
`cli-functional-core-shell` change). This document covers the technical
decisions that get the role-interface declaration, compile-time checks,
and Factory per-role fields landed without expanding into a wholesale
rewrite.

## Context

After `cli-functional-core-shell` lands, the codebase sits in a Tier 2
layout (`internal/core/`, `internal/git/`, `internal/output/`,
`internal/iostreams/`, `internal/cmdutil/`, `cmd/`). The composite
`git.Client` (defined in `internal/git/client.go`) is the canonical
injection point — it embeds `*reader` (read-side methods) and `*cliClient`
(write-side methods), so a single `*git.Client` value carries the entire
git I/O surface for downstream consumers.

Role interfaces per the `golang-structs-interfaces` skill have not yet been
declared in `internal/core/git.go`: today every method on the read- and
write-side concrete is reachable only through the composite. There is no
named contract that says "this is what `BranchReader` covers", no
compile-time guard that catches a method rename on the concrete, and no
narrow Factory field a command can request when it only needs a subset of
methods.

Concrete downsides at this state:

- A future read-only command has no narrow role to depend on; it must take
  the composite and inherit every write method by accident-of-typing.
- Renaming `OpenRepository` to `OpenRepo` on `*reader` breaks every call
  site through a runtime panic at first call, not at compile time.
- Documentation of which methods are "the branch read surface" or "the
  worktree write surface" lives only in godoc comments on the concrete,
  not in a discoverable interface.

`RepositoryLocator` (here `RepoLocator`) was already split to a single-method
interface in earlier work; it is the reference point for this change.

This change declares the role interfaces consumer-side in
`internal/core/git.go`, adds compile-time checks on the composite
`*git.Client` in `internal/git/client.go`, adds a single sentinel drift
metatest in `internal/git/client_test.go`, and adds 6 per-role lazy fields
to `cmdutil.Factory` alongside the existing `GitClient` field. It does not
introduce or reintroduce a service layer and does not reintroduce mocks.

## Goals / Non-Goals

**Goals:**

- Replace the implicit "composite `*git.Client` is everything" contract
  with six named role interfaces (`RepositoryOpener`, `BranchReader`,
  `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`),
  each at 1-3 methods except `WorktreeWriter` (4 methods, documented
  violation).
- Place every role interface declaration in `internal/core/git.go`
  (consumer-side), per the `golang-structs-interfaces` skill.
- Add `var _ core.Role = (*git.Client)(nil)` declarations in
  `internal/git/client.go` so method-set drift fails the build.
- Add a `_DriftCheck` sentinel in `internal/git/client_test.go` that embeds
  all six role fields; renaming or removing any role method fails the build.
- Add six per-role lazy function fields on `cmdutil.Factory` so commands
  that want a narrow role can opt into it. Keep the composite `GitClient`
  field unchanged.
- Preserve end-user CLI behavior (commands, flags, output, exit codes)
  exactly.

**Non-Goals:**

- Wholesale Tier 2 collapse (already done by `cli-functional-core-shell`).
- Service-layer reintroduction and constructor refactor (the service layer
  was collapsed deliberately; reintroducing it would undo that decision).
- Mock-tooling switch or test-fixture rewrite (no mocks exist; tests use
  real clients + `runF`).
- Per-method role interfaces (`WorktreeCreator`, `WorktreeDeleter`,
  `BranchDeleter`, etc.) — rejected for ceremony.
- Touching `internal/git/` implementations beyond adding compile-time
  checks and the drift metatest.

The cross-cutting deferrals (further splitting `WorktreeWriter`, the `moq`
switch — not applicable, per-cmd consumer-side placement, legacy spec
prefix migration) are tracked in `proposal.md` Non-goals.

## Decisions

### 1. Role interfaces placed consumer-side in `internal/core/git.go`

**Choice.** All six role interfaces declared in a single file
`internal/core/git.go` under the `core` package (where services consume
them).

**Rationale.** The `golang-structs-interfaces` skill explicitly recommends
"define interfaces where consumed; accept interfaces, return structs".
`core` is the natural consumer-side for any future service that consumes
the git role surface; the locked decision (consumer-side in
`internal/core/git.go`) rules out per-cmd placement.

**Alternatives considered:**

- Place roles in `internal/git/` (where the concrete lives) — rejected; the
  skill's "consumer-side" rule plus the Tier 2 depguard (`core` = stdlib +
  `samber/lo` only) make `internal/git/` the wrong home for consumer
  abstractions.
- Place each role in its consumer service file (`worktree.go`,
  `project.go`, …) — rejected; keeps the public surface co-located in one
  file for discoverability.
- Skip role interfaces entirely and rely on the composite — rejected; loses
  the compile-time drift discipline and the future-flexibility win.

### 2. Split the surface into four read-side roles

**Choice.** Four read-side roles:

| Role | Methods | Count |
|---|---|---|
| `RepositoryOpener` | `ValidateRepository` | 1 |
| `BranchReader` | `ListBranches`, `BranchExists` | 2 |
| `RepositoryReader` | `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo` | 3 |
| `RemoteReader` | `ListRemotes` | 1 |

**Rationale.** The split groups by semantic intent (validate, list branches,
inspect repo state, list remotes) rather than by accidental method name.
`RepositoryOpener` is at 1 method because `OpenRepository` is excluded (see
Decision 3). `RemoteReader` sits at 1 method because `ListRemotes` has no
natural co-mate at the same granularity. The 1-3 method rule is satisfied
for three of the four roles; `RepositoryReader` sits at 3 (the upper
bound), which is the natural ceiling for the inspection surface.

### 3. `OpenRepository` is not in `RepositoryOpener`

**Choice.** `core.RepositoryOpener` declares only `ValidateRepository(path)
error`. `OpenRepository(path) (*go-git.Repository, error)` is **not** in any
role interface; it stays as a concrete method on `*reader`, called only by
other `*reader` methods and by integration tests.

**Rationale.** Two independent reasons force this exclusion:

1. **No command consumes `OpenRepository`.** Every CLI command reaches the
   git I/O surface through higher-level methods (`ValidateRepository`,
   `ListBranches`, `GetRepositoryStatus`, `ListWorktrees`, etc.). The skill
   rule "don't design with interfaces, discover them" says: write the
   interface when a second consumer demands it, not before.
2. **`*go-git.Repository` is not exportable into `internal/core/`.** The
   `core-isolation` depguard (per `.golangci.yml`) restricts `internal/core/`
   to `$gostd + samber/lo`. A role interface method returning
   `*go-git.Repository` would either force `core` to import `go-git` (a
   depguard violation) or require a wrapper type (scope creep).

**Alternatives considered:**

- Wrap `*go-git.Repository` in a core type (e.g. `core.Repository`) —
  rejected for scope creep; no consumer needs the wrapper today, and the
  type's surface (`go-git.Repository.References()`, `repo.Storer`, …)
  extends well beyond what a role interface should expose.
- Move role interfaces out of `internal/core` into `internal/git/` so the
  import is allowed — rejected; violates the consumer-side rule and the
  Tier 2 depguard layers.

### 4. Split the write-side surface into two roles

**Choice.** Two write-side roles:

| Role | Methods | Count |
|---|---|---|
| `WorktreeWriter` | `CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees` | 4 |
| `BranchWriter` | `DeleteBranch`, `IsBranchMerged` | 2 |

**Rationale.** `BranchWriter` sits cleanly at 2 methods. `WorktreeWriter`
sits at 4 methods — a 1-3 rule violation. The methods form a tightly
coupled "worktree lifecycle" set: every code path that creates a worktree
also needs to list / delete / prune them, so further splitting adds
constructor parameters without reducing real coupling. The violation is
documented in the role's spec requirement and as a deferred follow-up.

### 5. The composite `git.Client` remains the canonical injection point

**Choice.** `cmdutil.Factory.GitClient func() (*git.Client, error)` is
retained unchanged. Per-role lazy fields are added alongside it; commands
that already use `f.GitClient()` keep working without modification.

**Rationale.** The composite was deliberately made the canonical injection
point in `cli-functional-core-shell` (Decision "no service-layer
indirection" + the requirement that `cmd/` calls git via the Factory). The
role interface segregation in this change is documentation, drift
discipline, and **future** opt-in narrowing — not a replacement of the
composite at every existing call site.

**Alternatives considered:**

- Retiring `f.GitClient()` in favor of per-role fields — rejected; no
  current command benefits (every git-consuming command needs both read
  and write methods, as the existing `cmd/list.go`, `cmd/create.go`,
  `cmd/delete.go`, `cmd/prune.go`, and `cmd/cd.go` show), and retirement
  would touch every command file for no call-site win.

### 6. Compile-time checks on the composite

**Choice.** Add a single block at the bottom of `internal/git/client.go`:

```go
var (
    _ core.RepositoryOpener  = (*git.Client)(nil)
    _ core.BranchReader      = (*git.Client)(nil)
    _ core.RepositoryReader  = (*git.Client)(nil)
    _ core.RemoteReader      = (*git.Client)(nil)
    _ core.WorktreeWriter    = (*git.Client)(nil)
    _ core.BranchWriter      = (*git.Client)(nil)
)
```

(Adjusted for the actual package alias; `internal/git/client.go` uses
`package git`.)

**Rationale.** The composite `*git.Client` embeds `*reader` and
`*cliClient`, so its method set is the union of both halves. A single
`var _` block on the composite covers all 14 role methods at once; per-role
checks against `(*reader)(nil)` and `(*cliClient)(nil)` would be redundant
(the composite promotes the same methods).

**Alternatives considered:**

- Per-role checks on each unexported concrete (`*reader`, `*cliClient`) —
  rejected; redundant with the composite check and would force reader.go
  and writer.go to know about role interfaces they don't otherwise need.
- Type-assert at the boundary to validate the composite satisfies all roles
  — rejected; compile-time check is the build-time version of the same
  guarantee.

### 7. Drift discipline is a single sentinel in `client_test.go`

**Choice.** Append a single sentinel struct to
`internal/git/client_test.go`:

```go
type _DriftCheck struct {
    _ core.RepositoryOpener
    _ core.BranchReader
    _ core.RepositoryReader
    _ core.RemoteReader
    _ core.WorktreeWriter
    _ core.BranchWriter
}
```

**Rationale.** A single struct embedding all six role fields against the
composite `*git.Client` is sufficient: `go build ./...` will fail with a
type-mismatch error the moment any method name or signature on the
composite drifts away from any role. Splitting the sentinel across two
files (read-side vs write-side) was the original design; it adds two files
for the same guarantee.

### 8. Factory per-role fields share the composite's cache

**Choice.** Each per-role lazy field reaches the same `*git.Client`
instance as `f.GitClient()`. Because `f.GitClient` is already implemented as
`sync.OnceValues` (in `NewFactory`), every per-role field that internally
calls `f.GitClient()` reuses the same cached concrete.

```go
f.GitClient = sync.OnceValues(func() (*git.Client, error) {
    if _, err := f.Config(); err != nil {
        return nil, err
    }
    return git.NewClient()
})

f.BranchReader = func() (core.BranchReader, error) {
    c, err := f.GitClient()
    if err != nil {
        return nil, err
    }
    return c, nil // *git.Client satisfies core.BranchReader via embedding
}
```

**Rationale.** Three `sync.Once`s (one per concrete, one per role) would
re-initialize the underlying LRU cache and break the contract that the
concrete is built once per command execution. Routing every per-role field
through `f.GitClient()` ensures one `sync.Once` per concrete.

**Alternatives considered:**

- One `sync.Once` per per-role field — rejected; would re-initialize the
  LRU cache (and possibly rebuild the client) the first time each role is
  touched.
- No `sync.Once` on per-role fields; rely on `f.GitClient()`'s `Once` —
  chosen (above).

### 9. No mock files; no test-fixture rename

**Choice.** `test/mocks/` is not created. Integration tests continue to use
real `*git.Client` instances via `git.NewClient()` (integration tier:
`//go:build integration`). Command unit tests continue to use the
`cmdutil.Factory` `runF` injection point (or a real `*git.Client` with a
temp dir) without mocks.

**Rationale.** `test/mocks/git_service_mock.go` does not exist; the role
segregation does not require it. The skill guidance ("use `moq` for 5+
method interfaces, hand-written doubles for smaller") does not apply
because no command relies on a mock for the git surface today — every
test that exercises git I/O does so through `git.NewClient()` against
real git repos (integration) or via the `runF` seam (command).

## Risks / Trade-offs

- **Factory field count growth** — Factory gains 6 lazy fields (from 1
  composite to 7 lazy). → Mitigation: the composite field stays;
  per-role fields are opt-in. Commands that want narrow typing can pick
  exactly the roles they need.
- **`WorktreeWriter` 4-method violation** — sits above the 1-3 rule.
  → Mitigation: documented in the role's spec requirement as a deferred
  follow-up; further split would split methods that are always used
  together.
- **Backward compatibility for external consumers** — none. → Mitigation:
  the `core` package is internal to the repo (`internal/`); the only
  out-of-tree consumer is `cmd/`, which keeps using the composite.
- **No call-site narrowing today** — every git-consuming command needs
  both read and write methods, so no command currently opts into a narrow
  role field. → Mitigation: this is acknowledged in `proposal.md` Why;
  per-role fields are for future read-only commands and as documentation
  of the role surface.
- **Lockstep with `cli-functional-core-shell`** — already applied and
  archived. → Mitigation: `proposal.md` documents the supersession;
  `tasks.md` does not assume any of the deleted packages.

## Migration Plan

8 numbered slices. Each slice ends with the build green and Tier 2
depguard satisfied.

1. **Setup** — confirm `cli-functional-core-shell` is archived; verify
   `internal/git/reader.go` carries the read-side method set and
   `internal/git/writer.go` carries the write-side method set; verify
   `internal/core/git.go` does not yet exist (it is created in slice 2).
2. **Role interface declarations** — create `internal/core/git.go` with
   the six role interfaces; `RepositoryOpener` is 1-method
   (`ValidateRepository` only — no `OpenRepository`).
3. **Compile-time satisfaction** — append the `var _ core.Role =
   (*git.Client)(nil)` block at the bottom of `internal/git/client.go`;
   `go build ./internal/git/...` clean.
4. **Drift metatest** — append the `_DriftCheck` sentinel to
   `internal/git/client_test.go`; `go build ./internal/git/...` and
   `go test ./internal/git/...` clean.
5. **Factory wiring (additive)** — append six per-role lazy fields to
   `cmdutil.Factory` in `internal/cmdutil/factory.go`; each routes
   through `f.GitClient()` for shared caching. Update `Init()` to touch
   the six new fields so initialization failures surface.
6. **Build clean** — `go build ./...` exits 0.
7. **Lint/vet/fmt** — `golangci-lint run`, `gofmt -l`, `goimports -l`,
   `go vet ./...` all clean.
8. **OpenSpec verification** — `openspec validate interface-segregation
   --strict --json` returns `valid: true`; `openspec status --change
   interface-segregation --json` returns `isPlanningComplete: true`.

**Rollback:** every slice is a discrete edit. `git revert` from the merge
commit restores the prior state. The compile-time interface checks are
the one point where a half-applied state breaks the build; staging them
in slice 3 keeps every earlier slice self-consistent.

## Open Questions

None. The decisions above are settled; the only deferred items (further
splitting `WorktreeWriter`, per-cmd consumer-side placement, legacy spec
prefix migration) are explicitly out of scope and do not change the
segregation contract.
