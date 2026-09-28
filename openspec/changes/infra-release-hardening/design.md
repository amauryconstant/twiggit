# Design

## Context

Commit `3d53c3b` already shipped the foundation layer (`go.mod 1.27.1`,
`mise.toml` tool pin, formatters block, `nolintlint`, `ci:coverage` task).
The remaining work is two-layered: (a) mechanical config polish across
`.golangci.yml`/`.dockerignore`/`Dockerfile.ci`/IDE files, and (b) CI and
release pipeline hardening where the architectural choices matter — TLS
DinD, goreleaser publish semantics, cosign signing mode, and matrix
configuration. This document focuses the (b) layer; the (a) layer is a
straightforward migration with no design tension. See `proposal.md` for
motivation, and `specs/infrastructure-release/spec.md` for the
behavioral contract each decision must satisfy.

The change ships in a single OpenSpec change. Tooling choices follow
the existing project standard (mise owns tools; `cosign` and
`govulncheck` are provisioned via the mise `[tools]` block, and
`govulncheck` is now wired into the `lint` job as the supply-chain
signal for twiggit's source dependencies). All choices below assume
that governance is unchanged.

## Goals / Non-Goals

**Goals:**

- Close REVIEW-B critical findings (B4, B5, B6) and most warn findings
  (B7, B9–B12, B15, B17, B19, B20, B22) with minimal CI-minute
  expansion.
- Preserve the existing `infrastructure-toolchain` contract: mise owns
  tools, `mise run verify` order is unchanged, govulncheck is now
  both developer-local + pre-commit AND a CI gate on the `lint` job.
- Keep the change's risk surface localized to `.gitlab-ci.yml`,
  `.goreleaser.yml`, and the Docker lint files. `cmd/` and `internal/`
  must remain untouched.

**Non-Goals:**

- Realize the B2, B3, B13, B14 deferred items (CodeQL SARIF,
  Dependabot/Renovate, homebrew token scope, SLSA provenance). Those
  are explicitly deferred per the proposal's Non-Goals section.
- Cut over to Buildah/Kaniko or another daemon-free builder; TLS DinD
  is the chosen path. Buildah is revisited only if (B6) reveals a
  second-order runtime problem.
- Modify `cmd/`, `internal/`, or any consumer of git operations.
- Change the `infra-release` Homebrew tap configuration beyond what
  `release.mode: append` requires.

## Decisions

### 1. TLS DinD via `DOCKER_TLS_CERTDIR` env var on the existing image

**Choice:** Keep the `docker:29.1.4-dind` service image unchanged; set
`DOCKER_TLS_CERTDIR: "/certs"` in the job `variables:` block. The
`dockerd-entrypoint.sh` script in the official `docker` image writes the
CA, server cert, and client cert into `/certs` automatically and binds
the daemon to TLS on port 2376. The client job sets `DOCKER_HOST:
tcp://docker:2376` and mounts `/certs/client` from the service.

**Rationale:** Docker image tags do not accept a `:variant` syntax
(`docker:29.1.4-dind:tls` returns 404 on Docker Hub; verified via the
Hub Registry API). The TLS-enabled mode is triggered by the
`DOCKER_TLS_CERTDIR` env var, which is the documented GitLab DinD-TLS
pattern (`https://docs.gitlab.com/ee/ci/docker/docker_in_docker.html`).
Mounting `/certs/client` (not the broader `/certs`) keeps the client
container's view of the cert set minimal; the daemon needs the full
set, but the client only needs the client cert + CA.

**Alternatives considered:**

- *Self-signed sidecar mounted into both jobs.* More brittle — runner
  images differ in OpenSSL behaviour and the cert must be regenerated
  per cert rotation. Rejected.
- *Buildah / Kaniko.* Daemon-free, but registry auth via AWS-style
  config and cache layer behaviour differ across registries; revisit
  only if TLS DinD runs into second-order issues. Rejected for this
  change.

### 2. Release-tag reuse gate: GitLab API preflight + `release.replace_existing_artifacts: false`

**Choice:** The `release` job's `before_script` adds a GitLab Releases
API preflight that fails when a release for the current tag already
exists. `.goreleaser.yml` keeps `release.mode: replace` (the release-notes
merge mode is unrelated to artifact overwrite) and gains
`release.replace_existing_artifacts: false` so goreleaser never silently
overwrites artifacts when invoked. The goreleaser `--fail-if-tag-exists`
flag proposed in the original draft does not exist in v2.18.2
(verified against `goreleaser/goreleaser` source at tag `v2.18.2`;
the only `--fail-*` flag is `--fail-fast` with different semantics);
the API preflight replaces it.

**Rationale:** Two goreleaser features were conflated in the original
draft. `release.mode` (`keep-existing | append | prepend | replace`) is
typed as `ReleaseNotesMode` and governs only how release notes merge
on re-tag; it does not affect artifact upload. Artifact overwrite
protection is the separate `ReplaceExistingArtifacts` field. The
preflight gates the case the user actually cares about (don't re-publish
to a tag that already has a release), and `replace_existing_artifacts:
false` is the goreleaser-native backstop for the case where the
preflight is bypassed (manual re-run of the `release` job).

**Alternatives considered:**

- *Use only `release.mode: append`.* Rejected: `append` governs release
  notes, not artifact overwrite. Implementation matches the proposal's
  stated intent only if both fields are set.
- *Use only the API preflight.* Considered; keeping
  `replace_existing_artifacts: false` provides defense-in-depth for the
  case where the API preflight is missing or skipped.
- *Use goreleaser `--fail-if-tag-exists`.* Rejected: the flag does not
  exist in v2.18.2.

### 3. cosign signing: keyless via GitLab OIDC, image-bundled binary

**Choice:** Use `cosign sign --yes` with no explicit key, relying on
the runner's GitLab OIDC token (`SIGSTORE_ID_TOKEN` is exposed via
`id_tokens` in job spec). The release job uses `goreleaser/goreleaser:v2.18.2`,
which bundles `cosign v3.1.3` at `/usr/bin/cosign`; no per-run install
is required. Fall back to `$COSIGN_KEY` only if OIDC is unavailable
on the runner fleet, via a `before_script` switch.

**Rationale:** Keyless signing is the upstream-recommended path for
short-lived CI jobs. GitLab.com shared runners expose OIDC; the
project already relies on GitLab for primary distribution, so the
key transparency log (Rekor) covers the same audience. The mise
`[tools]` pin uses the `github.com/sigstore/cosign/v3/cmd/cosign`
Go module path so local `cosign sign` matches the CI version
exactly. Using the image-bundled binary avoids `go install` overhead
on every release.

**Alternatives considered:**

- *Static `COSIGN_KEY` secret.* Lower complexity, but rotates
  manually, and revocation is not transparent. Use only if the
  runner OIDC is unavailable. Documented as fallback in `tasks.md`.
- *SLSA provenance alongside cosign.* Out of scope (B14 deferred).
  Rejected for this change.
- *Pin mise cosign to v2.* Rejected: would cause local-vs-CI
  version drift between developer signing and CI signing.

### 4. Single Go version from `go.mod` directive

**Choice:** The `test` job runs the toolchain version declared by the
`go.mod` `go` directive (currently `1.27.1`). No version matrix.

**Rationale:** twiggit is an n=1 CLI with pinned dependencies. AGENTS.md
records no n-1 compatibility policy. A two-entry matrix doubles CI
minutes per MR for a hypothetical cross-version compat regression that
the pinned toolchain (`mise` owns `go`, `golangci-lint`, `gopls`,
`goreleaser`, `govulncheck`, `cosign`) does not exercise in practice.
The toolchain set in `.mise/config.toml` already pins the same version
for local + CI; cross-version compat is a property of the dependency
graph, not of the language runtime.

**Alternatives considered:**

- *Two-entry matrix (directive + n-1 latest).* Doubles CI minutes for
  signal that does not exist. Rejected.
- *Three-entry matrix (directive + n-1 latest + n-1 minimum).* Doubles
  again. Rejected.

### 5. `-race` folded into the main `test` job

**Choice:** The `test` job runs `go test -race ./...` (a single `-race`
invocation covering all packages). The `mise run ci:coverage` task on
tagged releases provides the canonical coverage-and-race signal.

**Rationale:** A standalone `test:race` job duplicates what the main
`test` job can do with a flag. A fast `test:race:subset` covering
`internal/`+`cmd/` adds MR-gate cost without catching races the
fold-into-test pass does not already catch; the novel `-shuffle=on`
flag was paying for order-dep detection that the deterministic local
`mise run test:race` would have surfaced earlier. Tagged-release
coverage remains the source of truth.

**Alternatives considered:**

- *Standalone `test:race` + `test:race:subset`.* Doubled race-related
  MR cost with redundant coverage. Rejected.
- *Race only on tags, not on MRs.* Loses the MR-gate race signal.
  Rejected.

### 6. Mirror push uses plain `git push --follow-tags`, never `--force`

**Choice:** `mirror-to-github` removes `--force`. Main-branch sync
runs plain `git push github HEAD:$CI_COMMIT_BRANCH --tags`; tag
mirror runs `git push github $CI_COMMIT_TAG --follow-tags` only on
the explicit tag trigger (already `when: never`'d above).

**Rationale:** A non-forced push catches the destructive failure mode
that caused the original finding: a co-tenant or stale credential
cannot overwrite GitHub history. The trade-off is that a force-with-lease
correction now requires a manual local rebase followed by a non-forced
push.

**Alternatives considered:**

- *`git push --force-with-lease`.* Still allows the destructive
  overwrite on the local-side race condition; rejected.
- *Two-step push (mirror from a designated bot fork).* Cleaner
  historical invariant but adds a second remote. Not justified by the
  threat model. Rejected.

### 7. cmdutil composition-root migration: functional options + consumer-side interface

**Choice:** `cmdutil.NewFactory(opts ...FactoryOption) *Factory` accepts
three functional options (`WithVersion(string)`,
`WithConfigLoader(func() (*core.Config, error))`,
`WithGitClientFactory(func() (interface{}, error))`). The `Factory.GitClient`
field type changes from `*git.Client` to `interface{}` — `*git.Client`
satisfies it implicitly via Go's structural interface satisfaction, so
`cmdutil` no longer imports `internal/git`. `main.go` (composition root)
supplies the three options; `cmd/root.go` is unchanged and continues to
receive `*cmdutil.Factory` from `main.go`. The depguard narrow (`.golangci.yml`
`cmdutil:` `allow` drops `internal/{git,config,version,output}`) lands
after the refactor.

**Constraints** (normative; per project `openspec/config.yaml`
`rules.design` "Use SHALL/SHOULD for architectural constraints"):

- `internal/cmdutil` SHALL NOT import `internal/{git,config,version}` after
  this change lands.
- `main.go` SHALL be the sole composition root that constructs
  `config.NewManager()`, `git.NewClient()`, and reads `version.Version`.
- `cmdutil.Factory` SHALL expose only lazy function fields; no constructor
  may be invoked eagerly inside `cmdutil`.
- The default no-arg `cmdutil.NewFactory()` SHALL continue to compile
  for test literals in `factory_test.go`.
- `cmd/root.go` SHALL NOT gain imports of `internal/{git,config,version}`
  as a side effect of this change.

**Rationale:** The `golang-cli` Tier 2 convention locates composition-root
wiring in `main.go`, not `cmd/` — `cmd/root.go` assembles the Cobra tree
and receives a fully constructed `*cmdutil.Factory`. `golang-refactoring/structural.md
§1` recommends consumer-side interfaces for breaking import cycles:
declaring `interface{}` (or any small interface the consumer uses) in
`cmdutil` lets `*git.Client` satisfy it implicitly with zero changes at
any call site, because every Factory consumer that needs methods on the
composite narrows via the per-role fields (`RepoOpener`, `BranchReader`,
...) which already type as `core.*` interfaces. Functional options scale
as Factory deps grow (matching `golang-design-patterns`
functional-options-as-preferred) without breaking the existing test
seam: the default `NewFactory()` returns a Factory with the same eager
fields as today plus lazy fields that return a sentinel error if invoked
without options, so every existing `factory_test.go` literal keeps
compiling. This is the single code change in this PR; everything else
is config / CI YAML / Docker / linter config.

**Alternatives considered:**

- *Explicit constructor params* (`NewFactory(loadConfig, newGitClient, appVersion)`).
  Clearer for three deps; rejected because it forces every
  `factory_test.go` literal to thread real loaders, removing the test
  seam that `golang-cli` and `cmd/AGENTS.md` both rely on.
- *Move wiring into `cmd/root.go`.* Rejected per the `golang-cli` Tier 2
  composition-root principle (main.go owns wiring) and because `cmd/`
  would then need to import `internal/{git,config,version}`, broadening
  the cmd depguard allow-list and violating the constraint above.
- *Keep `*git.Client` field type.* Rejected — the depguard narrow fails
  because the `internal/git` import stays in `cmdutil/factory.go`.
- *Skip the depguard narrow.* Defers a known rule violation; rejected.
- *Land the depguard narrow without the refactor.* CI lint fails; rejected.
- *Move the offending implementations out of `cmdutil` (deeper refactor).*
  Best long-term fix, but the functional-options refactor already aligns
  `cmdutil` with the `golang-cli` skill contract (Factory + exit codes +
  persistent flags + consumer interfaces) without restructuring ownership
  of `config`/`git`/`version`. Rejected for this change; revisit if a
  larger `cmdutil` split is proposed later.

### 8. Pinned goreleaser image matches mise pin

**Choice:** `goreleaser/goreleaser:v2.18.2` everywhere it appears in
`.gitlab-ci.yml` (`goreleaser-dry-run` job, `release` job). The image
tag is updated only when `.mise/config.toml` `[tools] goreleaser`
moves; the `build-ci-image` trigger covers `.mise/config.toml` and
`.goreleaser.yml` so a goreleaser-version bump rebuilds the image
before downstream jobs run.

**Rationale:** The CI image and the goreleaser runtime share the
goreleaser contract. A drift between the two would cause the runtime
to disagree with the mise-pinned version, masking configuration
errors. The coupled `build-ci-image` trigger eliminates the drift
class at the image-rebuild step, with no per-job drift check needed.

**Alternatives considered:**

- *Always pull `goreleaser/goreleaser:latest` and validate via
  `goreleaser --version`.* Hides drift between mise pin and CI image
  for one cycle. Rejected.
- *Use the CI image's goreleaser.* Requires the CI image to embed
  goreleaser; current image does not. Not justified by gain.
- *Per-job `before_script` drift check* (parse `.mise/config.toml`,
  compare to `goreleaser --version`, fail on mismatch). Drift path is
  closed upstream by the `build-ci-image` trigger; per-job defense is
  redundant. Rejected.

## Risks / Trade-offs

**[R1] Tag re-use without artifact overwrite.** → Mitigate via the GitLab
Releases API preflight (fails the `release` job before goreleaser runs) and
goreleaser's `replace_existing_artifacts: false` field as a backstop. Local
`mise run release:tag` already exits 1 on existing tags (verified at
`.mise/tasks/release/tag` lines 32-35); no local change required. The
Homebrew cask pipe is independent of `release.mode` (verified against
goreleaser source: `internal/pipe/brew/brew.go` does not reference
`release.mode`); no change needed.

**[R2] TLS DinD cert distribution depends on the runner image behaviour.** →
The `dockerd-entrypoint.sh` in the official `docker` image writes the cert
set to `/certs` automatically when `DOCKER_TLS_CERTDIR` is set. Both
GitLab.com shared runners and self-hosted runners run the same image, so
behaviour is stable. If a self-hosted runner pins a different dockerd
that doesn't honour the env var, the fallback is a sidecar-issued cert
mounted into both jobs; not expected on the current runner fleet.

**[R3] cosign keyless depends on runner OIDC availability.** → GitLab.com shared runners expose OIDC; the CI YAML registers `id_tokens` for the relevant job. The release job uses the `cosign v3.1.3` binary bundled in the goreleaser image (no per-run install). If a self-hosted runner rejects OIDC, the fallback uses `COSIGN_KEY` (project CI variable). The fallback is a `before_script` switch, not a second pipeline.

**[R4] govulncheck on the `lint` job adds a build-time dependency scan.** → `govulncheck` is already in the CI image via `mise` (toolchain governance). One `govulncheck ./...` per MR + per tag; ~10s on the current cache. Real signal: vulnerabilities in called code from twiggit's actual dependencies (charm.land, samber/lo, go-git, etc.). The previously-considered Trivy image scan produced only false positives from vendored test fixtures and toolchain-binary CVEs; the source-code analysis is the right scope.

**[R5] cosign-signed SBOM verification breaks if the Rekor transparency log rejects an entry.** → Verified blobs are still downloadable; signing fails-soft with an explicit log line. The release job does not gate on cosign success in v1 (matches the spec's loose behaviour); v2 can tighten via a follow-up.

**[R6] Goreleaser image version vs mise pin drift.** → The CI image
(`goreleaser/goreleaser:v2.18.2`) and the mise pin (`goreleaser = "2.18.2"`)
are coupled by the `build-ci-image` trigger covering `.mise/config.toml`
and `.goreleaser.yml`. A goreleaser bump in either file rebuilds the
image before downstream jobs run; no per-job drift check is required.

## Open Questions

- **`homebrew_casks.token` rotation cadence** is not in scope (B13
  deferred) but the release job still uses `GITHUB_TOKEN` from the
  GitLab CI variable. If the token rotates mid-pipeline, the cask
  push fails. Reasonable behaviour for now; document as a known
  limitation.

- **`release:tag` task sub-invocation** (`.mise/config.toml` line 143
  references `release:tag` and `release:prepare`, neither of which is
  defined in the inline `[tasks]` block — both live as file-based
  scripts in `.mise/tasks/release/`). The change does not fix this,
  but the existing `release:tag` script already exits 1 on existing
  tags (line 32-35), so the local-side guarantee is intact. Flag for
  follow-up.
