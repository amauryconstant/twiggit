# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.15.1] - 2026-10-08



### Added

- Cover all 10 completion shells in `test/e2e/completion_test.go` (powershell, elvish, nushell, oil, tcsh, xonsh, cmd-clink on top of existing bash/zsh/fish); each block asserts exit 0 + shell-specific marker; pending `It` block for unsupported-shell scenario (cli-completion)

### Changed

- `mise run verify` runs golangci-lint exactly once: `lint:gated` invokes `lint:fix` (auto-fix on gate), redundant `lint:fix` entry removed from verify (it was shadowed by `lint:gated`) (ci-tooling)

### Fixed

- `twiggit completion zsh` rejected with `accepts 1 arg(s), received 0` because per-shell subcommands required one positional even though the shell name is encoded in `Use`; switch validator to `cobra.NoArgs`; pin with unit + e2e tests for bash/zsh/fish and the regression (cli-completion)
- `twiggit completion ksh` previously printed parent help and exited 0; replace parent `Args` with a whitelist validator returning `*core.ValidationError` on mismatch (parent gains a minimal `RunE` returning `cmd.Help()` so `cobra` still consults `Args` for unmatched subcommand args and preserves "show help when no shell given"); unlocks the unsupported-shell scenario in `openspec/specs/cli-completion/spec.md` (cli-completion)

## [0.15.0] - 2026-10-06


### Added

- Add `rebase` and `sync` commands for tracking worktree bases across projects, with per-worktree `twiggit.tracked-base` git config, pre/post-rebase and post-sync hook types, two new role interfaces (`Rebaser`, `BaseTracker`), and `[rebase]`/`[sync]` config sections (rebase-sync)

## [0.14.0] - 2026-10-05


### Breaking

- **BREAKING**: Consolidate CLI output and error handling onto `internal/output`; `--output` vocabulary narrows from `json|jsonl|table|plain` to `json|table|plain` (`text` and `jsonl` now exit 2), `--output json` shape changes from envelope to bare array, and `--output plain` emits headerless TSV (cli-cobra-output-format-unification)
- **BREAKING**: Drop `Get` prefix from 8 exported methods, rename `core` data types to drop `Git*` stutter and `*Info` suffix (`GitRepository`→`Repository`, `GitCommit`→`Commit`, `GitBranch`→`Branch`, `BranchInfo`→`Branch`, `WorktreeInfo`→`Worktree`, `RemoteInfo`→`Remote`, `GitDir`→`RepoDir`), remove 6 per-role lazy fields on `cmdutil.Factory`, and split `test/helpers` into 6 content-named packages (`test/worktree`, `test/shell`, `test/git`, `test/repo`, `test/golden`, `test/perf`) (naming-refactor-modernize)

### Added

- Group subcommands in `--help` output via `cmd.AddGroup` with four groups: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion) (cli-cobra-output-format-unification)
- `Tabular` projection interface in `internal/output` (`Header() []string`, `Rows() [][]string`); `cmd/setup.go` shared helper extracted from 5 `run*` functions (cli-cobra-output-format-unification, naming-refactor-modernize)
- Restore five shell-error sentinels in `internal/core/shell_errors.go` (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`) for `errors.Is` matching from `cmd/init.go` (test-observability-error-discipline)

### Changed

- Dispatch order in `internal/output/errors.go` reordered to `ValidationError` → `NotFoundError` → `OperationError` → `UsageError`; switch to `errors.AsType[T]`; `hintFor` and quiet-mode hint suppression absorbed from dead `cmd/error_formatter.go` (cli-cobra-output-format-unification)
- Unify debug logger into a single channel: `iostreams.NewLogger(io.Writer)` is the one constructor; `Factory.Logger`, `slog.Default()`, and `IOStreams.Logger` are the same `*slog.Logger` instance writing to `ios.ErrOut` (test-observability-error-discipline)
- Restore single-handling rule in `internal/git` adapters (`detectOpError`, `discoverProjects` stop calling `slog.Error`); debug logging moves to cmd boundary via `opts.IO.Logger.With("command", ...)` (test-observability-error-discipline)
- Preserve `*exec.ExitError` cause through every adapter failure: `git.NewCommandError` and 6 `NewWorktreeError`/`NewBranchError` call sites pass underlying `err` as `Cause` field (test-observability-error-discipline)
- Split `internal/git/context_resolver.go` (641 LOC) into `context_resolver.go` + `context_resolver_suggest.go`; collapse three `validate*Err` constructors into `validateFieldErr`; promote `reservedNames` to package-level `var` (naming-refactor-modernize)
- Adopt Go 1.26+ stdlib patterns: `slices.Sort` (×3), `wg.Go(...)` (×7), `strings.NewReplacer` (×2), `t.Setenv` (×2 files), strict `goleak.VerifyTestMain` in `test/concurrent`; rename unexported bool fields to `is/has/can` prefix; replace `fmt.Fprintf(os.Stderr,...)` parse-failure warning with `slog.Warn` (naming-refactor-modernize, test-observability-error-discipline)

### Removed

- `cmd/error_formatter.go` (~180 lines) and `cmd/output.go` (~88 lines) deleted after migrating to `internal/output`; `JSONLinesFormatter`, `FormatJSONL` const, and the `jsonl` switch arm removed (cli-cobra-output-format-unification)
- Drop dead code: `ProgressReporter.ReportProgress`, `CreateOptions.HookRunner`, `cmd/completion.go` redundant `ValidArgsFunction`, `cmd/util.go:ProgressReporter` field; replace `cmd.Find`+`RemoveCommand` dance with `CompletionOptions.DisableDefaultCmd = true` (cli-cobra-output-format-unification, naming-refactor-modernize)
- Drop dead error-classification code in `internal/git/errors.go`: `ErrorKindPermission`, `ErrorKindNotFound`, and the unreachable `Is(target error) bool` method (NotFound dispatch already walks cause chain via `*core.NotFoundError.Is()`) (test-observability-error-discipline)
- Remove duplicate `// Package git/core` comments from 4 files that already have `doc.go`; 6 per-role lazy fields on `cmdutil.Factory` removed; 3 `os.Setenv` call sites replaced with `t.Setenv` (test-observability-error-discipline, naming-refactor-modernize)

### Fixed

- GitLab CI image build: serialize tag+branch pipelines via `resource_group: ci-image-build` and `interruptible: false`; add `$CI_COMMIT_TAG` rule so tag pipelines rebuild the image when `Dockerfile.ci` or `.mise/config.toml` change; drop `.goreleaser.yml` from `changes:` (read at release time, not embedded in the image) (a977195)
- Fourteen mock-based tests in `internal/git` add `t.Cleanup(func() { mock.AssertExpectations(t) })` (12 in `writer_test.go`, 2 in `hook_runner_test.go`); suite methods switch from `require(s.T(), ...)` to `s.Require()`; `test/concurrent` adds strict `goleak.VerifyTestMain(m)` (test-observability-error-discipline)
- New `TestCLIClient_NonZeroExit_PreservesExecError` asserts `errors.As(err, &*exec.ExitError)` walks the chain (test-observability-error-discipline)
- Delete identical-twins `resolveFromWorktreeContext`; route both `core.ContextProject` and `core.ContextWorktree` through `resolveFromProjectContext`; remove dead `existingOnly` parameter on `addBranchSuggestions` (naming-refactor-modernize)
- Add `context.Context` first-parameter to `getProjectContextSuggestions`/`getWorktreeContextSuggestions`/`getOutsideGitContextSuggestions` so inner `context.Background()` calls honour caller cancellation (naming-refactor-modernize)

## [0.13.7] - 2026-09-29


### Fixed

- Release pipeline (GitLab CI + GitHub Actions): cosign v3 removed the `--output-signature` flag; the post-release shell loop that signed each `dist/*_sbom.spdx.json` with `cosign sign-blob --output-signature <sbom>.sig` failed on v0.13.5 (GitHub hard-aborted with `Error: must specify --bundle with --new-bundle-format`; GitLab misread the local file as an OCI image reference and hit Docker Hub `UNAUTHORIZED`). Move signing into goreleaser's native `signs:` block targeting `checksums.txt` via `cosign sign-blob --bundle` (cosign v3's single-file format combining signature + cert + Rekor inclusion into a `<file>.sigstore.json` JSON bundle). Per-SBOM signing is dropped because every SBOM hash is already in the signed checksum file, so a single signature transitively covers all artifacts. Consumers verify via `cosign verify-blob` against `checksums.txt` using the keyless OIDC + identity flags, then `sha256sum -c --ignore-missing checksums.txt` to transitively validate every archive and SBOM — see `AGENTS.md` "Release verification" for the GitLab and GitHub cert-identity strings. Drops the per-SBOM cosign shell loop from `.gitlab-ci.yml` and the `Sign SBOMs (OIDC keyless)` step from `.github/workflows/release.yml`; both pipelines now rely on goreleaser to invoke cosign as a child process with the runner's ambient OIDC token. AGENTS documents the transitive verification chain; `infrastructure-release` spec rewrites the GitHub Actions workflow requirement and the supply-chain-signal scenario around the new model.

### Changed

- Pin `syft = "latest"` in `.mise/config.toml` so `goreleaser release --snapshot` runs locally without going through the `goreleaser/goreleaser:v2.18.2` Docker image. Both surfaces pin to a moving `latest`; CI uses the image-bundled binary. AGENTS toolchain row updated to reflect the dual-surface pin.

## [0.13.6] - 2026-09-29



### Fixed

- GitLab CI release job: pin `GORELEASER_FORCE_TOKEN: gitlab` so the cross-platform `GITHUB_TOKEN` no longer collides with `GITLAB_TOKEN` at goreleaser startup ("multiple tokens found, but only one is allowed"); mark the workspace safe for in-container git (`git config --global --add safe.directory "$CI_PROJECT_DIR"` in `before_script`) so goreleaser can shell out to `git describe` under its root uid without hitting `dubious ownership`. Mirrors the GitHub Actions safe.directory fix shipped in v0.13.5.
- Release pipeline (GitLab CI + GitHub Actions): replace `cosign sign --yes` with `cosign sign-blob --yes --output-signature <sbom>.sig`. `cosign sign` treats the SBOM filename as an OCI image reference and tries to push to Docker Hub (HTTP 401 against `index.docker.io/v2/<sbom-path>/manifests/latest`); `cosign sign-blob` is the correct verb for file signing, and `--output-signature` writes a detached `.sig` sidecar next to the SBOM (sign-blob default is stdout). Consumers verify via `cosign verify-blob` against Rekor using the keyless OIDC + identity flags — see `AGENTS.md` "SBOM verification" for the GitLab and GitHub cert-identity strings.
- GitHub Actions release workflow: add a signing-mode fail-fast step that checks `ACTIONS_ID_TOKEN_REQUEST_TOKEN` (the GitHub Actions OIDC bearer — `SIGSTORE_ID_TOKEN` is not auto-populated here; cosign reads GitHub-native env vars directly) before goreleaser runs; exit 1 with `::error::` if neither that nor `COSIGN_KEY` is present. Mirrors the GitLab CI fail-fast guard in `before_script`.

### Removed

- Homebrew cask distribution: drop the `homebrew_casks` block from `.goreleaser.github.yml`, the Homebrew (macOS) install section from `README.md`, the `Homebrew tap` requirement from `openspec/specs/infrastructure-release/spec.md`, and the homebrew rows, cross-SCM invariant, and 401 troubleshooting entry from `AGENTS.md`. The pre-existing cask on `amoconst/homebrew-tap` will not auto-update; users who installed via `brew install amoconst/twiggit/twiggit` should switch to the install script or manual download before the next release.

## [0.13.5] - 2026-09-29



### Fixed

- GitHub Actions release workflow: run the release job inside the `goreleaser/goreleaser:v2.18.2` Docker image as the job container (`container:` directive, `--entrypoint /bin/sh`). v0.13.4 attempted to install `syft` via `anchore/sbom-action/install@v0`, but that action path doesn't exist (`anchore/sbom-action` is a single composite, no `/install` subdirectory) and the workflow errored at download time with `Can't find 'action.yml', 'action.yaml' or 'Dockerfile'`. A follow-up tarball-download step (`a595ad0`) was itself superseded by the container approach, which mirrors the GitLab CI release job and the sibling `airk` project's release job — the image bundles `goreleaser`, `syft`, `cosign`, and `gh`, so no per-tool install is needed. Drops the now-redundant `Install cosign` step and the `COSIGN_EXPERIMENTAL: 'true'` env flag (cosign v3 ships OIDC keyless signing out of experimental).

## [0.13.4] - 2026-09-29


### Fixed

- GitHub Actions release workflow: install `syft` v1.52.0 into `$PATH` via `anchore/sbom-action/install@v0` before goreleaser runs. `goreleaser-action@v6` ships only the goreleaser binary, so the `sboms:` step would fail with `exec: "syft": executable file not found in $PATH` and ship no SBOMs to the GitHub release. The GitLab CI release job is unaffected (the `goreleaser/goreleaser:v2.18.2` Docker image bundles both tools); bump `syft` in lockstep with goreleaser SBOM format expectations.

### Changed

- Distribution surface ownership now split at the goreleaser config layer: GitLab CI `release` job (`.goreleaser.yml`, `force_token: gitlab`) owns the GitLab release; GitHub Actions `release.yml` (`.goreleaser.github.yml`, `force_token: github`) owns the GitHub release and the Homebrew cask push. Both providers publish the same tarballs, checksums, and SBOMs for each `v*` tag — GitHub is no longer a discoverability stub for GitLab. `homebrew_casks` MUST stay in `.goreleaser.github.yml` only; goreleaser OSS cannot cross-SCM publish (the `token_type` field that would unlock it is Pro-only), and putting it back into `.goreleaser.yml` re-introduces the v0.13.2 401 because the branch-existence pre-flight uses the inferred SCM's API. AGENTS documents the split (toolchain row for `syft`, distribution ownership table, homebrew 401 troubleshooting entry); `infrastructure-release` spec adds the GitHub Actions workflow requirement (syft install, tag-format guard, per-provider preflight, cask ownership) and rewrites the GitLab/GitHub primary requirements around the new split.

## [0.13.3] - 2026-09-29


### Fixed

- GitHub Actions release workflow: replace the `github_token:` step input (silently ignored by `goreleaser-action@v6` — only `distribution`, `version`, `args`, `workdir`, `install-only` are valid inputs) with a `GITHUB_TOKEN` env block on the goreleaser step. The action forwards step `env:` to the goreleaser subprocess; without it the subprocess bails at "missing GITHUB_TOKEN, GITLAB_TOKEN and GITEA_TOKEN" before any artifact is built.
- Release pipeline: move `homebrew_casks` upload from `.goreleaser.yml` (GitLab) to `.goreleaser.github.yml`. Goreleaser OSS cannot cross-publish — the `homebrew_casks` branch existence pre-flight uses the inferred SCM's API, which is GitLab when `release.gitlab` is set. The target tap `amoconst/homebrew-tap` lives on GitHub, so the verify call hit `gitlab.com/api/v4/...` and returned 401 before the push. The `token_type` field that would unlock cross-SCM is Pro-only. GitHub Actions release job owns the push; the `homepage` and `url.template` still point at `gitlab.com/amoconst/twiggit` because that is where the binary artifacts actually live.

## [0.13.2] - 2026-09-29


### Fixed

- GitLab release job: pin `force_token: gitlab` in `.goreleaser.yml` so the cross-platform `GITHUB_TOKEN` env var (required by `homebrew_casks.token` template for the `amoconst/homebrew-tap` push) no longer triggers "multiple tokens found, but only one is allowed"
- GitHub Actions release workflow: pass `github_token: ${{ secrets.GITHUB_TOKEN }}` to `goreleaser-action@v6` (the v6 release dropped the auto-injection of `${{ github.token }}`, so the action saw zero SCM tokens and bailed at "missing GITHUB_TOKEN, GITLAB_TOKEN and GITEA_TOKEN"); switch to a single `-f .goreleaser.github.yml` arg and bump `actions/checkout` to `@v5` to silence the Node 20 deprecation warning
- `.goreleaser.github.yml`: self-contained (duplicates the shared build/archive/sbom/changelog blocks from `.goreleaser.yml`). v2.18.2 OSS has no `includes:` directive (Pro-only) and the `-f` flag takes only the last value via cobra `StringVarP`, so the previous override-file pattern silently loaded the stub and fell back to goreleaser defaults — `paths=.`, default arch including `linux_386`/`windows_386`, no SBOMs, no archive `name_template`. Keep both files in sync when touching the shared blocks.

## [0.13.1] - 2026-09-29


### Changed
- Release tag task now pushes `main` alongside the tag so the GitLab mirror-to-github job picks up the release commit (it skips tag pipelines)

### Fixed

- GitLab release job now forwards `GITHUB_TOKEN` to GoReleaser so the Homebrew cask upload to `amoconst/homebrew-tap` completes (template evaluation aborted the release before artifacts shipped)
- GitHub Actions release workflow now uses `sigstore/cosign-installer@v4.1.0` (the `aquasecurity/cosign-installer` repository was removed; cosign binary stays pinned at `v3.1.3`)
- `mise run release:*` subtasks (`prepare`, `check`, `tag`) no longer declare `sources`, so mise stops skipping them on subsequent invocations when `CHANGELOG.md` is unchanged; the `release` chain now always re-runs the validation/preview before the destructive tag step

## [0.13.0] - 2026-09-28

### Added

- Role interfaces (`Reader`, `Writer`, `Executor`) declared consumer-side in `internal/core/git.go` with compile-time satisfaction checks on `*git.Client`; per-role lazy fields on `cmdutil.Factory` (interface-segregation)
- TLS-enabled DinD CI image, non-forced GitHub mirror push, `replace_existing_artifacts: false` for releases, govulncheck as required CI gate (infra-release-hardening)
- GitHub Actions release workflow publishing GitHub Releases alongside GitLab artifacts (`.goreleaser.github.yml`, `.github/workflows/release.yml` on `v*` tags with `gh release view` preflight and OIDC keyless cosign SBOM signing)

### Changed

- `.golangci.yml` test-exclusion regex anchored; `staticcheck.checks` enumerates exclusions explicitly; depguard allow-lists narrowed for `cmdutil`; `.dockerignore` excludes `openspec/` and `.github/` (infra-release-hardening)

### Breaking

- **BREAKING**: Adopt sentinel-based error chain with `errors.Is`/`errors.As` walks; collapse CLI exit codes 3-6 into single failure code 1 — scripts keyed on per-resource codes break by design (foundation-error-chain)
- **BREAKING**: Invert architecture layers to `cmd → cmdutil → core`; delete `internal/application/`, `internal/service/`, `CompositeGitClient`, `application.GitClient` umbrella (architecture-layer-inversion)
- **BREAKING**: Collapse five-layer DDD-Light into Tier 2 (`golang-cli`) layout — new `internal/core/`, `internal/git/`, `internal/output/`, `internal/iostreams/`, `internal/cmdutil/`, `internal/config/`; rename `domain.X` → `core.X`; adopt Factory + IOStreams + Formatter composition pattern (cli-functional-core-shell)

## [0.12.0] - 2026-09-19

### Added

- Short flag `-m` for `delete --merged-only`
- Short flag `-d` for `prune --delete-branches`
- Preview of affected worktrees before confirmation in `prune --all`
- Configurable `Git.CLITimeout` and `Shell.HookTimeout` (replace previously hardcoded values)

### Changed

- `create` falls back to `config.Validation.DefaultSourceBranch` when `--source` is not specified
- `cd` not-found error includes the requested target and project context for easier diagnosis
- `check` mise task renamed to `verify` and now includes `build` step
- Release validation consolidated into shared library: `release:validate` replaced by `release:check` which supports initial release setup and dirty `CHANGELOG.md`

### Fixed

- Race condition in `PruneMergedWorktrees` result slice under concurrent pruning
- Nil `req.Context` panic risk in worktree creation
- Empty `ResolvedPath` panic risk in delete

## [0.11.0] - 2026-03-30

### Changed

- Centralized all interface definitions in application package (consolidated from domain/ and infrastructure/)
- Service layer now depends on interfaces from application/ following dependency inversion principle

### Added

- compile-time interface satisfaction checks to all implementations
- doc.go package documentation files
- govulncheck to pre-commit for vulnerability scanning
- depguard linter with domain package isolation rules

## [0.10.0] - 2026-03-18

### Added

- Golden file testing infrastructure for snapshot testing of list and error output
- Explicit error formatter strategy pattern with `errors.As()` matching
- Automatic cleanup registration for `RepoTestHelper` using `t.Cleanup()`
- Environment variable expansion in config paths (`$HOME` and `~` support)

### Changed

- Converted all layer tests (service, infrastructure, domain) from testify/suite to standard Go testing with `t.Run()` and `t.Cleanup()`
- Error formatter refactored to explicit strategy pattern with ordered matcher-formatter slice
- Build configuration updated to use `internal/version` package
- Go bumped to 1.26.1

### Fixed

- Golden file comparison normalization with TrimSpace on both actual and expected output
- Test helpers properly respect `HOME` env var for test isolation

## [0.9.2] - 2026-03-16

### Added

- Environment variable expansion in config paths (`ProjectsDirectory`, `WorktreesDirectory`, `BackupDir`)
- Support for `~` (tilde) expansion in config paths

## [0.9.1] - 2026-03-16

### Changed

- Go upgraded to 1.26.1

### Fixed

- E2E tests to respect HOME and use explicit config paths
- Path validation to ensure config paths stay under home directory

## [0.9.0] - 2026-03-16

### Added

- Command aliases: `ls` for `list`, `rm` for `delete`
- Short flags: `-a` for `list --all`, `-y/--yes` for `prune` auto-confirmation
- JSON output via `--output/-o` flag on `list` command
- Quiet mode with global `--quiet/-q` flag
- Progress reporting for `prune` command
- Shell completion enhancements: fuzzy matching, smart sorting, status indicators
- Granular exit codes: ExitCodeConfig (3), ExitCodeGit (4), ExitCodeValidation (5), ExitCodeNotFound (6)
- Panic recovery in main.go with user-friendly error messages
- Help text improvements with examples sections

### Changed

- Error messages simplified - removed internal operation names
- Mise tasks refactored to use explicit `run` arrays

### Fixed

- E2E test failures from output-scripting change
- Test isolation issues with HOME env var
- Cross-project branch completion

## [0.8.1] - 2026-03-12

### Changed

- `init` command default behavior changed from file installation to stdout output
  - Default mode: Print shell wrapper to stdout (eval-safe)
  - `-i, --install` flag enables file installation mode
  - `-c, --config` flag for custom config file path

## [0.8.0] - 2026-03-12

### Added

- Shell plugins for zsh (Oh My Zsh, antidote, zinit, znap with lazy-load completions)
- Shell plugins for bash (standalone and Bash-It integration)
- Shell plugins for fish (conf.d and Oh My Fish integration)
- Post-create hook execution for worktree setup
- HookRunner infrastructure to execute commands from `.twiggit.toml`
- Graceful interrupt handling for OpenSpec autonomous workflow

### Changed

- Pre-commit hooks refactored to use direct tool invocation
- GitLab CI pipeline optimized to prevent duplicate runs
- Validation jobs now run only on MRs and tags

### Fixed

- Completion command delegation to Carapace
- Duplicate brackets in shell wrapper template
- golangci-lint configuration for E2E build tags

## [0.7.0] - 2026-02-17

### Added

- Carapace integration for shell completion via `_carapace` command
- Configurable completion timeout (500ms default)
- LRU cache (25 repos) for git repositories
- TTL-based cache for git worktree validation
- ProjectSummary type for lightweight project enumeration
- `ListProjectSummaries` method on ProjectService

### Changed

- Dependencies updated: go-git v5.16.5, koanf v2.3.2, ginkgo v2.28.1, gomega v1.39.1
- Removed visible completion command (use `_carapace` instead)
- Version package consolidated into cmd layer

### Fixed

- Branch description tagging loop index bug
- Parent directory comparison edge case in IsPathUnder
- Error detection using typed ConfigError instead of string matching

## [0.6.0] - 2026-02-13

### Added

- Prune command for merged worktree cleanup with `--dry-run`, `--force`, `--delete-branches`, `--all` flags
- Protected branch protection (main, master, develop, staging, production)
- Cross-project pruning with confirmation prompt
- Navigation path output for shell integration
- OpenSpec extended skills for artifact management

## [0.5.6] - 2026-02-12

### Added

- Comprehensive codebase quality audit skill with modular patterns
- Depth levels: Minimal/Standard/Detailed/Maximum report granularity

### Changed

- Codebase quality refactored for modularity and flexibility

## [0.5.5] - 2026-02-12

### Added

- Standardized mock pattern with testify/mock (`.On()`/`.Return()`)
- Error handler tests for CLI error handling
- Git client tests for CompositeGitClient routing

### Changed

- Error handling standardized across domain/infrastructure/service layers with ValidationError types
- Test patterns unified to use Testify suites
- CLI flags standardized (`-C` for `--cd`, `-f` for `--force`)
- Infrastructure package flattened
- GoReleaser speed optimizations

### Removed

- Unused caching layer in ContextDetector
- Unused interfaces (GitRepositoryInterface, ProjectRepository)

## [0.5.4] - 2026-02-11

### Changed

- GoReleaser: Disabled Go module proxy

## [0.5.3] - 2026-02-11

### Added

- HTML coverage reports (`coverage.html`)
- XML coverage reports in Cobertura format (`coverage.xml`)

### Changed

- Gitignore updated for coverage file patterns

## [0.5.2] - 2026-02-11

### Added

- Shell auto-detection in `init` command from SHELL environment variable
- `init` command renamed from `setup-shell`
- `--check` flag to validate wrapper installation
- Verbose output with `-v` and `-vv` flags
- Interactive shell completions for bash, zsh, fish

### Changed

- `setup-shell` renamed to `init` command

## [0.5.1] - 2026-02-09

### Added

- Location-specific AGENTS.md documentation (cmd/, internal/, test/)
- OpenSpec workflow with specifications and change tracking
- Pre-commit hooks with worktree automation

### Changed

- Removed centralized `.ai/` directory in favor of co-located AGENTS.md files

## [0.4.0] - 2026-02-09

### Added

- Shell integration system with `setup-shell` command (bash, zsh, fish support)
- Context-aware navigation system with project/worktree detection
- `default_source_branch` configuration option
- `--source` flag for `create` command
- Unified error handling with structured suggestions

### Changed

- Configuration format migrated from YAML to TOML
- CLI command `switch` renamed to `cd`
- Shell integration enhanced with zfunctions support for zsh

### Fixed

- Shell wrapper functions properly preserve exit codes

### Security

- Path traversal protection against directory escape attacks

## [0.3.0] - 2026-02-09

### Changed

- Domain layer refactored to pure business logic
- Architecture properly separates domain and infrastructure layers
- Filesystem abstraction with dependency injection
- Infrastructure abstraction layer with mockable interfaces

## [0.2.0] - 2026-02-09

### Added

- Dependency injection container in `internal/infrastructure/deps.go`
- CLI installation task

### Changed

- All CLI commands refactored to accept dependencies
- Architecture refactored to use dependency injection pattern
- Native git command execution moved to service layer

## [0.1.13] - 2026-02-09

### Changed

- Docker support and modernized CI/CD pipeline
- Docker Buildx for multi-platform builds

### Fixed

- GitLab CI pipeline to prevent duplicate runs

## [0.1.7] - 2026-01-21

### Added

- Delete command for worktree removal
- Improved create command UX with automatic branch creation

## [0.1.3] - 2026-01-19

### Added

- Unified status and list commands
- Bare repository filtering to prevent discovery failures

## [0.1.0] - 2025-09-18 - Initial Release

### Added

- Core worktree management commands: `create`, `delete`, `list`, `cd`, `prune`
- Context-aware operation (auto-detects current project/worktree)
- Shell integration for directory navigation (`twiggit cd`)
- Comprehensive test coverage (unit, integration, E2E)
- Docker support and GitLab CI/CD pipeline
- GitHub mirroring capability
