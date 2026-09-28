# Proposal

## Why

GitLab CI owns build/test/lint for twiggit. Releases currently publish
to GitLab Releases plus a discoverability stub on GitHub Releases
(`.goreleaser/hooks/create-github-release.sh`), so GitHub users hit a
release page with no artifacts. Promote GitHub Releases to a first-class
publish target by running goreleaser on tag push via GitHub Actions,
mirroring the existing GitLab release job's preflight and cosign SBOM
signing so both providers publish identical, signed artifact sets.

## What Changes

- Add `.github/workflows/release.yml` triggered on `v*` tag push. Runs
  goreleaser with `release.github` set, prefights an existing GitHub
  release via `gh release view`, signs SBOMs with cosign keyless using
  GitHub OIDC, and mirrors the GitLab job's tag-format guard.
- Update `.goreleaser.yml`: add `release.github` block (`owner:
  amauryconstant`, `name: twiggit`); remove the
  `create-github-release.sh` reference from `before.hooks`.
- Delete `.goreleaser/hooks/create-github-release.sh`.
- Update `openspec/specs/infrastructure-release/spec.md`:
  - Replace "GitLab primary distribution" and "GitHub discoverability"
    requirements with a unified primary/first-class-target requirement.
  - Add "GitHub Actions release workflow" requirement covering the new
    workflow file, tag trigger, goreleaser invocation, preflight, and
    cosign keyless signing via `id-token: write`.
  - Extend "Release publish is non-overwriting on tag reuse" to cover
    the GitHub `gh release view` preflight.

GitLab CI release job is unchanged. GitLab stays canonical; GitHub
becomes a parallel first-class target.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `infrastructure-release`: GitHub distribution model flips from
  discoverability stub to first-class publish target; a GitHub Actions
  release workflow gains the same preflight and cosign guarantees as
  the GitLab release job.

## Impact

- `.github/workflows/release.yml` — new file.
- `.goreleaser.yml` — add `release.github` block; drop one
  `before.hooks` entry.
- `.goreleaser/hooks/create-github-release.sh` — deleted.
- `openspec/specs/infrastructure-release/spec.md` — three requirements
  revised (two edited, one extended).
- No Go source, no domain/service/cmd changes, no Homebrew tap
  changes. `homebrew_casks.repository.token` already reads
  `GITHUB_TOKEN`, which GitHub Actions provides automatically.
- No new secrets required on GitHub Actions: `GITHUB_TOKEN` is
  automatic, cosign uses keyless via `id-token: write`.
