# Tasks

## 1. Domain: rebase value objects, sentinels, roles

- [ ] 1.1 Add `RebaseRequest`, `RebaseResult`, `RebasedWorktree`, `RebaseOutcome` enum, `SyncRequest`, `SyncResult`, `SyncedBranch` to `internal/core/rebase.go`; verify by `go build ./internal/core/...`
- [ ] 1.2 Add `core.RebaseOutcome.String()` returning lowercase identifiers per variant; verify by `go test ./internal/core/ -run TestRebaseOutcome_String`
- [ ] 1.3 Add `ErrRebaseConflict`, `ErrRebaseInProgress`, `ErrBaseNotSet` to `internal/core/sentinels.go`; verify by `go build ./internal/core/...` and `go test ./internal/core/ -run TestSentinels_Rebase` asserting each `Error()` matches `^[a-z]` and ends with `[a-z0-9]` (no trailing punctuation)
- [ ] 1.4 Declare `Rebaser` (4 methods: `Rebase`, `Abort`, `Continue`, `Fetch`) and `BaseTracker` (2 methods: `SetTrackedBase`, `GetTrackedBase`) in `internal/core/git.go`; verify by `go build ./internal/core/...`
- [ ] 1.5 Write `internal/core/rebase_test.go` covering value-object zero values, outcome enum variants, and sentinel equality; verify by `go test ./internal/core/ -run TestRebase`
- [ ] 1.6 Extend `core.OperationError.Is` and `core.NotFoundError.Is` membership lists to walk `ErrRebaseConflict` and `ErrBaseNotSet` when `Op` carries the prefix `rebase.*` or `base.*` (per spec `core-errors/spec.md:33-44`); verify by `go test ./internal/core/ -run TestOperationError_Is_Rebase`

## 2. Domain: HookType extension and HookRunRequest widening

- [ ] 2.1 Add `HookTypePreRebase`, `HookTypePostRebase`, `HookTypePostSync` to the `core.HookType` enum in `internal/core/hook_types.go`; verify by `go build ./internal/core/...`
- [ ] 2.2 Add `RebaseBase`, `RebaseOldTip`, `RebaseNewTip`, `RebaseResult`, `SyncRemote`, `SyncBranch` optional string fields to `core.HookRunRequest`; verify by `go build ./...` (all existing callers compile unchanged)
- [ ] 2.3 Extend `internal/git/hook_runner.go` to set `TWIGGIT_REBASE_*` and `TWIGGIT_SYNC_*` env vars only when the hook type matches and the field is non-empty; verify by `go test ./internal/git/ -run TestHookRunner_EnvVars`
- [ ] 2.4 Extend `internal/core/hook_types_test.go` with assertions for every new `HookType` variant and for `HookRunRequest` round-trip with the new fields; verify by `go test ./internal/core/ -run TestHookType`
- [ ] 2.5 Extend `internal/git/hook_runner_test.go` to assert PreRebase hook failure short-circuits the run, and PostRebase/PostSync failures are recorded but non-fatal; verify by `go test ./internal/git/ -run TestHookRunner_FailureSemantics`
- [ ] 2.6 Extend `core.HookConfig` with `PreRebase []HookDefinition`, `PostRebase []HookDefinition`, `PostSync []HookDefinition` (koanf tags `pre-rebase` / `post-rebase` / `post-sync`); extend `internal/git/hook_runner.go` `Run` switch to dispatch each new variant from its slice; verify by `go test ./internal/core/ -run TestHookConfig_NewVariants` and `go test ./internal/git/ -run TestHookRunner_Dispatch`

## 3. Config: RebaseConfig and SyncConfig fields

- [ ] 3.1 Add `RebaseConfig{FetchOnAll bool, ConflictPolicy string}` and `SyncConfig{DefaultRemote string, PruneRemoteRefs bool, RebaseAfterSync bool}` to `core.Config` in `internal/core/config.go`; verify by `go build ./internal/core/...`
- [ ] 3.2 Add defaults in `core.DefaultConfig()` (FetchOnAll=false, ConflictPolicy="stop", SyncDefaultRemote="origin", SyncPruneRemoteRefs=true, SyncRebaseAfterSync=false); verify by `go test ./internal/core/ -run TestDefaultConfig`
- [ ] 3.3 Verify koanf loader reads `[rebase]` and `[sync]` blocks from the TOML config file and that `TWIGGIT_REBASE__FETCH_ON_ALL` / `TWIGGIT_SYNC__REBASE_AFTER_SYNC` env vars override file values; verify by `go test ./internal/config/ -run TestConfig_RebaseSync`

## 4. Infrastructure: Rebaser + BaseTracker adapter and error constructors

- [ ] 4.1 Add `Rebaser` methods (`Rebase`, `Abort`, `Continue`, `Fetch`) to `*cliClient` in a new `internal/git/rebaser.go`; route every call through the injected `CommandExecutor`; verify by `go build ./internal/git/...`
- [ ] 4.2 Add `BaseTracker` methods (`SetTrackedBase`, `GetTrackedBase`) to `*cliClient` in `internal/git/rebaser.go` using `git -C <wt> config --worktree`; verify by `go build ./internal/git/...`
- [ ] 4.3 Add a stderr-phrase detector (`isRebaseConflictStderr`, `isNothingToDoStderr`) in `internal/git/rebaser.go` mirroring `isNotFoundStderr`; verify by `go test ./internal/git/ -run TestRebaser_StderrPhrases`
- [ ] 4.4 Add `NewRebaseError` and `NewBaseTrackerError` constructors in `internal/git/errors.go` matching the `NewWorktreeError` / `NewBranchError` shape; verify by `go build ./internal/git/...`
- [ ] 4.5 Write `internal/git/rebaser_test.go` covering every `Rebaser` and `BaseTracker` method with `MockCommandExecutor` (`assert` for verifications, `require` for preconditions, `AssertExpectations(t)` in `t.Cleanup`); verify by `go test ./internal/git/ -run TestCliClient_Rebaser`

## 5. Infrastructure: composite satisfaction and drift sentinel

- [ ] 5.1 Add `var _ core.Rebaser = (*git.Client)(nil)` and `var _ core.BaseTracker = (*git.Client)(nil)` to `internal/git/client.go`; verify by `go build ./internal/git/...`
- [ ] 5.2 Extend the `_DriftCheck` sentinel in `internal/git/client_test.go` with `core.Rebaser` and `core.BaseTracker` fields; verify by `go build ./internal/git/...`
- [ ] 5.3 Add a table-driven test asserting all eight `var _` declarations are present in `client.go` and that any rename of a role method on `*cliClient` causes a compile error; verify by `go test ./internal/git/ -run TestDriftSentinel`

## 6. Presentation: tracked base persistence on create

- [ ] 6.1 In `cmd/create.go` `materialiseWorktree`, call `gitClient.SetTrackedBase(ctx, worktreePath, sourceBranch)` after `CreateWorktree` succeeds; log a warning via `opts.IO.Logger` when the call fails (do not fail the create); verify by `go build ./cmd/...`
- [ ] 6.2 Add a unit test in `cmd/create_test.go` asserting the warning is logged when `SetTrackedBase` returns an error and the worktree still exists; verify by `go test ./cmd/ -run TestCreate_SetTrackedBaseFailure`

## 7. Presentation: rebase cmd

- [ ] 7.1 Add `RebaseOptions`, `NewCmdRebase`, `runRebase` to `cmd/rebase.go` mirroring the prune pattern (Options struct, `runF` test seam, `RunE`, `SilenceUsage`/`SilenceErrors`); register `carapace.Gen(cmd).PositionalCompletion(actionWorktreeTarget(f, git.WithExistingOnly()))` for the `[project/branch]` positional; verify by `go build ./cmd/...`
- [ ] 7.2 Add `resolveRebaseTargets`, `resolveTrackedBase`, `dispatchRebaseHooks`, `emitRebaseOutput` to `cmd/rebase_helpers.go`; `dispatchRebaseHooks` returns the `*core.HookResult` and a bool indicating whether a PreRebase hook failure must abort the rebase; verify by `go build ./cmd/...`
- [ ] 7.3 Wire flags: positional `[project/branch]`, `--all`/`-a`, `--fetch`/`-f`, `--continue`/`-c`, `--abort`/`-A`, `--set-base <branch>`, `--force`/`-F`; the args validator (via `wrapArgsValidator`) MUST return `core.UsageError` when `--all` and a positional argument are both supplied; verify by `twiggit rebase --help` listing the flag set and `go test ./cmd/ -run TestRebase_AllAndPositional`
- [ ] 7.4 Register `rebase` in the `core` cobra group via `cmd/root.go`; verify by `twiggit --help` listing `rebase` under `Core:`
- [ ] 7.5 Write `cmd/rebase_helpers_test.go` (table-driven) covering `resolveTrackedBase` fallback to `Config.Validation.ProtectedBranches[0]`, `resolveRebaseTargets` honouring `[project/branch]` vs `--all` (and the UsageError when both are set), and `dispatchRebaseHooks` setting the right `HookRunRequest` fields AND returning `abort=true` when a `HookTypePreRebase` hook reports `IsSuccessful == false`; verify by `go test ./cmd/ -run TestRebaseHelpers`
- [ ] 7.6 Write `cmd/rebase_test.go` covering cwd rebase, named-wt rebase, `--all` fan-out with stop-on-first, `--fetch`, `--continue`, `--abort`, `--set-base`, dirty-wt refusal, fallback to protected branches, usage error outside git, single-target rebase prints the worktree path for the shell wrapper, a nothing-to-do rebase exits 0, and the `--all` + positional UsageError; verify by `go test ./cmd/ -run TestRebase`

## 8. Presentation: sync cmd

- [ ] 8.1 Add `SyncOptions`, `NewCmdSync`, `runSync` to `cmd/sync.go` mirroring the prune pattern; register `carapace.Gen(cmd).PositionalCompletion(actionWorktreeTarget(f, git.WithExistingOnly()))` for the `[project]` positional; verify by `go build ./cmd/...`
- [ ] 8.2 Add `resolveSyncTargets`, the shared `runRebaseWalk` invocation when `--rebase` is set, and the per-project fetch + `SyncedBranch` recording in `cmd/sync_helpers.go`; verify by `go build ./cmd/...`
- [ ] 8.3 Wire flags: positional `[project]`, `--all`/`-a`, `--remote <name>`, `--branch <name>`, `--rebase`/`-r`, `--fetch-only`/`-F`; verify by `twiggit sync --help` listing the flag set
- [ ] 8.4 Register `sync` in the `core` cobra group via `cmd/root.go`; verify by `twiggit --help` listing `sync` under `Core:`
- [ ] 8.5 Write `cmd/sync_helpers_test.go` (table-driven) covering `resolveSyncTargets` honouring `[project]` vs `--all` and the `--rebase` walk sharing `runRebaseWalk`; verify by `go test ./cmd/ -run TestSyncHelpers`
- [ ] 8.6 Write `cmd/sync_test.go` covering single-project sync, `--all`, `--remote`, `--branch`, `--rebase`, `--fetch-only`, and conflict stop-on-first when `--rebase` is set; verify by `go test ./cmd/ -run TestSync`

## 9. E2E and integration tests

- [ ] 9.1 Add `test/integration/rebase_test.go` covering real git: divergent branch rebase clean, conflict leaves mid-rebase, `--continue` after fix, `--abort` clears state, `--set-base` persists, PreRebase hook aborts rebase, PostRebase hook runs on clean only, fallback when tracked base missing; verify by `mise run [integration test for rebase]` (build tag `integration`)
- [ ] 9.2 Extend `test/repo/repo.go` with `CreateDivergentBranch`, `CreateDirtyWorktree`, `CreateMidRebase`, `AdvanceBase` helpers; verify by `go test ./test/repo/...`
- [ ] 9.3 Add `test/e2e/rebase_test.go` Ginkgo specs: cwd rebase, named-wt rebase, `--all` fan-out with stop-on-first, `--fetch`, `--continue`, `--abort`, `--set-base` (asserting `git -C <wt> config --worktree --get twiggit.tracked-base` returns the new base), dirty refuse, fallback, usage error outside git; verify by `mise run test:e2e -- --focus="rebase command"`
- [ ] 9.4 Add `test/e2e/sync_test.go` Ginkgo specs: single-project fetch, `--all`, `--remote`, `--branch`, `--rebase`, `--fetch-only`; verify by `mise run test:e2e -- --focus="sync command"`
- [ ] 9.5 Expand `test/e2e/workflows/rebase_workflow_test.go` with the create → work → rebase → push end-to-end flow and the shell-wrapper navigation assertion; verify by `mise run test:e2e -- --focus="rebase workflow"`

## 10. Docs

- [ ] 10.1 Add `twiggit rebase` and `twiggit sync` sections to `README.md` with examples for cwd, named wt, `--all`, `--fetch`, `--continue`, `--abort`, `--set-base`, dirty wash, fallback, hooks; verify by `gofmt -l README.md` and a grep for the new command names
- [ ] 10.2 Update `cmd/AGENTS.md` with the rebase and sync command shape (Options struct, NewCmd runF seam, helper-file list); verify by `gofmt -l cmd/AGENTS.md`
- [ ] 10.3 Update `openspec/AGENTS.md` only if a NEW outcome (not a rule change) emerged during implementation; otherwise skip; verify by `diff` showing no rule additions

## 11. Final verification

- [ ] 11.1 Run `mise run verify` and ensure format, lint:fix, lint:gated, vuln:check, tests, and build all pass; verify by exit 0 from `mise run verify`
- [ ] 11.2 Run `mise run test:race` and confirm no race conditions across the new tests; verify by exit 0 of `mise run test:race`
- [ ] 11.3 Run `go mod tidy && git diff --exit-code go.mod go.sum` to confirm no module drift; verify by empty diff