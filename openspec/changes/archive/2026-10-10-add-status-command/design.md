# Design

## Context

Twiggit's `internal/core` already declares a `core.WorktreeStatus` value type used by `cmd/delete.go:263 getWorktreeStatus` to refuse dirty worktrees. The type is narrow (clean/dirty + a `BranchStatus` summary string) and the helper lives in `cmd/` rather than the adapter layer — a layer leak that this change retires. The `core.RepositoryStatus.Ahead/Behind` fields are declared but never populated. The rebase command already implements a tracked-base resolution chain (per-worktree config → `Config.Validation.ProtectedBranches[0]` → `core.ErrBaseNotSet`) that status will mirror, so the two commands stay aligned on what "the base" means without inventing a parallel policy.

The new command's primary contract is read-only and best-effort: any per-worktree read failure is a row, not an error. This is the architectural shape that distinguishes status from `list` (which fails on the first read error and is therefore cheap and fast) and from `prune` (which refuses to delete a worktree it cannot read).

## Goals / Non-Goals

**Goals:**
- One canonical `Client.ReadWorktreeStatus` adapter method that all three consumers (`cmd/status`, `cmd/delete`, future prune optimisation) compose against.
- A clean extension of `core.WorktreeStatus` (six new fields with JSON tags) so the existing `core-types` data type absorbs the diagnostic view without a parallel type.
- A one-method role interface (`core.WorktreeStatusReader`) with the standard compile-time drift sentinel, so adding a method signature change breaks the build rather than silently diverging.
- Project-over-and-across output that mirrors `list`'s `--all` semantics, shares the `--output` vocabulary, and lands data on stdout / warnings on stderr per the `cli-output-formats` spec.

**Non-Goals:**
- New hook types — `status` is read-only.
- Mutating any worktree — `wtg rebase` is the right verb for that.
- Replacing `prune`'s `IsBranchMerged` walk in this slice; that refactor is a follow-up that consumes the new role.
- Remote fetch — ahead/behind is computed against the local base; `wtg rebase --fetch` is the existing way to refresh.
- TUI / interactive picker — the JSON output is the scripting surface.

## Decisions

### 1. `Client.ReadWorktreeStatus` lives on the composite `*Client`, not on `*reader` or `*cliClient`

**Choice:** the new method is a direct method on the composite `*Client` (not promoted from either embedded half).

**Rationale:** the read spans both halves. The dirty / clean state comes from go-git (`reader.RepositoryStatus`), the ahead/bebehind counts come from CLI (`git rev-list --count`), the merged check from CLI (`git branch --merged`, already in `cliClient.IsBranchMerged`), the base from CLI (`git config --worktree --get twiggit.tracked-base`, already in `cliClient.GetTrackedBase`), and the last-commit date from go-git (`reader.ListBranches`). Forcing the method onto one half would require the other half to be threaded through, defeating the consumer-side role pattern. The composite `*Client` already exists precisely for cross-cut methods that need both halves.

**Alternatives considered:**
- Put the method on `*cliClient` and have it call the reader via a passed-in interface. Rejected: inverts the embedding (the reader is the lower-dependency half; the cliClient should not know about it).
- Split the read into five separate role methods (`*Reader.IsClean`, `*Client.CountAheadBehind`, `*Client.IsMergedInto`, `*Client.LastCommitDate`, `*Client.GetTrackedBase`). Rejected: multiplies the role surface; the natural unit of work is "one worktree's diagnostic projection", not five individual facts.
- Add a third embedded half (`statusClient`). Rejected: a third half for one method is disproportionate; the existing composite absorbs the new responsibility.

The drift-sentinel comment in `internal/git/client.go` is amended to acknowledge that the composite now carries direct methods in addition to the promoted ones.

### 2. Per-worktree failures populate the row, not the error

**Choice:** the walk returns no error after a successful loop; each failed worktree carries `IsSkipped = true` and a `SkipReason`. A `*core.OperationError` returns only when the walk cannot run at all (config load, context resolution, project discovery).

**Rationale:** `status` is read-only. The user invokes it to see "where each worktree sits", and a single failed worktree should not hide the others. This is consistent with `kubectl get pods -o wide` showing a `STATUS` column with `Error` entries for unreachable nodes, and with `git status` showing nothing about a path that has been deleted mid-run. It also lets the user pipe the JSON output without the walk aborting on the first read error.

**Alternatives considered:**
- Return `errors.Join` of all per-worktree errors. Rejected: the join lives on the boundary that the user can act on, but the per-wt data is what they want; surfacing the error in `Main()` would convert the success case into an exit 1 and break the JSON pipeline.
- Aggregate per-wt errors and return the first. Rejected: hides information.
- Stop the walk on first failure. Rejected: opposite extreme; loses too much data.

### 3. `Config.Status` is a new sub-struct, not a flat `Config.StaleBehind` / `Config.StaleDays`

**Choice:** `core.Config.Status StatusConfig` with `StaleBehind int` and `StaleDays int`. TOML key prefix `[status]`.

**Rationale:** the existing pattern (`Config.Sync`, `Config.Rebase`, `Config.Validation`) groups related knobs under a sub-struct. The two stale flags are conceptually a unit ("when do I treat a worktree as stale?") and may grow (e.g. a future `--stale-no-fetch` flag or a `StaleAuthors` knob). A sub-struct is the natural container and matches the project's existing config shape. The TOML key prefix `[status]` is short and unambiguous.

**Alternatives considered:**
- Flat fields on `Config` (`StaleBehind int`, `StaleDays int`). Rejected: breaks the existing config grouping pattern; harder to extend.
- Put the thresholds on the `core.WorktreeStatus` type itself. Rejected: the value object must not carry configuration; the derivation is a function of the value and the config.

### 4. `--stale-behind 0` and `--stale-days 0` mean "disable this half of the heuristic"

**Choice:** a value of `0` on either flag or in either config field disables that half of the heuristic. The `IsStale` field is `true` only when at least one half trips.

**Rationale:** "disable" is the natural meaning of zero for a count. Setting both halves to zero disables the heuristic entirely, which is the only way a user can opt out without editing the type. Distinguishing "not set" from "set to 0" uses `cmd.Flags().Changed(name)` and falls through to the config default.

**Alternatives considered:**
- Use `*int` and treat `nil` as "unset". Rejected: pflag does not natively support `*int` without a custom flag; mixing two flag representations in one command is uglier than the `Changed()` check.
- Use a negative number to mean "disable". Rejected: negative counts are surprising in user-facing config.

### 5. `cmd/status` follows the existing `cmd/list` shape exactly, including the per-format projection pattern

**Choice:** `StatusOptions` mirrors `ListOptions` (Factory-injected IOStreams / Config / GitClient / Ctx / GlobalOptions / Logger; flags `All bool`, `StaleBehind int`, `StaleDays int`; `runF` test seam; `cobra.MaximumNArgs(1)`; Args validator). The `statusRows` Tabular projection lives in `cmd/status.go` and implements both `output.Tabular` (for `table`) and `json.Marshaler` (for the bare-array JSON), exactly as `worktreeRows` does for `list`.

**Rationale:** the project's `cli-output-formats` and `cli-flags-ergonomics` specs are clear that per-list-command output projections live in the cmd layer (presentation stays out of the core). Matching the `list` shape also keeps the diff minimal: a reviewer can compare `cmd/list.go` and `cmd/status.go` side by side and find the deltas (`worktreeRows` becomes `statusRows`, three columns become eight, additional flags). The `--all` and `--output` semantics are already battle-tested in `list`; status inherits them.

**Alternatives considered:**
- A shared `rows` helper between `list` and `status`. Rejected: the projections diverge (status has ahead/behind/base/etc., list does not); abstracting now is premature.

## Risks / Trade-offs

- **[Per-worktree read cost]** — three to four git calls per worktree, sequential. For a project with 50 worktrees that is ~200 calls (~2s on a cold cache). → Mitigation: defer the parallelisation; the existing pattern (sequential) is simpler and matches `prune` and `rebase --all`. Revisit if the field data shows the cost hurts the UX. The cache (`lru.Cache` on `*reader`) keeps the go-git path hot across calls.
- **[Stale heuristic false positives]** — a feature branch that intentionally lives for 31 days with 0 commits behind main is "stale" by the heuristic. → Mitigation: the column is informational, not actionable; no `prune --stale` exists yet to amplify the noise. The thresholds are user-tunable via flags and config.
- **[Layer violation in existing code]** — `cmd/delete.go:263 getWorktreeStatus` is in the wrong layer. → Mitigation: this change moves it to `internal/git/status.go` as `Client.ReadWorktreeStatus`; `cmd/delete.go` calls the new role. The mechanical refactor is part of the same change so the diff is honest.
- **[Direct method on composite `*Client`]** — the existing drift-sentinel comment claims the method set is the union of the two halves; the new method violates that claim. → Mitigation: the comment is amended; the compile-time guard is updated to acknowledge the direct method. A future refactor could split the read into per-fact methods on each half, but that costs more than it saves at one call site.
- **[JSON shape additions]** — adding fields to `core.WorktreeStatus` changes the JSON projection for any existing consumer. The only existing consumer is the in-tree `cmd/delete.go:296` projection, which builds a custom struct and never serialises. → Mitigation: zero external consumers; no migration. Documented in the spec for future consumers.
- **[`core.RepositoryStatus.Ahead/Behind` are now actually populated]** — this is a behavior change for any code that reads those fields. The only reader is the new `Client.ReadWorktreeStatus`. → Mitigation: zero existing readers; the field is filled only when the new role populates the projection.

## Migration Plan

No migration steps. The change is source-only and pre-1.0. No public API breaks (no exported types are removed or renamed; no CLI flags are removed). New command `status` joins the `core` group; existing commands (`list`, `prune`, `rebase`, `sync`, `create`, `delete`, `cd`, `init`, `version`, `completion`) are untouched in their surface.

Rollback: `git revert` of the change removes the new command, the new role, the type extensions, and the moved helper. The `core.RepositoryStatus.Ahead/Behind` fields return to their previous "declared, unpopulated" state (which is the current state pre-change). No data store, no config file, no on-disk state is touched.

## Open Questions

None. The five open questions surfaced in the exploration phase (default output columns, `LastCommitDate` source, `getWorktreeStatus` relocation, two-spec vs `MODIFIED`, stale defaults) were resolved with the user before the change was proposed. No new unknowns remain that would change the specs, the approach, or the task breakdown.
