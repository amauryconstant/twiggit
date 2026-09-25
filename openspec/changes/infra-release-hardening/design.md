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
the existing project standard (mise owns tools; `cosign` and `trivy` are
provisioned the same way `govulncheck` is — via the mise `[tools]`
block). All choices below assume that governance is unchanged.

## Goals / Non-Goals

**Goals:**

- Close REVIEW-B critical findings (B4, B5, B6) and most warn findings
  (B7–B12, B15–B23) with minimal CI-minute expansion.
- Preserve the existing `infrastructure-toolchain` contract: mise owns
  tools, `mise run verify` order is unchanged, govulncheck stays
  developer-local + pre-commit.
- Keep the change's risk surface localized to `.gitlab-ci.yml`,
  `.goreleaser.yml`, and the Docker lint files. `cmd/` and `internal/`
  must remain untouched.

**Non-Goals:**

- Realize the B1, B2, B3, B13, B14 deferred items (govulncheck CI,
  CodeQL SARIF, Dependabot/Renovate, homebrew token scope, SLSA
  provenance). Those are explicitly deferred per the proposal's
  Non-Goals section.
- Cut over to Buildah/Kaniko or another daemon-free builder; TLS DinD
  is the chosen path. Buildah is revisited only if (B6) reveals a
  second-order runtime problem.
- Modify `cmd/`, `internal/`, or any consumer of git operations.
- Change the `infra-release` Homebrew tap configuration beyond what
  `release.mode: append` requires.

## Decisions

### 1. TLS DinD uses `docker:dind:tls` service image, not custom cert sidecar

**Choice:** Replace `docker:29.1.4-dind` + `DOCKER_TLS_CERTDIR: ""`
with `docker:29.1.4-dind:tls`, leave `DOCKER_TLS_CERTDIR` at its
default (`/certs`), switch `DOCKER_HOST` to `tcp://docker:2376`, and
mount `/certs` from the service into the client.

**Rationale:** The TLS-enabled image is provided as a first-party
`docker:dind:tls` tag; it writes the CA, server cert, and client cert
into `/certs` and configures both sides to talk over TLS without any
custom cert generation. The setup is deterministic across GitLab.com
shared runners and self-hosted runners.

**Alternatives considered:**

- *Self-signed sidecar mounted into both jobs.* More brittle — runner
  images differ in OpenSSL behaviour and the cert must be regenerated
  per cert rotation. Rejected.
- *Buildah / Kaniko.* Daemon-free, but registry auth via AWS-style
  config and cache layer behaviour differ across registries; revisit
  only if TLS DinD runs into second-order issues. Rejected for this
  change.

### 2. Release mode guard: `--fail-if-tag-exists` + `release.mode: append`

**Choice:** `.goreleaser.yml:75` switches from `release.mode: replace`
to `release.mode: append`; `goreleaser release` in `.gitlab-ci.yml`
gains `--fail-if-tag-exists`.

**Rationale:** `--fail-if-tag-exists` is goreleaser's native guard
that fails the job when the tag is already published. Combined with
`append`, artifacts from a previous successful release are preserved
when a re-tag attempt happens. The two flags together express "tag
reuse must be a deliberate operation" without requiring CI to call
the GitLab Releases API pre-flight.

**Alternatives considered:**

- *Pre-flight `curl -sI` against the GitLab Releases API in
  `before_script`.* Slower and inconsistent with goreleaser's own
  semantics; rejected.
- *Keep `mode: replace` + add API guard.* Same as pre-flight but
  preserves the overwrite semantics; the proposal's analysis already
  concluded overwrite is unsafe. Rejected.

### 3. cosign signing: keyless via GitLab OIDC

**Choice:** Use `cosign sign --yes` with no explicit key, relying on
the runner's GitLab OIDC token (`SIGSTORE_ID_TOKEN` is exposed via
`id_tokens` in job spec). Fall back to `COSIGN_KEY` only if OIDC is
unavailable on the runner fleet.

**Rationale:** Keyless signing is the upstream-recommended path for
short-lived CI jobs. GitLab.com shared runners expose OIDC; the
project already relies on GitLab for primary distribution, so the
key transparency log (Rekor) covers the same audience.

**Alternatives considered:**

- *Static `COSIGN_KEY` secret.* Lower complexity, but rotates
  manually, and revocation is not transparent. Use only if the
  runner OIDC is unavailable. Documented as fallback in `tasks.md`.
- *SLSA provenance alongside cosign.* Out of scope (B14 deferred).
  Rejected for this change.

### 4. Go-version matrix keyed off `go.mod` directive + previous minor's latest patch

**Choice:** A two-entry matrix, populated by `.gitlab-ci.yml`
`variables:`. Entry 1 reads the `go` directive from `go.mod`; Entry 2
is the previous minor's latest patch. Matrix uses `fail-fast: false`
and a single, shared `script:`.

**Rationale:** Two entries cover the project-stated toolchain set
without doubling CI minutes proportionally. Reading the directive
from `go.mod` keeps the matrix in lockstep with the actual code; the
prior minor is the realistic floor for `golang.org/x/...` dependency
compatibility.

**Alternatives considered:**

- *Three-entry matrix (directive + n-1 latest + n-1 minimum).*
  Marginal benefit, doubles again. Rejected.
- *Single Go version (the directive only).* Insufficient — losing
  the cross-version coverage that the matrix is supposed to provide.
  Rejected.

### 5. Race jobs: standalone full matrix + fast subset, both `-race`

**Choice:** Two distinct jobs:
`test:race` runs `go test -race ./...` in a matrix cell;
`test:race:subset` runs `go test -race ./internal/... ./cmd/...` as a
faster pre-merge check. The tagged-release `coverage` job remains
the source of truth.

**Rationale:** The subset gives a fast race pass on every MR; the
full `-race` run on tags catches the longer path. The duplication is
intentional — the spec specifies both because the failure modes are
different (subset-jobs fail fast on common races; tagged release
fails on the rare races only the full run can detect).

**Alternatives considered:**

- *Single full `-race` job, no subset.* Faster to set up, but a long
  blocking wait on every MR. Rejected.
- *Subset-only on MRs.* Covers common cases but misses rare races.
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

### 7. Linter depguard stays narrow on `cmdutil` per the architectural contract

**Choice:** `.golangci.yml` `cmdutil:` `allow` drops entries for
`internal/git`, `internal/config`, `internal/output`, and
`internal/version`. Only `$gostd`, `lo`, `cobra`, `core`,
`iostreams` are imported by `internal/cmdutil/`. The composition
layer enforces imports via `cmd/`.

**Rationale:** This restores the depguard rule that broke when
`internal/output/table.go` was added (now wired through `iostreams`
+ `output`). The narrow allow-list matches the skill's contract.

**Alternatives considered:**

- *Keep the broad allow-list and revisit at end of the E change.*
  Defers a known rule violation; rejected.
- *Move the offending implementations out of `cmdutil`.*
  Best long-term fix, but those packages belong in `output`/`iostreams`
  already; rewiring them is a separate behavioural change. Not this
  change.

### 8. Pinned goreleaser image matches mise pin

**Choice:** `goreleaser/goreleaser:v2.18.2` everywhere it appears in
`.gitlab-ci.yml` (`goreleaser-dry-run` job, `release` job). The image
tag is updated only when `.mise/config.toml` `[tools] goreleaser`
moves, and the trigger for `build-ci-image` covers that case (B20).

**Rationale:** The CI image and the goreleaser runtime share the
goreleaser contract. A drift between the two would cause the
`goreleaser --version` check in the CI Dockerfile to disagree with
the runtime, masking configuration errors. Pinning both at the same
version with a coupled trigger (B20) eliminates the drift class.

**Alternatives considered:**

- *Always pull `goreleaser/goreleaser:latest` and validate via
  `goreleaser --version`.* Hides drift between mise pin and CI image
  for one cycle. Rejected.
- *Use the CI image's goreleaser.* Requires the CI image to embed
  goreleaser; current image does not. Not justified by gain.

## Risks / Trade-offs

**[R1] `release.mode: append` adds artifact folders to existing GitLab releases on tag re-use.** → Mitigate via `--fail-if-tag-exists` so re-use never silently appends; document in `tasks.md` that `mise run release:tag` must surface "tag already published" as a hard failure. Homebrew tap does not need to change because the re-tagged release would fail before goreleaser pushes the cask update.

**[R2] TLS DinD cert distribution depends on the runner image behaviour.** → The first-party `docker:dind:tls` image has stable behaviour across GitLab.com and self-hosted runners; if the project's runner fleet does not include `docker:dind:tls`, fall back to a sidecar-issued cert mounted into both jobs and document the workflow.

**[R3] cosign keyless depends on runner OIDC availability.** → GitLab.com shared runners expose OIDC; the CI YAML registers `id_tokens` for the relevant job. If a self-hosted runner rejects OIDC, the fallback uses `COSIGN_KEY` (project CI variable). The fallback is a `before_script` switch, not a second pipeline.

**[R4] Go-version matrix roughly doubles validate-stage CI minutes per MR.** → Acceptable cost for cross-version coverage; the matrix is two entries (`directive + n-1 latest`). For self-hosted runners, the matrix is configurable via `variables:` and can be reduced to one entry if cost is the priority.

**[R5] Race subset misses long-running packages outside `internal/`+`cmd/`.** → The full `-race` run lives on the tagged-release job; the subset is explicitly an MR-gate fast pass, not the canonical signal. If the subset scope changes, the spec's "fast race-detector subset" requirement re-evaluates; see `Open Questions`.

**[R6] cosign-signed SBOM verification breaks if the Rekor transparency log rejects an entry.** → Verified blobs are still downloadable; signing fails-soft with an explicit log line. The release job does not gate on cosign success in v1 (matches the spec's loose behaviour); v2 can tighten via a follow-up.

**[R7] B7 + B18 (`go mod tidy` gate) double-runs `go mod tidy` in the pipeline.** → Pipeline cost is one `go mod tidy` call per gate, total ~10s on the current cache. Acceptable; the gate catches a class of bugs (`go.mod` drift between local and CI) the test job cannot.

## Open Questions

- **Race-subset scope** (`./internal/...` + `./cmd/...`): does the
  project's test organization keep all goroutine-leaking code under
  those paths? If a future test under `test/integration/` is added
  that needs `-race`, the subset requirement changes. The current
  scope matches `cmd/AGENTS.md`'s e2e-only convention, so the risk is
  low and reviewable on the first integration test addition.

- **`homebrew_casks.token` rotation cadence** is not in scope (B13
  deferred) but the release job still uses `GITHUB_TOKEN` from the
  GitLab CI variable. If the token rotates mid-pipeline, the cask
  push fails. Reasonable behaviour for now; document as a known
  limitation.
