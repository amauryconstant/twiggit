# Proposal

## Why

The 2026-09-25 codebase audit (`REVIEW.md`) flagged two intertwined problems in the cmd layer: dead production code in `cmd/error_formatter.go` (~180 lines of strategy-pattern code reachable only from its own tests) and `cmd/output.go` (bespoke `OutputFormatter` types that shadow the generic `internal/output` package). The dead code hides three live spec-compliance bugs: dispatch order in `internal/output/errors.go` is the reverse of what `cli-error-formatting` mandates, the four-sentinel hint table required by spec only exists in the dead code, and quiet-mode hint suppression from spec is absent. The shadowed output package hides a vocabulary drift: `cmd/list.go` accepts `text|json` while `cli-output` mandates `json|jsonl|table|plain`, the bespoke JSON path wraps worktrees in a `{"worktrees":[…]}` envelope that no spec required, and `PlainFormatter` emits `fmt.Sprintln` per element instead of headerless TSV that the canonical CLI shape demands. Consolidating onto `internal/output` and absorbing the dead-code behavior closes these gaps in one mechanical pass.

## What Changes

- **Delete `cmd/error_formatter.go`** (~180 lines) and **`cmd/error_formatter_test.go`** (~260 lines of tests for unreachable code).
- **Absorb `hintFor` and quiet-mode hint suppression** into `internal/output/errors.go`, dispatching via `errors.Is` on existing `core.Err{Project,Worktree,Resolution,GitRepo}NotFound` sentinels (already defined in `internal/core/sentinels.go`).
- **Reorder dispatch in `internal/output/errors.go`** to `*core.ValidationError` → `*core.NotFoundError` → `*core.OperationError` → `*core.UsageError` to match the `cli-error-formatting` spec scenario "Specific matcher wins".
- **Switch error dispatch to `errors.AsType[T]`** (Go 1.26+, available on the current `go 1.27.1` toolchain); drop the `asType[T error]` workaround helper.
- **Delete `cmd/output.go`** (~88 lines of `OutputFormatter`/`TextFormatter`/`JSONFormatter`/`WorktreeJSON`/`WorktreeListJSON`) after migrating `cmd/list.go` to `internal/output.NewFormatter`.
- **Change `internal/output.NewFormatter` signature** from `(format string) Formatter` to `(format string) (Formatter, error)`; unknown/`text`/empty semantics per the modified `cli-output` spec.
- **Introduce `Tabular` interface in `internal/output/tabular.go`** with `Header() []string` and `Rows() [][]string`; rewrite `TableFormatter` and `PlainFormatter` to consume `Tabular`. `PlainFormatter` emits headerless TSV (per the canonical CLI shape in `golang-cli/references/output.md`).
- **Delete dead code**: `cmd/util.go:ProgressReporter.ReportProgress` method, `cmd/create.go:CreateOptions.HookRunner` field, `cmd/completion.go:41-43` redundant `ValidArgsFunction`.
- **Add `cmd.CompletionOptions.DisableDefaultCmd = true`** in `cmd/root.go`, removing the fragile `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand` dance around carapace's auto-registered completion.
- **Add `cmd.AddGroup(...)` calls** in `cmd/root.go` before the `AddCommand` block: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion/docgen) — cobra does not retroactively assign groups.
- **Wire a hidden `twiggit docgen <dir>` command** using `cobra/doc.GenMarkdownTree`; add a `mise run docgen` task that builds, runs docgen, and exits non-zero if generated differs.
- **BREAKING**: `twiggit list -o text` → previously rendered text (exit 0); now returns `core.UsageError` (exit 2). Spec `cli-output` mandates the narrowed vocabulary.
- **BREAKING**: `twiggit list -o json` previously emitted `{"worktrees":[{...}]}`; now emits a bare JSON array `[{...}]`. Scripts using `jq '.worktrees[]'` must switch to `jq '.[]'`. Spec `cli-output` JSON-shape table requires this for small collections.
- **BREAKING**: `twiggit list -o plain` previously emitted one `fmt.Sprintln` line per worktree (e.g., `branch -> path (modified)`); now emits headerless TSV (`branch<TAB>path<TAB>status`). The `Tabular` projection is the canonical shape; pipelines expecting per-element plain text must consume TSV instead.

## Capabilities

### New Capabilities

None. The hidden `docgen` command, `AddGroup` calls, and completion-dance simplification are implementation details that do not change user-visible behavior outside the three breaking changes above.

### Modified Capabilities

- `cli-output`: the `--output` vocabulary is narrowed to `json|jsonl|table|plain` with explicit rejection of the legacy `text` value; `plain` is redefined as headerless TSV; the `NewFormatter` signature gains the `(Formatter, error)` return; the existing `Formatter` interface requirement adds the constructor contract. New requirements are added for the `Tabular` projection interface and the bare-collection JSON shape.

`cli-error-formatting` is not modified: its scenarios already specify the dispatch order (Validation wins when both match), the per-sentinel hint table, and quiet-mode hint suppression. The change makes the implementation honor what `cli-error-formatting` already mandates; it does not introduce new requirement text.

## Impact

- **Source files (production)**:
  - `internal/output/tabular.go` — new file, defines `Tabular` interface
  - `internal/output/errors.go` — dispatch reorder, `hintFor`, quiet gating, `errors.AsType` adoption
  - `internal/output/formatter.go` — `TableFormatter`/`PlainFormatter` consume `Tabular`; `NewFormatter` signature change; `PlainFormatter` emits headerless TSV
  - `cmd/list.go` — migrate to `internal/output.NewFormatter`; reject `text` as `core.UsageError`; `worktreeRows` `Tabular` adapter; empty `--output` keeps bespoke human rendering
  - `cmd/root.go` — `CompletionOptions.DisableDefaultCmd = true`; `AddGroup` calls; hidden `docgen` command registration
  - `cmd/docgen.go` — new file, hidden command wrapping `cobra/doc.GenMarkdownTree`
  - `cmd/util.go` — drop `ProgressReporter.ReportProgress`
  - `cmd/create.go` — drop `CreateOptions.HookRunner` field
  - `cmd/completion.go` — drop redundant `ValidArgsFunction`
  - Delete: `cmd/error_formatter.go`, `cmd/output.go`
- **Source files (tests)**:
  - `internal/output/errors_test.go` — port and add: `TestFormatError_NotFoundHints` (4-sentinel table), `TestFormatError_QuietStripsHints`, `TestFormatError_ValidationWinsOverOperation` (specificity), `TestFormatError_UsageErrorFirst`
  - `internal/output/formatter_test.go` — add `TestNewFormatter_UnknownReturnsUsageError`, `TestNewFormatter_TextReturnsUsageError`, `TestNewFormatter_EmptyReturnsPlain`, `TestTableFormatter_Tabular`, `TestPlainFormatter_TabularTSV`, `TestPlainFormatter_NonTabularReturnsUsageError`
  - Delete: `cmd/error_formatter_test.go`
  - `test/e2e/` — regenerate golden files for `list -o plain` (shape changed from `Sprintln` per element to headerless TSV)
- **Tooling**: `mise.toml` — add `docgen` task (`go build`, run `twiggit docgen docs/`, `git diff --exit-code docs/`).
- **Documentation**: `cmd/AGENTS.md` — update "Error handling" section to reflect the new dispatch order (Validation → NotFound → Operation → Usage).
- **Spec deltas**: `openspec/changes/cli-cobra-output-format-unification/specs/cli-output/spec.md`.
- **Dependencies**: none. `cobra/doc` is already a transitive of `github.com/spf13/cobra`.
- **CI**: no change; `mise run verify` covers the touched layers.
- **Breaking changes**:
  - `twiggit list -o text` — now exits 2 with `Usage:` (was exit 0 with text output).
  - `twiggit list -o json` — shape changes from `{"worktrees":[…]}` to `[…]` (bare array).
  - `twiggit list -o plain` — shape changes from `fmt.Sprintln` per element to headerless TSV.
