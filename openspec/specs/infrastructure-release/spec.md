# Capability: Release and Distribution

## Purpose

CI/CD pipeline, release artifact distribution (GitLab primary, GitHub
discoverability, Homebrew tap), and contributor onboarding. This spec
is the canonical owner of release-process behavior; per-concern
details are cross-referenced.

## Requirements

### Requirement: CI pipeline

The CI pipeline SHALL build the binary, run lint, and run the full
test matrix on every push and pull request. Coverage enforcement and
artifact upload run on tagged releases.

#### Scenario: Push to main

- **WHEN** a commit is pushed to `main`
- **THEN** CI SHALL run `mise run lint` and `mise run test`
- **AND** SHALL upload a debug binary for manual verification

#### Scenario: Tagged release

- **WHEN** a tag matching `v*` is pushed
- **THEN** CI SHALL run GoReleaser
- **AND** SHALL publish artifacts to GitLab releases
- **AND** SHALL push a Homebrew formula to `amoconst/homebrew-tap`

### Requirement: GitLab primary distribution

GitLab releases SHALL host the primary artifacts (binary tarballs,
checksums, packages). GitHub SHALL host a discoverability page that
mirrors the release notes but SHALL NOT be the source of truth.

#### Scenario: Primary artifact on GitLab

- **WHEN** a release is published
- **THEN** the primary tarball, checksum, and SBOM SHALL be on
  GitLab releases

#### Scenario: GitHub discoverability

- **WHEN** a user visits the GitHub repo
- **THEN** releases page SHALL link to the GitLab release

### Requirement: Homebrew tap

GoReleaser SHALL publish a `twiggit` formula to
`amoconst/homebrew-tap` on every tag. The formula SHALL install the
binary, and on macOS SHALL remove the quarantine attribute after
install.

#### Scenario: Tap formula updated

- **WHEN** a `v*` tag is released
- **THEN** `homebrew-tap` SHALL receive a `twiggit.rb` update via
  GoReleaser's `brews:` block
- **AND** `brew upgrade twiggit` SHALL fetch the new version

#### Scenario: macOS quarantine

- **WHEN** a user installs via `brew install twiggit` on macOS
- **THEN** the formula SHALL call `xattr -d com.apple.quarantine`
  on the binary after install
- **AND** `twiggit` SHALL run without a Gatekeeper prompt

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
validation suite, and the project's commit conventions.

#### Scenario: New contributor setup

- **WHEN** a new contributor reads `CONTRIBUTING.md`
- **THEN** they SHALL find: toolchain install via `mise install`,
  pre-commit hook install via `pre-commit install`, and the
  `mise run verify` validation command

### Requirement: Test coverage policy

The project SHALL target ≥70% coverage on all packages. The cmd
package is tested exclusively via E2E (Ginkgo/Gomega); unit tests in
`cmd/` are not required and SHALL NOT be added. See `testing-e2e`.

#### Scenario: Coverage gate

- **WHEN** `mise run test:coverage` runs
- **THEN** coverage SHALL be reported per package
- **AND** packages below 70% SHALL be flagged (not blocking)

#### Scenario: No cmd/ unit tests

- **WHEN** developer inspects `cmd/`
- **THEN** only E2E tests (in `test/e2e/`) SHALL cover cmd behavior
- **AND** no `_test.go` files SHALL exist alongside the cmd source
  files (other than `cmd/error_formatter_test.go` and
  `cmd/util_test.go`, which test pure helpers)

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
