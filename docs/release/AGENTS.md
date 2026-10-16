# Release

`twiggit` ships to GitLab Releases (canonical) and GitHub Releases
(mirror). Tag pushes route through GitLab; the GitHub mirror is a
non-forced `git push --tags --force-with-lease`.

## Commands

| Command                  | Purpose               |
| ------------------------ | --------------------- |
| mise run release:check   | Release prerequisites |
| mise run release:dry-run | Test GoReleaser       |

## Distribution surface ownership

Split enforced at the goreleaser config layer:

| Surface              | CI owner                                   | Config                                           | What it publishes                                                                    |
| -------------------- | ------------------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------------------ |
| GitLab release       | GitLab CI `release` job (`.gitlab-ci.yml`) | `.goreleaser.yml` (`force_token: gitlab`)        | Tarballs, checksums, SBOMs to `gitlab.com/amoconst/twiggit/-/releases/v<tag>`        |
| GitHub release       | GitHub Actions `release.yml`               | `.goreleaser.github.yml` (`force_token: github`) | Tarballs, checksums, SBOMs to `github.com/amauryconstant/twiggit/releases/tag/<tag>` |
| Tag mirror to GitHub | GitLab CI `mirror-to-github` job           | `.gitlab-ci.yml`                                 | `git push --tags --force-with-lease` (tag pipelines: `when: never`)                  |

**Invariant:** The release job image MUST run a Go version that satisfies
`go.mod`'s `go` directive (currently `go 1.27.2`). Goreleaser pins
`GOTOOLCHAIN=local` for its child `go build` invocations, which disables
Go's toolchain auto-download — so the host Go has to already satisfy
go.mod or the `before:` hook fails with
`go: go.mod requires go >= 1.27.2 (running go 1.27.1; GOTOOLCHAIN=local)`.

The `goreleaser/goreleaser:v2.18.2` Docker image ships Go 1.27.1 (and
v2.18.3 / v2.19.0-nightly still ship 1.27.1) because goreleaser's pre-baked
`.apk` is compiled against the image's pinned Go. Using that image as the
release job container while go.mod's `go` directive is newer is a hard
trap. The release job therefore runs on a self-owned container image,
sourced differently per CI surface, that always satisfies go.mod's Go
directive.

- **GitLab CI** (`release`, `goreleaser-dry-run` in `.gitlab-ci.yml`)
  runs on `golang:1.27.2-alpine3.23` (matching `Dockerfile.ci`) and
  installs goreleaser/cosign/syft at job time via the
  `*install_release_tools` YAML anchor —
  `.mise/tasks/ci/install-release-tools.sh`. This is the unchanged
  runtime-installer pattern that worked at v0.15.x.
- **GitHub Actions** (`release` in `.github/workflows/release.yml`) runs
  on a self-owned image, built by the reusable workflow
  `.github/workflows/release-image.yml` from `Dockerfile.release`. That
  Dockerfile is a fork of the goreleaser project's released-image
  Dockerfile with `FROM golang:1.27.2-alpine3.23` (matching the Go
  requirement) and goreleaser/cosign/syft SHA-pinned at build time. This
  restores the bundled-image architecture that shipped v0.13.7 →
  v0.15.1 reliably on GitHub.

  The `release.yml` workflow calls `release-image.yml` as its first job
  (`uses: ./.github/workflows/release-image.yml`) and waits on it via
  `needs:`. The image workflow computes a content hash of
  `Dockerfile.release` (`sha256sum | cut -c1-16`), tags every build with
  `ghcr.io/amauryconstant/twiggit/release:src-<hash>`, and on a call
  where `:src-<hash>` already exists (i.e. the Dockerfile has not
  changed) it reuses the existing image and returns its digest without
  rebuilding — `force: true` (workflow_call input or workflow_dispatch
  checkbox) bypasses the skip. The release job then pulls the image by
  digest (`ghcr.io/amauryconstant/twiggit/release@<digest>`), which
  makes the release race-free against any parallel path-filter rebuild
  on main. Two `concurrency: cancel-in-progress: false` groups
  (`release-image`, `release`) serialize same-class rebuilds/releases so
  a v0.16.5 release already in flight is not cancelled by a v0.16.6
  push.

The two surfaces diverge on purpose: from v0.16.0 through v0.16.4 the
GitHub release job cascaded through five separate failures, each
introduced by the runtime-installer pattern (Go-version mismatch →
`go install` PATH trap → busybox bash shebang → no-git REST-API fallback
→ install-before-checkout chicken-and-egg). GitLab never exhibited the
cascade because its runner's git-aware helper populates the workspace
before `before_script` runs.

When bumping go.mod's `go` directive, bump both release surfaces in the
same change: `.gitlab-ci.yml` `release.job.image.name` (and the
`*release` / `goreleaser-dry-run` job-level `image:` lines) and
`Dockerfile.release` `FROM golang:…`. Do not reintroduce the
goreleaser-pinned image as a release runner.

**Invariant:** On GitHub Actions the release image ships `git` already,
so the checkout/install-order constraint from the v0.16.x cascade no
longer applies. If you switch the GitHub `release` job back to a
sparse-base image, restore the install-before-checkout ordering AND add
`fetch-tags: true` to `actions/checkout@v5`, or you re-introduce the
"current folder is not a git repository" failure.

**Invariant:** Release jobs install goreleaser, cosign, and syft from the
upstream prebuilt release artifacts via `.mise/tasks/ci/install-release-tools.sh`,
not via `go install`. The `.gitlab-ci.yml` global `variables:` block sets
`GOPATH: "$CI_PROJECT_DIR/.go"`, so a bare `go install` writes to
`$CI_PROJECT_DIR/.go/bin` — a directory not in `$PATH` — and the install
"succeeds" while the next line fails with `goreleaser: not found`. `go install`
has no `-o` flag (only `go build` does), and `GOBIN` only redirects within the
same trap if the destination also isn't on `$PATH`. Prebuilt is strictly
better: faster (~3 min saved per run by skipping the source compile), uses
the same artifact family as goreleaser's own Docker image, and can be
SHA-pinned against the upstream `checksums.txt` / `cosign_checksums.txt`.

Tool-version SHA pins live in two places and must move together:

| Tool       | GitLab CI (`install-release-tools.sh`)          | GitHub Actions (`Dockerfile.release`)                           |
| ---------- | ----------------------------------------------- | --------------------------------------------------------------- |
| goreleaser | `v2.18.2` SHA `0a96e…06d3`                      | `v2.18.2` SHA `0a96e…06d3` (`ARG GR_SHA`)                       |
| cosign     | `v3.1.3` SHA `4629c…f71`                        | `FROM gcr.io/projectsigstore/cosign:v3.1.3@sha256:9e5c2f…611b8` |
| syft       | `v1.51.1` (upstream install script, no SHA pin) | `FROM anchore/syft:v1.54.1@sha256:3eb53…9b1aa`                  |

The two pin sets are not identical today (syft `1.51.1` vs `1.54.1`); align
them toward one version per tool in a follow-up change. The GitLab CI
script is invoked via `.gitlab-ci.yml`'s `*install_release_tools` YAML
anchor (shared by `goreleaser-dry-run` and `release`); the Dockerfile is
invoked by `.github/workflows/release.yml`'s `build-image` job via
`uses: ./.github/workflows/release-image.yml`.

**CHANGELOG**: Auto-generated via `openspec-generate-changelog` after
archiving changes.

## Release verification

Each release publishes:

- `twiggit_<VERSION>_<OS>_<ARCH>.<ext>` archives (tar.gz / zip)
- `twiggit_<VERSION>_<OS>_<ARCH>_sbom.spdx.json` SPDX SBOMs (one per archive)
- `twiggit_<VERSION>_checksums.txt` containing sha256 of every archive and SBOM

The checksum file is signed by `cosign` keyless via OIDC (GitLab OIDC on
the GitLab release, GitHub Actions OIDC on the GitHub release;
image-bundled `cosign v3.1.3`). cosign v3 writes a single
`checksums.txt.sigstore.json` bundle (signature + cert + Rekor inclusion)
in `dist/` for debugging — it is NOT published as a release asset. Per-SBOM
signatures are intentionally omitted: every SBOM hash is already in the
signed checksum file, so a single signature transitively covers all
artifacts.

Verify the checksum file against the release's OIDC identity, then verify
every artifact against the signed checksum file:

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
# checksum file. This proves every artifact matches what was published.
sha256sum -c --ignore-missing twiggit_<VERSION>_checksums.txt
```

For OIDC-keyless verification cosign uses the ambient identity from the
TUF/Rekor transparency log; no public-key fingerprint is needed because
the signing cert is short-lived and rooted in Fulcio.
