# Tasks

## 1. Foundation polish (`.golangci.yml`, `.dockerignore`, `Dockerfile.ci`, mise, AGENTS.md)

- [ ] 1.1 Update `.golangci.yml` exclusions regex to use `^` anchors and a narrower surface, verify `golangci-lint run ./...` still exits 0 on a clean tree
- [ ] 1.2 Update `.golangci.yml` `staticcheck.checks` to enumerate the excluded checks (`-ST1000,-ST1003,-ST1005,-ST1016,-ST1020,-ST1021`), verify `staticcheck` passes on `internal/` and `cmd/`
- [ ] 1.3 Drop the redundant `twiggit/cmd` self-include from `.golangci.yml` `cmd:` depguard `allow`, verify `golangci-lint run` exits 0 (depguard still recognises same-package imports implicitly)
- [ ] 1.4 Add `.gitlab/`, `openspec/`, `.github/` to `.dockerignore`, verify `docker build --dry-run -f Dockerfile.ci . | wc -c` produces a measurably smaller context than before
- [ ] 1.5 Add `.cache-key` to `.dockerignore`, verify the file is omitted from any `docker build` context lookup
- [ ] 1.6 Add a non-root `USER` to `Dockerfile.ci` final stage (UID 1000 or equivalent), verify the runtime user can write to `/workspace/.gomodcache` and `/workspace/go` by either chowning the existing paths in the build stage or relocating them, verify `docker run --user <uid> <image> mise ls --json` exits 0
- [ ] 1.7 Update `AGENTS.md` and `CONTRIBUTING.md` to record that `mise install && pre-commit install` provisions the full toolchain (gopls, golangci-lint, goreleaser, govulncheck, cosign, pre-commit hooks); reference the developer-bootstrap section, verify a clean checkout runs `mise install && pre-commit install && mise run verify` end-to-end without manual steps
- [ ] 1.8 Switch `.mise/config.toml` `gopls:check` from `gopls check **/*.go` to `gopls check ./...`, verify `mise run gopls:check` exits 0 on a clean tree
- [ ] 1.9 Add `min_version = "2024.11.1"` (or current mise) to `.mise/config.toml` to pin the minimum mise release that the project's `[env] experimental = true` and Go backend syntax rely on, verify `mise doctor` no longer warns about the version
- [ ] 1.10 Add `"go:github.com/sigstore/cosign/v3/cmd/cosign" = "<pinned-v3-version>"` to `.mise/config.toml` `[tools]` (matching the `goreleaser/goreleaser:v2.18.2` image bundle's `cosign v3.1.3`), verify `mise install && cosign version` prints the pinned version

## 2. `.gitlab-ci.yml` defaults + mirror + DinD service

- [ ] 2.1 In `.gitlab-ci.yml` `build-ci-image`, keep the `docker:29.1.4-dind` service image and add `DOCKER_TLS_CERTDIR: "/certs"` to job `variables:` (per `dockerd-entrypoint.sh`; the `:tls` image variant does not exist). Set `DOCKER_HOST: tcp://docker:2376`. Mount `/certs/client` from the service into the client job (not the broader `/certs`), verify a test MR's `build-ci-image` job logs the TLS socket URL and `docker build` succeeds via TLS
- [ ] 2.2 Add `default:` block entries to `.gitlab-ci.yml` for `tags:`, `retry: { max: 2, when: [runner_system_failure, stuck_or_timeout_failure] }`, and `interruptible: true`, verify no existing job explicitly overrides these in ways that conflict (search for `tags:` / `retry:` / `interruptible:` overrides)
- [ ] 2.3 Remove `--force` from the `mirror-to-github` `git push` line, change to `git push github HEAD:$CI_COMMIT_BRANCH --tags`, verify the script syntactically accepts the change (lint and pipeline parse)
- [ ] 2.4 Set `interruptible: false` on the `release` job, verify the release rules block remains intact
- [ ] 2.5 Extend `mirror-to-github` rules to ensure `when: never` for tag triggers remains paired with non-forced push semantics on the main-branch sync (no other action beyond 2.3)

## 3. `.gitlab-ci.yml` lint + test gates + race jobs

- [ ] 3.1 Prefix the `lint` job script with `go mod tidy && git diff --exit-code`, verify the lint job fails locally when `go.mod` is dirty (`echo "module twiggit" >> go.mod`) and passes after `git checkout -- go.mod`. Add job-level `variables: { GOFLAGS: "-mod=readonly" }` so the lint command is reproducible across runners
- [ ] 3.2 Add `go mod tidy -diff` to the `test` job `before_script`, plus job-level `variables: { GOFLAGS: "-mod=readonly" }`, verify a synthetic drift fails the test job before tests run
- [ ] 3.3 Add a Go-version matrix to the `test` job: two entries keyed off `go.mod` `go` directive and the prior minor's latest patch, set `parallel: matrix` and `fail-fast: false`, verify both entries run on a test MR
- [ ] 3.4 Add a standalone `test:race` job that runs `go test -race ./...`, verify race-only failures do not block lint/matrix
- [ ] 3.5 Add a fast `test:race:subset` job that runs `go test -race -shuffle=on ./internal/... ./cmd/...`, verify the subset completes in the order of minutes on a test MR (target <10 min wall-clock). Note: `mise run test:race` locally stays deterministic (no `-shuffle=on`) to keep local runs reproducible

## 4. `.gitlab-ci.yml` supply-chain (Trivy, cosign, image triggers, drift check)

- [ ] 4.1 Insert `trivy image --exit-code 1 --severity CRITICAL,HIGH --no-progress` at the end of `build-ci-image` after both `docker push` invocations, verify the scan logs `CRITICAL`/`HIGH` findings against the published image and the step exits non-zero on a known-vulnerable tag
- [ ] 4.2 Add `"go:github.com/sigstore/cosign/v3/cmd/cosign" = "<pinned-v3-version>"` to `.mise/config.toml` `[tools]` (see also task 1.10)
- [ ] 4.3 Pin the goreleaser image to `goreleaser/goreleaser:v2.18.2` in both `goreleaser-dry-run` and `release` jobs, verify the value matches `.mise/config.toml:4`
- [ ] 4.4 Extend the `build-ci-image` trigger to include `.goreleaser.yml`, verify a synthetic edit to that file queues the image job
- [ ] 4.5 Remove the path-based trigger filter from `goreleaser-dry-run` (run on every MR), add `artifacts: { expire_in: 1 day }` to its artifacts, verify a no-op MR still runs the dry-run and the artifacts self-clean
- [ ] 4.6 In the `release` job: (a) add a `before_script` GitLab Releases API preflight that exits non-zero when `curl -sI ... /releases/v<tag>` returns 200 (tag already published); (b) add a `cosign sign --yes` step after `goreleaser release --clean` that signs each `*_sbom.spdx.json` artifact using the image-bundled `cosign v3.1.3`; (c) configure job-level `id_tokens: { SIGSTORE_ID_TOKEN: { aud: sigstore } }` for OIDC keyless signing with a `before_script` switch to `$COSIGN_KEY` when `$SIGSTORE_ID_TOKEN` is unset; (d) verify the cosign step prints a Rekor log entry URL on success
- [ ] 4.7 Add an image-drift coupling check in `goreleaser-dry-run` and `release` `before_script`: parse the pinned goreleaser version from `.mise/config.toml`, compare to `goreleaser --version` in the image, exit non-zero on mismatch. Verify a synthetic bump to `.mise/config.toml goreleaser = "2.19.0"` (without bumping the image) fails both jobs

## 5. `.goreleaser.yml` release publish mode

- [ ] 5.1 Keep `.goreleaser.yml:75` `release.mode: replace` (release-notes merge mode is unrelated to artifact overwrite) and add `release.replace_existing_artifacts: false`, verify `goreleaser check` exits 0 and the YAML schema recognises both values

## 6. Change validation

- [ ] 6.1 Run `openspec validate --changes infra-release-hardening --strict`, verify the spec/proposal/design/tasks are coherent and no missing cross-references
- [ ] 6.2 Run `mise run verify` locally, verify format, lint:fix, lint:gated, vuln:check, test, and build all exit 0
- [ ] 6.3 Push the branch and open a test MR, verify every new/modified CI job runs (`build-ci-image`, `lint` with go mod tidy gate, `test:race:subset`, `test` matrix entries, `goreleaser-dry-run` on every MR, `setup`, `default:` tags/retry/interruptible applied)
- [ ] 6.4 Confirm `cosign verify-blob --signature <sig> <sbom>` returns success on the SBOM published by the test tag, document the verification command in `AGENTS.md` or `CONTRIBUTING.md` if not already present

## 7. cmdutil refactor + depguard narrow (composition-root migration; prerequisite for the Group 1 depguard narrow)

- [ ] 7.1a Declare `type Client = interface{}` in `internal/cmdutil/factory.go` and change the `Factory.GitClient` field type from `func() (*git.Client, error)` to `func() (Client, error)`. Verify `golangci-lint run ./internal/cmdutil/...` exits 0 and that `*git.Client` satisfies `Client` implicitly (no call-site changes needed because callers narrow via per-role fields `RepoOpener`, `BranchReader`, etc., which already type as `core.*` interfaces)
- [ ] 7.1b Add `WithVersion(string) FactoryOption` to `internal/cmdutil/factory.go` and remove the `version.Version` read from `NewFactory`. The `AppVersion` field defaults to empty when no option is supplied. Verify `cmdutil.NewFactory()` still compiles and that `factory_test.go` `TestNewFactory_ReturnsNonNilWithSystemIOStreams` is updated to construct via `cmdutil.NewFactory(cmdutil.WithVersion("test"))` and remains green
- [ ] 7.1c Add `WithConfigLoader(func() (*core.Config, error)) FactoryOption` and `WithGitClientFactory(func() (cmdutil.Client, error)) FactoryOption` to `internal/cmdutil/factory.go`; remove the `config.NewManager()` and `git.NewClient()` calls from `NewFactory`. The default `Config` and `GitClient` fields return a sentinel error (e.g., `errors.New("cmdutil: no config loader wired")`) if invoked without options. Verify `golangci-lint run ./internal/cmdutil/...` no longer imports `internal/{git,config}` and that `cmd/` still compiles
- [ ] 7.1d Update `main.go` to call `cmdutil.NewFactory(cmdutil.WithVersion(version.Version), cmdutil.WithConfigLoader(loadConfig), cmdutil.WithGitClientFactory(newGitClient))` where `loadConfig` and `newGitClient` are small closures that call `config.NewManager().Load` and `git.NewClient()` respectively. Verify `go build ./... && go vet ./... && mise run test` all exit 0, and that `factory.Init()` still passes (the lazy fields still cache via `sync.OnceValues` inside the option closures)
- [ ] 7.1e Update `factory_test.go` to exercise both construction paths: the no-arg default (lazy fields return sentinel errors; `Init()` returns the joined errors) and the wired path (use `cmdutil.WithConfigLoader` to inject a deterministic stub and verify caching across calls). Verify all existing test scenarios pass plus the new ones
- [ ] 7.2 Narrow the `.golangci.yml` `cmdutil:` depguard `allow` list to `[$gostd, github.com/samber/lo, github.com/spf13/cobra, twiggit/internal/core, twiggit/internal/iostreams]` (drop `twiggit/internal/{output, git, config, version}`), verify `golangci-lint run ./internal/cmdutil/...` exits 0 and `cmd/` still compiles. Tasks 7.1a-7.1e MUST be complete before this lands; if any are not yet merged, the lint step fails immediately
