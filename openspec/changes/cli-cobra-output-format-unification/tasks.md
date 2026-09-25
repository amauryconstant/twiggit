# Tasks

## 1. `internal/output/errors.go` — spec compliance

- [ ] 1.1 Reorder the `FormatError` dispatch arms in `internal/output/errors.go` to `*core.ValidationError` → `*core.NotFoundError` → `*core.OperationError` → `*core.UsageError` and verify by `go build ./internal/output/...` succeeding and `TestFormatError_ValidationWinsOverOperation` (added in 1.7) passing.
- [ ] 1.2 Replace the `errors.As(err, new(*T))` + `errors.As(err, &oe)` double-call at each dispatch arm with `errors.AsType[*core.X](err)` (Go 1.26+, available after foundation change lands) and verify by `go build ./internal/output/...` succeeding with no `errors.As` left at the dispatch site.
- [ ] 1.3 Move `hintFor` from `cmd/error_formatter.go:23-35` into `internal/output/errors.go`, dispatching via `errors.Is(err, core.ErrProjectNotFound)` / `ErrWorktreeNotFound` / `ErrResolutionNotFound` / `ErrGitRepoNotFound` against the existing sentinels in `internal/core/sentinels.go`, and verify by `TestFormatError_NotFoundHints` (added in 1.7) passing for all four sentinels.
- [ ] 1.4 Gate `writeSuggestions` and `formatNotFoundError`'s hint append on `ios != nil && ios.Quiet` so hints are suppressed in quiet mode, and verify by `TestFormatError_QuietStripsHints` (added in 1.7) passing.
- [ ] 1.5 Add `TestFormatError_ValidationWinsOverOperation` to `internal/output/errors_test.go` asserting that an `&core.OperationError{Cause: &core.ValidationError{...}}` chain renders the `ValidationError` shape (field/value context present, no `op=` line), and verify by `go test -run TestFormatError_ValidationWinsOverOperation ./internal/output/...` passing.
- [ ] 1.6 Add `TestFormatError_NotFoundHints` to `internal/output/errors_test.go` as a table-driven test with one row per sentinel asserting the corresponding hint string from the spec table appears in the rendered output, and verify by `go test -run TestFormatError_NotFoundHints ./internal/output/...` passing.
- [ ] 1.7 Add `TestFormatError_QuietStripsHints` to `internal/output/errors_test.go` constructing `iostreams.Test()` with `Quiet: true` and asserting no `hint:` line appears in the rendered output for an error carrying suggestions, and verify by `go test -run TestFormatError_QuietStripsHints ./internal/output/...` passing.
- [ ] 1.8 Add `TestFormatError_UsageErrorFirst` to `internal/output/errors_test.go` asserting that `core.NewUsageError("missing flag", nil)` renders with a `Usage:` prefix, and verify by `go test -run TestFormatError_UsageErrorFirst ./internal/output/...` passing.
- [ ] 1.9 Force-fail: temporarily swap the dispatch order in `FormatError` so `*core.OperationError` is checked first, run `go test ./internal/output/...`, confirm `TestFormatError_ValidationWinsOverOperation` fails, restore the order, confirm green; verify the dispatch-order gate is live.

## 2. `internal/output/formatter.go` — signature change

- [ ] 2.1 Change `NewFormatter` signature in `internal/output/formatter.go` from `(format string) Formatter` to `(format string) (Formatter, error)`; unknown values (including the legacy `text`) return `(nil, core.NewUsageError("..."))`; the empty string returns `(PlainFormatter{}, nil)` to honor the global default; the four known values return `(JSONFormatter{}, nil)` / `(JSONLinesFormatter{}, nil)` / `(TableFormatter{}, nil)` / `(PlainFormatter{}, nil)` and verify by `go build ./internal/output/...` succeeding.
- [ ] 2.2 Add `TestNewFormatter_UnknownReturnsUsageError` to `internal/output/formatter_test.go` asserting `NewFormatter("xml")` returns `(nil, *core.UsageError)` and verify by `go test -run TestNewFormatter_UnknownReturnsUsageError ./internal/output/...` passing.
- [ ] 2.3 Add `TestNewFormatter_TextReturnsUsageError` to `internal/output/formatter_test.go` asserting `NewFormatter("text")` returns `(nil, *core.UsageError)` so the legacy value does not silently alias to plain and verify by `go test -run TestNewFormatter_TextReturnsUsageError ./internal/output/...` passing.
- [ ] 2.4 Add `TestNewFormatter_EmptyReturnsPlain` to `internal/output/formatter_test.go` asserting `NewFormatter("")` returns `(PlainFormatter{}, nil)` (not a UsageError) so the global default is honored and verify by `go test -run TestNewFormatter_EmptyReturnsPlain ./internal/output/...` passing.

## 3. `cmd/list.go` — migrate to `internal/output`

- [ ] 3.1 Replace the `--output` validation block at `cmd/list.go:62-69` to accept only `json|jsonl|table|plain`; reject `text` and any other non-empty value with `core.NewUsageError("invalid output format '%s': must be 'json', 'jsonl', 'table', or 'plain'", out)` and verify by `go build ./cmd/...` succeeding.
- [ ] 3.2 Replace the formatter selection at `cmd/list.go:129-138` with a call to `output.NewFormatter(format)`; handle the returned error by returning it from `runList` (it is already a `*core.UsageError`); verify by `go build ./cmd/...` succeeding.
- [ ] 3.3 Update `displayWorktrees` at `cmd/list.go:233-239` so the JSON / JSONL / Table / Plain paths call `formatter.Write(opts.IO.Out, worktrees)` directly on the `[]*core.WorktreeInfo` slice (no envelope wrapper); the default empty-format path stays bespoke with the existing `fmt.Fprintf` shape ("branch -> path (modified)(detached)") and verify by `go build ./cmd/...` succeeding.
- [ ] 3.4 Force-fail: temporarily wrap the JSON path in `WorktreeListJSON{Worktrees: ...}` again, run `go test -run TestList ./cmd/...` (or any e2e test that consumes `list -o json`), confirm the envelope-removal assertion fails, restore the bare-array write, confirm green; verify the envelope-drop gate is live.

## 4. Delete `cmd/output.go` and `cmd/error_formatter.go`

- [ ] 4.1 Delete `cmd/output.go` (~88 lines) and verify by `go build ./...` succeeding and `grep -r "OutputFormatter\|TextFormatter\b" cmd/` returning no matches outside the delete-target file.
- [ ] 4.2 Delete `cmd/error_formatter.go` (~180 lines) and verify by `go build ./...` succeeding and `grep -r "ErrorFormatter\b\|formatOperationError\b\|formatNotFoundError\b\|formatValidationError\b\|hintFor\b" cmd/` returning no matches outside the delete-target file.
- [ ] 4.3 Delete `cmd/error_formatter_test.go` (~260 lines) after the four unique scenarios have been ported to `internal/output/errors_test.go` in tasks 1.5–1.8 and verify by `go test ./cmd/...` succeeding with no orphan references.

## 5. `cmd/root.go` — Cobra polish

- [ ] 5.1 Add `cmd.CompletionOptions.DisableDefaultCmd = true` in `cmd/root.go` BEFORE the `AddCommand` block; delete the `cmd.Find([]string{"completion"})` + `cmd.RemoveCommand(completionCmd)` block at `cmd/root.go:73-77`; verify by `go build ./cmd/...` succeeding and `twiggit completion bash | head -5` printing the carapace bash snippet.
- [ ] 5.2 Register four `cmd.AddGroup(...)` calls in `cmd/root.go` BEFORE the `AddCommand` block: `"core"` (list/create/delete/prune), `"navigation"` (cd), `"setup"` (init), `"meta"` (version/completion/docgen); verify by `twiggit --help` displaying the four group headers and their respective commands.
- [ ] 5.3 Add a hidden `twiggit docgen <dir>` command in a new file `cmd/docgen.go`: `Hidden: true`, `Args: cobra.ExactArgs(1)`, `RunE` calls `doc.GenMarkdownTree(cmd.Root(), args[0])`; register it in `cmd/root.go` after the visible subcommands; verify by `twiggit docgen /tmp/twiggit-docs && ls /tmp/twiggit-docs/` listing `twiggit.md`, `twiggit_list.md`, etc.
- [ ] 5.4 Add a one-line comment above `carapace.Gen(cmd)` in `cmd/root.go` explaining that it registers the snippet generator used by `newCompletionCommand` and that `CompletionOptions.DisableDefaultCmd = true` prevents cobra from registering its own competing `completion` subcommand; verify by `go vet ./cmd/...` clean.

## 6. Dead code in `cmd/util.go`, `cmd/create.go`, `cmd/completion.go`

- [ ] 6.1 Delete the `ReportProgress` method at `cmd/util.go:52-58` and verify by `go build ./...` succeeding with no callers (confirmed by `grep -r "ReportProgress\b" cmd/ internal/`).
- [ ] 6.2 Delete the `HookRunner` field at `cmd/create.go:30` (and the unused `cmdutil.HookRunner` interface import if no other consumer remains in `cmd/`) and verify by `go build ./...` succeeding.
- [ ] 6.3 Delete the redundant `ValidArgsFunction` at `cmd/completion.go:41-43` (returns `nil, cobra.ShellCompDirectiveNoFileComp`; `Use: <shell>` + `ExactArgs(1)` already constrain the positional to the hardcoded shell list) and verify by `go build ./cmd/...` succeeding and `twiggit completion bash` still emitting the snippet.

## 7. `mise.toml` — docgen task

- [ ] 7.1 Add a `docgen` task to `.mise/config.toml` that runs `go build -o /tmp/twiggit-bin ./` then `/tmp/twiggit-bin docgen docs/` then `git diff --exit-code docs/`; verify by `mise run docgen` succeeding against an empty `docs/` directory and failing (non-zero exit) when a generated file is manually edited before re-running.

## 8. Spec sync and verification

- [ ] 8.1 Sync the `cli-output` spec delta from `openspec/changes/cli-cobra-output-format-unification/specs/cli-output/spec.md` into `openspec/specs/cli-output/spec.md` and verify by `openspec validate cli-output --type spec --strict` passing.
- [ ] 8.2 Run `mise run verify` and confirm format + lint + gopls + vuln + test + build all pass.
- [ ] 8.3 Run `go test ./...` and confirm every test file (unit + integration + concurrent + e2e) passes.
- [ ] 8.4 Run `./twiggit list -o text 2>&1; echo $?` and confirm the output begins with `Usage:` and the exit code is 2.
- [ ] 8.5 Run `./twiggit list -o json | head -c 50` and confirm the output begins with `[` (bare array, no `{"worktrees":` envelope).
- [ ] 8.6 Run `./twiggit --help` and confirm the output contains the four group headers `Core:`, `Navigation:`, `Setup:`, `Meta:` with the respective subcommands under each.
