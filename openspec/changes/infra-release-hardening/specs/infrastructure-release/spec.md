# Spec Delta

## ADDED Requirements

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

### Requirement: Mirror push is non-forced

The `mirror-to-github` job SHALL push the GitHub mirror without
`--force`. The push SHALL be plain `git push github HEAD:$CI_COMMIT_BRANCH --tags`
and SHALL fail rather than overwrite remote refs.

#### Scenario: Main-branch sync

- **WHEN** a commit is pushed to `main` and the mirror job runs
- **THEN** the push SHALL NOT include `--force`
- **AND** the push SHALL be `git push github HEAD:$CI_COMMIT_BRANCH --tags`
- **AND** the push SHALL succeed only if the remote ref fast-forwards
- **AND** a non-fast-forward local branch SHALL cause the job to fail
  instead of overwriting the GitHub ref

#### Scenario: Tag push

- **WHEN** a `v*` tag is mirrored via `--tags`
- **THEN** the tag SHALL be pushed without `--force`
- **AND** an existing tag at the same name SHALL cause the push to fail
  rather than overwrite the GitHub tag

### Requirement: Release publish is non-overwriting on tag reuse

The `release` job SHALL verify the tag has no existing GitLab release
before invoking goreleaser. `.goreleaser.yml` SHALL set
`release.replace_existing_artifacts: false` so goreleaser never silently
overwrites artifacts when invoked.

#### Scenario: First release of a tag

- **WHEN** CI runs `release` for a previously-unreleased `v*` tag
- **THEN** the GitLab Releases API preflight SHALL return non-200 for
  the tag
- **AND** goreleaser SHALL publish artifacts to the GitLab release URL
- **AND** the artifacts SHALL be downloadable from
  `gitlab.com/amoconst/twiggit/-/releases/v<tag>`

#### Scenario: Re-tag of an existing release

- **WHEN** CI runs `release` for a tag that already has a published
  release
- **THEN** the GitLab Releases API preflight SHALL return 200 for the
  tag
- **AND** CI SHALL fail the `release` job before goreleaser is invoked
- **AND** no artifact SHALL be silently overwritten

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
- **THEN** a `cosign sign` step SHALL sign each
  `*_sbom.spdx.json` artifact
- **AND** `cosign verify-blob` SHALL succeed for consumers retrieving
  the SBOM

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

