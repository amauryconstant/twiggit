# Proposal: Quality Modernization

## Why

The 19-skill audit in `REVIEW.md` surfaced ~400 findings across the
codebase. `foundation-error-chain` ships ~46 of them; `architecture-layer-inversion`
ships ~73. Roughly 580 findings remain — clustered around test reliability,
mechanical modernization, lint threshold tightening, dead-code sweep,
godoc coverage, dependency hygiene, and CI/Docker/Goreleaser hardening.
This change is the third wave of a sequential three-change sequence
(1 = error chain, 2 = layer inversion, 3 = quality modernization) and
lands the remainder on a stable post-Ch2 baseline. It is the highest-leverage
slice of the sequence because the test reliability work in this change
makes every future change safer to land.

## What Changes

- **Test reliability.** Adopt `t.Cleanup(mock.AssertExpectations)` as a
  universal mock-fixture contract; migrate 153 `context.Background()`
  literals in test files to `t.Context()`; replace `os.Chdir`+`defer
  os.Chdir` (14 sites across 4 files) with `t.Chdir`; drop `.Maybe()`
  audit (26 sites) in favor of explicit `t.Cleanup` verification;
  introduce a build-once `TestMain` in the binary tests so the binary
  compiles once instead of per test; add `goleak.VerifyTestMain`;
  migrate 7 `wg.Add(1); go func(); defer wg.Done()` blocks to
  `wg.Go` (Go 1.25); rename 2 local `func max` shadowing the
  builtin (1.21+) to `maxInt`; convert 3 `panic()` calls in
  `test/helpers/` to `t.Fatal` (signature change); absorb Change 1's
  missing `test/mocks/helpers.go` plus the `cmd_mocks.go` variadic fix
  (still unlanded); drop 9 `var _ interface{}` declarations and 4
  silent-pass `assert.NoError` + `t.Run` patterns.
- **Mock hygiene.** Drop the `ExpectedCalls = nil` reset pattern
  (15 sites) in favor of fresh mocks per subtest; remove the 4
  redundant `//nolint:wrapcheck` directives in the mock test file
  already excluded for `_test.go` files; tighten the mock import
  helpers so every `.On(...)` call in the test tree (~287 sites)
  has a paired `t.Cleanup(mock.AssertExpectations)`.
- **Modernization sweep.** Migrate 14 remaining `os.IsNotExist` calls
  to `errors.Is(err, os.ErrNotExist)` (Change 2 deletes the file
  that contained the highest-concentration sites); replace 3
  `strings.HasPrefix + strings.TrimPrefix` pairs with
  `strings.CutPrefix`; convert 16 `for i := 0; i < N; i++` to
  `for i := range N`; switch `interface{}` to `any` at 9 sites
  including the `SuggestionOption` callback type; thread the caller
  `ctx` through 4 `context.Background()` sites in
  `infrastructure/context_resolver.go`; split the 53-line
  `parseWorktreeList` state machine into a small typed parser;
  preallocate 6 nil-then-append slices via `make`.
- **CLI plumbing.** Wire `cmd.Context()` instead of
  `context.Background()` in every leaf `executeX` (12 sites across
  the cmd layer); set `SilenceUsage` and `SilenceErrors` on every
  leaf command (only `create` currently sets both); replace direct
  `os.Stderr` writes (`fmt.Fprintf`, `logv`) with `c.ErrOrStderr()`
  and accept `io.Writer` for injection in the `logv` helper;
  wrap the cobra command tree with `signal.NotifyContext` in
  `main.go` so SIGINT/SIGTERM propagate to `cmd.Context()`; switch
  the `main.go` panic recovery to `slog.Error` with
  `debug.Stack()`; consume the exit code returned by
  `HandleCLIError` instead of hardcoding `os.Exit(1)`.
- **Slog migration (extension).** Replace 2 silent skip-and-continue
  sites in `project_service.go` (ListProjects / ListProjectSummaries
  error-swallowing) with `slog.Error` (Change 2 covered
  `worktree_service.go`; this change extends the pattern to the
  project service); accept `io.Writer` in `cmd/util.go` `logv`
  so the cobra command can inject `c.ErrOrStderr()`.
- **Lint, naming, style.** Tighten `funlen.lines` 150→120 and
  `gocyclo.min-complexity` 25→13, staged per file in the same
  commit that fixes the function (no wholesale lower + grandfathered
  `//nolint`); convert 12 4-string-parameter constructor signatures
  in `internal/domain/errors.go` to struct literals; convert 25
  receiver-less methods in `internal/service/` to free functions
  (Fowler-shaped Extract Function); sweep dead code per
  `openspec/dead-code.md`; eliminate 42 `_, _ = fmt.Fprint*`
  discarded writes in the cmd layer by per-site classification
  (5 categories: navigation, warnings, progress, result detail,
  TTY prompt — each with a different propagation rule).
- **Documentation.** Add godoc comments to 33 exported methods/funcs
  (26 `Error()`/`Unwrap()` in domain errors, 4 exported
  infrastructure methods, 3 nit sites including `version.go`);
  add `llms.txt` at the repo root; repair README broken
  `twiggit init` snippets, the Go Report Card badge, the missing
  Demo/Features/Contributing/License sections; normalize CHANGELOG
  double blank lines; switch `internal/version/version.go` to
  ldflag-set unexported vars + `String()` accessor that trims
  trailing space; fix the Go version mismatch between
  `CONTRIBUTING.md` and the CHANGELOG.
- **Dependencies.** Add the `tool` directive for
  `govulncheck`; `go get -u=patch` `go-git`, `testify`, the
  `golang.org/x/{crypto,net,sys,text,sync}` packages; drop the
  duplicate `gopkg.in/yaml.v3` fork (keep `go.yaml.in/yaml/v3`
  only — the project resolves through `knadh/koanf`); add a
  `retract` block; pin the toolchain explicitly; `go mod tidy`.
- **CI / Docker / Goreleaser / Mise.** Add `govulncheck` step with
  baseline `govulncheck.json` artifact; wire `-race` into the test
  job; replace the warning-only coverage check with a non-zero
  exit below the threshold; add main-branch rules on lint/test/setup
  so direct pushes don't bypass cache + validate; pin the
  Dockerfile.ci base image by sha256; remove the `curl | sh`
  unpinned mise installer (download release tarball + sha256
  verify); fix the goreleaser GitHub owner mismatch
  (`.goreleaser.yml` says `amoconst`; the hook script targets
  `amauryconstant/twiggit` — the GitHub mirror lives at
  `github.com/amauryconstant/twiggit`, so change the goreleaser
  owner to `amauryconstant`); replace `|| echo` swallow in the
  goreleaser hook with `set -e` and explicit error; add
  `.github/dependabot.yml` and `.github/CODEOWNERS`; convert the
  mise `coverage.sh` warn-only to fail-below-threshold; add
  `-shuffle=on` to `test:unit`, `test:race`, and `test:e2e`
  tasks.

## Capabilities

### New Capabilities

- `cli-signal-handling`: cobra command tree wraps
  `signal.NotifyContext`; SIGINT and SIGTERM cancel
  `cmd.Context()`; cleanup of TTY-bound stdin readers
  (`confirmBulkPrune`) gates on `isatty` before reading; the
  TTY-gated path documents that pipes hang without
  `--yes`/`--force`.
- `domain-shell-wrappers`: pure domain function
  `ShellWrapper(shellType)` returning the eval-safe wrapper
  template; test fixtures call this instead of constructing a
  real `ShellInfrastructure` to seed mock return values; no
  infrastructure import from service tests for wrapper content.

### Modified Capabilities

- `cli-command-options-pattern`: add requirement that every leaf
  command sets `SilenceUsage = true` and `SilenceErrors = true`,
  so a `RunE` failure does not print the full usage block.
- `cli-quiet-mode`: add requirement that `logv` accepts an
  `io.Writer` parameter so tests can assert without writing to
  `os.Stderr`; the writer defaults to `c.ErrOrStderr()` when
  called from a cobra command.
- `cli-main-entry-point`: add scenarios for `signal.NotifyContext`
  cancellation, for slog-based panic recovery with `debug.Stack()`,
  and for the consumed exit code from `HandleCLIError` (rather
  than hardcoded `os.Exit(1)`).
- `domain-config-types`: scenarios for the deletion of the dead
  `ServiceConfig` fields (`CacheEnabled`, `ConcurrentOps`,
  `MaxConcurrent`, `Shell.Timeout`) — they remain in the struct
  for backward-compat read but are marked deprecated and ignored
  on load; add a scenario for the new `tool govulncheck`
  directive location.
- `infrastructure-config-manager`: add scenario that config-file
  writes use `os.Root` for path confinement when writing under
  `$HOME` (Go 1.24+); add scenario that `ProtectedBranches`
  is cloned via `slices.Clone` to avoid sharing the backing
  array with the caller.
- `infrastructure-context-resolver`: add scenarios for caller-`ctx`
  propagation through `gitService.ListWorktrees`,
  `GetRepositoryStatus`, `ListBranches`, and the
  `getWorktreeContextSuggestions` paths; document the 4 sites that
  previously held `context.Background()` and now thread the
  parent `ctx`.
- `infrastructure-git-client`: add scenario for bounded hash
  slicing (`min(7, len(hashStr))` instead of `hashStr[:7]`).
- `infrastructure-hook-runner`: add scenario that hook execution
  warnings propagate through `slog.Warn` rather than direct
  `os.Stderr` writes; document the env-var injection via
  `cmd.Env` instead of `sh -c "<exports>; <cmd>"` string concat.
- `infrastructure-release`: add scenarios for the govulncheck
  baseline (`govulncheck.json` committed; CI step uploads the
  diff); the `-race` test job requirement; the coverage
  threshold gate (non-zero exit below threshold); the
  digest-pinned Dockerfile.ci base image; the sha256-verified
  mise tarball installer; the goreleaser owner alignment
  (`amauryconstant` for GitHub artifacts); the `.github/dependabot.yml`
  config; the resource-group sharing across lint/test/setup jobs.
- `application-shell-service`: add requirement that shell-detection
  errors surface via `errors.Is(err, domain.ErrShellInferenceFailed)`
  so the cmd layer's typed-error dispatch (Change 1) classifies
  the failure correctly; document the absence of string-equality
  fallback paths.
- `application-worktree-management`: add requirement that
  `PruneWorktrees` errors are logged via `slog` before the
  branch delete proceeds, so a partial-failure state is visible
  in the result.
- `testing-e2e`: add scenarios for the build-once `TestMain`
  pattern (binary compiled into `t.TempDir()`, path exported
  via env var); the `-shuffle=on` test flag; the
  `goleak.VerifyTestMain` setup; the `t.Cleanup(mock.AssertExpectations)`
  fixture used by every mock-based test.
- `testing-golden-file-testing`: add scenario that the golden
  helper types its own `maxInt` rather than shadowing the
  Go 1.21+ builtin `max`.
- `testing-helpers`: add scenario that helper fixtures
  (`test/helpers/repo.go`, `test/helpers/git.go`) use `t.Fatal`
  instead of `panic` for precondition failures; signature change
  ripples to callers.

(13 modified + 2 new = 15 spec deltas total. The earlier estimate
of "13 total" undercounted the modified set; this number is the
real count after auditing every requirement that this change
modifies.)

## Impact

| Layer | Files | LOC delta |
|---|---|---|
| `cmd/` | `cd`, `create`, `delete`, `init`, `list`, `prune`, `util`, `version`, `completion`, `root`, `output`, `error_handler`, `error_formatter`, `suggestions` (~14 files) | +700 / -300 |
| `main.go` | 1 file | +20 / -10 |
| `internal/service/` | `worktree_service.go`, `project_service.go`, `context_service.go`, `shell_service.go` (4 files) | +600 / -400 (incl. 25 free-function refactor) |
| `internal/infrastructure/` | `cli_client.go`, `gogit_client.go`, `command_executor.go`, `context_detector.go`, `context_resolver.go`, `config_manager.go`, `hook_runner.go`, `shell_infra.go` (8 files) | +400 / -200 |
| `internal/domain/` | `errors.go`, `service_errors.go`, `shell_errors.go`, `shell_wrapper.go` (new), `pathutils.go`, `git_repo.go`, `git_types.go`, `hook_types.go`, `shell_results.go`, `service_results.go`, `version.go` (11 files) | +300 / -150 |
| `test/` | `test/mocks/helpers.go` (new), `test/mocks/cmd_mocks.go`, every `*_test.go` that uses mocks (~25 files), `test/helpers/*`, `test/integration/*`, `test/concurrent/*`, `test/e2e/fixtures/*` (~30 files) | +1,200 / -800 |
| `internal/application/` | `interfaces.go`, `AGENTS.md` | +50 / -10 |
| `AGENTS.md` (root + per-layer) | 5 files | +200 / -150 |
| `.golangci.yml`, `.gitlab-ci.yml`, `Dockerfile.ci`, `.goreleaser.yml`, `.goreleaser/hooks/*`, `.mise/tasks/ci/*`, `.mise/config.toml`, `go.mod`, `go.sum`, `.github/dependabot.yml` (new), `.github/CODEWNERS` (new), `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `llms.txt` (new) | 16 files | +400 / -150 |
| Specs | 13 modified + 2 new = 15 spec deltas via `openspec/specs/` | — |
| **Total** | **~115 files** | **+3,870 / -2,170 (net +1,700)** |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged (exit codes preserved; navigation paths still go to stdout; verbose output still gates on `--verbose`) |
| Public API of `internal/domain` | Adds `domain.ShellWrapper`; no removals |
| Public API of `internal/application` | No changes |
| Public API of `test/mocks` | Adds `test/mocks/helpers.go` `variadicArgs`; fixes the variadic mismatch on `MockContextService` |
| Tests | Universal mock cleanup fixture; build-once TestMain; `-shuffle=on` exposes previously-hidden order-dependent bugs in the 14 chdir sites and the 7 wg blocks; mock matcher swap (`context.Background()` → `mock.Anything`) may surface 10-30 previously-silent-pass tests |
| CI | New govulncheck baseline; -race in test job; coverage threshold gate; digest-pinned base; sha256-verified installer; goreleaser owner aligned with the actual GitHub mirror; Dependabot config committed |
| Spec deltas | 15 total (13 modified + 2 new) |
| Version | No version bump — Change 3 ships at the version of the most recent tagged release (0.12.0); the next user-visible release bumps from this baseline |
| Release sequencing | This change lands AFTER `foundation-error-chain` AND `architecture-layer-inversion` ship. Sequential, not parallel. Both prior changes must be merged first so this change's baseline includes the new error sentinels, the layer inversion, the `*Impl` renames, the LRU cache swap, and the depguard expansion. |

## Non-goals

- **Wholesale Tier 2 CLI collapse** (`main.go + cmd/ + internal/{core,
  git, output, iostreams, cmdutil}/`). Change 2 design §"Non-Goals"
  deferred this; Change 3 keeps the existing 5-layer convention.
- **Interface Segregation refactor** (split `GoGitClient` (8 methods)
  and `CLIClient` (6 methods) into 1-method role interfaces). The
  change keeps the role-interface shape Change 2 introduced; per-method
  roles would force 7+ spec migrations and a wholesale test rewrite.
- **Go toolchain major bump** (1.25.5 → 1.26+). Out of scope; the
  `tool` directive pins 1.25.5 explicitly. The `errors.AsType[T]`
  from Go 1.26 is unavailable; local helpers from Change 1 are
  sized to swap mechanically.
- **AI-driven PR review tooling** (Claude Code review, Copilot).
  Mentioned in REVIEW §18; deferred to a separate tooling change.
- **Codecov integration**. REVIEW §18 flagged absence; deferred
  pending a project-side threshold discussion.
- **Go version matrix** (CI matrix across `[1.25, 1.26, stable]`).
  Deferred until the toolchain bump lands.
- **The `simplify` range-over-int deduplication in the e2e retry
  fixture** (`test/e2e/fixtures/e2e_fixtures.go:176`) — flagged as
  a "synctest.Test (1.25) for determinism" candidate but optional.

## Captured decisions (from exploration)

| Decision | Choice | Rationale |
|---|---|---|
| Scope shape | Single change, 17 slices, tests-first ordering | User accepted; test reliability dominates change value |
| Tests-first ordering | Slices S4-S7 (test/) before S8-S13 (production) | Tests stabilize first; production modernization has a green baseline |
| Lint threshold strategy | Stage per-file in same commit as the fix | No wholesale lower + grandfathered `//nolint` |
| `t.Context()` migration | All 153 sites, single slice, three sub-buckets | B-a (140 in-fn literals), B-b (5 helper signatures), B-c (~30 mock matchers) |
| `_, _ = fmt.Fprint*` | Per-site 5-category classification (28 sites) + mechanical `cmd.ErrOrStderr` for logv (14 sites) | Different propagation rules per category |
| Goreleaser owner | Change `.goreleaser.yml:73` `amoconst` → `amauryconstant` | GitHub mirror is `amauryconstant/twiggit`; hook script already targets `amauryconstant/twiggit`; only goreleaser.yml is wrong |
| Spec deltas | 13 modified + 2 new = 15 total | `application-cmdutil` folds into `application-service-interfaces` (no new spec) |
| Version | No bump | Pure quality change; defer to next user-visible release |
| Change ordering | Sequential: foundation-error-chain → architecture-layer-inversion → quality-modernization | Both prior changes must ship first; this change's baseline depends on them |

## Open items

- The `openspec/config.yaml` parse error at line 134 (the embedded
  comment inside the `specs:` list item) persists. `openspec new
  change` printed a warning; `openspec validate` may misbehave until
  fixed. The fix is the same line move Change 1 task 1.2 intended;
  since Change 1 lands first, this is presumed fixed before
  Change 3 begins implementation. If still present at implementation
  start, fix as part of S1 (Setup).
- `test/mocks/helpers.go` does not exist yet (Change 1 task 5.1
  was scaffolded in design but never landed). Slice S1 absorbs
  this. If Change 1 is going to land first, its task 5.1 should
  also be completed; if not, S1 of this change creates the file.
- govulncheck first run on the current `go-git v5.16.5`,
  `mergo v1.0.2`, `x/crypto v0.48.0`, `x/sys v0.41.0` baseline
  will likely surface findings. The CI step is added AFTER the
  `go get -u=patch` step (slice S15); the baseline JSON is
  captured at the same commit. Any findings remaining after the
  patch bump are out of scope (separate CVE work).
