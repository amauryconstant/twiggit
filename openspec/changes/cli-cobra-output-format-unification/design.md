# Design

## Context

Two parallel output subsystems exist. `internal/output/` (the canonical home) provides generic `JSONFormatter`, `TableFormatter`, `PlainFormatter` types plus `errors.go:FormatError` which is the production error formatter wired into `main.go`. `cmd/output.go` and `cmd/error_formatter.go` (~268 lines) shadow this with bespoke types and a strategy-pattern error formatter that no production code path reaches; only their own test files exercise them.

Three spec-mandated behaviors live only in the dead code: (1) the per-sentinel hint table in `cli-error-formatting/spec.md:79-91`, (2) the "Specific matcher wins" registration order at `cli-error-formatting/spec.md:64-78` (the implementation has the reverse order in `internal/output/errors.go:34-53`), (3) the "Quiet mode strips hints" scenario at `cli-error-formatting/spec.md:107-111`. Consolidating onto `internal/output` and absorbing the dead-code behavior is one mechanical pass.

A separate output-vocabulary drift exists between `cli-output/spec.md` (mandates `json|jsonl|table|plain`, rejects `text`) and `cmd/list.go:62-74` (accepts `text|json`, rejects everything else as a plain `fmt.Errorf` exit 1). The bespoke JSON path in `cmd/output.go:43-65` wraps worktrees in `{"worktrees":[…]}` — no spec ever required the envelope.

`internal/output.PlainFormatter` (formatter.go:91-112) emits `fmt.Sprintln` per slice element, which is not the canonical CLI shape from `golang-cli/references/output.md:49` ("plain = headerless TSV, one record per line"). Migrating to `Tabular` projection + TSV aligns the implementation with the skill's table and makes list-style output scriptable via `cut`/`awk`.

The module targets Go 1.27.1 (`go.mod`); `errors.AsType[T]` (Go 1.26+) is available without prerequisite changes. The project is pre-1.0 and single-user; free to break anything.

## Goals / Non-Goals

**Goals:**
- Make `internal/output/errors.go` honor every requirement in `cli-error-formatting`.
- Make `cmd/list.go` honor every requirement in the modified `cli-output` capability, with output produced via the canonical `internal/output.NewFormatter` and the `Tabular` projection.
- Introduce the `Tabular` interface so both `TableFormatter` and `PlainFormatter` share the same input shape; `PlainFormatter` emits headerless TSV.
- Remove dead production code (`cmd/error_formatter.go`, `cmd/output.go`, `ProgressReporter.ReportProgress`, `CreateOptions.HookRunner`, redundant `ValidArgsFunction`) and the test code that targets it.
- Add `cmd.AddGroup` labels so `--help` output is scannable for the seven live subcommands.

**Non-Goals:**
- Switching all `opts.IO.Out` writes to `cmd.OutOrStdout()` — would regress the `iostreams.Test()` injection seam that 30+ test files depend on; `cmd/AGENTS.md` codifies the divergence.
- Renaming the `cmd/util.go` file or migrating to the testkit package layout.
- Reordering the `cli-error-formatting` dispatch table to put `OperationError` first — the spec mandates the current order, the implementation was wrong; no spec change.
- Bumping the module path from `twiggit` to a URL-shaped identifier.
- Bulk migration of `cli-`/`application-`/`domain-`/`infrastructure-` spec prefixes to the new `cli-`/`core-`/`git-`/`testing-` set; the current spec catalog carries both prefixes coexisting.
- Adopting `Tabular` for every list-style command. Only `list` adopts the projection here; `prune`/`cd`/etc. opt in when their output formatting is otherwise revisited.
- Keeping `jsonl` in the vocabulary. Streaming consumers pipe `jq -c .` over `json` instead.
- Adding a hidden `docgen` command or `mise run docgen` task. No CI gate runs it; no committed docs would result. Dropped entirely.

## Decisions

### Decision 1: Delete `cmd/error_formatter.go` and absorb behavior into `internal/output/errors.go`

**Choice**: Delete `cmd/error_formatter.go` (~180 lines), delete `cmd/error_formatter_test.go` (~260 lines), move `hintFor` into `internal/output/errors.go`, add quiet-mode gating, reorder dispatch, switch to `errors.AsType[T]`. Port the four unique behavioral test scenarios to `internal/output/errors_test.go`. Extract `shouldEmitHints(ios)` and `writeFieldContext(...)` helpers to single-source the quiet gate and the field-context rendering shared by `formatValidationError` and `formatOperationError`.

**Rationale**: The strategy-pattern types in `cmd/error_formatter.go` are unreachable from production (`main.go` calls `output.FormatError`). Three live spec behaviors live only in the dead code: (1) the hint table in `cli-error-formatting/spec.md:79-91`, (2) the "Specific matcher wins" registration order at `cli-error-formatting/spec.md:64-78` (the implementation has the reverse order in `internal/output/errors.go:34-53`), (3) the "Quiet mode strips hints" scenario at `cli-error-formatting/spec.md:107-111`. The fourth absorbed behavior is the `errors.AsType[T]` adoption (Go 1.26+) which drops the `asType[T error]` helper in `cmd/error_formatter.go:13-19`. Deletion removes ~440 lines of orphan code; the four ported test scenarios (~80 lines) cover the same behavior with less indirection.

**Alternatives considered**:
- *Keep both formatters, route production through `cmd`*: would require changing `main.go` to call `cmd.ErrorFormatter` instead of `output.FormatError`. Violates the canonical-home rule (errors formatted beside their origin) and creates two type families to maintain.
- *Use the dead code's strategy pattern via dependency injection*: would add a `Formatter` interface field to `main.go` and select one at boot. Pure ceremony for a single live implementation.

### Decision 2: Delete `cmd/output.go` and migrate `cmd/list.go` to `internal/output.NewFormatter`

**Choice**: Delete `cmd/output.go` (~88 lines). Rewrite `cmd/list.go:130-138` to call `internal/output.NewFormatter(format)` and pass the `Tabular` adapter directly to `formatter.Write(w, worktreeRows)`. Keep the bespoke plain-text rendering inline in `cmd/list.go`'s `displayWorktrees` for the empty-`--output` human default (Decision 5).

**Rationale**: `internal/output/formatter.go` already implements all three spec-mandated formatters (`json|table|plain`). The `cmd/output.go` `JSONFormatter` adds a top-level `{"worktrees":[…]}` envelope that no spec requires and that breaks `jq '.[]'` consumption; `internal/output.JSONFormatter` emits a bare array, which matches the JSON-shape table in `golang-cli/references/output.md:54-65`. The `WorktreeJSON{Branch, Path, Status}` projection is preserved: `cmd/list.go` builds a `worktreeRows` `Tabular` adapter that produces the same three fields. Bespoke text rendering for the empty-`--output` case stays beside `cmd/list.go`.

**Alternatives considered**:
- *Move bespoke rendering to `internal/output` as a `WorktreeFormatter`*: would require defining a `String() string` method on `*core.WorktreeInfo`, changing global representation (ripples into logs, debug output, test failure dumps). Too invasive for a one-command use case.
- *Keep `cmd/output.go` and only migrate the JSON path*: leaves `TextFormatter` orphaned and defeats the consolidation goal.

### Decision 3: Switch error dispatch to `errors.AsType[T]` (Go 1.26+)

**Choice**: At each dispatch arm in `internal/output/errors.go:FormatError`, call `errors.AsType[*core.X](err)`. Drop the `asType[T error]` helper from `cmd/error_formatter.go:13-19` without porting it.

**Rationale**: Go 1.26+ adds `errors.AsType[T](err)` which returns the typed error directly (no double-allocation, no nil guard needed). The module targets Go 1.27.1 (`go.mod`), so the primitive is available without prerequisite changes. Porting the `asType` helper would keep a Go 1.25 workaround in place for no benefit.

**Alternatives considered**:
- *Keep `asType[T error]` helper in `internal/output`*: works but adds an indirection layer that the stdlib already provides natively.

### Decision 4: Replace the carapace Find+Remove dance with `cmd.CompletionOptions.DisableDefaultCmd = true`

**Choice**: In `cmd/root.go`, add `cmd.CompletionOptions.DisableDefaultCmd = true` before `carapace.Gen(cmd)`. Delete the `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand(completionCmd)` block. Keep `newCompletionCommand` and `carapace.Gen(cmd)` unchanged.

**Rationale**: The dance exists because cobra auto-registers a `completion` subcommand, carapace auto-registers its own `completion` subcommand, and the project's custom `newCompletionCommand` is meant to replace both. `DisableDefaultCmd = true` is cobra's declarative way to skip the cobra-owned one; carapace's `_carapace` (hidden) remains the runtime path; `newCompletionCommand` remains the user-visible one. The `golang-cli/references/completion.md` reference documents this pattern explicitly.

**Alternatives considered**:
- *Walk to find carapace's `completion` and remove it, leaving cobra's*: cobra's auto-completion lacks the per-shell help text the custom command provides (see `getShellInstructions` at `cmd/completion.go:57-91`). Worse UX.
- *Replace the dance with an `addPreRun` hook*: doesn't solve the issue; the conflict is at command-registration time, not hook-chain time.

### Decision 5: Bespoke plain-text rendering stays inline in `cmd/list.go` for the empty-`--output` human default; `--output plain` goes through `PlainFormatter` TSV

**Choice**: The "branch -> path (modified)(detached)" output stays in `cmd/list.go:displayWorktrees` and fires only when `--output` is empty. Concretely: `runList` calls `output.NewFormatter(format)`; if the returned formatter is `nil` (empty format), `cmd/list.go` writes the bespoke rendering; otherwise it calls `formatter.Write(opts.IO.Out, worktreeRows)`. `--output plain` flows through `output.NewFormatter("plain")` → `PlainFormatter` → headerless TSV from the `worktreeRows` `Tabular` adapter.

**Rationale**: List-style commands keep a per-command interactive default when `--output` is not supplied; `--output plain` is the scriptable override. The two paths serve different audiences (interactive vs pipeline) and warrant different rendering strategies. `golang-cli/references/commands.md`: "Output beside the logic — the command knows what it produced and formats it with shared helpers from `internal/output`". The bespoke text shape is specific to one command's interactive default; promoting it to `internal/output` would force a global `Stringer` on `core.WorktreeInfo` that affects logging and debug dumps everywhere. The per-command default shape is captured in `cli-list` (deferred to `/osc-continue`) so future changes to `list`'s interactive rendering require an explicit spec update.

**Alternatives considered**:
- *Implement `String() string` on `*core.WorktreeInfo`, use `internal/output.PlainFormatter` for everything*: would work, but the global `Stringer` affects every site that formats a `WorktreeInfo` value (slog messages, debug `fmt.Sprintf`, test failure dumps). Out-of-scope side effects.

### Decision 6: `NewFormatter` returns `(Formatter, error)`; empty returns `(nil, nil)`

**Choice**: Change signature to `NewFormatter(format string) (Formatter, error)`. Unknown values (including the legacy `text`) return `(nil, *core.UsageError)`. Empty returns `(nil, nil)`; each command picks its own human default (Decision 5). The three known values (`json|table|plain`) return their respective formatters with a `nil` error.

**Rationale**: `golang-cli/references/output.md:43`: "No flag means the human default: styled, TTY-aware output written by the command itself." Pushing the empty-format decision to the caller is the skill-aligned pattern and removes a contradiction the prior design had (the prior design returned `(PlainFormatter{}, nil)` for empty, then `cmd/list.go` ignored the formatter and used bespoke rendering). The constructor now declares its own contract: unknown is usage-error, empty is defer-to-caller, known is the formatter.

**Alternatives considered**:
- *Keep nil-for-unknown, fix only `cmd/list.go`*: smaller diff, but the next caller will re-introduce the same bug (plain `fmt.Errorf` instead of `UsageError`). The signature is the right place to enforce.
- *Empty returns `(PlainFormatter{}, nil)`: every caller has to remember to ignore it for bespoke defaults. The `(nil, nil)` empty return makes the choice explicit at the call site.

### Decision 7: `twiggit list -o json` envelope dropped

**Choice**: JSON output for a collection emits a bare array `[{...}]`. The `WorktreeListJSON{Worktrees: []WorktreeJSON{...}}` envelope is removed.

**Rationale**: `golang-cli/references/output.md` JSON shape table: "A small collection — A JSON array". No spec ever mandated the envelope; the envelope was an implementation choice in `cmd/output.go:74-77`. Scripts consuming the old shape must update from `jq '.worktrees[]'` to `jq '.[]'`. The new shape is pinned by the `cli-list` MODIFIED requirement 5 (JSON array) and the `cli-output` ADDED requirement (bare-collection JSON shape).

**Alternatives considered**:
- *Keep envelope, add `--legacy-output` flag for back-compat*: pre-1.0 project; no external scripts in the wild depend on the envelope. Carrying a back-compat flag for a one-time shape change is dead weight.
- *Two-version deprecation cycle (warn → break)*: would require two OpenSpec changes and a release-cycle commitment. Project is pre-1.0; single cycle is appropriate.

### Decision 8: `twiggit list -o text` rejected as `core.UsageError`

**Choice**: The `text` value is rejected with `core.UsageError` (exit 2). No alias to `plain`.

**Rationale**: `cli-output/spec.md` specifies the accepted values; `text` is not among them. The implementation accepts it as a legacy quirk. The clean break is to reject per spec; scripts using `--output text` must switch to `--output plain`. Project is pre-1.0; no deprecation cycle needed.

**Alternatives considered**:
- *Accept `text` as an alias for `plain`*: preserves existing scripts but entrenches a value that the spec does not recognize. The change should make the implementation match the spec, not paper over the drift.

### Decision 9: `cmd.AddGroup` for subcommand discoverability

**Choice**: Register four groups via `cmd.AddGroup(...)` in `cmd/root.go` BEFORE the `AddCommand` block: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion). Per `golang-spf13-cobra` skill: groups must be registered before the commands that reference them.

**Rationale**: With seven live subcommands the help wall is hard to scan. Group labels make `--help` output match the user's mental model. Cosmetic; no behavior change beyond the help text. The grouping is captured in the new `cli-command-groups` capability (deferred to `/osc-continue`).

**Alternatives considered**:
- *No groups*: simpler, but `--help` output degrades as more commands land.

### Decision 10: Introduce `Tabular` interface in `internal/output/formatter.go`

**Choice**: Define `Tabular interface { Header() []string; Rows() [][]string }` co-located in `internal/output/formatter.go` next to `Formatter`. Both `TableFormatter` and `PlainFormatter` consume `Tabular` data. `cmd/list.go` provides a `worktreeRows` adapter that projects `[]*core.WorktreeInfo` to `Tabular` with `Header() = {"BRANCH","PATH","STATUS"}` and `Rows() = []string{branch, path, status}` per worktree.

**Rationale**: `golang-cli/references/output.md:42`: "`table` and `plain` need data implementing `Tabular` (`Header()`, `Rows()`); build that adapter in `cmd/` so presentation stays out of the core." The `Tabular` shape unifies the two formatters and gives scripts a stable contract for `cut`/`awk` consumption of `plain` output. Co-locating with `Formatter` matches the project file convention (interfaces beside primary consumer — see `internal/output/table.go` keeping `RenderTable` separate as a renderer, not as a separate-interface file). The first column is the most-stable identifier (`BRANCH`) per the skill convention; subsequent columns follow the same order in `Header()` and `Rows()`. Only `list` adopts the adapter here; other list-style commands opt in when their output formatting is otherwise revisited, so this change stays scoped.

**Alternatives considered**:
- *Keep raw `[][]string` for `TableFormatter` and add a separate `PlainTSVFormatter`*: leaves the two formatters with different input contracts and bypasses the canonical `Tabular` projection.
- *Define `Tabular` in `internal/output/tabular.go` as a separate file*: a 5-line interface alone in a file splits cohesion for no benefit. The project convention keeps interfaces beside their consumer (`Formatter` interface in `formatter.go`); `Tabular` follows the same rule.
- *Define `Tabular` in `internal/core` and add a `String() string` method on `*core.WorktreeInfo`*: forces a global `Stringer` (see Decision 5) that ripples into logging and debug output.

### Decision 11: `PlainFormatter` accepts only `Tabular`

**Choice**: `PlainFormatter.Write(w, data any)` requires `data` to satisfy `Tabular`. Non-`Tabular` data returns `(nil, *core.UsageError)`. Single-object commands that need raw text (e.g., `cd` printing a path to stdout) call `fmt.Fprintln(opts.IO.Out, path)` directly without going through `PlainFormatter`.

**Rationale**: Aligns with `golang-cli/references/output.md:42` ("`table` and `plain` need data implementing `Tabular`") and prevents the previous behavior where `PlainFormatter` emitted `fmt.Sprintln` per slice element (non-tabular, non-scriptable). One-object commands emit exactly one line; there is no row-vs-row distinction to script. Forcing them through `Tabular` would require a degenerate single-row `Tabular` adapter for no benefit.

**Alternatives considered**:
- *Keep `PlainFormatter` accepting `any` and dispatch on shape*: hidden complexity; TSV path requires a `[]string` projection; the spec gets ambiguous.

### Decision 12: Drop `jsonl` and `JSONLinesFormatter`

**Choice**: `jsonl` is removed from the vocabulary. The `JSONLinesFormatter` type, `FormatJSONL` const, and three `TestJSONLinesFormatter_*` tests are deleted. The package doc in `internal/output/doc.go` no longer mentions `jsonl`.

**Rationale**: Project is pre-1.0 and single-user. Streaming consumers pipe `jq -c .` over `json` for the same effect. The `JSONLinesFormatter` has no caller outside its own tests after the consolidation. Less surface, fewer scenarios to maintain, simpler vocabulary. The change is breaking for any external caller using `--output jsonl`; pre-1.0 single-user status justifies the clean break. Truly large/streaming collections (none in this codebase today — every list-style command returns a small collection) would reintroduce `JSONLinesFormatter` and the `jsonl` vocabulary entry; until then, `jq -c .` over `--output json` covers the use case.

**Alternatives considered**:
- *Keep `jsonl` as a hidden streaming path*: dead scaffolding; same problem as the removed `docgen` task.

## Risks / Trade-offs

- **Public API breaking change (envelope drop, `text` rejection, `plain` TSV shape, `jsonl` removal)**: scripts using the old shapes break. Mitigation: documented in `proposal.md` under "Breaking changes"; pre-1.0 single-user status means no deprecation cycle. Force-fail tests pin each new shape so a regression is caught at CI.
- **Dispatch reorder changes visible output for `OperationError{Cause: ValidationError}` wrappers**: previously rendered as `OperationError` (no field/value context); now renders as `ValidationError` (with field/value context). This IS the spec-mandated behavior. Mitigation: force-fail test `TestFormatError_ValidationWinsOverOperation` pins the new dispatch order.
- **`internal/output.NewFormatter` signature change ripples to every caller**: currently only `cmd/list.go` calls it; no other callers exist in the repo per `grep -r "output.NewFormatter"`. Mitigation: signature change is mechanical (one-line update per call site).
- **`DisableDefaultCmd = true` interacts with cobra version**: a future cobra major that renames the option breaks compilation. Mitigation: low risk; the option is stable since cobra v1.4 (2022).
- **`Tabular` interface ripples to every list-style command**. Mitigation: each adapter is one struct per command; the `Tabular` interface contract is stable.
- **`--output plain` shape change breaks existing scripts**: any pipeline consuming the previous `fmt.Sprintln`-per-element output must adopt TSV. Mitigation: e2e golden tests pin the new shape; pre-1.0 status means no deprecation cycle.
- **Bespoke plain-text rendering in `cmd/list.go` is now coupled to `core.WorktreeInfo` field changes**: if `WorktreeInfo` gains/loses fields, `displayWorktrees` must follow. Mitigation: low risk; the project is the only consumer of `WorktreeInfo`; the bespoke shape is captured in the `cli-list` ADDED requirement (deferred to `/osc-continue`) so future changes force a spec update.
- **Quiet-mode hint filter is gated by prefix `hint:` (lowercase)**: matches the existing live-code convention in `internal/output/errors.go:146`; the dead-code filter used `Hint:` (uppercase) and would silently miss live hints. Mitigation: extract `shouldEmitHints(ios)` helper as the single source for the gate; the `TestFormatError_QuietStripsHints` test pins lowercase absence.
- **`JSONLinesFormatter` removal is API-breaking for any external caller**: verified by `git grep -nE '\bJSONLinesFormatter\b|FormatJSONL\b' .` returning no production callers outside archive. Mitigation: grep gate in task 2.0.
- **`writeFieldContext` helper consolidation**: `formatValidationError` and `formatOperationError` previously had duplicate field-context rendering blocks (10 LOC). Mitigation: shared helper; both call sites tested.

## Migration Plan

This change has no deployment steps. It is a pre-1.0 source-only change. The breaking changes (`-o text` rejection, JSON envelope removal, `-o plain` TSV shape, `-o jsonl` removal) take effect on the next published build; no flag-gated compatibility path. Rollback is `git revert` of the change's merge commit.

## Open Questions

None. All decisions above are resolved.
