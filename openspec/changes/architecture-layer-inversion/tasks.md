# Tasks

## 1. Setup

- [ ] 1.1 Capture pre-change baseline by running `mise run test` and confirming the suite is green before any code edits; also confirm `slog.SetDefault(...)` runs once in `main.go` so service-layer `slog.Error` calls land in the configured handler
- [ ] 1.2 Confirm this change can land independently of other active changes by inspecting `openspec list --json`; SPECIFICALLY verify `quality-modernization` does not schedule overlapping edits to `.golangci.yml` or `internal/infrastructure/{cli_client,gogit_client,context_detector,context_resolver,config_manager,hook_runner,shell_infra,command_executor}.go`; reject if either change owns a slice before the other lands
- [ ] 1.3 Verify `github.com/hashicorp/golang-lru/v2` is in `go.mod` (used by slice 4c for `context_detector.go` LRU swap)
- [ ] 1.4 Verify `golangci-lint v2` is installed and accepts `linters.settings.nolintlint` (slice 7)

## 2. Domain layer expansion (slice 1)

- [ ] 2.1 Create `internal/domain/pathutils.go` with `ExtractProjectFromWorktreePath`, `NormalizePath`, `IsPathUnder` — change signatures to return plain `error` (no `domain.NewContextDetectionError` calls); verify by `go build ./internal/domain/...` clean
- [ ] 2.2 Create `internal/domain/git_repo.go` with `IsMainRepo`, `FindMainRepoByTraversal`, `FindGitDirByTraversal` reshaped to `(string, bool)` instead of `*string`, plus `GitDir` struct moved from `internal/infrastructure/git_utils.go`; verify by `go build ./internal/domain/...` clean
- [ ] 2.3 Create `internal/domain/shell_wrapper.go` with `func ShellWrapper(shellType domain.ShellType) (string, error)` exposing the same bash/zsh/fish wrapper templates that `infrastructure.shell_infra.GenerateWrapper` returned; verify by `go build ./internal/domain/...` clean
- [ ] 2.4 Add `PathTypeUnknown = iota 0` to `internal/domain/context.go` and shift `PathTypeProject`/`PathTypeWorktree`/`PathTypeInvalid` by +1; verify by `git grep -n 'PathType.*=.*[0-9]'` returning only the iota block (no integer literals at call sites)
- [ ] 2.5 Rename `domain.WorktreeInfo.Modified` to `IsModified`; rename `domain.ShellResult.Installed` to `IsInstalled`, `Skipped` to `IsSkipped`; rename `domain.HookResult.Executed` to `HasExecuted`, `Success` to `IsSuccessful`; rename `domain.NewErrorResult[T]` to `NewErrResult[T]`; verify by ALL of: (a) `gopls rename` reports zero unresolved references, (b) `go build ./...` clean, (c) `git grep -n 'Executed bool\|Success bool\|Modified bool\|Installed bool\|Skipped bool' internal/domain/` returns no matches, (d) `git grep -rn 'HookResult\.Executed\|HookResult\.Success\|WorktreeInfo\.Modified\|ShellResult\.Installed\|ShellResult\.Skipped' internal/ cmd/ test/` returns no matches (catches string-literal/template references gopls missed), (e) `gofmt -r 'HookResult.Executed -> HookResult.HasExecuted' -l ./...` returns no files
- [ ] 2.5a Verify all 20 domain error constructors retain their existing `NewXxxError` style (`NewValidationError`, `NewGitRepositoryError`, `NewGitWorktreeError`, `NewGitCommandError`, `NewConfigError`, `NewContextDetectionError`, `NewServiceError`, `NewWorktreeServiceError`, `NewProjectServiceError`, `NewNavigationServiceError`, `NewResolutionError`, `NewConflictError`, `NewShellAlreadyInstalledError`, `NewShellNotInstalledError`, `NewShellInvalidTypeError`, `NewShellInferenceError`, `NewShellDetectionError`, `NewShellWrapperError`, `NewShellConfigError`, `NewUsageError`); verify the `Error type taxonomy` table in `openspec/specs/domain-typed-errors/spec.md` keeps the `NewXxxError` column entries (per `golang-naming` skill `Error`-suffix rule for error types); verify by `go build ./...` clean
- [ ] 2.5b Verify the `domain-typed-errors/spec.md` table in slice 8 (spec deltas) matches the constructor list in task 2.5a (20 `NewXxxError` constructors, retained existing names); if a constructor was added/removed beyond the 20 listed, update both the table and the spec delta before either lands
- [ ] 2.6 Rename `domain.ValidationError.Context()` getter to `Detail()`; verify by `grep -rn 'ValidationError.*\.Context()' internal/ test/` returning no matches and `go build ./...` clean
- [ ] 2.7 Rewrite `internal/domain/pathutils_test.go` (NEW), `internal/domain/git_repo_test.go` (NEW), `internal/domain/shell_wrapper_test.go` (NEW) per project rule (tests AFTER implementation); add coverage for `FindGitDirByTraversal` `(string, bool)` shape; verify by `go test ./internal/domain/...` passing
- [ ] 2.8 Rewrite existing `internal/domain/git_types_test.go` and `internal/domain/hook_types_test.go` for the renamed fields and constructors; verify by `go test ./internal/domain/...` passing
- [ ] 2.9 Audit `domain.Result[T]` and the success/error constructors for slice aliasing: any field of type `[]T` carried inside `Result[T]` SHALL be cloned via `slices.Clone` at construction time to prevent the calling code from mutating internal state through the returned struct; verify by `go test ./internal/domain/...` passing for the new aliasing tests
- [ ] 2.10 Verify every `domain.New*ServiceError` and `domain.New*Git*Error` constructor in slice 1 wraps its cause with `%w` (not `%v`) and that `errors.Is(err, Err*NotFound)` reaches the per-resource sentinel per `domain-typed-errors` spec table; verify by adding `TestErrorsIsChain` table-driven test under `internal/domain/errors_test.go` running clean

## 3. Application interface split (slice 2)

- [ ] 3.1 Delete the `GitClient` interface (4 lines) at `internal/application/interfaces.go`; verify by `go build ./internal/application/...` clean and `git grep -n 'GitClient' internal/application/` returning only the role-interface references (umbrella gone)
- [ ] 3.2 Add `type RepoLocator interface { FindGitRepositories(dir string) ([]domain.GitDir, error) }` to `internal/application/interfaces.go`; verify by `go build ./internal/application/...` clean
- [ ] 3.3 Add compile-time interface check `var _ application.RepoLocator = (*infrastructure.RepoFinder)(nil)` at the bottom of `internal/infrastructure/repo_finder.go` (created in slice 5d); verify by `go build ./...` clean and the check survives `gopls rename` of the implementation type

## 4. Service layer inversion (slice 3)

- [ ] 4.1 Rewrite `internal/service/worktree_service.go`: drop the `"twiggit/internal/infrastructure"` import line; replace `gitService application.GitClient` field with `goGit application.GoGitClient` and `cli application.CLIClient` fields; update `NewWorktreeService` signature to take both; route read calls (`BranchExists`, `GetRepositoryStatus`) through `s.goGit` and mutation calls (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees`, `IsBranchMerged`, `DeleteBranch`) through `s.cli`; verify by `go build ./internal/service/...` clean
- [ ] 4.2 Delete the `mu sync.Mutex` field from `worktreeService` struct in `internal/service/worktree_service.go`; verify by `grep -n 'sync.Mutex' internal/service/worktree_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.3 Replace `hookResult, _ = s.hookRunner.Run(ctx, hookReq)` at `worktree_service.go` (around the post-create hook call site) with explicit error logging via `slog.Error` and continue; verify the surrounding branch does NOT also `return err` (single-handling rule from `golang-error-handling`: log OR return, never both); verify by `go build ./internal/service/...` clean and `go test ./internal/service/...` passing
- [ ] 4.4 Replace `_ = s.gitService.PruneWorktrees(...)` at the prune-failure logging site with `slog.Error` and continue (the partial-failure field on the result already captures the user-visible state); verify the surrounding branch does NOT also `return err` (single-handling rule); verify by `go build ./internal/service/...` clean and `go test ./internal/service/...` passing
- [ ] 4.5 Replace the linear `protectedBranches` scan at `worktree_service.go` with `slices.Contains(protectedBranches, branchName)`; verify by `go build ./internal/service/...` clean
- [ ] 4.6 Rewrite `internal/service/project_service.go`: drop the `"twiggit/internal/infrastructure"` import line; split the `gitService application.GitClient` field into `goGit application.GoGitClient` + `cli application.CLIClient`; add a `repoLocator application.RepoLocator` field; update `NewProjectService` signature to take both clients plus the locator; replace `infrastructure.FindGitRepositories(projectsDir, s.gitService)` calls with `s.repoLocator.FindGitRepositories(projectsDir)`; replace `infrastructure.FindMainRepoByTraversal`, `infrastructure.ExtractProjectFromWorktreePath`, `infrastructure.IsMainRepo` calls with the corresponding `domain.*` helpers; verify by `go build ./internal/service/...` clean
- [ ] 4.7 Replace every `os.IsNotExist(err)` in `internal/service/project_service.go` with `errors.Is(err, os.ErrNotExist)`; verify by `grep -rn 'os.IsNotExist' internal/service/` returning no matches
- [ ] 4.8 Remove the unused `contextService application.ContextService` field from `projectService` struct (verify no readers in `project_service.go` before deleting); verify by `grep -n 's.contextService' internal/service/project_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.9 Remove the unused `config` field from `contextService` struct in `internal/service/context_service.go` (verify no readers); verify by `grep -n 's.config' internal/service/context_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.10 Update `internal/service/shell_service.go` to use `domain.ShellWrapper(shellType)` instead of calling into `infrastructure.NewShellInfrastructure()`; verify by `go build ./internal/service/...` clean
- [ ] 4.11 Rewrite `internal/service/shell_service_test.go` to drop the `"twiggit/internal/infrastructure"` import and use `domain.ShellWrapper` for mock seed values; verify by `go test ./internal/service/...` passing
- [ ] 4.12 Rewrite `internal/service/worktree_service_test.go` and `internal/service/project_service_test.go` to construct two role mocks via `mocks.NewMockGitClientBundle()` and pass both to the new constructors; verify by `go test ./internal/service/...` passing
- [ ] 4.13 Rewrite `internal/service/doc.go` to remove the reference to `internal/infrastructure`; verify by `grep -F 'infrastructure' internal/service/doc.go` returning no matches

## 5. Infrastructure rewrite (slice 4 — split into 6 atomic sub-slices)

The slice-4 tasks are renumbered into 5a/5b-i/5b-ii/5c-i/5c-ii/5c-iii/5c-iv/5d/5e so
each sub-slice stays atomic AND fits within the `golang-refactoring` skill
100-500 lines per PR rule. Each sub-slice is purely structural or purely
behavioral, never both. Sub-slice ordering is fixed; within a sub-slice the
order of tasks is implementation order.

### 5a. Pure renames (structural)

- [ ] 5a.1 Rename `CLIClientImpl` → `CLIClient` and `NewCLIClientImpl` → `NewCLIClient` in `internal/infrastructure/cli_client.go`; update the compile-time check `var _ application.CLIClient = (*CLIClient)(nil)`; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5a.2 Rename `GoGitClientImpl` → `GoGitClient` and `NewGoGitClientImpl` → `NewGoGitClient` / `NewGoGitClientWithSizeImpl` → `NewGoGitClientWithSize` in `internal/infrastructure/gogit_client.go`; update the compile-time check; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5a.3 Rename `DefaultCommandExecutor` → `CommandExecutor` and `NewDefaultCommandExecutor` → `NewCommandExecutor` in `internal/infrastructure/command_executor.go`; verify by `gopls rename` zero unresolved references
- [ ] 5a.4 Rename `ContextDetectorImpl` → `ContextDetector` / `NewContextDetectorImpl` → `NewContextDetector`; rename `ContextResolverImpl` → `ContextResolver` / `NewContextResolverImpl` → `NewContextResolver`; rename `ConfigManagerImpl` → `ConfigManager` / `NewConfigManagerImpl` → `NewConfigManager`; rename `HookRunnerImpl` → `HookRunner` / `NewHookRunnerImpl` → `NewHookRunner`; rename `ShellInfrastructureImpl` → `ShellInfrastructure` / `NewShellInfrastructureImpl` → `NewShellInfrastructure`; update each compile-time check; verify by `grep -rn 'Impl.*= nil' internal/infrastructure/` returning no matches and `go build ./internal/infrastructure/...` clean
- [ ] 5a.5 Verify slice 5a: `git grep -n 'Impl.*= nil' internal/infrastructure/` returns no matches; `mise run verify` clean

### 5b-i. NewGoGitClient signature change (behavioral)

- [ ] 5b-i.1 Change `NewGoGitClient` and `NewGoGitClientWithSize` return types from `*GoGitClient` to `(*GoGitClient, error)` so the `lru.New` allocation error at lines around the constructor propagates instead of being discarded; verify by `go build ./internal/infrastructure/...` clean (callers updated in slice 6)
- [ ] 5b-i.2 Update `main.go` and the integration-test wiring to handle the new `error` return from both `NewGoGitClient*` constructors; verify by `go build ./...` clean and `mise run test:integration` passing

### 5b-ii. NewContextDetector signature change (behavioral)

- [ ] 5b-ii.1 Change `NewContextDetector` return type from `*ContextDetector` to `(*ContextDetector, error)` so the `lru.New` allocation error in `context_detector.go` propagates; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5b-ii.2 Update `main.go` and the integration-test wiring to handle the new `error` return from `NewContextDetector`; verify by `go build ./...` clean and `mise run test:integration` passing

### 5c-0. Type nil verification (precondition)

- [ ] 5c-0.1 Verify `command_executor.ExecuteWithTimeout` returns a concrete `*CommandResult` (not interface); refactor to concrete pointer first if interface today; verify by `grep -n 'ExecuteWithTimeout' internal/infrastructure/command_executor.go` showing concrete `*CommandResult` return type and `go build ./internal/infrastructure/...` clean

### 5c-i. Nil-guards (behavioral)

- [ ] 5c-i.1 Add `if result == nil { return <error> }` nil-guard before each `result.ExitCode` dereference in `cli_client.go` (6 sites); verify by `go test ./internal/infrastructure/...` passing
- [ ] 5c-i.2 Add `if cmdResult == nil { return <error> }` nil-guard before `cmdResult.ExitCode` at `hook_runner.go:137`; verify by `go test ./internal/infrastructure/...` passing

### 5c-ii. `errors.Is` migration + bound slice + prefix cut (behavioral)

- [ ] 5c-ii.1 Replace `os.IsNotExist(err)` with `errors.Is(err, os.ErrNotExist)` across ALL files in `internal/` (expanded scope — not just touched files); verify by `grep -rn 'os.IsNotExist' internal/` returning no matches
- [ ] 5c-ii.2 Replace `strings.HasPrefix(s, prefix) + strings.TrimPrefix(s, prefix)` with `strings.CutPrefix(s, prefix)` at the two flagged sites; verify by `go build ./...` clean
- [ ] 5c-ii.3 Bound the commit hash slice at `gogit_client.go` with `min(7, len(hashStr))` instead of `hashStr[:7]`; verify by `go test ./internal/infrastructure/...` passing for any short-hash fixture
- [ ] 5c-ii.4 Delete the `_ = remoteRef` workaround at `gogit_client.go` (drop the bound variable); verify by `grep -n '_ = remoteRef' internal/infrastructure/gogit_client.go` returning no matches

### 5c-iii. LRU cache swap (behavioral)

- [ ] 5c-iii.1 Replace the unbounded `map[string]cached` field at `internal/infrastructure/context_detector.go` with `lru.Cache[string, cached]` (size 256); verify by `go test ./internal/infrastructure/...` passing

### 5c-iv. Resolver + config refactor (behavioral)

- [ ] 5c-iv.1 Update `internal/infrastructure/context_resolver.go` to accept `goGit application.GoGitClient` + `cli application.CLIClient` (replacing the composite) and replace local `ProjectRef` struct with `domain.ProjectSummary`; verify by `go build ./...` clean
- [ ] 5c-iv.2 Update `internal/infrastructure/config_manager.go` `ProtectedBranches` deep-copy to use `slices.Clone`; verify by `go build ./...` clean

### 5d. Refactors (structural)

- [ ] 5d.1 Delete `internal/infrastructure/pathutils.go` (functions now live in `internal/domain/pathutils.go`); verify by `go build ./...` clean
- [ ] 5d.2 Delete `internal/infrastructure/git_utils.go`; verify by `go build ./...` clean
- [ ] 5d.3 Create `internal/infrastructure/repo_finder.go` with `FindGitRepositories(dir string, goGit application.GoGitClient) ([]domain.GitDir, error)` using the `domain.GitDir` type; add compile-time interface check `var _ application.RepoLocator = (*RepoFinder)(nil)` at the bottom of the file; POLICY: return a freshly allocated `[]domain.GitDir` via `slices.Clone` so callers cannot mutate the implementation's internal slice header; add unit test asserting that mutating the returned slice does not affect a second call; verify by `go build ./internal/infrastructure/...` clean and the new defensive-copy test passing
- [ ] 5d.4 Delete `internal/infrastructure/git_client.go` (146-line `CompositeGitClient`); verify by `go build ./...` clean
- [ ] 5d.5 Delete `internal/infrastructure/interfaces.go` (9-line placeholder); verify by `go build ./...` clean
- [ ] 5d.6 Collapse the five near-identical no-op result blocks in `hook_runner.go` into a single `noOpResult(req) *domain.HookResult` helper and call it from each branch; verify by `go build ./...` clean and `go test ./internal/infrastructure/...` passing
- [ ] 5d.7 Verify slice 5d: `git grep -n 'infrastructure\.' internal/service/` returns no matches (sanity check that the layer inversion is complete); `go build ./...` clean

### 5e. Test rewrites (anchor)

- [ ] 5e.1 Rewrite `internal/infrastructure/cli_client_test.go` for the new nil-guard semantics and renamed constructor; REQUIRED fixtures: (a) `ExecuteWithTimeout` returns `(*CommandResult, error)` with non-nil error and nil `*CommandResult` — exercises the nil-guard; (b) successful execution returns populated result; verify by `go test ./internal/infrastructure/cli_client_test.go` passing both cases
- [ ] 5e.2 Rewrite `internal/infrastructure/gogit_client_test.go` for the new constructor signature (returning `(*GoGitClient, error)`); verify by `go test ./internal/infrastructure/gogit_client_test.go` passing
- [ ] 5e.3 Verify slice 5e: `go test ./internal/infrastructure/...` passing

## 6. main.go rewiring (slice 5)

- [ ] 6.1 Update `main.go`: drop the `infrastructure.NewCompositeGitClient(...)` call; pass `goGitClient` and `cliClient` directly into `service.NewWorktreeService(...)` and `service.NewProjectService(...)`; construct the repo finder as `repoFinder := infrastructure.NewRepoFinder(goGitClient)` (concrete `*RepoFinder` returned per `golang-naming` "return structs" rule; assign to `application.RepoLocator` at the consumer site); pass both clients into `infrastructure.NewContextResolver(...)`; verify by `go build ./...` clean
- [ ] 6.2 Update `cmd/root.go` interface references for any renamed infrastructure types (likely none — `cmd/` consumes service interfaces, not infrastructure types); verify by `go build ./...` clean

## 7. Test mocks + mechanical rename (slice 6)

- [ ] 7.1 Rewrite `test/mocks/git_service_mock.go` to expose `MockGitClientBundle` struct with `MockGoGitClient *MockGoGitClient` and `MockCLIClient *MockCLIClient` fields (the two inner mocks retain their existing `*_test.go`-adjacent mock methods); verify by `go build ./test/mocks/...` clean
- [ ] 7.2 Add `MockRepoLocator` to `test/mocks/` for `application.RepoLocator`; verify by `go build ./test/mocks/...` clean
- [ ] 7.3 Rename `MockShellInfrastructureImpl` to `MockShellInfrastructure` (drop `Impl`) and any other `*Impl`-suffixed mock types; verify by `grep -rn 'Impl' test/mocks/` returning no `Mock*Impl` matches
- [ ] 7.4 Mechanical rename across `test/integration/`: replace `infrastructure.NewCLIClientImpl` → `NewCLIClient`, `NewGoGitClientImpl` → `NewGoGitClient`, `NewDefaultCommandExecutor` → `NewCommandExecutor`, `NewContextDetectorImpl` → `NewContextDetector`, `NewContextResolverImpl` → `NewContextResolver`, `NewConfigManagerImpl` → `NewConfigManager`, `NewHookRunnerImpl` → `NewHookRunner`, `NewShellInfrastructureImpl` → `NewShellInfrastructure`; verify by `mise run test:integration` passing (or `go test -tags=integration ./test/integration/...`)
- [ ] 7.5 Same mechanical rename in `test/concurrent/` and `test/e2e/fixtures/`; verify by `mise run test:race` and `go test -tags=e2e ./test/e2e/...` passing
- [ ] 7.6 Verify NO mechanical rename is required: confirm `test/integration/`, `test/concurrent/`, `test/e2e/fixtures/` keep calling `domain.New<Xxx>Error(` style constructors unchanged (Decision 19 settles constructor naming: `NewXxxError` is canonical, `NewXxxErr` rejected); verify by `git grep -rn 'domain\.New[A-Z][A-Za-z]*Err(' test/` returning no matches; the test suites above still pass

## 8. .golangci.yml (slice 7)

- [ ] 8.1 Extend `depguard` rules in `.golangci.yml`: add per-layer allowlists AND explicit `deny:` blocks (per `golang-cli-architecture` recommendation) for `internal/service/**`, `internal/application/**`, `internal/infrastructure/**`, `cmd/**`, `test/mocks/**`. Allowlist:
  - `domain` (existing): allow `$gostd`, `internal/domain`
  - `service` (new): allow `$gostd`, `internal/domain`, `internal/application`, `internal/service`; deny `twiggit/internal/infrastructure`, `twiggit/cmd`
  - `application` (new): allow `$gostd`, `internal/domain`, `internal/application`; deny `twiggit/internal/infrastructure`, `twiggit/internal/service`, `twiggit/cmd`
  - `infrastructure` (new): allow `$gostd`, `internal/domain`, `internal/application`, `internal/infrastructure`; deny `twiggit/internal/service`, `twiggit/cmd`
  - `cmd` (new): allow `$gostd`, `internal/domain`, `internal/application`, `internal/service`, `internal/infrastructure`, `internal/version`, `twiggit/cmd`; deny the above-listed reverse-direction imports for clarity in CI failure messages
  - Verify by `mise run lint` failing for any test file that imports a forbidden package (smoke test by adding `import "twiggit/internal/infrastructure"` to a `service/` test, confirming it fails, then removing)
- [ ] 8.2 Drop `gocognit` from `.golangci.yml` `linters.enable` list; verify by `mise run lint` clean
- [ ] 8.3 Add `nolintlint` block under `.golangci.yml` `linters.settings` with `require-explanation: true, require-specific: true`; verify by adding a bare `//nolint` to any file and confirming `mise run lint` flags it (smoke test)
- [ ] 8.4 Add `errcheck.check-type-assertions: true` under `linters.settings.errcheck`; verify by `mise run lint` clean
- [ ] 8.5 Remove the blanket `text: "Close.*is not checked"` exclusion; for each `Close()` call site that intentionally discards the error, add `//nolint:errcheck // <reason>` on the line above; verify by `mise run lint` clean
- [ ] 8.6 Remove the four redundant `//nolint:wrapcheck` directives in `command_executor_mock_test.go` (wrapcheck is excluded for `_test.go`); verify by `mise run lint` clean
- [ ] 8.7 Add `modernize: enable` to `.golangci.yml` `linters.enable` list (requires golangci-lint v2.6.0+, gated by Task 1.4); verify by `mise run lint` clean and `golangci-lint linters` listing `modernize` enabled

## 9. Verification (slice 9)

- [ ] 9.1 Run `mise run verify` and confirm lint + format + race + golden all pass; report any failures
- [ ] 9.2 Run `mise run test` and confirm the full suite is green; report any failing tests
- [ ] 9.3 Run `openspec validate architecture-layer-inversion --json` and confirm `valid: true, issues: []`
- [ ] 9.4 Run `git grep -n 'os.IsNotExist' internal/` returning no matches (sanity check for modernization completeness)
- [ ] 9.5 Run `git grep -n 'Impl.*= nil' internal/infrastructure/` returning no matches (sanity check for naming sweep completeness)
- [ ] 9.6 Run `git grep -n 'infrastructure\.' internal/service/` returning no matches (sanity check that the layer inversion is complete)
- [ ] 9.7 Run `git grep -n 'New[A-Z][A-Za-z]*Error('` returning no matches across `internal/`, `cmd/`, `test/` (sanity check for slice 1 Err-rename completeness)
