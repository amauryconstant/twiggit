# Design

## Context

`openspec/specs/infrastructure-release` currently describes GitLab as
the canonical release host and GitHub as a discoverability stub. The
discoverability stub is implemented in
`.goreleaser/hooks/create-github-release.sh`, invoked from
`.goreleaser.yml` `before.hooks`. It runs `gh release create` with no
assets; users landing on `github.com/amauryconstant/twiggit/releases`
see a release with no downloadables and a link back to GitLab.

GitLab CI already publishes a full GitLab Release (tarballs,
checksums, SBOMs) with cosign SBOM signing via GitLab OIDC. The
GitLab job also preflights existing releases via the GitLab Releases
API before invoking goreleaser, and `.goreleaser.yml` sets
`release.replace_existing_artifacts: false` so goreleaser itself
cannot overwrite artifacts.

Tag flow today: `git push origin vX.Y.Z` → GitLab runs `release` →
`mirror-to-github` pushes the tag to the GitHub remote. The GitHub
side currently stops at the tag push. We add a GitHub Actions
workflow that consumes the pushed tag and publishes a first-class
GitHub Release.

See `proposal.md` for motivation and
`openspec/changes/github-release-ci/specs/infrastructure-release/spec.md`
for requirements.

## Goals / Non-Goals

**Goals:**

- Promote GitHub Releases from a discoverability stub to a
  first-class publish target with the same artifact set as GitLab
  (tarballs, checksums, SBOMs).
- Reuse the existing tag transport (`mirror-to-github`) and pin the
  GitHub Actions workflow to the same toolchain the GitLab job uses.
- Mirror the GitLab job's safety properties on the GitHub side:
  tag-format guard, release-exists preflight, `replace_existing_artifacts:
  false`, cosign SBOM signing via provider OIDC.

**Non-Goals:**

- Moving CI/test/lint off GitLab. GitLab stays the build pipeline.
- Replacing the GitLab release job with the GitHub Actions job, or
  vice-versa. Both publish in parallel.
- Migrating Homebrew tap publishing to GitHub Actions. Homebrew
  publishing happens via goreleaser `homebrew_casks` during the
  goreleaser run; on GitHub Actions it would need a
  `HOMEBREW_TAP_GITHUB_TOKEN` secret with push rights on
  `amoconst/homebrew-tap`. The GitLab job already handles Homebrew
  publishing with its existing `GITHUB_TOKEN`; duplicating that to
  GitHub Actions is out of scope.
- Adding goreleaser `homebrew_casks.github` upload to the GitHub
  Actions run (would double-publish to Homebrew).
- Adding CI on GitHub Actions (lint/test/build). Release-only.

## Decisions

### 1. Separate goreleaser invocations per provider

**Choice:** GitLab `release` job runs goreleaser with `release.gitlab`;
GitHub Actions `release.yml` runs goreleaser with `release.github`.

**Rationale:** Each provider runs its own goreleaser invocation on its
native CI. No cross-token plumbing (no `GITHUB_TOKEN` in GitLab CI,
no `GITLAB_TOKEN` in GitHub Actions). Provider-native OIDC for cosign
keyless. Failure isolation: a GitHub-side failure does not roll back
the GitLab release and vice versa. The current `.goreleaser.yml` has
no `release.github` block, so adding one is additive.

**Alternatives considered:**

- *Single goreleaser run publishing to both*: requires
  `GITHUB_TOKEN` in GitLab CI and `GITLAB_TOKEN` in GitHub Actions;
  couples provider availability; one failure aborts the other
  provider's publish.
- *GitHub Actions only uploads artifacts built by GitLab job via
  `gh release upload`*: requires GitLab to ship dist artifacts
  across to GitHub (artifact download in GH workflow is workable but
  fragile, doubles the goreleaser build cost, and makes the GitHub
  Release depend on the GitLab job finishing first).

### 2. Goreleaser config: drop `create-github-release.sh` hook, add
`release.github`

**Choice:** Remove the `- bash .goreleaser/hooks/create-github-release.sh`
line from `.goreleaser.yml` `before.hooks`. Add a `release.github`
block pointing at `amauryconstant/twiggit`. Delete the hook script.

**Rationale:** The hook is a discoverability stub. Goreleaser's
native `release.github` block does the same job plus uploads
artifacts and SBOMs. Keeping the hook would race the native publish
and create a second empty GitHub Release on tag push.

**Alternatives considered:**

- *Keep the hook as a fallback*: goreleaser's native `release.github`
  is more reliable; the hook becomes dead code on success and a
  source of confusion on failure.

### 3. Tag transport stays on GitLab `mirror-to-github`

**Choice:** The GitLab `mirror-to-github` job continues to push the
tag to `github.com/amauryconstant/twiggit`. The new GitHub Actions
workflow triggers on the resulting `v*` tag push.

**Rationale:** The job already exists, runs on every push, and uses
the same `GITHUB_ACCESS_TOKEN` + `GITHUB_REPO_URL` pair the project
already maintains. Re-uses existing infrastructure and secrets.

**Alternatives considered:**

- *Re-tag from GitHub Actions on tag push*: redundant; the GitLab
  mirror job is the single tag-push path. Keeping it there keeps the
  flow direction `gitlab → github`.

### 4. Cosign keyless via provider OIDC

**Choice:** GitHub Actions workflow declares
`permissions: id-token: write` and uses `cosign sign --yes` with no
explicit key material. The GitLab job keeps its existing keyless flow
via GitLab `id_tokens.SIGSTORE_ID_TOKEN`.

**Rationale:** Mirrors the GitLab job's model. Two transparency log
entries per SBOM (one per provider OIDC issuer) is acceptable: the
SBOM content is identical, and consumers can verify with either
issuer's identity.

**Alternatives considered:**

- *Reuse a single cosign key on both providers*: requires long-lived
  key material in two CI systems; keyless is the documented
  twiggit contract in `infrastructure-release` and the SBOMs already
  use it.
- *Skip signing on the GitHub side*: asymmetric with GitLab and
  weakens the supply-chain signal; spec already requires cosign
  across the release pipeline.

### 5. Goreleaser dry-run + Homebrew publish unchanged

**Choice (dry-run):** GitLab `goreleaser-dry-run` continues to gate
MRs that touch release-affecting paths. No GitHub Actions dry-run
job.

**Choice (Homebrew):** Both the GitLab `release` job and the new
GitHub Actions `release.yml` run goreleaser with the shared
`.goreleaser.yml`, which contains a single `homebrew_casks` block.
Both invocations therefore push a formula update to
`amoconst/homebrew-tap` via `GITHUB_TOKEN`. The formula content is
reproducible from the tag commit (goreleaser renders the same
`twiggit.rb` on each run), so a double-push is content-identical and
idempotent. No gating via `--skip=brews` is required.

**Rationale:** Goreleaser formula updates are content-stable per
tag and the Homebrew tap already accepts the same `GITHUB_TOKEN`
from either CI provider. A single `--skip=brews` flag on the GitHub
Actions run would couple the two CI jobs' goreleaser invocations and
add a divergence surface; the idempotent double-push is simpler.
GitLab CI stays the build/test gate; GitHub Actions is release-only
per the non-goals. A separate dry-run job on GitHub Actions would
double the goreleaser cost on every PR with no additional signal.

## Risks / Trade-offs

- **[Drift between the two goreleaser invocations]** → Both pin the
  same `.goreleaser.yml` from the same tag commit, so artifact content
  is identical. Changelog/version come from `git` data, which is the
  same in both checkouts.
- **[GitHub Actions runner cold-start latency vs GitLab DinD]** →
  Acceptable; release is not latency-sensitive. No mitigation needed.
- **[Two transparency log entries per SBOM (GitLab + GitHub OIDC)]** →
  Documented in `infrastructure-release`. Consumers can verify with
  either issuer; the SBOM content is the same.
- **[mirror-to-github depends on `GITHUB_ACCESS_TOKEN` + `GITHUB_REPO_URL`]** →
  Both already set in the GitLab project CI/CD variables; no new
  secrets.
- **[Homebrew double-publish on every tag]** → Both goreleaser
  invocations push `amoconst/homebrew-tap` with identical formula
  content. Idempotent; no user-visible impact. Documented in
  Decision 5.
- **[No CI on GitHub mirror]** → Forks/PRs on the GitHub mirror get
  no CI feedback. Out of scope by design; acceptable for release-only
  scope.

## Migration Plan

1. Land the change in one MR: `.goreleaser.yml` diff,
   `.goreleaser/hooks/create-github-release.sh` deletion,
   `.github/workflows/release.yml` creation, spec delta.
2. Push a throwaway tag `v0.0.0-test1` to a scratch branch.
3. Confirm GitLab release job publishes to GitLab Releases with
   signed SBOM.
4. Confirm `mirror-to-github` pushes the tag.
5. Confirm GitHub Actions `release.yml` publishes to GitHub Releases
   with signed SBOM.
6. Confirm both releases carry identical artifact sets.
7. Re-push the same tag — both preflights fail cleanly, no overwrite.
8. Delete test artifacts from both providers.
9. Merge the MR. The next real tag triggers both publish paths.

Rollback: revert the MR. The `.goreleaser/hooks/create-github-release.sh`
deletion is non-recoverable from git history in the merged tree, so
keep the branch around for one release cycle if you want a safety net
to restore the stub. We do not keep the stub.

## Open Questions

None.
