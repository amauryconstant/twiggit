# Tasks

## 1. Foundation polish (`.golangci.yml`, `.dockerignore`, `Dockerfile.ci`, IDE, mise)

- [ ] 1.1 Update `.golangci.yml` exclusions regex to use `^` anchors and a narrower surface, verify `golangci-lint run ./...` still exits 0 on a clean tree
- [ ] 1.2 Update `.golangci.yml` `staticcheck.checks` to enumerate the excluded checks (`-ST1000,-ST1003,-ST1005,-ST1016,-ST1020,-ST1021`), verify `staticcheck` passes on `internal/` and `cmd/`
- [ ] 1.3 Drop the redundant `twiggit/cmd` self-include from `.golangci.yml` `cmd:` depguard `allow`, verify `golangci-lint run` exits 0 (depguard still recognises same-package imports implicitly)
- [ ] 1.4 Narrow the `.golangci.yml` `cmdutil:` depguard `allow` to `$gostd + lo + cobra + core + iostreams`, verify `golangci-lint run ./internal/cmdutil/...` exits 0 and that `cmd/` still compiles (any direct import in `internal/cmdutil/` of `internal/{output,git,config,version}` will now fail the linter and require rewiring through `cmd/`)
- [ ] 1.5 Add `.gitlab/`, `openspec/`, `.github/` to `.dockerignore`, verify `docker build --dry-run -f Dockerfile.ci . | wc -c` produces a measurably smaller context than before
- [ ] 1.6 Add `.cache-key` to `.dockerignore`, verify the file is omitted from any `docker build` context lookup
- [ ] 1.7 Add a non-root `USER` to `Dockerfile.ci` final stage (UID 1000 or equivalent), verify the runtime user can write to `/workspace/.gomodcache` and `/workspace/go` by either chowning the existing paths in the build stage or relocating them, verify `docker run --user <uid> <image> mise ls --json` exits 0
- [ ] 1.8 Create `.vscode/settings.json` with gopls enabled, `editor.formatOnSave` set to delegate to golangci-lint, and any per-language Markdown preferences, verify a developer opening the repo in VS Code inherits the settings
- [ ] 1.9 Update `install.sh` to provision `gopls` (matching `.mise/config.toml` `[tools] "go:golang.org/x/tools/gopls"`), verify `mise install && gopls version` prints the expected version on a clean checkout
- [ ] 1.10 Update `AGENTS.md` to record the editor/LSP/go-version triad and reference `.vscode/settings.json` as the project IDE baseline
- [ ] 1.11 Switch `.mise/config.toml` `gopls:check` from `gopls check **/*.go` to `gopls check ./...`, verify `mise run gopls:check` exits 0 on a clean tree

## 2. `.gitlab-ci.yml` defaults + mirror + DinD service

- [ ] 2.1 Replace the TLS-disabled DinD service in `.gitlab-ci.yml` `build-ci-image` with `docker:29.1.4-dind:tls`, set `DOCKER_TLS_CERTDIR: "/certs"`, switch `DOCKER_HOST: tcp://docker:2376`, mount `/certs` into the client job, verify a test MR's `build-ci-image` job logs the TLS socket URL and `docker build` succeeds
- [ ] 2.2 Add `default:` block to `.gitlab-ci.yml` setting `tags:`, `retry: 2`, and `interruptible: true`, verify no existing job explicitly overrides these in ways that conflict (search for `tags:` / `retry:` / `interruptible:` overrides)
- [ ] 2.3 Remove `--force` from the `mirror-to-github` `git push` line, change to `git push github HEAD:$CI_COMMIT_BRANCH --tags`, verify the script syntactically accepts the change (lint and pipeline parse)
- [ ] 2.4 Set `interruptible: false` on the `release` job, verify the release rules block remains intact
- [ ] 2.5 Extend `mirror-to-github` rules to ensure `when: never` for tag triggers remains paired with non-forced push semantics on the main-branch sync (no other action beyond 2.3)

## 3. `.gitlab-ci.yml` lint + test gates + race jobs

- [ ] 3.1 Prefix the `lint` job script with `go mod tidy && git diff --exit-code`, verify the lint job fails locally when `go.mod` is dirty (`echo "module twiggit" >> go.mod`) and passes after `git checkout -- go.mod`
- [ ] 3.2 Add `go mod tidy -diff` to the `test` job `before_script`, verify a synthetic drift fails the test job before tests run
- [ ] 3.3 Add a Go-version matrix to the `test` job: two entries keyed off `go.mod` `go` directive and the prior minor's latest patch, set `parallel: matrix` and `fail-fast: false`, verify both entries run on a test MR
- [ ] 3.4 Add a standalone `test:race` job that runs `go test -race ./...`, verify race-only failures do not block lint/matrix
- [ ] 3.5 Add a fast `test:race:subset` job that runs `go test -race ./internal/... ./cmd/...`, verify the subset completes in the order of minutes on a test MR (target <10 min wall-clock)

## 4. `.gitlab-ci.yml` supply-chain (Trivy, cosign, image triggers)

- [ ] 4.1 Insert a `trivy image` step at the end of `build-ci-image` after both `docker push` invocations, verify the scan logs `CRITICAL`/`HIGH` findings against the published image
- [ ] 4.2 Add `cosign` to `.mise/config.toml` `[tools]` (pinned version), verify `mise install` provisions the binary
- [ ] 4.3 Pin the goreleaser image to `goreleaser/goreleaser:v2.18.2` in both `goreleaser-dry-run` and `release` jobs, verify the value matches `.mise/config.toml:4`
- [ ] 4.4 Extend the `build-ci-image` trigger to include `.goreleaser.yml`, verify a synthetic edit to that file queues the image job
- [ ] 4.5 Remove the path-based trigger filter from `goreleaser-dry-run` (run on every MR), add `expire_in: 1 day` to its artifacts, verify a no-op MR still runs the dry-run and the artifacts self-clean
- [ ] 4.6 Add a `cosign sign` step after `goreleaser release --clean` in the `release` job; configure `id_tokens` for OIDC keyless signing, add a fallback `before_script` switch for `$COSIGN_KEY` when `$SIGSTORE_ID_TOKEN` is unset, verify the step prints a Rekor log entry URL on success

## 5. `.goreleaser.yml` release publish mode

- [ ] 5.1 Change `.goreleaser.yml:75` from `release.mode: replace` to `release.mode: append`, verify `goreleaser release --help` and the YAML schema still recognise the value
- [ ] 5.2 Confirm the `release` job's goreleaser invocation includes `--fail-if-tag-exists`, verify the dry-run with a known-tagged release runs without overwriting

## 6. Change validation

- [ ] 6.1 Run `openspec validate --change infra-release-hardening --strict`, verify the spec/proposal/design/tasks are coherent and no missing cross-references
- [ ] 6.2 Run `mise run verify` locally, verify format, lint:fix, lint:gated, vuln:check, test, and build all exit 0
- [ ] 6.3 Push the branch and open a test MR, verify every new/modified CI job runs (`build-ci-image`, `lint` with go mod tidy gate, `test:race:subset`, `test` matrix entries, `goreleaser-dry-run` on every MR, `setup`, `default:` tags/retry/interruptible applied)
- [ ] 6.4 On a synthetic `v0.0.0-test` tag, verify `release` job fails because of `--fail-if-tag-exists` after a first tag of the same name already produced a GitLab release in a previous run; document the failure mode in the change notes
- [ ] 6.5 Confirm `cosign verify-blob --signature <sig> <sbom>` returns success on the SBOM published by the test tag, document the verification command in `AGENTS.md` or `CONTRIBUTING.md` if not already present
