# Proposal

## Why

Twiggit's README declares "focus on rebase workflows", yet no `rebase` (or `sync`)
command exists — only the lifecycle verbs (list/create/delete/prune) plus `cd`.
Users rebase manually from each worktree (`git fetch && git rebase origin/main`),
which scales poorly across projects and re-introduces the human-error surface that
worktree management exists to remove. This change closes the philosophy-to-surface
gap with two new commands that share one inner machinery.

## What Changes

- Add `twiggit rebase [project/branch]` and `twiggit rebase --all` (project
  fan-out) with `--fetch`, `--continue`, `--abort`, `--set-base` verbs. Rebase
  operates offline by default; `--fetch` is opt-in. Conflicting worktrees halt
  fan-out and are reported with mid-rebase state intact.
- Add `twiggit sync [project]` and `twiggit sync --all` (project fan-out) with
  `--remote`, `--branch`, `--rebase`, `--fetch-only` flags. Sync fetches remote
  tracking refs into the project main worktree; `--rebase` runs the rebase
  walk afterwards so wts track the updated base.
- Persist each worktree's tracked base in git per-worktree config
  (`twiggit.tracked-base = <branch>`) so the rebase target survives path
  moves and is discoverable via `git config --worktree`. `twiggit create`
  writes the config on success; `rebase --set-base` mutates it on demand.
- Extend the `HookType` enum with `HookTypePreRebase`, `HookTypePostRebase`,
  `HookTypePostSync` and widen `HookRunRequest` with optional fields
  (`RebaseBase`, `RebaseOldTip`, `RebaseNewTip`, `RebaseResult`,
  `SyncRemote`, `SyncBranch`). Pre-rebase hook failure aborts the rebase;
  post-rebase and post-sync hook failures log warnings (work already done).
- Add two consumer-side role interfaces (`Rebaser`, `BaseTracker`) plus
  concrete methods on `*git.cliClient`; composite `*git.Client` exposes
  them via embedded promotion. Drift sentinel extended in
  `internal/git/client_test.go`.
- Add two config sections (`[rebase]`, `[sync]`) loaded via the existing
  koanf merge order (defaults → file → env → flags). Three new
  `core.Err*` sentinels join `core-errors` taxonomy
  (`ErrRebaseConflict`, `ErrRebaseInProgress`, `ErrBaseNotSet`).
- Add `rebase` and `sync` to the `core` cobra command group.

## Non-goals

- Cross-project fan-out for either command (deferred; project-only for v1).
- `twiggit rebase --onto <branch>` (users run `git rebase --onto` inside the
  worktree; covered by `cd` shell wrapper).
- `--force` rebase past dirty worktrees (dangerous; users stash, rebase,
  pop).
- Automated conflict resolution or interactive rebase script wrapper
  (`git rebase -i` is unchanged).
- Worktree workspaces / session groups / templates / TUI picker. These are
  distinct features surfaced in earlier exploration; each is a separate change
  if requested.

## Capabilities

### New Capabilities

- `cli-rebase`: command surface, flag set, output format, exit-code dispatch,
  and shell-wrapper navigation contract for `twiggit rebase`.
- `cli-sync`: command surface, flag set, output format, and exit-code
  dispatch for `twiggit sync`, including the shared rebase walk when
  `--rebase` is set.
- `core-rebase`: value objects (`RebaseRequest`, `RebaseResult`,
  `RebasedWorktree`, `RebaseOutcome` enum, `SyncRequest`, `SyncResult`,
  `SyncedBranch`) and the three new sentinels.
- `git-rebaser`: `core.Rebaser` (`Rebase`, `Abort`, `Continue`, `Fetch`)
  and `core.BaseTracker` (`SetTrackedBase`, `GetTrackedBase`) role
  interfaces, the CLI adapter on `*git.cliClient`, and the routing table
  entry that pins the CLI implementation per `git-client`.

### Modified Capabilities

- `core-hook-types`: add `HookTypePreRebase`, `HookTypePostRebase`,
  `HookTypePostSync` to the enum; widen `HookRunRequest` with the six
  optional fields listed above; extend the env-var table with
  `TWIGGIT_REBASE_BASE`, `TWIGGIT_REBASE_OLD_TIP`,
  `TWIGGIT_REBASE_NEW_TIP`, `TWIGGIT_REBASE_RESULT`,
  `TWIGGIT_SYNC_REMOTE`, `TWIGGIT_SYNC_BRANCH`. Additive; existing
  post-create callers compile without change.
- `core-git`: add the `Rebaser` and `BaseTracker` roles; expand the
  composite-satisfaction and drift-sentinel requirements to cover them.
- `git-config`: add `[rebase]` (`fetch_on_all`, `conflict_policy`) and
  `[sync]` (`default_remote`, `prune_remote_refs`, `rebase_after_sync`)
  sections, loaded via the existing koanf precedence.
- `cli-command-groups`: add `rebase` and `sync` to the `core` group.
- `core-errors`: document the three new sentinels and the per-entity
  dispatch (no new `core.Error` subtypes).

## Impact

### Production code (grouped by layer)

| Layer | File | Change |
|---|---|---|
| `cmd/` | `cmd/rebase.go` (new) | `RebaseOptions`, `NewCmdRebase`, `runRebase`, args validator |
| `cmd/` | `cmd/rebase_helpers.go` (new) | `resolveRebaseTargets`, `resolveTrackedBase`, `dispatchRebaseHooks`, `emitRebaseOutput` |
| `cmd/` | `cmd/sync.go` (new) | `SyncOptions`, `NewCmdSync`, `runSync`, project walk |
| `cmd/` | `cmd/sync_helpers.go` (new) | `resolveSyncTargets`, shared rebase-walk invocation |
| `cmd/` | `cmd/root.go` | register `rebase` + `sync` in `core` group (+4 LOC) |
| `cmd/` | `cmd/create.go` | call `gitClient.SetTrackedBase` after successful `CreateWorktree` (+5 LOC) |
| `internal/core/` | `core/rebase.go` (new) | request/result types + outcome enum |
| `internal/core/` | `core/hook_types.go` | 3 new enum variants + 6 optional fields |
| `internal/core/` | `core/config.go` | `RebaseConfig`, `SyncConfig` fields |
| `internal/core/` | `core/git.go` | `Rebaser`, `BaseTracker`, `RebaseOutcome` |
| `internal/core/` | `core/sentinels.go` | 3 new sentinels |
| `internal/git/` | `git/rebaser.go` (new) | `Rebaser` + `BaseTracker` CLI adapter on `*cliClient` |
| `internal/git/` | `git/client.go` | 2 new `var _ core.Role = (*git.Client)(nil)` declarations |
| `internal/git/` | `git/client_test.go` | drift sentinel struct embeds the 2 new roles |
| `internal/git/` | `git/errors.go` | `NewRebaseError`, `NewBaseTrackerError` |

### Tests (co-located with implementation)

| Test file | Role |
|---|---|
| `cmd/rebase_test.go`, `cmd/rebase_helpers_test.go` | unit, runF seam |
| `cmd/sync_test.go`, `cmd/sync_helpers_test.go` | unit, runF seam |
| `internal/core/rebase_test.go` | value-object + enum tests |
| `internal/core/hook_types_test.go` | extended enum + widened request |
| `internal/git/rebaser_test.go` | adapter tests with `MockCommandExecutor` |
| `internal/git/client_test.go` | drift sentinel extension |
| `test/integration/rebase_test.go` | real git CLI: clean rebase, conflict, continue, abort, set-base, hooks, fallback |
| `test/repo/repo.go` | extended helpers: `CreateDivergentBranch`, `CreateDirtyWorktree`, `CreateMidRebase`, `AdvanceBase` |

### E2E (golden + Ginkgo)

| E2E file | Role |
|---|---|
| `test/e2e/rebase_test.go` (new) | ~10 specs covering cwd, named wt, `--all`, `--fetch`, `--continue`, `--abort`, `--set-base`, dirty refuse, fallback, usage error outside git |
| `test/e2e/sync_test.go` (new) | ~6 specs covering single project, `--all`, `--remote`, `--branch`, `--rebase`, `--fetch-only` |
| `test/e2e/workflows/rebase_workflow_test.go` (expand existing) | end-to-end create → work → rebase → push flow with shell-wrapper navigation |

### Docs

| Doc | Role |
|---|---|
| `README.md` | new `twiggit rebase` / `twiggit sync` sections with examples |
| `cmd/AGENTS.md` | command-shape reference |
| `openspec/AGENTS.md` | workflow reference (no rule changes) |

### Risk and mitigation

- **Detecting rebase-conflict vs nothing-to-do from stderr** is git-version-dependent. Mitigation: centralise the detector in `internal/git/rebaser.go` (mirror `isNotFoundStderr` precedent) with table-driven tests per stderr phrase.
- **`git config --worktree` requires git ≥ 2.20** (Mar 2018). Mitigation: minimum git is implicit in twiggit's existing dependency surface; document the floor in `git-config` spec and fail with a clear message if missing.
- **`--all` halts on first conflict** may feel overly strict for users with many wts who want "skip conflict, do the rest". Mitigation: the `conflict_policy` config knob is reserved in `[rebase]`; `stop` is the v1 default. Future policy values (`continue`, `abort`) are deliberately deferred to avoid speculation.
- **Composite `*git.Client` grows** from 13 to 19 role methods. Mitigation: roles stay split across `*reader` and `*cliClient`; the composite surface is unchanged. No consumer is affected beyond the new rebase and search commands.
- **Hook contract extension breaks existing post-create callers** if `HookRunRequest` widening is mishandled. Mitigation: all six new fields are additive and string-empty by default; existing `PostCreate` requests compile unchanged. Compile-time fixture in `internal/core/hook_types_test.go` asserts every old `HookType` still produces a usable `HookRunRequest`.