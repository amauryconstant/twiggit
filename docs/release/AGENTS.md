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
v2.18.3 still ships 1.27.1) because goreleaser's pre-baked `.apk` is
compiled against the image's pinned Go. Using that image as the release
job container while go.mod's `go` directive is newer is a hard trap. The
release job therefore runs on `golang:1.27.2-alpine3.23` (matching the
`Dockerfile.ci` base) with goreleaser v2.18.2, syft v1.51.1, and cosign
v3.1.3 installed inside the job's `before_script`. This pattern applies
symmetrically to the GitLab CI `release` and `goreleaser-dry-run` jobs and
to the GitHub Actions `release` job.

When bumping go.mod's `go` directive, bump the release job image base in
the same change. Do not reintroduce the goreleaser-pinned image as a
release runner.

**Invariant:** Release jobs install goreleaser and cosign from the
upstream prebuilt release artifacts via `.mise/tasks/ci/install-release-tools`,
not via `go install`. The `.gitlab-ci.yml` global `variables:` block sets
`GOPATH: "$CI_PROJECT_DIR/.go"`, so a bare `go install` writes to
`$CI_PROJECT_DIR/.go/bin` — a directory not in `$PATH` — and the install
"succeeds" while the next line fails with `goreleaser: not found`. `go install`
has no `-o` flag (only `go build` does), and `GOBIN` only redirects within the
same trap if the destination also isn't on `$PATH`. Prebuilt is strictly
better: faster (~3 min saved per run by skipping the source compile), uses
the same artifact family as goreleaser's own Docker image, and can be
SHA-pinned against the upstream `checksums.txt` / `cosign_checksums.txt`.
Current pins: goreleaser `v2.18.2` SHA
`0a96edc9d9bc594e4a41cc4d59467c182062910ab24d9d1f6dd7b667d32606d3`, cosign
`v3.1.3` SHA `4629c757b7618056f8ddd7e2625ae9fdd94c0372a65049520bc7d9df9efc7f71`.
Bump the version and SHA in the same edit. The script is invoked from both
`.gitlab-ci.yml` (via the `*install_release_tools` YAML anchor, shared by
`goreleaser-dry-run` and `release`) and `.github/workflows/release.yml`.

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
