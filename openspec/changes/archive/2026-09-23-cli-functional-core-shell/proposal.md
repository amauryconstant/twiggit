# Proposal: Tier 2 Foundation

## Why

`twiggit` ships a single-binary CLI with 6 commands. The current five-layer
DDD-Light architecture (`cmd` → `application` → `service` → `domain`; `infrastructure`
→ `application` → `domain`) was inherited before the project's current size and
shape were clear. The `golang-cli-architecture` skill explicitly identifies this as
over-architected for a CLI of this size, recommending a Tier 2 (3-concern) layout:
Parse / Execute / Respond, with `internal/core/` (functional core), `internal/git/`
(I/O adapter), `internal/output/` (formatting), `internal/iostreams/` (TTY abstraction),
and `internal/cmdutil/` (composition + exit codes).

The `architecture-layer-inversion` change lands the existing five-layer convention
mechanically (depguard-enforced). This change collapses that convention into the
Tier 2 layout the skill recommends, adopting the Factory pattern, IOStreams
abstraction, command-pattern (Options + runF), and strict functional-core purity.

## What Changes

### Package migration

- **BREAKING** Create `internal/core/` (pure functional core: types, value objects, validation, errors, rules).
- **BREAKING** Create `internal/git/` (git I/O adapter: `client.go`, `reader.go`, `writer.go`, `errors.go`, `repo_finder.go`, `hook_runner.go`, `context_resolver.go`).
- **BREAKING** Create `internal/output/` (formatting: `formatter.go`, `errors.go`, `table.go`).
- **BREAKING** Create `internal/iostreams/` (TTY abstraction: `iostreams.go`, `styles.go`).
- **BREAKING** Create `internal/cmdutil/` (composition + exit codes: `factory.go`, `exit.go`, `json_flags.go`).
- `internal/version/` stays.
- **BREAKING** Move `internal/domain/*` → `internal/core/`. Rename `domain.X` → `core.X` via `gopls rename` (touches every `.go` file).
- **BREAKING** Move `internal/infrastructure/git_utils.go`, `gogit_client.go`, `cli_client.go`, `command_executor.go`, `repo_finder.go`, `hook_runner.go`, `context_resolver.go` → `internal/git/` (split into `client.go`, `reader.go`, `writer.go`, `errors.go`).
- **BREAKING** Move `internal/infrastructure/config_manager.go`, `context_detector.go` → `internal/config/` + `internal/git/` split.
- **BREAKING** Delete `internal/application/`. Service-interface contracts deferred to the `interface-segregation` change; `cmd/` consumes concrete types via Factory.
- **BREAKING** Delete `internal/service/`. Orchestration moves to `internal/core/` (pure) + `cmd/<command>.go` (I/O).
- **BREAKING** Delete `internal/infrastructure/` skeleton (files migrated above; helper remains if any leftover).

### Composition pattern

- Adopt `cmdutil.Factory` with lazy function fields and `sync.Once`-cached config.
- Adopt `iostreams.IOStreams` (`System()`, `Test()`, `IsStdoutTTY()`, `Styles()`, `Verbosef()`).
- Adopt `output.Formatter` (`json`, `table`, `plain`) for `--output` flag.
- Adopt `output.FormatError` with `errors.As` dispatch on `core.ValidationError`, `core.NotFoundError`, `core.OperationError`, `core.UsageError`.
- Adopt `core.Error` type hierarchy (`ValidationError`, `NotFoundError`, `OperationError`, `UsageError`).
- Adopt `cmdutil.ExitCodeFor(err)` → `ExitOK`/`ExitError`/`ExitUsage` (0/1/2 contract; matches `cli-error-formatting`).
- Adopt `cmd/<command>.go` pattern: `Options` struct + `NewCmd<Name>(f, runF)` + `run<Name>(opts) error`.
- Rewrite `main.go` as composition root: Factory init + `signal.NotifyContext` + panic recover.

### Dependencies

- Add `github.com/charmbracelet/lipgloss` (style rendering for `internal/iostreams/styles.go` and `internal/output/table.go`). Pin version deferred to apply time.

### Lint

- Adopt strict Tier 2 depguard (replaces the loose 5-layer rules from the layer-inversion change):
  - `internal/core/**`: allow `$gostd`, `github.com/samber/lo`. Deny `github.com/*`.
  - `internal/git/**`: allow `$gostd`, `github.com/samber/lo`, `internal/core`. Deny `github.com/*`.
  - `internal/output/**`: allow `$gostd`, `github.com/samber/lo`, `github.com/charmbracelet/lipgloss`, `internal/core`.
  - `internal/iostreams/**`: allow `$gostd`, `github.com/charmbracelet/lipgloss`, `github.com/samber/lo`.
  - `internal/cmdutil/**`: allow `$gostd`, `github.com/samber/lo`, `internal/core`, `internal/iostreams`. Deny `github.com/*`.
  - `cmd/**`: allow `$gostd`, `github.com/samber/lo`, `internal/core`, `internal/git`, `internal/config`, `internal/output`, `internal/iostreams`, `internal/cmdutil`, `internal/version`, `twiggit/cmd`.
- Keep `nolintlint` (`require-explanation: true, require-specific: true`) and `errcheck.check-type-assertions: true` from layer-inversion.
- Drop `gocognit` (redundant with `gocyclo` + `nestif`).

### Modernization sweep

- `os.IsNotExist` → `errors.Is(err, os.ErrNotExist)` across all touched files (extends layer-inversion sweep).
- `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` (extends layer-inversion sweep).
- Linear scans → `slices.Contains` / `slices.Clone` (extends layer-inversion sweep).
- `for i := 0; i < N; i++` → `for i := range N` (extends layer-inversion sweep).
- Swallowed errors → `slog.Error` and continue (extends layer-inversion sweep).

## Capabilities

### New Capabilities

- `cli-factory` — documents Factory pattern, lazy init contract, `sync.OnceValue`-cached config.
- `cli-iostreams` — documents IOStreams struct, TTY detection, stdout/stderr discipline, debug-vs-verbose split.
- `cli-output` — documents `--output` flag, Formatter interface (json/table/plain).
- `cli-exit-codes` — documents 0/1/2 exit code contract via `cmdutil.ExitCodeFor`.
- `core-types` — documents value-object pattern (`core.NewBranchName` etc.), validation pipeline (`Pipeline[T]`).
- `core-errors` — documents `ValidationError`, `NotFoundError`, `OperationError`, `UsageError` hierarchy; I/O-specific constructors live in their adapter packages.
- `core-paths` — documents path utilities in their new home at `internal/core/`.
- `git-config` — documents config loading migrated from `infrastructure-config-manager`.
- `git-client` — documents the composite `git.GitClient` for Tier 2; legacy `infrastructure-git-client` spec retains the two-role-interface design per deferred-migration non-goal.
- `git-context-resolver` — documents context-detection with `core.NormalizePath` / `core.IsPathUnder`; legacy `infrastructure-context-resolver` spec left intact per deferred-migration non-goal.
- `git-hook-runner` — documents `internal/git/hook_runner.go`; legacy `infrastructure-hook-runner` spec left intact per deferred-migration non-goal.
- `git-shell-detect` — documents the core/output split for shell logic; legacy `infrastructure-shell-detect` spec left intact per deferred-migration non-goal.

### Modified Capabilities

- `cli-error-formatting` — exit-code constants renamed `ExitCodeSuccess`/`ExitCodeError`/`ExitCodeUsage` → `ExitOK`/`ExitError`/`ExitUsage`; helper `GetExitCodeForError` → `cmdutil.ExitCodeFor`; per-resource formatter registry collapsed to `FormatError` dispatch on 4 core types; SIGINT → exit 130 / SIGTERM → exit 143 added.
- `cli-output-formats` — `--output` enum extended with `table`, `plain`, `jsonl`; `Formatter` interface contract added; default empty `--output` = plain.
- `cli-verbose-output` — boolean `Verbose` replaces the `-v`/`-vv` level distinction; `ios.Verbosef(format, ...)` replaces `logv(cmd, level, ...)`; separate `Logger *slog.Logger` channel added for `TWIGGIT_DEBUG`.
- `cli-main-entry-point` — main becomes thin composition root; config loading moves to `cmdutil.Factory.Config` lazy field; `GetExitCodeForError` → `cmdutil.ExitCodeFor`; old exit-code constants removed.
- `domain-typed-errors` — 20-type taxonomy collapsed to 4 core types (`ValidationError`, `NotFoundError`, `OperationError`, `UsageError`); 13 sentinels slimmed to 4 NotFound sentinels; `domain.NewValidationError(request, ...)` → `core.NewValidationError(field, value, message)`; shell/usage-specific sentinels removed.
- `domain-context-types` — `domain.ContextType` / `PathType` / `ResolutionResult` / `SuggestionOption` → `core.*` rename.
- `infrastructure-context-resolver` — detection surface scoped to priority chain + path-utils dep; `domain.ContextDetectionError` → `core.OperationError`; full resolver/suggestion coverage retained.
- `infrastructure-git-client` — "No composite umbrella interface SHALL exist" rule removed; composite `git.GitClient` is the only injection point; role interfaces become internal collaborators; `git.NewGoGitClient()` → `git.NewClient()`.
- `infrastructure-shell-detect` — `domain.ShellType` → `core.ShellType`; `ShellInfrastructure.ComposeWrapper` → `output.ComposeWrapper` in `internal/output/wrapper.go`; filesystem probing moved to `internal/git/shell_detect.go`.
- `infrastructure-hook-runner` — `application.HookRunner` interface → `cmdutil.HookRunner` declared in `internal/cmdutil/hook_runner_iface.go`; `domain.HookResult` → `core.HookResult`; `application.HookRunRequest` → `core.HookRunRequest`.
- `infrastructure-config-manager` — `domain.DefaultConfig()` → `core.DefaultConfig()`; `domain.ConfigError` (exit 3) → `core.OperationError` (exit 1); `NO_COLOR` env handling added.

## Impact

| Layer | Files | Lines (est.) |
|---|---|---|
| `internal/core/` | new package (~15 files: types, errors, validation, rules, paths, shell_wrapper) | ~1500 |
| `internal/git/` | new package (~8 files: client, reader, writer, errors, repo_finder, hook_runner, context_resolver) | ~1200 |
| `internal/output/` | new package (~4 files: formatter, errors, table, prompt) | ~300 |
| `internal/iostreams/` | new package (~2 files: iostreams, styles) | ~200 |
| `internal/cmdutil/` | new package (~3 files: factory, exit, json_flags) | ~250 |
| `cmd/` | refactor every command to Factory + Options + runF pattern (~10 files) | ~600 |
| `main.go` | composition root rewrite | ~50 |
| `internal/application/` | deleted | -250 |
| `internal/service/` | deleted | -700 |
| `internal/infrastructure/` | deleted (after migration) | -1100 |
| `internal/domain/` | deleted (after migration) | -600 |
| `.golangci.yml` | depguard rewrite (Tier 2) | ~80 |
| `go.mod` | add `github.com/charmbracelet/lipgloss` | ~1 |
| Specs | ~23 spec deltas (12 new, 11 modified) | — |
| Tests | mechanical rename + restructure per package | ~1500 |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged (commands, flags, output, exit codes preserved) |
| Public API of `internal/*` | Wholesale restructure; package names change (`domain` → `core`, `application`/`service`/`infrastructure` → `core`/`git`/`output`/`iostreams`/`cmdutil`) |
| Build | New dependency: lipgloss |
| Tests | All test files rename/move alongside source files; mechanical via `gopls rename` |
| Lint | depguard blocks reverse imports; strict Tier 2 rule for `internal/core/**` |

## Non-goals

- Interface Segregation Principle refactor — separate change; tracked as `interface-segregation`. The composite `git.GitClient` stays concrete in this change; the `interface-segregation` change promotes role interfaces to consumer-side interfaces.
- `moq` generation tooling switch — deferred.
- Receiver-less methods → free functions — deferred (quality change).
- `slog` migration across untouched `cmd/` files — deferred.
- Lint threshold tightening (`funlen`, `gocyclo`) — deferred.
- Wholesale modernization of `_ = fmt.Fprint*` sites in `cmd/` — deferred.
- Wholesale rename of `infrastructure-*` and `application-*` source directories to `git-*`/`cmdutil-*` prefixes — the legacy source dirs are deleted in this change, but the legacy spec paths (`openspec/specs/infrastructure-*`, `openspec/specs/application-*`) keep their directory names so existing cross-refs stay valid. The 11 MODIFIED deltas in this change reconcile the requirement text; the directory names update is a separate follow-up.

## Out-of-scope follow-ups

- `gopls rename` for the `domain.X` → `core.X` migration: needs a discrete slice at apply time. The `gopls` skill handles this safely (renames every call site atomically; refuses to break interface satisfaction).
- Build-time lipgloss version pin: deferred to apply time per Q5.
