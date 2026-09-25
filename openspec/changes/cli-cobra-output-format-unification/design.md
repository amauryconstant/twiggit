# Design

## Context

Two parallel output subsystems exist. `internal/output/` (the canonical home) provides generic `JSONFormatter`, `TableFormatter`, `PlainFormatter`, `JSONLinesFormatter` types plus `errors.go:FormatError` which is the production error formatter wired into `main.go`. `cmd/output.go` and `cmd/error_formatter.go` (~268 lines) shadow this with bespoke types and a strategy-pattern error formatter that no production code path reaches; only their own test files exercise them.

Three spec-mandated behaviors live only in the dead code: (1) the per-sentinel hint table in `cli-error-formatting/spec.md:79-91`, (2) the "Specific matcher wins" registration order at `cli-error-formatting/spec.md:64-78` (the implementation has the reverse order in `internal/output/errors.go:34-53`), (3) the "Quiet mode strips hints" scenario at `cli-error-formatting/spec.md:107-111`. Consolidating onto `internal/output` and absorbing the dead-code behavior is one mechanical pass.

A separate output-vocabulary drift exists between `cli-output/spec.md:13-44` (mandates `json|jsonl|table|plain`) and `cmd/list.go:67-69` (accepts `text|json`, rejects everything else as a plain `fmt.Errorf` exit 1). The bespoke JSON path in `cmd/output.go:43-65` wraps worktrees in `{"worktrees":[…]}` — no spec ever required the envelope.

The project targets Go 1.25.5 today (`go.mod`); the foundation change `cli-foundation-toolchain-and-lints` bumps to ≥1.26.2 to close a stdlib CVE window, which unlocks `errors.AsType[T]` (Go 1.26+). This change assumes the foundation change has landed first.

## Goals / Non-Goals

**Goals:**
- Make `internal/output/errors.go` honor every requirement in `cli-error-formatting`.
- Make `cmd/list.go` honor every requirement in `cli-output`, with output produced via the canonical `internal/output.NewFormatter`.
- Remove dead production code (`cmd/error_formatter.go`, `cmd/output.go`, `ProgressReporter.ReportProgress`, `CreateOptions.HookRunner`, redundant `ValidArgsFunction`) and the test code that targets it.
- Provide a `cobra/doc`-based generation path so help output is regenerable from a single `mise` task.
- Add `cmd.AddGroup` labels so `--help` output is scannable for the seven live subcommands.

**Non-Goals:**
- Switching all `opts.IO.Out` writes to `cmd.OutOrStdout()` — would regress the `iostreams.Test()` injection seam that 30+ test files depend on; `cmd/AGENTS.md` codifies the divergence.
- Renaming the `cmd/util.go` file or migrating to the testkit package layout — belongs in a separate refactor change.
- Reordering the `cli-error-formatting` dispatch table to put `OperationError` first — the spec mandates the current order, the implementation was wrong; no spec change.
- Reintroducing the legacy shell sentinels (`ErrShellAlreadyInstalled`, etc.) — belongs in the in-progress `test-observability-error-discipline` change (task 4.1).
- Bumping the module path from `twiggit` to a URL-shaped identifier — deferred to a separate change.
- Bulk migration of `cli-`/`application-`/`domain-`/`infrastructure-` spec prefixes to the new `cli-`/`core-`/`git-`/`testing-` set per AGENTS.md — deferred; current spec catalog carries both prefixes coexisting.

## Decisions

### Decision 1: Delete `cmd/error_formatter.go` and absorb behavior into `internal/output/errors.go`

**Choice**: Delete `cmd/error_formatter.go` (~180 lines), delete `cmd/error_formatter_test.go` (~260 lines), move `hintFor` into `internal/output/errors.go`, add quiet-mode gating, reorder dispatch, switch to `errors.AsType[T]`. Port the four unique test scenarios to `internal/output/errors_test.go`.

**Rationale**: The strategy-pattern types in `cmd/error_formatter.go` are unreachable from production (`main.go` calls `output.FormatError`). The three live spec behaviors (hint table, dispatch order, quiet suppression) are reachable only after the move. Deletion removes ~340 lines of orphan code; the four ported test scenarios (~80 lines) cover the same behavior with less indirection.

**Alternatives considered**:
- *Keep both formatters, route production through `cmd`*: would require changing `main.go` to call `cmd.ErrorFormatter` instead of `output.FormatError`. Violates the canonical-home rule (errors formatted beside their origin) and creates two type families to maintain.
- *Use the dead code's strategy pattern via dependency injection*: would add a `Formatter` interface field to `main.go` and select one at boot. Pure ceremony for a single live implementation.

### Decision 2: Delete `cmd/output.go` and migrate `cmd/list.go` to `internal/output.NewFormatter`

**Choice**: Delete `cmd/output.go` (~88 lines). Rewrite `cmd/list.go:129-138` to call `internal/output.NewFormatter(format)` and pass the `[]*core.WorktreeInfo` slice directly to `formatter.Write(w, worktrees)`. Keep bespoke plain-text rendering inline in `cmd/list.go`'s `displayWorktrees`.

**Rationale**: `internal/output/formatter.go` already implements all four spec-mandated formatters (`json|jsonl|table|plain`). The `cmd/output.go` `JSONFormatter` adds a top-level `{"worktrees":[…]}` envelope that no spec requires and that breaks `jq '.[]'` consumption; `internal/output.JSONFormatter` emits a bare array, which matches the JSON-shape table in `golang-cli/references/output.md`. Bespoke text rendering ("branch -> path (modified)") is one-command-specific; per `golang-cli/references/commands.md` ("output beside the logic"), it stays beside `cmd/list.go`.

**Alternatives considered**:
- *Move bespoke rendering to `internal/output` as a `WorktreeFormatter`*: would require defining a `String() string` method on `*core.WorktreeInfo`, changing global representation (ripples into logs, debug output, test failure dumps). Too invasive for a one-command use case.
- *Keep `cmd/output.go` and only migrate the JSON path*: leaves `TextFormatter` orphaned and defeats the consolidation goal.

### Decision 3: Switch error dispatch to `errors.AsType[T]` (Go 1.26+)

**Choice**: At each dispatch arm in `internal/output/errors.go:FormatError`, call `errors.AsType[*core.X](err)`. Drop the `asType[T error]` helper from `cmd/error_formatter.go:13-19` without porting it.

**Rationale**: Go 1.26+ adds `errors.AsType[T](err)` which returns the typed error directly (no double-allocation, no nil guard needed). The foundation change `cli-foundation-toolchain-and-lints` bumps `go.mod` to ≥1.26.2 to close the stdlib CVE window, so by the time this change lands the primitive is available. Porting the `asType` helper would keep a Go 1.25 workaround in place for no benefit.

**Alternatives considered**:
- *Keep `asType[T error]` helper in `internal/output`*: works but adds an indirection layer that the stdlib already provides natively.

### Decision 4: Replace the carapace Find+Remove dance with `cmd.CompletionOptions.DisableDefaultCmd = true`

**Choice**: In `cmd/root.go`, add `cmd.CompletionOptions.DisableDefaultCmd = true` before `carapace.Gen(cmd)`. Delete the `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand(completionCmd)` block. Keep `newCompletionCommand` and `carapace.Gen(cmd)` unchanged.

**Rationale**: The dance exists because cobra auto-registers a `completion` subcommand, carapace auto-registers its own `completion` subcommand, and the project's custom `newCompletionCommand` is meant to replace both. `DisableDefaultCmd = true` is cobra's declarative way to skip the cobra-owned one; carapace's `_carapace` (hidden) remains the runtime path; `newCompletionCommand` remains the user-visible one. The `golang-cli/references/completion.md` reference documents this pattern explicitly.

**Alternatives considered**:
- *Walk to find carapace's `completion` and remove it, leaving cobra's*: cobra's auto-completion lacks the per-shell help text the custom command provides (see `getShellInstructions` at `cmd/completion.go:57-91`). Worse UX.
- *Replace the dance with an `addPreRun` hook*: doesn't solve the issue; the conflict is at command-registration time, not hook-chain time.

### Decision 5: Bespoke plain-text rendering stays inline in `cmd/list.go`

**Choice**: The "branch -> path (modified)(detached)" output stays in `cmd/list.go:displayWorktrees`. No new `internal/output.WorktreeFormatter` type.

**Rationale**: `golang-cli/references/commands.md`: "Output beside the logic — the command knows what it produced and formats it with shared helpers from `internal/output`". The bespoke text shape is specific to one command; promoting it to `internal/output` would force a global `Stringer` on `core.WorktreeInfo` that affects logging and debug dumps everywhere.

**Alternatives considered**:
- *Implement `String() string` on `*core.WorktreeInfo`, use `internal/output.PlainFormatter`*: would work, but the global `Stringer` affects every site that formats a `WorktreeInfo` value (slog messages, debug `fmt.Sprintf`, test failure dumps). Out-of-scope side effects.

### Decision 6: `NewFormatter` returns `(Formatter, error)` instead of `Formatter`

**Choice**: Change signature to `NewFormatter(format string) (Formatter, error)`. Unknown/empty values return `(nil, *core.UsageError)`.

**Rationale**: `golang-cli/references/output.md`: "NewFormatter(format) returns a Formatter or a UsageError for unknown values — resolve it **before** doing any work". The current nil-returning signature forces callers to nil-check and emit their own UsageError, which `cmd/list.go` does correctly but with a plain `fmt.Errorf` (wrong type). Pushing the error into the constructor makes the spec contract enforceable and removes the divergence at the call site.

**Alternatives considered**:
- *Keep nil-for-unknown, fix only `cmd/list.go`*: smaller diff, but the next caller will re-introduce the same bug (plain `fmt.Errorf` instead of `UsageError`). The signature is the right place to enforce.

### Decision 7: `twiggit list -o json` envelope dropped

**Choice**: JSON output for a collection emits a bare array `[{...}]`. The `WorktreeListJSON{Worktrees: []WorktreeJSON{...}}` envelope is removed.

**Rationale**: `golang-cli/references/output.md` JSON shape table: "A small collection — A JSON array". No spec ever mandated the envelope; the envelope was an implementation choice in `cmd/output.go:74-77`. Scripts consuming the old shape must update from `jq '.worktrees[]'` to `jq '.[]'`.

**Alternatives considered**:
- *Keep envelope, add `--legacy-output` flag for back-compat*: pre-1.0 project; no external scripts in the wild depend on the envelope. Carrying a back-compat flag for a one-time shape change is dead weight.
- *Two-version deprecation cycle (warn → break)*: would require two OpenSpec changes and a release-cycle commitment. Project is pre-1.0; single cycle is appropriate.

### Decision 8: `twiggit list -o text` rejected as `core.UsageError`

**Choice**: The `text` value is rejected with `core.UsageError` (exit 2). No alias to `plain`.

**Rationale**: `cli-output/spec.md:13-44` specifies the accepted values; `text` is not among them. The implementation accepts it as a legacy quirk. The clean break is to reject per spec; scripts using `--output text` must switch to `--output plain`. Project is pre-1.0; no deprecation cycle needed.

**Alternatives considered**:
- *Accept `text` as an alias for `plain`*: preserves existing scripts but entrenches a value that the spec does not recognize. The change should make the implementation match the spec, not paper over the drift.

### Decision 9: Hidden `twiggit docgen` command + `mise run docgen` task

**Choice**: Add a hidden subcommand `twiggit docgen <dir>` that calls `cobra/doc.GenMarkdownTree(rootCmd, dir)`. Add a `docgen` task in `.mise/config.toml` that builds, runs docgen, and exits non-zero if the generated tree differs from a committed `docs/` (none committed in this change).

**Rationale**: `golang-spf13-cobra` skill documents `cobra/doc` as the canonical markdown generation path. Hidden = no help-noise; the task is opt-in. Generation is local-only; no docs are committed in this change.

**Alternatives considered**:
- *Commit generated docs to `docs/cli/`*: adds a `docs/` tree to the repo and a regeneration note in CONTRIBUTING.md. Larger scope; defer until a docs-site need actually appears.

### Decision 10: `cmd.AddGroup` for subcommand discoverability

**Choice**: Register four groups via `cmd.AddGroup(...)` in `cmd/root.go` BEFORE the `AddCommand` block: `core` (list/create/delete/prune), `navigation` (cd), `setup` (init), `meta` (version/completion/docgen). Per `golang-spf13-cobra` skill: groups must be registered before the commands that reference them.

**Rationale**: With seven live subcommands the help wall is hard to scan. Group labels make `--help` output match the user's mental model. Cosmetic; no behavior change beyond the help text.

**Alternatives considered**:
- *No groups*: simpler, but `--help` output degrades as more commands land.

## Risks / Trade-offs

- **Public API breaking change (envelope drop, `text` rejection)**: scripts using the old shape break. Mitigation: documented in `proposal.md` under "Impact > Breaking changes"; pre-1.0 status means no deprecation cycle. Force-fail tests pin the new shape so a regression to envelope is caught at CI.
- **Dispatch reorder changes visible output for `OperationError{Cause: ValidationError}` wrappers**: previously rendered as `OperationError` (no field/value context); now renders as `ValidationError` (with field/value context). This IS the spec-mandated behavior. Mitigation: force-fail test `TestFormatError_ValidationWinsOverOperation` pins the new dispatch order.
- **`internal/output.NewFormatter` signature change ripples to every caller**: currently only `cmd/list.go` calls it; no other callers exist in the repo per `grep -r "output.NewFormatter"`. Mitigation: signature change is mechanical (one-line update per call site).
- **`DisableDefaultCmd = true` interacts with cobra version**: a future cobra major that renames the option breaks compilation. Mitigation: low risk; the option is stable since cobra v1.4 (2022).
- **Bespoke plain-text rendering in `cmd/list.go` is now coupled to `core.WorktreeInfo` field changes**: if `WorktreeInfo` gains/loses fields, `displayWorktrees` must follow. Mitigation: low risk; the project is the only consumer of `WorktreeInfo`.
- **The `docgen` task requires `cobra/doc` import**: already a transitive of `github.com/spf13/cobra`, no new dependency.
- **Hidden `docgen` command shows in `twiggit --help` only when explicitly invoked**: cobra's `Hidden: true` flag suppresses the entry from help; the task is the discoverable surface.
- **Quiet-mode hint suppression may surface a different rendering for scripts**: tests in `cli-quiet-mode` already cover this; no new test surface needed.

## Migration Plan

This change has no deployment steps. It is a pre-1.0 source-only change. The two breaking changes (`-o text` rejection, JSON envelope removal) take effect on the next release; no flag-gated compatibility path. Rollback is a `git revert` of the merge commit.

## Open Questions

None. All decisions above are resolved; remaining unknowns (module path rename, spec prefix migration, error-kind taxonomy) belong to other changes.
