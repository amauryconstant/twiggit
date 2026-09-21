# Tasks

## 1. Setup (absorb Change 1 orphans)

- [ ] 1.1 Create `test/mocks/helpers.go` with `func variadicArgs(first interface{}, opts ...interface{}) []interface{}` and verify by `go build ./test/mocks/...` clean and `go vet ./test/mocks/...` clean
- [ ] 1.2 Fix `test/mocks/cmd_mocks.go:223` to call `m.Called(variadicArgs(partial, opts...)...)` instead of `m.Called(partial, opts)`; verify by `go build ./test/mocks/...` clean
- [ ] 1.3 Fix `test/mocks/cmd_mocks.go:233` to call `m.Called(variadicArgs(ctx, partial, opts...)...)` instead of `m.Called(ctx, partial, opts)`; verify by `go test ./test/mocks/...` passing (no test should regress)
- [ ] 1.4 Fix `openspec/config.yaml:134-140` parse error by moving the embedded `# Purity principles...` and `# Each rule is identified by...` comment lines out of the `specs:` list item to immediately above it (only if Change 1 task 1.2 has not landed); verify by `openspec list --json` printing no `Warning: could not parse` line
- [ ] 1.5 Capture pre-change baseline by running `mise run test` and confirming the suite is green before any code edits; record the green timestamp in the slice commit message

## 2. Lint config tightening

- [ ] 2.1 Add `formatters.enable: [gofumpt, goimports]` with `gofumpt.extra-rules: true` and add `issues.max-same-issues: 0` plus `max-issues-per-linter: 0` to `.golangci.yml`; verify by `mise run lint` clean and `gofumpt -d ./...` clean
- [ ] 2.2 Tighten `funlen.lines` 150→120 with `statements` 100→75 in `.golangci.yml` only after the first touched function in slice S11 is shortened to ≤120 lines; verify by `mise run lint` clean on the touched package
- [ ] 2.3 Tighten `gocyclo.min-complexity` 25→13 in `.golangci.yml` only after the first touched function in slice S11 is below complexity 13; verify by `mise run lint` clean on the touched package
- [ ] 2.4 Create `.github/dependabot.yml` with weekly Monday schedule monitoring `gomod`, `github-actions`, and `docker` ecosystems, with grouped patch+minor updates for dev-dependencies; verify by `dependabot.yml` validating via GitHub's schema (commit + push to a branch that has the file)
- [ ] 2.5 Create `.github/CODEOWNERS` with the maintainer team `@amoconst/core` owning `/cmd/`, `/internal/`, `/test/`, `/openspec/`, and root config files; verify by the file parsing under GitHub's CODEOWNERS rules

## 3. Slog migration extension

- [ ] 3.1 Replace the silent `continue` on `ListProjects` errors at `internal/service/project_service.go:75` with `slog.Error("list project", "path", path, "err", err)` followed by `continue`; verify by `go vet ./internal/service/...` clean and existing `project_service_test.go` still passes
- [ ] 3.2 Replace the silent `continue` on `ListProjectSummaries` errors at `internal/service/project_service.go:97` with `slog.Error("list project summary", "path", path, "err", err)` followed by `continue`; verify by `go test ./internal/service/...` passing
- [ ] 3.3 Change `cmd/util.go` `logv` (or `PrintV`) helper signature to accept an `io.Writer` parameter so the cobra command passes `c.ErrOrStderr()`; verify by `go build ./cmd/...` clean and existing `cmd/util_test.go` still passes after updating the test to pass `&bytes.Buffer{}`

## 4. t.Context in-fn literals (140 sites)

- [ ] 4.1 In `test/integration/prune_integration_test.go`, replace every `ctx := context.Background()` at function top with `ctx := t.Context()` (83 sites); verify by `go test -tags=integration ./test/integration/...` passing and `rg -c 'context\.Background' test/integration/prune_integration_test.go` returning 0
- [ ] 4.2 In `test/integration/git_clients_test.go`, replace 18 `context.Background()` literals with `t.Context()`; verify by `go test ./test/integration/git_clients_test.go` passing
- [ ] 4.3 In `test/integration/git_routing_test.go` (13 sites), `test/integration/project_discovery_test.go` (14 sites), `test/integration/hook_runner_test.go` (7 sites), and `test/integration/cmd/...` (the rest), replace literals; verify by `go test -tags=integration ./test/integration/...` passing and `rg -c 'context\.Background' test/integration/` returning 0
- [ ] 4.4 In `cmd/util_test.go`, replace the 6 `ctx := context.Background()` sites (lines 204, 242, 280, 310, 338, 369) with `ctx := t.Context()`; verify by `go test ./cmd/...` passing
- [ ] 4.5 In `cmd/init_test.go`, the 5 `shellService.On("...", context.Background(), ...)` matcher literals (lines 112, 146, 175, 208, 241) stay as-is for now (deferred to S6 where mock matchers are swapped together); no change in this slice

## 5. t.Context helper signatures (5 sites)

- [ ] 5.1 Update `test/helpers/worktree.go` helpers at lines 34, 50, 66, 82, 109 to accept `t *testing.T` as the first parameter; replace `context.WithTimeout(context.Background(), h.timeout)` with `context.WithTimeout(t.Context(), h.timeout)`; verify by `go build ./test/helpers/...` clean
- [ ] 5.2 Update every caller of the `test/helpers/worktree.go` functions across `test/integration/`, `test/concurrent/`, and `test/e2e/fixtures/` to pass `t` as the first argument; verify by `go test ./test/...` passing for the integration and concurrent suites
- [ ] 5.3 Verify the helpers' timeout propagation is unchanged by writing a quick test that captures the helper's ctx and asserts its `Done()` channel closes when `t` cancels; verify by the test passing

## 6. t.Context mock matcher swap (~30 sites)

- [ ] 6.1 Audit every `mockX.On("Method", context.Background(), ...)` site across `test/integration/` and the `cmd/` test files; produce a list of file:line for each site; verify the list is comprehensive (no `.On(` followed by `context.Background(` left unrecorded)
- [ ] 6.2 Replace each `context.Background()` matcher in mock expectations with `mock.Anything` for the context argument only (other arguments stay typed); verify by `go test ./test/integration/...` and `go test ./cmd/...` passing
- [ ] 6.3 Run the same test files WITHOUT `mock.Anything` first (i.e., before the swap) and capture the failing-test count; verify the swap eliminates all silent-pass failures (no test now passes because the matcher was wrong)
- [ ] 6.4 For mocks that genuinely verify a context with a deadline (search for `mock.MatchedBy(func(ctx context.Context) bool {` patterns), preserve the matcher; only swap sites that don't care about the specific context value; verify by per-site code review

## 7. Mock hygiene

- [ ] 7.1 In `test/mocks/helpers.go`, define `func newMockWithCleanup[T any](t *testing.T, ctor func() T) T` that constructs the mock, registers `t.Cleanup(func() { mock.AssertExpectations(t) })`, and returns it; verify by `go build ./test/mocks/...` clean
- [ ] 7.2 Add `func NewMockX(t *testing.T) *MockX` constructors for every mock type in `test/mocks/` that uses `.On(...)` in tests; verify by `go build ./test/mocks/...` clean
- [ ] 7.3 Update every test that creates a mock to call the new `NewMockX(t)` helper; verify by `rg 'mocks\.NewMock' test/ cmd/ internal/ --type go | wc -l` returning the same count as before the change (no test silently regresses)
- [ ] 7.4 Remove the 22 `.Maybe()` calls in `internal/service/worktree_service_test.go` and the 4 elsewhere across `internal/service/*_test.go`; verify by `rg '\.Maybe\(\)' test/ cmd/ internal/ --type go | wc -l` returning 0
- [ ] 7.5 Remove the 15 `mockX.ExpectedCalls = nil` resets in `internal/service/worktree_service_test.go`; verify by `rg 'ExpectedCalls = nil' test/ cmd/ internal/ --type go | wc -l` returning 0 and `go test ./internal/service/...` passing
- [ ] 7.6 Remove the 9 `var _ interface{} = mocks.NewMockXService()` declarations in `cmd/{create,delete,cd,list}_test.go`; verify by `rg 'var _ interface{}' cmd/ --type go | wc -l` returning 0 and `go build ./cmd/...` clean
- [ ] 7.7 Remove the 4 redundant `//nolint:wrapcheck` directives at `internal/infrastructure/command_executor_mock_test.go:24,26,33,35`; verify by `mise run lint` clean
- [ ] 7.8 Replace the 4 silent-pass `assert.NoError` + `t.Run` patterns at `internal/domain/errors_test.go:229→232`, `internal/domain/errors_test.go:242→245`, `internal/domain/config_test.go:29→32`, and `test/integration/cli_commands_test.go:84→89` with `require.NoError`; verify by the affected tests still passing

## 8. cmd.ErrOrStderr for logv progress (~14 sites)

- [ ] 8.1 Wire the new `io.Writer`-accepting verbose-log helper (from slice 3.3) to default to `c.ErrOrStderr()` when called from a cobra command; verify by `go test ./cmd/...` passing and a manual test that `cmd.SetErr(buffer)` captures the verbose output
- [ ] 8.2 Replace the 14 `_, _ = fmt.Fprint*` sites in `cmd/{util,init}.go` and the progress lines in `cmd/{prune,delete,create}.go` (Category C) with calls to the verbose-log helper; verify by `rg '_, _ = fmt\.F' cmd/{util,init,prune,delete,create}.go | wc -l` returning 14 fewer sites
- [ ] 8.3 Suppress errors from verbose-log calls (Category C is informational); verify by `go vet ./cmd/...` clean and tests still pass

## 9. fmt.Fprint per-site fixes (~28 sites)

- [ ] 9.1 Category A (navigation output, 4 sites): wrap `cmd/prune.go:127`, `cmd/delete.go:108`, `cmd/delete.go:187`, `cmd/create.go:125` in helper that captures the write error and returns it; verify by `go test ./cmd/...` passing and the navigation output still appears in success cases
- [ ] 9.2 Category B (warnings, ~12 sites): capture the error at `cmd/create.go:188-195` (failure summary) and `cmd/prune.go:103,127` and other warning sites; surface the error through the cmd layer's error path; verify by `go test ./cmd/...` passing
- [ ] 9.3 Category D (result detail, ~12 sites): in `cmd/prune.go:148-193` replace the local `errOut` variable with `c.ErrOrStderr()` directly; verify by `rg 'errOut' cmd/prune.go | wc -l` returning 0 and `go test ./cmd/...` passing
- [ ] 9.4 Category E (TTY prompt, 1 site): keep `cmd/prune.go:134` as-is; the `isatty` gate lands in slice 10.6; no change in this slice
- [ ] 9.5 Final sweep: `rg '_, _ = fmt\.F' cmd/ --type go | wc -l` should return 0; verify by the count

## 10. cmd modernization

- [ ] 10.1 Replace `ctx := context.Background()` at `cmd/create.go:61`, `cmd/delete.go:49`, `cmd/cd.go:45`, `cmd/prune.go:65`, `cmd/list.go:46`, `cmd/init.go:61` with `ctx := cmd.Context()`; verify by `rg 'context\.Background' cmd/{create,delete,cd,prune,list,init}.go | wc -l` returning 0
- [ ] 10.2 Set `SilenceUsage = true` and `SilenceErrors = true` on every leaf cobra command (`create`, `delete`, `cd`, `prune`, `list`, `init`, `completion`, `version`); verify by `rg 'SilenceUsage = true' cmd/ --type go | wc -l` returning ≥8 and the existing tests still pass
- [ ] 10.3 In `main.go`, wrap the cobra command execution in `signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)`; pass the resulting context to `cmd.ExecuteContext(ctx)`; verify by `go build ./...` clean and the existing e2e tests still pass
- [ ] 10.4 In `main.go`, replace the hardcoded `os.Exit(1)` with the exit code returned by `HandleCLIErrorWithCommand`; verify by manual run with a config error (`TWIGGIT_CONFIG=invalid` or similar) showing exit code 3
- [ ] 10.5 In `main.go`, replace the panic recovery handler's `fmt.Fprintf` with `slog.Error("panic recovered", "panic", r, "stack", string(debug.Stack()))` plus a one-line user-facing message on stderr; verify by `go vet ./...` clean and the existing tests still pass
- [ ] 10.6 In `cmd/prune.go`, gate `confirmBulkPrune`'s stdin read on `isatty(os.Stdin)`; when not a TTY and `--yes`/`--force` not set, exit with code 2 and a "TTY required" message; verify by manual run with `echo y | twiggit prune --all` exiting with code 2
- [ ] 10.7 In `main.go`, initialize `slog.Default()` with a handler that writes to `c.ErrOrStderr()` (resolved at first log call); log level INFO by default, DEBUG when `TWIGGIT_DEBUG=1`; verify by `go build ./...` clean

## 11. Infrastructure modernization

- [ ] 11.1 Migrate the 14 remaining `os.IsNotExist(err)` sites (`internal/infrastructure/{shell_infra.go:81,87,132,config_manager.go:77,context_detector.go:55,context_resolver.go:308,hook_runner.go:47,navigation_service.go:55}` and others Change 2 didn't touch) to `errors.Is(err, os.ErrNotExist)`; verify by `rg 'os\.IsNotExist' internal/ --type go | wc -l` returning 0
- [ ] 11.2 Replace the 3 `strings.HasPrefix + strings.TrimPrefix` patterns in `test/helpers/worktree.go:143-148,158-163,194-195` and `internal/infrastructure/cli_client.go:25-26,280-287,297-304` with `strings.CutPrefix`; verify by `rg 'TrimPrefix' test/helpers/worktree.go internal/infrastructure/cli_client.go | wc -l` returning 0
- [ ] 11.3 In `internal/infrastructure/context_resolver.go`, replace the 4 `context.Background()` sites (at the git service calls in `cr.gitService.ListWorktrees`, `cs.gitService.GetRepositoryStatus`, `cs.gitService.ListBranches`, `cs.gitService.ListWorktrees` in `getWorktreeContextSuggestions`) with the caller's `ctx`; verify by `rg 'context\.Background' internal/infrastructure/context_resolver.go | wc -l` returning 0
- [ ] 11.4 Split `internal/infrastructure/cli_client.go`'s 53-line `parseWorktreeList` into `parsePorcelainHeader(line) (porcelainHeader, bool)` and `accumulateField(current *porcelainWorktree, line)`; verify by `go test ./internal/infrastructure/...` passing and the new function sizes both ≤25 lines
- [ ] 11.5 Replace the local `ProjectRef` struct in `internal/infrastructure/context_resolver.go` with `domain.ProjectSummary`; verify by `rg 'ProjectRef' internal/ --type go | wc -l` returning 0
- [ ] 11.6 Bound the commit hash slice at `internal/infrastructure/gogit_client.go:336` with `min(7, len(hashStr))` instead of `hashStr[:7]`; verify by a unit test that calls `GetCommitInfo` with a hash shorter than 7 chars and observes no panic
- [ ] 11.7 Replace the remaining `_, _ = fmt.Fprint*` sites in `cmd/` (the 28 sites from S9 minus the 14 from S8) with category-appropriate handling; verify by `rg '_, _ = fmt\.F' cmd/ --type go | wc -l` returning 0
- [ ] 11.8 Preallocate the 6 nil-then-append slices via `make` at `cmd/error_formatter.go:88`, `internal/infrastructure/cli_client.go:270`, `internal/infrastructure/context_resolver.go:209,250,529`, `internal/service/navigation_service.go:83,118`, `internal/service/worktree_service.go:179`; verify by `rg '\[\]string\{' cmd/error_formatter.go internal/ --type go | wc -l` decreasing
- [ ] 11.9 In `internal/infrastructure/hook_runner.go`, replace the `sh -c "<exports>; <cmd>"` string concat with `exec.Command(cmd, args...)` and `cmd.Env = append(os.Environ(), envVars...)`; verify by `go test ./internal/infrastructure/hook_runner_test.go` passing and a manual test with a single quote in a path
- [ ] 11.10 Update 16 `for i := 0; i < N; i++` patterns across `test/concurrent/`, `test/helpers/`, `test/e2e/`, `internal/infrastructure/context_detector_test.go` to `for i := range N` (or `for range N` if the index is unused); verify by `rg 'for i := 0; i < ' test/ cmd/ internal/ --type go | wc -l` returning 0
- [ ] 11.11 Convert the 9 `interface{}` parameters/callbacks (`cmd/util.go:20,52`, `cmd/suggestions.go:67`, `test/helpers/performance.go:16,58`, `test/helpers/helpers_test.go:265,285`, the 9 `var _ interface{}` sites in `cmd/*_test.go` from S7.6) to `any`; verify by `rg 'interface\{\}' cmd/ test/ --type go | wc -l` returning 0 (after S7.6)
- [ ] 11.12 Replace 7 `wg.Add(1); go func(); defer wg.Done()` blocks in `test/concurrent/concurrent_test.go` (lines 124-145, 159-186, 226-242, 275-300, 302-..., 358-379, 381-...) with `wg.Go(func() { ... })`; verify by `go test -race ./test/concurrent/...` passing
- [ ] 11.13 Replace 2 local `func max(a, b int) int` shadows at `test/helpers/golden.go:117` and `test/e2e/golden_test.go:213` with `maxInt`; verify by `rg 'func max\(' test/ --type go | wc -l` returning 0
- [ ] 11.14 Convert 3 `panic()` calls in `test/helpers/repo.go:47,116` and `test/helpers/git.go:46` to `t.Fatal` (signature change to helpers requires updating callers); verify by `rg 'panic\(' test/helpers/ --type go | wc -l` returning 0 and `go test ./test/...` passing
- [ ] 11.15 Add a `build-once TestMain` at `main_test.go` that compiles the binary into `t.TempDir()` once, stashes the path in a package var, and exports it via `TWIGGIT_E2E_BINARY` env var; verify by `go test -tags=e2e ./test/e2e/...` running and the build happening exactly once (count `go build` invocations)
- [ ] 11.16 Add `goleak.VerifyTestMain` after `os.Exit(m.Run())` in `main_test.go` (or in a sub-package's `TestMain`); verify by a deliberate goroutine-leak test causing the suite to fail with the goroutine's stack trace

## 12. Ctor refactor (12 4-string ctors)

- [ ] 12.1 Convert `domain.NewValidationError(request, field, value, message string)` at `internal/domain/service_errors.go:60` to `domain.NewValidationError(spec ValidationErrorSpec)` with struct literal; verify by `go build ./...` clean
- [ ] 12.2 Convert `domain.NewWorktreeServiceError(worktreePath, branchName, operation, message string, cause error)` at `internal/domain/service_errors.go:139` to struct literal; verify by `go build ./...` clean
- [ ] 12.3 Convert `domain.NewProjectServiceError`, `domain.NewNavigationServiceError`, `domain.NewConflictError`, `domain.NewResolutionError`, `domain.NewShellError`, `domain.NewShellErrorWithCause`, `domain.NewServiceError`, `domain.NewGitCommandError`, `domain.NewGitWorktreeError`, `domain.NewContextDetectionError`, `domain.NewGitRepositoryError`, `domain.NewConfigError` (the rest) to struct literals; verify by `go build ./...` clean and `go vet ./...` clean
- [ ] 12.4 Update every call site in `internal/service/` and `cmd/` to use struct literal construction; verify by `rg 'NewValidationError\(' internal/ cmd/ --type go | wc -l` matching the number of struct-literal call sites
- [ ] 12.5 Add a per-ctor godoc comment naming each field's purpose (since the struct literal no longer has parameter names); verify by `rg 'NewValidationError' internal/domain/service_errors.go -A 5` showing godoc on the type

## 13. Free-function refactor (25 methods)

- [ ] 13.1 Convert 6 stateless methods in `internal/service/worktree_service.go` (`validateDeleteRequest`, `findProjectByWorktree`, `findMainRepoFromConfig`-style, `checkWorktreeSkip`, `findMainRepo*`, `filterSuggestionsForProject`-style) to free functions; verify by `go build ./...` clean and the methods' call sites updated via `gopls rename`
- [ ] 13.2 Convert the remaining 19 receiver-less methods across `internal/service/worktree_service.go`, `internal/service/project_service.go` (`findMainRepo*`), `internal/service/navigation_service.go` (`filterSuggestionsForProject`) to free functions; verify by `rg 'func \(s \*[A-Z][a-zA-Z]*\)' internal/service/ --type go | wc -l` decreasing
- [ ] 13.3 Sweep dead code per `openspec/dead-code.md`: remove unused struct fields, dead config keys, and orphan functions flagged by `staticcheck` (Unchecked loop, deadcode, unused linters); verify by `staticcheck ./...` clean
- [ ] 13.4 Verify the receiver-less refactor preserves the public interface: every `WorktreeService`/`ProjectService`/`NavigationService`/`ShellService` method on the interface is still implemented; verify by `mise run test` passing and `rg 'func \(.*WorktreeService\)' internal/service/worktree_service.go | wc -l` matching the interface method count

## 14. Docs

- [ ] 14.1 Add godoc comments to the 26 `Error()`/`Unwrap()` methods in `internal/domain/errors.go` (12 sites) and `internal/domain/service_errors.go` (14 sites); verify by `rg -L '//.*\nfunc \(e \*\w+Error\) Error' internal/domain/*.go | wc -l` returning 0 sites lacking a doc comment
- [ ] 14.2 Add godoc to the 4 exported methods in `internal/infrastructure/`: `DetectContext` (context_detector.go:47), `ResolveIdentifier` (context_resolver.go:182), `GetResolutionSuggestions` (context_resolver.go:203), `Run` (hook_runner.go:37); verify by `rg 'func .*\) (DetectContext|ResolveIdentifier|GetResolutionSuggestions|Run)\(' internal/infrastructure/ -B 1 | rg -v '^//'` returning 0 unmatched lines
- [ ] 14.3 Add godoc to the 3 nit sites: `HookResult.Executed`/`Success` (internal/domain/hook_types.go:21), `NormalizePath` (internal/infrastructure/pathutils.go:33 — wait, this moves to domain per Change 2; verify location), `version.go.String()`; verify by per-site godoc
- [ ] 14.4 Create `llms.txt` at the repository root with the project description, primary commands, and links to the spec catalog and AGENTS.md; verify by the file existing at `/llms.txt`
- [ ] 14.5 Fix the README broken `twiggit init ~/.zshrc` snippet (line 45) and the `twiggit init --shell=zsh` snippet (line 47) per the actual init command surface; verify by `rg 'init ~/.zshrc|init --shell' README.md | wc -l` returning 0
- [ ] 14.6 Add missing README sections: Demo, Features/Specification, Contributing link, Contributors, License link; verify by `rg '^## (Demo|Features|Contributors|License)' README.md | wc -l` returning 4
- [ ] 14.7 Fix the Go version mismatch between `CONTRIBUTING.md` (says `Go 1.25+`) and the latest CHANGELOG entry (says `1.26.1`); align both to the project's actual target (1.25.5 per `go.mod`); verify by `rg 'Go 1\.' CONTRIBUTING.md CHANGELOG.md` showing 1.25 in both
- [ ] 14.8 Normalize the CHANGELOG double blank lines at lines 8-10 (0.12.0) and 31 (0.11.0) to single blank lines; verify by `rg -B 1 '^$' CHANGELOG.md | wc -l` returning the expected single-blank pattern
- [ ] 14.9 Change `internal/version/version.go` to use ldflag-set unexported `version`, `commit`, `date` variables (uppercase names `Version`, `Commit`, `Date` → `version`, `commit`, `date`); update `String()` to trim trailing space; verify by `go build ./internal/version/...` clean and `version.go` no longer has package-level mutable exported vars

## 15. Dependencies

- [ ] 15.1 `go get -u=patch github.com/go-git/go-git/v5` and verify by `rg 'go-git/v5' go.mod` showing the bumped version; manually resolve any test breakages from the bump
- [ ] 15.2 `go get -u=patch github.com/stretchr/testify` and verify by `rg 'stretchr/testify' go.mod` showing the bumped version
- [ ] 15.3 `go get -u=patch golang.org/x/{crypto,net,sys,text,sync}` and verify by `go.sum` updating and `go build ./...` clean
- [ ] 15.4 Drop the duplicate `gopkg.in/yaml.v3` fork (keep `go.yaml.in/yaml/v3` only — the project resolves through `knadh/koanf` which uses the latter); verify by `rg 'gopkg.in/yaml.v3' go.mod go.sum | wc -l` returning 0
- [ ] 15.5 Add the `tool` directive for `golang.org/x/vuln/cmd/govulncheck` via `go get -tool golang.org/x/vuln/cmd/govulncheck@latest`; verify by `rg '^tool ' go.mod` showing the new directive
- [ ] 15.6 Pin the toolchain explicitly via the `// toolchain` directive in `go.mod` (or verify the existing `go 1.25.5` directive is the intended pin); verify by `go env GOTOOLCHAIN` matching
- [ ] 15.7 Add a `retract` block in `go.mod` covering any versions with known issues (consult the project's CVE list); verify by `rg '^retract ' go.mod` showing entries
- [ ] 15.8 Capture the govulncheck baseline: `go tool govulncheck ./...` and commit `govulncheck.json` as the baseline; verify by `govulncheck.json` existing at the repo root
- [ ] 15.9 `go mod tidy` and verify by `git diff go.mod go.sum` showing only the intentional changes (tool directive, drops, retract block)

## 16. CI / Docker / Goreleaser / Mise

- [ ] 16.1 `.gitlab-ci.yml`: add main-branch rules to `lint`, `test`, and `setup-cache` jobs so direct pushes to main run the full validation; verify by the YAML lint clean and a manual push to a test branch showing the new rules fire
- [ ] 16.2 `.gitlab-ci.yml`: add `-race` to the test job's `go test` invocation; verify by `rg '\-race' .gitlab-ci.yml | wc -l` returning ≥1
- [ ] 16.3 `.gitlab-ci.yml`: add the govulncheck step (after `mise run verify`) that runs `go tool govulncheck ./...` and uploads `govulncheck.json` as an artifact; verify by the YAML adding the new job
- [ ] 16.4 `.gitlab-ci.yml`: set `interruptible: false` on the release job explicitly; verify by `rg 'interruptible' .gitlab-ci.yml` showing the release job
- [ ] 16.5 `.gitlab-ci.yml`: add `resource_group: twiggit-ci` to lint/test/setup-cache jobs so concurrent pipelines serialize through cache writes; verify by `rg 'resource_group' .gitlab-ci.yml` matching the expected count
- [ ] 16.6 `.gitlab-ci.yml`: fix the mirror-to-github step to NOT embed `GITHUB_ACCESS_TOKEN` in the remote URL; use a credential helper or env-based auth; verify by `rg 'GITHUB_ACCESS_TOKEN' .gitlab-ci.yml` returning only the credential helper invocation (not a URL substring)
- [ ] 16.7 `.gitlab-ci.yml`: drop `--force` from the `git push --tags --force` step in the mirror job; verify by `rg '\-\-force' .gitlab-ci.yml` returning 0 in the mirror section
- [ ] 16.8 `Dockerfile.ci`: pin the base image by sha256 digest in addition to the tag (`golang:1.25.5-alpine3.23@sha256:...`); verify by `rg 'sha256:' Dockerfile.ci` showing the pinned digest
- [ ] 16.9 `Dockerfile.ci`: replace `curl https://mise.run | sh` with a release tarball download + sha256 verification step; verify by `rg 'curl.*mise\.run' Dockerfile.ci` returning 0
- [ ] 16.10 `.goreleaser.yml:73`: change `owner: amoconst` to `owner: amauryconstant`; verify by `rg 'owner:' .goreleaser.yml` showing the new value
- [ ] 16.11 `.goreleaser/hooks/create-github-release.sh`: add `set -euo pipefail`, replace `%s` with `%B` for the commit subject, quote `${TAG}`, and remove the `|| echo "Failed..."` swallow (let the script fail); verify by `rg 'echo.*Failed' .goreleaser/hooks/create-github-release.sh` returning 0 and `bash -n .goreleaser/hooks/create-github-release.sh` clean
- [ ] 16.12 `.mise/tasks/ci/coverage.sh`: replace the `awk "exit !(...)"` warning with an `exit 1` when coverage falls below 70%; verify by a deliberate test run with low coverage exiting non-zero
- [ ] 16.13 `.mise/config.toml`: add `-shuffle=on` to `test:unit`, `test:race`, and `test:e2e` tasks; verify by `rg '\-shuffle=on' .mise/config.toml` returning ≥3

## 17. Verification and archive

- [ ] 17.1 Run `mise run verify` and verify lint + format + race + golden all pass; report any failures in the slice commit message
- [ ] 17.2 Run `mise run test` (full suite including integration and concurrent) and verify the suite is green; report any failing tests
- [ ] 17.3 Run `openspec validate quality-modernization --json` and verify `valid: true, issues: []`
- [ ] 17.4 Run `git grep -n '_, _ = fmt\.'` and verify 0 matches in `cmd/` (no fmt.Fprint regressions)
- [ ] 17.5 Run `git grep -n 'os\.IsNotExist'` and verify 0 matches in `internal/`
- [ ] 17.6 Run `git grep -n 'context\.Background'` in `test/` and `cmd/` and verify 0 matches (the only remaining `context.Background()` use is for production code that genuinely wants a never-cancelled context, e.g., daemon-style listeners — none in this project)
- [ ] 17.7 Run `git grep -n 'Impl.*= nil'` in `internal/infrastructure/` and verify 0 matches (naming sweep completeness)
- [ ] 17.8 Run `git grep -n 'infrastructure\.' internal/service/` and verify 0 matches (layer inversion still holds)
- [ ] 17.9 Run `openspec-generate-changelog` and verify the new entry appears at the top of `CHANGELOG.md`
- [ ] 17.10 Run `openspec-archive-change quality-modernization --json` and verify the change moves from `openspec/changes/quality-modernization/` to `openspec/changes/archive/<date>-quality-modernization/` and the 16 spec deltas merge into `openspec/specs/<capability>/spec.md`
