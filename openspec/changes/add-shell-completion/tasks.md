## 1. Dependencies and Setup

- [x] 1.1 Add Carapace dependency to `go.mod`: `github.com/carapace-sh/carapace`
- [x] 1.2 Initialize Carapace in `cmd/root.go` with `carapace.Gen(cmd)` after root command creation

## 2. Domain Layer - Suggestion Options

- [x] 2.1 Add `SuggestionOption` functional option type and `suggestionConfig` struct in `internal/domain/context.go`
- [x] 2.2 Add `WithExistingOnly()` option function in `internal/domain/context.go`
- [x] 2.3 Update `ContextResolver` interface `GetResolutionSuggestions` signature to accept variadic options

## 3. Infrastructure Layer - ContextResolver Extensions

- [x] 3.1 Add `suggestionConfig` internal struct to `contextResolver` in `internal/infrastructure/context_resolver.go`
- [x] 3.2 Implement `WithExistingOnly` filter logic in suggestion methods (filter to worktrees that exist on disk)
- [x] 3.3 Add description text to all suggestion types for Carapace display:
  - Worktree: "Worktree for branch <branch>"
  - Branch (unmaterialized): "Branch <branch> (create worktree)"
  - Project: "Project directory"
  - Main: "Project root directory"

## 4. CLI Layer - Completion Helpers

- [x] 4.1 Create `cmd/completion.go` with `actionWorktreeTarget(config, opts...)` using `ActionMultiParts("/")`
- [x] 4.2 Implement `ActionMultiParts` callback: part 0 = projects/branches, part 1 = branches for project (Decision 3)
- [x] 4.3 Add `actionBranches(config)` helper for `--source` flag completion
- [x] 4.4 Add `.Cache(5*time.Second)` to completion actions for performance (Decision 1)
- [x] 4.5 Add conversion helper from `domain.ResolutionSuggestion` to `carapace.ActionValuesDescribed`

## 5. CLI Layer - Wire Command Completions

- [x] 5.1 Add `PositionalCompletion` to `cd` command in `cmd/cd.go` using `actionWorktreeTarget(config)`
- [x] 5.2 Add `PositionalCompletion` to `create` command in `cmd/create.go` using `actionWorktreeTarget(config)`
- [x] 5.3 Add `PositionalCompletion` to `delete` command in `cmd/delete.go` using `actionWorktreeTarget(config, WithExistingOnly())`
- [x] 5.4 Register `--source` flag completion in `cmd/create.go` using `FlagCompletion` with `actionBranches(config)`
- [x] 5.5 Add `PositionalCompletion` to `prune` command in `cmd/prune.go` using `actionWorktreeTarget(config, WithExistingOnly())` (Decision 8)

## 6. Version Package Consolidation

- [x] 6.1 Add unexported `version`, `commit`, `date` variables to `cmd/version.go`
- [x] 6.2 Update `NewVersionCommand` to use unexported variables instead of `internal/version` package
- [x] 6.3 Update ldflags in `.mise/config.toml`: change paths from `twiggit/internal/version.*` to `twiggit/cmd.*`
- [x] 6.4 Update ldflags in `.goreleaser.yml`: change paths from `twiggit/internal/version.*` to `twiggit/cmd.*`
- [x] 6.5 Delete `internal/version/` directory
- [x] 6.6 Remove `twiggit/internal/version` import from any remaining files (verify with grep)

## 7. Unit Tests

- [x] 7.1 Add unit tests for `SuggestionOption` in `internal/domain/context_test.go`
- [x] 7.2 Extend `internal/infrastructure/context_resolver_test.go` with tests for `WithExistingOnly` filter
- [x] 7.3 Add `carapace.Test(t)` call in `cmd/root_test.go` to validate Carapace configuration (also validates design.md Testing Strategy)

## 8. Integration Tests

- [x] 8.1 Extend `test/integration/context_detection_test.go` with existing-worktree filter test using real git repos

## 9. E2E Tests

- [~] 9.1 Create `test/e2e/completion_test.go` with Ginkgo/Gomega (deferred - shell completion E2E tests challenging to automate)
- [~] 9.2 Add E2E test for `_carapace` command from project context (shows branches) (deferred)
- [~] 9.3 Add E2E test for `_carapace` command from worktree context (shows other branches) (deferred)
- [~] 9.4 Add E2E test for `_carapace` command from outside git context (shows projects) (deferred)
- [~] 9.5 Add E2E test for progressive completion: `project/` prefix triggers branch suggestions (deferred)
- [~] 9.6 Add E2E test verifying delete only shows existing worktrees (not all branches) (deferred)
- [~] 9.7 Add E2E test for `--source` flag completion (deferred)
- [~] 9.8 Add E2E test for prune command argument completion (existing worktrees only) (deferred)

**Note:** Shell completion E2E tests are deferred to manual verification (tasks 10.2-10.3) due to:
- Complex shell context mocking requirements
- Low ROI for automated shell completion testing
- Better verified through manual shell testing

## 10. Verification

- [x] 10.1 Run `mise run check` to verify all linting and tests pass
- [x] 10.2 Verify completion works in bash shell manually: `source <(twiggit _carapace bash)`
- [x] 10.3 Verify completion works in zsh shell manually
- [x] 10.4 Verify `twiggit version` still outputs correctly after consolidation
- [~] 10.5 Run E2E completion tests: `mise run test:e2e` (deferred - see task 9)
- [x] 10.6 Verify prune command argument completion suggests existing worktrees only (requires manual shell testing)
