# Coordination note: architecture-layer-inversion → quality-modernization

**Status:** Architectural coordination memo. Not part of either change's spec or tasks.

**Author:** phase 0 of `architecture-layer-inversion` (preflight).

**Audience:** whoever next edits `openspec/changes/quality-modernization/`.

## Context

`architecture-layer-inversion` and `quality-modernization` (both active) touch the same files with overlapping intents. The two changes cannot merge cleanly if both edit the same lines. The user (project owner) decided to **defer `architecture-layer-inversion` until `quality-modernization` lands first**, then **amend `quality-modernization` tasks.md** to drop the now-redundant tasks.

## Tasks in `quality-modernization/tasks.md` that become redundant

If `architecture-layer-inversion` lands first (i.e. merged before quality-modernization's implementation work starts), the following quality-modernization tasks are already covered and should be **dropped** from that change's `tasks.md` (or marked as "covered by architecture-layer-inversion" with a pointer to that change):

| Task | Coverage |
|------|----------|
| 7.7 — Remove redundant `//nolint:wrapcheck` x4 in `command_executor_mock_test.go` | Covered by architecture-layer-inversion slice 8 (task 8.6) |
| 11.2 — `strings.CutPrefix` in `cli_client.go:25-26,280-287,297-304` | Covered by slice 5c-ii (tasks 5c-ii.2). Sites :280-287 and :297-304 are deferred per overlap note in slice 5c-ii commit body. |
| 11.5 — Replace local `ProjectRef` with `domain.ProjectSummary` in `context_resolver.go` | Covered by slice 5c-iv (task 5c-iv.1) |
| 11.6 — Bound commit hash slice at `gogit_client.go:336` with `min(7, len(...))` | Covered by slice 5c-ii (task 5c-ii.3) |

## Tasks in `quality-modernization/tasks.md` that have partial overlap

These tasks remain valid in principle but specific line numbers in their description have moved or been absorbed. The implementation phase of quality-modernization should re-grep the affected files before applying:

| Task | Partial overlap |
|------|-----------------|
| 2.1 — Add `gofumpt` + `goimports` formatters | Compatible with architecture-layer-inversion's depguard work, but the change must land AFTER `.golangci.yml` slice 7 lands (otherwise existing formatters fail mid-stream) |
| 2.2 — Tighten `funlen.lines` 150→120 | Conflicts with the `funlen.lines: 150` default architecture-layer-inversion keeps. Architecture-layer-inversion does NOT tighten thresholds. Quality-modernization can tighten after this change lands. |
| 2.3 — Tighten `gocyclo.min-complexity` 25→13 | Same as 2.2 — `gocyclo.min-complexity: 25` retained by architecture-layer-inversion |
| 11.3 — Replace `context.Background()` in `context_resolver.go` | Covered structurally by slice 5c-iv.1 (resolver rewrite). Quality-modernization task 11.3 must re-grep because the variable names (`cr.gitService`, `cs.gitService`) it references no longer exist post-inversion. |
| 11.4 — Split `parseWorktreeList` in `cli_client.go` | Architecture-layer-inversion adds nil-guards at 6 sites in the same file; quality-modernization's split is orthogonal but must run AFTER the nil-guard work to avoid merge conflicts on the same file |
| 11.8 — Preallocate slices at multiple sites | Sites `internal/service/worktree_service.go:179` and `internal/service/navigation_service.go:83,118` are absorbed by architecture-layer-inversion slice 4 (service rewrite). The other sites (cmd/, internal/infrastructure/) remain valid. |
| 11.9 — Rewrite `sh -c "<exports>; <cmd>"` in `hook_runner.go` | Orthogonal to architecture-layer-inversion's slice 5d.6 (no-op result block collapse). Both changes touch `hook_runner.go`; apply in this order: architecture-layer-inversion 5d.6 → quality-modernization 11.9. |
| 14.2 — Add godoc to exported infra methods | Remains valid; orthogonal. |

## Tasks in `quality-modernization/tasks.md` unaffected

Tasks 2.x (gofumpt/threshold tightening — except for the threshold conflict above), 7.x (other linter fixes), 11.1 (other `os.IsNotExist` sites not in slice 5c-ii), 11.7+ (other safety fixes), and all godoc tasks remain in scope.

## Ordering recommendation

If implementing both changes in series:

1. `quality-modernization` first, with the redundant tasks **kept** for now (they'll be removed during the architecture-layer-inversion implementation, OR post-merge during the archive pass).
2. Then `architecture-layer-inversion` removes the now-redundant tasks from this note as they get absorbed.

If implementing in parallel (not recommended — same-file edits cause conflicts):

- Slice 5d (`architecture-layer-inversion`) and tasks 11.4 / 11.9 (`quality-modernization`) cannot run in parallel — same files, different intents.
- All other slices are file-disjoint and can run in parallel IF each uses its own worktree.

## Open conflict (deferred to PR coordinator)

`internal/application/AGENTS.md` documents `application.GitClient` as a canonical interface. `architecture-layer-inversion` slice 2 deletes it. Either:

- architecture-layer-inversion slice 2 updates AGENTS.md (recommended).
- quality-modernization keeps the AGENTS.md as documentation of a non-existent interface (breaks readers).

architecture-layer-inversion slice 2 updates `internal/application/AGENTS.md` to drop the `GitClient (Composite)` row. quality-modernization should not re-introduce it.
