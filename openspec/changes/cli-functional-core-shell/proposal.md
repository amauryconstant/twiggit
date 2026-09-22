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

- Create `internal/core/` (pure functional core: types, value objects, validation, errors, rules).
- Create `internal/git/` (git I/O adapter: `client.go`, `reader.go`, `writer.go`, `errors.go`, `repo_finder.go`, `hook_runner.go`, `context_resolver.go`).
- Create `internal/output/` (formatting: `formatter.go`, `errors.go`, `table.go`, `prompt.go`).
- Create `internal/iostreams/` (TTY abstraction: `iostreams.go`, `styles.go`).
- Create `internal/cmdutil/` (composition + exit codes: `factory.go`, `exit.go`, `json_flags.go`).
- `internal/version/` stays.
- Move `internal/domain/*` → `internal/core/`. Rename `domain.X` → `core.X` via `gopls rename` (touches every `.go` file).
- Move `internal/infrastructure/git_utils.go`, `gogit_client.go`, `cli_client.go`, `command_executor.go`, `repo_finder.go`, `hook_runner.go`, `context_resolver.go` → `internal/git/` (split into `client.go`, `reader.go`, `writer.go`, `errors.go`).
- Move `internal/infrastructure/config_manager.go`, `context_detector.go` → `internal/config/` + `internal/git/` split.
- Delete `internal/application/`. Service-interface contracts move to `internal/core/` (consumer-side).
- Delete `internal/service/`. Orchestration moves to `internal/core/` (pure) + `cmd/<command>.go` (I/O).
- Delete `internal/infrastructure/` skeleton (files migrated above; helper remains if any leftover).

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

### Modified Capabilities

- `cli-factory` — new spec; documents Factory pattern, lazy init contract, sync.Once-cached config.
- `cli-iostreams` — new spec; documents IOStreams abstraction, TTY detection, stdout/stderr discipline.
- `cli-output` — new spec; documents `--output` flag, Formatter interface (json/table/plain).
- `cli-exit-codes` — new spec; documents 0/1/2 exit code contract via `cmdutil.ExitCodeFor`.
- `core-types` — new spec; documents value-object pattern (`core.NewBranchName` etc.), validation pipeline (`Pipeline[T]`).
- `core-errors` — new spec; documents `ValidationError`, `NotFoundError`, `OperationError`, `UsageError` hierarchy.
- `core-paths` — new spec; documents path utilities migrated from `domain/pathutils.go`.
- `git-client` — modify existing `infrastructure-git-client`; adapt to Tier 2 path references (`internal/git/`); preserve routing table + constructor-error requirements.
- `git-config` — new spec; documents config loading migrated from `infrastructure-config-manager`.
- `git-context-resolver` — modify existing `infrastructure-context-resolver`; adapt path references.
- `git-hook-runner` — modify existing `infrastructure-hook-runner`; adapt path references.
- `git-shell-detect` — modify existing `infrastructure-shell-detect`; adapt path references.

### New Capabilities

None. All changes are package migrations of existing capabilities or new specs that document newly-introduced abstractions.

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
| Specs | ~12 spec deltas (8 new, 4 modified) | — |
| Tests | mechanical rename + restructure per package | ~1500 |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged (commands, flags, output, exit codes preserved) |
| Public API of `internal/*` | Wholesale restructure; package names change (`domain` → `core`, `application`/`service`/`infrastructure` → `core`/`git`/`output`/`iostreams`/`cmdutil`) |
| Build | New dependency: lipgloss |
| Tests | All test files rename/move alongside source files; mechanical via `gopls rename` |
| Lint | depguard blocks reverse imports; strict Tier 2 rule for `internal/core/**` |

## Non-goals

- Interface Segregation Principle refactor — separate change; tracked as `interface-segregation`.
- `moq` generation tooling switch — deferred.
- Receiver-less methods → free functions — deferred (quality change).
- `slog` migration across untouched `cmd/` files — deferred.
- Lint threshold tightening (`funlen`, `gocyclo`) — deferred.
- Spec migration of legacy prefixes (`domain-*`, `infrastructure-*`, `application-*`) to `core-*`/`git-*`/`cli-*` — deferred; this change introduces new specs with Tier 2 prefixes only.
- Wholesale modernization of `_ = fmt.Fprint*` sites in `cmd/` — deferred.

## Out-of-scope follow-ups

- `gopls rename` for the `domain.X` → `core.X` migration: needs a discrete slice at apply time. The `gopls` skill handles this safely (renames every call site atomically; refuses to break interface satisfaction).
- Build-time lipgloss version pin: deferred to apply time per Q5.
