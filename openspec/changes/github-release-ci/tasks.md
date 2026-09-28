# Tasks

## 1. Goreleaser config

- [ ] 1.1 Add `release.github` block (`owner: amauryconstant`, `name: twiggit`) to `.goreleaser.yml` and verify goreleaser config validates via `goreleaser check`
- [ ] 1.2 Remove `- bash .goreleaser/hooks/create-github-release.sh` from `.goreleaser.yml` `before.hooks` and verify the hook reference is gone

## 2. Hook removal

- [ ] 2.1 Delete `.goreleaser/hooks/create-github-release.sh` and verify `git status` shows the deletion

## 3. GitHub Actions workflow

- [ ] 3.1 Create `.github/workflows/release.yml` with `on: push: tags: ['v*']`, `permissions: { contents: write, id-token: write }`, full-history checkout, pinned toolchain and cosign install, tag-format guard, `gh release view` preflight, `goreleaser release --clean`, and a `cosign sign` loop over `dist/*_sbom.spdx.json` — and verify the file exists at the target path
- [ ] 3.2 Run `actionlint` (or `yamllint`) over `.github/workflows/release.yml` and verify no schema or syntax findings
- [ ] 3.3 Verify the workflow triggers only on `v*` tag push by reviewing the `on:` block (no `pull_request`, no `push: branches: [main]`)

## 4. Local validation

- [ ] 4.1 Run `goreleaser release --snapshot --clean --skip=publish` locally and verify `dist/` contains the expected archives, checksums, and SBOMs (proves goreleaser config is internally consistent)
- [ ] 4.2 Run `mise run lint:check` (or equivalent golangci-lint invocation) to verify YAML/Go lint passes — and verify no new lint findings appear
- [ ] 4.3 Run `openspec validate --change github-release-ci` and verify the change validates clean

## 5. End-to-end release smoke test

- [ ] 5.1 Push a throwaway tag `v0.0.0-test1` to a scratch branch and verify the GitLab `release` job publishes to GitLab Releases with cosign-signed SBOM
- [ ] 5.2 Verify the GitLab `mirror-to-github` job pushes the tag to `github.com/amauryconstant/twiggit`
- [ ] 5.3 Verify the new `.github/workflows/release.yml` runs on the pushed tag, publishes to the GitHub Release for that tag with cosign-signed SBOM, and the artifact set matches GitLab (same filenames, same checksums, same SBOM content)
- [ ] 5.4 Re-push the same tag and verify both providers' preflights fail with a non-zero exit and no artifact is overwritten
- [ ] 5.5 Delete the throwaway releases and tag and verify `git push origin :refs/tags/v0.0.0-test1` removes the tag from both remotes
