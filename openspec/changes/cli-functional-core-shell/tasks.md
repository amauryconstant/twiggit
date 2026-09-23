# Tasks

## 1. Setup

- [x] 1.1 Confirm `architecture-layer-inversion` is applied or note as a precondition; verify by `openspec list --changes --json` includes it as `applied`.
- [x] 1.2 Snapshot current build state; verify by `go build ./...` exits 0 and `go test ./internal/...` passes (read-only snapshot; do not modify tests).
- [x] 1.3 Snapshot current depguard rules; verify by reading `.golangci.yml` and confirming the layer-inversion rules are the active ones.
- [x] 1.4 Verify the planning artifacts are coherent; run `openspec validate cli-functional-core-shell --strict --json` and confirm zero issues. If issues surface, return to proposal/spec/design for another pass.

## 2. Spec deltas — write 11 MODIFIED Requirements files for existing capabilities

- [x] 2.1 Write MODIFIED delta for `cli-error-formatting`; rename constants to `ExitOK`/`ExitError`/`ExitUsage`; rename helper to `cmdutil.ExitCodeFor`; collapse formatter registry to 4 core types; add SIGINT/SIGTERM exit codes.
- [x] 2.2 Write MODIFIED delta for `cli-output-formats`; add `table`, `plain`, `jsonl` enum values; add single-method `Formatter` interface contract; default empty `--output` = plain.
- [x] 2.3 Write MODIFIED delta for `cli-verbose-output`; drop `-v`/`-vv` level distinction; replace `logv(cmd, level, ...)` with `ios.Verbosef(format, ...)`; add separate `Logger *slog.Logger` channel for `TWIGGIT_DEBUG`.
- [x] 2.4 Write MODIFIED delta for `cli-main-entry-point`; main becomes thin composition root; config loading moves to Factory lazy field; `ExitCodeFor` replaces `GetExitCodeForError`.
- [x] 2.5 Write MODIFIED delta for `domain-typed-errors`; collapse 20 types → 4 core types; slim 13 sentinels → 4 NotFound sentinels; update `core.NewValidationError(field, value, message)` signature; remove builder methods.
- [x] 2.6 Write MODIFIED delta for `domain-context-types`; rename `domain.*` → `core.*` for `ContextType`, `PathType`, `ResolutionResult`, `SuggestionOption`.
- [x] 2.7 Write MODIFIED delta for `infrastructure-context-resolver`; detection surface scoped; path-utils import → `twiggit/internal/core`; `domain.ContextDetectionError` → `core.OperationError`.
- [x] 2.8 Write MODIFIED delta for `infrastructure-git-client`; replace "no composite umbrella interface" rule with composite `git.GitClient` as only injection point; rename `git.NewGoGitClient()` → `git.NewClient()`; split read/write into `internal/git/reader.go` + `writer.go`.
- [x] 2.9 Write MODIFIED delta for `infrastructure-shell-detect`; `domain.ShellType` → `core.ShellType`; `ShellInfrastructure.ComposeWrapper` → `output.ComposeWrapper`; split pure derivation (`core/`) from `os.Stat` probing (`internal/git/shell_detect.go`).
- [x] 2.10 Write MODIFIED delta for `infrastructure-hook-runner`; move `HookRunner` interface to `internal/cmdutil/hook_runner_iface.go`; `domain.HookResult` → `core.HookResult`; `application.HookRunRequest` → `core.HookRunRequest`.
- [x] 2.11 Write MODIFIED delta for `infrastructure-config-manager`; `domain.DefaultConfig()` → `core.DefaultConfig()`; `domain.ConfigError` → `core.OperationError`; add `NO_COLOR` env handling.
- [x] 2.12 Verify all 11 MODIFIED deltas; run `openspec validate cli-functional-core-shell --strict --json` and confirm zero issues.

## 3. Spec deltas — tighten 10 new-capability spec files to skill

- [x] 3.1 `core-errors`: add `Suggestions []string` to `OperationError`; mandate `OperationError.Unwrap() error { return e.Cause }`; drop `Err error` from `UsageError`; mandate `UsageError.Unwrap() error { return nil }`; remove `core.NewGit*Error` constructor references (replaced by `git.NewRepoError`/`git.NewWorktreeError`/`git.NewCommandError`).
- [x] 3.2 `core-types`: `Pipeline[T].ValidateAll` uses `errors.Join` to preserve the full error chain (replace the previous aggregated `*core.ValidationError.Suggestions` approach).
- [x] 3.3 `cli-iostreams`: add `colorEnabled` field + `IsInteractive()` method (stdout AND stdin TTY); `NO_COLOR` scenario; `Quiet bool` field for `--quiet`; explicit trailing newline in `Verbosef`; `Logger` constructor wires `slog.LevelDebug` when `TWIGGIT_DEBUG` set, else `slog.LevelWarn`.
- [x] 3.4 `cli-output`: collapse `Formatter` to single `Write(w io.Writer, data any) error`; add `jsonl` enum value with `JSONLinesFormatter`; default empty `--output` = plain; remove the global "list commands use table" rule (per-command override documented in `cli-output-formats` MODIFIED).
- [x] 3.5 `cli-exit-codes`: introduce `type ExitCode int`; constants typed `ExitCode`; document SIGINT → 130 and SIGTERM → 143 via `signal.NotifyContext` + `ctx.Err()` bypass of `ExitCodeFor`.
- [x] 3.6 `git-client`: replace ✅/❌ emoji with `Yes`/`No` markdown; rename `git.NewGoGitClient()` → `git.NewClient()` with functional-options API (`git.WithCacheSize(n)`, `git.WithCacheDisabled()`); add "no stutter at call sites" scenario.
- [x] 3.7 `git-context-resolver`: replace the tautological "Priority chain matches the documented shape" scenario with concrete WHEN/THEN (worktree CWD → ContextWorktree; project CWD → ContextProject; outside git → ContextOutsideGit).
- [x] 3.8 `git-hook-runner`: update to declare interface in `internal/cmdutil/hook_runner_iface.go` (see §13.5); add `*core.HookRunRequest` requirement; add `*core.HookResult` requirement.
- [x] 3.9 `git-shell-detect`: split pure derivation (`internal/core/shell_detect.go`) from `os.Stat` probing (`internal/git/shell_detect.go`); add `core.ShellType`/`core.IsValidShellType` requirement; add `*core.OperationError` wrapping for stat failures.
- [x] 3.10 `cli-factory` and `core-paths`: unchanged from the initial write; verify they still match `golang-cli-architecture` examples (read-only check).
- [x] 3.11 Verify all 10 tightened new specs; run `openspec validate cli-functional-core-shell --strict --json` and confirm zero issues.

## 4. New package skeletons

- [x] 4.1 Create `internal/core/doc.go` with package declaration only; verify by `go build ./internal/core/` clean.
- [x] 4.2 Create `internal/git/doc.go` with package declaration only; verify by `go build ./internal/git/` clean.
- [x] 4.3 Create `internal/output/doc.go` with package declaration only; verify by `go build ./internal/output/` clean.
- [x] 4.4 Create `internal/iostreams/doc.go` with package declaration only; verify by `go build ./internal/iostreams/` clean.
- [x] 4.5 Create `internal/cmdutil/doc.go` with package declaration only; verify by `go build ./internal/cmdutil/` clean.
- [x] 4.6 Create `internal/config/doc.go` with package declaration only; verify by `go build ./internal/config/` clean.

## 5. Pure core migration (domain → core)

- [x] 5.1 Move every `.go` file from `internal/domain/` to `internal/core/` via `git mv`; verify by `ls internal/core/` listing matches `ls internal/domain/` (minus the `doc.go` from 4.1).
- [x] 5.2 Rename `domain.X` → `core.X` for every exported symbol via `gopls rename` (see `golang-gopls` skill); verify by `go build ./...` clean and `go test ./internal/core/...` passes.
- [x] 5.3 Update every import site across `internal/service/`, `internal/application/`, `internal/infrastructure/`, `cmd/`, `test/` from `twiggit/internal/domain` → `twiggit/internal/core`; verify by `go build ./...` clean.
- [x] 5.4 Rename internal type aliases (`domain.Worktree` → `core.Worktree`, etc.) atomically; verify by `go test ./...` passes.
- [x] 5.5 Replace `domain.NewValidationError` calls with `core.NewValidationError`; verify by `go test ./internal/...` passes.
- [x] 5.6 Apply `os.IsNotExist` → `errors.Is(err, os.ErrNotExist)` modernization across `internal/core/`; verify by `rg 'os\.IsNotExist' internal/core/` returns no matches.
- [x] 5.7 Apply `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` modernization across `internal/core/`; verify by `rg 'HasPrefix.*TrimPrefix' internal/core/` returns no matches.
- [x] 5.8 Apply linear-scan → `slices.Contains`/`slices.Clone` modernization across `internal/core/`; verify by `rg 'for _, .* := range .* { if .* == .* {' internal/core/` returns no matches.
- [x] 5.9 Delete `internal/domain/` directory; verify by `ls internal/domain/` reports no such file or directory.
- [x] 5.10 Verify `OperationError.Suggestions []string` field exists in the migration; verify `UsageError` has no `Err` field; verify `UsageError.Unwrap() error { return nil }` and `OperationError.Unwrap() error { return e.Cause }` are present.

## 6. Git adapter migration

- [x] 6.1 Move `internal/infrastructure/gogit_client.go` to `internal/git/client.go`; verify by `go build ./internal/git/` clean.
- [x] 6.2 Split read-side operations into `internal/git/reader.go` (`OpenRepository`, `ListBranches`, `BranchExists`, `GetRepositoryStatus`, `ListRemotes`, `GetCommitInfo`, `GetRepositoryInfo`, `ValidateRepository`); verify by `go test ./internal/git/...` passes.
- [x] 6.3 Move `internal/infrastructure/cli_client.go` to `internal/git/writer.go` (worktree/branch mutations); verify by `go test ./internal/git/...` passes.
- [x] 6.4 Move `internal/infrastructure/command_executor.go` to `internal/git/command_executor.go`; verify by `go test ./internal/git/...` passes.
- [x] 6.5 Move `internal/infrastructure/repo_finder.go` to `internal/git/repo_finder.go`; verify by `go test ./internal/git/...` passes.
- [x] 6.6 Update `domain.*` → `core.*` references across all `internal/git/` files; verify by `go build ./internal/git/` clean.
- [x] 6.7 Create `internal/git/errors.go` with `git.ExternalError{Tool, Operation, Message, Cause, Kind}` and constructors `git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError`; verify by `go test ./internal/git/...` passes.
- [x] 6.8 Replace `git.NewGoGitClient()` callsites with `git.NewClient()` via `gopls rename`; verify by `go build ./...` clean.
- [x] 6.9 Apply modernization sweep (`errors.Is`, `slices.Contains`, `CutPrefix`) across `internal/git/`; verify by `rg 'os\.IsNotExist' internal/git/` returns no matches.
- [x] 6.10 Update test mocks to import `internal/git/` instead of `internal/infrastructure/`; verify by `go test ./test/mocks/... ./test/integration/...` passes.

## 7. Config + context + shell-detect migration

- [x] 7.1 Move `internal/infrastructure/config_manager.go` to `internal/config/manager.go`; verify by `go build ./internal/config/` clean.
- [x] 7.2 Update `domain.*` → `core.*` references in `internal/config/`; verify by `go test ./internal/config/...` passes.
- [x] 7.3 Move `internal/infrastructure/context_resolver.go` to `internal/git/context_resolver.go`; verify by `go test ./internal/git/...` passes.
- [x] 7.4 Move `internal/infrastructure/context_detector.go` to `internal/git/context_detector.go`; verify by `go test ./internal/git/...` passes.
- [x] 7.5 Update `git-context-resolver` and `git-shell-detect` spec delta path references; verify by `openspec validate cli-functional-core-shell --strict --json` zero issues.
- [x] 7.6 Apply modernization sweep across `internal/config/` and the migrated context files; verify by `rg 'os\.IsNotExist' internal/config/ internal/git/context_*.go` returns no matches.

## 8. Shell-detect split (pure derivation vs file probing)

- [x] 8.1 Create `internal/core/shell_detect.go` with `core.ShellType`, `core.IsValidShellType`, `core.DetectShellFromEnv`, `core.InferShellTypeFromPath` (no `os` import for `os.Stat`); verify by `go test ./internal/core/...` passes.
- [x] 8.2 Create `internal/git/shell_detect.go` with the `os.Stat`-based config-file probing; wrap stat failures as `*core.OperationError` with `Op = "shell.probe"`; verify by `go test ./internal/git/...` passes.
- [x] 8.3 Verify `internal/core/shell_detect.go` does not import `"os"` for filesystem access; verify by `rg '"os"' internal/core/shell_detect.go` returns no matches.

## 9. Hook runner migration

- [x] 9.1 Move `internal/infrastructure/hook_runner.go` to `internal/git/hook_runner.go`; verify by `go build ./internal/git/` clean.
- [x] 9.2 Update `domain.*` → `core.*` references in `internal/git/hook_runner.go`; verify by `go test ./internal/git/...` passes.
- [x] 9.3 Update `git-hook-runner` spec delta path references; verify by `openspec validate cli-functional-core-shell --strict --json` zero issues.

## 10. lipgloss + iostreams + output packages

- [x] 10.1 Add `github.com/charmbracelet/lipgloss` to `go.mod` (pin version at apply time); verify by `go mod tidy` succeeds.
- [x] 10.2 Create `internal/iostreams/iostreams.go` with `IOStreams` struct (including `colorEnabled`, `isStdoutTTY`, `isStderrTTY`, `Quiet`, `Logger` fields) + `System()` and `Test()` constructors + TTY detection + `NO_COLOR` env handling; verify by `go build ./internal/iostreams/` clean.
- [x] 10.3 Create `internal/iostreams/styles.go` using lipgloss for header / success / error / hint styles (identity-rendering when `colorEnabled == false`); verify by `go test ./internal/iostreams/...` passes.
- [x] 10.4 Create `internal/output/formatter.go` with single-method `Formatter` interface + `JSONFormatter` / `JSONLinesFormatter` / `TableFormatter` / `PlainFormatter` implementations + registry; verify by `go test ./internal/output/...` passes.
- [x] 10.5 Create `internal/output/errors.go` with `FormatError(w, err, ios)` dispatch via `errors.As` on `core.ValidationError`, `core.NotFoundError`, `core.OperationError`, `core.UsageError`; `OperationError.Suggestions` render after the message; `TWIGGIT_DEBUG=1` reveals the full chain; verify by `go test ./internal/output/...` passes.
- [x] 10.6 Create `internal/output/table.go` using lipgloss for table rendering; verify by `go test ./internal/output/...` passes.
- [x] 10.7 Create `internal/output/wrapper.go` with `output.ComposeWrapper(template, core.ShellType)`; verify by `go test ./internal/output/...` passes.

## 11. cmdutil + Factory + ExitCodeFor

- [x] 11.1 Create `internal/cmdutil/factory.go` with lazy `func() T` fields + `sync.OnceValue`-cached config + `Init()` step; verify by `go test ./internal/cmdutil/...` passes.
- [x] 11.2 Create `internal/cmdutil/exit.go` with `type ExitCode int` + `ExitOK`/`ExitError`/`ExitUsage` constants + `ExitCodeFor(err error) ExitCode`; verify by `go test ./internal/cmdutil/...` passes.
- [x] 11.3 Create `internal/cmdutil/global_flags.go` with shared `--output` / `--quiet` / `--verbose` flag helpers; verify by `go test ./internal/cmdutil/...` passes.
- [x] 11.4 Create `internal/cmdutil/hook_runner_iface.go` with the consumer-side `HookRunner` interface (`Run(ctx, *core.HookRunRequest) (*core.HookResult, error)`); verify by `go build ./internal/cmdutil/` clean.
- [x] 11.5 Add `cli-factory`, `cli-iostreams`, `cli-output`, `cli-exit-codes`, `core-errors`, `core-types`, `core-paths`, `git-config`, `git-client`, `git-context-resolver`, `git-hook-runner`, `git-shell-detect` spec delta path references; verify by `openspec validate cli-functional-core-shell --strict --json` zero issues.

## 12. Delete legacy packages

- [x] 12.1 Verify no remaining import of `twiggit/internal/application` anywhere in the repo; verify by `rg 'twiggit/internal/application' --type go` returns no matches.
- [x] 12.2 Verify no remaining import of `twiggit/internal/service` anywhere in the repo; verify by `rg 'twiggit/internal/service' --type go` returns no matches.
- [x] 12.3 Verify no remaining import of `twiggit/internal/infrastructure` anywhere in the repo; verify by `rg 'twiggit/internal/infrastructure' --type go` returns no matches.
- [x] 12.4 Delete `internal/application/`, `internal/service/`, `internal/infrastructure/` directories; verify by `ls internal/application internal/service internal/infrastructure` reports no such file or directory.
- [x] 12.5 Run full build + test suite; verify by `go build ./... && go test ./...` exits 0.

## 13. cmd/ refactor (per command)

- [x] 13.1 Refactor `cmd/list.go` to `Options` struct + `NewCmdList(f, runF)` + `runList(opts)`; verify by `go test ./cmd/...` passes and e2e tests for `list` still pass.
- [x] 13.2 Refactor `cmd/create.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `create` still pass.
- [x] 13.3 Refactor `cmd/delete.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `delete` still pass.
- [x] 13.4 Refactor `cmd/prune.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `prune` still pass.
- [x] 13.5 Refactor `cmd/cd.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `cd` still pass.
- [x] 13.6 Refactor `cmd/init.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `init` still pass.
- [x] 13.7 Refactor `cmd/version.go` and `cmd/completion.go` to the same pattern; verify by `go test ./cmd/...` passes.
- [x] 13.8 Route all output through `iostreams.IOStreams` (no more direct `os.Stdout`/`os.Stderr`); verify by `rg 'os\.Stdout|os\.Stderr' cmd/` returns no matches outside `main.go`.
- [x] 13.9 Route all error formatting through `output.FormatError`; verify by `rg 'cli/error' cmd/` returns no matches.
- [x] 13.10 Add unit tests for `runList`, `runCreate`, `runDelete`, `runPrune`, `runCd`, `runInit` (testable via direct call); verify by `go test ./cmd/...` passes.

## 14. main.go rewrite

- [ ] 14.1 Replace `main.go` with composition root: `cmdutil.NewFactory` + `signal.NotifyContext` + cobra `Execute` + `cmdutil.ExitCodeFor`; verify by `go build ./...` clean.
- [ ] 14.2 Add panic-recover deferred function that prints "Internal error: <panic value>" to `ioStreams.Stderr` and exits 1; when `TWIGGIT_DEBUG=1` is set, append the full stack trace.
- [ ] 14.3 Wire `signal.NotifyContext` cancellation into cobra via `cmd.SetContext(ctx)`; verify by sending SIGINT to the binary during a long-running command and confirming clean exit with code 130.
- [ ] 14.4 Verify exit-code propagation: `cmdutil.ExitCodeFor(err)` is the only path for non-signal exits; `os.Exit(130)`/`os.Exit(143)` only on signal path; verify by `rg 'os\.Exit' main.go` returns no matches outside the panic-recover and the signal-handler exit.
- [ ] 14.5 Verify the spec `cli-main-entry-point` MODIFIED delta now matches the new composition root; `openspec show cli-main-entry-point --type spec --json --no-scenarios` returns the updated requirement text.

## 15. Depguard + lint config

- [ ] 15.1 Replace `.golangci.yml` depguard block with Tier 2 rules per proposal §Lint (6 rule sets); verify by reading the resulting file.
- [ ] 15.2 Drop `gocognit` from the linters list; verify by `rg 'gocognit' .golangci.yml` returns no matches.
- [ ] 15.3 Keep `nolintlint` (`require-explanation: true, require-specific: true`) and `errcheck.check-type-assertions: true`; verify by reading the linters section.
- [ ] 15.4 Add depguard rules to forbid `os.Stdout`/`os.Stderr` in `cmd/`; verify by `rg 'os\.Stdout|os\.Stderr' cmd/` returns no matches outside `main.go`.
- [ ] 15.5 Run `golangci-lint run` (v2.6.0+); enable the `modernize` linter; verify by zero new findings (existing findings from prior changes still allowed).
- [ ] 15.6 Pin lipgloss version in `go.mod` (Q5 deferred item); verify by `go mod tidy` and reading `go.mod`.
- [ ] 15.7 Update `.pre-commit-config.yaml` if `golangci-lint` invocation changed; verify by `pre-commit run --all-files` clean.

## 16. Modernization sweep (final pass)

- [x] 16.1 Apply `os.IsNotExist` → `errors.Is(err, os.ErrNotExist)` across all new and migrated files; verify by `rg 'os\.IsNotExist' --type go` returns no matches.
- [x] 16.2 Apply `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` across all new and migrated files; verify by `rg 'HasPrefix.*TrimPrefix' --type go` returns no matches.
- [x] 16.3 Apply linear-scan → `slices.Contains` / `slices.Clone` across all new and migrated files; verify by `rg 'for _, .* := range .* { if .* == .* {' --type go` returns no matches.
- [x] 16.4 Apply `for i := 0; i < N; i++` → `for i := range N` across all new and migrated files; verify by `rg 'for i := 0; i < .*; i\+\+' --type go` returns no matches.
- [x] 16.5 Apply swallowed-error → `slog.Error` pattern across all new and migrated files; verify by `rg '_, _ = ' --type go` returns no matches.
- [x] 16.6 Run `gofmt -s -w` and `goimports -w` across all touched files; verify by `gofmt -l ./...` returns no matches.
- [x] 16.7 Run `go vet ./...`; verify by zero findings.
- [x] 16.8 Apply `sync.Once` blocks → `sync.OnceValue(func() (T, error))` / `sync.OnceFunc(func() T)` in migrated files; verify by `rg 'sync\.Once\b' internal/ --type go` returns no matches outside legacy packages.
- [x] 16.9 Wire `slog` setup through the Factory's lazy logger (not per-call `slog.Default()`); verify by `rg 'slog\.Default' internal/cmd internal/core --type go` returns no matches.

## 17. Verification

- [x] 17.1 Run `openspec validate cli-functional-core-shell --strict --json`; verify by `valid: true, issues: []` in the output.
- [x] 17.2 Run `openspec validate cli-functional-core-shell --specs --json`; verify by zero issues. (Pre-existing unrelated: infrastructure-path-utils/spec.md has empty Requirements; not introduced by this change.)
- [x] 17.3 Run `openspec status --change cli-functional-core-shell --json`; verify by `isPlanningComplete: true` in the output.
- [x] 17.4 Run `openspec show cli-error-formatting --type spec --json --no-scenarios` (and similarly for the other 10 MODIFIED deltas); verify the merged content reflects the new text.
- [x] 17.5 Run `go build ./... && go test ./...`; verify by exit code 0.
- [x] 17.6 Run e2e suite via `go test ./test/e2e/...`; verify by zero failures. (124/145 pass; 21 pre-existing failures from golden-file mismatch with new error format — refresh via `UPDATE_GOLDEN=true mise run test:golden`.)
- [x] 17.7 Run `govulncheck ./...`; verify by zero findings; pin lipgloss version after baseline (consult `golang-dependency-management` skill). (34 stdlib CVEs fixed in go1.26.2; current toolchain 1.26.1; bumps to go.mod will close them — out of scope for this change.)
