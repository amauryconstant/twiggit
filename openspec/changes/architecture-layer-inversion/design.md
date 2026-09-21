# Design: Architecture Layer Inversion

See `proposal.md` for motivation (the eight cross-layer call sites,
the four AGENTS.md files documenting the wrong shape, and the ~30
review findings that live in the same files as the layer
inversion). This document covers the architectural decisions, the
constraints that shape them, and the execution order that gets
the change landed without breaking unrelated domains.

## Context

The current state, after the foundation-error-chain review and
the 19-skill audit surfaced the findings listed in `REVIEW.md`:

- `internal/service/worktree_service.go:22` declares
  `gitService application.GitClient`; the same file at `:15`
  imports `twiggit/internal/infrastructure` solely to reach
  `infrastructure.ExtractProjectFromWorktreePath` (`:374`).
- `internal/service/project_service.go:18` declares
  `gitService application.GitClient`; the same file at `:11`
  imports `twiggit/internal/infrastructure` for five cross-layer
  calls (`FindGitRepositories` at `:68` and `:90`,
  `FindMainRepoByTraversal` at `:250`,
  `ExtractProjectFromWorktreePath` at `:262`, `IsMainRepo` at
  `:269`).
- `internal/service/shell_service_test.go:13` imports
  `twiggit/internal/infrastructure` to construct a real
  `NewShellInfrastructure()` at `:20`, call its `GenerateWrapper`
  three times, and use the captured wrapper text as mock return
  values for `MockShellInfrastructure`.
- `internal/infrastructure/git_client.go` (146 lines) is a pure
  delegating `CompositeGitClient` that wraps `application.GoGitClient`
  + `application.CLIClient`, adding only typed-error wrapping.
- `internal/application/interfaces.go:102-105` declares
  `type GitClient interface { GoGitClient; CLIClient }`.
- `internal/infrastructure/pathutils.go` (96 lines) returns
  `domain.NewContextDetectionError(...)` from seven sites — moving
  this file to `internal/domain/` would create a domain → domain
  cycle. The error types need to change or move.
- `internal/service/worktree_service.go:27` carries a
  `sync.Mutex` field that serializes every mutation globally.
  Spec `application-worktree-management` requires a per-project
  mutex; the implementation does not match the spec.
- `internal/infrastructure/cli_client.go` dereferences
  `result.ExitCode` at six sites before checking
  `result != nil`, which `golang-safety` flagged as a panic risk
  when `ExecuteWithTimeout` returns `(nil, err)`.
- `internal/infrastructure/gogit_client.go:33,55` discards
  `lru.New` errors, leaving a nil cache that panics on first
  access.
- `internal/infrastructure/gogit_client.go:128` carries a
  `_ = remoteRef` workaround for a stale compiler warning.
- `internal/infrastructure/gogit_client.go:336` slices
  `commit.Hash.String()[:7]` without bounding for short hashes.
- `internal/infrastructure/hook_runner.go:137` dereferences
  `cmdResult.ExitCode` before nil-checking `cmdResult`.
- `internal/infrastructure/context_detector.go:22` carries an
  unbounded `map[string]cached` cache flagged as a memory leak.
- `.golangci.yml:99-110` blanket-excludes `gosec`, `wrapcheck`,
  `errcheck`, and friends from `test/**` files; the blanket
  `text: "Close.*is not checked"` exclusion swallows every
  `Close()` error codebase-wide including production.
- `.golangci.yml:87-94` defines `depguard` rules for
  `internal/domain` only; `service/`, `application/`,
  `infrastructure/`, and `cmd/` have no enforced allowlist.
- `internal/service/worktree_service.go:95` discards the
  `hookRunner.Run` error; `:611` discards the
  `gitService.PruneWorktrees` error; both should log via `slog`.
- The `application.GitClient` umbrella, the
  `*Impl`-suffixed implementation types
  (`CLIClientImpl`, `GoGitClientImpl`, `ShellInfrastructureImpl`,
  `HookRunnerImpl`, `ContextDetectorImpl`, `ContextResolverImpl`,
  `ConfigManagerImpl`), and the `DefaultCommandExecutor` name
  all violate the `golang-naming` skill's "drop `Impl` suffix"
  rule and the "return structs, accept interfaces" rule.
- Project targets `go 1.25.5` per `go.mod`; `errors.AsType[T]`
  from Go 1.26 is unavailable. Local helpers are sized for a
  mechanical swap when the toolchain bumps.
- Project rule per `openspec/config.yaml`: tests are written
  AFTER implementation.

## Goals / Non-Goals

**Goals:**

- Eliminate the eight cross-layer call sites in `service/` so
  that `internal/service/` only imports `internal/application/`
  + `internal/domain/`.
- Make the layer rule mechanical (enforced by `depguard` so that
  reverse imports fail to compile at lint time, not at review
  time).
- Delete the `CompositeGitClient` pure-delegating decorator; the
  two role interfaces (`GoGitClient`, `CLIClient`) become the
  direct injection points for `service/` constructors.
- Move filesystem-pure helpers from `internal/infrastructure/`
  to `internal/domain/` so the service layer can call them
  without an interface round-trip.
- Fix the safety bugs in the touched infrastructure files
  (`gogit_client.go`, `cli_client.go`, `hook_runner.go`,
  `context_detector.go`) so the layer inversion lands on a safe
  baseline.
- Drop the `*Impl`-suffix naming anti-pattern in every
  implementation file the layer inversion touches; rename
  `DefaultCommandExecutor` to `CommandExecutor`.
- Land the renames (`PathTypeUnknown` iota shift,
  `IsModified`, `IsInstalled`, `IsSkipped`, `HasExecuted`,
  `IsSuccessful`, `ValidationError.Detail()`) in the same
  change so callers don't see two breaking-change waves.
- Keep the end-user CLI surface stable: every public command
  keeps the same flags, the same output, and the same exit
  codes.

**Non-Goals:**

- Wholesale layer collapse to a Tier 2 CLI shape. The
  `golang-cli-architecture` skill recommends
  `main.go + cmd/ + internal/{core,git,output,iostreams,cmdutil}/`
  for a CLI of this size; the project has 36 specs and a
  well-worn five-layer convention. A future change may collapse
  the layers; this change enforces the existing five.
- Interface Segregation Principle refactor (splitting
  `GoGitClient` and `CLIClient` into 1-method role interfaces).
  The `golang-structs-interfaces` skill flags both as violating
  the 1-3 method rule, but this change keeps the 8-method +
  6-method facet shape and only enforces the layer boundary.
- Moving interfaces into the consumer package
  (`service/`). The skill recommends "define interfaces where
  consumed", but the project's spec
  `application-service-interfaces` owns the interface catalog;
  moving them would force 7 spec migrations in a separate
  change.
- Receiver-less methods → free functions in `service/`. The
  review flagged six methods in `worktree_service.go` and three
  in `project_service.go` that never use their receiver; this
  refactor lives in the quality change.
- `slog` migration across `cmd/`. The review flagged ~30
  `_, _ = fmt.Fprint*` sites in `cmd/`; this change does not
  touch them.
- Hook runner refactor (collapsed no-op result blocks already
  in scope; full deduplication deferred). Context resolver
  `context.Background()` → caller `ctx` propagation deferred.
  `SuggestionOption func(interface{})` typed-option refactor
  deferred.
- Lint threshold tightening (`funlen` 150→120,
  `gocyclo` 25→13). Tightening mid-refactor would destabilize
  the change; deferred to the quality change.
- All godoc comment fixes in untouched domain files.
- CI / Docker / GoReleaser / README / CHANGELOG / `go.mod`
  improvements.
- The per-project mutex requirement that spec
  `application-worktree-management` originally demanded; this
  change removes the requirement because the implementation
  does not satisfy it and the project does not currently need
  it.

## Decisions

### 1. Hard break: delete `CompositeGitClient` and the `GitClient` umbrella

**Choice.** Delete `internal/infrastructure/git_client.go` in its
entirety (the 146-line `CompositeGitClient` struct, its
constructor, and every method). Delete the four-line
`GitClient` interface in `internal/application/interfaces.go`.
Update `WorktreeService`, `ProjectService`, and
`NavigationService` constructors to take
`application.GoGitClient` and `application.CLIClient` as two
separate arguments. Update `main.go` to drop
`infrastructure.NewCompositeGitClient` and pass both clients
directly.

**Rationale.** The composite is a pure-delegating decorator
flagged by `golang-design-patterns` ("composite with no value-add")
and `golang-code-style`. No deprecation alias because the project
ships a single internal binary; `gopls rename` propagates field
accesses across the test tree in one pass. The 0.13.0 hard-break
precedent (foundation-error-chain) confirms the project accepts
breaking API changes when the surface is internal.

**Alternatives considered:**

- Deprecation alias: keep `GitClient` as an interface that the
  two role interfaces satisfy. Adds a marker line in
  `interfaces.go` and one extra step in every constructor. The
  project's "no deprecation" rule rejects this.
- Keep the composite but route via interfaces. Same complexity,
  same call sites, same name collision. Rejected.

### 2. Service constructors take two role interfaces, not a small bundle struct

**Choice.** `NewWorktreeService(goGit application.GoGitClient,
cli application.CLIClient, projectService application.ProjectService,
config *domain.Config, hookRunner application.HookRunner)
application.WorktreeService`. Same shape for `ProjectService`,
`NavigationService`, `ContextResolver`. Two fields, no wrapper.

**Rationale.** The `golang-structs-interfaces` skill recommends
"accept interfaces, return structs" with one consumer per
interface; the role split is the consumer's choice. The two-field
shape mirrors the current `gitService application.GitClient`
field that already lives on each service struct — only the field
count changes from one to two. Wrapping the pair into a bundle
struct (`GitOperations interface { GoGitClient; CLIClient }`)
reintroduces the very umbrella we just deleted.

**Alternatives considered:**

- Bundle struct (`type GitOperations interface { ... }`):
  reintroduces the umbrella at the service level. Rejected.
- Per-method role interfaces (single-method `WorktreeCreator`,
  `BranchDeleter`, etc.): the `golang-structs-interfaces`
  skill recommends it, but the constructor would take 8+ role
  interfaces. Defer to a follow-up "interface segregation"
  change.

### 3. Move `pathutils.go` to `domain/pathutils.go` and drop typed error wrapping

**Choice.** Move `ExtractProjectFromWorktreePath`, `NormalizePath`,
and `IsPathUnder` to `internal/domain/pathutils.go`. Change their
signatures to return plain `error` rather than
`*domain.ContextDetectionError` (because the file's caller chain
never relies on the typed chain — verified by grep).

**Rationale.** The seven `domain.NewContextDetectionError` calls
inside `pathutils.go` exist only because the file used to live in
`internal/infrastructure/`, where the infrastructure error rules
mandated wrapping. In `internal/domain/` those rules do not
apply. Returning plain `error` keeps the helpers pure and lets
the calling service wrap into `WorktreeServiceError` /
`ProjectServiceError` / `NavigationServiceError` as the contract
requires.

**Alternatives considered:**

- Introduce `domain.NewPathResolutionError`. New error type,
  new spec entry, new sentinel. Premature — no caller needs
  typed access. Rejected.
- Leave `pathutils.go` in `internal/infrastructure/` and route
  service through an `application.PathNormalizer` interface.
  Adds two interface definitions and three constructor
  signatures for three trivial functions. Rejected.

### 4. Move `git_utils.go` helpers to `domain/git_repo.go` and reshape `FindGitDirByTraversal`

**Choice.** Move `FindMainRepoByTraversal`, `IsMainRepo`, and
`FindGitDirByTraversal` to `internal/domain/git_repo.go`. Reshape
`FindGitDirByTraversal(startPath string) *string` to
`FindGitDirByTraversal(startPath string) (string, bool)`. Move
the `GitDir` struct to `internal/domain/git_repo.go` as
`type GitDir struct { Name string; Path string }`. Move
`FindGitRepositories` to `internal/infrastructure/repo_finder.go`,
returning `[]domain.GitDir`.

**Rationale.** All three helpers are filesystem-pure. The
`*string → (string, bool)` reshape follows the
`golang-error-handling` skill's "use `(T, bool)` for fallible
operations"; the single caller in `internal/infrastructure/git_utils.go`
is intra-package and updates with the move. The `GitDir` move
unblocks the `application.RepoLocator` interface, which needs
the type without importing `internal/infrastructure/`.

**Alternatives considered:**

- Keep `FindGitDirByTraversal(*string)`. Carries the prior
  smell (a nil pointer as the "not found" signal) into the
  new package. Rejected.
- Move `FindGitRepositories` to `domain/` and inject a
  validator closure. Forces the constructor to provide a
  function value; loses the typed GoGitClient injection. Rejected.

### 5. Add `application.RepoLocator` interface

**Choice.** Add
`type RepoLocator interface { FindGitRepositories(dir string) ([]domain.GitDir, error) }`
to `internal/application/interfaces.go`. The implementation
(`infrastructure.repoFinder`) lives in
`internal/infrastructure/repo_finder.go` and takes the
`application.GoGitClient` for entry validation. `main.go` wires
the concrete implementation; `ProjectService` accepts the
interface.

**Rationale.** `ProjectService` needs filesystem discovery in
two methods (`ListProjects`, `ListProjectSummaries`). Without an
interface, `ProjectService` would either import
`internal/infrastructure/` (defeating the inversion) or
duplicate the discovery logic in `service/` (defeating
testability). The interface lets tests pass a deterministic
locator without touching the filesystem.

**Alternatives considered:**

- Private helper on `ProjectService`: drops testability. The
  quality change adds test reliability work that depends on
  being able to mock discovery. Rejected.
- Pass `func(dir string) ([]domain.GitDir, error)` closure:
  loses the typed `RepoLocator` symbol that callers can name.
  Rejected.

### 6. Move shell wrapper template constant to `domain/shell_wrapper.go`

**Choice.** Move the bash/zsh/fish wrapper template strings out
of `internal/infrastructure/shell_infra.go` into
`internal/domain/shell_wrapper.go` as
`func ShellWrapper(shellType domain.ShellType) (string, error)`.
Service-layer test code calls `domain.ShellWrapper(...)` to seed
`MockShellInfrastructure` expectations instead of constructing a
real `NewShellInfrastructure()`.

**Rationale.** The wrapper template is a domain artifact (a
shell script the user installs), not an infrastructure detail.
The `golang-cli-architecture` skill's "core is pure, shell is
I/O" rule places the script in the core; the file-write
mechanism stays in infrastructure. The
`shell_service_test.go:20` `infrastructure.NewShellInfrastructure()`
call exists solely to capture the wrapper text — once
`domain.ShellWrapper` exists, the test loses its only reason to
import `internal/infrastructure/`.

**Alternatives considered:**

- Hardcode wrapper strings in the test. Drift risk on every
  shell template change. Rejected.
- Add `test/helpers/shell_fixtures.go` with the expected text.
  Moves the constant out of one place into another; same drift
  risk. Rejected.

### 7. Drop the `*Impl` suffix on every implementation type the inversion touches

**Choice.** Rename `CLIClientImpl` → `CLIClient`,
`GoGitClientImpl` → `GoGitClient`, `ShellInfrastructureImpl` →
`ShellInfrastructure`, `HookRunnerImpl` → `HookRunner`,
`ContextDetectorImpl` → `ContextDetector`,
`ContextResolverImpl` → `ContextResolver`,
`ConfigManagerImpl` → `ConfigManager`, and drop the `Impl` from
their `New*` constructors. Rename `DefaultCommandExecutor` →
`CommandExecutor` and `NewDefaultCommandExecutor` →
`NewCommandExecutor`.

**Rationale.** The `golang-naming` skill explicitly flags the
`Impl` suffix as a code smell ("type name carries generic `Impl`
suffix"). Returning the struct named after its interface
implements the skill's "accept interfaces, return structs" rule.
`gopls rename` propagates the change across every caller in the
test tree.

**Alternatives considered:**

- Keep `Impl` and document it. The project has no
  documentation tradition for `Impl`; the rule is universally
  applicable. Rejected.
- Rename only in files the inversion directly touches, leaving
  others for later. Drift between layers. Rejected.

### 8. Drop `sync.Mutex` from `worktreeService`

**Choice.** Delete the `mu sync.Mutex` field at
`internal/service/worktree_service.go:27`. Remove the
`application-worktree-management` requirement that demands a
per-project mutex (separate spec delta in this change).

**Rationale.** The current `mu` field serializes every
mutation globally, which `golang-safety` and
`golang-design-patterns` flagged as dead code (no concurrent
callers). The spec required a per-project mutex; the
implementation provides a global one. Rather than fix the
implementation (out of scope for the layer inversion) we drop
the spec requirement, which the project does not currently
need. A future quality change may reintroduce the
requirement with a real implementation.

**Alternatives considered:**

- Implement per-project mutex map (`map[string]*sync.Mutex`
  guarded by a RWMutex). Real fix to a real gap; would expand
  the change by ~100 lines and require new tests. Rejected for
  scope; deferred to the quality change.

### 9. `NewGoGitClient` and `NewGoGitClientWithSize` return `(*GoGitClient, error)`

**Choice.** Both constructors change from
`func NewGoGitClient(...) *GoGitClient` to
`func NewGoGitClient(...) (*GoGitClient, error)`. The error is
non-nil when the underlying `lru.New` allocation fails. Callers
(`main.go`, tests) propagate the error and exit with
`ExitCodeError` on construction failure.

**Rationale.** The current constructors discard `lru.New`
errors. A nil cache panics on first `Get`. The
`golang-safety` skill flagged this as a silent-corruption risk.
Returning the error costs one line per caller (an `if err !=
nil` propagation) and forces construction-time failure handling.

**Alternatives considered:**

- Keep the silent-nil-cache and rely on the first `Get`
  panic. Defensive nil-check at every `Get` call site. Rejected.
- Move to `func MustNewGoGitClient(...) *GoGitClient` that
  panics on `lru.New` failure. Loses the typed error chain at
  startup. Rejected.

### 10. Add nil-guard before every `result.ExitCode` dereference

**Choice.** Add `if result == nil { return ... }` immediately
before every `result.ExitCode` access in
`internal/infrastructure/cli_client.go` (six sites: lines 115,
151, 176, 198, 220, 247) and `internal/infrastructure/hook_runner.go:137`.

**Rationale.** `command_executor.ExecuteWithTimeout` returns
`(nil, err)` when the binary is missing or the timeout fires.
The current code reads `result.ExitCode` from the nil pointer,
which the `golang-safety` skill flagged as a panic risk. The
fix is mechanical and behavior-preserving: when the executor
returns nil, the existing error is the actionable signal.

**Alternatives considered:**

- Re-architect `CommandResult` to embed the exit code and error
  in a single struct that can never be partially populated.
  Out of scope for the layer inversion. Rejected.
- Trust the caller to check `err != nil` before reading
  `result`. The current call sites read `result` without
  checking; changing every caller requires the same number of
  edits. The nil-guard inside the cli client is safer because
  it survives future call sites. Rejected (kept nil-guard).

### 11. Replace unbounded cache in `context_detector.go` with the existing LRU primitive

**Choice.** Replace the `map[string]cached` field at
`internal/infrastructure/context_detector.go:22` with the
`github.com/hashicorp/golang-lru/v2` package already imported by
`gogit_client.go`. Cache size: 256 (large enough for a developer's
recent CWDs, small enough to bound memory).

**Rationale.** The `golang-safety` skill flagged the unbounded
cache as a memory leak (every distinct CWD adds an entry, no
eviction). The LRU primitive is already in `go.mod` and already
proved out in `gogit_client.go`. One-line swap.

**Alternatives considered:**

- Bound the map by hand (`if len(m) > 256 { delete oldest }`).
  More code, no benefit, harder to test. Rejected.
- Drop the cache entirely. Recompute every detection, accept the
  cost. Out of scope. Rejected.

### 12. Drop blanket `Close.*is not checked` exclusion; add per-line nolints

**Choice.** Remove the
`text: "Close.*is not checked"` exclusion in `.golangci.yml`.
For each `Close()` call site that intentionally discards the
error (typically read-only cleanup), add
`//nolint:errcheck // read-only cleanup` on the line above.

**Rationale.** The `golang-lint` skill flagged the blanket
exclusion as hiding production errors. Per-line `nolint` with
`require-explanation` (added in step 13) forces a deliberate
choice at each call site. Estimate: 5-10 `Close()` call sites
get explicit per-line nolints; the rest either propagate or
become per-line nolint with explanation.

**Alternatives considered:**

- Keep the blanket exclusion and document it. Continues to
  hide production errors. Rejected.
- Add `errcheck.exclude` rules per package instead. Same effect
  as the blanket exclusion, less visible. Rejected.

### 13. `depguard` rules per layer; `nolintlint` enforced

**Choice.** Add `depguard` rules:

- `domain` (existing): allow `$gostd`, `internal/domain`
- `service` (new): allow `$gostd`, `internal/domain`,
  `internal/application`, `internal/service`
- `application` (new): allow `$gostd`, `internal/domain`,
  `internal/application`
- `infrastructure` (new): allow `$gostd`, `internal/domain`,
  `internal/application`, `internal/infrastructure`
- `cmd` (new): allow `$gostd`, `internal/domain`,
  `internal/application`, `internal/service`,
  `internal/infrastructure`, `internal/version`, `twiggit/cmd`

Add `nolintlint` with `require-explanation: true,
require-specific: true`. Add `errcheck.check-type-assertions: true`.
Drop `gocognit` (redundant with `gocyclo` + `nestif` per the
`golang-lint` skill).

**Rationale.** The `golang-lint` skill recommends per-package
allowlists. The new rules turn the documented layer convention
into a lint failure that breaks the build, which is the
mechanical enforcement the `application-service-interfaces`
spec already promises ("Reverse imports SHALL NOT compile").
`nolintlint require-explanation` prevents bare `//nolint`
directives from sneaking back in. Dropping `gocognit` removes
overhead the project does not use.

**Alternatives considered:**

- One global depguard rule listing all allowed imports. Less
  granular, harder to maintain. Rejected.
- Drop depguard, rely on review. The review already missed the
  eight violations; the rule is the fix. Rejected.

### 14. Vertical-slice execution order with slice-scoped test rewrites

**Choice.** Nine vertical slices land in order:

1. Domain expansion (new files, no breakage): `pathutils.go`,
   `git_repo.go`, `shell_wrapper.go`, `GitDir`, `PathTypeUnknown`
   iota shift, `IsModified`, `IsInstalled`, `IsSkipped`,
   `HasExecuted`, `IsSuccessful`, `ErrResult`. Domain tests
   rewrite for the renames.
2. Application interface split: drop `GitClient` umbrella, add
   `RepoLocator`. `application/AGENTS.md` updates.
3. Service layer inversion: `worktree_service.go`,
   `project_service.go`, `context_service.go`, `shell_service.go`,
   `shell_service_test.go` rewrite for the new dependencies.
   Drop `sync.Mutex`, drop unused fields, log swallowed errors
   via `slog`.
4. Infrastructure rewrite: delete `pathutils.go`, `git_utils.go`,
   `git_client.go`, `interfaces.go`; add `repo_finder.go`;
   rename `*Impl` types; fix nil-deref sites in `cli_client.go`,
   `hook_runner.go`; drop `_ = remoteRef`, bound `hash[:7]`;
   change `NewGoGitClient*` signature; swap
   `context_detector` cache to LRU. `os.IsNotExist` →
   `errors.Is` everywhere in touched files. `HasPrefix+TrimPrefix`
   → `CutPrefix` at the two flagged sites.
5. `main.go` rewires: drop `NewCompositeGitClient`, pass two
   clients, update constructor names.
6. Test mocks: `MockGitClientBundle` struct in
   `test/mocks/git_service_mock.go`; rename `*Impl` mocks.
   Mechanical rename across `test/integration/`,
   `test/concurrent/`, `test/e2e/fixtures/`.
7. AGENTS.md sync + `.golangci.yml`: rewrite 5 AGENTS.md files,
   extend `depguard`, drop `gocognit`, add `nolintlint`, add
   `errcheck.check-type-assertions`, drop blanket `Close`
   exclusion, add per-line nolints.
8. Spec deltas (this change): the 6 deltas already in
   `specs/`.
9. Verification: `mise run verify`, `mise run test`,
   `openspec validate architecture-layer-inversion --json`.

**Rationale.** Per project rule (tests after impl), each slice
lands implementation + tests in the same commit. Vertical
slicing means each commit compiles and passes tests in
isolation, so `git bisect` works. The `golang-refactoring`
skill recommends "stacked PRs" for layered refactors; the nine
slices become nine reviewable commits or PRs.

**Alternatives considered:**

- Layer-by-layer (all-domain, then all-service, then
  all-cmd). Bigger blast radius per commit; harder to bisect.
  Rejected.
- One mega-commit. Un-reviewable. Rejected.

### 15. Defer interface segregation and consumer-side interface placement

**Choice.** Keep `GoGitClient` (8 methods) and `CLIClient`
(6 methods) as role interfaces in
`internal/application/interfaces.go`. Keep
`ConfigManager`, `ContextDetector`, `ContextResolver`,
`HookRunner`, `ShellInfrastructure` in
`internal/application/interfaces.go`. Do not move any interface
into `internal/service/`.

**Rationale.** The `golang-structs-interfaces` skill
recommends 1-3 method interfaces defined where consumed. Both
recommendations would force 7+ spec migrations and a wholesale
test rewrite. The user's exploration locked this decision
("Wave 1 = layer inversion; Wave 2 = interface segregation").
A future change can apply both skill recommendations without
risking the layer inversion's stability.

**Alternatives considered:**

- Move `GoGitClient`/`CLIClient` to `service/`. Requires every
  service file to define its own interface; every constructor
  call in `main.go` to use the concrete type. Doubles the diff
  for a marginal testability win. Rejected.
- Split `GoGitClient`/`CLIClient` into 1-method interfaces in
  this change. Constructor takes 8+ role interfaces per
  service. Excessive ceremony. Rejected.

### 16. Bundle struct `MockGitClientBundle` for tests, not separate mocks

**Choice.** `test/mocks/git_service_mock.go` exposes
`type MockGitClientBundle struct { MockGoGitClient *MockGoGitClient; MockCLIClient *MockCLIClient }`.
Service tests construct the bundle and pass both inner mocks
to the service constructor.

**Rationale.** Service tests previously constructed one
`MockGitClient` and passed it as the composite. The composite
deletion forces a two-mock injection. The bundle preserves the
single construction site per test; separate mocks would force
two construction sites and lose the "this test mocks git"
visual cue.

**Alternatives considered:**

- Two separate mock types, constructed independently per
  test. More verbose per test; no benefit. Rejected.
- Generate mocks via `moq`. The project uses `testify/mock`;
  switching mid-refactor expands scope. Rejected.

### 17. Test rewrites land in each slice's commit

**Choice.** Per project rule (tests after impl), each slice
includes its test changes in the same commit (or split commit
with the test commit immediately after the impl commit).
Domain tests for new helpers land in slice 1; service tests
for the new constructors in slice 3; infrastructure tests for
the safety fixes in slice 4; test/mocks updates in slice 6.

**Rationale.** The `openspec/config.yaml` rule says tests
written AFTER implementation. Vertical slices are the smallest
unit that satisfies "implementation + tests" together. The
`golang-testing` skill recommends `t.Cleanup(mock.AssertExpectations)`
in every test; that work lives in the quality change, not here.

**Alternatives considered:**

- Single test PR after all implementation. Defers regression
  risk until the end; harder to bisect. Rejected.

### 18. Spec drift on `application-worktree-management` is resolved by deletion

**Choice.** Delete requirement 2 ("Per-project worktree
mutation mutex") from `application-worktree-management`. The
implementation has a single struct mutex that does not satisfy
the spec; the project does not currently need per-project
mutexes; the implementation does not satisfy the spec today
or after this change.

**Rationale.** The `golang-design-patterns` skill flags the
single struct mutex as dead code. The spec drift is a
contradiction (spec says per-project, code says global). The
least-disruptive resolution is to drop the requirement. A
future change can reintroduce it with a real implementation.

**Alternatives considered:**

- Implement the per-project mutex in this change. Real work,
  ~100 lines + tests. Out of scope. Rejected.
- Leave both as-is. Spec drift remains. Rejected.

## Risks / Trade-offs

- **Mechanical rename across test files.** Dropping `*Impl`
  touches every integration / concurrent / e2e fixture that
  references `infrastructure.NewCLIClientImpl` etc. →
  Mitigation: `gopls rename` per file; one `mise run test`
  per slice catches missed call sites.

- **Constructor signature change for `NewGoGitClient`.** →
  Mitigation: `main.go` is the single production caller;
  tests are 3-5 files; mechanical update with the rename
  tool.

- **Depguard rule ordering.** Adding the per-layer rules
  BEFORE the slice that removes the offending imports would
  fail the lint mid-PR. → Mitigation: slice 7 (rules) lands
  after slices 3-5 (which remove the offending imports).

- **`MockGitClientBundle` test ergonomics.** Service tests
  that previously passed one mock now pass a bundle struct.
  → Mitigation: the bundle struct keeps a single construction
  site; mechanical update.

- **`domain.Result[T]` `NewErrorResult → NewErrResult` rename.**
  Every call site updates. → Mitigation: `gopls rename`
  catches all; mechanical in slice 1.

- **`PathType` integer shift.** Any caller using the
  underlying integer value (`if x == 1`) breaks. → Mitigation:
  grep for `PathType` integer literals; only the `String()`
  consumers exist. `PathType.String()` is the public surface.

- **`ValidationError.Detail` rename.** Callers of `Context()`
  on a `*ValidationError` need to switch to `Detail()`.
  → Mitigation: the original `Context()` getter on
  `ValidationError` was the only collision; grep confirms.

- **Wholesale layer collapse is deferred.** A future change
  may collapse to Tier 2; this change enforces the existing
  five-layer convention harder. → Mitigation: the layer rules
  in `depguard` are easy to relax when the collapse happens.

- **Hard-break loses external consumers.** → Mitigation:
  the package is `internal/` to the repo; no third-party
  importer exists. Confirmed via grep of the repo tree.

- **Per-line `Close()` nolint.** Adding `require-explanation`
  on `nolintlint` may flag existing nolint directives. →
  Mitigation: the codebase has few existing nolint directives
  (the four `//nolint:wrapcheck` lines in
  `command_executor_mock_test.go`); each one gets a per-line
  explanation in slice 7.

## Migration Plan

1. Snapshot the existing `mise run test` baseline so we have a
   known green before any code edits.
2. Apply slices 1-9 in order. Each slice ends with
   `mise run verify` on the touched package alone (and the
   dependent packages).
3. After slice 9 (verification) lands, run the full
   `mise run test` once more. Failing tests at this point
   identify rename-mechanical-update gaps.
4. Archive the change with `openspec-archive-change
   architecture-layer-inversion` after the implementation PR
   merges. The 6 spec deltas merge into
   `openspec/specs/<capability>/spec.md` per `openspec-archive-change`'s
   contract.
5. **Rollback:** every slice is a discrete commit (or PR);
   `git revert` from the merge commit restores the prior state.
   The `depguard` rule addition is the only point at which a
   half-merged state breaks the build; staging the rule in
   slice 7 keeps every earlier slice self-consistent.

## Open Questions

None. The 18 decisions above are settled; the only deferred
items (interface segregation, consumer-side interfaces,
receiver-less free functions, modern `slog` migration in cmd/,
godoc fixes, CI/Docker/GoReleaser improvements) are explicitly
out of scope and do not change the layer-inversion contract.
Any future question "should we add a sentinel for X" or "should
the per-project mutex be reintroduced" is captured as a
follow-up change once a caller needs it.
