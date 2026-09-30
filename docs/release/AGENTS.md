# Release

`twiggit` ships to GitLab Releases (canonical) and GitHub Releases
(mirror). Tag pushes route through GitLab; the GitHub mirror is a
non-forced `git push --tags --force-with-lease`.

## Commands

| Command                   | Purpose          |
| ------------------------- | ---------------- |
| mise run release:check    | Release prerequisites |
| mise run release:dry-run  | Test GoReleaser  |

## Distribution surface ownership

Split enforced at the goreleaser config layer:

| Surface | CI owner | Config | What it publishes |
| ------- | -------- | ------ | ----------------- |
| GitLab release | GitLab CI `release` job (`.gitlab-ci.yml`) | `.goreleaser.yml` (`force_token: gitlab`) | Tarballs, checksums, SBOMs to `gitlab.com/amoconst/twiggit/-/releases/v<tag>` |
| GitHub release | GitHub Actions `release.yml` | `.goreleaser.github.yml` (`force_token: github`) | Tarballs, checksums, SBOMs to `github.com/amauryconstant/twiggit/releases/tag/<tag>` |
| Tag mirror to GitHub | GitLab CI `mirror-to-github` job | `.gitlab-ci.yml` | `git push --tags --force-with-lease` (tag pipelines: `when: never`) |

**Invariant:** The GitHub Actions release job MUST run inside the
`goreleaser/goreleaser:v2.18.2` Docker image (via the `container:` directive),
not via `goreleaser-action@v6`. `goreleaser-action@v6` installs only the
goreleaser binary; goreleaser's `sboms:` step then shells out to `syft` and
fails with `exec: "syft": executable file not found in $PATH`. Using the
Docker image as the job container bundles `goreleaser`, `syft`, `cosign`,
and `gh` — symmetric with the GitLab CI `release` job and with `airk`.

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
