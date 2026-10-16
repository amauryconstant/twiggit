# Tasks

## 1. Spec — create `openspec/specs/infrastructure-release/spec.md`

- [ ] 1.1 Run `openspec sync-specs --change consolidate-release-invariants` to write the delta's twenty ADDED Requirements into the live main spec at `openspec/specs/infrastructure-release/spec.md`. Verify by `openspec validate --specs` exiting 0 and `ls openspec/specs/infrastructure-release/spec.md` existing.
- [ ] 1.2 Confirm every restated requirement header from the archived delta (`openspec/changes/archive/2026-09-28-infra-release-hardening/specs/infrastructure-release/spec.md`) is preserved as-is in the new main spec. Verify by running `grep -E '^### Requirement:' openspec/changes/archive/2026-09-28-infra-release-hardening/specs/infrastructure-release/spec.md | sort > /tmp/archived.txt && grep -E '^### Requirement:' openspec/specs/infrastructure-release/spec.md | sort > /tmp/main.txt && diff /tmp/archived.txt /tmp/main.txt` and inspecting the diff (restated headers must be a superset; the cascade additions appear only in the main spec).
- [ ] 1.3 Confirm the new main spec carries a `## Purpose` block with `>= 50` characters. Verify by `openspec validate --strict --specs` exiting 0.

## 2. Documentation — root `AGENTS.md`

- [ ] 2.1 Append the **Release / CI invariants** section to root `AGENTS.md` (between the existing CLI Surface section and Conventions section, before the Subdirectory Guides table). Include all twenty bracketed `[release-*]` keys as a flat reference, each with a one-line rule and a parenthetical `(infrastructure-release)` cross-reference. Verify by `grep -c '^\*\*\[' AGENTS.md` returning `>= 20` and `grep -c '(infrastructure-release)' AGENTS.md` returning `>= 20`.
- [ ] 2.2 Append the **Release process** section to root `AGENTS.md` (immediately after the **Release / CI invariants** section). Include the distribution surface table (GitLab canonical / GitHub mirror / `mirror-to-github`), the `mise run release:*` commands, the cosign `verify-blob` recipe for both GitLab and GitHub OIDC identities, and the per-surface architecture note (GitLab runtime-installer / GitHub bundled-image). Verify by `grep -c '^## Release process' AGENTS.md` returning `1` and `grep -c 'cosign verify-blob' AGENTS.md` returning `>= 2` (one per OIDC identity).
- [ ] 2.3 Remove the `docs/release/AGENTS.md` row from the root `AGENTS.md` subdirectory guide table. Verify by `grep -c 'docs/release/AGENTS.md' AGENTS.md` returning `0`.
- [ ] 2.4 Confirm prettier formatting on the modified root `AGENTS.md`. Verify by `mise run format:md` exiting 0 (idempotent on already-formatted content).

## 3. Documentation — `openspec/AGENTS.md`

- [ ] 3.1 Add a one-line bullet to `openspec/AGENTS.md` §Active spec cross-references naming `infrastructure-release` with a one-sentence ownership summary. Verify by `grep -c 'infrastructure-release' openspec/AGENTS.md` returning `>= 1` and `openspec/AGENTS.md` containing the new bullet under the cross-references heading (not in another section).

## 4. Documentation — delete `docs/release/AGENTS.md` and the `docs/release/` directory

- [ ] 4.1 Delete `docs/release/AGENTS.md`. Verify by `test ! -f docs/release/AGENTS.md` exiting 0.
- [ ] 4.2 Delete the now-empty `docs/release/` directory. Verify by `test ! -d docs/release` exiting 0.

## 5. Validation

- [ ] 5.1 Run `openspec validate --change consolidate-release-invariants` and confirm exit 0. Verify by exit code only.
- [ ] 5.2 Run `mise run lint:md` against the modified markdown and confirm exit 0. Verify by exit code only.
- [ ] 5.3 Run `mise run verify` (full pipeline: format, lint:fix, lint:gated, vuln:check, test, build) and confirm exit 0. Verify by exit code only. No Go test is expected to change; the verify gate is a regression check on the docs change.
- [ ] 5.4 Confirm no `cmd/`, `internal/`, `.gitlab-ci.yml`, `.github/workflows/*`, `Dockerfile.ci`, `Dockerfile.release`, `.goreleaser.yml`, `mise/**`, or `go.mod`/`go.sum` change. Verify by `git diff --name-only` listing only `AGENTS.md`, `openspec/AGENTS.md`, the new `openspec/changes/consolidate-release-invariants/**`, and the new `openspec/specs/infrastructure-release/spec.md`.
