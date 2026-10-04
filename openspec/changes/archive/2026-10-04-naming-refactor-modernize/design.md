# Design

## Context

The `naming-refactor-modernize` change addresses 14 REVIEW findings
spanning `golang-naming`, `golang-project-layout`, `golang-structs-interfaces`,
`golang-refactoring`, `golang-modernize`, and `golang-documentation`.
See `proposal.md` for motivation. The shape of the work is constrained
by:

- The Tier 2 depguard in `.golangci.yml:214-233` forbids
  `internal/core/` from importing anything beyond `$gostd +
  samber/lo`. Type renames inside `core/` stay self-contained.
- The `golang-refactoring` skill's rename discipline (per archived
  `openspec/changes/.../tasks.md`): `gopls rename` plus `git grep`
  cross-check after every rename to catch string-literal references
  the language server misses (test descriptions, error messages,
  spec prose).
- Changes A through D (foundation, CI gates, test/error
  discipline, Cobra + dead-code) are in-flight as separate OpenSpec
  changes. E assumes their prerequisites land first or in parallel:
  `nolintlint` enabled, `gocritic`/`nilerr`/`paralleltest` active,
  Go toolchain ≥1.26.2.

The change is mechanical at its core (rename, split, swap, modernize)
but the cumulative blast radius is ~50 production files plus every
test file in the repository. Compile-time safety depends on strict
ordering.

## Goals / Non-Goals

**Goals:**

- Reduce REVIEW finding count from 311 toward the project's
  no-warning target.
- Bring naming in line with `golang-naming` skill without breaking
  the public CLI surface.
- Reduce structural complexity in `context_resolver.go`,
  `cmd/util.go`, and `test/helpers/` via file splits that group by
  concern.
- Modernize Go idioms (`slices.Sort`, `wg.Go`, `strings.NewReplacer`)
  to match the rest of the codebase's already-modern posture.
- Document the API surface correctly: spec deltas describe the
  post-change contract, code matches the spec.

**Non-Goals:**

- Renaming the module path `twiggit` → `github.com/amoconst/twiggit`
  (deferred to its own change).
- Bumping to Go 1.27 for `strings.CutLast` / `reflect.TypeAssert[T]`
  (E stops at ≥1.26.2).
- Adding fuzz / benchmark / example tests (opportunistic only).
- Removing dead code (`ProgressReporter.ReportProgress`,
  `CreateOptions.HookRunner`, `cmd/error_formatter.go`,
  `cmd/output.go`) — owned by change D in-flight.
- Creating `llms.txt` — explicitly declined by the user.
- Refactoring the `slog.SetDefault` / `Factory.Logger` dual-channel
  log path — owned by change C.
- Changing the CLI command surface (flags, arguments, exit codes,
  output formats preserved).

## Decisions

### Decision 1: Factory accessor collapse removes the six fields

**Choice**: Delete the 6 per-role fields from `cmdutil.Factory`
entirely. Callers needing role-narrowed access assign the
`*git.Client` returned by `f.GitClient()` to a local role-typed
variable.

**Rationale**: The fields had zero external callers (only
`factory_test.go` and `Factory.Init()`). They were 30 lines of
byte-identical wrapper code. `*git.Client` already satisfies every
role through embedded promotion, so no narrowing capability is lost.

**Alternatives considered**:
- Keep the fields, extract a single helper. Preserves the API
  surface but keeps YAGNI-violating dead fields with one fewer
  argument. Rejected: spec contract should match reality.
- Keep the fields, no helper. Rejected: 30 lines of identical code.

### Decision 2: `test/helpers` becomes six content-named packages

**Choice**: Split into `test/worktree/`, `test/shell/`,
`test/git/`, `test/repo/`, `test/golden/`, `test/perf/`. Three
production callers update imports.

**Rationale**: `golang-project-layout` skill flags `helpers` as the
canonical anti-pattern. The current `test/helpers/` already groups
by concern in separate files; the package boundary was the only
inconsistency. Three importers (`cmd/cd_test.go`, two `test/e2e/`
files) make the import-update cost trivial.

**Alternatives considered**:
- Keep `test/helpers`, rename file. Doesn't address the skill
  violation.
- Split into fewer packages (e.g., 2). Some files (`golden.go`,
  `performance.go`) don't share helpers; merging them would create
  a second `helpers`-shaped package.

### Decision 3: `context_resolver.go` splits into two files

**Choice**: Split into `internal/git/context_resolver.go` (resolver
functions) and `internal/git/context_resolver_suggest.go`
(suggestion builders). Cross-cutting filters
(`fuzzyMatch`, `matchesExclusionPatterns`, `containsPathTraversal`,
`parseCrossProjectReference`, `worktreeExists`,
`validatePathUnder`) stay in `context_resolver.go` because they
serve both halves.

**Rationale**: The file mixes three concerns: resolver, suggestion
builder, and cross-cutting filters. Two-way split is the cleanest
because filters are equally used by both halves; extracting them to
a third file creates artificial separation. Post-split, both files
fit under 350 LOC.

**Alternatives considered**:
- Three-file split (resolver, suggestion, filter). Rejected: filters
  don't form a stable boundary — suggestion code uses some, resolver
  uses others; splitting requires importing across files in the same
  package without boundary benefit.
- One file, rewrite. Rejected: 641 LOC exceeds the codebase's
  per-file norm (~250 LOC median).

### Decision 4: `Get` prefix drops on returns-of-noun methods only

**Choice**: `GetConfig` → `Config`, `GetRepositoryStatus` →
`RepositoryStatus`, `GetRepositoryInfo` → `Repository`,
`GetCommitInfo` → `Commit`, `GetResolutionSuggestions` →
`ResolutionSuggestions`, `GetShellType` → `ShellType`. Imperative
mutating methods (`DeleteWorktree`, `CreateWorktree`,
`IsBranchMerged`, `DeleteBranch`) keep their verb names.

**Rationale**: `golang-naming` skill: "Go omits `Get` — `Config()`
reads naturally." Verbs on mutators stay because they describe the
operation, not the return value. `ResolveIdentifier` keeps its verb
because it describes an action.

**Alternatives considered**:
- Drop `Get` on mutators too (`DeleteWorktree` → `WorktreeDeletion`).
  Rejected: loses imperative clarity; convention in stdlib (`os.Remove`,
  `client.Do`) keeps verbs.
- Keep `Get` everywhere. Rejected: violates skill rule.

### Decision 5: `RepositoryStatus` keeps its name; only the `Get` prefix drops

**Choice**: `core.RepositoryStatus` stays. `GetRepositoryStatus`
becomes `RepositoryStatus`.

**Rationale**: `Status` alone is too generic — it could shadow a
future `OperationStatus`, `BuildStatus`, or `ContextStatus`. The
return type name matches the method name (`RepositoryStatus(ctx)
(RepositoryStatus, error)`), a Go idiom (`time.Time.String()`,
`http.Response.Header()`). Stutter is avoided because
`RepositoryStatus` describes a domain object, not the noun-of-noun
problem the `Info`-suffix rename solves.

**Alternatives considered**:
- Rename to `Status`. Rejected: collides semantically with future
  status types.
- Rename to `RepoStatus`. Rejected: re-introduces stutter.

### Decision 6: `GitDir` renames to `RepoDir`, not `Dir`

**Choice**: `core.GitDir` → `core.RepoDir`.

**Rationale**: The type describes a discovered git directory on
disk (`{Name, Path}`). `Dir` alone collides conceptually with
`os.DirFS`, `filepath.Dir`, and any future directory descriptor.
`RepoDir` aligns with existing project vocabulary: `RepoOpener`,
`RepoFinder`, `core.IsMainRepo`.

**Alternatives considered**:
- `Dir`. Rejected: too generic.
- `GitDir` (unchanged). Rejected: violates `core-types` delta
  eliminating `Git*` stutter.

### Decision 7: `errors_legacy.go` renames to `errors_demoted.go`

**Choice**: `internal/core/errors_legacy.go` →
`internal/core/errors_demoted.go`.

**Rationale**: The file's own header says "Demoted git / config /
context error constructors. Each previously concrete error type
collapses to OperationError; Op names the source." The name
`errors_legacy.go` implies dead code; the constructors are still
actively called from `internal/output/shell_infra.go`,
`internal/git/shell_detect.go`, and `cmd/create.go`. `errors_op.go`
would over-narrow (file holds 5 constructors for different
  `Op` values, not just `OperationError`).

**Alternatives considered**:
- `errors_op.go`: over-narrows the file (5 constructors for distinct `Op` values, not just `OperationError`).
- `errors.go`: collides with the existing `internal/core/errors.go`.
- Keep `errors_legacy.go`: the file's own header says the demoted constructors are still actively called — the name implies dead code and misleads new readers.

### Decision 8: Bool fields use `is`/`has`/`can` prefix

**Choice**: `colorEnabled` → `isColorEnabled`; `existingOnly` →
`isExistingOnly`; `valid` → `isValid`; `cacheEnabled` →
`isCacheEnabled`; `ProgressReporter.quiet` → `isQuiet`.

**Rationale**: `golang-naming` skill: `is/has/can` prefix on
bool fields. Aligns with existing `IOStreams.isStdoutTTY`,
`isStdinTTY`, `isStderrTTY`.

**Alternatives considered**:
- Drop the `is` prefix on `existingOnly` to keep it short. Rejected:
  drifts from sibling bools.

### Decision 9: Compile-safe rename order

**Choice**: Apply renames in this order:
1. Leaf code in `core/validation.go` (`C11`, `C12`).
2. Mechanical nits and godoc (`C13`, `C14`).
3. Bool prefix (`C6`).
4. Type rename (`C7`) split into atomic alias-driven commits per
   Decision 11: 4a `GitBranch` → `Branch` via type alias; 4b
   `BranchInfo` → `Branch` via type alias; 4c other type renames
   (`GitRepository`, `GitDir`, `GitCommit`, `WorktreeInfo`,
   `RemoteInfo`, `CommitInfo`); 4d compile-time data-type fixture in
   `internal/core/`. `Get*` method names stay unchanged so the
   composite `*reader` and `*cliClient` continue to satisfy the role
   interfaces; the role interfaces now reference the new return
   types.
5. Method rename (`C5`) alone as a separate commit: `gopls rename`
   for the 8 `Get*` methods on `*reader` + `*cliClient` plus the
   role interface method names in `internal/core/git.go`. Returns
   typecheck against the already-renamed `*core.Repository` etc.
6. Factory collapse (`C1`) alone as a third commit: delete the 6
   per-role lazy fields, update `Factory.Init()`, update
   `factory_test.go` to use the specific role interface form.
7. `errors_legacy.go` rename (`C8`).
8. 5-step setup extraction (`C4`).
9. Identical-twins delete (`C3`).
10. `context_resolver.go` split (`C2`).
11. `test/helpers` split (`C9`, single atomic PR per the user's
    chosen sequencing).
12. `cmd/util.go` split (`C10`).

**Rationale**: Each step keeps `go build ./...` green. Splitting the
rename cluster into three atomic single-category commits (per the
`golang-refactoring` skill's 100-500 LOC + single-category rules) keeps
each PR reviewable and bisectable. The compile-safety argument for a
single commit ("role interfaces never reference absent types") is
preserved by sequencing: type renames come first, method renames use
the new return types, Factory collapse uses the new method names. Renames
that produce partial states (e.g., renaming `BranchInfo` before
renaming `GetRepositoryInfo` to its new return type) would compile
fail and force bisecting — the sequencing above avoids that.

**Alternatives considered**:
- Rename everything in one commit. Rejected: bisection impossible
  on failure; review diff too large.
- Rename only types first, then methods in a later change. Rejected:
  splits a single coherent rename into two MRs without benefit.

### Decision 10: Suggestion methods gain `context.Context` first-parameter

**Choice**: `getProjectContextSuggestions`, `getWorktreeContextSuggestions`,
`getOutsideGitContextSuggestions` all take `ctx context.Context`
as their first parameter. Internal `context.Background()` calls
are replaced with the caller's `ctx`.

**Rationale**: Today the methods take `ctx *core.Context` (a domain
type) but no `context.Context` (stdlib type for cancellation). The
4 `context.Background()` sites discard caller cancellation. Adding
the param restores SIGINT propagation through suggestion building.

**Alternatives considered**:
- Pass `ctx` via the `suggestionConfig` struct. Rejected: pollutes
  the config with a non-option concern.
- Leave `context.Background()` and rely on the suggestion timeout
  in `cmd/suggestions.go`. Rejected: the REVIEW finding stands;
  SIGINT should cancel work promptly.

### Decision 11: Type aliases for `GitBranch` and `BranchInfo` collapse

**Choice**: Rename `GitBranch` → `Branch` and `BranchInfo` →
`Branch` via a two-step type-alias migration. In the first step,
add `type GitBranch = Branch` to `internal/core/git_types.go`,
update all `GitBranch` call sites in `internal/git/`, `cmd/`, and
tests to use `Branch`, and drop the alias in a follow-up commit.
Repeat for `BranchInfo` (which has a separate collision surface
— its return type on `ListBranches` and its use in golden
fixtures). The two sequences are independent and land in
separate commits.

**Rationale**: Both source types share the same target name.
A direct rename would require one commit to break the build
(the package would temporarily contain two `Branch` types in
the same file, or a `Branch` and a stale `GitBranch` whose
methods reference a not-yet-existing `Branch` field). A type
alias (`type Old = New`) is the officially-blessed Go mechanism
for gradual code repair: the old and new names stay
interchangeable while callers migrate incrementally. Each
commit stays green; `gopls rename` works against the alias
without complaint.

**Alternatives considered**:
- Sequential commits, no aliases. Rejected: the intermediate
  state has `GitBranch` referencing the new `Branch` while
  `BranchInfo` is still being used, which compiles only if the
  two old types share a field set. They don't (`BranchInfo` has
  `IsCurrent`, `Remote`, `Commit`, `Author`, `Date`;
  `GitBranch` has a narrower read-side shape). Without aliases,
  one of the two renames has to land in a single big-bang
  commit.
- Big-bang single commit renaming both. Rejected: violates
  the `golang-refactoring` rule against mixing structural and
  behavioral changes, and the bisection cost is high.

### Decision 12: Gopatch rewrite rules drive mechanical transforms [SUPERSEDED]

> **Superseded post-implementation.** The mechanical transforms landed
> via direct `gopls rename` per cluster (Decision 9 ordering), one commit
> per category. The `// +gopatch` rule-emission and golden-test work
> described below was not performed. The audit trail is the per-cluster
> commit log on `main` plus the cross-type-rename grep at task 4.5. No
> follow-up is required: the renames are complete and verified by
> `git grep` returning zero matches.

**Original choice (not executed)**: Emit one `gopatch` rewrite rule per
mechanical transform in this change: Get-prefix drop (group 5), each
per-type rename (groups 4.1-4.3a-f), bool-field prefix (group 3), and
each modern idiom (groups 2.1-2.3: `sort.Strings` → `slices.Sort`,
`wg.Add(1) + defer wg.Done()` → `wg.Go`, `strings.ReplaceAll` chain →
`strings.NewReplacer`). Each rule ships with a golden test that
asserts the rewrite against a representative input (sample fixture in
`internal/cmdutil/factory_test.go` or `test/golden/`, named via the
rule's `// +gopatch` marker).

**Original rationale**: The change touches ~50 production files with
cross-cutting mechanical patterns. `gopls rename` per file leaves no
cross-file audit trail and forces hand-checking each site; `gopatch`
rules are reviewable as a single diff, re-runnable against the full
tree, and testable against golden inputs — exactly what the
`golang-refactoring` skill prescribes for "a change that recurs across
many sites". The setup cost (generating one rule per transform, wiring
the `// +gopatch` golden-test marker) pays back across the ~50-file
blast radius. Sequencing: rules ship in task 0.6 (Migration Plan
step 7), before any rename or modernize commit lands, so each commit
can `gopatch apply` against its rule and `git grep` cross-check.

**Alternatives considered (re-evaluated post-implementation)**:
- Stick with `gopls rename` per file. **Selected in practice**: the
  per-cluster commit log + the post-cluster `git grep` cross-checks
  (tasks 4.5 / 7.1 / 15.3) provided a sufficient audit trail for this
  change's blast radius without the setup cost of `gopatch` rules.
- `gofmt -r` rewrites. Rejected: rule expressiveness is too limited
  for the type-alias gradual repair and role-method signature updates
  that this change requires.
- One-off `sed`/`awk` scripts. Rejected: no AST awareness, no
  golden-test mechanism, no Go-language-server-style interface-
  satisfaction guard.

## Risks / Trade-offs

- **Large blast radius** (~50 production files, every test file) →
  Apply renames via `gopls rename` per file, then `git grep` for
  string-literal references in spec prose, error messages, and
  test names. Each rename commit ends with `go build ./... &&
  go test ./...` green.
- **Linter might re-flag renamed symbols** (e.g., `gocritic`
  preferring `GetX` style) → Coordinate with change A which
  enables the linter; if conflicts emerge, prefer the skill
  verdict.
- **External test mocks** in `test/mocks/` may hard-code old method
  names → Update mocks alongside the role interface renames in
  the same commit set.
- **depguard allow-list in `.golangci.yml:257`** lists `twiggit/cmd`
  self-import — irrelevant for E; no depguard change required for
  `test/<domain>/` packages because tests are excluded from
  depguard rules.
- **`cmd/cd_test.go:12` imports `test/helpers`** — single-file
  update as part of `C9`. Risk: missed importers compile-fail; run
  `go build ./...` after the `C9` commit.
- **Spec delta wording drift** — the four delta files describe
  post-change behaviour but real code may diverge during
  implementation → run `openspec validate --strict` after each
  commit that touches a spec'd surface.
- **`slog.SetDefault` / `Factory.Logger` dual channel** not fixed
  by E (owned by change C) → No interaction expected; E doesn't
  touch logger wiring.

## Migration Plan

1. Apply changes A through D in their existing OpenSpec change
   branches; merge them before E starts.
2. Branch `naming-refactor-modernize` from `main` after A-D merge.
3. Run safety-net coverage gate: `go test -cover
   ./internal/core/... ./internal/git/... ./internal/cmdutil/...`
   and confirm ≥80% line coverage on each package. If below,
   characterize-test the under-covered branches before group 4
   starts.
4. Run `golangci-lint run --enable-only modernize ./...` and
   capture baseline output (`tmp/modernize-baseline.txt`) for
   comparison after each modernization commit.
5. Run `go mod tidy && git diff --exit-code` and confirm no
   `godebug` lines in `go.mod` pin removed keys
   (`asynctimerchan`, `tlsunsafeekm`, `tlsrsakex`, `tls3des`,
   `tls10server`, `x509keypairleaf`, `gotypesalias`).
6. Verify `git show main:openspec/specs/core-git/spec.md | grep
   -c WorktreeWriter` returns ≥ 1 (the existing follow-up note
   for the 4-method `WorktreeWriter` is present). If 0, manually
   add a one-line follow-up to the existing `core-git` spec
   before this change's delta applies.
7. Decision 12 originally specified installing `gopatch` and emitting
   one rewrite rule per mechanical transform before any rename
   commit. **Superseded** post-implementation: transforms landed via
   direct `gopls rename` per Decision 9 ordering, and the per-cluster
   cross-check greps (tasks 4.5 / 7.1 / 15.3) provided the audit
   trail. No `gopatch` setup work was performed.
8. Land commits in the order in Decision 9, gated by Sign-off
   Gates 1, 2, and 3 (per the user-chosen sequencing). Each
   commit:
   - Compiles (`go build ./...`).
   - Passes tests (`go test ./...`).
   - Runs `mise run lint` clean.
   - Carries a single cluster (C1-C14) so bisection stays useful.
   - Includes its own tag/reflection grep verification.
8. After all commits land, run `openspec sync specs --change
   naming-refactor-modernize` to merge the four delta files into
   the canonical specs.
9. Run `openspec archive --change naming-refactor-modernize` to
   finalize.
10. Run `osx-generate-changelog` to add the entry to `CHANGELOG.md`.

> Step 7 was renumbered: Pre-merge verification moves into the
> existing step 6; `gopatch` tooling setup is the new step 7.
> The follow-on steps 8-10 are renumbered from the previous 8-10.

## Open Questions

None. All decisions locked in the explore-mode discussion above.

## Deferred Work

Out of scope for this change; tracked for future OpenSpec changes:

- **`WorktreeWriter` interface split** — sits at 4 methods
  (`CreateWorktree`, `DeleteWorktree`, `ListWorktrees`,
  `PruneWorktrees`), exceeding the 1-3 method shape recommended by
  the `golang-structs-interfaces` skill. The split is deferred to
  a future OpenSpec change; this change keeps the 4-method shape
  to limit blast radius. Already noted in the existing `core-git`
  spec under the "WorktreeWriter exposes the four worktree-lifecycle
  methods" requirement as a follow-up.
- **`errors.AsType[T]` adoption** — Go 1.27+ generic API for
  typed error chain inspection. Project target is ≥1.26.2
  (Non-Goals); this change does not promote the bump. Defer until
  the Go 1.27 toolchain upgrade lands.
- **`strings.CutLast` / `reflect.TypeAssert[T]` adoption** — Go 1.27+
  stdlib improvements. Same defer rationale as `errors.AsType[T]`.
