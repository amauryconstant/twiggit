# Proposal

## Why

The v0.16.0 → v0.16.8 release cascade (ten CI-touching patches in five days) exposed ten evergreen invariants that currently live only as historical prose in `docs/release/AGENTS.md`. The same period also surfaced a ghost spec: the 194-line `infrastructure-release` delta created by the 2026-09-28 `infra-release-hardening` change was never synced into `openspec/specs/`, leaving the canonical home of release/CI invariants ambiguous. This change consolidates both into a single main spec (`infrastructure-release`) and folds the prose mirror plus operator-facing recipes into the root `AGENTS.md`, eliminating `docs/release/AGENTS.md`.

## What Changes

- Create `openspec/specs/infrastructure-release/spec.md` (main spec) — synced from the 2026-09-28 archived delta plus ten new ADDED Requirements derived from the v0.16 cascade. The archived delta file under `openspec/changes/archive/` is **not** edited; this change writes the canonical wording fresh.
- Append a **Release / CI invariants** block to root `AGENTS.md` — bracketed `[release-*]` prose mirror following the same pattern `openspec/AGENTS.md` uses for `[code-conv]` rules. Cross-references `infrastructure-release` as the authority.
- Append a **Release process** block to root `AGENTS.md` — operator-facing only (distribution surface table, `mise run release:*` commands, cosign `verify-blob` recipe, per-surface GitLab-vs-GitHub architecture note). No invariants.
- Remove the `docs/release/AGENTS.md` row from the root `AGENTS.md` subdirectory table.
- Delete `docs/release/AGENTS.md` and the now-empty `docs/release/` directory.

No code, CI, Docker, workflow, or `go.mod` change. The archived change folder under `openspec/changes/archive/2026-09-28-infra-release-hardening/` is read-only reference material — it is not edited.

## Capabilities

### New Capabilities

- `infrastructure-release`: canonical home for twiggit's CI and release pipeline invariants. Owns the goreleaser/cosign/syft toolchain pin contract, the GitLab and GitHub Actions release surface contracts, release-tag reuse protection, module-graph freshness gating, and the per-surface architecture decisions governing how the two CI vendors publish the same release. The 2026-09-28 `infra-release-hardening` change created this delta but never synced it; this change performs the sync and adds the missing v0.16 cascade invariants.

### Modified Capabilities

None. No existing spec has its requirements changed. `version-package` is adjacent (build-time ldflags) but its requirements are unchanged.

## Impact

### Layer rollup

This is a docs/spec change. No `cmd/`, no `internal/`, no `.gitlab-ci.yml`, no `.github/workflows/*`, no `Dockerfile.ci`, no `Dockerfile.release`, no `mise/*`, no `go.mod`. The change is observable only as a re-homing of canonical invariants (spec + root `AGENTS.md`) and a deletion of `docs/release/AGENTS.md`.

### Files modified

```
AGENTS.md (root)                                      — Append "Release / CI invariants"
                                                        and "Release process" sections;
                                                        remove docs/release/AGENTS.md row
                                                        from subdir table
openspec/changes/consolidate-release-invariants/      — New change folder (proposal,
                                                        design, tasks, specs delta)
openspec/specs/infrastructure-release/spec.md         — New main spec (synced from
                                                        the 2026-09-28 delta plus
                                                        10 cascade-derived requirements)
openspec/AGENTS.md                                    — One-line addition to "Active
                                                        spec cross-references" list
                                                        naming infrastructure-release
```

### Files deleted

```
docs/release/AGENTS.md                                — Content moves to root AGENTS.md
docs/release/                                         — Empty after deletion; removed
```

### Files unchanged (within scope, asserted)

- `openspec/changes/archive/2026-09-28-infra-release-hardening/**` — read-only reference material per user constraint. Not edited.
- `cmd/**`, `internal/**`, `go.mod`, `go.sum` — implementation untouched.
- `.gitlab-ci.yml`, `.github/workflows/**`, `Dockerfile.ci`, `Dockerfile.release`, `.mise/**`, `.goreleaser.yml` — CI/release behavior untouched. This change is purely a docs/spec re-homing; no release job, no workflow step, no YAML line is altered.

### Operational impact

None at runtime. The verification gate (`mise run verify`) still applies because the change touches markdown files covered by `mise run lint:md` and prettier formatting; both gates must pass on the new content.

### Compatibility

No user-visible break. The content of `docs/release/AGENTS.md` is preserved in root `AGENTS.md`; any agent that previously read `docs/release/AGENTS.md` now reads the equivalent sections in root `AGENTS.md` instead. The bracketed `[release-*]` keys are a new convention in this repo (mirroring `[code-conv]` in `openspec/AGENTS.md`); they do not change any existing rule.

### Risks and mitigations

- **[R1] Sync wording diverges from the archived delta.** Mitigated by sourcing the wording from the archived delta at `openspec/changes/archive/2026-09-28-infra-release-hardening/specs/infrastructure-release/spec.md` and preserving every requirement intent verbatim — the new main spec is a re-statement, not a rewrite. Any reviewer can diff the two.
- **[R2] Per-surface architecture permission (Requirement: surface divergence MAY) reads as weaker than the archived prescriptive rule.** Mitigated by the cascade evidence in the design rationale: prescriptive unification was the failure mode that produced the five-environment cascade. The permissive form is the evergreen lesson.
- **[R3] Operator recipes in root `AGENTS.md` look like invariants.** Mitigated by the section heading ("Release process") and by the explicit invariant/process split in the design rationale. Recipes are surfaced in their own block, separately from the bracketed invariant rules.
