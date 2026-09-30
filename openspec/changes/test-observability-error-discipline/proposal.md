# Proposal

## Why

The 2026-09-25 codebase audit (`REVIEW.md`, 311 findings, 27 critical) revealed three intertwined regressions: (1) fourteen mock-based tests in `internal/git` register expectations but never call `AssertExpectations`, silently passing even when production never invokes the mock; (2) the Logger surface is split between `slog.SetDefault` in `main.go` and `Factory.Logger` (wired to `io.Discard`), so the documented `TWIGGIT_DEBUG` contract for project code is dead; (3) error handling has lost `*exec.ExitError` cause chains at six `NewWorktreeError`/`NewBranchError` call sites, two adapter functions log AND return the same error (single-handling rule violation), and one cmd-side call still string-matches `oe.Op == "shell.already_installed"` against a sentinel-set that was removed. Each item is mechanically small; together they break the confidence the test suite and debug logger are supposed to provide, and they violate existing spec contracts that the project has otherwise treated as canonical.

## What Changes

- **Unify the debug logger into a single channel.** `iostreams.NewLogger(io.Writer)` becomes the one constructor. `Factory.Logger` returns the same `*slog.Logger` as `IOStreams.Logger`; `main.go` calls `slog.SetDefault` on that instance exactly once. The handler writes to `ios.ErrOut` (not `os.Stderr` directly) so test buffers and color/TTY gating work uniformly.
- **Restore the single-handling rule in `internal/git` adapters.** `detectOpError` (context detection) and `discoverProjects` (cross-project resolver) stop calling `slog.Error`; they return the wrapped `*core.OperationError` only. Debug logging moves to the cmd boundary via `opts.IO.Logger.With("command", ...)`.
- **Preserve `*exec.ExitError` cause through every adapter failure.** `git.NewCommandError` and the six `NewWorktreeError`/`NewBranchError` call sites in `internal/git/writer.go` pass the underlying `err` (or `result.Err`) as the `Cause` field, not `nil`. A new test asserts `errors.As(err, &*exec.ExitError)` walks the chain.
- **Restore five shell-error sentinels in `internal/core/shell_errors.go`.** `ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`. Each is matched via `errors.Is` from cmd code, including `cmd/init.go:222` which currently string-matches `oe.Op`.
- **Drop the dead error-classification code in `internal/git/errors.go`.** `ErrorKindPermission` is never produced by `classifyKind`; `ErrorKindNotFound` is never set by any caller (the `Is()` method that consults it is therefore unreachable); the entire `Is(target error) bool` method is removed. NotFound detection already walks the cause chain through `*core.NotFoundError.Is()` per the baseline spec.
- **Fix test discipline.** Fourteen tests add `t.Cleanup(func() { mock.AssertExpectations(t) })` (twelve in `internal/git/writer_test.go`, two in `internal/git/hook_runner_test.go`). Two test files (`test/integration/config_test.go`, `internal/config/manager_test.go`) migrate `os.Setenv`/`os.Unsetenv` to `t.Setenv`. `test/concurrent/concurrent_test.go` adds `TestMain(m)` with `goleak.VerifyTestMain(m)` (strict, no `IgnoreTopFunction`). Suite methods switch `require(s.T(), ...)` to `s.Require()`, `assert(s.T(), ...)` to `s.Assert()`, and the shadowed `require := require.New(t)` renames to `must := require.New(t)` per `golang-stretchr-testify` convention.
- **Replace the `fmt.Fprintf(os.Stderr, ...)` parse-failure warning in `internal/git/hook_runner.go` with `slog.Warn`** through the unified logger channel.
- **Stop double-wrapping `OnceValues` errors in `Factory.Init`.** The `fmt.Errorf("cmdutil: ...: %w", err)` wraps are removed; cached errors flow through `errors.Join` directly. Each init failure carries a single wrapper, not three.
- **Document the debug ergonomics.** Add a one-paragraph `GODEBUG` section to `README.md` (`gctrace=1`, `schedtrace=10000`, `asyncpreemptoff=1`) and a `.vscode/launch.json` Delve config for the built binary.

No public CLI surface changes. No exit-code contract changes. No new dependencies (`go.uber.org/goleak` is the only new transitive test dependency).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-iostreams`: requirement 5 (`Logger` gating and level) is rewritten to specify that `iostreams.NewLogger(io.Writer)` is the single construction point, the handler writes to `IOStreams.ErrOut`, and the cmdutil command factory's logger accessor, the default `slog` logger, and `IOStreams.Logger` are the same `*slog.Logger` instance (pointer-equality test). The `Test()` returns-buffers requirement is rewritten alongside so the test path also constructs the singleton and exposes a non-nil Logger.
- `domain-typed-errors`: requirement 8 (sentinel catalog) is amended to restore five shell-error sentinels (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`) alongside the four NotFound sentinels, with the canonical `errors.Is` matching pattern that `cmd/init.go` already documents in code-comment but does not yet implement.

## Non-Goals

- No change to the public CLI surface, exit-code contract (`cli-exit-codes`), or `cli-error-formatting` dispatch.
- No `ValidationError` cause field; spec `core-errors` `req 1` keeps `Unwrap() error` returning `nil`.
- No `Factory` per-role accessor collapse (Change E).
- No `errors_legacy.go` rename (Change E).
- No `cmd.OutOrStdout()` migration (Change D).
- No linter config additions (`nolintlint`, `gocritic`, `testifylint`) — Change A.
- No module path rename (`twiggit` → `github.com/amoconst/twiggit`) — deferred.

## Risks

- **Logger writer change is observable in test stderr capture** — tests that previously relied on debug output going to the host's `os.Stderr` will now see it in the test buffer instead. → Mitigation: tests that assert on debug output must read `iostreams.Test()`'s stderr buffer, not `os.Stderr`.
- **Five new sentinels expand the package surface** — the prior archive deliberately removed sentinels to force matching through `errors.Is` against a typed error. → Mitigation: sentinels ARE the `errors.Is` contract; restoring them enables it. The seven subtype structs stay removed per design.md §Decision 4.
- **Strict goleak will flag any test that leaks goroutines by accident**, including any third-party goroutine started indirectly. → Mitigation: when a test fails on goleak output, identify the stack and either fix the leak or add a targeted `IgnoreTopFunction` for that specific stack only. No blanket ignores.
- **Cause-chain test (`errors.As` on `*exec.ExitError`) is sensitive to `command_executor.go` preserving the error** — if a future change drops the cause again, the test fails loudly. → Mitigation: the test is part of the verifier gate (`go test -run TestNonZeroExit ./internal/git/...`).
- **`Factory.Init` no longer prefixes init failures with `cmdutil:`** — a script or test that greps stderr for `cmdutil:` will break. → Mitigation: no such consumer exists in the repo; the underlying `*core.OperationError` already exposes `Op` which carries the same diagnostic value.
- **`ErrorKindPermission` and `ErrorKindNotFound` removal is a no-op today** — but could surprise a future caller expecting the enum to include either class. → Mitigation: single-file deletions in `internal/git/errors.go`; the diff is reviewable. NotFound dispatch still works through `*core.NotFoundError.Is()` walking the cause chain, which the baseline spec already mandates.
- **`slog.Default()` test-buffer bypass in `internal/git/hook_runner.go`** — the hook production-write warning is logged via `slog.Default()` (production code over test-buffer observability), so test-side assertions on this path read host stderr, not the test buffer. → Accepted gap; tests that need to assert on this path are authored as E2E tests with stderr captured at the process boundary.

## Impact

- **Source files (production):**
  - `internal/iostreams/iostreams.go` — `NewLogger` signature + writer plumbing
  - `internal/cmdutil/factory.go` — `Logger` field returns singleton; `Init` drops double-wrap
  - `main.go` — `slog.SetDefault` once; panic recover unchanged
  - `internal/git/writer.go` — six cause-chain sites
  - `internal/git/command_executor.go` — `NewCommandError` cause preserved
  - `internal/git/context_detector.go` — drop `slog.Error`
  - `internal/git/context_resolver.go` — drop `slog.Error`
  - `internal/git/hook_runner.go` — `slog.Warn` instead of `fmt.Fprintf`
  - `internal/git/errors.go` — drop `ErrorKindPermission`, drop `ErrorKindNotFound`, remove the `Is(target error) bool` method
  - `internal/core/shell_errors.go` — five sentinels added back
  - `cmd/init.go` — string-match replaced with `errors.Is`
- **Source files (tests):**
  - `internal/git/writer_test.go` — twelve `t.Cleanup` additions + new `TestCLIClient_NonZeroExit_PreservesExecError`
  - `internal/git/hook_runner_test.go` — two `t.Cleanup` additions
  - `test/integration/config_test.go` — `t.Setenv` migration
  - `internal/config/manager_test.go` — `t.Setenv` migration
  - `test/concurrent/concurrent_test.go` — `TestMain` with `goleak.VerifyTestMain(m)`
  - `internal/core/{hook_types,git_types,service_results}_test.go` — `require` shadow renamed to `must`
- **Spec deltas:** two files under `openspec/changes/test-observability-error-discipline/specs/`.
- **Documentation / IDE:** `README.md` (GODEBUG section, ~10 lines); `.vscode/launch.json` (new, ~30 lines).
- **Dependencies:** `go.uber.org/goleak` (test-only indirect, already used by many Go projects; safe across all supported platforms).
- **CI:** no change; `mise run verify` already covers the touched layers.
- **Breaking changes:** none externally. The five restored shell sentinels are additive. The Logger writer change is internal: callers that constructed `IOStreams.Logger` directly still get the gated handler; the only observable difference is that test-mode stderr capture now sees the debug output if `TWIGGIT_DEBUG=1` is set in the test process.
