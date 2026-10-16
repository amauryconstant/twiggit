# Spec Delta

## Purpose

Canonical home for twiggit's CI and release pipeline invariants. Owns the goreleaser/cosign/syft toolchain pin contract, the GitLab and GitHub Actions release surface contracts, release-tag reuse protection, module-graph freshness gating, and the per-surface architecture decisions governing how the two CI vendors publish the same release.

## ADDED Requirements

### Requirement: Release job image satisfies go.mod Go directive

The release job image SHALL run a Go version that satisfies `go.mod`'s `go` directive. Goreleaser pins `GOTOOLCHAIN=local` for child `go build` invocations, which disables Go toolchain auto-download; the host Go has to already satisfy `go.mod` or the goreleaser `before:` hook fails. A bump to `go.mod`'s `go` directive SHALL bump the release job image in the same change on every release surface.

#### Scenario: Image satisfies go.mod

- **WHEN** a release job runs
- **THEN** the Go version reported by `go version` SHALL be `>=` the version declared by `go.mod`'s `go` directive

#### Scenario: Image stale relative to go.mod

- **WHEN** `go.mod`'s `go` directive is bumped to a version newer than the release image's pinned Go
- **THEN** the goreleaser `before:` hook SHALL fail with a Go-toolchain error
- **AND** the release job SHALL fail

### Requirement: Release tools installed from upstream prebuilt artifacts

Release-time tools (goreleaser, cosign, syft) SHALL be installed from upstream prebuilt release artifacts SHA-pinned against the upstream `checksums.txt` (or equivalent), extracted to a directory on `$PATH` (typically `/usr/local/bin`). `go install` SHALL NOT be used to install release-time tools because `go install` has no `-o` flag and the project-level `GOPATH` override traps the binary off-`$PATH`.

#### Scenario: Prebuilt binary installed

- **WHEN** the release job's install step runs
- **THEN** each release tool SHALL be located under `/usr/local/bin`
- **AND** each release tool's reported version SHALL match the pinned version

#### Scenario: SHA mismatch

- **WHEN** the downloaded artifact's SHA256 does not match the pinned value
- **THEN** the install step SHALL exit non-zero
- **AND** the release job SHALL fail before goreleaser runs

### Requirement: Alpine scripts use `#!/bin/sh`

Scripts that run inside Alpine-based release containers SHALL start with `#!/bin/sh` and SHALL NOT use `set -o pipefail` (a bash/ksh extension). Busybox `tar` SHALL be invoked with `-f` explicitly. Bash MAY be installed by the script itself, but SHALL NOT be assumed at script-execution time.

#### Scenario: Alpine release container script

- **WHEN** a release job runs on an Alpine-based image and the install script is the first step
- **THEN** the script SHALL execute under busybox `sh`, not bash
- **AND** any `tar` extraction SHALL pass `-f` explicitly

### Requirement: Sparse-base release container installs git before checkout

On GitHub Actions, when the release job runs on a sparse-base container that does not bundle `git`, the install step that provides `git` SHALL run BEFORE `actions/checkout@v5`. `actions/checkout@v5` SHALL be configured with `fetch-tags: true`. GitLab CI is exempt because its runner populates the workspace before `before_script` runs.

#### Scenario: GitHub Actions sparse-base release

- **WHEN** the GitHub Actions release job runs on a container without bundled `git`
- **THEN** `git` SHALL be on `$PATH` before `actions/checkout@v5` executes
- **AND** `actions/checkout@v5` SHALL receive `fetch-tags: true`

#### Scenario: GitLab CI release

- **WHEN** the GitLab CI release job runs
- **THEN** the install-before-checkout ordering rule SHALL NOT apply
- **AND** the runner-provided `.git/` SHALL satisfy goreleaser's git-repository requirement

### Requirement: Preflight checks enumerate expected status codes

Preflight checks against external APIs (release existence, image existence) SHALL enumerate the expected status codes (typically `200` for "exists" and `404` for "absent"). Any other status code (e.g. `401`, `403`, `5xx`) SHALL fail the job loudly. Silent preflights that return non-success without aborting the pipeline are prohibited.

#### Scenario: Tag-exists preflight returns 200

- **WHEN** the release job preflight checks the GitLab Releases API for the current tag
- **AND** the API returns `200`
- **THEN** the preflight SHALL exit non-zero with an error message naming the existing release

#### Scenario: Tag-exists preflight returns 404

- **WHEN** the release job preflight checks the GitLab Releases API for the current tag
- **AND** the API returns `404`
- **THEN** the preflight SHALL exit zero and the release SHALL proceed

#### Scenario: Tag-exists preflight returns unexpected code

- **WHEN** the release job preflight checks the GitLab Releases API for the current tag
- **AND** the API returns any code other than `200` or `404`
- **THEN** the preflight SHALL exit non-zero with an error message naming the unexpected status

### Requirement: Workflow `uses:` declares permission superset of callee

A workflow that calls another workflow via `uses:` SHALL declare a workflow-level `permissions:` block that is a superset of every scope the callee workflow requests. GitHub Actions aborts reusable-workflow calls with `startup_failure` when the caller's token scope is narrower than the callee's. Adding a scope to a reusable workflow's `permissions:` SHALL add the same scope to every caller in the same change.

#### Scenario: Callee requests packages write

- **WHEN** a caller workflow uses a reusable workflow whose `permissions:` requests `packages: write`
- **THEN** the caller's workflow-level `permissions:` SHALL also grant `packages: write`

#### Scenario: Caller narrower than callee

- **WHEN** a caller workflow's `permissions:` is a strict subset of the callee's `permissions:`
- **THEN** GitHub Actions SHALL abort the workflow with `startup_failure`

### Requirement: docker/build-push-action GHA cache pairs with setup-buildx-action

A workflow that uses `docker/build-push-action` with `cache-to: type=gha` SHALL also configure `docker/setup-buildx-action` (docker-container driver) before the build step. The default docker driver rejects GHA cache export with "Cache export is not supported for the docker driver".

#### Scenario: Buildx with GHA cache

- **WHEN** a workflow invokes `docker/build-push-action` with `cache-to: type=gha`
- **THEN** the workflow SHALL also invoke `docker/setup-buildx-action` beforehand

#### Scenario: Buildx without setup-buildx-action

- **WHEN** a workflow invokes `docker/build-push-action` with `cache-to: type=gha`
- **AND** the workflow does not invoke `docker/setup-buildx-action`
- **THEN** the build SHALL fail with "Cache export is not supported for the docker driver"

### Requirement: Release consumer pulls registry image by digest

When a release consumer (the downstream job that runs goreleaser) references a registry image built earlier in the same workflow, the consumer SHALL pull by digest (`ghcr.io/.../image@sha256:...`), not by tag (e.g. `:latest`). Digest reference eliminates the race against any parallel image rebuild and guarantees byte-for-byte reproducibility.

#### Scenario: Consumer references digest

- **WHEN** a release consumer depends on an image built by an earlier job
- **THEN** the consumer's image reference SHALL include a `@sha256:` digest

#### Scenario: Consumer references mutable tag

- **WHEN** a release consumer depends on an image built by an earlier job
- **AND** the consumer's image reference uses `:latest` or any other mutable tag
- **THEN** the workflow SHALL be vulnerable to a parallel rebuild overwriting the image between build and consume

### Requirement: docker/build-push-action SHALL NOT combine push-by-digest with named tags

`docker/build-push-action` SHALL NOT combine `outputs: type=image,push-by-digest=true` with named `tags:` inputs. Buildx refuses "tagged ref … by digest" — a `:tag` pointing to a content digest is contradictory. Tag-based push is sufficient; `digest` is exposed as a step output regardless of push method.

#### Scenario: push-by-digest with named tags

- **WHEN** a workflow configures `docker/build-push-action` with both `outputs: type=image,push-by-digest=true` and a `tags:` input listing one or more named tags
- **THEN** the push SHALL fail with "can't push tagged ref … by digest"

#### Scenario: push-by-digest without named tags

- **WHEN** a workflow configures `docker/build-push-action` with `outputs: type=image,push-by-digest=true` and no named `tags:`
- **THEN** the push SHALL succeed and the caller MAY consume the digest from the step output

### Requirement: Per-surface CI architecture MAY diverge

Per-surface CI architecture (GitLab CI vs GitHub Actions) MAY diverge when a unified architecture has demonstrated environmental cascades on one surface. The divergence SHALL be documented in the root `AGENTS.md` Release process section, naming the surface, the pattern used, and the failure mode it avoids. The intent is to permit the architecture that survives on each surface, not to enforce identical pipelines.

#### Scenario: Cascading unified architecture

- **WHEN** a unified release pipeline has cascaded through three or more distinct environmental failures on one CI surface
- **THEN** the project MAY switch that surface to a divergent architecture that has not exhibited the cascade
- **AND** the other surface SHALL be allowed to keep its existing pattern if it has not cascaded

#### Scenario: Documented divergence

- **WHEN** two CI surfaces use divergent architectures
- **THEN** the root `AGENTS.md` Release process section SHALL name both patterns and the failure mode each one avoids

### Requirement: TLS-enabled DinD service

The `build-ci-image` job SHALL enable TLS on the Docker-in-Docker service. A non-empty `DOCKER_TLS_CERTDIR` SHALL be set in the job `variables:` block; the daemon SHALL present a TLS certificate on port 2376; the client SHALL connect via `tcp://docker:2376` with the client cert set mounted.

#### Scenario: TLS handshake between client and service

- **WHEN** `build-ci-image` runs
- **THEN** `DOCKER_TLS_CERTDIR` SHALL be a non-empty path
- **AND** `docker build` SHALL complete via the TLS socket on port 2376

#### Scenario: Plaintext DinD

- **WHEN** `build-ci-image` runs without `DOCKER_TLS_CERTDIR` set
- **THEN** the daemon SHALL listen on the plaintext port 2375
- **AND** a co-tenant runner SHALL be able to drive `docker build` against the registry

### Requirement: Mirror push is non-forced

The `mirror-to-github` job SHALL push the GitHub mirror without `--force`. The push SHALL be plain `git push github HEAD:$CI_COMMIT_BRANCH --tags` and SHALL fail rather than overwrite remote refs.

#### Scenario: Main-branch sync

- **WHEN** a commit is pushed to `main` and the mirror job runs
- **THEN** the push SHALL NOT include `--force`
- **AND** the push SHALL succeed only if the remote ref fast-forwards

#### Scenario: Non-fast-forward local branch

- **WHEN** the mirror job's local branch cannot fast-forward the remote ref
- **THEN** the push SHALL fail
- **AND** the remote ref SHALL NOT be overwritten

### Requirement: Release publish is non-overwriting on tag reuse

The `release` job SHALL verify the tag has no existing GitLab release before invoking goreleaser. The goreleaser configuration SHALL set `release.replace_existing_artifacts: false` so goreleaser never silently overwrites artifacts when invoked. The preflight gate SHALL treat `200` (exists) as fatal and `404` (absent) as proceed; other codes SHALL also be fatal.

#### Scenario: First release of a tag

- **WHEN** CI runs `release` for a previously-unreleased tag
- **THEN** the preflight SHALL return a non-success code
- **AND** goreleaser SHALL publish artifacts to the GitLab release URL

#### Scenario: Re-tag of an existing release

- **WHEN** CI runs `release` for a tag that already has a published release
- **THEN** the preflight SHALL return `200`
- **AND** CI SHALL fail the `release` job before goreleaser is invoked

### Requirement: CI supply-chain signals

The CI pipeline SHALL run `govulncheck` on twiggit's source code at the `lint` stage and SHALL sign release SBOMs with cosign. The release job's cosign signing mode SHALL prefer OIDC keyless signing via the runner's OIDC identity and SHALL fall back to a `COSIGN_KEY` env var when OIDC is unavailable.

#### Scenario: govulncheck on source

- **WHEN** the `lint` job runs
- **THEN** a `govulncheck ./...` step SHALL run after `go mod tidy`
- **AND** unfixed vulnerabilities SHALL exit the step non-zero

#### Scenario: cosign SBOM signature

- **WHEN** the `release` job finishes `goreleaser release`
- **THEN** each `*_sbom.spdx.json` artifact SHALL be signed by cosign
- **AND** `cosign verify-blob` SHALL succeed for a consumer with the right certificate identity

#### Scenario: cosign signing mode

- **WHEN** the `release` job runs and the runner exposes an OIDC identity
- **THEN** cosign SHALL sign via OIDC keyless signing
- **AND** when OIDC is unavailable and `$COSIGN_KEY` is set, cosign SHALL sign via the keyful path

### Requirement: Pinned CI tooling images

CI jobs that require a goreleaser container SHALL pin the image to a specific tag. The pinned version SHALL match the goreleaser entry in `.mise/config.toml`. The `build-ci-image` trigger SHALL cover `.mise/config.toml` and `.goreleaser.yml` so a goreleaser-version bump rebuilds the image before downstream jobs run.

#### Scenario: Goreleaser pin in job image

- **WHEN** the `goreleaser-dry-run` or `release` job runs
- **THEN** the job image SHALL be a pinned tag (e.g. `golang:1.27.2-alpine3.23`)
- **AND** the pinned image SHALL NOT be `latest`

#### Scenario: Goreleaser version bump rebuilds image

- **WHEN** `.mise/config.toml` `goreleaser` entry is bumped
- **THEN** the `build-ci-image` job SHALL queue before downstream jobs
- **AND** the rebuilt image SHALL carry the new goreleaser version

### Requirement: Module-graph freshness gate

The CI pipeline SHALL verify that the module graph declared by `go.mod`/`go.sum` matches the source under review. The check SHALL run in the `lint` job before `golangci-lint run`.

#### Scenario: Lint gate

- **WHEN** the `lint` job runs
- **THEN** it SHALL first execute `go mod tidy && git diff --exit-code`
- **AND** an unstaged `go.mod`/`go.sum` drift SHALL fail the lint job

### Requirement: Pipeline defaults for jobs

The CI `.gitlab-ci.yml` `default:` block SHALL declare the retry policy and interruptibility expectation that apply to every job unless a job overrides them. Test failures SHALL NOT be retried by the default policy.

#### Scenario: Default retry applies

- **WHEN** a job does not override `retry:`
- **THEN** runner-system or stuck-or-timeout failures SHALL be retried up to the `default:` `retry:` count
- **AND** test failures SHALL NOT be retried by the default policy

#### Scenario: Default interruptibility

- **WHEN** a job does not override `interruptible:`
- **THEN** a new pipeline on the same ref SHALL be able to cancel the in-flight job (interruptible: true)

### Requirement: Release job is non-interruptible

The `release` job SHALL set `interruptible: false` so a concurrent pipeline on the same ref cannot abort a publish mid-flight.

#### Scenario: Release in flight

- **WHEN** `release` is running and a new pipeline starts on the same ref
- **THEN** the new pipeline SHALL NOT cancel `release`
- **AND** `release` SHALL complete naturally or fail on its own merits

### Requirement: CI image triggers include release tooling

The `build-ci-image` job SHALL trigger on changes to `.goreleaser.yml` and `.mise/config.toml` in addition to its existing triggers, so a goreleaser-version bump or goreleaser-config edit rebuilds the CI image automatically.

#### Scenario: Goreleaser config edit

- **WHEN** a commit modifies `.goreleaser.yml`
- **THEN** `build-ci-image` SHALL queue before downstream jobs
- **AND** the rebuilt image SHALL carry the goreleaser version expected by `.gitlab-ci.yml`

### Requirement: Goreleaser dry-run on release-affecting merge requests

The `goreleaser-dry-run` job SHALL run on merge requests that touch release-affecting paths (`cmd/`, `internal/`, `.goreleaser.yml`, `go.mod`, `go.sum`). Snapshot artifacts SHALL expire within one day.

#### Scenario: Release-affecting trigger

- **WHEN** a merge request touches a release-affecting path
- **THEN** `goreleaser-dry-run` SHALL run
- **AND** the result SHALL gate the MR

#### Scenario: Docs-only trigger

- **WHEN** a merge request touches only docs or non-release paths
- **THEN** `goreleaser-dry-run` SHALL NOT run

#### Scenario: Snapshot retention

- **WHEN** `goreleaser-dry-run` publishes snapshot artifacts
- **THEN** the artifacts SHALL expire within one day
