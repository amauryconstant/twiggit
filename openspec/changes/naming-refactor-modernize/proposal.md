# Proposal

## Why

The codebase accumulated naming drift, type stutter, and dead code surfaces
across 14 REVIEW findings from `golang-naming`, `golang-project-layout`,
`golang-structs-interfaces`, `golang-refactoring`, `golang-modernize`, and
`golang-documentation` surveys. Mechanical refactors and modernizations can
now land because changes A through D in the broader REVIEW remediation
plan are either complete or in-flight as separate OpenSpec changes, leaving
a stable base for the renaming and structural cleanups. Skills take
precedence over existing specs: where a current spec mandates a name the
skill rejects, the spec is updated, not the code.

## What Changes

- **BREAKING** Drop the `Get` prefix from 8 exported methods:
  `GetConfig`, `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo`,
  `GetResolutionSuggestions`, `GetShellType` (×3). Methods become
  `Config`, `RepositoryStatus`, `Repository`, `Commit`,
  `ResolutionSuggestions`, `ShellType`.
  **Migration**: rename every call site with `gopls rename`; update
  tests and the role-interface signatures in `internal/core/git.go`;
  `git grep` for the old `Get*` names must return zero production-code
  matches.
- **BREAKING** Rename `core` package types to drop `Git*` stutter and
  `*Info` suffix: `GitRepository` → `Repository`, `GitDir` → `RepoDir`,
  `GitCommit` → `Commit`, `GitBranch` → `Branch`, `BranchInfo` →
  `Branch`, `WorktreeInfo` → `Worktree`, `RemoteInfo` → `Remote`,
  `CommitInfo` → `Commit`. `RepositoryStatus` keeps its name; only the
  `Get` prefix drops.
  **Migration**: `gopls rename` per file; update `internal/core/git.go`
  role signatures and `internal/git/reader.go` return types in the
  same commit set; mocks and golden files update atomically; `git grep`
  for the old type names must return zero production-code matches.
- **BREAKING** Remove the 6 per-role lazy fields on `cmdutil.Factory`
  (`RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`,
  `WorktreeWriter`, `BranchWriter`). Callers needing role-narrowed
  access assign `*git.Client` (returned by `Factory.GitClient()`) to a
  local role-typed variable. `*git.Client` satisfies every role through
  embedded promotion.
  **Migration**: replace every `f.<Role>()` call with
  `client, err := f.GitClient(); if err != nil { return err }; var <role> core.<Role> = client`.
  Errors propagate per the single-handling rule (no `_` discard).
  Update `Factory.Init()` to drop the six per-role touch calls and
  `internal/cmdutil/factory_test.go` to use the specific role interface
  form (e.g., `core.BranchReader(client)`).
- **BREAKING** Split the `test/helpers` package into content-named
  packages: `test/worktree`, `test/shell`, `test/git`, `test/repo`,
  `test/golden`, `test/perf`. Three production callers
  (`cmd/cd_test.go`, `test/e2e/infrastructure_verification_test.go`,
  `test/e2e/fixtures/e2e_fixtures.go`) update their imports.
  **Migration**: move the six source files plus
  `worktree_coverage_test.go` into the new content-named packages;
  dissolve or relocate `helpers_test.go`; update the three production
  callers plus any test-file importers.
- Split `internal/git/context_resolver.go` (641 LOC) into
  `context_resolver.go` + `context_resolver_suggest.go`. Resolver
  functions stay in the parent file; suggestion builders
  (`addMainSuggestion`, `addWorktreeSuggestions`, `addBranchSuggestions`,
  `addProjectSuggestions`) move to the suggestion file. The dead
  `existingOnly` parameter on `addBranchSuggestions` is removed.
- Delete the identical-twins `resolveFromWorktreeContext`; route both
  `core.ContextProject` and `core.ContextWorktree` cases through
  `resolveFromProjectContext`.
- Extract the 5-step `Config → GitClient → ContextDetector → filepath.Abs
  → DetectContext` setup duplicated across 5 `cmd/run*` functions into a
  shared helper in `cmd/setup.go` (new file).
- Add `context.Context` first-parameter to the suggestion methods
  (`getProjectContextSuggestions`, `getWorktreeContextSuggestions`,
  `getOutsideGitContextSuggestions`) so the inner `context.Background()`
  calls honour caller cancellation.
- Replace hand-rolled `intToStr` in `errors_legacy.go` with `strconv.Itoa`.
  Rename file to `errors_demoted.go` to match its in-file purpose comment.
- Collapse three near-identical `validateBranchErr` / `validateProjectErr`
  / `validateShellErr` constructors into one parameterised
  `validateFieldErr(field, value, msg, sug)`.
- Promote the `reservedNames` map in `core/validation.go` to a package-level
  `var` so it is allocated once at package init.
- Split `cmd/util.go` (93 LOC, four concerns) into `cmd/progress.go`
  (`ProgressReporter`), `cmd/writer.go` (`ignoreWriter`, `writeOrIgnore`),
  and `cmd/util.go` (retains `verbosef` and `wrapArgsValidator`).
- Rename unexported bool fields to `is/has/can`-prefixed form:
  `colorEnabled` → `isColorEnabled`, `existingOnly` → `isExistingOnly`,
  `valid` → `isValid`, `cacheEnabled` → `isCacheEnabled`,
  `ProgressReporter.quiet` → `isQuiet`.
- Modernize idioms: `sort.Strings` → `slices.Sort` (×3 in
  `cmd/suggestions.go`); `wg.Add(1) + defer wg.Done()` → `wg.Go(...)` (×7
  in `test/concurrent/concurrent_test.go`); `strings.ReplaceAll` chains →
  `strings.NewReplacer` (×2 in `internal/core/shell_wrapper.go` and
  `test/e2e/helpers/test_id_generator.go`).
- Remove 4 duplicate `// Package git/core` comments from files that
  already have a canonical `doc.go`. Add godoc to 6 exported `*.Error()`
  methods in `core/errors.go`, `core/usage_error.go`, and the
  `ContextResolver.ResolveIdentifier` + `ResolutionSuggestions` methods.

## Non-Goals

- Renaming the module path `twiggit` → `github.com/amoconst/twiggit`
  (deferred to its own change).
- Bumping to Go 1.27 for `strings.CutLast` / `reflect.TypeAssert[T]`
  (this change stops at ≥1.26.2).
- Adding fuzz / benchmark / example tests (opportunistic only).
- Removing dead code (`ProgressReporter.ReportProgress`,
  `CreateOptions.HookRunner`, `cmd/error_formatter.go`,
  `cmd/output.go`) — owned by change D in-flight.
- Creating `llms.txt` — explicitly declined.
- Refactoring the `slog.SetDefault` / `Factory.Logger` dual-channel
  log path — owned by change C.
- Changing the CLI command surface (flags, arguments, exit codes,
  output formats preserved).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `core-git` — MODIFIED `### Requirement: RepositoryReader exposes
  GetRepositoryStatus, GetRepositoryInfo, and GetCommitInfo`
  (rename methods to drop the `Get` prefix and update the return-type
  references to the renamed data types from `core-types`).
  REMOVED `### Requirement: Factory per-role fields are additive to
  composite` and `### Requirement: Factory lazy-field errors are
  propagated, not swallowed` (per-role fields no longer exist; the
  second is folded into the broader error pass-through pattern).
- `core-types` — MODIFIED data-type surface to drop `Git*` stutter
  and `*Info` suffix per the rename table. The migration uses type
  aliases (`type GitBranch = Branch`, then `type BranchInfo = Branch`)
  so each commit stays green. ADDS a new requirement distinguishing
  data types (`Repository`, `Branch`, `Worktree`, `Commit`, `Remote`,
  `RepoDir`) from capability interfaces (`RepositoryOpener`,
  `BranchReader`, etc.) — same identifier namespace, different
  semantic category. ADDS a new requirement requiring a compile-time
  fixture in `internal/core/` that exercises every renamed data
  type's constructor or zero-value usage.
- `cli-factory` — MODIFIED `### Requirement: Factory exposes lazy
  function fields` to drop the 6 per-role entries and explicitly
  state that role interfaces are obtained via type assertion on
  `*git.Client` returned by `Factory.GitClient()`. ADDS a new
  requirement for the type-assertion caller pattern and a new
  requirement stating that `Factory.Init()` joins the lazy-field
  construction failures with `errors.Join`.
- `testing-helpers` — REMOVED `### Requirement: Worktree helpers`
  and `### Requirement: Shell helpers` (the package-path reference
  was the anti-pattern; content moves to `test/worktree` and
  `test/shell`). MODIFIED `### Requirement: Automatic resource
  cleanup` and `### Requirement: t.Helper() for error reporting` to
  apply to every new `test/<domain>/` package. ADDS a new
  requirement listing the 6 content-named packages
  (`test/worktree`, `test/shell`, `test/git`, `test/repo`,
  `test/golden`, `test/perf`) and forbidding the `helpers` / `util` /
  `common` / `misc` / `support` package names.

## Impact

**Production code** (core domain types + error contract):
- `internal/core/validation.go` — collapse three `validate*Err`
  constructors into one `validateFieldErr`; promote `reservedNames`
  to a package-level `var`.
- `internal/core/git.go` — role-interface return-type updates
  (`core-types` rename target).
- `internal/core/git_types.go` — `GitRepository` → `Repository`,
  `GitDir` → `RepoDir`, `GitCommit` → `Commit`, `GitBranch` →
  `Branch` (with `type GitBranch = Branch` alias during migration).
- `internal/core/git_repo.go` — `BranchInfo` → `Branch` (with
  `type BranchInfo = Branch` alias during migration), `WorktreeInfo`
  → `Worktree`, `RemoteInfo` → `Remote`, `CommitInfo` → `Commit`.
- `internal/core/errors.go`, `internal/core/usage_error.go` — add
  godoc to `*.Error()` methods.
- `internal/core/errors_legacy.go` → renamed to
  `internal/core/errors_demoted.go`.

**Production code** (git I/O adapter + cmdutil):
- `internal/git/context_resolver.go` — split into
  `context_resolver.go` + `context_resolver_suggest.go`; delete
  `resolveFromWorktreeContext`; add `ctx context.Context` first
  parameter to suggestion methods.
- `internal/git/reader.go` — return-type updates to match `core-types`
  renames.
- `internal/git/client.go` — six `var _ core.Role = (*git.Client)(nil)`
  compile-time fixtures verified after method rename.
- `internal/cmdutil/factory.go` — delete the 6 per-role lazy fields
  (`RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`,
  `WorktreeWriter`, `BranchWriter`); update `Factory.Init()` to drop
  the 6 touch calls.
- `cmd/util.go` — split into `cmd/progress.go` (`ProgressReporter`),
  `cmd/writer.go` (`ignoreWriter`, `writeOrIgnore`), and
  `cmd/util.go` (retains `verbosef` and `wrapArgsValidator`).
- `cmd/setup.go` (new) — extract the 5-step setup duplicated across
  5 `run*` functions.
- `cmd/create.go`, `cmd/delete.go`, `cmd/prune.go`, `cmd/cd.go`,
  `cmd/list.go` — caller updates for the Get-prefix drop and the
  per-role Factory collapse.
- `cmd/suggestions.go` — `sort.Strings` → `slices.Sort`; verify
  `[]string` input type at each call site.

**Tests**:
- `internal/cmdutil/factory_test.go` — use specific role interface
  form (`core.BranchReader(client)`), not the `core.Role` form.
- `internal/git/client_test.go` — return-type updates to match
  `core-types` renames.
- `cmd/cd_test.go` — import path update (now imports
  `test/worktree` and `test/shell`).
- `test/concurrent/concurrent_test.go` — `wg.Add(1) + defer wg.Done()`
  → `wg.Go(...)` at 7 sites.
- `test/e2e/infrastructure_verification_test.go`,
  `test/e2e/fixtures/e2e_fixtures.go` — import path updates.
- `test/e2e/helpers/test_id_generator.go` — `strings.ReplaceAll`
  chain → `strings.NewReplacer`.
- `test/mocks/` — mock method names updated alongside role-interface
  renames.
- `test/golden/` — golden fixtures updated to match renamed types
  and methods.
- `test/integration/` — return-type updates.

**Build tooling / CI**:
- `.golangci.yml` — `nolintlint`, `gocritic`, `nilerr`,
  `paralleltest` enabled (prerequisite — change A).

**Toolchain**:
- Go ≥1.26.2 (prerequisite — change A) for stdlib vuln closure;
  `slices`, `samber/lo`, `strings.NewReplacer`, `sync.WaitGroup.Go`
  already used elsewhere in the codebase.
- Go 1.27 `godebug` cleanup precondition: no `godebug` lines in
  `go.mod` pinning removed keys (`asynctimerchan`, `tlsunsafeekm`,
  `tlsrsakex`, `tls3des`, `tls10server`, `x509keypairleaf`,
  `gotypesalias`).

**Risks and mitigations**:
- **Per-role Factory field removal touches 5 caller sites** in
  `cmd/create.go`, `cmd/prune.go`, `cmd/delete.go`, `cmd/cd.go`,
  `cmd/list.go` → Apply via a `gopatch` rule (design Decision 12);
  verify `go build ./... && go test ./cmd/... ./internal/cmdutil/...`
  after the rename to confirm no site is missed.
- **test/helpers package split breaks 3 importers**
  (`cmd/cd_test.go`, `test/e2e/infrastructure_verification_test.go`,
  `test/e2e/fixtures/e2e_fixtures.go`) and any other test file
  importing `twiggit/test/helpers` → Land as a single atomic PR
  (tasks 12.1-12.4) per design Decision 9 step 11;
  `git grep -nE 'twiggit/test/helpers' -- '*.go'` must return zero
  matches before the commit ships.
- **Type rename blast radius touches ~50 production files** —
  `gopls rename` misses string-literal references in error
  messages, golden fixtures, and spec prose → Each rename commit
  ends with a `git grep` cross-check (tasks 4.5, 7.1, 15.3) for
  the full rename table.
- **Gopatch rules need their own golden tests** before trusting
  them across the rename cluster → Add `internal/cmdutil/factory_test.go`
  and `test/golden/` fixture asserts as golden tests for the
  gopatch-generated edits (the gopatch rule's `// +gopatch` marker
  names the golden test it must satisfy).
- **Modernize strategy gaps** (`slices.Sort`, `wg.Go`,
  `strings.NewReplacer`, `t.Context()`) introduce modernize-lint
  findings only after the linter is enabled (change A prerequisite)
  → Each modernization task in 2.1-2.7 verifies
  `golangci-lint run --enable-only modernize ./...` shows no delta
  vs `tmp/modernize-baseline.txt` captured in task 0.3.

**External surface**:
- No external Go consumers depend on the module (module path is bare
  `twiggit`; the rename to a URL-shaped module path is a separate
  deferred change).
- No CLI command surface changes — flags, arguments, output formats,
  exit codes all preserved.
