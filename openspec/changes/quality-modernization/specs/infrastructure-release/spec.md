# Spec Delta

## ADDED Requirements

### Requirement: Govulncheck gate in CI

The CI pipeline SHALL run `go tool govulncheck ./...` as a step
in the validate or test stage. The step SHALL fail the pipeline
when `govulncheck` reports any finding that does not appear in
the committed baseline file `govulncheck.json`. The baseline
file SHALL be regenerated whenever dependencies are patched and
the new finding set is reviewed.

#### Scenario: New vulnerability surfaces after dependency bump
- **WHEN** the project bumps a dependency and
  `govulncheck ./...` reports a new advisory
- **THEN** the CI pipeline fails until either the dependency
  is fixed or the advisory is reviewed and added to the
  baseline with a documented justification

### Requirement: Test job runs with race detector

The CI test job SHALL invoke the test suite with `-race`. The
`-race` flag SHALL be applied to the unit, integration, and
E2E test tasks so that data races are detected in any layer.

#### Scenario: Concurrent test exposes race
- **WHEN** a concurrent test (under `test/concurrent/`)
  contains a data race that the unit test suite does not
  detect
- **THEN** the `-race`-enabled CI run fails with a stack
  trace of the race

### Requirement: Coverage threshold gate is non-zero

The CI coverage step SHALL exit with a non-zero status when
the filtered (internal-packages-only) coverage falls below the
configured threshold. The threshold SHALL be defined in the
coverage script and SHALL be at least 70%. The warning-only
behavior is deprecated.

#### Scenario: Coverage below threshold fails CI
- **WHEN** a change reduces filtered coverage to 65%
- **THEN** the coverage step exits non-zero and the pipeline
  fails before the release stage

### Requirement: Dockerfile.ci base image is digest-pinned

The `Dockerfile.ci` SHALL pin the base image by `@sha256:...`
digest in addition to (or instead of) the versioned tag. The
digest SHALL be updated only via a deliberate dependency
review; CI SHALL detect digest drift between the tag and
the locked digest.

#### Scenario: Base image digest is locked
- **WHEN** `Dockerfile.ci` references `golang:1.25.5-alpine3.23`
- **THEN** the `FROM` line also pins the resolved
  `@sha256:...` digest and any change to the tag without
  updating the digest is flagged in code review

### Requirement: Mise installer verified by sha256

`Dockerfile.ci` SHALL download the mise release tarball from
the official GitHub release, verify its `sha256sum` against a
value pinned in the file, and only then extract the binary.
The unpinned `curl https://mise.run | sh` pattern SHALL NOT
appear.

#### Scenario: Installer step verifies sha256
- **WHEN** `Dockerfile.ci` installs mise
- **THEN** the script fetches the tarball, computes
  `sha256sum`, compares against the pinned value, and aborts
  on mismatch with a non-zero exit

### Requirement: Goreleaser owner matches GitHub mirror

The GoReleaser config SHALL set the GitHub owner to the
organization or user that owns the GitHub mirror of the
project. The GitHub mirror's repository URL is the canonical
target for release artifacts and the homebrew tap formula
source. The GoReleaser config and the hook script
(`.goreleaser/hooks/create-github-release.sh`) SHALL name
the same owner and repository.

#### Scenario: Owner mismatch surfaces as a release failure
- **WHEN** `.goreleaser.yml` names one owner and the hook
  script targets a different owner
- **THEN** the homebrew formula URL or the GitHub release
  creation fails; the project SHALL run `mise run
  release:dry-run` and verify both paths resolve before
  tagging a release

### Requirement: Goreleaser hook fails on GitHub API errors

`.goreleaser/hooks/create-github-release.sh` SHALL exit with a
non-zero status when the GitHub release API returns a
non-2xx response. The `|| echo "Failed..."` swallow pattern
SHALL NOT appear; the script SHALL use `set -euo pipefail`
and propagate errors.

#### Scenario: GitHub API returns 5xx
- **WHEN** the hook script calls the GitHub release API and
  receives a 5xx response
- **THEN** the script exits non-zero and the goreleaser
  pipeline marks the step failed

### Requirement: Dependabot config committed

A `.github/dependabot.yml` file SHALL exist at the repository
root and SHALL configure Dependabot to monitor the project's
primary dependency sources: `gomod`, `github-actions`, and
`docker` (for `Dockerfile.ci`). The schedule SHALL be weekly
on Monday. Grouped updates SHALL merge patch and minor bumps
for non-security dev-dependencies into a single PR.

#### Scenario: Dependabot opens weekly PR
- **WHEN** a Monday passes and new patch versions exist for
  any monitored dependency
- **THEN** Dependabot opens a PR (or grouped PR) within the
  same business day

### Requirement: Main-branch CI rules

The `.gitlab-ci.yml` workflow rules SHALL include the main
branch in the cache-rebuild, lint, and test jobs. Direct
pushes to main SHALL run the full validation suite (lint +
test + setup cache) the same way merge requests do.

#### Scenario: Direct push to main triggers full validation
- **WHEN** a developer force-pushes (or fast-forwards) main
- **THEN** the lint, test, and cache jobs all run and the
  pipeline does not short-circuit on missing rules

### Requirement: Resource group shared by lint, test, setup

The `.gitlab-ci.yml` SHALL define a resource group named
`twiggit-ci` that the lint, test, and setup-cache jobs join.
Concurrent pipelines SHALL serialize through the resource
group so cache writes do not race.

#### Scenario: Two pipelines race for cache
- **WHEN** two pipelines start within the same minute and
  both want to rebuild the mise/go cache
- **THEN** the second pipeline waits for the first to release
  the `twiggit-ci` resource group before starting its own
  cache-write step
