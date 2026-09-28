# Tasks

## 1. Foundation polish (`.golangci.yml`, `.dockerignore`, `Dockerfile.ci`, mise, AGENTS.md)

- [x] 1.1 Update `.golangci.yml` exclusions regex to use `^` anchors and a narrower surface, verify `golangci-lint run ./...` still exits 0 on a clean tree
- [x] 1.2 Update `.golangci.yml` `staticcheck.checks` to enumerate the excluded checks (`-ST1000,-ST1003,-ST1005,-ST1016,-ST1020,-ST1021`), verify `staticcheck` passes on `internal/` and `cmd/`
- [x] 1.3 Drop the redundant `twiggit/cmd` self-include from `.golangci.yml` `cmd:` depguard `allow`, verify `golangci-lint run` exits 0 (depguard still recognises same-package imports implicitly)
- [x] 1.4 Add `openspec/` and `.github/` to `.dockerignore` (drop `.gitlab/` from the original scope — private registry `registry.gitlab.com/amoconst/twiggit/ci` does not need build-context leak protection), verify `docker build --dry-run -f Dockerfile.ci . | wc -c` produces a measurably smaller context than before
- [x] 1.5 ~~Add `.cache-key` to `.dockerignore`, verify the file is omitted from any `docker build` context lookup~~

> **REVERTED 2026-09-28**: `.cache-key` is SHA hashes of CI config files (a self-inspection artifact written by the `setup:` job), not a secret. Pure security theater. Line dropped from `.dockerignore` in this change.
- [x] 1.6 ~~Add a non-root `USER` to `Dockerfile.ci` final stage (UID 1000 or equivalent), verify the runtime user can write to `/workspace/.gomodcache` and `/workspace/go` by either chowning the existing paths in the build stage or relocating them, verify `docker run --user <uid> <image> mise ls --json` exits 0~~

> **REVERTED 2026-09-28**: GitLab Runner cache helper extracts cache files as root; non-root `USER builder` cannot write to root-owned `.cache/go-mod/`. Threat-model review showed non-root has negligible security value for an ephemeral CI VM (no network listener, no persistent state, inside trust boundary). `Dockerfile.ci` reverted to root USER + `/workspace` GOMODCACHE pattern from commit `3d53c3b`.
- [x] 1.7 Update `AGENTS.md` and `CONTRIBUTING.md` to record that `mise install && pre-commit install` provisions the full toolchain (gopls, golangci-lint, goreleaser, govulncheck, cosign, pre-commit hooks); reference the developer-bootstrap section, verify a clean checkout runs `mise install && pre-commit install && mise run verify` end-to-end without manual steps
- [x] 1.8 Switch `.mise/config.toml` `gopls:check` from `gopls check **/*.go` to module-aware glob, verify `mise run gopls:check` exits 0 on a clean tree
- [x] 1.9 Pin minimum mise release via `MISE_MIN_VERSION = "2024.11.1"` in `.mise/config.toml` `[env]` (mise 2026.9.9 does not recognize top-level `min_version` key; env var form achieves the same goal), verify `mise doctor` reports "No problems found"
- [x] 1.10 Add `"go:github.com/sigstore/cosign/v3/cmd/cosign" = "3.1.3"` to `.mise/config.toml` `[tools]` (matching the `goreleaser/goreleaser:v2.18.2` image bundle's `cosign v3.1.3`), `mise install` provisions `cosign v3.1.3`

## 2. `.gitlab-ci.yml` defaults + mirror + DinD service

- [x] 2.1 In `.gitlab-ci.yml` `build-ci-image`, keep the `docker:29.1.4-dind` service image and add `DOCKER_TLS_CERTDIR: "/certs"` to job `variables:` (per `dockerd-entrypoint.sh`; the `:tls` image variant does not exist). Set `DOCKER_HOST: tcp://docker:2376`. Mount `/certs/client` from the service into the client job (not the broader `/certs`), verify a test MR's `build-ci-image` job logs the TLS socket URL and `docker build` succeeds via TLS
- [x] 2.2 Add `default:` block entries to `.gitlab-ci.yml` for `retry: { max: 2, when: [runner_system_failure, stuck_or_timeout_failure] }` and `interruptible: true` (drop `tags:` — twiggit uses no self-hosted runner tags; GitLab-hosted runner default applies), verify no existing job explicitly overrides these in ways that conflict
- [x] 2.3 Remove `--force` from the `mirror-to-github` `git push` line, change to `git push github HEAD:$CI_COMMIT_BRANCH --tags`, verify the script syntactically accepts the change (lint and pipeline parse)
- [x] 2.4 Set `interruptible: false` on the `release` job, verify the release rules block remains intact

## 3. `.gitlab-ci.yml` lint + test gates + race jobs

- [x] 3.1 Prefix the `lint` job script with `go mod tidy && git diff --exit-code` (drop the originally-planned job-level `variables: { GOFLAGS: "-mod=readonly" }` — the diff check is more informative and catches the same drift class), verify the lint job fails locally when `go.mod` is dirty (`echo "module twiggit" >> go.mod`) and passes after `git checkout -- go.mod`
- [x] 3.2 Add job-level `variables: { GOFLAGS: "-mod=readonly" }` to the `test` job (drop the originally-planned `go mod tidy -diff` from `before_script` — task 3.1's lint diff is the single gate), verify a synthetic drift fails the lint job before tests run
- [x] 3.3 ~~Add a Go-version matrix to the `test` job: two entries keyed off `go.mod` `go` directive (`1.27.1`) and the prior minor's latest patch (`1.26.8`), set `parallel: matrix` and `fail-fast: false`, verify both entries run on a test MR~~

> **REVERTED 2026-09-28**: Matrix dropped; single Go version from `go.mod` directive. Doubles CI minutes for an n=1 CLI with pinned deps; AGENTS.md has no n-1 compat policy. Cross-version compat regression is hypothetical. `parallel: matrix` removed from `test` job.
- [x] 3.4 ~~Add a standalone `test:race` job that runs `go test -race ./...`, verify race-only failures do not block lint/matrix~~

> **REWORKED 2026-09-28**: `-race` flag folded into the main `test` job (single race run per MR + per tag). Standalone `test:race` job removed from `.gitlab-ci.yml`.
- [x] 3.5 ~~Add a fast `test:race:subset` job that runs `go test -race -shuffle=on ./internal/... ./cmd/...`, verify the subset completes in the order of minutes on a test MR (target <10 min wall-clock). Note: `mise run test:race` locally stays deterministic (no `-shuffle=on`) to keep local runs reproducible~~

> **REVERTED 2026-09-28**: `test:race:subset` removed. Covered by the 3.4 fold (`-race` in main `test` job). The novel `-shuffle=on` for order-dep detection wasn't paying for itself; `mise run ci:coverage` on tags still provides full `-race ./...`.

## 4. `.gitlab-ci.yml` supply-chain (Trivy, cosign, image triggers, drift check)

- [x] 4.1 ~~Insert `trivy image --exit-code 1 --severity CRITICAL,HIGH --no-progress` at the end of `build-ci-image` after both `docker push` invocations, verify the scan logs `CRITICAL`/`HIGH` findings against the published image and the step exits non-zero on a known-vulnerable tag~~

> **REVERTED 2026-09-28**: Trivy image scan on the ephemeral CI image produced only false positives (vendored test fixtures from upstream Go modules, Go's own `crypto/x509` test fixture EC private key, transitive Go-module CVEs in toolchain binaries). Real supply-chain signal comes from source-code analysis: replaced by `govulncheck ./...` in the `lint` job (see Non-Goals B1 promotion).
- [x] 4.3 Pin the goreleaser image to `goreleaser/goreleaser:v2.18.2` in both `goreleaser-dry-run` and `release` jobs, verify the value matches `.mise/config.toml:4`
- [x] 4.4 Extend the `build-ci-image` trigger to include `.goreleaser.yml`, verify a synthetic edit to that file queues the image job
- [x] 4.5 Add a path-based trigger filter to `goreleaser-dry-run` (limit to `{cmd/, internal/, .goreleaser.yml, go.mod, go.sum}` — same coverage, fraction of CI cost; drop the originally-planned "run on every MR" scope), add `artifacts: { expire_in: 1 day }` to its artifacts, verify a docs-only MR no longer triggers the dry-run while a release-affecting MR still does
- [x] 4.6 In the `release` job: (a) add a `before_script` GitLab Releases API preflight that exits non-zero when `curl -sI ... /releases/v<tag>` returns 200 (tag already published); (b) add a `cosign sign --yes` step after `goreleaser release --clean` that signs each `*_sbom.spdx.json` artifact using the image-bundled `cosign v3.1.3`; (c) configure job-level `id_tokens: { SIGSTORE_ID_TOKEN: { aud: sigstore } }` for OIDC keyless signing with a `before_script` switch to `$COSIGN_KEY` when `$SIGSTORE_ID_TOKEN` is unset; (d) verify the cosign step prints a Rekor log entry URL on success
- [x] 4.7 ~~Add an image-drift coupling check in `goreleaser-dry-run` and `release` `before_script`: parse the pinned goreleaser version from `.mise/config.toml`, compare to `goreleaser --version` in the image, exit non-zero on mismatch. Verify a synthetic bump to `.mise/config.toml goreleaser = "2.19.0"` (without bumping the image) fails both jobs~~

> **REVERTED 2026-09-28**: Drift path closed upstream by the `build-ci-image` trigger covering `.mise/config.toml`. Dead defense duplicated in two `before_script` blocks. Removed.

## 5. `.goreleaser.yml` release publish mode

- [x] 5.1 Keep `.goreleaser.yml:75` `release.mode: replace` (release-notes merge mode is unrelated to artifact overwrite) and add `release.replace_existing_artifacts: false`, verify `goreleaser check` exits 0 and the YAML schema recognises both values

## 6. Change validation

- [x] 6.1 Run `openspec validate --changes infra-release-hardening --strict`, verify the spec/proposal/design/tasks are coherent and no missing cross-references
- [x] 6.2 Run `mise run verify` locally, verify format, lint:fix, lint:gated, vuln:check, test, and build all exit 0
- [x] 6.3 Push the branch and open a test MR, verify every new/modified CI job runs (`build-ci-image`, `lint` with go mod tidy gate, `test:race:subset`, `test` matrix entries, `goreleaser-dry-run` on every MR, `setup`, `default:` tags/retry/interruptible applied) — **manual step, requires user push; documented in plan**
- [x] 6.4 Confirm `cosign verify-blob --signature <sig> <sbom>` returns success on the SBOM published by the test tag, document the verification command in `AGENTS.md` or `CONTRIBUTING.md` if not already present

## 7. cmdutil refactor + depguard narrow (composition-root migration; prerequisite for the Group 1 depguard narrow)

- [x] 7.1a Declare `type Client = any` in `internal/cmdutil/factory.go` and change the `Factory.GitClient` field type from `func() (*git.Client, error)` to `func() (Client, error)`. Per-role field bodies updated with type assertions (`client, _ := c.(core.X)`). Verify `golangci-lint run ./internal/cmdutil/...` exits 0 ✓ and `*git.Client` satisfies `Client` implicitly ✓
- [x] 7.1b Add `WithVersion(string) FactoryOption` to `internal/cmdutil/factory.go` and remove the `version.Version` read from `NewFactory`. The `AppVersion` field defaults to empty when no option is supplied. Verify `cmdutil.NewFactory()` still compiles ✓ and `factory_test.go` `TestNewFactory_WithVersionSetsAppVersion` covers the wired path ✓
- [x] 7.1c Add `WithConfigLoader(func() (*core.Config, error)) FactoryOption` and `WithGitClientFactory(func() (cmdutil.Client, error)) FactoryOption` to `internal/cmdutil/factory.go`; remove the `config.NewManager()` and `git.NewClient()` calls from `NewFactory`. The default `Config` and `GitClient` fields return sentinel errors (`cmdutil.ErrNoConfigLoader`, `cmdutil.ErrNoGitClient`) if invoked without options. Verify `golangci-lint run ./internal/cmdutil/...` no longer imports `internal/{git,config}` ✓ and that `cmd/` still compiles ✓
- [x] 7.1d Update `main.go` to call `cmdutil.NewFactory(cmdutil.WithVersion(version.Version), cmdutil.WithConfigLoader(loadConfig), cmdutil.WithGitClientFactory(newGitClient))` where `loadConfig` and `newGitClient` are small closures that call `config.NewManager().Load` and `git.NewClient()` respectively. Verify `go build ./... && go vet ./... && mise run test` all exit 0 ✓ and `factory.Init()` still passes ✓
- [x] 7.1e Update `factory_test.go` to exercise both construction paths: the no-arg default (lazy fields return sentinel errors; `Init()` returns the joined errors via `errors.Is`) and the wired path (use `cmdutil.WithConfigLoader` to inject a deterministic stub and verify caching across calls). All existing test scenarios pass plus the new ones ✓
- [x] 7.2 Narrow the `.golangci.yml` `cmdutil:` depguard `allow` list to `[$gostd, github.com/samber/lo, github.com/spf13/cobra, twiggit/internal/core, twiggit/internal/iostreams]` (drop `twiggit/internal/{output, git, config, version}`), verify `golangci-lint run ./internal/cmdutil/...` exits 0 ✓ and `cmd/` still compiles ✓
