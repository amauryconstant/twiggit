# Tasks

## 1. Logger unification

- [x] 1.1 Change `iostreams.NewLogger` to accept an `io.Writer` parameter, keep the `TWIGGIT_DEBUG` env-var read inside, return a single `*slog.Logger` instance and verify by `go build ./...` and `go test ./internal/iostreams/...` passing.
- [x] 1.2 Wire `iostreams.System()` to call `NewLogger(os.Stderr)` and `iostreams.Test()` to call `NewLogger(testErrBuf)` so debug writes land in the test buffer; verify by `go test ./internal/iostreams/...` passing.
- [x] 1.3 Update `cmdutil.Factory.Logger` to return `iostreams.NewLogger(f.IOStreams.ErrOut)` and verify by `go build ./internal/cmdutil/...` and `go test ./internal/cmdutil/...` passing.
- [x] 1.4 Update `main.go` to call `slog.SetDefault(factory.Logger())` exactly once before `rootCmd.Execute()` and verify by `go build .` succeeding.
- [x] 1.5 Add `TestLoggerPointer` in `internal/iostreams/iostreams_test.go` asserting `System().Logger == NewLogger(os.Stderr)` pointer-equality and `Test().Logger == NewLogger(testErrBuf)` pointer-equality; verify by `go test -run TestLoggerPointer ./internal/iostreams/...` passing.

## 2. Single-handling rule restoration

- [x] 2.1 Remove the `slog.Error` call from `detectOpError` in `internal/git/context_detector.go` and verify by `go build ./internal/git/...` and `go test ./internal/git/...` passing.
- [x] 2.2 Remove the `slog.Error` call from `discoverProjects` in `internal/git/context_resolver.go` and verify by `go build ./internal/git/...` and `go test ./internal/git/...` passing.
- [x] 2.3 Trace the cmd-side callers of `detector.DetectContext` and apply boundary debug logging at each site. Confirmed call sites: `cmd/list.go:113`, `cmd/cd.go:100`, `cmd/create.go:126`, `cmd/delete.go:113`, `cmd/prune.go:127`, `cmd/suggestions.go:192`. Each site SHALL add `opts.IO.Logger.With("command", cmd.Name()).Debug("detect failed", "err", err)` immediately after the `DetectContext` error check (or be refactored through a small `cmd` helper that adds the boundary logger once); verify by `TWIGGIT_DEBUG=1 ./twiggit list`, `cd`, `create`, `delete`, `prune` each producing one debug line per failed detection (not two).

## 3. Cause-chain preservation

- [x] 3.1 Update `git.NewCommandError("non-zero-exit", ...)` in `internal/git/command_executor.go` to pass the underlying `err` as the cause instead of `nil` and verify by `go build ./internal/git/...` passing.
- [x] 3.2 Update the six `NewWorktreeError`/`NewBranchError` call sites at `internal/git/writer.go` lines 137, 183, 206, 230, 257, 283 to pass `result.Err` as the cause instead of `nil` and verify by `go build ./internal/git/...` passing.
- [x] 3.3 Add `TestCLIClient_NonZeroExit_PreservesExecError` in `internal/git/writer_test.go` asserting `errors.As(err, &*exec.ExitError)` walks the chain and verify by `go test -run TestCLIClient_NonZeroExit ./internal/git/...` passing.

## 4. Shell sentinels restoration

- [x] 4.1 Add five sentinels (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`, `ErrInvalidShellType`, `ErrInferenceFailed`, `ErrDetectionFailed`) as `var ErrXxx = errors.New("core: ...")` package variables in `internal/core/shell_errors.go` and verify by `go build ./internal/core/...` passing.
- [x] 4.2 Wire the existing shell-error constructors in `internal/core/shell_errors.go` to set `Cause` to the matching sentinel when their conditions trigger and verify by `go build ./internal/core/...` and `go test ./internal/core/...` passing.
- [x] 4.3 Replace the `errors.As+Op` match at `cmd/init.go:222` with `errors.Is(err, core.ErrShellAlreadyInstalled)` and verify by `go build ./cmd/...` and `go test ./cmd/...` passing.
- [x] 4.4 Add `TestShellAlreadyInstalled_Is` in `internal/core/shell_errors_test.go` asserting `errors.Is(err, core.ErrShellAlreadyInstalled)` returns true for a wrapped `OperationError` and false for an unrelated error; verify by `go test -run TestShellAlreadyInstalled ./internal/core/...` passing.

## 5. `ErrorKindPermission` and `ErrorKindNotFound` removal

- [x] 5.1 Drop the `ErrorKindPermission` constant from the `ErrorKind` iota in `internal/git/errors.go` and remove the constant from any references and verify by `go build ./...` succeeding with no unused-constant warnings.
- [x] 5.2 Drop the `ErrorKindNotFound` constant from the `ErrorKind` iota and remove the `Is(target error) bool` method on `*ExternalError` entirely. NotFound dispatch continues to walk through `*core.NotFoundError.Is()` over the cause chain (the baseline path); verify by `go build ./...` succeeding with no unused-constant or unused-method warnings and `go vet ./...` clean.
- [x] 5.3 Run `go test ./internal/git/...`, `go test ./cmd/...`, and `go test -tags=integration ./test/integration/...` and verify all tests still pass after the enum + method shrink.

## 6. `Factory.Init` double-wrap removal

- [x] 6.1 Drop the `fmt.Errorf("cmdutil: <role> init: %w", err)` wrappers in `internal/cmdutil/factory.go` `Factory.Init` and rely on `errors.Join` to surface each cached error and verify by `go build ./internal/cmdutil/...` and `go test ./internal/cmdutil/...` passing.
- [x] 6.2 Run `go test ./cmd/...` and verify the init-failure exit code path still maps to `cmdutil.ExitError` (1) via `cmdutil.ExitCodeFor`.

## 7. `hook_runner.go` `slog.Warn` migration

- [x] 7.1 Replace the `fmt.Fprintf(os.Stderr, ...)` warning at `internal/git/hook_runner.go:60` with `slog.Default().Warn("hook config parse warning", ...)` (the slog channel from main.go's `SetDefault`) and verify by `go build ./internal/git/...` and `go test ./internal/git/...` passing.

## 8. Mock verification idiom

- [x] 8.1 Add `t.Cleanup(func() { mockExecutor.AssertExpectations(t) })` to each of the twelve mock-based test bodies in `internal/git/writer_test.go` (tests: `TestCLIClient_CreateWorktree`, `TestCLIClient_CreateWorktree_WithExistingBranch`, `TestCLIClient_CreateWorktree_Failure`, `TestCLIClient_DeleteWorktree`, `TestCLIClient_DeleteWorktree_WithForce`, `TestCLIClient_ListWorktrees`, `TestCLIClient_PruneWorktrees`, `TestCLIClient_IsBranchMerged`, `TestCLIClient_DeleteBranch`, `TestCLIClient_Timeout`, `TestWriteSideFailure_OpIsGitWorktree`, `TestWriteSideFailure_OpIsGitBranch`) and verify by `go test ./internal/git/...` passing.
- [x] 8.2 Add `t.Cleanup(...)` to the two mock-based test bodies in `internal/git/hook_runner_test.go` (`TestHookRunner_Run_CommandFailure_ContinuesAndCollectsFailures`, `TestHookRunner_Run_EnvironmentVariablesSet`) and verify by `go test ./internal/git/...` passing.
- [x] 8.3 Verify the failure mode: deliberately delete one `mockExecutor.On(...)` registration and confirm the corresponding test fails with an `AssertExpectations` error, then restore the registration and confirm green.

## 9. `os.Setenv` → `t.Setenv` migration

- [x] 9.1 Replace `os.Setenv` + manual `defer os.Setenv(original, original)` patterns in `test/integration/config_test.go` with `t.Setenv(...)` and verify by `go test -tags=integration ./test/integration/...` passing.
- [x] 9.2 Replace the same pattern in `internal/config/manager_test.go` and verify by `go test ./internal/config/...` passing.

## 10. `goleak.VerifyTestMain` placement

- [x] 10.1 Add `func TestMain(m *testing.M) { goleak.VerifyTestMain(m); os.Exit(m.Run()) }` at the top of `test/concurrent/concurrent_test.go` and add the `go.uber.org/goleak` test import; verify by `go test -tags=concurrent ./test/concurrent/...` passing.
- [x] 10.2 Run `go mod tidy` and verify `go.mod` and `go.sum` are consistent with the new test dependency.
- [x] 10.3 Verify the failure mode: deliberately remove a `wg.Wait()` in `TestConcurrentListOperations` and confirm the test fails with a `goleak` report naming the leaked goroutine, then restore the wait and confirm green.

## 11. Suite helper and shadowed-import cleanup

- [x] 11.1 Replace `require(s.T(), ...)` with `s.Require()` and `assert(s.T(), ...)` with `s.Assert()` across all suite methods in `test/concurrent/concurrent_test.go` and verify by `go test -tags=concurrent ./test/concurrent/...` passing.
- [x] 11.2 Rename `require := require.New(t)` to `must := require.New(t)` in `internal/core/hook_types_test.go`, `internal/core/git_types_test.go`, and `internal/core/service_results_test.go`; update each call site to use `must.NoError(...)` etc. and verify by `go test ./internal/core/...` passing.

## 12. Documentation and IDE

- [x] 12.1 Add a one-paragraph `GODEBUG` section to `README.md` listing the common env vars (`gctrace=1`, `schedtrace=10000`, `asyncpreemptoff=1`) and the typical use cases for each and verify the file renders cleanly.
- [x] 12.2 Add `.vscode/launch.json` with a Delve launch configuration for the built `twiggit` binary, including pre-launch `go build` task and a `TWIGGIT_DEBUG=1` environment variable and verify by opening the project in VS Code with the Go extension that the launch entry appears.

## 13. Spec deltas and verification

- [x] 13.1 Sync the `cli-iostreams` spec delta under `specs/cli-iostreams/spec.md` to `openspec/specs/cli-iostreams/spec.md` and verify by `openspec validate cli-iostreams --type spec --strict` passing.
- [x] 13.2 Sync the `domain-typed-errors` spec delta under `specs/domain-typed-errors/spec.md` to `openspec/specs/domain-typed-errors/spec.md` and verify by `openspec validate domain-typed-errors --type spec --strict` passing.
- [x] 13.3 Run `mise run verify` and confirm format + lint + gopls + vuln + test + build all pass.

## 14. Force-fail and force-leak verification

- [x] 14.1 Temporarily delete one `mockExecutor.On(...)` registration, run the targeted test, confirm it fails with `AssertExpectations` error, then restore the registration and confirm green; verify the verification gate "tests fail when mocks don't fire" is live.
- [x] 14.2 Temporarily remove a `wg.Wait()` in `test/concurrent/concurrent_test.go`, run the concurrent suite, confirm `goleak` reports the leak with a stack, then restore the wait and confirm green; verify the verification gate "leaks are caught" is live.
- [x] 14.3 Temporarily revert `cmd/init.go` to the `errors.As+Op` string-match, run `go test ./cmd/...` and `TWIGGIT_DEBUG=1 ./twiggit init bash`, confirm the sentinel-match test fails, then restore the `errors.Is` and confirm green; verify the verification gate "string-matches die" is live.
