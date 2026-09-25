# Proposal

## Why

The 2026-09-25 codebase audit (`REVIEW.md`) flagged two intertwined problems in the cmd layer: dead production code in `cmd/error_formatter.go` (~180 lines of strategy-pattern code reachable only from its own tests) and `cmd/output.go` (bespoke `OutputFormatter` types that shadow the generic `internal/output` package). The dead code hides three live spec-compliance bugs: dispatch order in `internal/output/errors.go` is the reverse of what `cli-error-formatting` mandates, the four-sentinel hint table required by spec only exists in the dead code, and quiet-mode hint suppression from spec is absent. The shadowed output package hides a vocabulary drift: `cmd/list.go` accepts `text|json` while spec mandates `json|jsonl|table|plain`, and JSON output wraps worktrees in a `{"worktrees":[…]}` envelope that no spec ever required. Consolidating onto `internal/output` and absorbing the dead-code behavior closes these gaps in one mechanical pass.

## What Changes

- **Delete `cmd/error_formatter.go`** (~180 lines) and **`cmd/error_formatter_test.go`** (~260 lines of tests for unreachable code).
- **Absorb `hintFor` and quiet-mode hint suppression** into `internal/output/errors.go`, dispatching via `errors.Is` on existing `core.Err{Project,Worktree,Resolution,GitRepo}NotFound` sentinels (already defined in `internal/core/sentinels.go`).
- **Reorder dispatch in `internal/output/errors.go`** to `*core.ValidationError` → `*core.NotFoundError` → `*core.OperationError` → `*core.UsageError` to match the `cli-error-formatting` spec scenario "Specific matcher wins".
- **Switch error dispatch to `errors.AsType[T]`** (Go 1.26+, available after the foundation change bumps `go.mod`); drop the `asType[T error]` workaround helper that exists only because the project targets Go 1.25.
- **Delete `cmd/output.go`** (~88 lines of `OutputFormatter`/`TextFormatter`/`JSONFormatter`/`WorktreeJSON`/`WorktreeListJSON`) after migrating `cmd/list.go` to `internal/output.NewFormatter(format)`.
- **Change `internal/output.NewFormatter` signature** from `(format string) Formatter` to `(format string) (Formatter, error)`; unknown/empty returns `*core.UsageError`.
- **Delete dead code**: `cmd/util.go:ProgressReporter.ReportProgress` method, `cmd/create.go:CreateOptions.HookRunner` field, `cmd/completion.go:41-43` redundant `ValidArgsFunction`.
- **Add `cmd.CompletionOptions.DisableDefaultCmd = true`** in `cmd/root.go`, removing the fragile `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand` dance around carapace's auto-registered completion (matches the canonical pattern in `golang-cli/references/completion.md`).
- **Add `cmd.AddGroup(...)` calls** in `cmd/root.go` before the `AddCommand` block: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion/docgen) — cobra does not retroactively assign groups.
- **Wire a hidden `twiggit docgen <dir>` command** using `cobra/doc.GenMarkdownTree`; add a `mise run docgen` task that builds, runs docgen, and exits non-zero if generated differs.
- **BREAKING**: `twiggit list -o text` → previously rendered text (exit 0); now returns `core.UsageError` (exit 2). Spec `cli-output` mandates the narrowed vocabulary.
- **BREAKING**: `twiggit list -o json` previously emitted `{"worktrees":[{...}]}`; now emits a bare JSON array `[{...}]`. Scripts using `jq '.worktrees[]'` must switch to `jq '.[]'`. Spec `cli-output` JSON-shape table requires this for small collections.

## Capabilities

### New Capabilities

None. The hidden `docgen` command, `AddGroup` calls, and completion-dance simplification are implementation details that do not change user-visible behavior outside the two breaking changes above.

### Modified Capabilities

- `cli-output`: the `--output` vocabulary is narrowed to `json|jsonl|table|plain` and the unknown-value error type becomes `core.UsageError`. The JSON shape for collections is a bare array (envelope removed). `NewFormatter` returns `(Formatter, error)` instead of `Formatter`.

`cli-error-formatting` is not modified: its scenarios already specify the post-change dispatch order (Validation wins when both match), the per-sentinel hint table, and quiet-mode hint suppression. The change makes the implementation honor what `cli-error-formatting` already mandates; it does not introduce new requirement text.

## Impact

- **Source files (production)**:
  - `internal/output/errors.go` — dispatch reorder, `hintFor`, quiet gating, `errors.AsType` adoption
  - `internal/output/formatter.go` — `NewFormatter` signature change
  - `cmd/list.go` — migrate to `internal/output.NewFormatter`; reject `text` as `core.UsageError`; bespoke plain-text rendering moves inline into `displayWorktrees`
  - `cmd/root.go` — `CompletionOptions.DisableDefaultCmd = true`; `AddGroup` calls; hidden `docgen` command registration
  - `cmd/docgen.go` — new file, hidden command wrapping `cobra/doc.GenMarkdownTree`
  - `cmd/util.go` — drop `ProgressReporter.ReportProgress`
  - `cmd/create.go` — drop `CreateOptions.HookRunner` field
  - `cmd/completion.go` — drop redundant `ValidArgsFunction`
  - Delete: `cmd/error_formatter.go`, `cmd/output.go`
- **Source files (tests)**:
  - `internal/output/errors_test.go` — port and add: `TestFormatError_NotFoundHints` (4-sentinel table), `TestFormatError_QuietStripsHints`, `TestFormatError_ValidationWinsOverOperation` (specificity), `TestFormatError_UsageErrorFirst`
  - `internal/output/formatter_test.go` — add `TestNewFormatter_UnknownReturnsUsageError`, `TestNewFormatter_EmptyReturnsUsageError`
  - Delete: `cmd/error_formatter_test.go`
- **Tooling**: `mise.toml` — add `docgen` task (`go build`, run `twiggit docgen docs/`, `git diff --exit-code docs/`).
- **Spec deltas**: `openspec/changes/cli-cobra-output-format-unification/specs/cli-error-formatting/spec.md`, `specs/cli-output/spec.md`.
- **Dependencies**: none. `cobra/doc` is already a transitive of `github.com/spf13/cobra`.
- **CI**: no change; `mise run verify` covers the touched layers.
- **Breaking changes**:
  - `twiggit list -o text` — now exits 2 with `Usage:` (was exit 0 with text output).
  - `twiggit list -o json` — shape changes from `{"worktrees":[…]}` to `[…]` (bare array).
