# Twiggit - OpenCode Reference

Pragmatic git worktree CLI. Functional-core + I/O-adapter layout.

## Architecture

| Package               | Role                                                                                                    |
| --------------------- | ------------------------------------------------------------------------------------------------------- |
| `internal/core/`      | Pure value objects, errors, validation. No I/O.                                                         |
| `internal/git/`       | I/O adapter: client, writer, hook runner, context resolver, `status.go` per-worktree diagnostic reader. |
| `internal/config/`    | koanf config loader.                                                                                    |
| `internal/output/`    | Formatters + error renderer.                                                                            |
| `internal/iostreams/` | TTY, IOStreams, slog singleton.                                                                         |
| `internal/cmdutil/`   | Factory, ExitCodeFor, persistent flags.                                                                 |
| `cmd/` + `main.go`    | Cobra tree + composition root.                                                                          |

## Config knobs

| Knob                            | Default                                              | Source                  |
| ------------------------------- | ---------------------------------------------------- | ----------------------- |
| `validation.protected_branches` | `["main","master","develop","staging","production"]` | `git-config`            |
| `status.stale_behind`           | `20`                                                 | `[status] stale_behind` |
| `status.stale_days`             | `30`                                                 | `[status] stale_days`   |

Depguard owns the import rules (see `.golangci.yml`).

## Commands

| Task              | Command                                |
| ----------------- | -------------------------------------- |
| Verify everything | `mise run verify`                      |
| Run all tests     | `mise run test`                        |
| Run e2e           | `mise run test:e2e`                    |
| Lint              | `mise run lint:check`                  |
| Format            | `mise run format`                      |
| Lint markdown     | `mise run lint:md`                     |
| Format markdown   | `mise run format:md`                   |
| Build             | `mise run build`                       |
| Update goldens    | `UPDATE_GOLDEN=true mise run test:e2e` |

## CLI Surface

| Command   | Alias | Purpose                                                                 |
| --------- | ----- | ----------------------------------------------------------------------- |
| `list`    | `ls`  | List worktrees                                                          |
| `create`  | -     | New worktree + post-create hooks                                        |
| `delete`  | `rm`  | Remove worktree (+ optional branch)                                     |
| `prune`   | -     | Delete merged worktrees                                                 |
| `cd`      | -     | Print worktree path                                                     |
| `rebase`  | -     | Rebase current/named/all worktrees onto tracked base                    |
| `status`  | -     | Per-worktree diagnostic view (ahead/behind, base, merged, dirty, stale) |
| `sync`    | -     | Fetch remote tracking refs (optionally + rebase)                        |
| `init`    | -     | Shell wrapper                                                           |
| `version` | -     | Build version                                                           |

## Conventions

- **Bool prefix.** `is/has/can` on every bool field. `Was*` on past-tense bools (`WasDeleted`).
- **Single-handling.** Adapter returns wrapped errors only; logging at cmd boundary via `opts.IO.Logger`.
- **Sentinels.** Twelve `core.Err*` live in `internal/core/sentinels.go`. Match via `errors.Is`, never by `Op` string.
- **Factory roles.** No per-role lazy fields. Callers do `f.GitClient()` then assign to a role-typed local.
- **Tracked base.** Worktree's rebase target persists at `twiggit.tracked-base` per-wt git config (`git config --worktree`). Written by `wtg create`; mutated by `wtg rebase --set-base`.
- **Exit codes.** 0=ok, 1=error, 2=usage. `cmdutil.ExitCodeFor` is the single mapping.

## Release / CI invariants

The canonical rules live in the `infrastructure-release` spec. This block is the prose mirror; the spec is authoritative. Each key is grep-discoverable.

**Toolchain**

- **[release-go-directive]** Release job image SHALL run a Go version satisfying `go.mod`'s `go` directive. Goreleaser pins `GOTOOLCHAIN=local` for child `go build` invocations, disabling toolchain auto-download. A bump to `go.mod`'s `go` directive SHALL bump the release image (`.gitlab-ci.yml` `release.job.image.name` and `Dockerfile.release` `FROM golang:…`) in the same change. (infrastructure-release)
- **[release-tools-prebuilt]** Release tools (goreleaser, cosign, syft) SHALL be installed from upstream prebuilt artifacts SHA-pinned against `checksums.txt`, extracted to `/usr/local/bin`. `go install` SHALL NOT be used: it has no `-o` flag and the project-level `GOPATH` override traps the binary off-`$PATH`. (infrastructure-release)
- **[release-pinned-images]** CI tooling images (goreleaser, base images) SHALL be pinned to a specific tag, not `latest`. The pinned goreleaser version SHALL match `.mise/config.toml`. (infrastructure-release)

**Install scripts**

- **[release-alpine-sh]** Scripts that run inside Alpine release containers SHALL start with `#!/bin/sh` and SHALL NOT use `set -o pipefail` (bash/ksh extension). Busybox `tar` SHALL use `-f` explicitly. Bash MAY be installed by the script itself, but SHALL NOT be assumed at execution time. (infrastructure-release)
- **[release-install-before-checkout]** On GitHub Actions, when the release job runs on a sparse-base container, `git` SHALL be installed BEFORE `actions/checkout@v5`, which SHALL receive `fetch-tags: true`. GitLab CI is exempt — its runner populates the workspace before `before_script`. (infrastructure-release)

**GitHub Actions specifics**

- **[release-permission-superset]** A workflow that calls another via `uses:` SHALL declare a workflow-level `permissions:` block that is a superset of every scope the callee requests. Adding a scope to a reusable workflow SHALL add it to every caller in the same change. (infrastructure-release)
- **[release-setup-buildx]** A workflow that uses `docker/build-push-action` with `cache-to: type=gha` SHALL also configure `docker/setup-buildx-action` (docker-container driver) beforehand. The default docker driver rejects GHA cache export. (infrastructure-release)
- **[release-digest-reference]** When a release consumer depends on a registry image built earlier in the same workflow, the consumer SHALL pull by digest (`@sha256:…`), not by mutable tag. Digest reference eliminates the race against any parallel rebuild. (infrastructure-release)
- **[release-no-digest-with-tags]** `docker/build-push-action` SHALL NOT combine `outputs: type=image,push-by-digest=true` with named `tags:` inputs. Buildx refuses "tagged ref … by digest". (infrastructure-release)

**Preflights and module graph**

- **[release-preflight-enumerate]** Preflight checks against external APIs SHALL enumerate expected status codes (typically `200` for exists, `404` for absent). Any other code SHALL fail the job. Silent preflights are prohibited. (infrastructure-release)
- **[release-module-freshness]** The `lint` job SHALL run `go mod tidy && git diff --exit-code` before `golangci-lint run`. Unstaged `go.mod`/`go.sum` drift fails the job. (infrastructure-release)

**Security gates**

- **[release-govulncheck]** The `lint` job SHALL run `govulncheck ./...`. Unfixed vulnerabilities in called code exit the step non-zero and fail the job. (infrastructure-release)
- **[release-cosign-signing]** The `release` job SHALL sign SBOMs with cosign. OIDC keyless signing is preferred; the job SHALL fall back to `$COSIGN_KEY` when OIDC is unavailable. (infrastructure-release)
- **[release-tls-dind]** The `build-ci-image` job SHALL enable TLS on the Docker-in-Docker service via `DOCKER_TLS_CERTDIR: "/certs"`. Plaintext DinD on port 2375 is prohibited. (infrastructure-release)

**Mirror and release job**

- **[release-mirror-non-forced]** The `mirror-to-github` job SHALL push without `--force`. Non-fast-forward local branches SHALL fail rather than overwrite. (infrastructure-release)
- **[release-non-overwriting]** The `release` job SHALL verify the tag has no existing GitLab release via API preflight before invoking goreleaser. `.goreleaser.yml` SHALL set `release.replace_existing_artifacts: false`. (infrastructure-release)
- **[release-non-interruptible]** The `release` job SHALL set `interruptible: false` so a concurrent pipeline on the same ref cannot abort a publish mid-flight. (infrastructure-release)
- **[release-dry-run-path-filter]** The `goreleaser-dry-run` job SHALL run only on merge requests that touch release-affecting paths (`cmd/`, `internal/`, `.goreleaser.yml`, `go.mod`, `go.sum`). Snapshot artifacts SHALL expire within one day. (infrastructure-release)

**CI image and pipeline defaults**

- **[release-ci-image-triggers]** The `build-ci-image` job SHALL trigger on changes to `.goreleaser.yml` and `.mise/config.toml` so a goreleaser-version bump rebuilds the image before downstream jobs run. (infrastructure-release)
- **[release-pipeline-defaults]** The `.gitlab-ci.yml` `default:` block SHALL declare the retry policy (test failures not retried) and `interruptible: true` for every job unless overridden. (infrastructure-release)

**Architecture**

- **[release-surface-divergence]** Per-surface CI architecture (GitLab CI vs GitHub Actions) MAY diverge when a unified architecture has cascaded through three or more environmental failures on one surface. The divergence SHALL be documented in the Release process section below, naming the pattern used and the failure mode it avoids. (infrastructure-release)

## Release process

Operator-facing recipes — no invariants in this section.

**Distribution surface**

| Surface              | CI owner                                   | Config                                           | Publishes to                                                                         |
| -------------------- | ------------------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------------------ |
| GitLab release       | GitLab CI `release` job (`.gitlab-ci.yml`) | `.goreleaser.yml` (`force_token: gitlab`)        | Tarballs, checksums, SBOMs to `gitlab.com/amoconst/twiggit/-/releases/v<tag>`        |
| GitHub release       | GitHub Actions `release.yml`               | `.goreleaser.github.yml` (`force_token: github`) | Tarballs, checksums, SBOMs to `github.com/amauryconstant/twiggit/releases/tag/<tag>` |
| Tag mirror to GitHub | GitLab CI `mirror-to-github` job           | `.gitlab-ci.yml`                                 | `git push --tags` (tag pipelines: `when: never`)                                     |

**Commands**

| Command                    | Purpose               |
| -------------------------- | --------------------- |
| `mise run release:check`   | Release prerequisites |
| `mise run release:dry-run` | Test GoReleaser       |

**Verify a release**

Each release publishes `twiggit_<VERSION>_<OS>_<ARCH>.<ext>` archives, `twiggit_<VERSION>_<OS>_<ARCH>_sbom.spdx.json` SPDX SBOMs, and `twiggit_<VERSION>_checksums.txt` (sha256 of every archive and SBOM). The checksum file is signed by `cosign` keyless via OIDC. Step 1 verifies the checksum file against the release's OIDC identity; step 2 verifies every artifact against the signed checksum file.

```bash
# Step 1: verify checksums.txt was signed by this repo's release job.
# GitLab release (signed by GitLab CI release job)
cosign verify-blob \
  --certificate-identity 'https://gitlab.com/amoconst/twiggit//.gitlab-ci.yml@refs/tags/<TAG>' \
  --certificate-oidc-issuer 'https://gitlab.com' \
  twiggit_<VERSION>_checksums.txt

# GitHub release (signed by GitHub Actions release workflow)
cosign verify-blob \
  --certificate-identity 'https://github.com/amauryconstant/twiggit/.github/workflows/release.yml@refs/tags/<TAG>' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  twiggit_<VERSION>_checksums.txt

# Step 2: verify every downloaded archive + SBOM against the signed
# checksum file.
sha256sum -c --ignore-missing twiggit_<VERSION>_checksums.txt
```

**Per-surface architecture**

The two release surfaces intentionally diverge. GitLab CI uses a runtime installer pattern: the `release` and `goreleaser-dry-run` jobs run on `golang:1.27.2-alpine3.23` and install goreleaser/cosign/syft at job time via `.mise/tasks/ci/install-release-tools.sh`. GitHub Actions uses a self-owned bundled image: `.github/workflows/release-image.yml` builds `ghcr.io/amauryconstant/twiggit/release` from `Dockerfile.release` (a fork of the goreleaser project's image with `FROM golang:1.27.2-alpine3.23`) and the `release.yml` job pulls the image by digest. The divergence exists because the unified runtime-installer pattern cascaded through five environmental failures on GitHub Actions (Go version mismatch, `go install` PATH trap, busybox bash shebang, no-git REST-API fallback, install-before-checkout ordering) while remaining stable on GitLab CI. Per-surface divergence is permitted by `[release-surface-divergence]`.

## Subdirectory Guides

| File                                                     | Purpose                        |
| -------------------------------------------------------- | ------------------------------ |
| [cmd/AGENTS.md](cmd/AGENTS.md)                           | Cobra patterns, command tree   |
| [openspec/AGENTS.md](openspec/AGENTS.md)                 | OpenSpec workflow + spec rules |
| [test/AGENTS.md](test/AGENTS.md)                         | Test organization              |
| [test/integration/AGENTS.md](test/integration/AGENTS.md) | Testify suites                 |
| [test/e2e/AGENTS.md](test/e2e/AGENTS.md)                 | Ginkgo/Gomega CLI testing      |
| [test/concurrent/AGENTS.md](test/concurrent/AGENTS.md)   | Race detector patterns         |
| [internal/version/AGENTS.md](internal/version/AGENTS.md) | Build-time ldflags             |
