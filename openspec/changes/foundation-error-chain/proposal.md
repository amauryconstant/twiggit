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
- **`BREAKING`** Rewrite `internal/domain/shell_errors.go` as six concrete
  subtypes (`ShellAlreadyInstalledError`, `ShellNotInstalledError`,
  `ShellInvalidTypeError`, `ShellInferenceError`, `ShellDetectionError`,
  `ShellWrapperError`, `ShellConfigError`) sharing a common base struct, each
  implementing `Unwrap() error` returning `Cause error` and `Is(target error)
  bool` matching its sentinel. Drop the unused `ErrConfigFileNotWritable`
  constant and the `Code string` field.
- **`BREAKING`** Rename `Cause error` to `Err error` on all thirteen domain
  error types to align with stdlib convention (e.g., `*os.PathError.Err`,
  `*net.OpError.Err`). Mechanical rename via gopls.
- Replace substring-based `CategorizeError` in `cmd/error_handler.go` with a
  sentinel dispatch: `errors.Is(err, Err*NotFound)` collapses to
  `ExitCodeNotFound`; typed `errors.As` dispatches the remaining error
  categories. Drop the `strings.Contains(errStr, "invalid")` fallback and
  replace `IsCobraArgumentError`'s seven-substring list with cobra typed
  errors (`*cobra.FlagError`, `cobra.ErrSubCommandRequired`,
  `pflag.ErrorType*`).
- Replace the IIFE `func() *T { ... _ = errors.As(...); return target }()`
  pattern in `cmd/error_formatter.go` with a private `asType[T error](err)
  (T, bool)` generic helper, register formatters for the six shell subtypes,
  and add a hint table that returns resource-specific hints (e.g., `twiggit
  list --all` for `ErrProjectNotFound`).
- Add `test/mocks/helpers.go` with a `variadicArgs(first interface{}, opts
  ...interface{}) []interface{}` helper. Fix
  `test/mocks/cmd_mocks.go:222-238` so variadic options are expanded before
  reaching `m.Called`.
- Update `internal/service/AGENTS.md` and `cmd/AGENTS.md` to use the
  `errors.Is(err, domain.ErrX)` idiom and the `asType[T]` helper
  respectively; delete the previous examples that document the substring bug.
- Drop the `💡` emoji from `ValidationError.Error()`. Add `func (e
  *ValidationError) Unwrap() error { return nil }` so the type satisfies the
  existing `domain-typed-errors` requirement that all domain errors
  implement `Unwrap`.
- Fix the `openspec/config.yaml` parse error at lines 134-140 (move the
  `# Purity principles...` comment out of the list item so the YAML parser
  stops ignoring the file on every CLI invocation). Delete the empty
  stranded `openspec/changes/spec-restructure-all-categories/` folder.
- Bump version to **0.13.0** (hard break, no deprecation aliases). End-user
  CLI surface is unchanged; the breaking surface is the public API of
  `internal/domain/*Error` types and the mock package.

## Capabilities

### New Capabilities

None. The sentinel taxonomy is owned by the existing `domain-typed-errors`
spec; new sentinels and shell subtypes are additions to that capability's
requirement surface.

### Modified Capabilities

- `domain-typed-errors`: the `IsNotFound() bool` substring requirement is
  replaced with a sentinel-walk requirement; the `ShellError` constructor is
  replaced with six concrete subtypes; the sentinel catalog becomes a
  first-class requirement under the capability's `## Requirements` section.
- `cli-error-formatting`: the dispatch algorithm in
  `GetExitCodeForError` is reformulated around sentinel and typed-error
  matches; a per-resource hint table is added; the IIFE formatter pattern
  is replaced with the `asType[T]` helper and caller formatters now handle
  nil safely.

## Impact

| Layer | Files | Lines |
|---|---|---|
| `internal/domain/` | `sentinels.go` (new), `errors.go`, `service_errors.go`, `shell_errors.go` | ~350 |
| `internal/service/` | `shell_service.go` | ~15 |
| `internal/application/` | unchanged | — |
| `cmd/` | `error_handler.go`, `error_formatter.go`, `delete.go` | ~180 |
| `test/mocks/` | `helpers.go` (new), `cmd_mocks.go` | ~25 |
| Tests (after impl, per project rule) | `domain/errors_test.go`, `domain/service_errors_test.go`, `domain/shell_errors_test.go`, `cmd/error_handler_test.go`, `cmd/error_formatter_test.go` | rewritten |
| Specs | `openspec/specs/domain-typed-errors/spec.md`, `openspec/specs/cli-error-formatting/spec.md` | rewritten via deltas |
| AGENTS.md | `internal/service/AGENTS.md`, `cmd/AGENTS.md` | 2 sections |
| Setup | `openspec/config.yaml` (1-line move), delete empty `spec-restructure-all-categories/` | — |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged (exit codes for the four NotFound categories still collapse to 6). |
| Public API of `domain/*` types | Breaks: `IsNotFound()` removed on six types, `Cause` renamed to `Err`, `ShellError` split into six subtypes. |
| Tests | Existing assertion-on-substring tests must be rewritten (drop `IsNotFound` cases, add sentinel-walk cases). |
| Version | Bump 0.13.0 (hard break). |

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
