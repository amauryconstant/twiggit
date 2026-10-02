# Tasks

## 0. Prerequisites (precondition only — no work performed)

- [ ] 0.1 Confirm changes A, B, C, D have been archived to `main`
  (verify by `openspec list --json` returning empty for changes A, B, C,
  D, i.e. all four present in `openspec/changes/archive/`); halt and
  surface to user if any are not yet archived. Verify `mise run verify`
  on `main` is green before starting group 4.
- [ ] 0.2 Run `go test -cover ./internal/core/... ./internal/git/... ./internal/cmdutil/...`
  and verify the report shows ≥80% line coverage on each of the three
  packages. If below, characterize-test the under-covered branches
  before starting group 4.
- [ ] 0.3 Run `golangci-lint run --enable-only modernize ./...` and
  capture the baseline output. Store in `tmp/modernize-baseline.txt`
  for later comparison after each modernization commit.
- [ ] 0.4 Run `go mod tidy && git diff --exit-code` and verify no
  `godebug` lines in `go.mod` pin removed keys (`asynctimerchan`,
  `tlsunsafeekm`, `tlsrsakex`, `tls3des`, `tls10server`,
  `x509keypairleaf`, `gotypesalias`).
- [ ] 0.5 Run `git show main:openspec/specs/core-git/spec.md | grep
  -c WorktreeWriter` and verify the count is ≥ 1 (existing follow-up
  note). If 0, manually add a one-line follow-up to the existing
  `core-git` spec before this change's delta applies.
- [ ] 0.6 Install `gopatch` per design Decision 12 (pin in
  `.mise/config.toml` `tools` to match the project's pin-binary-tools
  rule); emit one rewrite rule per mechanical transform named in
  Decision 12 (Get-drop, each per-type rename in group 4, bool
  prefix in group 3, modern idioms in group 2), each rule carrying
  a `// +gopatch` golden-test marker pointing at a fixture in
  `internal/cmdutil/factory_test.go` or `test/golden/`. Verify each
  rule's golden test passes against a representative input before
  any rule ships in a commit.

## 1. Validation cleanup (leaf code)

- [ ] 1.1 Collapse `validateBranchErr`, `validateProjectErr`, `validateShellErr` into a single `validateFieldErr(field, value, msg, sug)` constructor in `internal/core/validation.go`. The constructor MUST return `*core.ValidationError` (unwrapped, per the project rule that keeps `core.ValidationError` unwrapped). Verify `go build ./...` compiles, `go test ./internal/core/...` passes, and `grep -nE 'validateFieldErr.*ValidationError' internal/core/validation_test.go` returns a match.
- [ ] 1.2 Promote `reservedNames` map literal to a package-level `var` in `internal/core/validation.go` and verify `go test ./internal/core/...` still passes (no behavioural change; addresses REVIEW finding on per-call re-allocation).

## 2. Mechanical modernization + documentation polish

- [ ] 2.1 Replace `sort.Strings` with `slices.Sort` at the three sites in `cmd/suggestions.go` using the `gopatch` rule from task 0.6; confirm `go build ./...` compiles, `cmd/*_test.go` passes, and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt` (captured in task 0.3).
- [ ] 2.2 Replace `wg.Add(1) + defer wg.Done()` patterns with `wg.Go(...)` at the seven sites in `test/concurrent/concurrent_test.go` using the `gopatch` rule from task 0.6; confirm `go test -race ./test/concurrent/...` passes and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [ ] 2.3 Replace the two `strings.ReplaceAll` chains in `internal/core/shell_wrapper.go` and `test/e2e/helpers/test_id_generator.go` with a single `strings.NewReplacer(...).Replace(...)` call per site using the `gopatch` rule from task 0.6; confirm the respective tests still produce byte-identical output and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [ ] 2.4 Replace the hand-rolled `intToStr` function in `internal/core/errors_legacy.go` with `strconv.Itoa` and verify `go build ./...` compiles (output identical, no behaviour change) and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [ ] 2.5 Add godoc comments to `core.ValidationError.Error()`, `core.NotFoundError.Error()`, `core.OperationError.Error()`, `core.UsageError.Error()`, `ContextResolver.ResolveIdentifier`, and `ContextResolver.ResolutionSuggestions` and verify `go doc ./internal/core/... ./internal/git/...` renders the new prose.
- [ ] 2.6 Remove the four duplicate `// Package git/core` comments from files that already have a canonical `doc.go` and verify `go vet ./...` reports no warning.
- [ ] 2.7 Replace `context.Background()` calls in `*_test.go` files with
  `t.Context()` at every site where the test would benefit from
  automatic test-lifecycle cancellation. Verify `git grep -nE
  'context\.Background\(\)' -- '*_test.go' | wc -l` is strictly less
  than the count before this task (some test sites legitimately use
  `context.Background()` for non-test-scoped goroutines), and
  `golangci-lint run --enable-only modernize ./...` introduces no new
  findings vs `tmp/modernize-baseline.txt`.

## 3. Bool field prefix normalization

- [ ] 3.1 Rename `colorEnabled` → `isColorEnabled`, `existingOnly` → `isExistingOnly`, `valid` → `isValid`, `cacheEnabled` → `isCacheEnabled`, and `ProgressReporter.quiet` → `isQuiet` (with corresponding getter on the struct if exported callers depend on it) and verify `go build ./...` compiles and the affected `_test.go` files pass.
- [ ] 3.2 Run `grep -rE '\b(colorEnabled|existingOnly|cacheEnabled)\b' internal/ cmd/ --include '*.go'` after the rename and verify the only hits are the new `is*` form.

## 4. Type rename (atomic alias-driven commits)

This group splits the type renames into atomic commits driven by
the `type Old = New` alias strategy (design Decision 11). The
`Get*` method names stay unchanged so `*reader` and `*cliClient`
continue to satisfy the role interfaces; groups 5 and 6 drop the
`Get` prefix and remove the Factory per-role fields.

### Sign-off gate 1

Before starting any task in this group, the implementer SHALL
present: (a) the full type-rename list per commit, (b) the
alias-strategy sequencing for the Branch collision, (c) the
role-interface return-type update scope, (d) the new compile-time
data-type fixture in `internal/core/`. The user explicitly
approves the plan before 4.1 begins.

- [ ] 4.1 `GitBranch` → `Branch` via type alias. Add `type GitBranch =
  Branch` to `internal/core/git_types.go`. Update every `GitBranch`
  reference in `internal/git/`, `cmd/`, `internal/cmdutil/`, and
  tests to `Branch` using `gopls rename`. Drop the alias. Verify
  `go build ./...` compiles, `go test ./internal/core/... ./internal/git/... ./cmd/...` passes, and `git grep -nE '\bGitBranch\b' -- '*.go' | wc -l` returns 0.
- [ ] 4.2 `BranchInfo` → `Branch` via type alias. Add `type
  BranchInfo = Branch` to `internal/core/git_types.go`. Update
  every `BranchInfo` reference in `internal/git/`, `cmd/`,
  `internal/cmdutil/`, and tests (including golden files in
  `test/golden/` and integration fixtures) to `Branch`. Drop the
  alias. Verify `go build ./...` compiles, `go test ./...` passes,
  and `git grep -nE '\bBranchInfo\b' -- '*.go' '*.json' '*.yaml' | wc -l` returns 0.
- [ ] 4.3a `core.GitRepository` → `core.Repository` rename. Apply the
  `gopatch` rule from task 0.6, update the role-interface return-type
  signatures in `internal/core/git.go` to reference `core.Repository`,
  and verify `go build ./...` compiles, `go test ./internal/core/...
  ./internal/git/...` passes, and `git grep -nE '\bGitRepository\b'
  -- '*.go' '*.json' '*.tmpl' '*.yaml'` returns 0 production-code
  matches.
- [ ] 4.3b `core.GitDir` → `core.RepoDir` rename. Apply the `gopatch`
  rule from task 0.6 and verify `go build ./...` compiles, `go test
  ./internal/core/... ./internal/git/...` passes, and `git grep -nE
  '\bGitDir\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'` returns 0.
- [ ] 4.3c `core.GitCommit` → `core.Commit` rename. Apply the
  `gopatch` rule from task 0.6 and verify `go build ./...` compiles,
  `go test ./internal/core/... ./internal/git/...` passes, and `git
  grep -nE '\bGitCommit\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'`
  returns 0.
- [ ] 4.3d `core.WorktreeInfo` → `core.Worktree` rename. Apply the
  `gopatch` rule from task 0.6 and verify `go build ./...` compiles,
  `go test ./internal/core/... ./internal/git/...` passes, and `git
  grep -nE '\bWorktreeInfo\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'`
  returns 0.
- [ ] 4.3e `core.RemoteInfo` → `core.Remote` rename. Apply the
  `gopatch` rule from task 0.6 and verify `go build ./...` compiles,
  `go test ./internal/core/... ./internal/git/...` passes, and `git
  grep -nE '\bRemoteInfo\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'`
  returns 0.
- [ ] 4.3f `core.CommitInfo` → `core.Commit` rename. Apply the
  `gopatch` rule from task 0.6 and verify `go build ./...` compiles,
  `go test ./internal/core/... ./internal/git/...` passes, and `git
  grep -nE '\bCommitInfo\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'`
  returns 0.
- [ ] 4.4 Add the compile-time data-type fixture required by the
  new `core-types` requirement: a `TestDataTypeFixture` function
  in `internal/core/data_type_fixture_test.go` with one named
  subtest per renamed data type
  (`TestDataTypeFixture/Repository`, `/Branch`, `/Worktree`,
  `/Commit`, `/Remote`, `/RepoDir`), each subtest exercising the
  type's zero value or constructor so a removal or rename without
  fixture update surfaces as a compile error local to the
  failing subtest. Verify `go test ./internal/core/... -run
  TestDataTypeFixture` passes and `go vet ./...` reports no
  warning.
- [ ] 4.5 Final cross-type-rename grep. Run
  `git grep -nE '\b(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'` and verify zero production-code matches (golden files and tag values may legitimately reference old names; investigate each).

## 5. Method rename (commit 2: Get drop)

This group commits the `Get` prefix removal. The composite `*reader`
and `*cliClient` methods, the role-interface method names, and every
call site move together so each step stays green.

- [ ] 5.1 Drop the `Get` prefix on `GetConfig`, `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo`, `GetResolutionSuggestions`, and the three `GetShellType` methods using `gopls rename`. Update the role-interface method declarations in `internal/core/git.go` in the same commit. The rename scope MUST include at minimum: `cmd/delete.go`, `cmd/prune.go`, `cmd/create.go` (call sites of `GetRepositoryStatus` / `GetRepositoryInfo`), `internal/core/git.go` (role-interface signatures), `internal/git/reader.go` (return types), `internal/git/client.go`, `internal/git/client_test.go`, `internal/cmdutil/factory.go`, `internal/cmdutil/factory_test.go`, and `test/mocks/` (any mock that hard-codes the old method names). Verify `go build ./...` compiles, `go test ./...` passes, `git grep -nE '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)\b' -- '*.go'` returns zero production-code matches, and `git grep -nE '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)\b' -- '*.json' '*.tmpl' '*.yaml'` returns zero tag/reflection matches.
- [ ] 5.2 Verify the six `var _ core.Role = (*git.Client)(nil)` declarations in `internal/git/client.go` use the renamed methods (no code change required if gopls rename propagated) and verify `go build ./internal/git/...` reports no compile-time role-satisfaction failure.

## 6. Factory collapse (commit 3: per-role field removal)

This group deletes the six speculative per-role lazy fields and updates
callers to type-assert on the composite `*git.Client`. Each call site
gains one identifier (`var <role> core.<Role> = client`) and zero new
lazy-cache surface.

### Sign-off gate 2

Before starting 6.1, the implementer SHALL present: (a) the new
caller pattern (`client, err := f.GitClient(); if err != nil { return
err }; var br core.BranchReader = client`) at each of the 5 call
sites in `cmd/`, (b) the updated `Factory.Init()` shape (no role
touch calls, just `Config`/`GitClient`/`Logger`), (c) the updated
`internal/cmdutil/factory_test.go` showing the specific role
interface form. The user explicitly approves the plan before 6.1
begins.

- [ ] 6.1 Delete the six per-role lazy fields (`RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) from `cmdutil.Factory` and update `Factory.Init()` to drop the six touch calls. Update `internal/cmdutil/factory_test.go` to use the specific role interface form (e.g., `core.BranchReader(client)`), NOT the `core.Role(client)` form. Verify `go build ./...` compiles and `go test ./internal/cmdutil/... ./cmd/...` passes.

## 7. Final cross-check (commit 4: rename sweep complete)

- [ ] 7.1 Run `git grep -nE '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)\b'` and `git grep -nE '\b(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)\b'` and verify zero matches in production code (`internal/`, `cmd/`, `test/`).

## 8. Error file rename

- [ ] 8.1 Rename `internal/core/errors_legacy.go` → `internal/core/errors_demoted.go` (matches the in-file purpose comment). All 5 demoted constructors (`NewGitRepositoryError`, `NewGitWorktreeError`, `NewGitCommandError`, `NewConfigError`, `NewContextDetectionError`) MUST keep their existing return types (`*core.OperationError` per the project pattern) — the rename is filename-only, the constructor signatures stay. Verify `go build ./...` compiles, the 5 constructors remain reachable from their existing callers in `internal/output/shell_infra.go`, `internal/git/shell_detect.go`, and `cmd/create.go`, and `go test ./internal/core/... ./internal/output/... ./internal/git/... ./cmd/...` passes.

## 9. Command setup helper extraction

- [ ] 9.1 Extract the 5-step `Config → GitClient → ContextDetector → filepath.Abs → DetectContext` setup duplicated across `runCreate`, `runDelete`, `runList`, `runCd`, `runPrune` into a new `cmd/setup.go` helper `detectContext(f)` returning `(*core.Context, *git.Client, error)` and verify `go build ./...` compiles and every `cmd/*_test.go` still passes.
- [ ] 9.2 Run `grep -nE 'f\.GitClient\(\)|NewContextDetector|DetectContext' cmd/*.go` and verify each file has exactly one call site to the new helper.

## 10. Identical-twins resolution delete

- [ ] 10.1 Write a characterization test in
  `internal/git/context_resolver_test.go` (or a new
  `context_resolver_identical_twins_test.go`) asserting that
  `resolveFromProjectContext(ctx, CoreContextProject)` and
  `resolveFromProjectContext(ctx, CoreContextWorktree)` produce
  identical `*core.Result` values (or both return the same
  `(result, error)` pair) for a representative input set. Verify
  the test passes against `main` BEFORE this group deletes the
  `resolveFromWorktreeContext` function — proves the deletion is
  behavior-preserving. (`golang-refactoring` + `golang-safety` rule:
  load safety review when changing code logic, not just shape.)
- [ ] 10.2 Delete `resolveFromWorktreeContext` from
  `internal/git/context_resolver.go` and route both
  `core.ContextProject` and `core.ContextWorktree` cases in
  `ResolveIdentifier` through `resolveFromProjectContext` and
  verify `go test ./internal/git/...` passes (the characterization
  test from 10.1 must continue to pass without modification).
- [ ] 10.3 Run `grep -nE 'resolveFromWorktreeContext' internal/git/*.go` and verify zero production-code matches remain.

> Group 10 was renumbered: the characterization test inserted
> before the delete; existing 10.1 became 10.2 and existing 10.2
> became 10.3.

## 11. Context resolver file split

- [ ] 11.1 Move `addMainSuggestion`, `addWorktreeSuggestions`, `addBranchSuggestions`, `addProjectSuggestions`, `getProjectContextSuggestions`, `getWorktreeContextSuggestions`, `getOutsideGitContextSuggestions` into a new `internal/git/context_resolver_suggest.go` and verify `go build ./...` compiles.
- [ ] 11.2 Drop the dead `_ *suggestionConfig` parameter from `addBranchSuggestions` (already unused) and verify the suggestion tests in `internal/git/` still cover the same scenarios.
- [ ] 11.3 Add a `ctx context.Context` first-parameter to the three
  `get*ContextSuggestions` methods and replace every
  `context.Background()` site in `context_resolver.go` and
  `context_resolver_suggest.go` with the caller's `ctx` (run
  `git grep -nE 'context\.Background\(\)' internal/git/context_resolver*.go`
  to confirm the exact site count; if not 4, update design
  Decision 10 to match). Add a `ctx context.Context`
  first-parameter to any helper in `resolver.go` currently called
  from a `get*ContextSuggestions` method (`worktreeExists`,
  `validatePathUnder`, `parseCrossProjectReference`,
  `matchesExclusionPatterns`, etc.) and replace every
  `context.Background()` site in those helpers with the new
  parameter, so the cancellation chain reaches the leaf calls.
  Verify `go test -race ./internal/git/...` passes
  (cancellation propagates through suggestion building).
- [ ] 11.4 Reorder the file so `ContextResolver` (the public type alias) appears before the private `contextResolver` struct and verify `gopls references` shows the alias still resolves callers correctly.

## 12. Test helpers package split (single atomic PR)

This group lands in a single atomic PR. No intermediate broken
state: the move and the caller-import updates happen together.
The `helpers_test.go` disposition is move-whole to
`test/helpers_test/` with package `helpers_test`; per-package
dissolve is deferred to a follow-up.

### Sign-off gate 3

Before starting 12.1, the implementer SHALL present: (a) the
6 package boundary decisions (which helpers go in which package),
(b) the 3 caller import updates, (c) the `helpers_test.go`
disposition (move-whole to `test/helpers_test/` with package
`helpers_test`; per-package dissolve deferred to a follow-up).
The user explicitly approves the plan before 12.1 begins.

- [ ] 12.1 Create `test/worktree/`, `test/shell/`, `test/git/`,
  `test/repo/`, `test/golden/`, `test/perf/` directories. Move the
  corresponding source files (`worktree.go`, `shell.go`, `git.go`,
  `repo.go`, `golden.go`, `performance.go`) and
  `worktree_coverage_test.go` into them with their package
  declarations updated to match the directory name. Move
  `test/helpers/helpers_test.go` to
  `test/helpers_test/helpers_test.go` with package `helpers_test`.
  Update every import of `twiggit/test/helpers` in
  `cmd/cd_test.go`, `test/e2e/infrastructure_verification_test.go`,
  `test/e2e/fixtures/e2e_fixtures.go`, and any other test file to
  the new content-named package path (use `gopls rename` on the
  import path). Verify `go build ./...` compiles, `go test ./...`
  passes, and `golangci-lint run ./...` reports no warning.
- [ ] 12.2 Delete the now-empty `test/helpers/` directory and
  verify `ls test/helpers/` returns "No such file or directory";
  `ls test/{worktree,shell,git,repo,golden,perf}/` shows all six
  directories with their respective source files.
- [ ] 12.3 Run `git grep -nE 'twiggit/test/helpers'` and verify
  zero production-code or test-code matches remain (only the
  archived spec prose in `openspec/changes/archive/` may still
  reference the old path).
- [ ] 12.4 Add package content to each new `test/<domain>/`
  package per the project's `golang-documentation` rule (package
  comment MUST exist): one-line `// Package <name> ...` at the top
  of the package's primary source file for `test/worktree`,
  `test/shell`, `test/git`, `test/repo`, `test/golden`,
  `test/perf`, and `test/helpers_test`. Add a smoke test in
  `test/package_boundary_test.go` (build-tag-free, runs in the
  default suite) that asserts: no directory under `test/` matches
  the forbidden names `helpers`, `util`, `common`, `misc`, or
  `support`; the package name in each new `test/<domain>/` package
  matches its directory name; and a representative helper from
  each new package is importable. Verify `go test ./test/...`
  passes and `golangci-lint run ./...` reports no warning.

## 13. cmd/util.go split

- [ ] 13.1 Move `ProgressReporter`, `NewProgressReporter`, `Report`, `ReportProgress` into a new `cmd/progress.go` and verify `go build ./...` compiles.
- [ ] 13.2 Move `ignoreWriter` and `writeOrIgnore` into a new `cmd/writer.go` and verify `go build ./...` compiles.
- [ ] 13.3 Leave `verbosef` and `wrapArgsValidator` in `cmd/util.go` (now ~30 LOC) and verify every `cmd/*.go` import continues to resolve correctly (no import path changes required because all three files share the `cmd` package).

## 14. Spec sync and archival

- [ ] 14.1 Run `openspec sync specs --change naming-refactor-modernize` and verify the four delta files (`core-git`, `core-types`, `cli-factory`, `testing-helpers`) merge cleanly into the canonical `openspec/specs/<id>/spec.md` files.
- [ ] 14.2 Run `openspec validate --strict` against the synced specs and verify zero violations.
- [ ] 14.3 Run `openspec archive --change naming-refactor-modernize` and verify the change directory moves to `openspec/changes/archive/` and `CHANGELOG.md` gains the entry after `osx-generate-changelog` runs.

## 15. Final verification

- [ ] 15.1 Run `golangci-lint fmt ./...` then `mise run verify` (format + lint:fix + lint:gated + vuln:check + test + build) and verify zero errors and zero warnings.
- [ ] 15.2 Run `go test -race ./...` and verify zero data-race reports.
- [ ] 15.3 Run `git grep -nE '\b(Get(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)|(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)|errors_legacy\.go|test/helpers/)\b' -- '*.go' '*.json' '*.tmpl' '*.yaml'` and verify the only matches are in `openspec/changes/archive/` (historical change artifacts) and the new spec prose acknowledging the rename.

## 16. Follow-ups (no implementation here)

- [ ] 16.1 Split `WorktreeWriter` into a 1-3 method shape (e.g.,
  `worktree-creator` + `worktree-lifecycle-reader`) in a future OpenSpec
  change. This change keeps the 4-method shape (`CreateWorktree`,
  `DeleteWorktree`, `ListWorktrees`, `PruneWorktrees`) to limit blast
  radius. The 1-3 method recommendation comes from the
  `golang-structs-interfaces` skill; the existing `core-git` spec
  already flags the deferred split as a follow-up. Tracked in a future
  proposal; no implementation in this change.