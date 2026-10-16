# Design

## Context

The v0.16.0 → v0.16.8 release cascade produced ten CI-touching patches in five days. Each patch was failure-driven; the only structural exceptions were the mise task migration and a prettier markdown pass. The cascade exposed evergreen invariants that now live only as historical prose in `docs/release/AGENTS.md` (177 lines, mixing invariants with operator recipes). The same period left an orphaned spec: the 194-line `infrastructure-release` delta created by the 2026-09-28 `infra-release-hardening` change was never synced into `openspec/specs/`.

Constraints shaping this design:

- The archived change folder under `openspec/changes/archive/2026-09-28-infra-release-hardening/` is read-only historical material; it SHALL NOT be edited.
- The change MUST NOT alter `.gitlab-ci.yml`, `.github/workflows/*`, `Dockerfile.ci`, `Dockerfile.release`, `.goreleaser.yml`, `mise/**`, or `go.mod`. CI/release behavior is preserved exactly.
- The repo already has a working precedent for invariant discipline — `style-enforcement` is the spec, `openspec/AGENTS.md` is the prose mirror with `[bracketed-key]` rules. The release/CI invariants SHOULD follow the same pattern, not invent a new structure.
- The four canonical category prefixes (`cli-`, `core-`, `git-`, `testing-`) do not cover CI/release surfaces. The archived delta uses `infrastructure-release` as a flat name; this design continues that naming. No new category prefix is added to `openspec/config.yaml` because `infrastructure-release` is a single capability, not a category — the prefix-as-category rule applies only to specs that share a layer.

## Goals / Non-Goals

**Goals:**

- Re-home twiggit's release/CI invariants into a canonical main spec at `openspec/specs/infrastructure-release/spec.md` with twenty testable requirements (ten from the archived delta + ten from the v0.16 cascade).
- Add a prose mirror to root `AGENTS.md` so agents reading the project primer see the invariants at-hand, without leaving the canonical spec.
- Fold operator-facing recipes (distribution surface table, `mise run release:*` commands, cosign `verify-blob` recipe, per-surface architecture note) into the root `AGENTS.md` Release process section.
- Delete `docs/release/AGENTS.md` and the now-empty `docs/release/` directory.
- Preserve the existing root `AGENTS.md` primer structure (architecture table, config knobs, commands, CLI surface, conventions, subdirectory guide) and only append the two new sections.

**Non-Goals:**

- Introducing any new release-time behavior. No CI YAML change, no workflow step change, no Dockerfile change, no goreleaser config change. This is a docs/spec re-homing only.
- Editing the archived delta at `openspec/changes/archive/2026-09-28-infra-release-hardening/`. The new main spec is written fresh, drawing on the archived wording as reference.
- Reframing the ten cascade-derived requirements as anything other than evergreen invariants. The cascade commits remain in git history; the spec rules are stated forward, not backward.
- Adding new tests. The change is docs/spec; `mise run verify` runs against the modified markdown to ensure prettier and golangci-lint still pass, but no Go test is added.
- Renaming any existing spec. `version-package` is adjacent (build-time ldflags) but its requirements are unchanged.
- Adding the `infrastructure-` token to `openspec/config.yaml`'s canonical prefix list. `infrastructure-release` is one capability, not a category; the prefix-as-category rule applies only to specs that share a layer.

## Decisions

### 1. Sync wording: re-state from the archived delta, do not modify the archive

**Choice:** The new main spec writes twenty ADDED Requirements (ten restated from the archived delta + ten new from the cascade) at `openspec/specs/infrastructure-release/spec.md`. The archived delta at `openspec/changes/archive/2026-09-28-infra-release-hardening/specs/infrastructure-release/spec.md` is not edited.

**Rationale:** Per user constraint ("Do not edit archived files, only the live spec"), the archived change folder is read-only. The new main spec is the live spec. Restating the ten archived requirements preserves their intent verbatim — any reviewer can diff the archived delta against the new main spec to confirm zero semantic drift. This is the same pattern as a code change that re-implements an archived prototype: the prototype stays for history, the new implementation is authoritative.

**Alternatives considered:**

- _Edit the archived delta to match the new main spec._ Rejected per user constraint and against the principle that archived change folders are immutable historical record.
- _Copy the archived delta's spec.md verbatim into `openspec/specs/infrastructure-release/spec.md` without the ten cascade additions._ Rejected — the cascade learnings are the primary motivation for this change; a pure sync without additions does not close the gap.
- _Create a separate spec for the cascade learnings and cross-reference._ Rejected — splitting the invariants across two specs makes the canonical home ambiguous, the same gap this change was meant to close.

### 2. Bracketed-key prefix: `[release-*]`

**Choice:** The prose mirror in root `AGENTS.md` uses bracketed keys prefixed `[release-*]`, matching the long-form convention already established in `openspec/AGENTS.md` for `[code-conv]` rules.

**Rationale:** The existing prose mirror (`openspec/AGENTS.md` §Code conventions enforced via golang-* skills) uses long-form bracketed keys (`[errors-checked]`, `[run-e-only]`, etc.). The short forms (`[rel-*]`, `[relci-*]`) would break the grep-for-rule convention agents already use on `[code-conv]` keys. Consistency wins; the verbosity is bounded — only twenty keys, in one section.

**Alternatives considered:**

- _Short forms `[rel-*]` or `[relci-*]`._ Rejected for inconsistency with the existing convention.
- _No bracketed keys, prose-only._ Rejected — bracketed keys are the project-wide convention for agent-grep rules; breaking the convention makes the new invariants less discoverable.

### 3. Per-surface divergence rule: permissive ("MAY diverge") rather than prescriptive

**Choice:** Requirement: Per-surface CI architecture MAY diverge codifies the divergence permission as `MAY`, not `SHALL`. The cascade is the evergreen evidence that prescriptive unification causes cascading failures.

**Rationale:** The v0.16 cascade showed that a unified architecture (sparse `golang:1.27.2-alpine3.23` base + per-run installer) cascaded through five distinct environmental failures on GitHub Actions while the same unified pattern was stable on GitLab CI. Prescribing unification would have re-introduced the cascade on the next go.mod bump. Permissive divergence lets each surface keep the architecture that has demonstrably survived.

**Alternatives considered:**

- _Prescriptive: "Release surfaces SHALL NOT diverge."_ Rejected — the cascade IS the failure mode this prescription would lock in.
- _Documented-as-exception: "Divergence requires a tracked ADR."_ Rejected — adds ceremony without changing the invariant. The evergreen rule is "MAY diverge if the unified architecture cascades"; the ADR is unnecessary because the cascade is the evidence.

### 4. Cosign recipe placement: root `AGENTS.md`, not a separate `docs/release/VERIFICATION.md`

**Choice:** The cosign `verify-blob` recipe, the `mise run release:*` commands, the distribution surface table, and the per-surface architecture note all live in the root `AGENTS.md` "Release process" section. No `docs/release/` subtree survives.

**Rationale:** Per user constraint ("Find a way to get rid of docs/release/AGENTS.md file"), the operator-facing content must move. Folding it into root `AGENTS.md` preserves a single canonical home and avoids re-creating a doc subtree whose only content is a cosign recipe. The recipe is short (~20 lines of `cosign verify-blob` blocks) and sits naturally next to the invariants it implements.

**Alternatives considered:**

- _New `docs/release/VERIFICATION.md` for the cosign recipe, separate from root `AGENTS.md`._ Rejected — preserves the `docs/release/` subtree in violation of the user's "get rid of" instruction.
- _Move the cosign recipe into `openspec/specs/infrastructure-release/spec.md` as scenarios._ Rejected — recipes are operator commands, not testable behavior contracts. Scenarios describe inputs and expected outcomes; recipes describe shell invocations. Mixing them pollutes the spec.

### 5. Cross-reference update in `openspec/AGENTS.md`

**Choice:** Add a one-line entry to `openspec/AGENTS.md` §Active spec cross-references naming `infrastructure-release` with a one-sentence ownership summary. No other change to `openspec/AGENTS.md`.

**Rationale:** The "Active spec cross-references" list is the project-standard mechanism for surfacing new specs. Without the entry, agents following `openspec/AGENTS.md` would not learn that `infrastructure-release` exists as a main spec until they ran `openspec list --specs`. The entry costs one line and matches the existing pattern (`core-worktree-status`, `cli-status`).

**Alternatives considered:**

- _Skip the cross-reference update; rely on `openspec list --specs` for discovery._ Rejected — the cross-reference list exists precisely to make newly-promoted specs discoverable without a CLI round-trip.
- _Restructure the canonical ownership table in `openspec/config.yaml`._ Rejected — out of scope. The table in `config.yaml` already documents that `infrastructure-release` would be a single capability, not a category.

## Risks / Trade-offs

**[R1] Sync wording diverges from the archived delta.** A reviewer might find subtle wording drift between the archived delta and the new main spec, even though the new main spec preserves every archived intent. Mitigated by sourcing each restated requirement from the archived delta verbatim, except for grammar fixes that preserve intent. The cascade-derived requirements (#11–#20) are entirely new; there is no archived wording to compare them against. The change's `tasks.md` includes a verification step that grep-compares each restated requirement header against the archived delta header.

**[R2] Permissive divergence ("MAY diverge") reads as weaker than the archived prescriptive rules.** The archived `infrastructure-release` delta is fully prescriptive (TLS SHALL, non-forced SHALL, govulncheck SHALL). The new main spec retains all ten archived requirements as prescriptive `SHALL` rules; only the cascade-derived "per-surface architecture" requirement is permissive. Mitigated by the design rationale documenting the cascade as the evidence base; an agent reviewing the spec sees one permissive rule surrounded by nine prescriptive rules.

**[R3] Operator recipes in root `AGENTS.md` look like invariants.** Mixing bracketed invariant rules with cosign `verify-blob` commands in the same file could mislead an agent into treating the cosign commands as enforceable rules. Mitigated by the section heading ("Release process") being visually distinct from "Release / CI invariants", and by the explicit invariant/process split in the section intros.

**[R4] Bracketed keys introduce a new convention surface in root `AGENTS.md`.** Agents that pattern-match `[code-conv]` for `openspec/AGENTS.md` may not immediately pattern-match `[release-*]` in root `AGENTS.md`. Mitigated by the section heading naming the convention and by `openspec/AGENTS.md` §Active spec cross-references pointing at `infrastructure-release`, which carries the canonical key list.

**[R5] `openspec/AGENTS.md` cross-reference update creates a third edited file in a docs-only change.** The cross-reference is a single-line addition; the edit is mechanical and verifiable. Mitigated by including the verification step in `tasks.md`.

## Migration Plan

For source-only pre-1.0 changes, an explicit "no migration steps" line applies: this change does not deploy runtime artifacts, does not introduce or remove APIs, does not change CLI surface. No data migration; no rollback plan.

The sequence is: implement docs/spec edits → run `openspec validate --change consolidate-release-invariants` → run `openspec validate --specs` (after sync) → run `mise run verify` to confirm prettier and golangci-lint still pass on the modified markdown → archive the change → run `openspec sync-specs` to write the main spec into `openspec/specs/infrastructure-release/spec.md`.

## Open Questions

None. All four questions from the planning conversation (bracketed-key prefix, divergence rule permission, recipe placement, cross-reference update) are resolved by Decisions 2–5 above.
