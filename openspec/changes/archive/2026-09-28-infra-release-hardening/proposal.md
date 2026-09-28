# Proposal

## Why

Commit `3d53c3b` ("Adopt Go 1.27 modernizers and mise-only tool ownership")
closed most of change **A** (foundation: go.mod bump, mise tool pin, linter
expansion, `formatters:` block, `ci:coverage` task, infra-toolchain spec) but
left mechanical residue in `.golangci.yml` and `.dockerignore`, and shipped
no part of change **B** (CI security gates). The REVIEW titled
`REVIEW.md:2026-09-25` flagged six critical findings under
`infrastructure-release` — TLS-disabled DinD, `--force` mirror push,
`release.mode: replace` overwrite, plus missing govulncheck CI job / CodeQL
/ Dependabot — of which this change ships the TLS DinD, non-forced
mirror, and `replace_existing_artifacts: false` remedies, plus promotes
the previously-deferred govulncheck CI gate to active. The remaining
items (B2 CodeQL, B3 Dependabot, B13 homebrew token scope, B14 SLSA
provenance) remain explicitly deferred.

## What Changes

### Foundation polish (no spec delta — `infrastructure-toolchain` already governs this surface)

- **A2** `.dockerignore` excludes `openspec/` and `.github/` so the build
  context sent to `docker build` does not ship planning artifacts
  (`.gitlab/` dropped from the original scope — the project's CI image
  registry is private and does not need build-context leak protection).
- **A4** `.golangci.yml` test-exclusion regex gains `^` anchors and a
  narrower surface so it does not catch unrelated `*_test.go` paths in
  subpackages.
- **A5** `staticcheck.checks: ["all,-ST1000"]` enumerates the excluded
  checks explicitly (e.g. `["all,-ST1000,-ST1003,-ST1005,-ST1016,-ST1020,-ST1021"]`)
  so future re-enablement is a one-line change.
- **A6** `.golangci.yml` `cmd:` depguard `allow` drops the redundant
  self-include of `twiggit/cmd`; depguard already allows same-package
  imports implicitly.
- **A7** `.golangci.yml` `cmdutil:` depguard allow-list narrows per the
  skill contract — `cmdutil` is for Factory + exit codes + persistent
  flags + consumer interfaces; direct imports of `internal/{output,git,config,version}`
  are removed in favor of the documented `main.go` composition root.
  The narrow is gated on a `cmdutil.NewFactory` refactor that moves the
  composition-root wiring out of `cmdutil` and into `main.go`. The new
  `NewFactory(opts ...FactoryOption)` signature accepts three functional
  options: `WithVersion(string)` (replaces the eager `version.Version`
  read), `WithConfigLoader(func() (*core.Config, error))`, and
  `WithGitClientFactory(func() (interface{}, error))`. The `Factory.GitClient`
  field type changes from `*git.Client` to `interface{}` so `*git.Client`
  satisfies it implicitly (Go's structural interface satisfaction) without
  `cmdutil` importing `internal/git`. `main.go` calls
  `cmdutil.NewFactory(cmdutil.WithVersion(version.Version), cmdutil.WithConfigLoader(loadConfig), cmdutil.WithGitClientFactory(newGitClient))`
  where `loadConfig` and `newGitClient` close over `config.NewManager()`
  and `git.NewClient()`. The default no-arg `NewFactory()` call returns a
  Factory with the same eager fields as today (`IOStreams`, `Context`,
  `Executable`) plus lazy fields that return a sentinel error if invoked
  — preserving every existing test literal in `factory_test.go`. This is
  sequenced after the rest of the foundation polish in Group 7 because
  the depguard narrow immediately fails `golangci-lint run ./internal/cmdutil/...`
  until the imports drop.
- **A10** `.mise/config.toml` `gopls:check` switches from
  `gopls check **/*.go` (filesystem glob) to `gopls check ./...`
  (module-aware), so it respects `go.mod` boundaries.
- **A11** `.mise/config.toml` adds `cosign` to `[tools]` using the
  `github.com/sigstore/cosign/v3/cmd/cosign` Go module path so local dev
  matches the `cosign v3.1.3` bundled in `goreleaser/goreleaser:v2.18.2`.
- **A12** `.mise/config.toml` adds a `min_version` directive to pin
  the minimum mise release that the project's `[env] experimental = true`
  and Go backend syntax rely on.
- **A13** `AGENTS.md` and `CONTRIBUTING.md` developer-bootstrap section
  records that `mise install && pre-commit install` provisions gopls,
  golangci-lint, goreleaser, govulncheck, cosign, and pre-commit hooks
  on a clean checkout (replaces the proposal's A9 `install.sh` edit,
  which targeted the end-user binary installer in error).

### CI / release hardening (modifies `infrastructure-release`)

- **B4** `mirror-to-github` job drops `--force` from `git push`. The push
  becomes non-forced for main-branch syncs and accepts `--tags` only on
  the explicit tag trigger (which already routes through `when: never`
  for `mirror-to-github`).
- **B5** `.goreleaser.yml` keeps `release.mode: replace` (release-notes
  merge mode is unrelated to artifact overwrite) and adds
  `release.replace_existing_artifacts: false` so a re-tag never silently
  overwrites artifacts. The release job's `before_script` adds a GitLab
  Releases API preflight (`curl -sI` against `/releases/v<tag>`) that
  fails the job when a release for the tag already exists. The goreleaser
  flag `--fail-if-tag-exists` does not exist in v2.18.2; replaced with
  the preflight gate.
- **B6** `.gitlab-ci.yml` `build-ci-image` enables TLS on the existing
  `docker:29.1.4-dind` service by setting `DOCKER_TLS_CERTDIR: "/certs"`
  in job `variables:` (the `:tls` image variant does not exist; TLS is
  triggered by the env var per `dockerd-entrypoint.sh`). The client sets
  `DOCKER_HOST: tcp://docker:2376` and mounts `/certs/client` from the
  service into the client container.
- **B7** `.gitlab-ci.yml` `lint` job runs `go mod tidy && git diff --exit-code`
  before `golangci-lint run` so the diff between `go.mod`/`go.sum` and
  source is caught at the same gate.
- **B9** `.gitlab-ci.yml` `test` job runs `go test -race ./...` so the
  race detector flags failures at the same job as the unit tests (folded
  in from a previously-planned standalone `test:race` job; the
  `mise run ci:coverage` task keeps the canonical coverage signal on
  tagged releases).
- **B11** `.gitlab-ci.yml` `release` job adds a `cosign sign` step after
  `goreleaser release --clean` (keyless via GitLab OIDC using the
  image-bundled `cosign v3.1.3` at `/usr/bin/cosign`). The job configures
  `id_tokens: { SIGSTORE_ID_TOKEN: { aud: sigstore } }` for OIDC and a
  `before_script` switch for `$COSIGN_KEY` when OIDC is unavailable.
- **B12** `.gitlab-ci.yml` `goreleaser-dry-run` and `release` jobs pin
  the goreleaser image to `goreleaser/goreleaser:v2.18.2` matching
  `.mise/config.toml:4`; unversioned `goreleaser/goreleaser` is removed.
- **B15** `.gitlab-ci.yml` `mirror-to-github` extends its rule set: the
  `when: never` for tag triggers (already present) is paired with an
  explicit `interruptible: false` for the main-branch sync, and the
  push switches to `git push --follow-tags` (no `--force`).
- **B16** `.gitlab-ci.yml` `goreleaser-dry-run` scopes its trigger to
  release-affecting paths (`cmd/`, `internal/`, `.goreleaser.yml`,
  `go.mod`, `go.sum`) so a docs-only MR does not pay for a goreleaser
  dry-run that cannot fail.
- **B17** `.gitlab-ci.yml` `default:` block sets explicit `retry:`
  (2 max for transient errors) and `interruptible: true` (overridable
  per-job). `tags:` is dropped from the default block — twiggit uses no
  self-hosted runner tags; the GitLab-hosted runner default applies.
- **B19** `.gitlab-ci.yml` `release` job sets `interruptible: false`
  so an upstream cancel cannot abort the publish mid-flight.
- **B20** `.gitlab-ci.yml` `build-ci-image` trigger extends to
  `.goreleaser.yml` and `.mise/config.toml` so a goreleaser-version bump
  that needs a fresh image rebuilds the CI image automatically (drift
  between the goreleaser pin and the CI image is closed at the
  image-rebuild step; no per-job drift check is required).
- **B22** `.gitlab-ci.yml` `goreleaser-dry-run` artifacts use
  `expire_in: 1 day` so the snapshot churns do not fill runner disk.
- **B24** `.gitlab-ci.yml` `lint` job adds a `govulncheck ./...` step
  after `go mod tidy` so unfixed vulnerabilities in called code from
  twiggit's source dependencies (charm.land, samber/lo, go-git, etc.)
  fail the `lint` job. `govulncheck` is already in the CI image via
  `mise` (toolchain governance); no per-run install is required.
  This is the previously-deferred B1 promoted to active; it replaces
  the originally-planned Trivy image scan on the CI image, which
  produced only false positives from vendored test fixtures and
  toolchain-binary CVEs (see `design.md` Decisions rationale).

## Capabilities

### New Capabilities

None. No new runtime behavior is introduced; all changes live inside
existing CI/release surfaces or as tooling polish.

### Modified Capabilities

- `infrastructure-release`: gains requirements covering TLS-enabled
  DinD service, non-forced mirror push, GitLab Releases API preflight
  gate on release-tag reuse plus `release.replace_existing_artifacts:
  false`, `govulncheck` on the `lint` job, cosign SBOM sign, pinned
  goreleaser image, `-race` flag in the `test` job, `go mod tidy`
  gate in the `lint` job, path-filtered goreleaser-dry-run, and
  pipeline defaults (`retry:` scoped to transient failures,
  `interruptible:`).

## Non-Goals

- **B2** CodeQL / SARIF publish — deferred. gosec inside
  golangci-lint covers SAST locally today; surfacing results in the
  GitLab Security tab requires a future SARIF-upload job.
- **B3** Dependabot / Renovate — deferred. Manual `go get -u=patch`
  cadence is documented in `CONTRIBUTING.md`; an automated bot is its
  own change because it requires `.github/` (for Dependabot) or
  `.gitlab/renovate.json5` (for Renovate) config and a CI hook to
  surface PRs back.
- **B13** `homebrew_casks.token` scope documentation — deferred. The
  broad-scope token concern stays until the token-rotation policy is
  re-evaluated; out of scope for the release-safety pass.
- **B14** Full SLSA provenance — deferred. Requires sigstore/cosign
  key infrastructure beyond B11's SBOM sign. Will follow once the
  signing path stabilizes.

## Impact

### Layer rollup

Per the `golang-cli` Tier 2 convention, `main.go` is the composition root
and `cmd/` is the Cobra tree assembly layer. Every change in this PR lives
in `main.go` (Group 7 wiring migration) or in repo-root config / CI YAML /
Docker / linter config (Groups 1-6). No `cmd/` command file or `internal/core`
/ `internal/git` / `internal/config` / `internal/output` / `internal/iostreams`
package is touched. The single Go-code change is the cmdutil refactor
(Group 7), which moves responsibility without changing observable behavior.

### Files modified

```
.gitlab-ci.yml                  — B4, B6, B7, B9, B11, B12, B15, B16,
                                   B17, B19, B20, B22, B24
.dockerignore                   — A2
.golangci.yml                   — A4, A5, A6, A7 (A7 sequenced after
                                   Group 7)
.goreleaser.yml                 — B5
.mise/config.toml               — A10, A11, A12
AGENTS.md                       — A13
CONTRIBUTING.md                 — A13
internal/cmdutil/factory.go     — Group 7 (refactor: lazy constructors
                                   via functional options)
main.go                         — Group 7 (composition-root wiring of
                                   config/git/version)
internal/cmdutil/factory_test.go — Group 7 (test seam: exercise both
                                   default and wired construction paths)
openspec/specs/infrastructure-release/spec.md     — proposal→specs phase
openspec/changes/infra-release-hardening/specs/infrastructure-release/
                                                  — delta spec
```

### Files unchanged (within this change's typical scope)

- `cmd/` (all command files and `cmd/root.go`) — implementation
  untouched. `cmd/root.go` continues to receive `*cmdutil.Factory` from
  `main.go` without owning its construction; the `CommandConfig = cmdutil.Factory`
  type alias at `cmd/root.go:16` stays unchanged.
- `internal/{core,git,config,output,iostreams}` — implementation
  untouched. The `cmdutil` refactor moves responsibility for constructing
  `config.NewManager()` and `git.NewClient()` into `main.go`; the
  packages themselves are unchanged.
- `go.mod`, `go.sum` — already current.
- All pre-commit hooks — already aligned with `infrastructure-toolchain`.

### Operational impact

- CI minutes drop: single Go version replaces the originally-planned
  two-entry matrix, and the `-race` flag folds into the main `test`
  job (no standalone `test:race` or `test:race:subset`). The
  `goreleaser-dry-run` job is now path-filtered to release-affecting
  MRs, so docs-only MRs no longer pay the goreleaser cost.
- TLS DinD (B6) requires GitLab Runner DinD TLS cert distribution
  via `DOCKER_TLS_CERTDIR: "/certs"`. Self-hosted runners emit the certs
  automatically through `dockerd-entrypoint.sh`; GitLab.com shared
  runners behave the same.
- Release-tag reuse protection (B5) requires the GitLab Releases API
  preflight in the `release` job `before_script`. A re-tag of an existing
  tag fails the preflight before goreleaser runs; no local `release:tag`
  update is required because the local task already exits 1 on existing
  tags (`.mise/tasks/release/tag` lines 32-35).
- Homebrew tap behavior is independent of `release.mode` and is
  governed by `homebrew_casks.skip_upload`; no change required.
- `govulncheck ./...` (B24) runs in the `lint` job on every MR + tag.
  `govulncheck` is already in the CI image via `mise` (no per-run
  install). Real supply-chain signal: vulnerabilities in called code
  from twiggit's actual dependencies. The previously-considered
  Trivy image scan (B10, since dropped) scanned the ephemeral CI
  image and produced only false positives from vendored test fixtures
  and toolchain-binary CVEs.

### Compatibility

- **BREAKING (CI-side only)**: `.goreleaser.yml` gains
  `release.replace_existing_artifacts: false`. Existing GitLab releases
  on prior tags remain; subsequent re-tags no longer overwrite artifact
  binaries. Consumers that re-tag an upstream consumer's prebuilt binary
  must delete the prior GitLab release first.
- **BREAKING (CI-side only)**: `goreleaser/goreleaser` image pin (B12).
  If a runner image does not include `v2.18.2`, the build pulls the
  official pinned image.
- Non-breaking: DinD TLS (B6), `govulncheck` in lint (B24), cosign
  (B11), `-race` fold into test (B9), `cmdutil` factory refactor
  (Group 7), and all A-post items affect CI/toolchain only or are
  internal rewirings with no externally observable behavior change.
