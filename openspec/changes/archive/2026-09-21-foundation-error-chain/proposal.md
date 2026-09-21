# Proposal: Sentinel-based error chain foundation

## Why

The cmd layer's error dispatch (`cmd/error_handler.go`) classifies errors by
substring match, the `domain-typed-errors` spec mandates the same broken
behavior in its requirements, `internal/service/AGENTS.md` documents that
substring pattern as the recommended approach, and a variadic mismatch at
`test/mocks/cmd_mocks.go:222-238` silently swallows assertion failures in
roughly thirty tests. These four artifacts — code, spec, doc, and mock —
are mutually out of sync, and the failure mode (silent misclassification,
false-pass tests, brittle error copy) is observable to end users.

The proposed fix is anchored on `errors.Is` / `errors.As` walks against a
deterministic sentinel catalog and a private `asType[T]` generic helper
that replaces the IIFE pattern. As part of this alignment we also collapse
the existing seven-code CLI exit-code table (0-6) to a three-code dispatch
(0/1/2) per the `golang-cli-architecture` skill's exit-code discipline:
per-resource NotFound categories remain distinguishable through the
formatter's Actionable hints requirement, not through per-resource exit
codes. Scripts and tests keyed on codes 3-6 break by design.

## What Changes

- **`BREAKING`** Introduce `internal/domain/sentinels.go` with twelve sentinel
  errors: four resource-specific NotFound sentinels (`ErrGitRepoNotFound`,
  `ErrWorktreeNotFound`, `ErrProjectNotFound`, `ErrResolutionNotFound`) and
  eight shell sentinels (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`,
  `ErrInvalidShellType`, `ErrShellInferenceFailed`, `ErrShellDetectionFailed`,
  `ErrWrapperGeneration`, `ErrWrapperInstallation`, `ErrConfigFileNotFound`).
  Each message is `domain: <resource> <state>`.
- **`BREAKING`** Replace `IsNotFound() bool` substring matching on six error
  types with an `Is(target error) bool` method that participates in
  `errors.Is` walks against the new sentinels. Delete the
  `strings.Contains` implementations.
- **`BREAKING`** Rewrite `internal/domain/shell_errors.go` as seven concrete
  subtypes (`ShellAlreadyInstalledError`, `ShellNotInstalledError`,
  `ShellInvalidTypeError`, `ShellInferenceError`, `ShellDetectionError`,
  `ShellWrapperError`, `ShellConfigError`) sharing a common base struct, each
  implementing `Unwrap() error` returning `Err error` and `Is(target error)
  bool` matching its sentinel. Drop the unused `ErrConfigFileNotWritable`
  constant and the `Code string` field.
- **`BREAKING`** Rename `Cause error` to `Err error` on all thirteen domain
  error types to align with stdlib convention (e.g., `*os.PathError.Err`,
  `*net.OpError.Err`). Mechanical rename via gopls.
- **`BREAKING`** Reduce CLI exit-code dispatch from seven codes (0-6) to
  three codes (0/1/2). `GetExitCodeForError` returns `ExitCodeUsage` (2)
  for typed cobra usage errors, `ExitCodeError` (1) for any other non-nil
  error, `ExitCodeSuccess` (0) for nil. Per-resource NotFound categories
  remain distinguishable through the formatter's Actionable hints
  requirement. Drop the `ExitCodeConfig`, `ExitCodeGit`,
  `ExitCodeValidation`, `ExitCodeNotFound` constants and their dispatch
  branches. Drop the substring fallback (`strings.Contains(errStr,
  "invalid")`) in `CategorizeError`.
- `BREAKING` Replace `IsCobraArgumentError`'s seven-substring list in
  `cmd/error_handler.go` with a typed walk against `*domain.UsageError`.
  Cobra/pflag flag-parse errors are wrapped at the cmd boundary
  (`cmd/root.go`'s `SetFlagErrorFunc` via `domain.UsageWrap`); cmd-side
  flag-combination failures are constructed directly with
  `domain.NewUsageError`. The previous shape (typed cobra/pflag concrete
  types plus `cmd.ErrFlagUsage`) is replaced by a single
  `errors.As(err, &*domain.UsageError{})` check. Each `cmd/*.go` command
  declares a `cobra.Args:` validator (`ExactArgs`, `MinimumNArgs`,
  `MaximumNArgs`, or `MatchAll`) matching its actual arg shape so the
  substring path is unreachable for syntactic arg-shape failures. Flag-value
  errors are wrapped at the cmd boundary via `domain.UsageWrap` so the
  dispatch becomes a single `errors.As(err, &*domain.UsageError{})` walk.
- Replace the IIFE `func() *T { ... _ = errors.As(...); return target }()`
  pattern in `cmd/error_formatter.go` with a private `asType[T error](err)
  (T, bool)` generic helper, register formatters for the seven shell subtypes,
  and add a hint table that returns resource-specific hints (e.g., `twiggit
  list --all` for `ErrProjectNotFound`).
- Add `test/mocks/helpers.go` with a `variadicArgs(first interface{}, opts
  ...interface{}) []interface{}` helper. Fix
  `test/mocks/cmd_mocks.go:222-238` so variadic options are expanded before
  reaching `m.Called`.
- Update `internal/service/AGENTS.md` and `cmd/AGENTS.md` to use the
  `errors.Is(err, domain.ErrX)` idiom and the `asType[T]` helper
  respectively; delete the previous examples that document the substring bug.
  `cmd/AGENTS.md:58-69` and `AGENTS.md:233-237` drop the rows for exit
  codes 3-6.
- Drop the `💡` emoji from `ValidationError.Error()`. Add `func (e
  *ValidationError) Unwrap() error { return nil }` so the type satisfies the
  existing `domain-typed-errors` requirement that all domain errors
  implement `Unwrap`.
- Fix the `openspec/config.yaml` parse error at lines 134-140 (move the
  `# Purity principles...` comment out of the list item so the YAML parser
  stops ignoring the file on every CLI invocation). Delete the empty
  stranded `openspec/changes/spec-restructure-all-categories/` folder.
- Bump version deferred to the next change that surfaces a release-pinned
  tag; this change deliberately keeps `internal/version/version.go` at
  the build-time-injected `dev` value. End-user CLI surface changes:
  exit codes 3-6 collapse to 1; usage-error categories remain at 2;
  per-resource distinction moves to the formatter hint layer. The
  other breaking surface is the public API of `internal/domain/*Error`
  types and the mock package.

## Capabilities

### New Capabilities

None. The sentinel taxonomy is owned by the existing `domain-typed-errors`
spec; new sentinels and shell subtypes are additions to that capability's
requirement surface.

### Modified Capabilities

- `domain-typed-errors`: the `IsNotFound() bool` substring requirement is
  replaced with a sentinel-walk requirement; the `ShellError` constructor is
  replaced with seven concrete subtypes; the sentinel catalog becomes a
  first-class requirement under the capability's `## Requirements` section;
  the canonical exit-code mapping collapses to three codes (0/1/2).
- `cli-error-formatting`: the dispatch algorithm in
  `GetExitCodeForError` is reformulated around the new
  `*domain.UsageError` type (with `domain.ErrUsageFlag` sentinel) →
  `ExitCodeUsage` (2) and a default `ExitCodeError` (1); a per-resource
  hint table is added; the IIFE formatter pattern is replaced with the
  `asType[T]` helper and caller formatters now handle nil safely.
- `cli-main-entry-point`: the "Specific exit code honored" scenario in
  the Exit code propagation requirement references `ExitCodeValidation (5)`,
  which is removed under the new 3-code dispatch. The scenario body is
  updated to assert that the binary propagates the exit code returned by
  `GetExitCodeForError`, which for `*domain.ValidationError` is now
  `ExitCodeError` (1), and the requirement body clarifies that the entry
  point does not override non-zero codes to 1.

## Non-goals

- Dead config field removal (`CacheEnabled`, `ConcurrentOps`, `MaxConcurrent`,
  `Shell.Timeout`, etc.). Deferred to the next quality change.
- Layer inversion (`internal/service` depending only on `internal/application`).
  Deferred to a separate architecture change.
- Modernization sweep (`for i := range N`, `strings.CutPrefix`,
  `errors.Is(err, os.ErrNotExist)`, `wg.Go`, `t.Context()`, etc.). Deferred.
- `os.Root` migration, slog migration, `internal/infrastructure/cli_client.go`
  nil-guard fix (six sites), `cmd/version.go Run` → `RunE`, removing the
  unused `sync.Mutex` on `worktreeService`. All deferred to follow-up changes.
- Generalization of the IsNotFound string-equality check into a generic helper
  across the error taxonomy beyond what `Is(target)` on each concrete type
  already provides.
- Granular exit codes (3-6). Brought into the change's scope as a Goal;
  reverse-out of any prior change that reintroduced them.
- Reverting the public API break via deprecation aliases. The rename and
  `IsNotFound` removal land hard per the project rule (no deprecation
  aliases in releases); the version tag that ships the break is owned by
  the next change that surfaces a release-pinned tag.

## Impact

| Layer | Files | Lines |
|---|---|---|
| `internal/domain/` | `sentinels.go` (new), `errors.go`, `service_errors.go`, `shell_errors.go`, `usage_error.go` (new) | ~410 |
| `internal/service/` | `shell_service.go` | ~15 |
| `internal/application/` | unchanged | — |
| `cmd/` | `error_handler.go`, `error_formatter.go`, `delete.go` + every `cmd/*.go` Args-validator addition | ~200 |
| `test/mocks/` | `helpers.go` (new), `cmd_mocks.go` | ~25 |
| Tests (after impl, per project rule) | `domain/errors_test.go`, `domain/service_errors_test.go`, `domain/shell_errors_test.go`, `cmd/error_handler_test.go`, `cmd/error_formatter_test.go`, `test/e2e/list_test.go` | rewritten |
| Specs | `openspec/specs/domain-typed-errors/spec.md`, `openspec/specs/cli-error-formatting/spec.md`, `openspec/specs/cli-main-entry-point/spec.md` (delta to be created via `/osc-continue`) | rewritten via deltas |
| AGENTS.md | `internal/service/AGENTS.md`, `cmd/AGENTS.md`, root `AGENTS.md` (troubleshooting table) | ~3 sections |
| Setup | `openspec/config.yaml` (1-line move), delete empty `spec-restructure-all-categories/` | — |

| Aspect | Effect |
|---|---|
| End-user CLI | Exit codes collapse from 0-6 to 0-2. Non-usage failures exit 1 (was 1, 3, 4, 5, or 6 depending on category). Usage errors exit 2 (unchanged). Per-resource NotFound distinction is preserved in the formatter's hint layer. Scripts and CI pipes keyed on `$? -eq 5` etc. break by design. |
| Public API of `domain/*` types | Breaks: `IsNotFound()` removed on six types, `Cause` renamed to `Err`, `ShellError` split into seven subtypes, `ExitCodeConfig`/`ExitCodeGit`/`ExitCodeValidation`/`ExitCodeNotFound` constants removed. **Addition:** `domain.UsageError` type with `NewUsageError(message, err)` and `UsageWrap(err)` constructors, plus `domain.ErrUsageFlag` sentinel (new 13th sentinel in the catalog). |
| Tests | Existing assertion-on-substring tests must be rewritten (drop `IsNotFound` cases, add sentinel-walk cases; collapse per-exit-code assertions to the three-code table). |
| Version | Deferred — build-time `dev` until the next release-pinned tag. |
