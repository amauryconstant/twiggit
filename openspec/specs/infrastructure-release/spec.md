# Capability: Release and Distribution

## Purpose

CI/CD pipeline, release artifact distribution (GitLab primary, GitHub
discoverability), and contributor onboarding. This spec
is the canonical owner of release-process behavior; per-concern
details are cross-referenced.

Tool ownership (which file pins which tool) and the composition of
`mise run verify` are owned by `infrastructure-toolchain`; this spec
references those contracts instead of restating them.

## Requirements

### Requirement: CI pipeline

The CI pipeline SHALL build the binary, run lint, and run the full
test matrix on every push and pull request. Coverage enforcement and
artifact upload run on tagged releases.

#### Scenario: Push to main

- **WHEN** a commit is pushed to `main`
- **THEN** CI SHALL run `mise run lint:gated` and `mise run test`
- **AND** SHALL upload a debug binary for manual verification

#### Scenario: Tagged release

- **WHEN** a tag matching `v*` is pushed
- **THEN** the GitLab CI `release` job SHALL run GoReleaser against
  `.goreleaser.yml`
- **AND** SHALL publish artifacts to GitLab releases
- **AND** the GitHub Actions release workflow SHALL publish the
  GitHub release per the `GitHub Actions release workflow`
  requirement

### Requirement: GitLab primary + GitHub first-class distribution

GitLab releases SHALL host the canonical release artifacts (binary
tarballs, checksums, SBOMs) and SHALL be the first publisher per
tag. GitHub releases SHALL also host the full release artifacts
(tarballs, checksums, SBOMs) as a parallel first-class target. The
two providers SHALL publish the same artifact set for the same tag;
neither SHALL be a discoverability stub for the other.

#### Scenario: Primary artifact on GitLab

- **WHEN** a release is published
- **THEN** the primary tarball, checksum, and SBOM SHALL be on
  GitLab releases

#### Scenario: GitHub first-class target

- **WHEN** a user visits the GitHub repo
- **THEN** the GitHub release SHALL host the same tarballs, checksums,
  and SBOMs as the GitLab release for the tag
- **AND** the GitHub release SHALL NOT be a discoverability stub
  pointing only at GitLab

### Requirement: GitHub Actions release workflow

A `.github/workflows/release.yml` workflow SHALL publish a GitHub
release with full artifacts when a `v*` tag is pushed. The workflow
SHALL set `permissions: contents: write` and
`permissions: id-token: write` (cosign keyless), run the job inside
the `goreleaser/goreleaser:v2.18.2` Docker image as its container
so that `goreleaser`, `syft`, `cosign`, and `gh` are on `$PATH`
without per-step installation, check out the repo with full history,
validate the tag format (`vX.Y.Z`), preflight any existing GitHub
release via `gh release view`, run
`goreleaser release --clean -f .goreleaser.github.yml`, and sign
every `dist/*_sbom.spdx.json` artifact with
`cosign sign-blob --yes --output-signature <file>.sig`.

#### Scenario: Tag push triggers the workflow

- **WHEN** a tag matching `v[0-9]+\.[0-9]+\.[0-9]+$` is pushed
- **THEN** the workflow SHALL run on the tag commit
- **AND** SHALL publish tarballs, checksums, and SBOMs to the GitHub
  release for that tag

#### Scenario: Replay of an existing tag fails cleanly

- **WHEN** the workflow runs for a tag whose GitHub release already
  exists
- **THEN** the preflight step SHALL exit non-zero
- **AND** goreleaser SHALL NOT be invoked
- **AND** no GitHub release artifact SHALL be overwritten

#### Scenario: SBOM signatures on GitHub

- **WHEN** goreleaser finishes on GitHub Actions
- **THEN** every `dist/*_sbom.spdx.json` SHALL be signed with
  `cosign sign-blob --yes --output-signature <file>.sig`
- **AND** `cosign verify-blob` SHALL succeed for each signed SBOM

#### Scenario: Tag format guard

- **WHEN** a tag not matching `vX.Y.Z` is pushed
- **THEN** the workflow SHALL exit non-zero before goreleaser runs
- **AND** no GitHub release SHALL be created

#### Scenario: syft present for SBOM cataloging

- **WHEN** the workflow runs the goreleaser step
- **THEN** `syft` SHALL be installed at the pinned version on `$PATH`
- **AND** goreleaser SHALL successfully catalog archives into
  `dist/*_sbom.spdx.json` artifacts

### Requirement: Release validation hooks

Two `mise` tasks SHALL gate releases:
`mise run release:check` (full release prerequisites: environment,
clean tree, CHANGELOG, GoReleaser config, version bump calculation)
and `mise run release:dry-run` (test GoReleaser config without publishing).

#### Scenario: Check before tag

- **WHEN** developer runs `mise run release:check`
- **THEN** task SHALL run environment, CHANGELOG, GoReleaser, and
  version bump validations
- **AND** SHALL fail if the working tree is dirty (CHANGELOG.md excluded)
- **AND** SHALL fail if not on `main` branch
- **AND** SHALL preview the tag that would be created

#### Scenario: Dry run before publish

- **WHEN** developer runs `mise run release:dry-run`
- **THEN** GoReleaser SHALL run with `--skip-publish`
- **AND** no artifacts SHALL be uploaded

### Requirement: Contributor onboarding

`CONTRIBUTING.md` SHALL cover: `mise install` for toolchain setup,
`pre-commit install` for hooks, `mise run verify` for the full
validation suite, and the project's commit conventions. The
toolchain provisioned by `mise install` SHALL include (at minimum):
`go`, `golangci-lint`, `goreleaser`, `ginkgo`, `gopls`,
`govulncheck`, `gocover-cobertura`, `gocovmerge`, and `pre-commit`.
The exact mapping of which file pins which tool is the concern of
`infrastructure-toolchain`.

#### Scenario: New contributor setup

- **WHEN** a new contributor reads `CONTRIBUTING.md`
- **THEN** they SHALL find: toolchain install via `mise install`,
  pre-commit hook install via `pre-commit install`, and the
  `mise run verify` validation command
- **AND** the documented toolchain SHALL match the set enforced by
  `infrastructure-toolchain`

### Requirement: cmd tested exclusively via E2E

The `cmd/` package SHALL be tested exclusively via E2E
(Ginkgo/Gomega); unit tests in `cmd/` SHALL NOT be added.

#### Scenario: No cmd/ unit tests

- **WHEN** a developer inspects `cmd/`
- **THEN** only E2E tests (in `test/e2e/`) SHALL cover cmd behavior
- **AND** no `_test.go` files SHALL exist alongside the cmd source
  files (other than `cmd/error_formatter_test.go` and
  `cmd/util_test.go`, which test pure helpers)

See `testing-e2e`.

### Requirement: Unit-test conventions

Unit tests across the codebase SHALL follow:

- Table-driven tests with descriptive names
- `t.Run()` sub-tests for parameterized cases
- `t.Cleanup()` for resource teardown (not `defer t.Cleanup`)
- No `testify/suite` (use plain `testing.T`)
- No `init()`-based setup

See `CONTRIBUTING.md` for the full convention guide.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Golden-file testing

Golden-file tests SHALL use a `CompareGolden(actual, name)` helper
that compares `actual` against
`test/golden/<name>.golden`. The `UPDATE_GOLDEN=1` env var SHALL
rewrite the golden file in place.

#### Scenario: Match golden

- **WHEN** `CompareGolden` is called and the actual matches the file
- **THEN** the test SHALL pass

#### Scenario: Mismatch

- **WHEN** `CompareGolden` is called and the actual differs
- **THEN** the test SHALL fail with a unified diff

#### Scenario: Update golden

- **WHEN** `UPDATE_GOLDEN=1` is set
- **THEN** `CompareGolden` SHALL write the actual to the golden file
- **AND** the test SHALL pass

See `testing-golden-file-testing`.

### Requirement: E2E suite

End-to-end tests SHALL use Ginkgo/Gomega, run against the built
binary in a temp directory, and cover all user-visible commands plus
edge cases (corrupted repo, bare repo, submodule, detached HEAD).
See `testing-e2e` and `testing-edge-case-fixtures`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: TLS-enabled DinD service

The `build-ci-image` job SHALL enable TLS on the Docker-in-Docker
service by setting `DOCKER_TLS_CERTDIR: "/certs"` in job `variables:`.
The DinD service SHALL present a TLS certificate on port 2376 under
the `dockerd-entrypoint.sh` contract, and the client SHALL connect
via `tcp://docker:2376` with the client cert set mounted at
`/certs/client`.

#### Scenario: TLS handshake succeeds between client and service

- **WHEN** `build-ci-image` runs
- **THEN** the CI SHALL NOT expose `DOCKER_HOST: tcp://docker:2375`
- **AND** `DOCKER_TLS_CERTDIR` SHALL be a non-empty path
- **AND** the `docker build` invocation SHALL complete via the TLS
  socket on port 2376
- **AND** a co-tenant runner SHALL NOT be able to drive `docker build`
  or `docker push` against the registry without the DinD cert bundle

### Requirement: Mirror push uses --force-with-lease

The `mirror-to-github` job SHALL push the GitHub mirror using
`git push --force-with-lease` so a divergent GitHub ref can be reconciled
without overwriting concurrent updates. The push SHALL be
`git push github HEAD:$CI_COMMIT_BRANCH --tags --force-with-lease` after
explicitly fetching `github/$CI_COMMIT_BRANCH` so the lease tracks the
remote state. GitLab is the canonical reference; GitHub follows.

#### Scenario: Main-branch sync with divergent GitHub ref

- **WHEN** a commit is pushed to `main` and the mirror job runs
- **AND** `github/main` has diverged from `origin/main` (e.g. an earlier
  mirror attempt failed non-fast-forward)
- **THEN** the job SHALL fetch `github/$CI_COMMIT_BRANCH` before pushing
- **AND** the push SHALL include `--force-with-lease`
- **AND** the push SHALL succeed and overwrite the divergent GitHub ref

#### Scenario: Main-branch sync when GitHub ref has moved concurrently

- **WHEN** a commit is pushed to `main` and the mirror job runs
- **AND** `github/main` has been updated since the local fetch
- **THEN** the push SHALL fail with the lease-mismatch error
- **AND** the job SHALL NOT overwrite the concurrent update

#### Scenario: Tag push

- **WHEN** a `v*` tag is mirrored via `--tags`
- **AND** a tag with the same name does NOT exist on GitHub
- **THEN** the tag SHALL be pushed without `--force`
- **AND** an existing tag at the same name SHALL cause the push to fail
  rather than overwrite the GitHub tag

### Requirement: Release publish is non-overwriting on tag reuse

Each release pipeline SHALL verify the tag has no existing release on
its target provider before invoking goreleaser. The GitLab CI
`release` job SHALL preflight the GitLab Releases API; the GitHub
Actions `release.yml` workflow SHALL preflight via `gh release view`.
Both `.goreleaser.yml` and `.goreleaser.github.yml` SHALL set
`release.replace_existing_artifacts: false` so goreleaser never
silently overwrites artifacts when invoked.

#### Scenario: First release of a tag on GitLab

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
- **AND** goreleaser SHALL publish artifacts to the GitHub release
- **AND** the artifacts SHALL be downloadable from
  `github.com/amauryconstant/twiggit/releases/tag/<tag>`

#### Scenario: Re-tag of an existing release on either provider

- **WHEN** CI runs release for a tag that already has a published
  release on either provider
- **THEN** that provider's preflight SHALL fail
- **AND** goreleaser SHALL NOT be invoked for that provider
- **AND** no artifact SHALL be silently overwritten on that provider

### Requirement: CI supply-chain signals

The CI pipeline SHALL run `govulncheck` on twiggit's source code at the
`lint` stage and SHALL sign release SBOMs.

#### Scenario: govulncheck on source

- **WHEN** `lint` runs
- **THEN** a `govulncheck ./...` step SHALL run after `go mod tidy`
- **AND** unfixed vulnerabilities in called code SHALL exit the step
  non-zero
- **AND** the non-zero exit SHALL fail the `lint` job

#### Scenario: cosign SBOM signature

- **WHEN** `release` finishes `goreleaser release --clean`
- **THEN** a `cosign sign-blob` step SHALL sign each
  `*_sbom.spdx.json` artifact with a detached `.sig` sidecar
- **AND** `cosign verify-blob` SHALL succeed for consumers verifying
  the SBOM via the Rekor transparency log

### Requirement: Pinned CI tooling images

CI jobs that require a goreleaser container SHALL pin the image to a
specific tag. The pinned version SHALL match the goreleaser entry in
`.mise/config.toml` `[tools]`. The `build-ci-image` trigger SHALL cover
`.mise/config.toml` and `.goreleaser.yml` so a goreleaser-version bump
rebuilds the image before downstream jobs run.

#### Scenario: Goreleaser dry-run job

- **WHEN** `goreleaser-dry-run` runs
- **THEN** the job image SHALL be a pinned tag, not the moving
  `latest`
- **AND** a goreleaser minor-version drift SHALL NOT alter CI behavior
  silently

#### Scenario: Goreleaser release job

- **WHEN** `release` runs
- **THEN** the job image SHALL be the same pinned tag

### Requirement: Module-graph freshness gate

The CI pipeline SHALL verify that the module graph declared by
`go.mod`/`go.sum` matches the source under review. The check SHALL run
in the `lint` job before `golangci-lint run`.

#### Scenario: Lint gate

- **WHEN** `lint` runs
- **THEN** it SHALL first execute `go mod tidy && git diff --exit-code`
- **AND** an unstaged `go.mod`/`go.sum` drift SHALL fail the lint job

### Requirement: Pipeline defaults for jobs

The CI `.gitlab-ci.yml` `default:` block SHALL declare the retry policy
and interruptibility expectation that apply to every job unless a job
overrides them.

#### Scenario: Default retry applies

- **WHEN** a job does not override `retry:`
- **THEN** runner-system or stuck-or-timeout failures SHALL be retried
  up to the `default:` `retry:` count
- **AND** test failures SHALL NOT be retried by the default policy

#### Scenario: Default interruptibility

- **WHEN** a job does not override `interruptible:`
- **THEN** a new pipeline on the same ref SHALL be able to cancel the
  in-flight job (interruptible: true) unless the job explicitly opts
  out

### Requirement: Release job is non-interruptible

The `release` job SHALL set `interruptible: false` so a concurrent
pipeline on the same ref cannot abort a publish mid-flight.

#### Scenario: Release in flight

- **WHEN** `release` is running and a new pipeline starts on the same
  ref
- **THEN** the new pipeline SHALL NOT cancel `release`
- **AND** `release` SHALL complete naturally or fail on its own
  merits

### Requirement: CI image triggers include release tooling

The `build-ci-image` job SHALL trigger on changes to
`.goreleaser.yml` in addition to its existing triggers, so a
goreleaser-version bump rebuilds the CI image automatically.

#### Scenario: Goreleaser config edit

- **WHEN** a commit modifies `.goreleaser.yml`
- **THEN** `build-ci-image` SHALL queue before downstream jobs
- **AND** the rebuilt image SHALL carry the goreleaser version
  expected by `.gitlab-ci.yml`

### Requirement: Goreleaser dry-run on release-affecting merge requests

The `goreleaser-dry-run` job SHALL run on merge requests that touch
release-affecting paths (`cmd/`, `internal/`, `.goreleaser.yml`,
`go.mod`, `go.sum`). The dry-run SHALL produce snapshot artifacts
under a short retention window to avoid runner disk pressure.

#### Scenario: Release-affecting trigger

- **WHEN** a merge request touches a release-affecting path
- **THEN** `goreleaser-dry-run` SHALL run
- **AND** the result SHALL gate the MR

#### Scenario: Docs-only trigger

- **WHEN** a merge request touches only docs or non-release paths
- **THEN** `goreleaser-dry-run` SHALL NOT run

#### Scenario: Artifact retention

- **WHEN** `goreleaser-dry-run` publishes snapshot artifacts
- **THEN** the artifacts SHALL expire within one day
- **AND** older snapshots SHALL NOT accumulate
