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
- **BREAKING** Rename `core` package types to drop `Git*` stutter and
  `*Info` suffix: `GitRepository` → `Repository`, `GitDir` → `RepoDir`,
  `GitCommit` → `Commit`, `GitBranch` → `Branch`, `BranchInfo` →
  `Branch`, `WorktreeInfo` → `Worktree`, `RemoteInfo` → `Remote`,
  `CommitInfo` → `Commit`. `RepositoryStatus` keeps its name; only the
  `Get` prefix drops.
- **BREAKING** Remove the 6 per-role lazy fields on `cmdutil.Factory`
  (`RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`,
  `WorktreeWriter`, `BranchWriter`). Callers needing role-narrowed
  access assign `*git.Client` (returned by `Factory.GitClient()`) to a
  local role-typed variable. `*git.Client` satisfies every role through
  embedded promotion.
- **BREAKING** Split the `test/helpers` package into content-named
  packages: `test/worktree`, `test/shell`, `test/git`, `test/repo`,
  `test/golden`, `test/perf`. Three production callers
  (`cmd/cd_test.go`, `test/e2e/infrastructure_verification_test.go`,
  `test/e2e/fixtures/e2e_fixtures.go`) update their imports.
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

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `core-git`: Remove the requirement that `cmdutil.Factory` exposes
  per-role lazy fields (`RepoOpener`, `BranchReader`, `RepositoryReader`,
  `RemoteReader`, `WorktreeWriter`, `BranchWriter`). Rename the methods
  listed in the `RepositoryReader` role interface requirement
  (`GetRepositoryStatus` → `RepositoryStatus`, `GetRepositoryInfo` →
  `Repository`, `GetCommitInfo` → `Commit`). Update the role interfaces
  to reflect the renamed return types from `core-types`.
- `core-types`: Rename `GitRepository` → `Repository`, `GitDir` → `RepoDir`,
  `GitCommit` → `Commit`, `GitBranch` → `Branch`, `BranchInfo` → `Branch`,
  `WorktreeInfo` → `Worktree`, `RemoteInfo` → `Remote`, `CommitInfo` →
  `Commit`. Add a requirement distinguishing data types
  (`Repository`, `Branch`, `Worktree`) from role interfaces
  (`RepositoryOpener`, `BranchReader`) — same identifier namespace,
  different semantic category.
- `cli-factory`: Update the requirement listing lazy fields to drop the 6
  per-role entries. Add a requirement that role interfaces are obtained
  via type assertion on `*git.Client` returned by `Factory.GitClient()`.
- `testing-helpers`: Rewrite the existing `Worktree helpers` and
  `Shell helpers` requirements to describe content rather than the
  `test/helpers` package path. Add requirements for the new packages
  `test/golden`, `test/perf`, and `test/repo`. The package path is no
  longer mandated by the spec; spec describes what the helpers do, not
  where they live.

## Impact

- Production code in `internal/core/`, `internal/git/`, `internal/config/`,
  `internal/cmdutil/`, `internal/output/`, `cmd/`.
- Tests across `cmd/`, `test/integration/`, `test/e2e/`,
  `test/concurrent/`, `internal/git/`, `internal/core/`,
  `internal/config/`.
- Build tooling: `mise run verify` gate; `golangci-lint` v2 with
  `nolintlint`, `gocritic`, `nilerr`, `paralleltest` enabled
  (prerequisite — change A).
- Toolchain: Go ≥1.26.2 (prerequisite — change A) for stdlib vuln
  closure; `slices`, `samber/lo`, `strings.NewReplacer` already used
  elsewhere in the codebase.
- No external Go consumers depend on the module (module path is bare
  `twiggit`; the rename to a URL-shaped module path is a separate
  deferred change).
- No CLI command surface changes — flags, arguments, output formats,
  exit codes all preserved.
