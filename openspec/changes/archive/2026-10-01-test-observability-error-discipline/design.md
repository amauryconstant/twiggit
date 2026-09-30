# Design

## Context

The codebase carries a Tier-2 CLI architecture (factory + lazy deps + 3-code exit contract) but three subsystems have drifted out of their documented contracts:

1. **Logger topology**: `main.go` calls `slog.SetDefault(...)` with a stderr text handler gated by `TWIGGIT_DEBUG`; `iostreams.NewLogger()` independently re-reads `TWIGGIT_DEBUG` and writes a parallel stderr handler; `cmdutil.Factory.Logger` constructs a third handler wired to `io.Discard` at `LevelDebug` regardless of the env var. The `cli-iostreams` spec already mandates a single channel (`req 5`); production code violates it.
2. **Error-chain loss**: `git.NewCommandError` and the six `NewWorktreeError`/`NewBranchError` call sites in `internal/git/writer.go` pass `nil` as `cause` on non-zero-exit. `errors.Is(err, ...)` cannot reach the originating `*exec.ExitError`. Two adapter functions (`detectOpError`, `discoverProjects`) log AND return the wrapped error — single-handling rule violation from `golang-error-handling` rule 7.
3. **Test discipline**: 14 testify mock tests register `.On(...)` calls but never call `AssertExpectations(t)`; tests pass even when production never invokes the mock. Two test files use `os.Setenv`/`os.Unsetenv` with manual restore instead of `t.Setenv`. `test/concurrent/concurrent_test.go` lacks `goleak.VerifyTestMain`. Four `_test.go` files shadow `require` with `require := require.New(t)` instead of the skill-recommended `must`.

The single most consequential file is `cmd/init.go:222`, which string-matches `oe.Op == "shell.already_installed"` — but `domain-typed-errors` spec `req 8` explicitly removed those sentinels in the prior archive, leaving the cmd-side call site in a half-migrated state.

## Goals / Non-Goals

**Goals:**

- Bring `iostreams.NewLogger`, `Factory.Logger`, `main.go`'s `slog.SetDefault`, and `IOStreams.Logger` into one construction chain with one env-var read.
- Restore the cause chain on every git-adapter failure so `errors.Is(err, *exec.ExitError)` and `errors.As(err, &exitErr)` succeed at every layer.
- Restore five shell-error sentinels and replace the one remaining `Op` string-match call site with `errors.Is`.
- Establish the test-verification idiom (`t.Cleanup(mock.AssertExpectations)`, `t.Setenv`, `goleak.VerifyTestMain`, `s.Assert()`/`s.Require()`, `must := require.New(t)`) across the touched test files.
- Remove dead code (`ErrorKindPermission` constant) and double-wrap patterns (`Factory.Init`).
- Document the `TWIGGIT_DEBUG` and `GODEBUG` ergonomics; ship a Delve launch config.

**Non-Goals:**

- No change to the public CLI surface, exit-code contract (`cli-exit-codes`), or `cli-error-formatting` dispatch.
- No `ValidationError` cause field; spec `core-errors` `req 1` keeps `Unwrap() error` returning `nil`.
- No `Factory` per-role accessor collapse (Change E).
- No `errors_legacy.go` rename (Change E).
- No `cmd.OutOrStdout()` migration (Change D).
- No linter config additions (`nolintlint`, `gocritic`, `testifylint`) — Change A.
- No module path rename (`twiggit` → `github.com/amoconst/twiggit`) — deferred.

## Decisions

### 1. Single Logger construction: `iostreams.NewLogger(io.Writer)` with the writer passed at the boundary

**Choice.** `iostreams.NewLogger` gains a single `io.Writer` parameter. `iostreams.System()` passes `os.Stderr`; `iostreams.Test()` passes the test stderr `bytes.Buffer`. `Factory.Logger` returns the same `*slog.Logger` pointer that `IOStreams.Logger` holds. `main.go` calls `slog.SetDefault(factory.Logger())` exactly once before dispatch.

**Rationale.** The `golang-cli/references/logging.md` skill prescribes `slog.New(slog.NewTextHandler(f.IOStreams.ErrOut, ...))` — the writer is `ErrOut`, not `os.Stderr`. The current `iostreams.NewLogger()` writes to `os.Stderr` directly, which bypasses the test-mode capture buffer and the color/TTY gating that `IOStreams.ErrOut` provides. Passing the writer at the construction boundary lets each entry point (System, Test, future constructors) decide what to write through, while the level/env-var logic stays in one place.

**Alternatives considered.**

- Keep `NewLogger()` zero-arg with hardcoded `os.Stderr`: rejected — leaks past `iostreams.Test()` capture; tests cannot observe debug output.
- Drop `Factory.Logger` and let cmd code call `iostreams.NewLogger` directly: rejected — duplicates the construction pattern across every cmd, and the singleton guarantee becomes unenforceable.

### 2. Cause-chain preservation: pass the underlying `*exec.ExitError` (or `result.Err`) into every `ExternalError`-shaped constructor

**Choice.** `git.NewCommandError("non-zero-exit", ..., err)` accepts the `*exec.ExitError` from `c.executor.ExecuteWithTimeout`. The six `NewWorktreeError`/`NewBranchError` call sites at `internal/git/writer.go` lines 137, 183, 206, 230, 257, 283 pass `result.Err` (the `*exec.ExitError` attached to the `CommandResult`) instead of `nil`. The `*core.OperationError` embedded in `ExternalError` carries the same `Cause`, so `errors.As(err, &*exec.ExitError)` and `errors.Is(err, ...)` succeed end-to-end.

**Rationale.** `golang-error-handling` rule 5: "MUST use `errors.Is` for sentinel matching and `errors.As`/`errors.AsType` for typed chain inspection". `golang-cli/references/errors.md` principle 3: "Each layer wraps with its own context — `fmt.Errorf("reading config %s: %w", path, err)` — and the boundary prints the full chain." Without the cause preserved, the formatter cannot show the originating CLI exit code or stderr under `TWIGGIT_DEBUG=1`.

**Alternatives considered.**

- Encode the exit code into the `Message` field of `OperationError`: rejected — breaks the chain contract; callers cannot inspect `*exec.ExitError` programmatically.
- Define a new `git.CommandFailure` typed error carrying `ExitCode`, `Command`, `Args`, `Stderr`, `Cause`: rejected — adds a new external error type for a property the existing `*exec.ExitError` already carries; constructor surface grows.

### 3. Single-handling rule restoration: drop the `slog.Error` call from `detectOpError` and `discoverProjects`

**Choice.** Remove `slog.Error(...)` from `internal/git/context_detector.go:75-86` (`detectOpError`) and `internal/git/context_resolver.go:583` (`discoverProjects`). Both functions return the wrapped `*core.OperationError` only; the cmd layer logs at the boundary via `opts.IO.Logger.With("command", cmd.Name()).Debug("adapter failed", "err", err)` if desired.

**Rationale.** `golang-error-handling` rule 7 ("Errors MUST be either logged OR returned, NEVER both") and `golang-cli/references/errors.md` principle 2 ("Each error is logged or returned, never both"). The current code logs at the adapter and the formatter re-renders at the cmd layer, producing two outputs for one error event. Logging at the boundary uses the existing `IOStreams.Logger` plumbing from Decision 1, including `With(...)` for trace correlation per `golang-cli/references/logging.md` ("`logger := opts.IO.Logger.With("command", cmd.Name())`").

**Alternatives considered.**

- Keep `slog.Error` and suppress the formatter when the logger already saw it: rejected — the formatter cannot know whether the logger fired without a side channel; the single-handling rule is the cleaner invariant.
- Add a per-error "logged" sentinel to the error struct: rejected — couples error types to log state; cross-process serialization breaks.

### 4. Shell sentinels: restore five sentinels, keep the seven subtypes removed

**Choice.** Restore five package-level sentinels in `internal/core/shell_errors.go`: `ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`. Each sentinel is matched from cmd code via `errors.Is`. `cmd/init.go:222` is rewritten from `errors.As(err, &oe) && oe.Op == "shell.already_installed"` to `errors.Is(err, core.ErrShellAlreadyInstalled)`. The seven shell error subtypes (`ShellAlreadyInstalledError`, etc.) removed in the prior archive stay removed; sentinels flow through existing `*core.OperationError` wrappers, not via new concrete struct types.

**Rationale.** `domain-typed-errors` `req 8` and the prior archive both removed these sentinels — but `cmd/init.go:222` was missed in that migration and still string-matches `oe.Op`. The spec rule "Callers SHALL identify these sentinels exclusively through `errors.Is`" cannot be satisfied when the sentinels do not exist. Restoring the sentinels is the minimum change that lets the one surviving call site migrate cleanly. The subtypes stay removed per the same archive — sentinels are a smaller surface and don't reintroduce the per-type `Is()` methods that were also removed.

**Alternatives considered.**

- Add an `Op`-to-sentinel mapping to `*core.OperationError.Is`: rejected — creates an indirection table for one call site; `errors.Is` against a sentinel is more idiomatic.
- Define a typed shell-error interface (`core.ShellError`): rejected — reopens the seven-subtype architecture that the prior archive deliberately closed.
- Leave `cmd/init.go:222` as a string-match: rejected — keeps the explicit spec violation.

### 5. `ErrorKindPermission` removal: drop the dead constant

**Choice.** Remove `ErrorKindPermission` from the `ErrorKind` iota in `internal/git/errors.go`. `classifyKind` keeps the `context.DeadlineExceeded` → `ErrorKindTimeout` mapping; everything else stays `ErrorKindOther`.

**Rationale.** The constant is declared but never assigned. The `os.IsPermission()` heuristic that would produce it is fragile across platforms (no equivalent on Windows without `golang.org/x/sys/windows`), and the typed `OperationError.Is` mapping from `domain-typed-errors` already gives callers a path to permission failures via custom sentinels if they need one. YAGNI applies.

**Alternatives considered.**

- Wire `os.IsPermission` plus a Windows-specific branch: rejected — adds a platform fork for a constant no caller currently inspects.
- Mark the constant deprecated: rejected — there is no caller to migrate.

### 6. Mock verification idiom: inline `t.Cleanup(AssertExpectations)`

**Choice.** Each of the 14 test bodies that constructs a `MockCommandExecutor` adds `t.Cleanup(func() { mockExecutor.AssertExpectations(t) })` immediately after the `.On(...)` registrations.

**Rationale.** `golang-stretchr-testify` SKILL "Common Mistakes" lists "Forgetting `AssertExpectations(t)`" as the leading mock misuse. `golang-testing` SKILL rule 6 and `golang-cli/references/concurrency.md` both recommend "goleak in `TestMain`" as the canonical placement for cross-cutting test invariants — the same reasoning applies to mock verification: a single placement per mock that always runs at test teardown, regardless of which assertions fired. The cleanup-registered idiom is one line, lives next to the mock setup, and is the most local enforcement possible without changing the `NewMockCommandExecutor` signature.

**Alternatives considered.**

- Wrap `NewMockCommandExecutor` to auto-register `AssertExpectations`: rejected — 14 call sites change; the wrapper adds a `*testing.T` parameter that the current signature does not need.
- Lint-time enforcement via `testifylint expect-expect-assertions`: deferred to Change A — the rule requires a linter not currently enabled.

### 7. `Factory.Init` double-wrap removal: drop the `cmdutil: ...: %w` wraps

**Choice.** `Factory.Init` collects errors with `errors.Join` only. The current `fmt.Errorf("cmdutil: <role> init: %w", err)` wrappers are removed. Callers that want context see the underlying error's own `Op`, `Entity`, and `Message` fields.

**Rationale.** The wrapped error is already a `*core.OperationError` (or another typed error) carrying full context; the `cmdutil:` prefix adds noise without information. The `errors.Join` pattern gives every init failure independently, and the formatter already walks `Op` and `Message` correctly.

**Alternatives considered.**

- Keep wrappers but use `errors.Join(fmt.Errorf("cmdutil: ...: %w", err), ...)` so the wrap appears as its own join entry: rejected — same as dropping them; the wrapper is structural noise.

### 8. `goleak.VerifyTestMain(m)` placement and strictness

**Choice.** Add `func TestMain(m *testing.M) { goleak.VerifyTestMain(m); os.Exit(m.Run()) }` at the top of `test/concurrent/concurrent_test.go`. No `IgnoreTopFunction` filters.

**Rationale.** `golang-testing` SKILL rule 6: "Packages with goroutines SHOULD use `goleak.VerifyTestMain` in `TestMain` to detect goroutine leaks". `golang-cli/references/concurrency.md` line 97: "go.uber.org/goleak in `TestMain` — expect no goroutines left after a command returns". Strict mode catches library leaks at the cost of one verification per test; the alternative (ignore lists) masks real regressions.

**Alternatives considered.**

- `VerifyNone(t)` per-test: rejected — leaks from a non-test-main goroutine in a goroutine that returned successfully would not be caught.
- `goleak.IgnoreCurrent()` + `IgnoreTopFunction("...")`: rejected — pre-emptive ignores for an unknown library surface; strict first, ignore on demand.

### 9. `t.Setenv` migration and `must := require.New(t)` rename

**Choice.** Replace `os.Setenv` + `defer os.Setenv(origKey, origVal)` patterns in `test/integration/config_test.go` and `internal/config/manager_test.go` with `t.Setenv` (auto-restored on test cleanup). Rename `require := require.New(t)` to `must := require.New(t)` in the three `_test.go` files where the import is shadowed. Suite methods in `test/concurrent/concurrent_test.go` switch `require(s.T(), ...)` to `s.Require()` and `assert(s.T(), ...)` to `s.Assert()`.

**Rationale.** `golang-testing` SKILL: `t.Setenv` is the standard library replacement that "automatically restores" the original value, plays correctly with `t.Parallel()`, and removes the manual defer boilerplate. `golang-stretchr-testify` SKILL: "Name them `is` and `must`" is the skill convention; the shadowing of the `require` import is the antithesis of that convention.

**Alternatives considered.**

- Add `t.Parallel()` to the migrated tests: deferred — these tests share filesystem state via `t.TempDir()` and may not be safely parallelizable; out of scope for this change.

### 10. `go.uber.org/goleak` dependency

**Choice.** Add `go.uber.org/goleak` as a test-only dependency.

**Rationale.** The skill explicitly recommends it; the import path is standard and stable; the package is widely used (Kubernetes, Prometheus, gRPC all depend on it transitively). No production code changes its import graph.

**Alternatives considered.**

- Hand-rolled goroutine leak detection via `runtime.NumGoroutine()` snapshots: rejected — flaky, race-prone, no stack attribution.

### 11. Remove `ErrorKindNotFound` and `ExternalError.Is()`

**Choice.** Drop the `ErrorKindNotFound` constant from the `ErrorKind` iota in `internal/git/errors.go` and remove the `Is(target error) bool` method on `*ExternalError` in its entirety. `classifyKind` keeps the `context.DeadlineExceeded` → `ErrorKindTimeout` mapping and otherwise returns `ErrorKindOther`. The four per-resource NotFound sentinels continue to match via `*core.NotFoundError.Is()` walking the cause chain — exactly the path the baseline spec already mandates.

**Rationale.** `grep -nE 'Kind\s*[:=]\s*ErrorKindNotFound'` finds no setter anywhere in the module. The `Is()` method at `internal/git/errors.go:85-99` short-circuits with `if e.Kind != ErrorKindNotFound { return false }`, so it can never match `core.ErrGitRepoNotFound` or `core.ErrWorktreeNotFound` in practice. Removing the constant alone leaves `Is()` as dead code; removing `Is()` alone leaves an unreachable constant. Both go. The typed-error path the baseline spec mandates (`*core.NotFoundError.Is()` over the cause chain) already covers every concrete call site: `ExternalError.Unwrap()` returns `[OperationError, Cause]`; when `Cause` is a `*core.NotFoundError`, `errors.Is(err, core.ErrXNotFound)` walks to it directly. Same logic as Decision 5 for `ErrorKindPermission` — declared but never assigned, fragile across platforms, and the typed `OperationError` mapping already covers the only legitimate caller path.

**Alternatives considered.**

- Wire a read-side method (e.g. `OpenRepository`) to set `Kind = ErrorKindNotFound` when the go-git error chain indicates a missing repo — rejected; the typed-error path already covers callers, and adding a `Kind` bridge duplicates the `*core.NotFoundError.Is()` work.
- Mark `Is()` deprecated — rejected; no caller to migrate (no setter = no path that produces a working match).
- Move `Is()` to a method on `*core.OperationError` keyed on `Op` prefix — rejected; spec rule says NotFound dispatch is solely a `*core.NotFoundError` property, and the indirection table grows for no observable gain.

## Risks / Trade-offs

- **Logger writer change is observable in test stderr capture**: tests that previously relied on debug output going to the host's `os.Stderr` will now see it in the test buffer instead. → Mitigation: tests that assert on debug output must read `iostreams.Test()`'s stderr buffer, not `os.Stderr`.
- **Five new sentinels expand the package surface**: the prior archive deliberately removed sentinels to force all matching through `errors.Is` against a typed error. → Mitigation: sentinels ARE the `errors.Is` contract; restoring them enables it. The seven subtype structs stay removed.
- **Strict goleak will flag any test that leaks goroutines by accident**, including any third-party goroutine started indirectly. → Mitigation: when a test fails on goleak output, identify the stack and either fix the leak or add a targeted `IgnoreTopFunction` for that specific stack only. No blanket ignores.
- **Cause-chain test (`errors.As` on `*exec.ExitError`) is sensitive to `command_executor.go` preserving the error**: if a future change drops the cause again, the test fails loudly. → Mitigation: the test is part of the verifier gate (`go test -run TestNonZeroExit ./internal/git/...`).
- **`Factory.Init` no longer prefixes init failures with `cmdutil:`**: a script or test that greps stderr for `cmdutil:` will break. → Mitigation: no such consumer exists in the repo (`grep -r "cmdutil:" .` returns no callers); the formatter already shows the underlying error's `Op` which carries the same diagnostic value.
- **`ErrorKindPermission` removal is a no-op today** but could surprise a future caller expecting the enum to include a permission class. → Mitigation: this is a single-file deletion in `internal/git/errors.go`; the diff is reviewable and the package's CHANGELOG / archive history can document the removal.
- **`slog.Default()` test-buffer bypass in `internal/git/hook_runner.go`** — the hook production-write warning is logged via `slog.Default()` (production code over test-buffer observability, per the user's intent). Test-side assertions on this path read host stderr, not the test buffer. → Accepted gap; tests that need to assert on this path are authored as E2E tests with stderr captured at the process boundary.

## Migration Plan

This is a code-only change with no deployable artifacts, schema, or wire-format changes. The migration is the change itself:

1. Land Logger unification first; pointer-equality test in `internal/iostreams/iostreams_test.go` confirms single-source.
2. Land single-handling rule fixes (`detectOpError`, `discoverProjects`); cmd-side logging at the boundary replaces the dropped adapter logs.
3. Land cause-chain fix (`NewCommandError` + six writer.go sites); add `TestCLIClient_NonZeroExit_PreservesExecError`.
4. Land shell sentinels restoration; migrate `cmd/init.go:222`.
5. Land `ErrorKindPermission` and `ErrorKindNotFound` removal (drop the constant + the `Is()` method).
6. Land `Factory.Init` double-wrap removal.
7. Land `hook_runner.go:60` `slog.Warn` migration.
8. Land test discipline changes (14 AssertExpectations + t.Setenv + goleak + `must` rename).
9. Land docs (README GODEBUG section, `.vscode/launch.json`).

**Rollback:** revert the change. No data migration; no schema. The branch can be force-pushed-as-fast-forwarded only on a feature branch.

## Open Questions

None. All material decisions resolved during the planning conversation. The remaining unknowns (e.g. exact `goleak.IgnoreTopFunction` strings, exact text of the GODEBUG documentation paragraph) are non-blocking and can be filled in during implementation within the chosen approach.
