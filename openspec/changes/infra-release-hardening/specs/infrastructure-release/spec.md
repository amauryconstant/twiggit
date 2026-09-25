# Spec Delta

## ADDED Requirements

### Requirement: TLS-enabled DinD service

The `build-ci-image` job SHALL use the TLS-enabled Docker-in-Docker
service image rather than the plain TCP service. The DinD service
SHALL present a TLS certificate under a non-empty
`DOCKER_TLS_CERTDIR`, and the client SHALL connect via `tcp://docker:2376`
with the certificate bundle mounted into the client container.

#### Scenario: TLS handshake succeeds between client and service

- **WHEN** `build-ci-image` runs
- **THEN** the CI SHALL NOT expose `DOCKER_HOST: tcp://docker:2375`
- **AND** the `docker build` invocation SHALL complete via the TLS
  socket
- **AND** a co-tenant runner SHALL NOT be able to drive `docker build`
  or `docker push` against the registry without the DinD cert bundle

### Requirement: Mirror push is non-forced

The `mirror-to-github` job SHALL push the main-branch mirror without
`--force`. Tag mirroring SHALL be gated by the same non-forced push
form, scoped to the explicit tag trigger.

#### Scenario: Main-branch sync

- **WHEN** a commit is pushed to `main` and the mirror job runs
- **THEN** the push SHALL NOT include `--force`
- **AND** the push SHALL succeed only if the remote ref fast-forwards
- **AND** a non-fast-forward local branch SHALL cause the job to fail
  instead of overwriting the GitHub ref

#### Scenario: Tag mirror

- **WHEN** a `v*` tag is pushed
- **THEN** the tag SHALL be pushed without `--force`
- **AND** an existing tag at the same name SHALL cause the job to fail
  rather than overwrite the GitHub tag

### Requirement: Release publish is non-overwriting on tag reuse

The `release` job SHALL invoke goreleaser in `release.mode: append`
configuration. The goreleaser invocation SHALL pass
`--fail-if-tag-exists` so an accidental re-tag hard-fails the release
job instead of silently replacing the artifacts of an existing
release.

#### Scenario: First release of a tag

- **WHEN** CI runs `release` for a previously-unreleased `v*` tag
- **THEN** goreleaser SHALL publish artifacts to the GitLab release
  URL
- **AND** the artifacts SHALL be downloadable from
  `gitlab.com/amoconst/twiggit/-/releases/v<tag>`

#### Scenario: Re-tag of an existing release

- **WHEN** CI runs `release` for a tag that already has a published
  release
- **THEN** goreleaser SHALL exit non-zero
- **AND** CI SHALL fail the `release` job
- **AND** no artifact SHALL be silently overwritten

### Requirement: CI supply-chain scans

The CI pipeline SHALL scan published images and release artifacts for
vulnerabilities and SHALL sign release SBOMs.

#### Scenario: Trivy scan on the CI image

- **WHEN** `build-ci-image` finishes `docker push`
- **THEN** a `trivy image` step SHALL scan the published CI image
- **AND** `CRITICAL` and `HIGH` findings SHALL cause the downstream
  `validate`-stage jobs to gate

#### Scenario: cosign SBOM signature

- **WHEN** `release` finishes `goreleaser release --clean`
- **THEN** a `cosign sign` step SHALL sign each
  `*_sbom.spdx.json` artifact
- **AND** `cosign verify-blob` SHALL succeed for consumers retrieving
  the SBOM

### Requirement: Pinned CI tooling images

CI jobs that require a goreleaser container SHALL pin the image to a
specific tag. The pinned version SHALL match the goreleaser entry in
`.mise/config.toml` `[tools]`.

#### Scenario: Goreleaser dry-run job

- **WHEN** `goreleaser-dry-run` runs
- **THEN** the job image SHALL be a pinned tag, not the moving
  `latest`
- **AND** a goreleaser minor-version drift SHALL NOT alter CI behavior
  silently

#### Scenario: Goreleaser release job

- **WHEN** `release` runs
- **THEN** the job image SHALL be the same pinned tag
- **AND** the version SHALL be discoverable via `goreleaser --version`
  in `before_script`

### Requirement: Go-version matrix in CI

The CI pipeline SHALL execute the test suite under a matrix of Go
versions covering the toolchain set: the directive version declared in
`go.mod` plus the prior supported minor's latest patch. The matrix
SHALL set `fail-fast: false` so a single matrix entry does not cancel
the others.

#### Scenario: Matrix entries

- **WHEN** CI runs on a merge request
- **THEN** the test job SHALL include entries for the `go.mod`
  directive version and the previous minor's latest patch
- **AND** each entry SHALL run `mise run test` (or equivalent) and
  report independently

#### Scenario: Failure isolation

- **WHEN** one matrix entry fails
- **THEN** the other entries SHALL continue to completion
- **AND** the job SHALL report aggregate success only when every entry
  succeeds

### Requirement: Standalone race-detector job

The CI pipeline SHALL execute a `go test -race` job decoupled from
coverage. Race failures SHALL surface at that job's log instead of
being folded into coverage noise.

#### Scenario: Race failure visibility

- **WHEN** a race detector finding is introduced
- **THEN** the standalone race job SHALL fail
- **AND** the failure message SHALL point to the race-detected data
  race, distinct from coverage-shaping noise

### Requirement: Module-graph freshness gate

The CI pipeline SHALL verify that the module graph declared by
`go.mod`/`go.sum` matches the source under review. The check SHALL run
before lint and before the test job in the `validate` stage.

#### Scenario: Lint gate

- **WHEN** `lint` runs
- **THEN** it SHALL first execute `go mod tidy && git diff --exit-code`
- **AND** an unstaged `go.mod`/`go.sum` drift SHALL fail the lint job

#### Scenario: Test gate

- **WHEN** `test` runs
- **THEN** its `before_script` SHALL run `go mod tidy -diff`
- **AND** a missing tidy SHALL fail the test job before any test
  executes

### Requirement: Pipeline defaults for jobs

The CI `.gitlab-ci.yml` `default:` block SHALL declare the runner
selection, retry policy, and interruptibility expectation that apply
to every job unless a job overrides them.

#### Scenario: Default tags and retry apply

- **WHEN** a job does not override `tags:`
- **THEN** the job SHALL run on the runner tag declared at `default:`
- **AND** transient errors SHALL be retried up to the `default:`
  `retry:` count

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

### Requirement: Goreleaser dry-run runs on every merge request

The `goreleaser-dry-run` job SHALL run on every merge request,
independent of which paths the MR touches. The dry-run SHALL produce
snapshot artifacts under a short retention window to avoid runner
disk pressure.

#### Scenario: Path-agnostic trigger

- **WHEN** any merge request is opened
- **THEN** `goreleaser-dry-run` SHALL run
- **AND** the result SHALL gate the MR

#### Scenario: Artifact retention

- **WHEN** `goreleaser-dry-run` publishes snapshot artifacts
- **THEN** the artifacts SHALL expire within one day
- **AND** older snapshots SHALL NOT accumulate

### Requirement: Fast race-detector subset

The CI pipeline SHALL execute a fast `-race` subset covering the
`internal/` and `cmd/` trees before the slower full `mise run
ci:coverage` job. The full `-race` run remains required for tagged
releases.

#### Scenario: Default branch race subset

- **WHEN** CI runs on a merge request or main-branch push
- **THEN** a `test:race:subset` job SHALL run `go test -race
  ./internal/... ./cmd/...`
- **AND** the subset SHALL complete on the order of minutes

#### Scenario: Tagged release race coverage

- **WHEN** CI runs on a `v*` tag
- **THEN** `mise run ci:coverage` SHALL run and include the `-race`
  flag for `./...`
- **AND** the tagged-release job SHALL fail if `-race` reports a
  failure even when the subset passed

### Requirement: Mirror job's main-branch push is non-forced

The `mirror-to-github` job's main-branch sync SHALL NOT use
`--force`. Tag mirroring SHALL use `--follow-tags` (no `--force`)
under the existing `when: never` exclusion for the tag trigger.

#### Scenario: Push protection

- **WHEN** the mirror job runs for a main-branch sync
- **THEN** the push SHALL be plain `git push github HEAD:$CI_COMMIT_BRANCH --tags`
- **AND** a non-fast-forward remote SHALL reject the push
- **AND** the job SHALL fail instead of overwriting GitHub history
