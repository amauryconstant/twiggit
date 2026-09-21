# Tasks

## 1. Setup

- [ ] 1.1 Capture pre-change baseline by running `mise run test` and confirming the suite is green before any code edits
- [ ] 1.2 Confirm `openspec list --json` reports `foundation-error-chain` as the only active change (so this change can land independently)
- [ ] 1.3 Verify `github.com/hashicorp/golang-lru/v2` is in `go.mod` (used by slice 4 for `context_detector.go` LRU swap)

## 2. Domain layer expansion (slice 1)

- [ ] 2.1 Create `internal/domain/pathutils.go` with `ExtractProjectFromWorktreePath`, `NormalizePath`, `IsPathUnder` — change signatures to return plain `error` (no `domain.NewContextDetectionError` calls); verify by `go build ./internal/domain/...` clean
- [ ] 2.2 Create `internal/domain/git_repo.go` with `IsMainRepo`, `FindMainRepoByTraversal`, `FindGitDirByTraversal` reshaped to `(string, bool)` instead of `*string`, plus `GitDir` struct moved from `internal/infrastructure/git_utils.go`; verify by `go build ./internal/domain/...` clean
- [ ] 2.3 Create `internal/domain/shell_wrapper.go` with `func ShellWrapper(shellType domain.ShellType) (string, error)` exposing the same bash/zsh/fish wrapper templates that `infrastructure.shell_infra.GenerateWrapper` returned; verify by `go build ./internal/domain/...` clean
- [ ] 2.4 Add `PathTypeUnknown = iota 0` to `internal/domain/context.go` and shift `PathTypeProject`/`PathTypeWorktree`/`PathTypeInvalid` by +1; verify by `git grep -n 'PathType.*=.*[0-9]'` returning only the iota block (no integer literals at call sites)
- [ ] 2.5 Rename `domain.WorktreeInfo.Modified` to `IsModified`; rename `domain.ShellResult.Installed` to `IsInstalled`, `Skipped` to `IsSkipped`; rename `domain.HookResult.Executed` to `HasExecuted`, `Success` to `IsSuccessful`; rename `domain.NewErrorResult[T]` to `NewErrResult[T]`; verify by `gopls rename` reporting zero unresolved references and `go build ./...` clean
- [ ] 2.6 Rename `domain.ValidationError.Context()` getter to `Detail()`; verify by `grep -rn 'ValidationError.*\.Context()' internal/ test/` returning no matches and `go build ./...` clean
- [ ] 2.7 Rewrite `internal/domain/pathutils_test.go` (NEW), `internal/domain/git_repo_test.go` (NEW), `internal/domain/shell_wrapper_test.go` (NEW) per project rule (tests AFTER implementation); add coverage for `FindGitDirByTraversal` `(string, bool)` shape; verify by `go test ./internal/domain/...` passing
- [ ] 2.8 Rewrite existing `internal/domain/git_types_test.go` and `internal/domain/hook_types_test.go` for the renamed fields; verify by `go test ./internal/domain/...` passing

## 3. Application interface split (slice 2)

- [ ] 3.1 Delete the `GitClient` interface (4 lines) at `internal/application/interfaces.go:101-105`; verify by `go build ./internal/application/...` clean and `git grep -n 'GitClient' internal/application/` returning only the role-interface references
- [ ] 3.2 Add `type RepoLocator interface { FindGitRepositories(dir string) ([]domain.GitDir, error) }` to `internal/application/interfaces.go`; verify by `go build ./internal/application/...` clean
- [ ] 3.3 Update `internal/application/AGENTS.md` interface table to remove `GitClient (Composite)` row and add `RepoLocator` row; verify by `grep -F 'GitClient (Composite)' internal/application/AGENTS.md` returning no matches

## 4. Service layer inversion (slice 3)

- [ ] 4.1 Rewrite `internal/service/worktree_service.go`: drop the `"twiggit/internal/infrastructure"` import line; replace `gitService application.GitClient` field with `goGit application.GoGitClient` and `cli application.CLIClient` fields; update `NewWorktreeService` signature to take both; route read calls (`BranchExists`, `GetRepositoryStatus`) through `s.goGit` and mutation calls (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees`, `IsBranchMerged`, `DeleteBranch`) through `s.cli`; verify by `go build ./internal/service/...` clean
- [ ] 4.2 Delete the `mu sync.Mutex` field from `worktreeService` struct in `internal/service/worktree_service.go:27`; verify by `grep -n 'sync.Mutex' internal/service/worktree_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.3 Replace `hookResult, _ = s.hookRunner.Run(ctx, hookReq)` at `worktree_service.go:95` with explicit error handling via `slog.Error` and continue (no error return); verify by `go build ./internal/service/...` clean
- [ ] 4.4 Replace `_ = s.gitService.PruneWorktrees(...)` at `worktree_service.go:611` with `slog.Error` and continue (the partial-failure field on the result already captures the user-visible state); verify by `go build ./internal/service/...` clean
- [ ] 4.5 Replace the linear `protectedBranches` scan at `worktree_service.go:622-630` with `slices.Contains(protectedBranches, branchName)`; verify by `go build ./internal/service/...` clean
- [ ] 4.6 Rewrite `internal/service/project_service.go`: drop the `"twiggit/internal/infrastructure"` import line; split the `gitService application.GitClient` field into `goGit application.GoGitClient` + `cli application.CLIClient`; add a `repoLocator application.RepoLocator` field; update `NewProjectService` signature to take both clients plus the locator; replace `infrastructure.FindGitRepositories(projectsDir, s.gitService)` at lines 68 and 90 with `s.repoLocator.FindGitRepositories(projectsDir)`; replace `infrastructure.FindMainRepoByTraversal`, `infrastructure.ExtractProjectFromWorktreePath`, `infrastructure.IsMainRepo` at lines 250, 262, 269 with `domain.FindMainRepoByTraversal`, `domain.ExtractProjectFromWorktreePath`, `domain.IsMainRepo`; verify by `go build ./internal/service/...` clean
- [ ] 4.7 Replace `os.IsNotExist(err)` at `project_service.go:221` with `errors.Is(err, os.ErrNotExist)`; verify by `grep -n 'os.IsNotExist' internal/service/project_service.go` returning no matches
- [ ] 4.8 Remove the unused `contextService application.ContextService` field from `projectService` struct (verify no readers in `project_service.go` before deleting); verify by `grep -n 's.contextService' internal/service/project_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.9 Remove the unused `config` field from `contextService` struct in `internal/service/context_service.go` (verify no readers); verify by `grep -n 's.config' internal/service/context_service.go` returning no matches and `go build ./internal/service/...` clean
- [ ] 4.10 Update `internal/service/shell_service.go` to use `domain.ShellWrapper(shellType)` instead of calling into `infrastructure.NewShellInfrastructure()`; verify by `go build ./internal/service/...` clean
- [ ] 4.11 Rewrite `internal/service/shell_service_test.go` to drop the `"twiggit/internal/infrastructure"` import and use `domain.ShellWrapper` for mock seed values; verify by `go test ./internal/service/...` passing
- [ ] 4.12 Rewrite `internal/service/worktree_service_test.go` and `internal/service/project_service_test.go` to construct two role mocks via `mocks.NewMockGitClientBundle()` and pass both to the new constructors; verify by `go test ./internal/service/...` passing
- [ ] 4.13 Rewrite `internal/service/doc.go` to remove the reference to `internal/infrastructure`; verify by `grep -F 'infrastructure' internal/service/doc.go` returning no matches

## 5. Infrastructure rewrite (slice 4)

- [ ] 5.1 Delete `internal/infrastructure/pathutils.go`; verify by `go build ./...` clean (the functions now live in `internal/domain/pathutils.go`)
- [ ] 5.2 Delete `internal/infrastructure/git_utils.go`; verify by `go build ./...` clean
- [ ] 5.3 Create `internal/infrastructure/repo_finder.go` with `FindGitRepositories(dir string, goGit application.GoGitClient) ([]domain.GitDir, error)` using the `domain.GitDir` type; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5.4 Delete `internal/infrastructure/git_client.go` (146-line `CompositeGitClient`); verify by `go build ./...` clean
- [ ] 5.5 Delete `internal/infrastructure/interfaces.go` (9-line placeholder); verify by `go build ./...` clean
- [ ] 5.6 Rename `CLIClientImpl` to `CLIClient` and `NewCLIClientImpl` to `NewCLIClient` in `internal/infrastructure/cli_client.go`; update the compile-time check `var _ application.CLIClient = (*CLIClientImpl)(nil)` accordingly; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5.7 Rename `GoGitClientImpl` to `GoGitClient` and `NewGoGitClientImpl` to `NewGoGitClient` / `NewGoGitClientWithSizeImpl` to `NewGoGitClientWithSize` in `internal/infrastructure/gogit_client.go`; update the compile-time check; verify by `go build ./internal/infrastructure/...` clean
- [ ] 5.8 Change `NewGoGitClient` and `NewGoGitClientWithSize` return types from `*GoGitClient` to `(*GoGitClient, error)` so the `lru.New` allocation error at lines 33 and 55 propagates instead of being discarded; verify by `go build ./...` clean (callers updated in slice 5)
- [ ] 5.9 Delete the `_ = remoteRef` workaround at `gogit_client.go:128` (drop the bound variable); verify by `grep -n '_ = remoteRef' internal/infrastructure/gogit_client.go` returning no matches
- [ ] 5.10 Bound the commit hash slice at `gogit_client.go:336` with `min(7, len(hashStr))` instead of `hashStr[:7]`; verify by `go test ./internal/infrastructure/...` passing for any short-hash fixture
- [ ] 5.11 Add `if result == nil { return <error> }` nil-guard before each `result.ExitCode` dereference in `cli_client.go` at lines 115, 151, 176, 198, 220, 247; verify by `go test ./internal/infrastructure/...` passing
- [ ] 5.12 Add `if cmdResult == nil { return <error> }` nil-guard before `cmdResult.ExitCode` at `hook_runner.go:137`; verify by `go test ./internal/infrastructure/...` passing
- [ ] 5.13 Rename `DefaultCommandExecutor` to `CommandExecutor` and `NewDefaultCommandExecutor` to `NewCommandExecutor` in `internal/infrastructure/command_executor.go`; verify by `gopls rename` zero unresolved references
- [ ] 5.14 Replace `os.IsNotExist(err)` with `errors.Is(err, os.ErrNotExist)` in every touched infrastructure file (`cli_client.go`, `gogit_client.go`, `config_manager.go`, `context_resolver.go`, `shell_infra.go`, `hook_runner.go`, `navigation_service.go`, `project_service.go` — but only the parts the layer inversion rewrites); verify by `grep -rn 'os.IsNotExist' internal/infrastructure/` returning no matches
- [ ] 5.15 Replace `strings.HasPrefix(s, prefix) + strings.TrimPrefix(s, prefix)` with `strings.CutPrefix(s, prefix)` at `cli_client.go:25-26` and `config_manager.go:29`; verify by `go build ./...` clean
- [ ] 5.16 Replace the unbounded `map[string]cached` field at `internal/infrastructure/context_detector.go:22` with `lru.Cache[string, cached]` (size 256, already in `go.mod`); verify by `go test ./internal/infrastructure/...` passing
- [ ] 5.17 Rename `ContextDetectorImpl` to `ContextDetector` and `NewContextDetectorImpl` to `NewContextDetector`; same for `ContextResolver`, `ConfigManager`, `HookRunnerImpl`, `ShellInfrastructureImpl`; update their compile-time checks; verify by `gopls rename` clean
- [ ] 5.18 Update `internal/infrastructure/context_resolver.go` to accept `goGit application.GoGitClient` + `cli application.CLIClient` (replacing the composite) and replace local `ProjectRef` struct with `domain.ProjectSummary`; verify by `go build ./...` clean
- [ ] 5.19 Update `internal/infrastructure/config_manager.go` `ProtectedBranches` deep-copy comment to use `slices.Clone`; verify by `go build ./...` clean
- [ ] 5.20 Collapse the five near-identical no-op result blocks in `hook_runner.go:37-95` into a single `noOpResult(req) *domain.HookResult` helper and call it from each branch; verify by `go build ./...` clean and `go test ./internal/infrastructure/...` passing
- [ ] 5.21 Rewrite `internal/infrastructure/cli_client_test.go` for the new nil-guard semantics and renamed constructor; verify by `go test ./internal/infrastructure/cli_client_test.go` passing
- [ ] 5.22 Rewrite `internal/infrastructure/gogit_client_test.go` for the new constructor signature (returning `(*GoGitClient, error)`); verify by `go test ./internal/infrastructure/gogit_client_test.go` passing

## 6. main.go rewiring (slice 5)

- [ ] 6.1 Update `main.go`: drop the `infrastructure.NewCompositeGitClient(...)` call (line 43); pass `goGitClient` and `cliClient` directly into `service.NewWorktreeService(...)` (line 53) and `service.NewProjectService(...)` (line 50); add `repoLocator := infrastructure.NewRepoFinder(goGitClient)` and pass it to `NewProjectService`; pass both clients into `infrastructure.NewContextResolver(...)` (line 46); verify by `go build ./...` clean
- [ ] 6.2 Update `cmd/root.go` interface references for any renamed infrastructure types (likely none — `cmd/` consumes service interfaces, not infrastructure types); verify by `go build ./...` clean

## 7. Test mocks + mechanical rename (slice 6)

- [ ] 7.1 Rewrite `test/mocks/git_service_mock.go` to expose `MockGitClientBundle` struct with `MockGoGitClient *MockGoGitClient` and `MockCLIClient *MockCLIClient` fields (the two inner mocks retain their existing `*_test.go`-adjacent mock methods); verify by `go build ./test/mocks/...` clean
- [ ] 7.2 Add `MockRepoLocator` to `test/mocks/` for `application.RepoLocator`; verify by `go build ./test/mocks/...` clean
- [ ] 7.3 Rename `MockShellInfrastructureImpl` to `MockShellInfrastructure` (drop `Impl`) and any other `*Impl`-suffixed mock types; verify by `grep -rn 'Impl' test/mocks/` returning no `Mock*Impl` matches
- [ ] 7.4 Mechanical rename across `test/integration/`: replace `infrastructure.NewCLIClientImpl` → `NewCLIClient`, `NewGoGitClientImpl` → `NewGoGitClient`, `NewDefaultCommandExecutor` → `NewCommandExecutor`, `NewContextDetectorImpl` → `NewContextDetector`, `NewContextResolverImpl` → `NewContextResolver`, `NewConfigManagerImpl` → `NewConfigManager`, `NewHookRunnerImpl` → `NewHookRunner`, `NewShellInfrastructureImpl` → `NewShellInfrastructure`; verify by `mise run test:integration` passing (or `go test -tags=integration ./test/integration/...`)
- [ ] 7.5 Same mechanical rename in `test/concurrent/` and `test/e2e/fixtures/`; verify by `mise run test:race` and `go test -tags=e2e ./test/e2e/...` passing

## 8. AGENTS.md sync + .golangci.yml (slice 7)

- [ ] 8.1 Rewrite `internal/application/AGENTS.md` interface table to remove `GitClient (Composite)` row and add `RepoLocator` row; verify by `grep -F 'Composite' internal/application/AGENTS.md` returning no matches
- [ ] 8.2 Rewrite `internal/service/AGENTS.md` examples to use `application.GoGitClient` and `application.CLIClient` as separate constructor params (no composite, no `infrastructure.*` imports); verify by `grep -F 'infrastructure.' internal/service/AGENTS.md` returning no matches
- [ ] 8.3 Rewrite `internal/infrastructure/AGENTS.md` GitClient routing section to drop `*Impl` from implementation-type names; verify by `grep -F 'Impl' internal/infrastructure/AGENTS.md` returning no matches
- [ ] 8.4 Update `internal/domain/AGENTS.md` to document the new helpers in `pathutils.go`, `git_repo.go`, `shell_wrapper.go` and the `IsModified`, `IsInstalled`, `IsSkipped`, `HasExecuted`, `IsSuccessful` field renames; verify by `grep -F 'PathTypeUnknown' internal/domain/AGENTS.md` returning a match
- [ ] 8.5 Update `test/mocks/AGENTS.md` to document `MockGitClientBundle` and `MockRepoLocator`; verify by `grep -F 'MockGitClientBundle' test/mocks/AGENTS.md` returning a match
- [ ] 8.6 Extend `depguard` rules in `.golangci.yml`: add allowlists for `internal/service/**`, `internal/application/**`, `internal/infrastructure/**`, `cmd/**`, `test/mocks/**` per design decision 13; verify by `mise run lint` failing for any test file that imports a forbidden package (smoke test by adding `import "twiggit/internal/infrastructure"` to a `service/` test, confirming it fails, then removing)
- [ ] 8.7 Drop `gocognit` from `.golangci.yml` `linters.enable` list; verify by `mise run lint` clean
- [ ] 8.8 Add `nolintlint` block under `.golangci.yml` `linters.settings` with `require-explanation: true, require-specific: true`; verify by adding a bare `//nolint` to any file and confirming `mise run lint` flags it (smoke test)
- [ ] 8.9 Add `errcheck.check-type-assertions: true` under `linters.settings.errcheck`; verify by `mise run lint` clean
- [ ] 8.10 Remove the blanket `text: "Close.*is not checked"` exclusion at `.golangci.yml:129-130`; for each `Close()` call site that intentionally discards the error, add `//nolint:errcheck // <reason>` on the line above (estimate 5-10 sites); verify by `mise run lint` clean
- [ ] 8.11 Remove the four redundant `//nolint:wrapcheck` directives at `internal/infrastructure/command_executor_mock_test.go:24,26,33,35` (wrapcheck is excluded for `_test.go`); verify by `mise run lint` clean

## 9. Verification (slice 9)

- [ ] 9.1 Run `mise run verify` and confirm lint + format + race + golden all pass; report any failures
- [ ] 9.2 Run `mise run test` and confirm the full suite is green; report any failing tests
- [ ] 9.3 Run `openspec validate architecture-layer-inversion --json` and confirm `valid: true, issues: []`
- [ ] 9.4 Run `git grep -n 'os.IsNotExist'` returning no matches in `internal/` (sanity check for modernization completeness)
- [ ] 9.5 Run `git grep -n 'Impl.*= nil'` in `internal/infrastructure/` returning no matches (sanity check for naming sweep completeness)
- [ ] 9.6 Run `git grep -n 'infrastructure\.' internal/service/` returning no matches (sanity check that the layer inversion is complete)
