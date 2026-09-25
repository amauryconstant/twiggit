# Tasks

## 1. Validation cleanup (leaf code)

- [ ] 1.1 Collapse `validateBranchErr`, `validateProjectErr`, `validateShellErr` into a single `validateFieldErr(field, value, msg, sug)` constructor in `internal/core/validation.go` and verify `go build ./...` compiles and `go test ./internal/core/...` passes.
- [ ] 1.2 Promote `reservedNames` map literal to a package-level `var` in `internal/core/validation.go` and verify `go test ./internal/core/...` still passes (no behavioural change; addresses REVIEW finding on per-call re-allocation).

## 2. Mechanical modernization + documentation polish

- [ ] 2.1 Replace `sort.Strings` with `slices.Sort` at the three sites in `cmd/suggestions.go` and verify `go build ./...` compiles and `cmd/*_test.go` passes.
- [ ] 2.2 Replace `wg.Add(1) + defer wg.Done()` patterns with `wg.Go(...)` at the seven sites in `test/concurrent/concurrent_test.go` and verify `go test -race ./test/concurrent/...` passes.
- [ ] 2.3 Replace the two `strings.ReplaceAll` chains in `internal/core/shell_wrapper.go` and `test/e2e/helpers/test_id_generator.go` with a single `strings.NewReplacer(...).Replace(...)` call per site and verify the respective tests still produce byte-identical output.
- [ ] 2.4 Replace the hand-rolled `intToStr` function in `internal/core/errors_legacy.go` with `strconv.Itoa` and verify `go build ./...` compiles (output identical, no behaviour change).
- [ ] 2.5 Add godoc comments to `core.ValidationError.Error()`, `core.NotFoundError.Error()`, `core.OperationError.Error()`, `core.UsageError.Error()`, `ContextResolver.ResolveIdentifier`, and `ContextResolver.ResolutionSuggestions` and verify `go doc ./internal/core/... ./internal/git/...` renders the new prose.
- [ ] 2.6 Remove the four duplicate `// Package git/core` comments from files that already have a canonical `doc.go` and verify `go vet ./...` reports no warning.

## 3. Bool field prefix normalization

- [ ] 3.1 Rename `colorEnabled` → `isColorEnabled`, `existingOnly` → `isExistingOnly`, `valid` → `isValid`, `cacheEnabled` → `isCacheEnabled`, and `ProgressReporter.quiet` → `isQuiet` (with corresponding getter on the struct if exported callers depend on it) and verify `go build ./...` compiles and the affected `_test.go` files pass.
- [ ] 3.2 Run `grep -rE '\b(colorEnabled|existingOnly|cacheEnabled)\b' internal/ cmd/ --include '*.go'` after the rename and verify the only hits are the new `is*` form.

## 4. Type rename + Get-prefix removal + Factory collapse

This group is a single commit set because the renames are mutually
dependent: type renames change return-type signatures, method renames
follow, and the Factory collapse references the renamed role methods.

- [ ] 4.1 Rename `core.GitRepository` → `core.Repository`, `core.GitDir` → `core.RepoDir`, `core.GitCommit` → `core.Commit`, `core.GitBranch` → `core.Branch`, `core.BranchInfo` → `core.Branch`, `core.WorktreeInfo` → `core.Worktree`, `core.RemoteInfo` → `core.Remote`, `core.CommitInfo` → `core.Commit` using `gopls rename` per file and verify `go build ./...` compiles and every test in `internal/core/` + `internal/git/` passes.
- [ ] 4.2 Drop the `Get` prefix on `GetConfig`, `GetRepositoryStatus`, `GetRepositoryInfo`, `GetCommitInfo`, `GetResolutionSuggestions`, and the three `GetShellType` methods using `gopls rename` and verify `go build ./...` compiles.
- [ ] 4.3 Update the six `var _ core.Role = (*git.Client)(nil)` declarations in `internal/git/client.go` to use the renamed methods (no code change required if gopls rename propagated) and verify `go build ./internal/git/...` reports no compile-time role-satisfaction failure.
- [ ] 4.4 Delete the six per-role lazy fields (`RepoOpener`, `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`, `BranchWriter`) from `cmdutil.Factory` and update `Factory.Init()` to drop the six touch calls and verify `go build ./...` compiles and `internal/cmdutil/factory_test.go` passes (after updating the test to use `core.Role(client)` form).
- [ ] 4.5 Run `git grep -nE '\bGet(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)\b'` and `git grep -nE '\b(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)\b'` and verify zero matches in production code (`internal/`, `cmd/`, `test/`).

## 5. Error file rename

- [ ] 5.1 Rename `internal/core/errors_legacy.go` → `internal/core/errors_demoted.go` (matches the in-file purpose comment) and verify `go build ./...` compiles and the 5 demoted constructors (`NewGitRepositoryError`, `NewGitWorktreeError`, `NewGitCommandError`, `NewConfigError`, `NewContextDetectionError`) remain reachable from their existing callers.

## 6. Command setup helper extraction

- [ ] 6.1 Extract the 5-step `Config → GitClient → ContextDetector → filepath.Abs → DetectContext` setup duplicated across `runCreate`, `runDelete`, `runList`, `runCd`, `runPrune` into a new `cmd/setup.go` helper `detectContext(f)` returning `(*core.Context, *git.Client, error)` and verify `go build ./...` compiles and every `cmd/*_test.go` still passes.
- [ ] 6.2 Run `grep -nE 'f\.GitClient\(\)|NewContextDetector|DetectContext' cmd/*.go` and verify each file has exactly one call site to the new helper.

## 7. Identical-twins resolution delete

- [ ] 7.1 Delete `resolveFromWorktreeContext` from `internal/git/context_resolver.go` and route both `core.ContextProject` and `core.ContextWorktree` cases in `ResolveIdentifier` through `resolveFromProjectContext` and verify `go test ./internal/git/...` passes (no scenario currently relies on divergent behaviour).
- [ ] 7.2 Run `grep -nE 'resolveFromWorktreeContext' internal/git/*.go` and verify zero production-code matches remain.

## 8. Context resolver file split

- [ ] 8.1 Move `addMainSuggestion`, `addWorktreeSuggestions`, `addBranchSuggestions`, `addProjectSuggestions`, `getProjectContextSuggestions`, `getWorktreeContextSuggestions`, `getOutsideGitContextSuggestions` into a new `internal/git/context_resolver_suggest.go` and verify `go build ./...` compiles.
- [ ] 8.2 Drop the dead `_ *suggestionConfig` parameter from `addBranchSuggestions` (already unused) and verify the suggestion tests in `internal/git/` still cover the same scenarios.
- [ ] 8.3 Add a `ctx context.Context` first-parameter to the three `get*ContextSuggestions` methods and replace the four `context.Background()` sites with the caller's `ctx` and verify `go test -race ./internal/git/...` passes (cancellation propagates through suggestion building).
- [ ] 8.4 Reorder the file so `ContextResolver` (the public type alias) appears before the private `contextResolver` struct and verify `gopls references` shows the alias still resolves callers correctly.

## 9. Test helpers package split

- [ ] 9.1 Create `test/worktree/`, `test/shell/`, `test/git/`, `test/repo/`, `test/golden/`, `test/perf/` directories and move the corresponding source files (`worktree.go`, `shell.go`, `git.go`, `repo.go`, `golden.go`, `performance.go`) into them with their package declarations updated to match the directory name.
- [ ] 9.2 Delete the now-empty `test/helpers/` directory and verify `go build ./...` fails with import errors (expected intermediate state) before the next step.
- [ ] 9.3 Update the three production callers (`cmd/cd_test.go`, `test/e2e/infrastructure_verification_test.go`, `test/e2e/fixtures/e2e_fixtures.go`) to import the new content-named packages and verify `go build ./...` compiles and `go test ./...` passes.
- [ ] 9.4 Run `ls test/helpers/` and verify the directory does not exist; run `ls test/{worktree,shell,git,repo,golden,perf}/` and verify all six directories exist.

## 10. cmd/util.go split

- [ ] 10.1 Move `ProgressReporter`, `NewProgressReporter`, `Report`, `ReportProgress` into a new `cmd/progress.go` and verify `go build ./...` compiles.
- [ ] 10.2 Move `ignoreWriter` and `writeOrIgnore` into a new `cmd/writer.go` and verify `go build ./...` compiles.
- [ ] 10.3 Leave `verbosef` and `wrapArgsValidator` in `cmd/util.go` (now ~30 LOC) and verify every `cmd/*.go` import continues to resolve correctly (no import path changes required because all three files share the `cmd` package).

## 11. Spec sync and archival

- [ ] 11.1 Run `openspec sync specs --change naming-refactor-modernize` and verify the four delta files (`core-git`, `core-types`, `cli-factory`, `testing-helpers`) merge cleanly into the canonical `openspec/specs/<id>/spec.md` files.
- [ ] 11.2 Run `openspec validate --strict` against the synced specs and verify zero violations.
- [ ] 11.3 Run `openspec archive --change naming-refactor-modernize` and verify the change directory moves to `openspec/changes/archive/` and `CHANGELOG.md` gains the entry after `osx-generate-changelog` runs.

## 12. Final verification

- [ ] 12.1 Run `mise run verify` (format + lint:fix + lint:gated + vuln:check + test + build) and verify zero errors and zero warnings.
- [ ] 12.2 Run `go test -race ./...` and verify zero data-race reports.
- [ ] 12.3 Run `git grep -nE '\b(Get(RepositoryStatus|RepositoryInfo|CommitInfo|ShellType|ResolutionSuggestions|Config)|(GitRepository|GitDir|GitCommit|GitBranch|BranchInfo|WorktreeInfo|RemoteInfo|CommitInfo)|errors_legacy\.go|test/helpers/)\b'` and verify the only matches are in `openspec/changes/archive/` (historical change artifacts) and the new spec prose acknowledging the rename.
