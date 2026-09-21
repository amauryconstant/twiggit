# Design: Sentinel-based error chain foundation

See `proposal.md` for motivation (the four-way drift between code, spec,
doc, and mock). This document covers the architectural decisions, the
constraints that shape them, and the execution order that gets the change
landed without breaking unrelated domains.

## Context

The current state, after twenty Golang skill audits surfaced the
findings listed in `REVIEW.md`:

- `internal/domain/errors.go` defines `GitRepositoryError`,
  `GitWorktreeError`, `GitWorktreeError`, `ConfigError`, and
  `ContextDetectionError`. Each has an `IsNotFound() bool` method that
  scans `Message` for substrings (`"not found"`, `"does not exist"`,
  `"no such file or directory"`). Same pattern repeats on
  `WorktreeServiceError`, `ProjectServiceError`,
  `NavigationServiceError`, and `ResolutionError` in
  `internal/domain/service_errors.go`.
- `internal/domain/shell_errors.go` carries `ShellError.Code string`
  plus nine `const`-declared `Err*` string keys
  (`ErrShellAlreadyInstalled`, `ErrShellNotInstalled`,
  `ErrInvalidShellType`, `ErrConfigFileNotFound`,
  `ErrConfigFileNotWritable`, `ErrWrapperGeneration`,
  `ErrWrapperInstallation`, `ErrInferenceFailed`,
  `ErrShellDetectionFailed`). Callers compare with `==`
  (`internal/service/shell_service.go:105,141`).
- `cmd/error_handler.go:113-125` checks `IsNotFound()` on three of the
  six types; `:166-168` falls back to substring matching against
  `err.Error()`; `:175-195` matches cobra errors via a seven-substring
  list.
- `cmd/error_formatter.go:113-117,138-142,159-164,177-181` use a
  `func() *T { target := &T{}; _ = errors.As(err, &target); return
  target }()` IIFE; review flags this as a latent nil-deref.
- `internal/service/AGENTS.md` documents the substring pattern as the
  "right" idiom and `cmd/AGENTS.md` documents the IIFE pattern as the
  recipe. Two source-tree doc files actively teach the bug.
- `test/mocks/cmd_mocks.go:222-238` calls `m.Called(partial, opts)` on a
  method with signature `opts ...domain.SuggestionOption`. `opts`
  reaches `m.Called` as a single slice argument; mocks never match the
  registered expectations.
- `openspec/config.yaml:134-140` has a list item with embedded comment
  text that the YAML parser does not accept; every `openspec`
  invocation prints `Warning: could not parse ... ignoring it`.
- `openspec/changes/spec-restructure-all-categories/` is an empty
  folder from a previous attempt.
- Project targets `go 1.25.5` (per `go.mod`); `errors.AsType[T]` from
  Go 1.26 is not available. The local helper is sized to slot in cleanly
  when the toolchain bumps.
- Project rule per `openspec/config.yaml`: tests are written AFTER
  implementation.

## Goals / Non-Goals

**Goals:**

- Replace substring-based error identification with `errors.Is` walks
  against deterministic sentinels.
- Eliminate the four-way drift (code ↔ spec ↔ AGENTS.md ↔ mock) so any
  one source of truth changing forces the others to update.
- Land the fix in a way that proves itself per slice (vertical-slice
  order) before compounding to the rest of the taxonomy.
- Keep the end-user CLI surface stable: non-usage failures still exit 1,
  usage errors still exit 2, panic recovery still exits 1.
- Reduce CLI exit-code dispatch from seven codes to three (0/1/2).
  Per-resource NotFound distinctions move to sentinel-driven hint
  discrimination in the formatter (`cli-error-formatting`'s Actionable
  hints requirement). Aligns with the `golang-cli-architecture` skill's
  exit-code discipline (sysexits-safe range; consumer-script reality).

**Non-Goals:**

- Layer inversion (`internal/service` ↛ `internal/infrastructure`),
  modernization sweep (`for i := range N`, `strings.CutPrefix`, etc.),
  and dead-config-field removal. Each is a separate follow-up change.
- Granular exit codes (3-6). Brought into the change's scope as a Goal;
  reverse-out of any prior change that reintroduced them.
- A `ShellError` interface in the `domain` package. The package rule
  permits only types and errors (`internal/domain/AGENTS.md`); seven
  concrete subtypes without a shared interface stays consistent with
  that rule and avoids polymorphic dispatch where none is needed.
- Method `Is(target error)` returning `false` for any target other than
  this type's own sentinel. Doing so would silently absorb unrelated
  sentinels.

## Decisions

### 1. Sentinel participation via `Is(target error) bool` + `Unwrap() error`

**Choice.** Each of the twelve domain wrapper types (six service
errors + six shell subtypes) implements both
`Is(target error) bool` and `Unwrap() error`. `Is(target)` returns
`target == <this type's sentinel>`. `Unwrap()` returns the `Err`
field. `ValidationError` returns `nil` from `Unwrap()` (terminal).

**Rationale.** `errors.Is(err, target)` walks the chain by calling
`Is(target)` first, then `Unwrap()` on the result. Walking the chain
through `Unwrap()` alone is insufficient when a sentinel needs to be
visible despite a wrapped `Cause` (which is the common case after this
change). Walking through `Is` alone breaks the chain for any wrapped
cause a caller might want to inspect. Combining the two produces a
chain that satisfies both `errors.Is(err, ErrXNotFound)` and
`errors.Is(err, originalCause)` simultaneously.

**Alternatives considered:**

- **`Unwrap()` returning the sentinel when `Cause` is nil.** Rejected.
  When `Cause` is set (the production case), `errors.Is` walks past the
  sentinel and never matches.
- **`Unwrap()` returning `errors.Join(sentinel, Cause)`.** Rejected.
  The `Error()` formatter then renders both, producing noisy multi-line
  messages for what was previously one line.
- **Separate `NotFound` constructor (`NewWorktreeNotFound`).**
  Rejected. Forces callers to choose a constructor before knowing which
  fields they have, and double the number of constructors per type.

### 2. Per-resource NotFound sentinels drive formatter hint discrimination

**Choice.** Four resource-specific sentinels
(`ErrGitRepoNotFound`, `ErrWorktreeNotFound`, `ErrProjectNotFound`,
`ErrResolutionNotFound`) all map to `ExitCodeError` (1) at the cmd
boundary. The `cmd/error_formatter.go` hint table discriminates between
them via `switch errors.Is(err, <sentinel>) { ... }`, returning the
resource-appropriate hint (e.g., `twiggit list --all` for
`ErrProjectNotFound`, `twiggit list` for `ErrWorktreeNotFound`).

**Rationale.** Granular exit codes (3-6) were removed per the project's
exit-code discipline. The four NotFound categories remain distinct for
both programmatic matching (`errors.Is`) and user-facing hint
discrimination, but they collapse to a single, single-bit
success/failure signal at the OS level. Operators get better hints
without growing the exit-code table past `golang-cli-architecture`'s
3-code contract.

**Alternatives considered:**

- **One generic `ErrNotFound`.** Rejected. Loses hint precision; cmd
  layer cannot choose a resource-specific hint without re-introducing
  typed `errors.As` chains that we'd be trying to delete.

### 3. ShellError as seven concrete subtypes with shared base struct

**Choice.** Seven structs
(`ShellAlreadyInstalledError`, `ShellNotInstalledError`,
`ShellInvalidTypeError`, `ShellInferenceError`,
`ShellDetectionError`, `ShellWrapperError`, `ShellConfigError`)
embed a `shellErrorBase struct { ShellType, Context, Err }` and each
implements its own `Is(target)` and `Unwrap()`. The `Code string`
field is removed; `ErrConfigFileNotWritable` (no live callers per
review §14) is deleted.

**Rationale.** Concrete subtypes make each shell failure mode
individually matchable via `errors.As`, which is what the formatter
pattern requires for typed extraction. Sharing the base struct keeps
the per-type boilerplate down without reintroducing a
`domain.ShellError` interface (which would violate the
`internal/domain` package rule against interfaces). The dead
`ErrConfigFileNotWritable` is removed at the same point rather than
left as a future footgun.

**Alternatives considered:**

- **One `ShellError` struct with a `Kind` enum (per `golang-design-patterns`).**
  Rejected. Brings back the string-equality path that this change deletes
  (`shellErr.Kind == KindAlreadyInstalled`). The seven subtypes trade
  ergonomics for type-driven dispatch, which is the whole point of the
  sentinel + `errors.As` pattern adopted in this change.
- **A `domain.ShellError` interface implemented by seven subtypes.**
  Rejected. The package rule permits only types and errors in
  `domain/`; adding an interface here forces a follow-up rule
  violation that Change 2 (architecture) would then need to undo.

### 4. `Cause error` → `Err error` rename happens in this change

**Choice.** All thirteen domain error types rename their wrap-field
from `Cause error` to `Err error`, including constructor parameters
where they are exported, plus the `Unwrap()` body, plus every read
site (`e.Cause` becomes `e.Err`).

**Rationale.** Stdlib uses `Err error` on every comparable wrapper
(`*os.PathError.Err`, `*net.OpError.Err`, `*fs.PathError.Err`,
`*url.Error.Err`). Holding the project on `Cause` for one more release
means a rename lands as part of the next non-error-chain change,
mixing concerns. Doing it here while the rest of the file is being
touched is mechanical and `gopls rename` keeps call sites in sync.

**Alternatives considered:**

- **Defer to the follow-up quality change.** Rejected. Multiplies
  touch-points: the error-chain change already touches every error
  type file, and the quality change will touch most of them again for
  unrelated reasons.

### 5. Sentinel messages include the `domain:` package prefix

**Choice.** All twelve sentinel messages are
`"domain: <resource> <state>"` (e.g.,
`errors.New("domain: worktree not found")`,
`errors.New("domain: shell wrapper already installed")`).

**Rationale.** The `golang-naming` skill explicitly recommends
package-prefixed sentinel messages. Stdlib precedent (e.g.,
`redis.Nil = errors.New("redis: nil")`) reinforces it. The previous
project-wide convention (no prefix) lives on existing `Error()`
strings and is not regressed; new sentinels adopt the better style.
Low-cardinality messages stay stable per `golang-error-handling`
guidance — the prefix is constant, the `<state>` is bounded to a
finite enum.

**Alternatives considered:**

- **No prefix (`errors.New("worktree not found")`).** Rejected.
  Harder to identify origin in logs that contain multiple packages'
  errors; fights the naming skill's explicit recommendation.

### 6. Private `asType[T error](err error) (T, bool)` helper in cmd

**Choice.** A private generic
`func asType[T error](err error) (T, bool) { var target T; if
!errors.As(err, &target) { return target, false }; return target,
true }` lives in `cmd/error_formatter.go`. All four formatters call
it instead of the IIFE pattern. Same name as Go 1.26's
`errors.AsType[T]` so the migration is mechanical.

**Rationale.** Go 1.26 ships `errors.AsType[T](err) (T, bool)` which
has identical semantics. With `go 1.25.5` locked in `go.mod`, the
feature is unavailable today; a local helper is a single ten-line
file. When the toolchain bumps to 1.26, the formatters swap their
callers without changing the call-site shape.

**Alternatives considered:**

- **Inline `errors.As` at every formatter call site.** Rejected.
  Repeats the target-type declaration per site and re-introduces the
  nil risk the IIFE pattern already had.
- **`internal/cmdutil/errors.go` so other layers can adopt later.**
  Rejected for now. No other layer needs it in this change; premature
  generalization. Promotes easily on first need.

### 7. Execute as vertical slice, not layer-by-layer

**Choice.** Land `WorktreeServiceError` end-to-end first (sentinel,
`Is`, `Unwrap`, renamed `Err`, deleted old `IsNotFound`, rewritten
test for that one type), validate that `cmd/delete.go:102-104`
behaves correctly with `errors.Is`, then replicate the pattern
mechanically for `GitRepositoryError`, `GitWorktreeError`,
`ProjectServiceError`, `NavigationServiceError`,
`ResolutionError`, and the seven shell subtypes. Service-layer
(`shell_service.go`), cmd-layer (`error_handler.go`,
`error_formatter.go`), and mock-layer (`cmd_mocks.go`) changes land
after the domain layer proves itself.

**Rationale.** Per project rule (`openspec/config.yaml`), tests ship
AFTER implementation. A layer-by-layer commit wave means every test
that asserts on the new behavior doesn't exist yet when the code
lands. Vertical slicing keeps the size of each unverified chunk
small and gives a concrete shape (`WorktreeServiceError`,
`ErrWorktreeNotFound`, `cmd/delete.go:102-104`) before
the pattern replicates.

**Alternatives considered:**

- **Layer-by-layer (all-domain, then all-service, then all-cmd).**
  Rejected. Larger blast radius per commit; harder to bisect if
  `Is()` semantics prove wrong.

### 8. ValidationError: terminal, no emoji, plain `Error()` text

**Choice.** `ValidationError.Error()` returns the existing message
format without the `💡` glyph on each suggestion. The formatter
prefixes suggestions with `Hint:` as it does today. Add
`func (e *ValidationError) Unwrap() error { return nil }` to satisfy
the spec requirement (`domain-typed-errors`) that all domain error
types implement `Unwrap() error`.

**Rationale.** The project rule "no emoji in code" (per
`cmd/AGENTS.md` style guidance and the `golang-naming` skill's
"error strings are fully lowercase — including acronyms") rejects
emoji in user-facing strings. The formatter already prefixes each
suggestion with `Hint:` so the emoji added no information. The
`golang-naming` skill also requires error strings to carry no
trailing punctuation; `ValidationError.Error()` SHALL emit the
message text without a trailing `.` (and without any other terminal
punctuation). Terminal `Unwrap()` is correct: `ValidationError` has
no `Cause` field and its chain ends at itself.

**Alternatives considered:**

- **Move emoji to the formatter.** Rejected. Lets the formatter
  decide emoji vs plain text, but also lets callers decide emoji, and
  that is not a CLI-owned concern. Domain emits clean text.

### 9. Setup fixes bundled into the implementation tasks

**Choice.** Fixing the `openspec/config.yaml` parse error (move the
`# Purity principles...` comment out of the list item) and deleting
the empty `openspec/changes/spec-restructure-all-categories/`
folder are the first two implementation tasks, before any code
changes.

**Rationale.** The parse error prints on every `openspec` CLI
invocation and was confirmed by `openspec list --json` at the
start of the change. Leaving it set forces the team to read warnings
for the rest of the work. The empty folder is unattributed content
artifact from a prior attempt; removing it is one-line cleanup that
prevents future confusion.

**Alternatives considered:**

- **Defer setup fixes to a separate change.** Rejected. Splits a
  one-line move plus one `git rm -r` across two PRs for no benefit;
  the warnings remain a pain point during this change's development.

## Risks / Trade-offs

- **Forgetting the `Is()` method on a new wrapper leaks the bug.**
  Mitigation: per-type test table in
  `internal/domain/errors_test.go` and the corresponding sentinel-walk
  table in `cmd/error_handler_test.go` verify each wrapper's `Is()`
  participates before the per-resource hint switch is trusted.
- **`Cause` → `Err` rename touches every error-type file and every
  test that reads the field directly.** Mitigation: `gopls rename`
  propagates the rename across all packages and refuses to leave
  dangling field accesses. Build + `mise run lint` catches anything
  it misses.
- **Internal scripts and tests key on exit codes 3-6 today.**
  Mitigation: `cmd/error_handler_test.go:18-98` and
  `test/e2e/list_test.go:180,199` collapse to assertions on
  `ExitCodeError` (1) and `ExitCodeUsage` (2). `cmd/AGENTS.md:58-69`
  and `AGENTS.md:233-237` drop rows for codes 3-6. `mise run verify`
  (task 8.1) catches any consumer we missed.
- **`internal/service/AGENTS.md` and `cmd/AGENTS.md` re-drift if
  next change authors repeat patterns without checking the spec.**
  Mitigation: spec catalog at `openspec/specs/domain-typed-errors/spec.md`
  is the single source-of-truth requirement table for sentinel
  participation; CI does not currently enforce spec conformance, so
  this is a review-process risk rather than a compile-time one.
- **Multi-package reuse of `asType[T]` likely arises in Change 2
  (architecture) once Service tests adopt it.** Mitigation: helper is
  private to `cmd/` for now; promote to `internal/cmdutil/` on first
  cross-package need.
- **Hard break loses external consumers.** Mitigation: this package
  is `internal/` to the repo; no third-party importer exists.
  Confirmed via grep of the repo tree.
- **Test rewrite waves land after every code commit during
  implementation.** Mitigation: vertical-slice task order means the
  rewrite for one type precedes the rewrite for the next; the per-type
  test provides immediate regression protection.

## Migration Plan

1. Snapshot the existing `mise run test` baseline so we have a known
   green before any code changes.
2. Apply tasks in the order listed under the proposal's "Setup" plus
   "vertical-slice" sequence. Each task ends with `mise run verify`
   (lint + race + golden) on the touched package alone.
3. After all vertical slices complete, run the full `mise run test`
   once more. Failing tests at this point identify sentinel-walk
   coverage gaps from per-slice rewrites.
4. The version bump that ships the hard-break surface is deferred to the
   next change that owns a release-pinned tag. The hard-break surface is:
   `Cause`→`Err` field rename on thirteen existing wrapper types, deletion
   of `IsNotFound() bool` on six types, and `ShellError` split into seven
   concrete subtypes. The exit-code alignment to 3-code dispatch is a
   consumer-visible policy change but does not add new API surface;
   scripts keyed on codes 3-6 break by design. `internal/version/version.go`
   stays at the build-time-injected `dev` value for this change. Regenerate
   `CHANGELOG.md` via `openspec-generate-changelog` after
   `openspec-archive-change`.
5. **Rollback:** every step in this change is a discrete commit;
   `git revert` from the merge commit restores the prior state. The
   setup fixes (config.yaml + folder deletion) are reversible with
   `git revert` + folder restoration.

## Open Questions

None. Decisions 1 through 9 are settled; any future question
"should we add a sentinel for X" is captured as a follow-up change
once we see a caller that needs it. The per-resource NotFound
granularity is the largest open axis, and the four-sentinel hint
discrimination plan is committed.
