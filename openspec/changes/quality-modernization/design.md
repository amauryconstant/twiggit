# Design: Quality Modernization

See `proposal.md` for motivation (the ~580 REVIEW.md findings that
remain after `foundation-error-chain` and `architecture-layer-inversion`
ship) and `specs/` for the requirement surface (16 spec deltas across
the changed behavior). This document covers the architectural
decisions, the constraints that shape them, the execution order,
and the testing strategy that gets the change landed without
breaking unrelated domains.

## Context

The project after Change 1 + Change 2 lands has the following
characteristics that shape this design:

- Project targets `go 1.25.5` per `go.mod`; Go 1.26 features
  (`errors.AsType[T]`) are unavailable. Local helpers from
  Change 1 are sized to swap mechanically when the toolchain
  bumps.
- Project rule per `openspec/config.yaml`: tests are written
  AFTER implementation. Vertical slicing is the smallest unit
  that satisfies "implementation + tests" together.
- Layer rule is mechanical after Change 2: `depguard` blocks
  reverse imports. `internal/service/` cannot import
  `internal/infrastructure/` directly.
- 287 `.On(...)` call sites across the test tree have **zero**
  paired `mock.AssertExpectations` calls. This is the universal
  no-assert gap and the most leveraged single fix in this change.
- 153 `context.Background()` literals in test files. Three
  distinct structural buckets: in-fn literals (~140), helper
  signatures (~5), and mock matchers (~30). The matcher bucket
  is the one that exposes silent-pass bugs.
- 14 `os.IsNotExist(err)` sites remain after Change 2 deletes
  `git_utils.go` and fixes `project_service.go:221`.
- 12 leaf `executeX` functions in `cmd/` call
  `context.Background()` instead of `cmd.Context()`.
- 42 `_, _ = fmt.Fprint*` discarded write errors split into 5
  semantic categories with 3 different propagation rules.
- 25 receiver-less methods in `internal/service/` are Fowler-
  shaped Extract Function candidates.
- 12 4-string constructor signatures in `internal/domain/errors.go`
  are mechanically convertible to struct literals.
- `test/mocks/helpers.go` does **not** exist yet (Change 1 task
  5.1 was scaffolded in design but never landed). The variadic
  bug at `test/mocks/cmd_mocks.go:223,233` is still present.
- `openspec/config.yaml:134` has a YAML parse error from an
  embedded list-item comment that every `openspec` CLI invocation
  warns about. Either Change 1 task 1.2 lands first, or this
  change's slice S1 absorbs the fix.
- `.goreleaser.yml:73` owner `amoconst` vs hook script's
  GitHub target `amauryconstant/twiggit` is a real release
  blocker; the GitHub mirror lives at
  `github.com/amauryconstant/twiggit`.
- govulncheck's first run on the current dependency baseline
  (`go-git v5.16.5`, `mergo v1.0.2`, `x/crypto v0.48.0`,
  `x/sys v0.41.0`) will likely surface advisories. The CI
  gate is added AFTER `go get -u=patch`.

## Goals / Non-Goals

**Goals:**

- Land all 17 slices in dependency order (tests first, then
  production, then docs/deps/CI/specs) so each commit compiles
  and passes tests in isolation. `git bisect` works at every
  point.
- Add the `t.Cleanup(mock.AssertExpectations)` fixture so every
  mock-based test reports unfulfilled expectations at test end.
  This is the highest-leverage single fix in the change.
- Eliminate the silent-pass pattern (4 sites) and the
  order-dependent test pattern (`os.Chdir` + `defer`, 14 sites)
  so `-shuffle=on` runs cleanly.
- Tighten lint thresholds (`funlen` 150→120, `gocyclo` 25→13)
  per-file in the same commit that fixes the function. No
  grandfathered `//nolint`.
- Thread `cmd.Context()` and `signal.NotifyContext` through
  every leaf command so SIGINT/SIGTERM cancel in-flight work.
- Land the goreleaser owner fix and the digest-pinned Docker
  base so release artifacts target the actual GitHub mirror.
- Add govulncheck + baseline + Dependabot + `CODEOWNERS` so
  dependency drift is detected automatically.

**Non-Goals:**

- Wholesale Tier 2 CLI collapse (`cmd/{name}/main.go +
  internal/{core,git,output,iostreams,cmdutil}/`). Deferred
  per Change 2 design §"Non-Goals".
- Interface Segregation Principle refactor on `GoGitClient`
  (8 methods) and `CLIClient` (6 methods). Constructor would
  take 8+ role interfaces. Deferred.
- Receiver-less → free functions in `test/` helpers. The change
  applies the pattern in `internal/service/` only; test helpers
  are added on demand.
- The Go 1.26 `errors.AsType[T]` migration. Local
  `cmd/error_formatter.go` `asType[T]` from Change 1 stays in
  place until the toolchain bumps.
- AI-driven PR review tooling (Claude Code, Copilot).
- Codecov integration; project-side coverage threshold
  discussion; Go version CI matrix.

## Decisions

### 1. Mock cleanup fixture is helper-based, not per-test boilerplate

**Choice.** `test/mocks/helpers.go` exposes a constructor
pattern per mock type: `func NewMockX(t *testing.T) *MockX` that
constructs the mock and registers `t.Cleanup(func() {
mock.AssertExpectations(t) })`. Tests construct mocks via the
helper; the cleanup is implicit.

**Rationale.** Per-test boilerplate (`t.Cleanup(mock.AssertExpectations)`
in every test) is mechanical and easy to forget. A constructor
helper makes the cleanup contract part of the mock's public
API; a test that calls `mocks.NewMockX(t)` cannot forget to
register cleanup.

**Alternatives considered:**

- **Suite-based assertion** (`suite.Run` + `suite.TearDownAll`):
  the project does not use `testify/suite` per the
  `infrastructure-release` spec rule. Switching to suite
  breaks the convention. Rejected.
- **Per-test inline `t.Cleanup`**: works but requires the
  reviewer to spot missing calls. Helper-based is mechanical.
  Rejected.
- **Run-after-each-subtest hook**: Go's `testing` package does
  not expose this; would require build-tag magic. Rejected.

### 2. t.Context migration splits into 3 sub-slices

**Choice.** Three slices with distinct risk profiles:
S4 (in-fn literals, 140 sites, mechanical), S5 (helper
signatures, 5 sites, signature change), S6 (mock matchers, ~30
sites, may surface silent-pass bugs).

**Rationale.** The three buckets have different verification
shapes. S4 is provably correct by inspection (replacing one
identifier). S5 changes function signatures, so callers must
update — mechanical via `gopls rename`. S6 changes mock
expectation matching; a matcher swap from
`context.Background()` to `mock.Anything` removes a
type-check that may have been hiding unfulfilled-call bugs.
S6 lands last so its failure surface is isolated.

**Alternatives considered:**

- **Single slice with all 153**: hides the silent-pass
  exposure inside a "mechanical sweep." Reviewers cannot tell
  which changes are risk-bearing. Rejected.
- **`mock.MatchedBy(func(ctx) bool { return ctx != nil })`
  everywhere in S6**: preserves a type-check but is more
  verbose than `mock.Anything`. Same risk profile. Rejected
  (use `mock.Anything` for simplicity).

### 3. `_, _ = fmt.Fprint*` is per-site classified, not swept

**Choice.** 42 sites split into 5 categories. The fix differs
per category:
- A (navigation output, 4 sites): capture error, return on
  write failure (cd target path is user-critical).
- B (warnings, ~12 sites): capture error, surface via the
  command's error path.
- C (progress / logv / hints, ~14 sites): switch to
  `c.ErrOrStderr()`, suppress errors (informational).
- D (result detail, ~12 sites): keep `errOut` but ensure the
  cobra writer is used; classify as either A or B based on
  whether the detail is critical or informational.
- E (TTY prompt, 1 site at `cmd/prune.go:134`): keep as-is,
  gate on `isatty` upstream.

**Rationale.** A single rule ("propagate all errors") makes
progress messages fatal and breaks the verbose-log helper's
contract. A single rule ("suppress all") hides real errors
in the navigation path. Per-site classification matches the
user's intent: navigation paths and warnings are critical;
progress is informational.

**Alternatives considered:**

- **All-or-nothing**: rejected because the 5 categories have
  different user-impact profiles.
- **`io.Writer` injection everywhere with no error capture**:
  mechanical but loses category A's safety. Rejected.

### 4. Receiver-less methods → free functions, in `service/` only

**Choice.** Convert 25 methods in `internal/service/` whose
receiver is unused (`s.` does not appear) to free functions
in the same package. Per-method Extract Function with
`gopls rename` to propagate call sites.

**Rationale.** The Fowler-shaped refactor isolates pure
helpers from stateful methods. The package boundary stays
the same; the call sites stay the same except the receiver
is dropped. `gopls rename` keeps the test tree in sync.

**Alternatives considered:**

- **Move to a new `internal/service/internal/` sub-package**:
  forces a new import path; more disruption than the refactor
  warrants. Rejected.
- **Leave as methods**: keeps the smell; review re-flags it.
  Rejected.

### 5. 4-string ctors → struct literals, mechanical

**Choice.** Convert 12 constructors in
`internal/domain/errors.go` (e.g., `NewValidationError(request,
field, value, message string)`) to struct literals. The
constructor still exists but takes a struct; call sites
migrate to `&ValidationError{Request: ..., Field: ...,
Value: ..., Message: ...}`. `gopls rename` may not work
across signature shape changes; manual update per call site.

**Rationale.** Adjacent same-typed parameters are a known
mistake-magnet (`createBranch("main", "feature")` reads
backwards). Struct literals make call sites self-documenting.

**Alternatives considered:**

- **Functional options pattern**: more flexible but
  higher-ceremony; the project has 4-string ctors, not 8+.
  Rejected.
- **Leave as-is**: review re-flags. Rejected.

### 6. parseWorktreeList splits into a typed parser

**Choice.** Replace the 53-line state-machine parser with a
small `parsePorcelain` helper that uses `bufio.Scanner`,
a struct-of-counters for the current worktree header, and
returns `[]domain.WorktreeInfo`. Two functions instead of
one: `parsePorcelainHeader(line) (porcelainHeader, bool)`
and `accumulateField(current *porcelainWorktree, line)`.

**Rationale.** The current parser mutates a local
`currentWorktree` pointer across loop iterations and
silently skips malformed entries (the `currentWorktree != nil`
guard). A typed parser eliminates the pointer aliasing and
makes the failure mode visible.

**Alternatives considered:**

- **Parser combinator library** (e.g., `participle`):
  external dep for a 53-line parser; overkill. Rejected.
- **Use `git worktree list --porcelain --format=%...`**:
  git CLI doesn't expose this format; stuck parsing porcelain.
  Rejected.

### 7. Lint thresholds tighten per-file, in the same commit as the fix

**Choice.** Each function that exceeds the new threshold is
split (or shortened) in the same commit that lowers the
threshold for that linter setting. No wholesale lower + add
per-line `//nolint` for grandfathered code.

**Rationale.** The user accepted "stage per-file in same
commit as the fix" in exploration. CI never breaks mid-PR;
the threshold is always at or below the current code's
worst-violation.

**Alternatives considered:**

- **Wholesale lower + grandfathered `//nolint`**: faster but
  leaves ~30 permanent lint suppressions. Rejected.
- **Leave thresholds loose**: review re-flags on every
  audit. Rejected.

### 8. Goreleaser owner aligns with GitHub mirror

**Choice.** `.goreleaser.yml:73` `owner: amoconst` → `owner:
amauryconstant`. The hook script already targets
`amauryconstant/twiggit`. The change resolves the mismatch
in favor of the actual GitHub mirror.

**Rationale.** The homebrew tap formula URL is built from the
goreleaser owner + repo. If owner is `amoconst` and the GitHub
mirror is `amauryconstant/twiggit`, the formula URL resolves
to a non-existent tap. Aligning goreleaser.yml with the
actual mirror unblocks release.

**Alternatives considered:**

- **Move the hook script to `amoconst`**: requires the GitHub
  mirror to live at `amoconst/twiggit`. It doesn't. Rejected.
- **Keep both, document the inconsistency**: drift continues.
  Rejected.

### 9. Coverage threshold is 70%

**Choice.** `mise run verify` and CI fail when filtered
coverage drops below 70% (non-zero exit). The
`infrastructure-release` spec already requires `≥70%` on all
packages; this change makes the gate mechanical.

**Rationale.** The current `coverage.sh` only emits a
warning, so coverage drift is invisible to CI. A 70% threshold
is consistent with the existing spec requirement.

**Alternatives considered:**

- **80%**: stricter but may block legitimate refactors; no
  evidence the project operates at 80% today. Rejected.
- **Project-side codecov.yml**: deferred (Non-Goals).

### 10. TestMain builds once via package var, exports via env var

**Choice.** `main_test.go` defines a `TestMain(m *testing.M)`
that compiles the binary into `t.TempDir()` once, stashes
the path in a package-level variable, and exports it via
`os.Setenv("TWIGGIT_E2E_BINARY", binaryPath)`. Each test
reads the env var (or the package var).

**Rationale.** The current code calls `buildTestBinary`
inside each of 6 test functions — 6 redundant builds per
`go test` invocation. `TestMain` centralizes the build;
the env var is the cross-package access path.

**Alternatives considered:**

- **Package var only**: works within `main_test.go` but
  tests in other packages cannot reach it. Rejected.
- **Env var only, no package var**: works but every test
  reads `os.Getenv`; the package var saves a syscall. Both
  are kept.

### 11. t.Chdir replaces os.Chdir + defer os.Chdir everywhere

**Choice.** All 14 `os.Chdir` + `defer os.Chdir` patterns
become `t.Chdir(tempDir)`. Per-test cleanup is automatic;
order-dependent tests fail under `-shuffle=on`.

**Rationale.** The 14 sites are spread across 4 files
(`internal/service/worktree_service_test.go`,
`test/integration/prune_integration_test.go`,
`test/integration/path_utilities_test.go`,
`test/integration/context_detection_test.go`). They are the
biggest source of order-dependent test failure under shuffle.

**Alternatives considered:**

- **Per-subtest `os.Chdir` with `defer t.Chdir`-equivalent**:
  Go 1.24+ has `t.Chdir`; no need to roll our own. Rejected.

### 12. Slog migration extends to project_service.go only

**Choice.** Slog replaces 2 silent skip-and-continue sites in
`project_service.go` (`ListProjects`, `ListProjectSummaries`).
Change 2 already covers `worktree_service.go`.

**Rationale.** The same pattern (silent swallow in a service
loop) recurs in `project_service.go`. Change 3 closes the
loop. Other layers (cmd/) use `slog` via `logv` and the
main entry point; service-layer is the gap.

**Alternatives considered:**

- **Migrate all services to slog in one slice**: scope creep;
  not all services emit logs. Rejected.
- **Leave project_service.go silent**: review re-flags.
  Rejected.

### 13. Slices land in tests-first order

**Choice.** Slices S4-S7 (test reliability) land before
S8-S13 (production modernization). S14-S17 (docs/deps/CI/specs)
land last.

**Rationale.** Production modernization depends on a stable
test baseline. If slice 8 (free-function refactor) breaks a
test, bisect points at the refactor, not at a
mid-test-rewrite baseline. Tests-first means production
commits always have a green test baseline.

**Alternatives considered:**

- **Production-first**: each production commit has a green
  baseline, but mid-test-rewrite commits have ambiguous
  failures. Rejected.

### 14. Dependency bump before govulncheck gate

**Choice.** `go get -u=patch` runs first (slice S15),
producing a clean `go.mod`. The govulncheck baseline
(`govulncheck.json`) is captured at the same commit. The CI
step that enforces the gate lands in slice S16.

**Rationale.** Without the patch bump, the first govulncheck
run surfaces advisories that block the gate's introduction.
Bumping first produces a clean baseline; the gate enforces
*future* regressions, not pre-existing ones.

**Alternatives considered:**

- **Gate first, deal with findings later**: blocks the merge
  until each finding is individually addressed. Rejected.

## Risks / Trade-offs

- **`test/mocks/helpers.go` does not exist yet**. → S1
  creates the file as part of absorbing Change 1's missing
  pieces. If Change 1 is going to land first, its task 5.1
  should also complete; S1's scope is then reduced to a
  sanity check.
- **`openspec/config.yaml:134` parse error**. → S1 includes
  the fix (move the embedded comment out of the list item)
  if Change 1's task 1.2 hasn't landed. Same line move.
- **`-shuffle=on` exposes order-dependent tests**. → S4
  fixes the `os.Chdir` sites and `wg.Go` migrations before
  shuffle lands. Shuffle may still surface 5-10 more
  order-dependent tests; these are tracked as bugs to fix
  in subsequent PRs.
- **Mock matcher swap exposes 10-30 silent-pass tests**. →
  S6 is dedicated to this; the failures are collected, the
  tests are fixed in S6 (not deferred).
- **Receiver-less → free function refactor (25 methods)
  cascades**. → S13 is the only slice that does this; S12
  (struct-literal ctors) lands first to absorb related
  mechanical churn. If S13 exceeds ~500 LOC delta, it splits
  into S13a (extract functions for stateless helpers) and
  S13b (extract functions that share a small helper struct).
- **govulncheck first run surfaces CVEs beyond the patch
  bump**. → The CI gate's failure is reviewed case-by-case;
  any post-pump findings become follow-up CVE work tracked
  in `openspec/dead-code.md` or a new CVE file.
- **`Dockerfile.ci` digest pin requires resolving the
  current `golang:1.25.5-alpine3.23` sha256**. → The
  resolution is part of S16; if the digest cannot be obtained
  offline, the slice notes the constraint and pins only the
  tag (with a comment).
- **Mise installer sha256**. → Same: resolved at S16
  implementation time against the GitHub release manifest.
- **Goreleaser hook `set -euo pipefail` may surface other
  shell-quoting bugs**. → The hook script is small; any
  other bugs are tracked as follow-up.
- **Homebrew tap URL depends on goreleaser owner**. → Verify
  the homebrew-tap repo state at S16 implementation; if the
  tap formula already exists under a different owner, the
  migration is a separate homebrew-tap PR.
- **Per-site judgment on 42 fmt.Fprint sites slows S9**. →
  The slice is sized at ~150 LOC mechanical + ~100 LOC
  judgment; per-site decisions are documented inline as
  godoc-style comments on each `Fprintln` call.
- **153 t.Context mechanical change may have unexpected
  side effects**. → S4 is dedicated to in-fn literals only;
  S5 changes helper signatures; S6 changes mock matchers.
  Each is independently revertible.
- **Coverage threshold of 70% may block legitimate
  refactors**. → The threshold is configurable in
  `coverage.sh`; if a refactor requires lowering it
  temporarily, the change notes the temporary nature.
- **25 receiver-less refactor touches many call sites**. →
  S13 is dedicated; the diff is large but each call-site
  change is mechanical via `gopls rename` once the receiver
  is removed.

## Migration Plan

1. **Confirm sequencing**: this change lands AFTER
   `foundation-error-chain` AND `architecture-layer-inversion`
   merge. Sequential, not parallel. The pre-S1 baseline
   includes:
   - Domain error sentinels (`domain.ErrXxxNotFound`, the 6
     shell subtypes, `asType[T]` helper).
   - Layer inversion: `*Impl` suffix dropped in touched files,
     `CompositeGitClient` deleted, `GoGitClient` + `CLIClient`
     injected as separate fields, `RepoLocator` added,
     `domain.ShellWrapper` exposed.
   - Lint config: `gocognit` dropped, `nolintlint` enforced,
     `errcheck.check-type-assertions: true`, `depguard`
     per-layer rules, blanket `Close.*is not checked`
     exclusion replaced by per-line `//nolint`.
2. **Slice S1 absorbs orphans**: create `test/mocks/helpers.go`,
   fix the variadic bug at `cmd_mocks.go:223,233`. Fix the
   `openspec/config.yaml:134` parse error if Change 1's
   task 1.2 hasn't landed.
3. **Slices S2-S17 land in order**, each ending with
   `mise run verify` on the touched packages (and the
   dependent packages).
4. **S17 closes the change**: `openspec validate
   quality-modernization --json` reports `valid: true,
   issues: []`. The 16 spec deltas merge into
   `openspec/specs/<capability>/spec.md` per `archive`'s
   contract.
5. **Rollback**: every slice is a discrete commit (or PR).
   `git revert` from the merge commit restores the prior
   state. The depguard rule addition (S2) and the goreleaser
   owner fix (S16) are the only points where a half-merged
   state breaks the build; staging them at the end keeps
   every earlier slice self-consistent.

## Open Questions

- **Coverage threshold exact value**: the
  `infrastructure-release` spec mandates "≥70%". The
  `coverage.sh` script currently hardcodes the threshold as
  60.0% (which becomes the warning line). S16 sets it to
  70.0% explicitly. If project reality is closer to 65%, the
  threshold may need to drop with a documented rationale.
  Resolved at S16 implementation time by reading
  `coverage_filtered.txt` baseline.
- **`Dockerfile.ci` digest value**: requires `docker pull
  golang:1.25.5-alpine3.23 && docker images --digests` at
  S16 implementation time. The repo may not have Docker
  available; an alternative is `crane digest
  golang:1.25.5-alpine3.23`. Resolved at S16.
- **Mise installer sha256**: requires fetching the GitHub
  release manifest at S16 implementation time. The pinned
  version is the one referenced in the existing
  `.gitlab-ci.yml` `get_mise_version` step.
- **Govulncheck baseline acceptance**: any findings remaining
  after `go get -u=patch` are tracked in
  `openspec/dead-code.md` or a new
  `openspec/cve-baseline.md` with rationale per finding.
  Resolved at S15 implementation time.
- **153 t.Context may include sites that should not
  migrate**: e.g., a sub-test that intentionally shares a
  parent test's context for cancellation propagation.
  S4 includes a per-site review pass before bulk migration;
  any excluded sites are documented inline.
- **Mock matcher swap may not be safe for every site**:
  some mocks may be verifying that the production code
  passes a specific context value (e.g., a context with a
  deadline). S6 review preserves these by using
  `mock.MatchedBy(func(ctx context.Context) bool { ... })`
  with the deadline check, falling back to `mock.Anything`
  only for sites that genuinely don't care.
