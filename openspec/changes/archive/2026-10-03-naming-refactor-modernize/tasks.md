# Tasks

## 0. Prerequisites (precondition only — no work performed)

- [x] 0.1 Confirm changes A, B, C, D have been archived to `main`
  (verify by `openspec list --json` returning empty for changes A, B, C,
  D, i.e. all four present in `openspec/changes/archive/`); halt and
  surface to user if any are not yet archived. Verify `mise run verify`
  on `main` is green before starting group 4.
- [x] 0.2 Run `go test -cover ./internal/core/... ./internal/git/... ./internal/cmdutil/...`
  and verify the report shows ≥80% line coverage on each of the three
  packages. If below, characterize-test the under-covered branches
  before starting group 4.
- [x] 0.3 Run `golangci-lint run --enable-only modernize ./...` and
  capture the baseline output. Store in `tmp/modernize-baseline.txt`
  for later comparison after each modernization commit.
- [x] 0.4 Run `go mod tidy && git diff --exit-code` and verify no
  `godebug` lines in `go.mod` pin removed keys (`asynctimerchan`,
  `tlsunsafeekm`, `tlsrsakex`, `tls3des`, `tls10server`,
  `x509keypairleaf`, `gotypesalias`).
- [x] 0.5 Run `git show main:openspec/specs/core-git/spec.md | grep
  -c WorktreeWriter` and verify the count is ≥ 1 (existing follow-up
  note). If 0, manually add a one-line follow-up to the existing
  `core-git` spec before this change's delta applies.
- [x] 0.6 Dropped per user decision: rely on `gopls rename` + hand-edit
  + per-task `git grep` cross-checks. No `gopatch` install, no rule
  emission, no `tools/gopatch/` directory.

## 1. Validation cleanup (leaf code)

- [x] 1.1 Collapse `validateBranchErr`, `validateProjectErr`, `validateShellErr` into a single `validateFieldErr(field, value, msg, sug)` constructor in `internal/core/validation.go`. The constructor MUST return `*core.ValidationError` (unwrapped, per the project rule that keeps `core.ValidationError` unwrapped). Verify `go build ./...` compiles, `go test ./internal/core/...` passes, and `grep -nE 'validateFieldErr.*ValidationError' internal/core/validation_test.go` returns a match.
- [x] 1.2 Promote `reservedNames` map literal to a package-level `var` in `internal/core/validation.go` and verify `go test ./internal/core/...` still passes (no behavioural change; addresses REVIEW finding on per-call re-allocation).

## 2. Mechanical modernization + documentation polish

- [x] 2.1 Replace `sort.Strings` with `slices.Sort` at the three sites in `cmd/suggestions.go`; confirm `go build ./...` compiles, `cmd/*_test.go` passes, and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt` (captured in task 0.3).
- [x] 2.2 Replace `wg.Add(1) + defer wg.Done()` patterns with `wg.Go(...)` at the seven sites in `test/concurrent/concurrent_test.go`; confirm `go test -race ./test/concurrent/...` passes and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [x] 2.3 Replace the two `strings.ReplaceAll` chains in `internal/core/shell_wrapper.go` and `test/e2e/helpers/test_id_generator.go` with a single `strings.NewReplacer(...).Replace(...)` call per site; confirm the respective tests still produce byte-identical output and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [x] 2.4 Replace the hand-rolled `intToStr` function in `internal/core/errors_legacy.go` with `strconv.Itoa` and verify `go build ./...` compiles (output identical, no behaviour change) and `golangci-lint run --enable-only modernize ./...` introduces no new findings vs `tmp/modernize-baseline.txt`.
- [x] 2.5 Add godoc comments to `core.ValidationError.Error()`, `core.NotFoundError.Error()`, `core.OperationError.Error()`, `core.UsageError.Error()`, `ContextResolver.ResolveIdentifier`, and `ContextResolver.GetResolutionSuggestions` and verify `go doc ./internal/core/... ./internal/git/...` renders the new prose. The 4 Error() methods render via `go doc`; ResolveIdentifier/GetResolutionSuggestions live on the unexported contextResolver struct surfaced via the public ContextResolver alias (task 11.4 will reorder types to improve surface visibility).
- [x] 2.6 Remove the four duplicate `// Package git/core` comments from files that already have a canonical `doc.go` and verify `go vet ./...` reports no warning.
- [x] 2.7 Replace `context.Background()` calls in `*_test.go` files with
  `t.Context()` (or `s.T().Context()` for testify suite methods) at every
  site where the test would benefit from automatic test-lifecycle
  cancellation. Count went from 86 → 1 (the remaining 1 is in the
  `.opencode/skills/golang-cli/assets/examples/create_test.go` reference
  asset, not a project test). `golangci-lint run --enable-only modernize
  ./...` still reports 0 issues vs `tmp/modernize-baseline.txt`.

## 3. Bool field prefix normalization

- [x] 3.1 Rename `colorEnabled` → `isColorEnabled`, `existingOnly` → `isExistingOnly`, `valid` → `isValid`, `cacheEnabled` → `isCacheEnabled`, and `ProgressReporter.quiet` → `isQuiet` (with corresponding getter on the struct if exported callers depend on it) and verify `go build ./...` compiles and the affected `_test.go` files pass.
- [x] 3.2 Run `grep -rE '\b(colorEnabled|existingOnly|cacheEnabled)\b' internal/ cmd/ --include '*.go'` after the rename and verify the only hits are the new `is*` form. Zero hits remain.

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

- [x] 4.1 No-op: `core.GitBranch` does not exist in the codebase; only
  `BranchInfo` (now renamed `Branch` in 4.2) was ever present. No
  type alias migration needed. (One historical mention of "GitBranch"
  remains in `internal/git/writer_test.go` as a test function name
  `TestWriteSideFailure_OpIsGitBranch` — that name documents an Op
  string, not the type.)
- [x] 4.2 `BranchInfo` → `Branch` via direct rename. No `GitBranch` to
  alias-collide with. Updated 6 files. `git grep -nE '\bBranchInfo\b'
  -- '*.go' '*.json' '*.yaml'` returns 0.
- [x] 4.3a `core.GitRepository` → `core.Repository` direct rename. 16
  files updated. `git grep -nE '\bGitRepository\b'` returns 0.
- [x] 4.3b `core.GitDir` → `core.RepoDir` direct rename. 3 files
  updated. `git grep -nE '\bGitDir\b'` returns 0.
- [x] 4.3c No-op: `core.GitCommit` does not exist; only `CommitInfo`
  (renamed in 4.3f) was ever present.
- [x] 4.3d `core.WorktreeInfo` → `core.Worktree` direct rename. 16
  files updated. `git grep -nE '\bWorktreeInfo\b'` returns 0.
- [x] 4.3e `core.RemoteInfo` → `core.Remote` direct rename. 7 files
  updated. `git grep -nE '\bRemoteInfo\b'` returns 0.
- [x] 4.3f `core.CommitInfo` → `core.Commit` direct rename. 4 files
  updated. `git grep -nE '\bCommitInfo\b'` returns 0.
- [x] 4.4 Added `internal/core/data_type_fixture_test.go` with
  `TestDataTypeFixture` and 6 named subtests (Repository, Branch,
  Worktree, Commit, Remote, RepoDir). `go test ./internal/core/...
  -run TestDataTypeFixture` passes; `go vet ./...` reports no
  warning.
- [x] 4.5 Final cross-grep: `git grep -nE '\b(GitRepository|GitDir|
  GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)
  \b'` returns 0 production-code matches. Only the test-function name
  `TestWriteSideFailure_OpIsGitBranch` mentions "GitBranch" as an
  identifier suffix (not a type reference).

## 5. Method rename (commit 2: Get drop)

This group commits the `Get` prefix removal. The composite `*reader`
and `*cliClient` methods, the role-interface method names, and every
call site move together so each step stays green.

- [x] 5.1 Drop the `Get` prefix on `GetConfig`, `GetRepositoryStatus`,
  `GetRepositoryInfo`, `GetCommitInfo`, `GetResolutionSuggestions`,
  and the three `GetShellType` methods using `gopls rename`. Update the
  role-interface method declarations in `internal/core/git.go` in the
  same commit. The 3 `GetShellType` methods (on SetupShellRequest,
  ValidateInstallationRequest, GenerateWrapperRequest) hit a
  field/method name collision with the existing exported `ShellType`
  field on each Request type. Per the user's call, resolved by
  unexporting the 3 fields to `shellType` and adding 3 constructors
  (`NewSetupShellRequest`, `NewValidateInstallationRequest`,
  `NewGenerateWrapperRequest`). The `RequestWithShellType` interface
  is updated to require `ShellType() ShellType` instead of
  `GetShellType()`. Renamed 15 files. `git grep -nE
  '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|
  ResolutionSuggestions|Config)\b' -- '*.go' '*.json' '*.tmpl'
  '*.yaml'` returns 0 matches.
- [x] 5.2 The six `var _ core.Role = (*git.Client)(nil)` declarations
  in `internal/git/client.go` are unchanged (they reference the role
  interfaces, not the renamed methods directly). `go build
  ./internal/git/...` reports no compile-time role-satisfaction
  failure; `go vet` and `go test ./internal/git/...` clean.

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

- [x] 6.1 Delete the six per-role lazy fields (`RepoOpener`, `BranchReader`,
  `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`)
  from `cmdutil.Factory` and update `Factory.Init()` to drop the six
  touch calls. Update `internal/cmdutil/factory_test.go` to use the
  specific role interface form (e.g., `core.BranchReader(client)`),
  NOT the `core.Role(client)` form. Per Gate 2 finding: the 5 cmd
  caller sites the design anticipated do not exist — the cmd layer
  already uses `f.GitClient()` directly. Only `factory.go` and
  `factory_test.go` needed changes. `go build ./...` compiles;
  `go test ./...` passes.

## 7. Final cross-check (commit 4: rename sweep complete)

- [x] 7.1 Run `git grep -nE '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)\b'` and `git grep -nE '\b(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)\b'` and verify zero matches in production code (`internal/`, `cmd/`, `test/`). Both greps return zero matches.

## 8. Error file rename

- [x] 8.1 Rename `internal/core/errors_legacy.go` → `internal/core/errors_demoted.go` (matches the in-file purpose comment). All 5 demoted constructors (`NewGitRepositoryError`, `NewGitWorktreeError`, `NewGitCommandError`, `NewConfigError`, `NewContextDetectionError`) MUST keep their existing return types (`*core.OperationError` per the project pattern) — the rename is filename-only, the constructor signatures stay. Verify `go build ./...` compiles, the 5 constructors remain reachable from their existing callers in `internal/output/shell_infra.go`, `internal/git/shell_detect.go`, and `cmd/create.go`, and `go test ./internal/core/... ./internal/output/... ./internal/git/... ./cmd/...` passes.

## 9. Command setup helper extraction

- [x] 9.1 Extract the 5-step `Config → GitClient → ContextDetector → filepath.Abs → DetectContext` setup duplicated across `runCreate`, `runDelete`, `runList`, `runCd`, `runPrune` into a new `cmd/setup.go` helper `detectContext(opts.Config, opts.GitClient)` returning `(*core.Context, *git.Client, error)` and verify `go build ./...` compiles and every `cmd/*_test.go` still passes. The helper takes the two function-typed accessor functions rather than a Factory pointer because each Options type stores them as fields rather than methods, and Go interfaces require methods (not function-typed fields) for interface satisfaction.
- [x] 9.2 Run `grep -nE 'f\.GitClient\(\)|NewContextDetector|DetectContext' cmd/*.go` and verify each file has exactly one call site to the new helper. Each of the 5 cmd files has exactly one.

## 10. Identical-twins resolution delete

- [x] 10.1 Characterization test (`TestResolveFromProjectContext_AndWorktreeAreComparable` in
  `context_resolver_test.go`, written as part of §0.2 coverage work)
  asserts that `resolveFromProjectContext(ctx, identifier)` and
  `resolveFromWorktreeContext(ctx, identifier)` produce identical
  results for `identifier` in `{main, feature-1, otherproj/branch-x}`
  with a Worktree context. The test passed before this group deleted
  the duplicate.
- [x] 10.2 Deleted `resolveFromWorktreeContext` from
  `internal/git/context_resolver.go`. `ResolveIdentifier` now routes
  both `core.ContextProject` and `core.ContextWorktree` cases
  through `resolveFromProjectContext`. Characterization test from
  10.1 continues to pass without modification.
- [x] 10.3 `grep -nE 'resolveFromWorktreeContext' internal/git/*.go`
  returns zero production-code matches (only the test reference in
  `context_resolver_test.go:411` — test was also updated to the new
  shape).

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

- [x] 12.1 Created `test/worktree/`, `test/shell/`, `test/git/`,
  `test/repo/`, `test/golden/`, `test/perf/`, `test/helpers_test/`.
  Moved source files with package declarations updated to match
  directory name. Moved `test/helpers/helpers_test.go` to
  `test/helpers_test/helpers_test.go` with package `helpers_test`
  (added `doc.go` for the external _test package to satisfy Go
  build). Updated 3 caller import paths:
  - `cmd/cd_test.go`: `twiggit/test/helpers` → `twiggit/test/git` (githelpers alias)
  - `test/e2e/fixtures/e2e_fixtures.go`: → `twiggit/test/git` (githelpers alias)
  - `test/e2e/infrastructure_verification_test.go`: → `twiggit/test/worktree` (worktreehelpers alias)
  Two tests in `helpers_test.go` that referenced unexported
  fields of ShellTestHelper were softened (internal-field assertions
  removed; covered by future test/shell internal tests in a
  follow-up). `go build ./...` compiles, `go test ./...` passes.
- [x] 12.2 `test/helpers/` is gone (`ls test/helpers/` returns
  no such file); `ls test/{worktree,shell,git,repo,golden,perf}/`
  shows all six directories with their source files.
- [x] 12.3 `git grep -nE 'twiggit/test/helpers'` returns 0
  production-code or test-code matches.
- [x] 12.4 Package comments added to each new `test/<domain>/`
  package. `test/package_boundary_test.go` enforces three rules:
  no top-level directory under `test/` matches the forbidden
  names (`helpers`, `util`, `common`, `misc`, `support`);
  each `test/<domain>/` package name matches its directory;
  each new package is importable. `go test ./test/...` passes.

## 13. cmd/util.go split

- [x] 13.1 Moved `ProgressReporter`, `NewProgressReporter`, `Report` into
  a new `cmd/progress.go`. (No `ReportProgress` exists in the current
  codebase — the dead-code removal was done by an earlier change.)
- [x] 13.2 Moved `ignoreWriter` and `writeOrIgnore` into a new
  `cmd/writer.go`.
- [x] 13.3 `verbosef` and `wrapArgsValidator` stay in `cmd/util.go`
  (now 36 LOC). All three share the `cmd` package; no import path
  changes required.

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