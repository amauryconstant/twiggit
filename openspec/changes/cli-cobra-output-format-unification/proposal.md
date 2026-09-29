# Proposal

## Why

The codebase audit flagged two intertwined problems in the cmd layer: dead production code in `cmd/error_formatter.go` (~180 lines of strategy-pattern code reachable only from its own tests) and `cmd/output.go` (bespoke `OutputFormatter` types that shadow the generic `internal/output` package). The dead code hides three live spec-compliance bugs: dispatch order in `internal/output/errors.go` is the reverse of what `cli-error-formatting` mandates, the four-sentinel hint table required by spec only exists in the dead code, and quiet-mode hint suppression from spec is absent. The shadowed output package hides a vocabulary drift: `cmd/list.go` accepts `text|json` while the spec mandates `json|jsonl|table|plain` (we narrow this further to `json|table|plain` in this change), the bespoke JSON path wraps worktrees in a `{"worktrees":[…]}` envelope that no spec required, and `PlainFormatter` emits `fmt.Sprintln` per element instead of headerless TSV that the canonical CLI shape demands. Consolidating onto `internal/output` and absorbing the dead-code behavior closes these gaps in one mechanical pass.

This is a pre-1.0 single-user tool; free to break anything.

## What Changes

- **Delete `cmd/error_formatter.go`** (~180 lines) and **`cmd/error_formatter_test.go`** (~260 lines of tests for unreachable code).
- **Absorb `hintFor` and quiet-mode hint suppression** into `internal/output/errors.go`, dispatching via `errors.Is` on existing `core.Err{Project,Worktree,Resolution,GitRepo}NotFound` sentinels (already defined in `internal/core/sentinels.go`).
- **Reorder dispatch in `internal/output/errors.go`** to `*core.ValidationError` → `*core.NotFoundError` → `*core.OperationError` → `*core.UsageError` to match the `cli-error-formatting` spec scenario "Specific matcher wins".
- **Switch error dispatch to `errors.AsType[T]`** (Go 1.26+, available on the current `go 1.27.1` toolchain); drop the `asType[T error]` workaround helper.
- **Delete `cmd/output.go`** (~88 lines of `OutputFormatter`/`TextFormatter`/`JSONFormatter`/`WorktreeJSON`/`WorktreeListJSON`) after migrating `cmd/list.go` to `internal/output.NewFormatter`.
- **Change `internal/output.NewFormatter` signature** from `(format string) Formatter` to `(format string) (Formatter, error)`; empty format returns `(nil, nil)` so each command picks its own human default; unknown values (including the legacy `text`) return `(nil, *core.UsageError)`.
- **Drop `jsonl` from the vocabulary**: vocabulary narrows from `json|jsonl|table|plain` to `json|table|plain`. `JSONLinesFormatter` type, `FormatJSONL` const, and the three `TestJSONLinesFormatter_*` tests are deleted.
- **Introduce `Tabular` interface** co-located in `internal/output/formatter.go` next to `Formatter`, with `Header() []string` and `Rows() [][]string`; rewrite `TableFormatter` and `PlainFormatter` to consume `Tabular`. `PlainFormatter` emits headerless TSV (per the canonical CLI shape in `golang-cli/references/output.md`).
- **Delete dead code**: `cmd/util.go:ProgressReporter.ReportProgress` method, `cmd/create.go:CreateOptions.HookRunner` field, `cmd/completion.go:41-43` redundant `ValidArgsFunction`.
- **Add `cmd.CompletionOptions.DisableDefaultCmd = true`** in `cmd/root.go`, removing the fragile `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand` dance around carapace's auto-registered completion.
- **Add `cmd.AddGroup(...)` calls** in `cmd/root.go` before the `AddCommand` block: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion) — cobra does not retroactively assign groups.
- **BREAKING**: `twiggit list -o text` → previously rendered text (exit 0); now returns `core.UsageError` (exit 2). Spec `cli-output` mandates the narrowed vocabulary.
- **BREAKING**: `twiggit list -o json` previously emitted `{"worktrees":[{...}]}`; now emits a bare JSON array `[{...}]`. Scripts using `jq '.worktrees[]'` must switch to `jq '.[]'`. Spec `cli-output` JSON-shape table requires this for small collections.
- **BREAKING**: `twiggit list -o plain` previously emitted one `fmt.Sprintln` line per worktree (e.g., `branch -> path (modified)`); now emits headerless TSV (`branch<TAB>path<TAB>status`). The `Tabular` projection is the canonical shape; pipelines expecting per-element plain text must consume TSV instead.
- **BREAKING**: `twiggit list -o jsonl` → now returns `core.UsageError` (exit 2). The vocabulary drops `jsonl`; streaming consumers pipe `jq -c .` over `json` instead.

## Capabilities

### New Capabilities

- `cli-command-groups`: `--help` output groups subcommands via `cmd.AddGroup`. Four groups: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion).

### Modified Capabilities

- `cli-output`: the `--output` vocabulary is narrowed from `json|jsonl|table|plain` to `json|table|plain` with explicit rejection of the legacy `text` value; `plain` is redefined as headerless TSV; the `NewFormatter` signature gains the `(Formatter, error)` return with empty-format returning `(nil, nil)`. New requirements are added for the `Tabular` projection interface, the bare-collection JSON shape, and the constructor contract. The existing "Default (no --output) falls through to plain" requirement is MODIFIED to defer the human default to the per-command spec.
- `cli-output-formats`: the `--output` vocabulary drops `jsonl`; the JSON shape pin changes from envelope (`{"worktrees":[…]}`) to bare array (`[…]`) for collections; the "tabular text by default for listings" requirement is MODIFIED to defer the per-command interactive default to the owning command's spec.
- `cli-list`: the `--output json` requirement is MODIFIED from "JSON object" to "JSON array" so the shape is unambiguous; a new requirement declares `list`'s bespoke interactive default rendering (`BRANCH -> PATH` with conditional ` (modified)` / ` (detached)` suffixes) for empty `--output`.

`cli-error-formatting` is not modified: its scenarios already specify the dispatch order (Validation wins when both match), the per-sentinel hint table, and quiet-mode hint suppression. The change makes the implementation honor what `cli-error-formatting` already mandates; it does not introduce new requirement text.

## Impact

- **Source files (production)**:
  - `internal/output/formatter.go` — `Tabular` interface added next to `Formatter`; `TableFormatter`/`PlainFormatter` consume `Tabular`; `NewFormatter` signature change; `JSONLinesFormatter` type, `FormatJSONL` const, and the `jsonl` switch arm removed; `PlainFormatter` emits headerless TSV
  - `internal/output/errors.go` — dispatch reorder, `hintFor`, quiet gating, `errors.AsType` adoption, `shouldEmitHints` and `writeFieldContext` helpers (single source for quiet gate and field-context rendering)
  - `internal/output/doc.go` — drop `jsonl` from package doc
  - `cmd/list.go` — migrate to `internal/output.NewFormatter`; reject `text` as `core.UsageError`; `worktreeRows` `Tabular` adapter; empty `--output` keeps bespoke human rendering
  - `cmd/root.go` — `CompletionOptions.DisableDefaultCmd = true`; `AddGroup` calls
  - `cmd/util.go` — drop `ProgressReporter.ReportProgress`
  - `cmd/create.go` — drop `CreateOptions.HookRunner` field
  - `cmd/completion.go` — drop redundant `ValidArgsFunction`
  - Delete: `cmd/error_formatter.go`, `cmd/output.go`
- **Source files (tests)**:
  - `internal/output/errors_test.go` — port and add: `TestFormatError_NotFoundHints` (4-sentinel table), `TestFormatError_QuietStripsHints`, `TestFormatError_ValidationWinsOverOperation` (specificity), `TestFormatError_UsageErrorFirst`
  - `internal/output/formatter_test.go` — drop `TestJSONLinesFormatter_*`; add `TestNewFormatter_UnknownReturnsUsageError`, `TestNewFormatter_TextReturnsUsageError`, `TestNewFormatter_EmptyReturnsNil`, `TestTableFormatter_Tabular`, `TestPlainFormatter_TabularTSV`, `TestPlainFormatter_NonTabularReturnsUsageError`
  - Delete: `cmd/error_formatter_test.go`
  - `test/e2e/` — regenerate golden files for `list -o plain` (shape changed from `Sprintln` per element to headerless TSV)
- **Documentation**: `cmd/AGENTS.md` — update "Error handling" section to reflect the new dispatch order (Validation → NotFound → Operation → Usage) and reference `cli-error-formatting` for the per-sentinel hint table.
- **Spec deltas** (all included in this update):
  - `openspec/changes/cli-cobra-output-format-unification/specs/cli-output/spec.md` — MODIFIED + ADDED (existing Req 2 restated to drop `JSONLinesFormatter`; new requirements for `Tabular` projection, narrowed vocabulary, per-command default, bare-collection JSON shape, constructor contract, shell completion)
  - `openspec/changes/cli-cobra-output-format-unification/specs/cli-output-formats/spec.md` — REMOVED + ADDED (vocabulary narrows from `json|table|plain|jsonl` to `json|table|plain`; per-command interactive default rule; shell completion)
  - `openspec/changes/cli-cobra-output-format-unification/specs/cli-list/spec.md` — REMOVED + MODIFIED + ADDED (envelope JSON shape dropped; status indicators restated to capture the `BRANCH -> PATH` interactive default; bare JSON array output)
  - `openspec/changes/cli-cobra-output-format-unification/specs/cli-command-groups/spec.md` — ADDED (new capability for `--help` group labels)
- **Dependencies**: none. No new imports.
- **CI**: no change; `mise run verify` covers the touched layers.
- **Breaking changes**:
  - `twiggit list -o text` — now exits 2 with `Usage:` (was exit 0 with text output).
  - `twiggit list -o json` — shape changes from `{"worktrees":[…]}` to `[…]` (bare array).
  - `twiggit list -o plain` — shape changes from `fmt.Sprintln` per element to headerless TSV.
  - `twiggit list -o jsonl` — now exits 2 with `Usage:` (vocabulary drops jsonl).
