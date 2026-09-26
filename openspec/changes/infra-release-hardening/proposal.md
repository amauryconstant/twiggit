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
/ Dependabot — three of which this change ships the remedies for. The
remaining three (B1, B2, B3) plus B13 (homebrew token scope) and B14 (SLSA
provenance) are explicitly deferred.

## What Changes

### Foundation polish (no spec delta — `infrastructure-toolchain` already governs this surface)

- **A1** `Dockerfile.ci` final stage runs as a non-root user; document the
  UID/permission model alongside the existing `USER nonroot` convention.
- **A2** `.dockerignore` excludes `.gitlab/`, `openspec/`, `.github/` so the
  build context sent to `docker build` does not ship planning artifacts.
- **A3** `.gomodcache` (and any other persistent CI paths) gain a known
  owner before the `USER nonroot` cutover in A1; chown or relocate to the
  runtime user.
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
- **B8** `.gitlab-ci.yml` adds a Go-version matrix that exercises the
  toolchain set: `go.mod` directive version (currently `1.27.1`) plus
  the prior supported minor (currently `1.26.x` latest patch);
  `fail-fast: false` so a single-broken-matrix entry does not cancel
  the others.
- **B9** `.gitlab-ci.yml` adds a standalone `test:race` CI job (currently
  `-race` only runs as a flag inside `mise run ci:coverage`), so race
  failures surface at their own job instead of being folded into
  coverage noise.
- **B10** `.gitlab-ci.yml` `build-ci-image` adds a `trivy image` scan
  step after the `docker push`; findings gate the `validate` stage for
  subsequent jobs.
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
- **B16** `.gitlab-ci.yml` `goreleaser-dry-run` no longer scopes its
  trigger to `.goreleaser.yml`/`go.mod`/`cmd`/`internal` changes — it
  runs on every MR so a refactor that touches the release binary is
  caught before merge.
- **B17** `.gitlab-ci.yml` `default:` block sets explicit `tags:`,
  `retry:` (2 max for transient errors), and `interruptible: true`
  (overridable per-job).
- **B18** `.gitlab-ci.yml` `test` job `before_script` adds a
  `go mod tidy -diff` verification so uncommitted module-graph drift
  fails the job before tests run.
- **B19** `.gitlab-ci.yml` `release` job sets `interruptible: false`
  so an upstream cancel cannot abort the publish mid-flight.
- **B20** `.gitlab-ci.yml` `build-ci-image` trigger extends to
  `.goreleaser.yml` so a goreleaser-version bump that needs a fresh
  image rebuilds the CI image automatically.
- **B21** `.dockerignore` excludes `.cache-key` (the cache-key file
  written by `setup:` job) so it does not leak into the docker build
  context.
- **B22** `.gitlab-ci.yml` `goreleaser-dry-run` artifacts use
  `expire_in: 1 day` so the snapshot churns do not fill runner disk.
- **B23** `.gitlab-ci.yml` adds a fast `test:race:subset` job that
  runs `go test -race ./internal/... ./cmd/...` (per-target) so the
  default branch has a quick race pass; the existing
  `mise run ci:coverage` keeps the full race run for tagged releases.

## Capabilities

### New Capabilities

None. No new runtime behavior is introduced; all changes live inside
existing CI/release surfaces or as tooling polish.

### Modified Capabilities

- `infrastructure-release`: gains requirements covering TLS-enabled
  DinD service, non-forced mirror push, GitLab Releases API preflight
  gate on release-tag reuse plus `release.replace_existing_artifacts:
  false`, Trivy image scan, cosign SBOM sign, pinned goreleaser image,
  Go-version matrix, standalone race CI job, `go mod tidy` gate,
  matrix-less dry-run, and pipeline defaults (`tags:`, `retry:` scoped
  to transient failures, `interruptible:`).

## Non-Goals

- **B1** govulncheck CI job — deferred. `mise run vuln:check` is the
  developer-local + pre-commit gate today; promoting it to CI is a
  follow-up change once the linter expansion in change A has settled.
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
.gitlab-ci.yml                  — B4, B6, B7, B8, B9, B10, B11, B12,
                                  B15, B16, B17, B18, B19, B20, B22,
                                  B23
Dockerfile.ci                   — A1, A3
.dockerignore                   — A2, B21
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

- CI minutes roughly double for matrix expansion (B8) and
  fast-subset race (B23). Acceptable; the matrix failure is the
  release-blocking signal, not the race subset.
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

### Compatibility

- **BREAKING (CI-side only)**: `.goreleaser.yml` gains
  `release.replace_existing_artifacts: false`. Existing GitLab releases
  on prior tags remain; subsequent re-tags no longer overwrite artifact
  binaries. Consumers that re-tag an upstream consumer's prebuilt binary
  must delete the prior GitLab release first.
- **BREAKING (CI-side only)**: `goreleaser/goreleaser` image pin (B12).
  If a runner image does not include `v2.18.2`, the build pulls the
  official pinned image.
- Non-breaking: DinD TLS (B6), Trivy (B10), cosign (B11), matrix
  (B8), `cmdutil` factory refactor (Group 7), and all A-post items
  affect CI/toolchain only or are internal rewirings with no
  externally observable behavior change.
