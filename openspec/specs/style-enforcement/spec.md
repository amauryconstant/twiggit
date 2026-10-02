# Capability: Style Enforcement

## Purpose

Project-wide code conventions codified as testable Requirements. This spec
is the authority for must/must-not language; `openspec/AGENTS.md` and the
`golang-*` skills mirror it for situational awareness. Each bracketed key
in `openspec/AGENTS.md` (other than orchestrator keys) maps 1:1 to a
Requirement here. Mechanical enforcement lives in `.golangci.yml`;
judgment enforcement happens in code review and the AI agent's apply phase.

## Requirements

### Requirement: Identifier casing follows MixedCaps

Identifiers SHALL use `MixedCaps` or `mixedCaps`; underscores in identifier
names are prohibited. Acronyms SHALL be cased consistently within the same
identifier (`HTTPServer`, not `HttpServer`; `userID`, not `userId`).

#### Scenario: Underscored identifier is rejected

- **WHEN** a new identifier is introduced in `internal/`, `cmd/`, or `test/`
- **THEN** `go vet` and `golangci-lint` SHALL reject any identifier whose
  name contains an underscore

### Requirement: Names do not stutter package qualifiers

A name SHALL NOT repeat information already carried by its package qualifier
or surrounding scope. Interfaces exported from `internal/core/git` SHALL
NOT begin with `Git`; constructors on `core.BranchName` SHALL NOT be
named `core.NewBranchNameOfBranch`. The package path is part of the name.

#### Scenario: Role interface names omit the package qualifier

- **WHEN** the six role interfaces declared by `core-git` are enumerated
- **THEN** no role SHALL begin with `Git` followed by the role's noun
  (`GitBranchReader` is forbidden; `BranchReader` is correct)

### Requirement: Enum-typed fields declare an Unknown sentinel at iota 0

Every enum-typed value object (`ShellType`, `ContextType`, `PathType`,
`HookType`, and any future discriminated value) SHALL declare an explicit
`Unknown` or `Invalid` variant at `iota` position 0. The zero value of
the type SHALL NOT collide with a valid value so that an uninitialized
field is detectable at runtime and serialization.

#### Scenario: Unknown sentinel is the zero value

- **WHEN** a `core.ShellType` is declared without explicit initialization
- **THEN** the zero value SHALL equal `ShellTypeUnknown` and SHALL NOT be
  treated as a usable shell type by `git-shell-detect`

### Requirement: Error symbols follow prefix-suffix conventions

Sentinel error variables SHALL use the `Err` prefix
(`ErrBranchNotFound`, `ErrShellAlreadyInstalled`). Error types SHALL use
the `Error` suffix (`*OperationError`, `*ValidationError`). Constructors
SHALL use the `New` prefix (`NewValidationError`, `NewBranchName`). The
same rule applies to all packages under `internal/core/`.

#### Scenario: Naming is enforced by review

- **WHEN** a new sentinel, error type, or constructor is introduced
- **THEN** its name SHALL match `^Err[A-Z][A-Za-z0-9]*$`,
  `^Error$`, or `^New[A-Z][A-Za-z0-9]*$` per its kind

### Requirement: Interfaces are declared at the consumer side

Interfaces SHALL be defined in the package that consumes them, not the
package that produces the concrete implementation. Concrete types SHALL
return structs; consumer packages SHALL define the minimal interface they
require.

#### Scenario: Role interfaces live in the consumer package

- **WHEN** `core-git` declares the six role interfaces (`RepositoryOpener`,
  `BranchReader`, `RepositoryReader`, `RemoteReader`, `WorktreeWriter`,
  `BranchWriter`)
- **THEN** `internal/git/client.go` SHALL satisfy those interfaces via
  compile-time `var _ T = (*Impl)(nil)` drift sentinels, NOT by
  re-declaring them

### Requirement: Role-interface method sets are guarded by a drift sentinel

Role interfaces declared in `core-git` SHALL have their method sets
enforced by a compile-time drift sentinel (`var _ T = (*Impl)(nil)`) in
the implementing package. Signature changes in the interface SHALL fail
the build until every implementing package is updated.

#### Scenario: Drift sentinel blocks stale implementations

- **WHEN** a method is removed from the `BranchReader` interface in
  `core-git`
- **THEN** the `var _ BranchReader = (*Client)(nil)` sentinel in
  `internal/git/client.go` SHALL fail compilation

### Requirement: Returned errors are always checked

Every function that returns an error SHALL have its error either
returned, logged, or explicitly assigned to a named variable. Discarding
with `_` is prohibited except in `defer` cleanups where the secondary
error is genuinely uninteresting.

#### Scenario: Discarded error is a lint violation

- **WHEN** `golangci-lint` runs with `errcheck` enabled
- **THEN** any `_, _ := f()` or `_, _ = f()` form SHALL fail the lint

### Requirement: Wrapped errors preserve context

Error chains SHALL use `fmt.Errorf("context: %w", err)` so the chain retains
both the originating error and the operation context. The exceptions are
the four canonical error types (`*core.ValidationError`, `*core.UsageError`,
`*core.NotFoundError`, `*core.OperationError`) which are returned unwrapped
because their constructors already carry context.

#### Scenario: Wrapping pattern is enforced

- **WHEN** a function returns an error originating from a third-party or
  external call
- **THEN** the wrapping SHALL match `fmt.Errorf("op: %w", err)` form so
  `errors.Is` and `errors.AsType` can inspect the chain

### Requirement: Error messages are lowercase without trailing punctuation

Error strings SHALL be lowercase, SHALL NOT end with `.`, `!`, or `?`, and
SHALL NOT duplicate context that the wrapping layer adds. Log records
SHALL likewise carry lowercase messages and lowercase attribute keys.

#### Scenario: Format violation is caught by review

- **WHEN** an error message ends with a period or carries a capitalized
  first letter
- **THEN** the AI agent's apply phase SHALL reject the change with the
  precise rule violated

### Requirement: Sentinel matching uses errors.Is; typed walks use errors.AsType[T]

Error inspection SHALL use `errors.Is` for sentinel matching and
`errors.AsType[T]` (Go 1.27+ generic API) for typed chain inspection.
Substring matching on `err.Error()` is prohibited. Legacy `errors.As(err,
&*core.X{})` form SHALL be migrated to `errors.AsType[*core.X](err)`.

#### Scenario: Modern inspection API is preferred

- **WHEN** a function inspects an error chain for a known typed error
- **THEN** it SHALL use `errors.AsType[*core.OperationError](err)` rather
  than `errors.As(err, &*core.OperationError{})` or string matching

### Requirement: Each error is handled exactly once

An error SHALL be either logged OR returned, never both. Adapter code
in `internal/git/` returns wrapped errors only; logging happens at the
cmd boundary via the per-command logger. The `core.ValidationError` and
`*core.UsageError` are returned unwrapped and never logged before
return.

#### Scenario: Boundary logging is the single handling site

- **WHEN** an adapter function returns a wrapped error
- **THEN** the adapter SHALL NOT also log that error; the cmd boundary
  is the single logging site

### Requirement: Expected failures return errors; panic is for programmer errors only

`panic` SHALL NOT be used for expected error conditions. `panic` is
reserved for programmer errors (impossible invariants, failed assertions
on must-be-true conditions) and `Must*` constructors whose contract
guarantees non-failure. Any expected failure path SHALL return an error
and let the caller decide how to render it.

#### Scenario: Expected failure returns error

- **WHEN** a function encounters an expected failure (validation,
  not-found, usage misuse)
- **THEN** it SHALL return one of the four canonical error types or a
  wrapping error; `panic` SHALL NOT be the cause

### Requirement: defer Close() follows successful resource acquisition

`defer Close()` SHALL be placed immediately after a successful resource
acquisition (file, network connection, `*exec.Cmd`, `*go-git.Repository`,
logger). Deferred close SHALL NOT be placed before the acquisition check
so resources are not closed that were never opened.

#### Scenario: Placement is enforceable

- **WHEN** a `*go-git.Repository` is acquired through `git.PlainOpen` in
  any reader
- **THEN** `defer repo.Close()` SHALL be placed immediately after the
  acquisition line, before any other statement on the happy path

### Requirement: External calls carry a context deadline

Every external call (git subprocess, shell command, filesystem walk, hook
invocation) SHALL carry a caller-supplied or package-default deadline
(`context.WithTimeout` or `BulkDeadline`). Subprocess cancellation SHALL
send `Process.Kill()` so partially-blocked calls release descriptors.

#### Scenario: Cancellation propagates

- **WHEN** a hook subprocess exceeds `Config.Shell.HookTimeout` seconds
- **THEN** the runner SHALL send `Process.Kill()` and record the failure
  with `TimedOut: true` so the next hook still runs

### Requirement: Type assertions use comma-ok form

Type assertions SHALL use `v, ok := x.(T)` so a runtime mismatch returns
`ok=false` rather than panicking. Bare `v := x.(T)` is prohibited except
in `Must*` constructors whose contract guarantees the assertion succeeds.

#### Scenario: Bare assertion is a lint violation

- **WHEN** `golangci-lint` runs with `staticcheck` enabled
- **THEN** bare `x.(T)` form SHALL fail the lint

### Requirement: Maps are initialized before write

Maps SHALL be initialized (`make(map[K]V)`) before any write. Writing to
a nil map panics. Reading from a nil map is permitted and returns the
zero value.

#### Scenario: Nil-map write is a panic

- **WHEN** code attempts `m[k] = v` on an uninitialized `m`
- **THEN** the runtime SHALL panic with `assignment to entry in nil map`

### Requirement: Exported slice/map returns are defensive copies

Exported functions returning slices or maps whose backing storage is
shared with the caller MUST return defensive copies (`slices.Clone`,
`maps.Clone`). Incidental per-call returns (constructed within the
function and not referenced elsewhere) SHOULD return defensive copies
for reviewer safety.

#### Scenario: Cross-defensive copies prevent aliasing

- **WHEN** a function returns a `[]core.WorktreeInfo` whose backing
  array is shared with internal cache state
- **THEN** the return SHALL be a `slices.Clone` of the source slice so
  the caller cannot mutate cache via the return

### Requirement: Maps are not accessed concurrently without synchronization

Maps SHALL NOT be accessed concurrently without external synchronization
(`sync.Mutex`, `sync.RWMutex`) or use of `sync.Map`. The `git.client`
LRU cache SHALL document its locking strategy. Concurrent read-only
access is permitted only when the map is not being written.

#### Scenario: Concurrent write is a race

- **WHEN** two goroutines write to the same map without a lock
- **THEN** `go test -race ./...` SHALL flag the data race

### Requirement: init() is avoided in favor of explicit lazy constructors

`init()` SHALL NOT be used for package-level state setup. Use explicit
constructors (`NewX(...)`) or lazy initialization via
`sync.OnceValue[T]` (single-value) or `sync.OnceValues[T]` (multi-value).
The `internal/cmdutil.Factory` lazy fields use `sync.OnceValues` so the
config is loaded at first use and the result is cached.

#### Scenario: init() is reviewable

- **WHEN** a new package needs state shared across instances
- **THEN** it SHALL provide a constructor returning a struct holding the
  shared state; `init()` SHALL NOT be used

### Requirement: Test preconditions use require; verifications use assert

Testify's `require` SHALL be used for preconditions (setup, error checks,
file existence). `assert` SHALL be used for verifications on the result.
Mixing randomly is prohibited because failing preconditions under `assert`
produces noisy downstream diffs and masked failures.

#### Scenario: Precondition failure aborts the test

- **WHEN** `require.NoError(t, err)` fails inside a setup step
- **THEN** the test SHALL abort immediately rather than produce a noisy
  cascade of subsequent `assert` failures

### Requirement: Testify argument order preserves (expected, actual)

Testify's `require.Equal(t, expected, actual)` and equivalent forms
SHALL preserve the `(expected, actual)` argument order so diff output
reads top-to-bottom (`expected` first, `actual` second). Swapping the
order produces confusing diff output that misleads reviewers.

#### Scenario: Argument order is reviewable

- **WHEN** `require.Equal(t, actual, expected)` is committed
- **THEN** the code reviewer SHALL flag it and request the swap to
  `(expected, actual)`

### Requirement: Mock-based tests assert expectations at teardown

Every mock-based test SHALL end with `mock.AssertExpectations(t)` (or
`mock.AssertExpectations(GinkgoT())` for Ginkgo suites) so missing or
unexpected interactions fail the suite rather than pass silently.

#### Scenario: Missing expectations fail the test

- **WHEN** a mock's `.On(...)` expectation is configured but never satisfied
- **THEN** `mock.AssertExpectations(t)` SHALL report the unsatisfied
  expectation and fail the test

### Requirement: Tests verify observable behavior

Tests SHALL verify observable behavior (exit codes, stdout/stderr bytes,
file system side effects, return values) rather than internal
implementation details (private field values, internal call counts).
Refactors that preserve observable behavior SHALL NOT break tests.

#### Scenario: Internal refactor does not break tests

- **WHEN** the body of an internal helper is rewritten while preserving
  inputs, outputs, and side effects
- **THEN** the test suite SHALL pass without modification

### Requirement: Cobra commands use RunE only

Cobra commands MUST use `RunE func(cmd *cobra.Command, args []string) error`
and SHALL NOT use `Run func(cmd *cobra.Command, args []string)` because
`Run` cannot return errors. `RunE` SHALL be the entry point for every
subcommand.

#### Scenario: Run is a lint violation

- **WHEN** a cobra command declares `Run: func(...) { ... }`
- **THEN** code review SHALL reject it; `RunE` is the only permitted
  entry-point declaration

### Requirement: Cobra root suppresses duplicate error rendering

The cobra root SHALL set `SilenceUsage: true` and `SilenceErrors: true`
so cobra's auto-rendering does not duplicate the formatter's output.
Errors SHALL be rendered exactly once by `output.FormatError` at the
cmd boundary.

#### Scenario: Duplicate error is reviewable

- **WHEN** the cobra root omits `SilenceUsage` or `SilenceErrors`
- **THEN** the user SHALL see cobra's auto-rendering in addition to the
  formatter's output; code review SHALL reject the omission

### Requirement: Command output routes through cmd.Out/ErrOrStdout

Command handlers SHALL write data via `cmd.OutOrStdout()` and errors via
`cmd.ErrOrStderr()` (or `opts.IO.Out` / `opts.IO.ErrOut`). Direct use
of `os.Stdout` / `os.Stderr` is prohibited so that tests can capture
output via `iostreams.Test()` and the binary's stdout/stderr are not
bypassed.

#### Scenario: Direct os.Stdout is a lint violation

- **WHEN** `cmd.RunE` writes to `os.Stdout` or `os.Stderr` directly
- **THEN** `golangci-lint` SHALL flag it via `gocheckcompilerdirectives`
  or review SHALL require the change to `cmd.OutOrStdout()`

### Requirement: Argument count is validated by cobra Args

Positional argument count SHALL be validated via cobra `Args` validators
(`cobra.ExactArgs(N)`, `cobra.MinimumNArgs(N)`, `cobra.MaximumNArgs(N)`).
`len(os.Args)` and `len(args)` SHALL NOT appear inside `RunE`. Counting
arguments outside `Args` duplicates the cobra contract.

#### Scenario: len(args) is a lint violation

- **WHEN** `RunE` body contains `len(args)` or `len(cmd.Flags().Args())`
- **THEN** code review SHALL reject it and require a cobra `Args`
  validator

### Requirement: nolint directives specify the linter name

`//nolint` directives SHALL specify the linter name (e.g.,
`//nolint:errcheck`, `//nolint:gocritic`). Bare `//nolint` is prohibited
because it suppresses ALL linters at the line and undermines the
lint-driven code-quality contract.

#### Scenario: Bare nolint is reviewable

- **WHEN** a `//nolint` directive without a linter name is committed
- **THEN** code review SHALL reject it; the linter name SHALL follow
  the colon

### Requirement: nolint directives include a justification

`//nolint:<linter>` directives SHALL be followed by a justification
comment explaining why the suppression is necessary. Reviewers SHALL be
able to read the reason without context-switching to a design doc.

#### Scenario: Justification is reviewable

- **WHEN** a `//nolint:errcheck` directive is committed without a
  justification
- **THEN** code review SHALL reject it and request a justification
  comment

### Requirement: Security linter suppressions require strong justification

Security linters (`gosec`, `bodyclose`, `sqlclosecheck`,
`noctxcheck/`) SHALL NOT be suppressed without a strong justification
documented in the directive and a corresponding test that demonstrates
the suppressed condition cannot occur. Bare `//nolint:gosec` is
prohibited regardless of context.

#### Scenario: Security suppression is reviewable

- **WHEN** a security linter directive is silenced
- **THEN** the suppression SHALL include a justification comment AND
  a TestArtifact identifier naming the test that proves the condition is
  unreachable

### Requirement: go.sum is committed

`go.sum` SHALL be committed to the repository. It records cryptographic
checksums of every dependency version and protects against malicious
version drift between developers and CI.

#### Scenario: go.sum drift fails CI

- **WHEN** `git status` shows `go.sum` modifications
- **THEN** the developer SHALL commit them alongside `go.mod`; CI's
  `git diff --exit-code` gate SHALL fail otherwise

### Requirement: go mod tidy gates commits

`go mod tidy` SHALL be run before every commit that changes dependencies.
CI SHALL fail on module drift via `go mod tidy && git diff --exit-code`.
Untidied commits SHALL be rejected by pre-commit hooks via
`mod-tidy-pre-commit`.

#### Scenario: Untidy commit fails CI

- **WHEN** a commit introduces a dependency that `go mod tidy` would
  remove (or vice versa)
- **THEN** the CI module-drift gate SHALL fail the pipeline

### Requirement: Standard library is preferred over new dependencies

Before proposing a new dependency, the agent SHALL evaluate whether the
Go standard library already covers the use case. New third-party
dependencies SHALL be justified in the proposal's Impact section with a
sentence explaining why stdlib is insufficient.

#### Scenario: Stdlib-first is reviewable

- **WHEN** a proposal introduces a new third-party dependency
- **THEN** the proposal SHALL include a one-sentence explanation of why
  stdlib is insufficient for the same use case

### Requirement: AI agents confirm before adding dependencies

AI agents SHALL ask the user for confirmation before running `go get` to
add any new dependency. Adding a dependency is a public-API decision
that ripples across `go.mod`, `go.sum`, vendoring, and CI cache, and
SHALL NOT be made without explicit user approval.

#### Scenario: Unauthorized go get is reviewable

- **WHEN** an AI agent commits a new direct dependency without prior
  user confirmation
- **THEN** the commit SHALL be reverted and the agent SHALL NOT retry
  without re-asking

### Requirement: Binary tools are pinned in mise or go.mod tool directive

Executable tools MUST be pinned in `.mise/config.toml` (mise backends) or
`go.mod`'s `tool` directive (Go 1.24+). `@latest` is prohibited for
CI-affecting tools (`govulncheck`, `cosign`, `syft`, `golangci-lint`,
`gopls`). New tools SHALL be added with explicit version pins.

#### Scenario: @latest pin is a lint violation

- **WHEN** `.mise/config.toml` or `go.mod` references a tool with
  `@latest` or no version qualifier
- **THEN** code review SHALL reject it and require a pinned version