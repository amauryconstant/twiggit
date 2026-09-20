# Tasks

## 1. Setup

- [ ] 1.1 Capture pre-change baseline by running `mise run test` and recording that the suite is green before any code edits
- [ ] 1.2 Fix the `openspec/config.yaml` parse error at lines 134-140 by moving the `# Purity principles...` and `# Each rule is identified by...` comment lines out of the `specs:` list item to immediately above it (or removing them), then verify by running `openspec list --json` and confirming no `Warning: could not parse` line is printed
- [ ] 1.3 Delete the empty stranded `openspec/changes/spec-restructure-all-categories/` folder with `git rm -r` and verify by running `ls openspec/changes/` and confirming only `foundation-error-chain/`, `archive/`, and `README.md` remain

## 2. Domain layer: sentinels and first slice

- [ ] 2.1 Create `internal/domain/sentinels.go` with the twelve `errors.New(...)` sentinel variables listed in the `domain-typed-errors` spec's Sentinel catalog requirement, verifying via `go build ./internal/domain/...` that the file compiles
- [ ] 2.2 Refactor `WorktreeServiceError` end-to-end: rename `Cause` field to `Err`, update `Unwrap()` body to return `e.Err`, add `Is(target error) bool` returning `target == domain.ErrWorktreeNotFound`, delete the `IsNotFound() bool` method and its `strings.Contains` body; verify by `gopls rename` reporting zero unresolved references and `go vet ./internal/domain/...` clean
- [ ] 2.3 Apply the same `Cause`→`Err` rename + `Is(target) bool` addition to `ProjectServiceError` (matching `ErrProjectNotFound`), `NavigationServiceError` (matching `ErrResolutionNotFound`), and `ResolutionError` (matching `ErrResolutionNotFound`); verify by `go build ./...` reporting zero compile errors
- [ ] 2.4 Apply the same rename + `Is(target) bool` to `GitRepositoryError` (matching `ErrGitRepoNotFound`) and `GitWorktreeError` (matching `ErrWorktreeNotFound`); verify by `go build ./...` clean
- [ ] 2.5 Add `func (e *ValidationError) Unwrap() error { return nil }`, drop the `💡` glyph from `ValidationError.Error()` so the message never contains a `💡` rune; verify by `grep -c '💡' internal/domain/service_errors.go` returning `0`
- [ ] 2.6 Rewrite `internal/domain/shell_errors.go`: introduce a private `shellErrorBase struct { ShellType, Context, Err }`, drop the `Code string` field and the `ErrConfigFileNotWritable` constant, define six concrete subtypes (`ShellAlreadyInstalledError`, `ShellNotInstalledError`, `ShellInvalidTypeError`, `ShellInferenceError`, `ShellDetectionError`, `ShellWrapperError`, `ShellConfigError`) each embedding the base struct and implementing `Is(target error) bool` matching its sentinel plus `Unwrap() error` returning the `Err` field; verify by `go build ./internal/domain/...` clean and by running `openspec validate foundation-error-chain --json` reporting `valid: true`

## 3. Service layer

- [ ] 3.1 Update `internal/service/shell_service.go`: replace the two `errors.As(err, &shellErr) && shellErr.Code == ErrShellAlreadyInstalled/ErrShellNotInstalled` checks with `errors.Is(err, domain.ErrShellAlreadyInstalled)` / `errors.Is(err, domain.ErrShellNotInstalled)`, and replace `domain.NewShellError(domain.ErrInvalidShellType, ...)` constructors with the per-subtype constructor `domain.NewShellInvalidTypeError(...)`; verify by `go build ./internal/service/...` clean and `go vet ./...` clean
- [ ] 3.2 Run `mise run lint:fix` and verify zero diffs remain in non-import order lint output

## 4. cmd layer

- [ ] 4.1 Replace the substring fallback block in `cmd/error_handler.go` `CategorizeError` (the `errStr := strings.ToLower(err.Error())` and the `strings.Contains(errStr, "invalid")` check) with `errors.Is`-based dispatch: an opening `switch { case errors.Is(err, domain.ErrGitRepoNotFound), errors.Is(err, domain.ErrWorktreeNotFound), ... : return ErrorCategoryNotFound }` covering the four NotFound sentinels; verify by `go build ./cmd/...` clean
- [ ] 4.2 Replace the seven-string-substring `IsCobraArgumentError` body in `cmd/error_handler.go` with typed cobra/pflag error checks: `errors.As(err, &*cobra.FlagError{})`, `errors.Is(err, cobra.ErrSubCommandRequired)`, and `errors.As(err, &*pflag.Error{Code: pflag.ErrRequired})` walks; verify by `go build ./cmd/...` clean
- [ ] 4.3 Add the private `asType[T error](err error) (T, bool)` generic helper to `cmd/error_formatter.go`, replacing the four `func() *T { target := &T{}; _ = errors.As(err, &target); return target }()` IIFE blocks in `formatValidationError`, `formatWorktreeError`, `formatProjectError`, and `formatServiceError` with `target, ok := asType[*domain.T](err); if !ok { return "" }; ...`; verify by `go build ./cmd/...` clean
- [ ] 4.4 Register formatters in `NewErrorFormatterWithOptions` for the six shell subtypes plus `GitRepositoryError`, `GitWorktreeError`, `GitCommandError`, `NavigationServiceError`, `ResolutionError`, and `ConflictError`, in the order specified by the `cli-error-formatting` spec's Type-matched dispatch requirement (specific before generic), verifying by `go build ./cmd/...` clean
- [ ] 4.5 Implement the per-resource hint table in `cmd/error_formatter.go`: replace the literal `"Use 'twiggit list' to see available worktrees"` (and the project/worktree split elsewhere in the registry) with a `switch` over `errors.Is(err, sentinel)` returning the four hints from the `cli-error-formatting` spec's Actionable hints requirement; verify by `go build ./cmd/...` clean
- [ ] 4.6 Update `cmd/delete.go` lines 102-104: replace the three `errors.As + IsNotFound()` checks with a single `if errors.Is(err, domain.ErrWorktreeNotFound) { ... }`; verify by `go build ./cmd/...` clean
- [ ] 4.7 Run `go test ./cmd/...` and verify the package compiles and pre-existing tests still pass before the rewrite in task 7

## 5. test/mocks layer

- [ ] 5.1 Create `test/mocks/helpers.go` with `func variadicArgs(first interface{}, opts ...interface{}) []interface{} { args := []interface{}{first}; return append(args, opts...) }`; verify by `go build ./test/mocks/...` clean and `go vet ./test/mocks/...` clean
- [ ] 5.2 Fix `test/mocks/cmd_mocks.go` lines 222-238: in `MockContextService.GetCompletionSuggestions` and `MockContextService.GetCompletionSuggestionsFromContext`, replace `args := m.Called(partial, opts)` (and the three-arg variant) with `args := m.Called(variadicArgs(partial, opts...)...)`; verify by `go build ./test/mocks/...` clean
- [ ] 5.3 Run `mise run lint:fix` and verify zero diffs remain in the mocks package

## 6. AGENTS.md updates

- [ ] 6.1 Update `internal/service/AGENTS.md` to replace the example that uses `worktreeErr.Message == "worktree not found in any project"` with one showing `if errors.Is(err, domain.ErrWorktreeNotFound) { ... }`; verify by `grep -F 'Message ==' internal/service/AGENTS.md` returning no matches
- [ ] 6.2 Update `cmd/AGENTS.md` to replace the IIFE recipe in the "Adding a new error type" section with one describing the `asType[T error](err) (T, bool)` helper; verify by `grep -F 'func() *' cmd/AGENTS.md` returning no matches

## 7. Tests (per project rule: written AFTER implementation)

- [ ] 7.1 Rewrite `internal/domain/errors_test.go`: drop the `TestGitRepositoryError_IsNotFound` and `TestGitWorktreeError_IsNotFound` cases, add `TestErrSentinels_WalkThroughWraps` (table-driven: each NotFound sentinel wrapped through each corresponding error type returns true via `errors.Is`, each non-matching sentinel returns false); verify by `go test ./internal/domain/...` passing
- [ ] 7.2 Rewrite `internal/domain/service_errors_test.go`: drop `TestWorktreeServiceError_IsNotFound`, add table-driven tests covering `errors.Is(err, ErrWorktreeNotFound)`, `errors.Is(err, ErrProjectNotFound)`, `errors.Is(err, ErrResolutionNotFound)` walks for `WorktreeServiceError`, `ProjectServiceError`, `NavigationServiceError`, `ResolutionError`; verify by `go test ./internal/domain/...` passing
- [ ] 7.3 Rewrite `internal/domain/shell_errors_test.go`: assert each subtype's `Is(target)` returns true only for its own sentinel, each subtype's `Unwrap()` returns its `Err` field, and the Sentinel catalog identifiers and messages in `internal/domain/sentinels.go` exactly match the spec table; verify by `go test ./internal/domain/...` passing
- [ ] 7.4 Rewrite `cmd/error_handler_test.go`: drop the substring-based tests, add `TestCategorizeError_NotFoundSentinelsCollapse` (each of the four NotFound sentinels produces `ErrorCategoryNotFound`), `TestCategorizeError_CobraTypedErrors` (typed cobra errors produce `ErrorCategoryCobra`), `TestCategorizeError_NavigationServiceErrorNotFound` (the previously-missing category resolves to `ErrorCategoryNotFound`); verify by `go test ./cmd/...` passing
- [ ] 7.5 Rewrite `cmd/error_formatter_test.go`: drop the IIFE tests, add per-formatter registration coverage including the six new shell subtypes plus the four NotFound sentinels returning the correct hint from the hint table; verify by `go test ./cmd/...` passing
- [ ] 7.6 Run `mise run test` and verify the full suite is green and that the new sentinel-walk assertions across `internal/domain` and `cmd` confirm both `errors.Is(err, ErrXNotFound)` and `errors.As(err, &typedErr)` returns

## 8. Verification and release

- [ ] 8.1 Run `mise run verify` and verify lint, format, race, and golden-file tests all pass
- [ ] 8.2 Run `openspec validate foundation-error-chain --json` and verify `valid: true, issues: []` remains
- [ ] 8.3 Bump `internal/version/version.go` to `0.13.0`, update any version strings in `README.md` and `CHANGELOG.md`, verify with `git grep -F '0.12'` returning only historical references
- [ ] 8.4 Run `openspec-generate-changelog` and verify the new entry appears at the top of `CHANGELOG.md` with the change archived
- [ ] 8.5 Run `openspec-archive-change foundation-error-chain --json` and verify the change moves from `openspec/changes/foundation-error-chain/` to `openspec/changes/archive/<date>-foundation-error-chain/` and that the two spec deltas are merged into `openspec/specs/domain-typed-errors/spec.md` and `openspec/specs/cli-error-formatting/spec.md` respectively
