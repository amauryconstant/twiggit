# Design: Tier 2 Foundation

## Principles

Every interactive command is scriptable: flags bypass prompts; prompting is gated on an interactive TTY (and emits `core.UsageError` on missing required input when not interactive).

## Context

`twiggit` is a single-binary Cobra CLI with 6 commands (`list`, `create`, `delete`, `prune`, `cd`, `init`) plus 2 hidden commands (`_carapace`, `version`). It uses a five-layer DDD-Light architecture: `cmd/` → `internal/application/` (interfaces) → `internal/service/` (orchestration) → `internal/domain/` (pure) and `internal/infrastructure/` (git, config, shell, paths). The `architecture-layer-inversion` change has been planned to land this layout with a depguard-enforced convention.

The `golang-cli-architecture` skill (`tier-mapping.md`) flags this layering as over-architected for a CLI of this size and recommends a Tier 2 (3-concern) layout: **Parse → Execute → Respond**, physically expressed as `internal/core/` (pure functional core), `internal/git/` (I/O adapter), `internal/output/` (formatting), `internal/iostreams/` (TTY abstraction), `internal/cmdutil/` (composition + exit codes), with `cmd/<command>.go` as the wiring layer. Tier 2's strict rule is `internal/core/` imports only Go stdlib + `samber/lo`; everything else flows through the composition root.

This change is the consolidation that lands the skill's recommendation, collapses the mechanical depguard rules from `architecture-layer-inversion`, and adopts the Factory + IOStreams + command-pattern (`Options` + `runF`) composition style that gh, kubectl, and docker use.

## Goals / Non-Goals

**Goals:**
- Collapse the 5-layer convention into a 3-concern Tier 2 layout that matches the `golang-cli-architecture` skill recommendation.
- Move every pure type/rule from `internal/domain/` to `internal/core/`; rename `domain.X` → `core.X` atomically via `gopls rename`.
- Move every git/config/shell/hook I/O adapter from `internal/infrastructure/` to `internal/git/` (split into `client.go`, `reader.go`, `writer.go`, `errors.go`).
- Adopt `cmdutil.Factory` (lazy function fields + `sync.Once`-cached config) as the single composition root for all commands.
- Adopt `iostreams.IOStreams` so commands stop touching `os.Stdout`/`os.Stderr` directly and gain TTY-aware rendering.
- Adopt `output.Formatter` (`json`/`table`/`plain`) and `output.FormatError` (`errors.As` dispatch) so the `--output` flag and error formatting share one formatter registry.
- Adopt `cmdutil.ExitCodeFor` (0/1/2 contract) and rename `ExitCodeSuccess`/`ExitCodeError`/`ExitCodeUsage` → `ExitOK`/`ExitError`/`ExitUsage`. Keep behavior identical to `cli-error-formatting`.
- Adopt `cmd/<command>.go` pattern: `Options` struct + `NewCmd<Name>(f, runF)` + `run<Name>(opts) error`.
- Rewrite `main.go` as a thin composition root: Factory init → `signal.NotifyContext` → cobra dispatch → exit-code propagation.
- Adopt strict Tier 2 depguard in `.golangci.yml`.
- Extend the modernization sweep (errors.Is, slices.Contains, CutPrefix, slog.Error) into every new package.

**Non-Goals:**
- Interface Segregation Principle refactor (separate `interface-segregation` change).
- `moq` generation tooling switch (deferred; testify retained).
- Receiver-less methods → free functions (deferred quality change).
- `slog` migration across untouched `cmd/` files (deferred).
- Lint threshold tightening for `funlen`/`gocyclo` (deferred).
- Spec migration of legacy prefixes (`domain-*`, `infrastructure-*`, `application-*`) to Tier 2 prefixes (deferred; this change introduces new specs only).
- Wholesale modernization of `_ = fmt.Fprint*` sites in `cmd/` (deferred).

## Decisions

### 1. Package mapping (5-layer → Tier 2)

**Choice:** Five new packages, one kept, four deleted.

| Old | New | Role |
|---|---|---|
| `internal/domain/` | `internal/core/` | Pure types, value objects, errors, validation, rules |
| `internal/application/` | (deleted) | Service interfaces deferred to `interface-segregation` change; `cmd/` consumes concrete types via Factory |
| `internal/service/` | (deleted; orchestration folded into `internal/core/` pure + `cmd/<command>.go` I/O) | Orchestration becomes `core.OrchestrateX` + `cmd runX` |
| `internal/infrastructure/` | `internal/git/` + `internal/config/` + `internal/output/` + `internal/iostreams/` + `internal/cmdutil/` | I/O adapters split by concern |
| `internal/version/` | (kept) | Build-time version injection |

**Rationale:** The Tier 2 layout matches `golang-cli-architecture`'s recommendation. Pure types get a single home (`core`); I/O adapters split by the resource they touch (git, config, output formatting, TTY, composition); command wiring lives at `cmd/`. The `core` package becomes the only one a reader needs to understand to reason about the business rules. The `application/` interfaces are not reintroduced here; `cmd/` consumes concrete types via Factory lazy fields, and ISP-shaped consumer-side interfaces are introduced by the `interface-segregation` change.

**Alternatives considered:**
- Keep `domain` package name → rejected; the skill explicitly recommends `core` to signal "functional core, imperative shell."
- Keep `application`/`service` split → rejected; the ISP refactor is a separate change, but the 5-layer split itself is the over-architecture the skill flags.
- Move service interfaces to consumer packages (e.g., `internal/core/worktree_iface.go`) → deferred to `interface-segregation`.

### 2. `domain.X` → `core.X` rename via `gopls rename`

**Choice:** Use `gopls rename` (via `golang-gopls` skill) to rename the package identifier across every `.go` file in the repo atomically.

**Rationale:** `gopls rename` walks every reference including test files, mocks, and the `_test.go` files in `test/integration/` and `test/e2e/`. It refuses to break interface satisfaction, so a half-migrated symbol causes a build failure immediately, not a silent runtime regression. A `find … -exec sed -i` approach cannot guarantee interface satisfaction.

**Alternatives considered:**
- `find . -exec sed -i 's|internal/domain|internal/core|g'` → rejected; no compile-time guarantee; misses mocks generated by `mockery`.
- Manual rename with `git mv` per file → rejected; too slow; error-prone for ~50 files.
- New `core` package aliases `domain` → rejected; ships two parallel APIs forever.

### 3. `cmdutil.Factory` pattern (lazy function fields, `sync.Once`-cached config)

**Choice:** A single `Factory` struct with lazy `func() T` fields (one per dependency) and a `sync.Once` for the config that wraps the first read. Commands receive `*Factory` via the `NewCmd<Name>(f, runF)` constructor and call fields as `f.Config()` / `f.GitClient()` / etc.

**Rationale:** Lazy fields give unit-test injection points without an interface explosion (the `interface-segregation` change handles that later). `sync.Once` ensures the config is loaded exactly once even when multiple commands access it through the same Factory. This is the same pattern `gh` and `kubectl` use.

**Alternatives considered:**
- Constructor-injected `*ServiceContainer` (the current 5-layer approach) → rejected; couples every command to every dependency.
- Functional options (`NewCmd(opts ...Option)`) → rejected; command wiring is a code path, not a config knob.
- Global singletons → rejected; breaks test isolation.

### 4. `iostreams.IOStreams` abstraction

**Choice:** A `IOStreams` interface (and concrete `System()` / `Test()` constructors) exposing `Stdout`, `Stderr`, `In`, `IsStdoutTTY()`, `IsStderrTTY()`, `Styles()`, `Verbosef()`. The `Styles()` method returns a `*Styles` (lipgloss-rendered) for table headers, success, error, hint, etc.

**Rationale:** Commands currently mix `fmt.Fprintf(os.Stderr, …)` and direct TTY checks via `isatty`. The IOStreams abstraction lets unit tests assert on `bytes.Buffer` output, lets the formatter detect non-TTY contexts (CI, pipe), and centralizes the lipgloss dependency so `internal/output/` doesn't have to import it directly in every file.

**Alternatives considered:**
- Pass `*bytes.Buffer` everywhere → rejected; loses TTY semantics; duplicates the same plumbing in every command.
- Use `pterm` or `fatih/color` instead of `lipgloss` → rejected; lipgloss is the de-facto Go CLI renderer; matches the skill's recommendation.

### 5. `output.Formatter` interface for `--output` flag

**Choice:** A `Formatter` interface with `FormatJSON(w io.Writer, v any) error` and `FormatTable(w io.Writer, headers []string, rows [][]string) error` plus a registry keyed on `--output` value (`json` | `table` | `plain`). Default (no flag) falls through to `table` for list commands, `plain` for everything else.

**Rationale:** The current `cli-output-formats` spec already commits to json/table; the Formatter interface makes it enforceable as a single point of extension. The `plain` value is the new addition (no formatting, raw human-readable rendering) and lets scripts opt out of pretty-printing without re-parsing table borders.

**Alternatives considered:**
- String-based dispatch in each command (`switch output { case "json": …`) → rejected; scatters the logic; impossible to test.
- A `Renderable` interface per command (each command knows its own schema) → rejected; doubles the implementation surface; doesn't compose with the `--output` flag the user sees.

### 6. `output.FormatError` with `errors.As` dispatch

**Choice:** A single `FormatError(w io.Writer, err error, ios *iostreams.IOStreams)` function that walks the error chain via `errors.As` (in order: `*core.ValidationError`, `*core.NotFoundError`, `*core.OperationError`, `*core.UsageError`) and renders the matching branch with hint text. `core.OperationError.Suggestions []string` renders after the user-facing message. `TWIGGIT_DEBUG=1` causes the full wrapped chain to print after the user message.

**Rationale:** The current `cli-error-formatting` spec registers per-error-type formatters explicitly; `FormatError` keeps that explicit-registration rule but moves the registry to `internal/output/` where it belongs (it's a presentation concern, not a domain concern). The dispatch order is identical to `cli-error-formatting`'s. Per-resource NotFound hints attach via the `NotFoundError` branch.

**Alternatives considered:**
- Keep formatter registration in `cmd/` → rejected; cmd/ should be wiring only.
- Generic `formatAny[T](err T)` → rejected; loses the ordered dispatch contract.

### 7. `core.Error` type hierarchy

**Choice:** Four types in `internal/core/errors.go`:

```go
type ValidationError struct { Field, Value, Message string; Suggestions []string }
func (e *ValidationError) Error() string { ... }   // lowercase, no emoji, no trailing punctuation
func (e *ValidationError) Unwrap() error { return nil }

type NotFoundError   struct { Entity, Name string }
func (e *NotFoundError) Error() string { ... }
func (e *NotFoundError) Unwrap() error { return nil }
func (e *NotFoundError) Is(target error) bool { ... }  // matches one of 4 NotFound sentinels

type OperationError  struct { Op, Message string; Cause error; Suggestions []string }
func (e *OperationError) Error() string { ... }
func (e *OperationError) Unwrap() error { return e.Cause }   // chain walks to cause

type UsageError      struct { Message string }   // NO Err field per skill
func (e *UsageError) Error() string { ... }
func (e *UsageError) Unwrap() error { return nil }   // typed-nil-safe per golang-safety
```

**Rationale:** Collapses the 20-type taxonomy in `domain-typed-errors` to 4 types by category, then carries resource/field/op context as struct fields. The cmd-side formatter sees per-resource distinction via `NotFoundError.Entity`. I/O-adapter-specific constructors live in `internal/git/errors.go` as `*git.ExternalError` whose embedded core type walks to `*core.OperationError`. `UsageError` carries no `Err` field (per `golang-cli-architecture` §05-errors); explicit `Unwrap() error { return nil }` avoids the typed-nil interface trap (`golang-safety`).

**Alternatives considered:**
- Keep all 20 types → rejected; over-modeled for a CLI.
- One flat `Error` type with `Kind` enum → rejected; loses `errors.As` typed-walk contract that `cli-error-formatting` relies on.
- `UsageError.Err` for cobra-wrapped usage errors → rejected; skill says UsageError is terminal. Cobra wrap uses `core.NewUsageError("wrapped: " + cobraErr.Error())` instead.

### 8. `cmdutil.ExitCodeFor` (0/1/2 contract)

**Choice:** `func ExitCodeFor(err error) int` returning `ExitOK` (0), `ExitError` (1), `ExitUsage` (2). Dispatch order: `errors.As(err, &*core.UsageError{})` → `ExitUsage`; non-nil → `ExitError`; nil → `ExitOK`. Same logic as `GetExitCodeForError` in `cli-error-formatting`; new names; new location.

**Rationale:** The skill recommends the symbolic names; the new package owns the helper. The contract is preserved (same numbers, same dispatch order), so scripts that key on exit codes are unaffected.

**Alternatives considered:**
- Keep names `ExitCodeSuccess`/`ExitCodeError`/`ExitCodeUsage` → rejected; the skill uses `ExitOK`/`ExitError`/`ExitUsage`; convention wins.
- Move the helper into `core/` → rejected; `core/` must stay pure (no exit-code constants tied to `os.Exit`).

### 9. `cmd/<command>.go` pattern (Options + runF)

**Choice:** Each command file declares `type listOptions struct{ Output string; All bool; … }`, `func NewCmdList(f *cmdutil.Factory, runF func(opts *listOptions) error) *cobra.Command`, and `func runList(opts *listOptions) error`. `NewCmdList` binds flags; `runList` is the testable business entrypoint.

**Rationale:** This is gh's pattern. Splitting wiring (`NewCmdList`) from execution (`runList`) means tests can call `runList(opts)` directly without spawning cobra, and the cobra-specific concerns (flag parsing, shell completion, help text) stay isolated.

**Alternatives considered:**
- Keep the current `RunE` inline pattern → rejected; makes commands hard to unit test.
- Generic `runCobra[Opts any](f *Factory, opts Opts, run func(Opts) error)` → rejected; type-erases the cobra help text.

### 10. `main.go` composition root

**Choice:** ~50 lines: `func main()` calls `f := cmdutil.NewFactory(ioStreams)`, defers panic recovery that prints to `ioStreams.Stderr` and exits 1, sets up `signal.NotifyContext` for SIGINT/SIGTERM cancellation that propagates into cobra via `cmd.SetContext(ctx)`, then executes `cmd.NewRootCmd(f).Execute()`. Exit code is `cmdutil.ExitCodeFor(err)`. The root command SHALL set `SilenceErrors: true` and `SilenceUsage: true` so error formatting and exit-code mapping stay in `main.go`.

**Rationale:** The skill's "thin main" recommendation. All real logic lives in `cmd/` and `core/`; `main.go` is just wiring.

**Alternatives considered:**
- Keep `main.go` doing config loading and validation → rejected; the entry point should be a dispatcher, not a loader.
- Add a separate `internal/cli/` package for main → rejected; unnecessary indirection for a single function.

### 11. lipgloss dependency

**Choice:** Add `github.com/charmbracelet/lipgloss` to `go.mod`. Pin version deferred to apply time per Q5.

**Rationale:** lipgloss is the de-facto Go CLI style renderer; `internal/iostreams/styles.go` and `internal/output/table.go` need it. Tier 2 depguard allows it only in those two packages (`internal/output/**`, `internal/iostreams/**`), nowhere else.

**Alternatives considered:**
- Use `fatih/color` → rejected; terminal-only, no style composition.
- Use `pterm` → rejected; heavy dependency footprint; overlaps with lipgloss.
- Skip styling entirely → rejected; degrades UX of table output and error hints.

### 12. Strict Tier 2 depguard

**Choice:** Rewrite the depguard rules in `.golangci.yml` to match Tier 2 (per proposal §Lint). Drop `gocognit` (redundant with `gocyclo` + `nestif`). Keep `nolintlint` (`require-explanation: true, require-specific: true`) and `errcheck.check-type-assertions: true`.

**Rationale:** The depguard rules are the only thing that prevents regression to the 5-layer convention. Strict Tier 2 rules (`internal/core/**` = stdlib + samber/lo only) make the functional-core purity a property of the build, not a convention.

**Alternatives considered:**
- Loose rules with `// allowed:` comments → rejected; convention without enforcement has rotted in this repo before (pre-layer-inversion).
- Keep `gocognit` → rejected; redundant lint rule, false positives on otherwise-clean code.

### 13. Modernization sweep (errors.Is, slices.Contains, CutPrefix, etc.)

**Choice:** Extend the `architecture-layer-inversion` modernization sweep to every new and migrated file. Specifically:
- `os.IsNotExist(err)` → `errors.Is(err, os.ErrNotExist)`
- `strings.HasPrefix(s, p) && strings.TrimPrefix(s, p)` → `strings.CutPrefix(s, p)`
- Linear scans → `slices.Contains` / `slices.Clone`
- `for i := 0; i < N; i++` → `for i := range N`
- `_, _ = f(), g()` swallowed-error pairs → `if err := f(); err != nil { slog.Error(...) }`

**Rationale:** Each of these is a code-quality improvement, and applying them during the migration means no half-modernized files. The sweep is mechanical and verifiable by grep.

**Alternatives considered:**
- Defer the modernization sweep → rejected; leaves the codebase inconsistent (some files modern, some not) and forces a second mass-rename pass later.
- Add even more modernizations (`slices.Sort` over `sort.Slice`, etc.) → rejected; out of scope for this change; defer.

### 14. Migration order (new packages first → migrate code → delete legacy → main.go)

**Choice:** Phased slice order (see Migration Plan). Critical invariant: at every step the build stays green and depguard rules are satisfied.

**Rationale:** The `gopls rename` skill handles the `domain.X` → `core.X` migration atomically, but only after the new `core` package exists. The phased order (create skeletons → migrate code → switch imports → delete legacy → main.go) lets CI catch breakage at each commit.

**Alternatives considered:**
- One-shot rewrite → rejected; impossible to bisect; high blast radius on the rename.
- Bottom-up (start with `core`) → chosen; matches the skill's "pure types first, adapters second" guidance.

### 15. Test migration strategy

**Choice:** Test files move with their source files (mechanical `git mv`). Mocks regenerated via existing mockery patterns. New unit tests for new packages (Factory, IOStreams, Formatter, FormatError) added after implementation per project convention (`tests-written-after-implementation`).

**Rationale:** Mechanical migration keeps test coverage continuous; new unit tests anchor behavior for the new abstractions (Factory, IOStreams, Formatter).

**Alternatives considered:**
- Rewrite all tests → rejected; no behavior change, so existing tests still apply.
- Defer new unit tests → rejected; breaks the project's "tests written after implementation" rule for the new abstractions.

### 16. HookRunner interface placement (consumer-side, shared)

**Choice:** The `HookRunner` interface is declared in `internal/cmdutil/hook_runner_iface.go` (consumer-side, shared). The implementation in `internal/git/hook_runner.go` satisfies it via `var _ cmdutil.HookRunner = (*git.HookRunnerImpl)(nil)`. Per `golang-cli-architecture` §"Architecture": "Interfaces where consumed — define an interface next to the code that calls it." `cmdutil` is the shared consumer that every `cmd/<command>.go` reaches into; the interface lives there rather than per-command.

**Rationale:** Skill rule says interface next to consumer; in this CLI there is no `application/` package to host it, and per-command interface duplication would explode without the `interface-segregation` change. `cmdutil` is the shared consumer seam (Factory + ExitCodeFor + HookRunner); each cmd/<command>.go consumes `cmdutil.HookRunner` and the implementation lives in the I/O adapter.

**Alternatives considered:**
- Interface next to implementation (next to `HookRunnerImpl` in `internal/git/hook_runner.go`) → rejected; violates "interfaces where consumed".
- Per-command interface in each `cmd/<command>.go` file → rejected; duplicates the same interface in 6+ files; deferred to `interface-segregation` for proper ISP decomposition.
- Move interface to a dedicated `internal/ports/` package → rejected; introduces a new package for one type.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Package-rename blast radius (~50 `.go` files, ~1500 test lines) | `gopls rename` walks every reference atomically; refuses to break interface satisfaction; build is the verification gate at each slice. |
| lipgloss is a new build-time dependency (security surface, supply chain) | Pin to a specific version at apply time per Q5; depguard restricts it to two packages; review the version's release notes before pinning. |
| Strict Tier 2 depguard could break the build mid-migration | Phased migration; depguard rules enabled only after every new package compiles; the layer-inversion depguard stays active until the strict rules are ready. |
| `core-errors` collapses 20 domain-error types to 4 `core.Error` subtypes; existing tests check specific types | Migrate tests alongside source; if a test asserts on a specific `domain.XError` type, replace with the corresponding `core.YError` assertion in the same slice. |
| 11 MODIFIED-delta spec files (existing capabilities) introduce a wider review surface for the PR | Per-spec commits; CI gate at each slice; the `openspec validate --strict --json` gate runs at the end of the planning phase before `/osc-apply-change`. |
| IOStreams adds an indirection layer; tests that bypass it (use `fmt.Println` directly) won't see TTY behavior | Lint rule: forbid `os.Stdout`/`os.Stderr` in `cmd/`; `errcheck` + `nolintlint` keeps the rule tight. |
| Command-pattern (`runF`) change touches every command file | Slice-by-slice migration per command; e2e tests catch behavior drift; CLI surface is identical (verified by `cli-create`, `cli-prune`, etc. existing specs). |
| `cli-error-formatting` spec contract changes (constants renamed, location moved) | Update `cli-error-formatting` in the same slice as the helper move; exit-code contract (0/1/2) is preserved so scripts are unaffected. |
| Factory's lazy fields could mask initialization bugs (silent NPE if a field is called before its initializer is set) | Add a `Factory.Init()` step that calls every field once at startup in `main.go`; tests cover the "lazy field is nil before first call" contract. |

## Migration Plan

Phased slice order (~14 slices). Each slice ends with the build green and depguard satisfied.

1. **Setup** — baseline: confirm `architecture-layer-inversion` is in place (or note as a precondition); verify openspec state.
2. **Package skeletons** — create empty `internal/core/`, `internal/git/`, `internal/output/`, `internal/iostreams/`, `internal/cmdutil/`, `internal/config/` with `package …` and a doc-only `doc.go`.
3. **Core migration (domain → core)** — `git mv internal/domain/* internal/core/`; mechanical rename of types; `gopls rename domain.X → core.X` for every exported symbol; update all import sites; delete `internal/domain/`.
4. **Git adapter migration (infrastructure/git_*)** — move `gogit_client.go`, `cli_client.go`, `command_executor.go`, `repo_finder.go` to `internal/git/`; split `gogit_client.go` into `client.go` (constructor) + `reader.go` (read ops); split `cli_client.go` into `writer.go` (write ops) + `errors.go`.
5. **Context resolver migration** — move `context_resolver.go` + `context_detector.go` to `internal/git/context_resolver.go`; update types from `domain.*` to `core.*`.
6. **Hook runner migration** — move `hook_runner.go` to `internal/git/hook_runner.go`.
7. **Config migration** — move `config_manager.go` to `internal/config/manager.go` (koanf wiring + XDG resolution + env expansion); split `git_config.go` (the `domain.GitConfig` block) into `internal/git/config.go` if any.
8. **Shell detect migration** — move `shell_detect.go` to `internal/core/shell_detect.go` (it's pure logic).
9. **Lipgloss + iostreams + output** — add `github.com/charmbracelet/lipgloss`; create `internal/iostreams/iostreams.go` + `styles.go`; create `internal/output/formatter.go` + `errors.go` + `table.go` + `prompt.go`.
10. **cmdutil + Factory** — create `internal/cmdutil/factory.go` (lazy fields + `sync.Once`) + `exit.go` (`ExitCodeFor`) + `json_flags.go`.
11. **Delete legacy packages** — delete `internal/application/`, `internal/service/`, `internal/infrastructure/` after every consumer has been migrated; verify build + tests.
12. **cmd/ refactor (per command)** — for each of `list.go`, `create.go`, `delete.go`, `prune.go`, `cd.go`, `init.go`, `version.go`: introduce `Options` struct + `NewCmd<Name>(f, runF)` + `run<Name>(opts)`; rewire all error formatting through `output.FormatError`; all output through `output.Formatter` and `iostreams.IOStreams`.
13. **main.go rewrite** — replace `main.go` with the composition root (Factory init → signal context → cobra dispatch → exit code).
14. **Depguard + lint** — replace layer-inversion depguard rules with strict Tier 2 rules; drop `gocognit`; run `golangci-lint run`; resolve any new findings via `//nolint:…` with `// allowed: …` explanation.

Rollback strategy: this is a single PR; rollback is `git revert`. The depguard rule change is the highest-risk item; if the strict rules fail, the slice is to revert just `.golangci.yml` and re-land the depguard as a follow-up.

## Open Questions

None. Locked decisions (resolved during the planning phase):

- **Q1** — 20-type collapse to 4 core types approved. Drives `core-errors` shape and `domain-typed-errors` MODIFIED delta.
- **Q2** — `Formatter` single-method interface (`Write(w io.Writer, data any) error`) approved. Drives `cli-output` shape and `cli-output-formats` MODIFIED delta.
- **Q3** — `HookRunner` interface lives in `internal/cmdutil/hook_runner_iface.go` (consumer-side, shared) approved. Drives `git-hook-runner` and `infrastructure-hook-runner` MODIFIED deltas.
- **Q4** — I/O error wrappers live in `internal/git/errors.go` as `git.ExternalError{Tool, Operation, Message, Cause, Kind}` per `golang-cli-architecture` §05-errors. Constructors `git.NewRepoError`, `git.NewWorktreeError`, `git.NewCommandError` return `*git.ExternalError` whose embedded core type walks to `*core.OperationError`.
- **Q5** — lipgloss version pin deferred to apply time (task 12.6).

The `golang-cli-architecture` skill's Tier 2 recommendation fully determines the approach.
