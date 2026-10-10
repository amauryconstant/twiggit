# Proposal

## Why

Twiggit's user-facing surface lets the user create, list, prune, rebase, and sync worktrees, but offers no single command that answers "where does each worktree sit relative to its tracked base right now?" `list` reports three columns (branch, path, dirty status). `prune` catches merged worktrees but cannot see ahead/behind or staleness. The rebase walk can run blind onto a stale base. The user is left assembling the picture by hand: `cd wt && git fetch && git rev-list --count main..HEAD && git log -1 --format=%ci` per worktree. The README's "focus on rebase workflows" intent is undermined by the missing delta view that makes rebase decisions informed.

`core.WorktreeStatus` already exists and is partially populated by `cmd/delete.go` for the dirty check, but the function lives in `cmd/` (a layer leak) and the projection is narrow (dirty + a `BranchStatus` string). No role interface, no command, no rich projection, no integration with the rebase tracked-base chain that already exists.

## What Changes

- Add a new `status` command in the `core` group that emits a per-worktree projection of `core.WorktreeStatus`: branch, path, ahead, behind, base, merged, dirty, last-commit-date, stale.
- Add a `WorktreeStatusReader` role interface (one method) in `internal/core/git.go`; implement it as `Client.ReadWorktreeStatus` in `internal/git/status.go` on the composite `*Client` (the read spans go-git and CLI halves).
- Extend `core.WorktreeStatus` with six new fields: `Base`, `IsMerged`, `IsStale`, `LastCommitDate`, `IsSkipped`, `SkipReason`. Populate the existing `RepositoryStatus.Ahead` and `Behind` so the `BranchStatus` string is meaningful.
- Move the existing `getWorktreeStatus` helper from `cmd/delete.go` into the new `Client.ReadWorktreeStatus` so the read path has one canonical implementation. `cmd/delete.go` calls the new role.
- Add a `Config.Status` sub-struct with `StaleBehind` (default 20) and `StaleDays` (default 30) knobs; expose `--stale-behind` / `--stale-days` flags on the new command.
- Add a compile-time drift sentinel `var _ core.WorktreeStatusReader = (*Client)(nil)` in `internal/git/client.go`.

No breaking changes. No deletions. No changes to existing CLI surface.

## Capabilities

### New Capabilities

- `cli-status`: the `twiggit status` command surface — flags (`--all`, `--output`, `--stale-behind`, `--stale-days`), positional argument, output shapes (default wide layout, `json` bare array, `table` aligned, `plain` headerless TSV), per-worktree best-effort error contract, exit-code contract, command-group placement.
- `core-worktree-status`: the `core.WorktreeStatus` value type extensions, the `WorktreeStatusReader` role interface, the `IsStale` derivation, the per-worktree best-effort skip contract, the drift sentinel placement.

### Modified Capabilities

None. `core-types` already documents `core.WorktreeStatus` only as a service result; the extensions and the new role live under the new `core-worktree-status` capability. `core-git` cross-references the new role from its role-interface segregation requirement.

## Impact

### Production (cmd + internal + config)

- `cmd/status.go` (new): `StatusOptions`, `NewCmdStatus`, `runStatus`, `statusRows` Tabular projection, `MarshalJSON` for the bare-array JSON shape.
- `cmd/delete.go` (modify): replace the local `getWorktreeStatus` (lines 262-303) with a call to `client.ReadWorktreeStatus` via the new role; remove the local helper.
- `cmd/root.go` (modify): register `NewCmdStatus(f, nil)` in the `core` group alongside `list`/`create`/`delete`/`prune`.
- `internal/core/git.go` (modify): add `WorktreeStatusReader` interface declaration (one method, per consumer-side role pattern).
- `internal/core/service_results.go` (modify): extend `WorktreeStatus` with `Base string`, `IsMerged bool`, `IsStale bool`, `LastCommitDate time.Time`, `IsSkipped bool`, `SkipReason string`; add JSON tags to every exported field.
- `internal/core/config.go` (modify): add `StatusConfig` sub-struct (`StaleBehind int`, `StaleDays int`) under `Config.Status`; wire defaults in `DefaultConfig`.
- `internal/git/client.go` (modify): add `var _ core.WorktreeStatusReader = (*Client)(nil)`; amend the existing drift-sentinel comment to acknowledge the direct method on the composite.
- `internal/git/status.go` (new): `func (c *Client) ReadWorktreeStatus(ctx, cfg, repoPath, wtPath) (core.WorktreeStatus, error)`; populate all fields; return best-effort per-wt error contract.

### Tests (unit + integration + e2e + golden)

- `internal/core/service_results_test.go` (modify): zero-value safety for the new fields, JSON round-trip, `IsStale` derivation table-driven over six cases (fresh, behind, aged, both, neither, disabled thresholds).
- `internal/git/status_test.go` (new): mocked `CommandExecutor`; ahead/behind parse from `git rev-list --count` stdout, merged-detection path, base fallback chain (tracked-base → protected[0] → error), per-wt skip on `IsBranchMerged` failure.
- `internal/git/client_test.go` (modify): extend the role-satisfaction drift test to cover `WorktreeStatusReader`.
- `cmd/status_test.go` (new): parse-level (every flag), run-level with `runF` injection and fakes; golden for the four output shapes (default wide, `json`, `table`, `plain`); outside-git → `UsageError`; `--all` + positional mutually exclusive.
- `test/integration/status_test.go` (new, `//go:build integration`): real git repo with main + 2 features (one ahead, one merged, one clean, one with `twiggit.tracked-base` set, one dirty).
- `test/e2e/status_test.go` (new, `//go:build e2e`): CLI smoke + golden (`UPDATE_GOLDEN=true mise run test:e2e`).
- `test/golden/status/` (new): captured output for the four formats.

### Docs (README + AGENTS.md)

- `README.md` (modify): add `twiggit status` to the Quick Start and a brief section describing the columns and the stale heuristic.
- `openspec/AGENTS.md` (modify): mention the new role and the move of `getWorktreeStatus` in the `core-git` cross-reference note.
- `AGENTS.md` (modify): update the architecture table to list `internal/git/status.go` under the git I/O adapter row and the `Status` config knob row.

### CI

- No CI changes. The existing `mise run verify` pipeline (format + lint:fix + lint:gated + vuln:check + test + build) covers the new code via the standard test targets.

## Non-goals

- No new hook types. `status` is read-only.
- No changes to `prune`'s `IsBranchMerged` walk; the new adapter could replace it (saves N calls per prune), deferred to a follow-up.
- No `wtg status --rebase-needed` filter; v0.15 ships the rows first, the filter is a v0.16 candidate.
- No TUI / interactive picker. The JSON output is the scripting surface.
- No remote fetch. The ahead/behind numbers are computed against the local base. `wtg rebase --fetch` is the existing way to refresh before status.
- No `--base` override; the rebase command's tracked-base fallback chain is the single source of truth. Status mirrors it.
