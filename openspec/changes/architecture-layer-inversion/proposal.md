# Proposal: Architecture Layer Inversion

## Why

`internal/service/` imports `internal/infrastructure/` at 8 call sites
(verified by grep) to reach pure helpers and one delegating facade
(`CompositeGitClient`). The dependency rule
(`service → application → domain`) is documented in
`internal/application/AGENTS.md` but enforced by no tool; layer drift
is observable in code and in the four AGENTS.md files that document
the wrong shape. The Go-CLI architecture, naming, lint, safety, and
design-patterns skills flag ~30 review findings that live in the same
files the inversion touches, so the inversion and the cleanup land
together.

## What Changes

### Layer inversion core

- Move `internal/infrastructure/pathutils.go` to
  `internal/domain/pathutils.go`. Functions return plain `error` (no
  typed chain), removing the `ContextDetectionError` mismatch.
- Move `FindMainRepoByTraversal`, `IsMainRepo`, `FindGitDirByTraversal`
  (with signature change `*string` → `(string, bool)`) to
  `internal/domain/git_repo.go`.
- Move `FindGitRepositories` to
  `internal/infrastructure/repo_finder.go`; it now returns
  `[]domain.GitDir` (the type moves to `internal/domain/git_repo.go`).
- Add `application.RepoLocator` interface exposing
  `FindGitRepositories(dir) ([]domain.GitDir, error)`.
- Move the shell wrapper template constant out of
  `internal/infrastructure/shell_infra.go` into
  `internal/domain/shell_wrapper.go` so tests assert on a domain
  function instead of constructing real infrastructure.
- **BREAKING** Delete `CompositeGitClient` (`internal/infrastructure/git_client.go`)
  and the `application.GitClient` umbrella interface
  (`internal/application/interfaces.go:102-105`).
- Service constructors take `application.GoGitClient` and
  `application.CLIClient` as two separate fields
  (WorktreeService, ProjectService, NavigationService). `ContextResolver`
  constructor also splits.
- `main.go` drops `infrastructure.NewCompositeGitClient` and wires both
  clients directly into services.
- Delete `internal/infrastructure/interfaces.go` (placeholder).

### Service layer cleanup

- `internal/service/worktree_service.go` drops the `infrastructure`
  import; deletes the dead `sync.Mutex` field; switches
  `hookResult, _ = s.hookRunner.Run(...)` to log via `slog` and
  continue (no behavior change); switches the swallowed prune error
  at line 611 to `slog`; switches the linear
  `protectedBranches` scan to `slices.Contains`.
- `internal/service/project_service.go` drops the `infrastructure`
  import; replaces `os.IsNotExist(err)` with `errors.Is(err,
  os.ErrNotExist)` at line 221; verifies and removes the unused
  `contextService` field.
- `internal/service/context_service.go` removes the unused `config`
  field.
- `internal/service/shell_service.go` and its test stop using
  `infrastructure.NewShellInfrastructure()` for fixture capture;
  they call `domain.ShellWrapper(shellType)` instead.
- `internal/service/doc.go` rewritten.

### Infrastructure rewrite (in touched files)

- **BREAKING** `NewGoGitClient` and `NewGoGitClientWithSize` now
  return `(*GoGitClient, error)` because `lru.New` failure is no
  longer discarded (silently nil cache, panic risk on next call).
- `internal/infrastructure/cli_client.go` adds nil-guard before
  `result.ExitCode` at lines 115, 151, 176, 198, 220, 247 and the
  hook-runner nil-guard at `hook_runner.go:137`. Renames
  `CLIClientImpl` → `CLIClient` (drops `Impl` suffix per skill).
- `internal/infrastructure/gogit_client.go` deletes the
  `_ = remoteRef` workaround at line 128; bounds the hash slice at
  line 336 via `min(7, len(s))`; renames `GoGitClientImpl` →
  `GoGitClient`.
- `internal/infrastructure/command_executor.go` renames
  `DefaultCommandExecutor` → `CommandExecutor` and
  `NewDefaultCommandExecutor` → `NewCommandExecutor`; extracts
  the lowercase + substring heuristic into one helper.
- `internal/infrastructure/context_detector.go` swaps the unbounded
  `map[string]cached` for the bounded LRU already used by the git
  client; renames the implementation type to drop `Impl`.
- `internal/infrastructure/context_resolver.go` replaces the local
  `ProjectRef` with `domain.ProjectSummary`; renames the
  implementation type.
- `internal/infrastructure/config_manager.go` switches the
  "deep copy" of `ProtectedBranches` to `slices.Clone`; renames
  the implementation type.
- `internal/infrastructure/hook_runner.go` collapses the five
  near-identical no-op result blocks into `noOpResult(req)`; the
  six env-export blocks iterate over a struct slice; renames the
  implementation type.
- `internal/infrastructure/shell_infra.go` renames the
  implementation type and removes the wrapper template constant
  (moved to domain).
- `os.IsNotExist(err)` → `errors.Is(err, os.ErrNotExist)` across
  all touched infrastructure files.
- `strings.HasPrefix + TrimPrefix` → `strings.CutPrefix` at
  `cli_client.go:25-26` and `config_manager.go:29`.
- `for i := 0; i < N; i++` → `for i := range N` in touched files.

### Domain renames

- **BREAKING** `domain.PathType` adds `PathTypeUnknown = iota 0`;
  existing values shift by +1.
- `domain.WorktreeInfo.Modified bool` → `IsModified bool`.
- `domain.ShellResult.Installed bool` → `IsInstalled bool`;
  `Skipped bool` → `IsSkipped bool`.
- `domain.HookResult.Executed bool` → `HasExecuted bool`;
  `Success bool` → `IsSuccessful bool`.
- `domain.Result[T]` constructors: `NewErrorResult[T]` →
  `NewErrResult[T]`.
- **BREAKING** `domain.ValidationError.Context()` → `Detail()`
  (the name collided with `domain.Context` at call sites).

### Test mocks

- `test/mocks/git_service_mock.go` becomes `MockGitClientBundle`
  exposing `MockGoGitClient` and `MockCLIClient` as two named
  fields. Existing tests pass both mocks to the service
  constructor.
- `test/mocks/shell_infrastructure_mock.go` and any other mock
  naming `*Impl` is renamed to drop the suffix.
- `test/integration/`, `test/concurrent/`, `test/e2e/fixtures/`
  update for renamed constructors (mechanical via `gopls rename`).

### Lint config

- Drop `gocognit` from `.golangci.yml` (redundant with `gocyclo` +
  `nestif`).
- Add `nolintlint` with `require-explanation: true,
  require-specific: true`.
- Add `errcheck.check-type-assertions: true`.
- Extend `depguard` rules:
  - `domain` — allow `$gostd`, `internal/domain` (unchanged)
  - `service` — allow `$gostd`, `internal/domain`,
    `internal/application`, `internal/service`
  - `application` — allow `$gostd`, `internal/domain`,
    `internal/application`
  - `infrastructure` — allow `$gostd`, `internal/domain`,
    `internal/application`, `internal/infrastructure`
  - `cmd` — allow `$gostd`, `internal/domain`,
    `internal/application`, `internal/service`,
    `internal/infrastructure`, `internal/version`, `twiggit/cmd`
- Drop the blanket `text: "Close.*is not checked"` exclusion; add
  per-line `//nolint:errcheck // reason` at the ~5–10 `Close()`
  call sites that intentionally discard the error.

### AGENTS.md sync

Rewrite to current shape:
- `internal/application/AGENTS.md` — interface table reflects
  the two-role constructor and the new `RepoLocator`.
- `internal/service/AGENTS.md` — examples use `domain.*` and
  `application.*` only.
- `internal/infrastructure/AGENTS.md` — interface renames +
  `*Impl` drops documented.
- `internal/domain/AGENTS.md` — new helpers, renames, PathTypeUnknown.
- `test/mocks/AGENTS.md` — bundle struct documented.

## Capabilities

### New Capabilities

None. `RepoLocator` and the new `domain.ShellWrapper` fit under
existing `application-service-interfaces` and a future
`domain-shell-wrappers` spec (deferred — no behavior change
because the wrapper content is byte-for-byte unchanged).

### Modified Capabilities

- `application-service-interfaces` — drop `GitClient` umbrella;
  document role interfaces (`GoGitClient` + `CLIClient`) and the
  new `RepoLocator`; add DI rule that reverse imports SHALL NOT
  compile (existing rule, made mechanical).
- `application-worktree-management` — drop the "composite"
  wording from requirement 7; drop requirement 2 (per-project
  mutex) since the single `sync.Mutex` field on
  `worktreeService` is dead code and the spec requirement is
  unfulfilled.
- `application-project-service` — no requirement change. The
  `RepoLocator` injection is internal to `ProjectService`
  implementation; the public contract stays the same.
- `infrastructure-git-client` — drop the "composite GitClient is
  the only injection point" sentence in requirement 1; document
  the role-interface split.
- `infrastructure-path-utils` — document the move of
  `ExtractProjectFromWorktreePath`, `NormalizePath`,
  `IsPathUnder` from `internal/infrastructure/pathutils.go` to
  `internal/domain/pathutils.go` (the public contracts are
  unchanged).
- `domain-context-types` — add `PathTypeUnknown = iota 0`;
  existing values shift by +1.
- `domain-hook-types` — `HookResult.Executed` → `HasExecuted`;
  `HookResult.Success` → `IsSuccessful`.
- `domain-typed-errors` — `ValidationError.Context()` →
  `Detail()`; `NewErrorResult[T]` → `NewErrResult[T]`.

## Impact

| Layer | Files | Lines |
|---|---|---|
| `internal/domain/` | `pathutils.go` (new), `git_repo.go` (new), `shell_wrapper.go` (new), `git_types.go`, `shell_results.go`, `hook_types.go`, `service_results.go`, `service_errors.go`, `context.go` | ~600 |
| `internal/application/` | `interfaces.go` | ~80 |
| `internal/infrastructure/` | `pathutils.go` (delete), `git_utils.go` (delete), `git_client.go` (delete), `interfaces.go` (delete), `repo_finder.go` (new), `cli_client.go`, `gogit_client.go`, `command_executor.go`, `context_detector.go`, `context_resolver.go`, `config_manager.go`, `hook_runner.go`, `shell_infra.go` | ~1100 |
| `internal/service/` | `worktree_service.go`, `project_service.go`, `context_service.go`, `shell_service.go`, `shell_service_test.go`, `doc.go` | ~250 |
| `cmd/`, `main.go` | `main.go`, `cmd/root.go` (interface ref) | ~60 |
| `test/` | `test/mocks/*`, `test/integration/*`, `test/concurrent/*`, `test/e2e/fixtures/*` (mechanical renames) | ~400 |
| `.golangci.yml` | depguard rules + gocognit drop + nolintlint | ~50 |
| AGENTS.md | 5 files rewritten | ~500 |
| Specs | 7 spec deltas via `openspec/specs/<id>/spec.md` | — |
| Tests (after impl, per project rule) | `internal/domain/*_test.go` (new), `internal/service/*_test.go` (rewritten), `internal/infrastructure/*_test.go` (rewritten) | ~1000 |

| Aspect | Effect |
|---|---|
| End-user CLI | Unchanged |
| Public API of `internal/application` | `GitClient` umbrella removed; `RepoLocator` added; `NewGoGitClient*` returns `error` |
| Public API of `internal/domain` | `PathType` enum values shift +1; `ValidationError.Context` → `Detail`; `Result[T]` constructor renamed; three boolean fields renamed |
| Public API of `internal/infrastructure` | `*Impl` types renamed; `CompositeGitClient` deleted |
| Tests | Mechanical rename across `test/integration`, `test/concurrent`, `test/e2e/fixtures`; new domain test files |
| Lint | depguard blocks reverse imports; existing service → infrastructure imports fail until slice 3 lands |
| Version | Hard break (release bump handled outside this change) |

## Non-goals

- Wholesale layer collapse (Tier 2 CLI shape) — separate future
  change.
- Interface Segregation Principle refactor (splitting
  `GoGitClient` / `CLIClient` into 1-method roles) — separate
  future change; this change keeps the 8-method + 6-method facet
  shape and only enforces the layer boundary.
- Receiver-less methods → free functions in `service/` —
  separate refactor.
- All modernization in `cmd/` (~30 sites for `_, _ = fmt.Fprint*`,
  `cmd.Context()`, `signal.NotifyContext`) — deferred.
- All godoc comment fixes in untouched domain files — deferred
  to the quality change.
- Lint threshold tightening (`funlen`, `gocyclo`) that would
  trigger pre-existing code — deferred.
- CI / Docker / GoReleaser / README / CHANGELOG / `go.mod`
  improvements — deferred.
- Hook runner refactor (duplicated result blocks → single helper)
  beyond the no-op result collapse — deferred.
- Context resolver ctx propagation (currently uses
  `context.Background()`) — deferred; flag in design.md as
  known carry-over.
- `SuggestionOption func(interface{})` typed-option refactor —
  deferred.
