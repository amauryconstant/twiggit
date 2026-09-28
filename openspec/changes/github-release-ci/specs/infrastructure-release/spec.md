# Spec Delta

## RENAMED Requirements

- FROM: `### Requirement: GitLab primary distribution`
- TO: `### Requirement: GitLab primary + GitHub first-class distribution`

## ADDED Requirements

### Requirement: GitHub Actions release workflow

A `.github/workflows/release.yml` workflow SHALL publish a GitHub
Release with full artifacts when a `v*` tag is pushed. The workflow
SHALL set `permissions: contents: write` and
`permissions: id-token: write` (cosign keyless), check out the repo
with full history, install the pinned toolchain and the pinned cosign
binary, validate the tag format (`vX.Y.Z`), preflight any existing
GitHub release via `gh release view`, run `goreleaser release --clean`,
and sign every `dist/*_sbom.spdx.json` artifact with `cosign sign`.

#### Scenario: Tag push triggers the workflow

- **WHEN** a tag matching `v[0-9]+\.[0-9]+\.[0-9]+$` is pushed
- **THEN** the workflow SHALL run on the tag commit
- **AND** SHALL publish tarballs, checksums, and SBOMs to the GitHub
  Release for that tag

#### Scenario: Replay of an existing tag fails cleanly

- **WHEN** the workflow runs for a tag whose GitHub Release already
  exists
- **THEN** the preflight step SHALL exit non-zero
- **AND** goreleaser SHALL NOT be invoked
- **AND** no GitHub Release artifact SHALL be overwritten

#### Scenario: SBOM signatures on GitHub

- **WHEN** goreleaser finishes on GitHub Actions
- **THEN** every `dist/*_sbom.spdx.json` SHALL be signed with
  `cosign sign --yes`
- **AND** `cosign verify-blob` SHALL succeed for each signed SBOM

#### Scenario: Tag format guard

- **WHEN** a tag not matching `vX.Y.Z` is pushed
- **THEN** the workflow SHALL exit non-zero before goreleaser runs
- **AND** no GitHub Release SHALL be created

## MODIFIED Requirements

### Requirement: GitLab primary + GitHub first-class distribution

GitLab releases SHALL host the canonical release artifacts (binary
tarballs, checksums, packages) and SHALL be the first publisher per
tag. GitHub releases SHALL also host the full release artifacts
(tarballs, checksums, SBOMs) as a parallel first-class target. The
two providers SHALL publish the same artifact set for the same tag;
neither SHALL be a discoverability stub for the other.

#### Scenario: Primary artifact on GitLab

- **WHEN** a release is published
- **THEN** the primary tarball, checksum, and SBOM SHALL be on
  GitLab releases

#### Scenario: GitHub discoverability

- **WHEN** a user visits the GitHub repo
- **THEN** the GitHub Release SHALL host the same tarballs, checksums,
  and SBOMs as the GitLab release for the tag
- **AND** the GitHub Release SHALL NOT be a discoverability stub
  pointing only at GitLab

### Requirement: Release publish is non-overwriting on tag reuse

The release pipeline SHALL verify the tag has no existing release on
the target provider before invoking goreleaser on that provider. The
GitLab `release` job SHALL preflight the GitLab Releases API; the
GitHub Actions `release.yml` workflow SHALL preflight via
`gh release view`. `.goreleaser.yml` SHALL set
`release.replace_existing_artifacts: false` so goreleaser never
silently overwrites artifacts when invoked.

#### Scenario: First release of a tag

- **WHEN** CI runs the GitLab `release` job for a previously-unreleased
  `v*` tag
- **THEN** the GitLab Releases API preflight SHALL return non-200 for
  the tag
- **AND** goreleaser SHALL publish artifacts to the GitLab release URL
- **AND** the artifacts SHALL be downloadable from
  `gitlab.com/amoconst/twiggit/-/releases/v<tag>`

#### Scenario: First release of a tag on GitHub

- **WHEN** the GitHub Actions `release.yml` workflow runs for a
  previously-unreleased `v*` tag
- **THEN** `gh release view` SHALL report no existing release for the
  tag
- **AND** goreleaser SHALL publish artifacts to the GitHub Release
- **AND** the artifacts SHALL be downloadable from
  `github.com/amauryconstant/twiggit/releases/tag/<tag>`

#### Scenario: Re-tag of an existing release

- **WHEN** CI runs release for a tag that already has a published
  release on either provider
- **THEN** that provider's preflight SHALL fail
- **AND** goreleaser SHALL NOT be invoked for that provider
- **AND** no artifact SHALL be silently overwritten on that provider
