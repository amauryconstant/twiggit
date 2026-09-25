# Proposal

## Why

The 2026-09-25 codebase audit (`REVIEW.md`, 311 findings, 27 critical) revealed three intertwined regressions: (1) fourteen mock-based tests in `internal/git` register expectations but never call `AssertExpectations`, silently passing even when production never invokes the mock; (2) the Logger surface is split between `slog.SetDefault` in `main.go` and `Factory.Logger` (wired to `io.Discard`), so the documented `TWIGGIT_DEBUG` contract for project code is dead; (3) error handling has lost `*exec.ExitError` cause chains at six `NewWorktreeError`/`NewBranchError` call sites, two adapter functions log AND return the same error (single-handling rule violation), and one cmd-side call still string-matches `oe.Op == "shell.already_installed"` against a sentinel-set that was removed. Each item is mechanically small; together they break the confidence the test suite and debug logger are supposed to provide, and they violate existing spec contracts that the project has otherwise treated as canonical.

## What Changes

- **Unify the debug logger into a single channel.** `iostreams.NewLogger(io.Writer)` becomes the one constructor. `Factory.Logger` returns the same `*slog.Logger` as `IOStreams.Logger`; `main.go` calls `slog.SetDefault` on that instance exactly once. The handler writes to `ios.ErrOut` (not `os.Stderr` directly) so test buffers and color/TTY gating work uniformly.
- **Restore the single-handling rule in `internal/git` adapters.** `detectOpError` (context detection) and `discoverProjects` (cross-project resolver) stop calling `slog.Error`; they return the wrapped `*core.OperationError` only. Debug logging moves to the cmd boundary via `opts.IO.Logger.With("command", ...)`.
- **Preserve `*exec.ExitError` cause through every adapter failure.** `git.NewCommandError` and the six `NewWorktreeError`/`NewBranchError` call sites in `internal/git/writer.go` pass the underlying `err` (or `result.Err`) as the `Cause` field, not `nil`. A new test asserts `errors.As(err, &*exec.ExitError)` walks the chain.
- **Restore five shell-error sentinels in `internal/core/shell_errors.go`.** `ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`. Each is matched via `errors.Is` from cmd code, including `cmd/init.go:222` which currently string-matches `oe.Op`.
- **Drop the dead `ErrorKindPermission` constant in `internal/git/errors.go`.** `classifyKind` never produces it; the `os.IsPermission` heuristic is fragile across platforms.
- **Fix test discipline.** Fourteen tests add `t.Cleanup(func() { mock.AssertExpectations(t) })` (twelve in `internal/git/writer_test.go`, two in `internal/git/hook_runner_test.go`). Two test files (`test/integration/config_test.go`, `internal/config/manager_test.go`) migrate `os.Setenv`/`os.Unsetenv` to `t.Setenv`. `test/concurrent/concurrent_test.go` adds `TestMain(m)` with `goleak.VerifyTestMain(m)` (strict, no `IgnoreTopFunction`). Suite methods switch `require(s.T(), ...)` to `s.Require()`, `assert(s.T(), ...)` to `s.Assert()`, and the shadowed `require := require.New(t)` renames to `must := require.New(t)` per `golang-stretchr-testify` convention.
- **Replace the `fmt.Fprintf(os.Stderr, ...)` parse-failure warning in `internal/git/hook_runner.go` with `slog.Warn`** through the unified logger channel.
- **Stop double-wrapping `OnceValues` errors in `Factory.Init`.** The `fmt.Errorf("cmdutil: ...: %w", err)` wraps are removed; cached errors flow through `errors.Join` directly. Each init failure carries a single wrapper, not three.
- **Document the debug ergonomics.** Add a one-paragraph `GODEBUG` section to `README.md` (`gctrace=1`, `schedtrace=10000`, `asyncpreemptoff=1`) and a `.vscode/launch.json` Delve config for the built binary.

No public CLI surface changes. No exit-code contract changes. No new dependencies (`go.uber.org/goleak` is the only new transitive test dependency).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-iostreams`: requirement 5 (`Logger` gating and level) is rewritten to specify that `iostreams.NewLogger(io.Writer)` is the single construction point, the handler writes to `IOStreams.ErrOut`, and `Factory.Logger`, the default `slog` logger, and `IOStreams.Logger` are the same `*slog.Logger` instance (pointer-equality test).
- `domain-typed-errors`: requirement 8 (sentinel catalog) is amended to restore five shell-error sentinels (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`) alongside the four NotFound sentinels, with the canonical `errors.Is` matching pattern that `cmd/init.go` already documents in code-comment but does not yet implement.

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
  - `internal/git/errors.go` — drop `ErrorKindPermission`
  - `internal/core/shell_errors.go` — five sentinels added back
  - `cmd/init.go` — string-match replaced with `errors.Is`
- **Source files (tests):**
  - `internal/git/writer_test.go` — twelve `t.Cleanup` additions + new `TestCLIClient_NonZeroExit_PreservesExecError`
  - `internal/git/hook_runner_test.go` — two `t.Cleanup` additions
  - `test/integration/config_test.go` — `t.Setenv` migration
  - `internal/config/manager_test.go` — `t.Setenv` migration
  - `test/concurrent/concurrent_test.go` — `TestMain` with `goleak.VerifyTestMain(m)`
  - `internal/core/{hook_types,git_types,service_results,service_errors}_test.go` — `require` shadow renamed to `must`
- **Spec deltas:** two files under `openspec/changes/test-observability-error-discipline/specs/`.
- **Documentation / IDE:** `README.md` (GODEBUG section, ~10 lines); `.vscode/launch.json` (new, ~30 lines).
- **Dependencies:** `go.uber.org/goleak` (test-only indirect, already used by many Go projects; safe across all supported platforms).
- **CI:** no change; `mise run verify` already covers the touched layers.
- **Breaking changes:** none externally. The five restored shell sentinels are additive. The Logger writer change is internal: callers that constructed `IOStreams.Logger` directly still get the gated handler; the only observable difference is that test-mode stderr capture now sees the debug output if `TWIGGIT_DEBUG=1` is set in the test process.
