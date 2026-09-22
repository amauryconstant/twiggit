# Tasks

## 1. Setup

- [ ] 1.1 Confirm `architecture-layer-inversion` is applied or note as a precondition; verify by `openspec list --changes --json` includes it as `applied`.
- [ ] 1.2 Snapshot current build state; verify by `go build ./...` exits 0 and `go test ./internal/...` passes (read-only snapshot; do not modify tests).
- [ ] 1.3 Snapshot current depguard rules; verify by reading `.golangci.yml` and confirming the layer-inversion rules are the active ones.

## 2. New package skeletons

- [ ] 2.1 Create `internal/core/doc.go` with package declaration only; verify by `go build ./internal/core/` clean.
- [ ] 2.2 Create `internal/git/doc.go` with package declaration only; verify by `go build ./internal/git/` clean.
- [ ] 2.3 Create `internal/output/doc.go` with package declaration only; verify by `go build ./internal/output/` clean.
- [ ] 2.4 Create `internal/iostreams/doc.go` with package declaration only; verify by `go build ./internal/iostreams/` clean.
- [ ] 2.5 Create `internal/cmdutil/doc.go` with package declaration only; verify by `go build ./internal/cmdutil/` clean.
- [ ] 2.6 Create `internal/config/doc.go` with package declaration only; verify by `go build ./internal/config/` clean.

## 3. Pure core migration (domain → core)

- [ ] 3.1 Move every `.go` file from `internal/domain/` to `internal/core/` via `git mv`; verify by `ls internal/core/` listing matches `ls internal/domain/` (minus the `doc.go` from 2.1).
- [ ] 3.2 Rename `domain.X` → `core.X` for every exported symbol via `gopls rename` (see `golang-gopls` skill); verify by `go build ./...` clean and `go test ./internal/core/...` passes.
- [ ] 3.3 Update every import site across `internal/service/`, `internal/application/`, `internal/infrastructure/`, `cmd/`, `test/` from `twiggit/internal/domain` → `twiggit/internal/core`; verify by `go build ./...` clean.
- [ ] 3.4 Rename internal type aliases (`domain.Worktree` → `core.Worktree`, etc.) atomically; verify by `go test ./...` passes.
- [ ] 3.5 Replace `domain.NewValidationError` calls with `core.NewValidationError`; verify by `go test ./internal/...` passes.
- [ ] 3.6 Replace `domain.NewGitRepositoryError` etc. with `core.NewGitRepositoryError`; verify by `go test ./internal/git/... ./internal/service/...` passes.
- [ ] 3.7 Apply `os.IsNotExist` → `errors.Is(err, os.ErrNotExist)` modernization across `internal/core/`; verify by `rg 'os\.IsNotExist' internal/core/` returns no matches.
- [ ] 3.8 Apply `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` modernization across `internal/core/`; verify by `rg 'HasPrefix.*TrimPrefix' internal/core/` returns no matches.
- [ ] 3.9 Apply linear-scan → `slices.Contains`/`slices.Clone` modernization across `internal/core/`; verify by `rg 'for _, .* := range .* { if .* == .* {' internal/core/` returns no matches.
- [ ] 3.10 Delete `internal/domain/` directory; verify by `ls internal/domain/` reports no such file or directory.

## 4. Git adapter migration

- [ ] 4.1 Move `internal/infrastructure/gogit_client.go` to `internal/git/client.go`; verify by `go build ./internal/git/` clean.
- [ ] 4.2 Split read-side operations into `internal/git/reader.go` (`OpenRepository`, `ListBranches`, `BranchExists`, `GetRepositoryStatus`, `ListRemotes`, `GetCommitInfo`, `GetRepositoryInfo`, `ValidateRepository`); verify by `go test ./internal/git/...` passes.
- [ ] 4.3 Move `internal/infrastructure/cli_client.go` to `internal/git/writer.go` (worktree/branch mutations); verify by `go test ./internal/git/...` passes.
- [ ] 4.4 Move `internal/infrastructure/command_executor.go` to `internal/git/command_executor.go`; verify by `go test ./internal/git/...` passes.
- [ ] 4.5 Move `internal/infrastructure/repo_finder.go` to `internal/git/repo_finder.go`; verify by `go test ./internal/git/...` passes.
- [ ] 4.6 Update `domain.*` → `core.*` references across all `internal/git/` files; verify by `go build ./internal/git/` clean.
- [ ] 4.7 Create `internal/git/errors.go` for `core.NewGitRepositoryError` / `core.NewGitWorktreeError` / `core.NewGitCommandError` constructors; verify by `go test ./internal/git/...` passes.
- [ ] 4.8 Apply modernization sweep (`errors.Is`, `slices.Contains`, `CutPrefix`) across `internal/git/`; verify by `rg 'os\.IsNotExist' internal/git/` returns no matches.
- [ ] 4.9 Update `git-client` spec delta path references from `internal/infrastructure/` to `internal/git/`; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.
- [ ] 4.10 Update test mocks to import `internal/git/` instead of `internal/infrastructure/`; verify by `go test ./test/mocks/... ./test/integration/...` passes.

## 5. Config + context migration

- [ ] 5.1 Move `internal/infrastructure/config_manager.go` to `internal/config/manager.go`; verify by `go build ./internal/config/` clean.
- [ ] 5.2 Update `domain.*` → `core.*` references in `internal/config/`; verify by `go test ./internal/config/...` passes.
- [ ] 5.3 Add `git-config` spec delta (`git-config/spec.md`) with new requirements for koanf + XDG + env precedence; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.
- [ ] 5.4 Move `internal/infrastructure/context_resolver.go` to `internal/git/context_resolver.go`; verify by `go test ./internal/git/...` passes.
- [ ] 5.5 Move `internal/infrastructure/context_detector.go` to `internal/git/context_detector.go`; verify by `go test ./internal/git/...` passes.
- [ ] 5.6 Update `git-context-resolver` spec delta path references; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.
- [ ] 5.7 Apply modernization sweep across `internal/config/` and the migrated context files; verify by `rg 'os\.IsNotExist' internal/config/ internal/git/context_*.go` returns no matches.

## 6. Hook runner + shell wrapper migration

- [ ] 6.1 Move `internal/infrastructure/hook_runner.go` to `internal/git/hook_runner.go`; verify by `go build ./internal/git/` clean.
- [ ] 6.2 Update `domain.*` → `core.*` references and `application.HookRunner` interface satisfaction in `internal/git/hook_runner.go`; verify by `go test ./internal/git/...` passes.
- [ ] 6.3 Update `git-hook-runner` spec delta path references; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.
- [ ] 6.4 Move shell-detect logic to `internal/core/shell_detect.go` (it's pure); verify by `go test ./internal/core/...` passes.
- [ ] 6.5 Update `git-shell-detect` spec delta path references from `internal/infrastructure/` to `internal/output/` + `internal/core/`; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.

## 7. lipgloss + iostreams + output packages

- [ ] 7.1 Add `github.com/charmbracelet/lipgloss` to `go.mod` (pin version at apply time); verify by `go mod tidy` succeeds.
- [ ] 7.2 Create `internal/iostreams/iostreams.go` with `IOStreams` interface + `System()` and `Test()` constructors + TTY detection; verify by `go build ./internal/iostreams/` clean.
- [ ] 7.3 Create `internal/iostreams/styles.go` using lipgloss for header / success / error / hint styles; verify by `go test ./internal/iostreams/...` passes.
- [ ] 7.4 Create `internal/output/formatter.go` with `Formatter` interface + `json` / `table` / `plain` implementations + registry; verify by `go test ./internal/output/...` passes.
- [ ] 7.5 Create `internal/output/errors.go` with `FormatError` dispatch via `errors.As` on `core.ValidationError`, `core.NotFoundError`, `core.OperationError`, `core.UsageError`; verify by `go test ./internal/output/...` passes.
- [ ] 7.6 Create `internal/output/table.go` using lipgloss for table rendering; verify by `go test ./internal/output/...` passes.
- [ ] 7.7 Create `internal/output/prompt.go` for confirmation prompts (used by `--force` / `--yes` paths); verify by `go test ./internal/output/...` passes.

## 8. cmdutil + Factory

- [ ] 8.1 Create `internal/cmdutil/factory.go` with lazy `func() T` fields + `sync.Once`-cached config + `Init()` step; verify by `go test ./internal/cmdutil/...` passes.
- [ ] 8.2 Create `internal/cmdutil/exit.go` with `ExitCodeFor(err) int` + `ExitOK`/`ExitError`/`ExitUsage` constants; verify by `go test ./internal/cmdutil/...` passes.
- [ ] 8.3 Create `internal/cmdutil/json_flags.go` with shared `--output` / `--quiet` / `--verbose` flag helpers; verify by `go test ./internal/cmdutil/...` passes.
- [ ] 8.4 Add `cli-factory` spec delta (`cli-factory/spec.md`) documenting the lazy-init + `sync.Once` contract; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.
- [ ] 8.5 Add `cli-iostreams` spec delta (`cli-iostreams/spec.md`) documenting TTY detection + stdout/stderr discipline; verify by `openspec validate cli-functional-core-shell --specs --json` zero issues.

## 9. Delete legacy packages

- [ ] 9.1 Verify no remaining import of `twiggit/internal/application` anywhere in the repo; verify by `rg 'twiggit/internal/application' --type go` returns no matches.
- [ ] 9.2 Verify no remaining import of `twiggit/internal/service` anywhere in the repo; verify by `rg 'twiggit/internal/service' --type go` returns no matches.
- [ ] 9.3 Verify no remaining import of `twiggit/internal/infrastructure` anywhere in the repo; verify by `rg 'twiggit/internal/infrastructure' --type go` returns no matches.
- [ ] 9.4 Delete `internal/application/`, `internal/service/`, `internal/infrastructure/` directories; verify by `ls internal/application internal/service internal/infrastructure` reports no such file or directory.
- [ ] 9.5 Run full build + test suite; verify by `go build ./... && go test ./...` exits 0.

## 10. cmd/ refactor (per command)

- [ ] 10.1 Refactor `cmd/list.go` to `Options` struct + `NewCmdList(f, runF)` + `runList(opts)`; verify by `go test ./cmd/...` passes and e2e tests for `list` still pass.
- [ ] 10.2 Refactor `cmd/create.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `create` still pass.
- [ ] 10.3 Refactor `cmd/delete.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `delete` still pass.
- [ ] 10.4 Refactor `cmd/prune.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `prune` still pass.
- [ ] 10.5 Refactor `cmd/cd.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `cd` still pass.
- [ ] 10.6 Refactor `cmd/init.go` to the same pattern; verify by `go test ./cmd/...` passes and e2e tests for `init` still pass.
- [ ] 10.7 Refactor `cmd/version.go` and `cmd/_carapace.go` to the same pattern; verify by `go test ./cmd/...` passes.
- [ ] 10.8 Route all output through `iostreams.IOStreams` (no more direct `os.Stdout`/`os.Stderr`); verify by `rg 'os\.Stdout|os\.Stderr' cmd/` returns no matches outside `main.go`.
- [ ] 10.9 Route all error formatting through `output.FormatError`; verify by `rg 'cli/error' cmd/` returns no matches.
- [ ] 10.10 Add unit tests for `runList`, `runCreate`, `runDelete`, `runPrune`, `runCd`, `runInit` (testable via direct call); verify by `go test ./cmd/...` passes.

## 11. main.go rewrite

- [ ] 11.1 Replace `main.go` with composition root: `cmdutil.NewFactory` + `signal.NotifyContext` + cobra `Execute` + `cmdutil.ExitCodeFor`; verify by `go build ./...` clean.
- [ ] 11.2 Add panic-recover deferred function that prints to `ioStreams.Stderr` and exits 1; verify by injecting a panic in a test command and confirming the binary exits 1 with the expected stderr message.
- [ ] 11.3 Wire `signal.NotifyContext` cancellation into cobra via `cmd.SetContext(ctx)`; verify by sending SIGINT to the binary during a long-running command and confirming clean exit.
- [ ] 11.4 Verify exit-code propagation: `cmdutil.ExitCodeFor(err)` is the only return path from `main`; verify by `rg 'os\.Exit' main.go` returns no matches outside the panic-recover.
- [ ] 11.5 Update `cli-main-entry-point` spec to reference the new composition root (deferred to follow-up spec delta; out of scope for this change's 12 deltas); verify by `openspec validate --strict --json` zero issues.

## 12. Depguard + lint config

- [ ] 12.1 Replace `.golangci.yml` depguard block with Tier 2 rules per proposal §Lint (6 rule sets); verify by reading the resulting file.
- [ ] 12.2 Drop `gocognit` from the linters list; verify by `rg 'gocognit' .golangci.yml` returns no matches.
- [ ] 12.3 Keep `nolintlint` (`require-explanation: true, require-specific: true`) and `errcheck.check-type-assertions: true`; verify by reading the linters section.
- [ ] 12.4 Add depguard rules to forbid `os.Stdout`/`os.Stderr` in `cmd/`; verify by `rg 'os\.Stdout|os\.Stderr' cmd/` returns no matches outside `main.go`.
- [ ] 12.5 Run `golangci-lint run`; verify by zero new findings (existing findings from prior changes still allowed).
- [ ] 12.6 Pin lipgloss version in `go.mod` (Q5 deferred item); verify by `go mod tidy` and reading `go.mod`.
- [ ] 12.7 Update `.pre-commit-config.yaml` if `golangci-lint` invocation changed; verify by `pre-commit run --all-files` clean.

## 13. Modernization sweep (final pass)

- [ ] 13.1 Apply `os.IsNotExist` → `errors.Is(err, os.ErrNotExist)` across all new and migrated files; verify by `rg 'os\.IsNotExist' --type go` returns no matches.
- [ ] 13.2 Apply `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` across all new and migrated files; verify by `rg 'HasPrefix.*TrimPrefix' --type go` returns no matches.
- [ ] 13.3 Apply linear-scan → `slices.Contains` / `slices.Clone` across all new and migrated files; verify by `rg 'for _, .* := range .* { if .* == .* {' --type go` returns no matches.
- [ ] 13.4 Apply `for i := 0; i < N; i++` → `for i := range N` across all new and migrated files; verify by `rg 'for i := 0; i < .*; i\+\+' --type go` returns no matches.
- [ ] 13.5 Apply swallowed-error → `slog.Error` pattern across all new and migrated files; verify by `rg '_, _ = ' --type go` returns no matches.
- [ ] 13.6 Run `gofmt -s -w` and `goimports -w` across all touched files; verify by `gofmt -l ./...` returns no matches.
- [ ] 13.7 Run `go vet ./...`; verify by zero findings.

## 14. Verification

- [ ] 14.1 Run `openspec validate cli-functional-core-shell --json`; verify by `valid: true, issues: []` in the output.
- [ ] 14.2 Run `openspec status --change cli-functional-core-shell --json`; verify by `isPlanningComplete: true` in the output.
- [ ] 14.3 Run `openspec list --specs --json` and confirm the 8 new specs (`cli-factory`, `cli-iostreams`, `cli-output`, `cli-exit-codes`, `core-types`, `core-errors`, `core-paths`, `git-config`) are listed; verify by reading the JSON output.
- [ ] 14.4 Run `go build ./... && go test ./...`; verify by exit code 0.
- [ ] 14.5 Run e2e suite via `go test ./test/e2e/...`; verify by zero failures (e2e suite untouched by this change; e2e is the integration safety net for the cmd/ refactor).
