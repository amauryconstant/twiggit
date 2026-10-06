# Design

## Context

See `proposal.md` for motivation. The technical constraints that shape
this design:

- `*git.Client` is a composite that embeds `*reader` and `*cliClient`;
  role methods are reached via embedded promotion. Adding two new
  roles extends the embedded field set; no other surface changes.
- The current `cli-factory` spec explicitly bans per-role lazy fields
  on `cmdutil.Factory` (enforced by
  `internal/cmdutil/factory_naming_test.go:36`). Role methods MUST be
  reached by type-asserting `Factory.GitClient()` to `*git.Client` and
  using the embedded methods directly.
- `internal/core/` is I/O-free (`core-isolation` depguard rule). Role
  interfaces live there consumer-side; concrete implementations live
  in `internal/git/`.
- The existing worktree-lifecycle skip buckets (current / protected /
  unmerged) in `cmd/prune.go` are the canonical template for the
  rebase walk; sync reuses the same walk when `--rebase` is set.
- `core.HookRunRequest` already carries 7 fields. Adding 6 more is
  additive and optional; existing post-create callers MUST compile
  unchanged.
- The rebase/sync command surface must mirror `twiggit prune` so users
  see one consistent walk / skip / report shape across the lifecycle
  commands.

## Goals / Non-Goals

**Goals:**

- One shared inner rebase machinery drives both `twiggit rebase` and
  `twiggit sync --rebase`.
- Stop-on-first conflict semantics that match `git rebase`'s mental
  model, with `--continue` and `--abort` verbs that map to `git
  rebase --continue` / `git rebase --abort`.
- Tracked base discoverable via stock git tooling
  (`git config --worktree --get twiggit.tracked-base`).
- Hook integration that lets users stash before rebase and run
  follow-up commands after a successful rebase or sync.
- Every change additive or shielded by a new role method so no existing
  consumer or test breaks.

**Non-Goals (design-level):**

- Detecting or resolving conflicts programmatically. We rely on git's
  own conflict machinery; the cmd layer just reports.
- Caching or memoising rebase outcomes across runs. Each invocation
  spawns a fresh `git rebase`.
- A long-running watcher that auto-rebases when the base branch moves.
  Both commands are one-shot.
- Cross-project fan-out. Both commands operate on a single project
  context (current or named). Cross-project sync is a deliberate
  follow-up.

## Decisions

### Decision 1: Persist tracked base in git per-worktree config

- **Choice**: Each worktree's tracked base lives at
  `git config --worktree twiggit.tracked-base <branch>` (file path
  `$GIT_DIR/worktrees/<id>/config.worktree`). `twiggit create` writes
  it after a successful worktree add; `twiggit rebase --set-base`
  writes it on demand; `twiggit rebase` reads it.
- **Rationale**: Lives with the branch (survives worktree moves,
  removed on `git worktree remove`), discoverable via stock git
  tooling, no twiggit data dir needed, and no collision with git's own
  config namespaces (`twiggit.*` is reserved).
- **Alternatives considered**:
  - **`$XDG_DATA_HOME/twiggit/tracked-base/<project>/<branch>`** —
    would survive any git operation but breaks if a worktree is moved
    (path-dependent) and adds a parallel sidecar. Rejected as the
    fallback only when git rejects the config key (it does not).
  - **`.twiggit-base` sidecar in the worktree** — easy to read but
    requires custom file handling and survives branch deletion, which
    is the opposite of what we want.

### Decision 2: Two new roles, not one

- **Choice**: `core.Rebaser` (4 methods: `Rebase`, `Abort`, `Continue`,
  `Fetch`) and `core.BaseTracker` (2 methods: `SetTrackedBase`,
  `GetTrackedBase`) are declared in `internal/core/git.go`. Both are
  satisfied by `*cliClient` and exposed through `*Client` via embedded
  promotion. The drift sentinel in `client_test.go` embeds both.
- **Rationale**: `Rebaser` groups the rebase lifecycle (4 methods is
  the same shape as `WorktreeWriter`); `BaseTracker` is a tight
  read/write pair for per-wt config. Splitting keeps each role near
  the recommended 1-3 method ceiling while letting the rebase cmd
  compose both via the same composite client.
- **Alternatives considered**:
  - **Single `Rebaser` with all 7 methods** — exceeds the 1-3
    recommendation; mixes two unrelated concerns (rebase lifecycle +
    per-wt config I/O). Rejected.
  - **Extend `BranchWriter` (currently 2 methods)** — couples rebase
    with branch mutation; BranchWriter is already on a different
    lifecycle stage. Rejected as a coupling smell.

### Decision 3: Reuse prune's walk + skip-bucket pattern

- **Choice**: `runRebase` and the rebase half of `runSync --rebase`
  share `resolveRebaseTargets`, `resolveTrackedBase`,
  `dispatchRebaseHooks`, and `emitRebaseOutput` from
  `cmd/rebase_helpers.go`. The walk iterates targets, classifies each
  into `RebasedWorktree` with an `Outcome` and optional `SkipReason`,
  and aggregates into a `RebaseResult` mirroring
  `core.PruneWorktreesResult`. `runSync --rebase` calls the same walk
  and appends its result to the `SyncResult.RebasedBranches` slice.
- **Rationale**: Consistency with `prune` means users see one mental
  model for batch walk / conflict / report shape across the lifecycle
  commands. Existing helper functions (`detectContext`, `resolve*`,
  `discoverProject`) carry over without modification.
- **Alternatives considered**:
  - **New walk pattern for rebase** — divergent shape, harder to
    review, harder to keep error messages consistent. Rejected.
  - **Compose rebase + sync as a single CLI verb with subcommands**
    (`twiggit workflow rebase`) — adds a layer without changing
    capability. Rejected as speculative architecture.

### Decision 4: Stop-on-first conflict with `--continue` / `--abort` verbs

- **Choice**: `--all` halts iteration on the first
  `RebaseOutcomeConflicted`. Remaining worktrees are reported with
  `SkipReason: "prior conflict"`. The user runs `twiggit rebase
  --continue <wt>` after fixing the conflict, or `twiggit rebase
  --abort <wt>` to abandon. Exit code is 1 on conflict, 0 on clean
  run.
- **Rationale**: Matches `git rebase`'s mental model (one rebase at a
  time, the user drives resolution). Stopping on the first conflict
  prevents cascading badness and gives the user one actionable item.
- **Alternatives considered**:
  - **Continue-past-conflict** — leaves multiple worktrees in
    mid-rebase state; the user has to fix each. Higher cognitive load,
    rarely what users want. Configured as a deferred
    `ConflictPolicy` knob for follow-up.
  - **Abort-on-any-conflict** — discards work. Worse than
    stop-on-first because the user loses the partial progress on
    earlier wts.

### Decision 5: Offline by default; `--fetch` opt-in

- **Choice**: `twiggit rebase` does not touch the network unless
  `--fetch` is supplied. `twiggit sync` is the verb for refreshing
  remote tracking refs and is a separate concern. `twiggit sync
  --rebase` may then rebase onto the freshly fetched base.
- **Rationale**: Users running rebase in CI or on a plane expect
  offline operation by default. The two-verb split
  (`rebase` offline, `sync` online) makes the intent explicit and
  matches git's `fetch` / `rebase` split.
- **Alternatives considered**:
  - **Fetch-on-every-rebase** — surprises users in CI and on slow
    networks. Rejected.
  - **Fetch configured globally via `[rebase].fetch_on_all`** — keeps
    the per-command default offline but lets teams opt in for
    `--all`. Implemented as `RebaseConfig.FetchOnAll` (default false,
    reserved for follow-up; not surfaced as a flag in v1).

### Decision 6: Hook contract extended additively

- **Choice**: `core.HookType` enum gains `HookTypePreRebase`,
  `HookTypePostRebase`, `HookTypePostSync`. `core.HookRunRequest`
  gains six optional string fields (`RebaseBase`, `RebaseOldTip`,
  `RebaseNewTip`, `RebaseResult`, `SyncRemote`, `SyncBranch`). The
  runner sets the corresponding `TWIGGIT_REBASE_*` or
  `TWIGGIT_SYNC_*` env vars only when the hook type matches and the
  field is non-empty.
- **Rationale**: The existing `HookTypePostCreate` and the
  `HookRunRequest`'s 7 existing fields are unchanged; existing callers
  and config keys compile and load without change. New hooks are additive.
- **Alternatives considered**:
  - **Fork `HookRunRequest` into per-type variants** — breaks the
    single runner type and forces every caller to switch on
    `HookType`. Rejected.
  - **Encode rebase state in `SourceBranch` and ad-hoc fields** —
    loses type clarity and breaks the env-var table. Rejected.

### Decision 7: PreRebase hook failure aborts the rebase; PostRebase / PostSync hook failures are warnings

- **Choice**: `HookTypePreRebase` non-zero exit blocks the rebase
  (signal, not warning). `HookTypePostRebase` and `HookTypePostSync`
  non-zero exit logs a warning but the rebase/sync outcome is
  preserved.
- **Rationale**: Pre-rebase hooks are an explicit "block rebase until
  you have done X" signal (e.g., stash). Their failure means the
  pre-condition is not satisfied; continuing would lose data.
  Post-rebase hooks are "after success, do X"; their failure is a
  secondary concern (e.g., `npm install` failed) that the user can
  re-run. Mirrors the post-create hook philosophy.
- **Alternatives considered**:
  - **All hook failures are warnings** — same as today; loses the
    pre-rebase-as-gate semantic that users want. Rejected.

### Decision 8: Sync reuses the rebase walk

- **Choice**: `twiggit sync --rebase` calls `runRebaseWalk` with the
  same opts-derived targets and aggregates the rebase result into
  `SyncResult.RebasedBranches`. `twiggit sync` without `--rebase`
  never calls the rebase walk.
- **Rationale**: One inner machinery, two surface verbs. Avoids the
  fan-out walking logic duplicating between rebase and sync.
- **Alternatives considered**:
  - **Sync has its own rebase walk** — duplicated logic, drift risk.
    Rejected.
  - **Sync doesn't rebase; users run `twiggit rebase --all` after
    `twiggit sync`** — two commands for the common workflow; forces
    users to remember the verb pair. Rejected.

### Decision 9: Conflict-detection from stderr phrase table

- **Choice**: The rebase adapter detects `RebaseOutcomeConflicted`
  by parsing stderr for the git rebase conflict phrase (`"could not
  apply"` or `"Resolve all conflicts manually"`). Detects
  `RebaseOutcomeNothingToDo` from `"up to date"` or `"Your branch is
  up to date"`. The detector lives in `internal/git/rebaser.go` with
  table-driven unit tests per phrase, mirroring the existing
  `isNotFoundStderr` precedent in `internal/git/writer.go`.
- **Rationale**: Git's exit codes do not distinguish "conflict
  paused" from "fetch failed"; the stderr phrase is the
  authoritative signal. Centralised detector prevents future
  git-version skew from breaking one of two callers.
- **Alternatives considered**:
  - **Detect via `.git/rebase-merge/` directory presence** — fragile
    (path varies by rebase type) and only available mid-rebase.
    Rejected.
  - **Wrap `git rebase` in a custom script that emits JSON** — adds a
    runtime dependency and is git-version-specific. Rejected.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| `git config --worktree` requires git ≥ 2.20 (Mar 2018); older git errors with an unclear message | The adapter returns a `core.NewRebaseError("set", "git config --worktree requires git 2.20+", ...)`; the cmd layer surfaces it with an actionable message. The minimum git version is documented in `git-config` and `README.md`. |
| Stop-on-first conflict feels overly strict when many wts have diverged in the same way | `core.RebaseConfig.ConflictPolicy` is reserved as a config knob; the v1 default is `stop`. A future policy value (`continue`) is a follow-up change. |
| Per-wt config survives `git worktree remove` only if the user runs `wtg delete` (which calls `git worktree remove` then `git branch -D`); a stray manual `rm -rf` of the worktree leaves the tracking ref pointing at nothing | `BaseTracker.GetTrackedBase` returns `("", nil)` when the config key is absent; the cmd layer treats empty as "missing" and falls back to `ProtectedBranches[0]`. No crash. |
| The new `Rebaser` role at 4 methods exceeds the 1-3 method recommendation | Same shape as the existing `WorktreeWriter` (which the `core-git` spec explicitly accepts as a tight lifecycle set); documents the rationale via an inline note in `core-git/spec.md`. |
| Hook contract widening breaks an existing consumer that switched on `HookType` only | The widening is purely additive; existing `HookRunRequest{HookType: PostCreate, ...}` literals compile unchanged. A compile-time fixture in `internal/core/hook_types_test.go` asserts every old `HookType` still produces a usable `HookRunRequest`. |
| Composite `*git.Client` grows from 14 methods (6 roles) to 20 methods (8 roles), or equivalently from 6 to 8 roles | Roles stay split across `*reader` and `*cliClient`; the composite surface is unchanged. The drift sentinel is the only site that enumerates the surface; consumers reach role methods via embedded promotion. No consumer refactor required. |
| `twiggit sync --all` walk races with a concurrent `wtg create` adding a new worktree | The walk reads `/workspace/`.` once per project; new wtgs created mid-walk are reported as "skipped (added during enumeration)" with the existing `MissingXAction` skip-bucket pattern. Idempotent and conservative. |
| Rebase-conflict stderr phrase is git-version-dependent | The detector lives in one place (`internal/git/rebaser.go`) with a phrase table; future git-version skew updates one site with one test per phrase. Mirrors `isNotFoundStderr`. |

## Migration Plan

No migration steps. This is a pre-1.0 source-only addition; existing
behaviour is unchanged. Users who have built automation on the bare
`twiggit create` / `twiggit prune` lifecycle commands see no
behaviour change. Existing `PostCreate` hooks continue to fire; new
`PreRebase` / `PostRebase` / `PostSync` hooks are opt-in via the
`.twiggit.toml` config file.

## Open Questions

None. All design-level questions were resolved during exploration and
captured in the proposal's locked decisions; deferrable variants
(continue-past-conflict, cross-project fan-out, `--onto` flag,
`--force` past dirty wt) are explicit non-goals.