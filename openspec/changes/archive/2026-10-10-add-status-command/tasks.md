# Tasks

## 1. Core domain: extend `core.WorktreeStatus` and add `Config.Status`

- [x] 1.1 Extend `core.WorktreeStatus` in `internal/core/service_results.go` with the six new fields (`Base string`, `IsMerged bool`, `IsStale bool`, `LastCommitDate time.Time`, `IsSkipped bool`, `SkipReason string`) and JSON struct tags on every exported field; verify by reading the file back and confirming the new fields appear with the expected tags.
- [x] 1.2 Add a `StatusConfig` struct (`StaleBehind int`, `StaleDays int`) in `internal/core/config.go` and a `Status StatusConfig` field on `core.Config`; verify by `go build ./internal/core/` succeeding.
- [x] 1.3 Set the default values (`StaleBehind=20`, `StaleDays=30`) in `core.DefaultConfig()`; verify by `go test ./internal/core/ -run TestDefaultConfig` passing once that test is added or the existing default-coverage test still passes.
- [x] 1.4 Add `StatusConfig` to `core.Config.Validate()` (no validation rules beyond non-negative; document in the comment block); verify by `go test ./internal/core/ -run TestConfig_Validate` passing.
- [x] 1.5 Add unit tests in `internal/core/service_results_test.go` for the new fields: zero-value safety, JSON round-trip, and a table-driven `IsStale` derivation covering fresh/behind/aged/both/neither/disabled; verify by `go test ./internal/core/ -run TestWorktreeStatus` passing.
- [x] 1.5a Add `func (s *WorktreeStatus) ComputeIsStale(cfg *core.Config) bool` and `func (s *WorktreeStatus) Dirty() bool` in `internal/core/service_results.go`; verify by `go build ./internal/core/` succeeding and the existing `data_type_fixture_test.go` still passing (no type removed).
- [x] 1.6 Add a koanf tag round-trip test confirming `[status] stale_behind` and `[status] stale_days` flow into `Config.Status`; verify by `go test ./internal/core/ -run TestStatusConfig` passing.

## 2. Role interface: declare `WorktreeStatusReader` and add the drift sentinel

- [x] 2.1 Add the `WorktreeStatusReader` interface (one method: `ReadWorktreeStatus(ctx context.Context, cfg *core.Config, repoPath, wtPath string) (core.WorktreeStatus, error)`) to `internal/core/git.go` next to the existing role interfaces; verify by `go build ./internal/core/` succeeding and a grep for the new method on the role block.
- [x] 2.2 Add `var _ core.WorktreeStatusReader = (*Client)(nil)` in `internal/git/client.go` next to the existing drift sentinels; verify by `go build ./internal/git/` failing until the implementation lands (negative check) and then succeeding once 3.1 lands.
- [x] 2.3 Amend the drift-sentinel comment block in `internal/git/client.go` to acknowledge that the composite now carries direct methods in addition to promoted ones; verify by reading the comment back and confirming the new sentence appears.

## 3. Adapter: implement `Client.ReadWorktreeStatus`

- [x] 3.1 Add `internal/git/status.go` with `func (c *Client) ReadWorktreeStatus(ctx, cfg, repoPath, wtPath) (core.WorktreeStatus, error)`; verify by `go build ./internal/git/` succeeding with the new file present and the drift sentinel from 2.2 satisfied.
- [x] 3.2 Inside `ReadWorktreeStatus`: read dirty state via `c.RepositoryStatus` (go-git), populate the existing `core.WorktreeStatus.IsClean` / `HasUncommittedChanges` fields, and copy the populated `RepositoryStatus` (including `Ahead` / `Behind` set to zero initially) into the projection; verify by writing a table-driven unit test in `internal/git/status_test.go` that drives a mocked `CommandExecutor` and asserts the field set.
- [x] 3.3 Inside `ReadWorktreeStatus`: compute ahead/behind via `git rev-list --count <base>..<wt-branch>` and `<wt-branch>..<base>`, parse the integer stdout, and write the result into `RepositoryStatus.Ahead` and `Behind`; verify by a unit test that asserts the parse against three fixture stdout values (`"5\n"`, `"0\n"`, `"12\n"`).
- [x] 3.4 Inside `ReadWorktreeStatus`: call `c.IsBranchMerged(ctx, wtPath, wt.Branch)` (first arg is the worktree path per `cmd/delete.go:209`; the parameter name `repoPath` in the role signature is misleading) and set `IsMerged` accordingly; on failure set `IsSkipped=true`, `SkipReason=<lowercase reason>`, and return the row without error; verify by a unit test where the mock executor returns non-zero for the merge check and the resulting row carries `IsSkipped=true`.
- [x] 3.5 Inside `ReadWorktreeStatus`: resolve the `Base` via the rebase tracked-base chain (per-worktree `git config --worktree --get twiggit.tracked-base` first, then `Config.Validation.ProtectedBranches[0]`); verify by three unit tests covering (a) tracked-base present, (b) no tracked-base + fallback present, (c) no tracked-base + empty protected list, asserting `Base` and `IsSkipped` per the spec.
- [x] 3.6 Inside `ReadWorktreeStatus`: look up the worktree's branch via `c.ListBranches` and populate `LastCommitDate` from the matching `core.Branch.Date`; on no match leave the zero time and do not mark `IsSkipped`; verify by a unit test where the mock branch list contains (and does not contain) the target branch.
- [x] 3.7 The adapter SHALL NOT derive `IsStale`; the value is left at the zero (`false`) and the cmd layer fills it per row via `row.ComputeIsStale(cfg)`; verify by a unit test that confirms the adapter-populated `core.WorktreeStatus` carries `IsStale = false`.
- [x] 3.8 Inside `ReadWorktreeStatus`: set `LastChecked = time.Now()`; verify by a unit test that asserts the field is non-zero and within the last second of the test's wall clock.
- [x] 3.9 Add a sentinel-passthrough test in `internal/git/client_test.go` that confirms `*Client` satisfies `core.WorktreeStatusReader` at compile time and at runtime (calling the method returns a populated value); verify by `go test ./internal/git/ -run TestClient_WorktreeStatusReader` passing.

## 4. Refactor: move `getWorktreeStatus` out of `cmd/delete.go` and onto the new role

- [x] 4.1 Delete the local `getWorktreeStatus` function and the `now` helper from `cmd/delete.go`; verify by `go build ./cmd/` failing with "undefined: getWorktreeStatus" (negative check) until 4.2 lands.
- [x] 4.2 Replace the call site in `cmd/delete.go` (around the `getWorktreeStatus` invocation in the dirty-check) with a call to `client.ReadWorktreeStatus(ctx, cfg, target.ProjectPath, target.WorktreePath)` and read the `IsClean` / `HasUncommittedChanges` fields from the returned `core.WorktreeStatus`; verify by `go test ./cmd/ -run TestDelete` passing.
- [x] 4.3 Verify the dirty-check behaviour of `cmd/delete` is unchanged across the three states (clean, modified, detached); verify by `go test ./cmd/ -run TestDelete_Dirty` and the existing `cmd/delete_test.go` dirty-path tests passing.

## 5. New command: `cmd/status.go` and its projections

- [x] 5.1 Add `cmd/status.go` with `StatusOptions` (Factory-injected IOStreams / Config / GitClient / Ctx / GlobalOptions / Logger, flags `All bool`, `StaleBehind int`, `StaleDays int`); verify by `go build ./cmd/` succeeding.
- [x] 5.1a In `runStatus`, fill `IsStale` per row via `row.ComputeIsStale(cfg)` after the adapter returns; verify by a unit test where one worktree has `Behind >= StaleBehind` and the corresponding row carries `IsStale=true`.
- [x] 5.2 Add `NewCmdStatus(f, runF)` with `cobra.MaximumNArgs(1)`, the `--all` / `--output` / `--stale-behind` / `--stale-days` flags, the `runF` test seam, and registration in the `core` group in `cmd/root.go`; verify by `go build ./cmd/` succeeding and `twiggit --help` listing `status` in the `Core:` section when the binary runs.
- [x] 5.3 Implement `runStatus(opts)` to (a) resolve the formatter before the walk and return `*core.UsageError` on bad values, (b) detect context (current project / `--all` / positional), (c) discover the target worktrees via the same `RepoFinder` path `list` uses; verify by `go test ./cmd/ -run TestStatus_OutsideGitUsageError` and `TestStatus_UnknownProject` passing.
- [x] 5.4 Implement the per-project walk: iterate the project's worktrees, call `client.ReadWorktreeStatus` for each, collect rows, and emit per-row warnings on stderr (suppressed under `--quiet`); verify by a unit test using `runF` injection with a fake `core.WorktreeStatusReader` that returns one skipped row among three; the captured stdout contains three rows, the captured stderr contains one warning, and the test asserts exit 0.
- [x] 5.5 Add the `statusRows` Tabular projection in `cmd/status.go` (Header `[]string{"BRANCH", "PATH", "AHEAD", "BEHIND", "BASE", "MERGED", "DIRTY", "STALE"}`, Rows with the matching string values per row); verify by a unit test that asserts the header order and a sample row's column count.
- [x] 5.6 Add a `MarshalJSON` on `statusRows` that emits the bare array shape per `cli-output`; verify by a unit test that decodes the JSON and asserts the array length, the per-object fields, and the `omitempty` on `SkipReason`.
- [x] 5.7 Implement the default-output path (no `--output` flag) using `output.RenderTable(opts.IO, statusRows.Header(), statusRows.Rows(), opts.IO)` (the existing formatter helper); verify by a unit test that compares the rendered string to a golden fixture in `test/golden/status/default.golden`.
- [x] 5.8 Wire the formatter-driven path (`json|table|plain`) through `output.NewFormatter` and the `statusRows` Tabular / JSON implementations; verify by three unit tests, one per format, each comparing to a captured golden fixture (`test/golden/status/{json,table,plain}.golden`).
- [x] 5.9 Add `cmd/status_test.go` parse-level tests covering every flag, `--all` + positional mutual exclusion (returning `*core.UsageError`), and the outside-git → `UsageError` path; verify by `go test ./cmd/ -run TestStatus_Parse` passing.
- [x] 5.10 Add the help-text assertion test that confirms `twiggit status --help` lists the four flags (`--all`, `--output`, `--stale-behind`, `--stale-days`); verify by `go test ./cmd/ -run TestStatus_Help` passing.

## 6. Integration: divergent / merged / clean / dirty / stale-by-age fixtures

- [x] 6.1 Add `test/integration/status_test.go` (`//go:build integration`) with a helper that builds a real git repo containing one main, one feature-ahead, one feature-merged, one feature-clean, one feature-dirty, and one feature-with-tracked-base; verify by `go test -tags=integration -run TestStatus_Integration_AllShapes` passing.
- [x] 6.2 Cover the five `core.WorktreeStatus` derivations on real git repos: ahead/behind counts, merged detection, dirty detection, tracked-base fallback, and stale-by-age with `Config.Status.StaleDays=1`; verify by `go test -tags=integration ./test/integration/ -run TestStatus_Integration` passing.

## 7. E2E: smoke tests and goldens for the four output shapes

- [x] 7.1 Add `test/e2e/status_test.go` (`//go:build e2e`) with smoke tests for `twiggit status`, `twiggit status --all`, `twiggit status --output json`, `twiggit status --output table`, and `twiggit status --output plain` against a built binary on a real fixture repo; verify by `go test -tags=e2e ./test/e2e/ -run TestStatus_E2E` passing.
- [x] 7.2 Add golden fixtures under `test/golden/status/` for each of the four output shapes and the per-skip-warning case; verify by `UPDATE_GOLDEN=true mise run test:e2e` regenerating the goldens and `go test -tags=e2e -run TestStatus_E2E` passing on the re-run without `UPDATE_GOLDEN`.
- [x] 7.3 Add a `test/e2e/status_test.go` case for the `wtg status | jq` pipeline (mirrors the `cli-output-formats` stream-separation scenario); verify by the captured JSON document having no interleaved stderr bytes (assert on the captured stdout of the e2e helper).
- [x] 7.4 Add a `test/e2e/status_test.go` case for the outside-git → `UsageError` exit-2 path; verify by the e2e helper asserting exit code 2 and a stderr substring naming both options (`--all` and positional).

## 8. Docs: AGENTS.md, README, and openspec AGENTS.md

- [x] 8.1 Update the top-level `AGENTS.md` architecture table to list `internal/git/status.go` under the git I/O adapter row and a `Status` row under the config knobs; verify by reading the table back and confirming both rows appear.
- [x] 8.2 Update `openspec/AGENTS.md` to mention the new `core-worktree-status` and `cli-status` specs in the spec-organization cross-reference list and to note the move of `getWorktreeStatus` out of `cmd/delete.go`; verify by grep on `core-worktree-status` and `cli-status` finding one match each in the file.
- [x] 8.3 Add a `twiggit status` section to `README.md` under the existing command list, with one short example and a one-line description of the columns; verify by reading the section back and confirming the columns named in the section match the spec's JSON field set.
- [x] 8.4 Run `mise run verify` end-to-end and confirm format / lint:fix / lint:gated / vuln:check / test / build all pass; verify by `mise run verify` exiting 0.
