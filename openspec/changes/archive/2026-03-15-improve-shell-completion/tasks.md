## 1. Domain Layer - Configuration and Types

- [x] 1.1 Add exclusion pattern fields to `CompletionConfig` in `internal/domain/config.go`
- [x] 1.2 Add `IsCurrent`, `IsDirty`, `Remote`, `StyleHint` fields to `ResolutionSuggestion` in `internal/domain/context.go`
- [x] 1.3 Add default exclusion pattern config (empty slices) in `DefaultConfig()`

## 2. Infrastructure Layer - Context Resolver Enhancements

- [x] 2.1 Add `addProjectSuggestions()` helper method to `contextResolver` in `internal/infrastructure/context_resolver.go`
- [x] 2.2 Call `addProjectSuggestions()` from `getProjectContextSuggestions()` (exclude current project)
- [x] 2.3 Call `addProjectSuggestions()` from `getWorktreeContextSuggestions()` (exclude current project)
- [x] 2.4 Implement `fuzzyMatch()` function for subsequence matching in `context_resolver.go`
- [x] 2.5 Update `addWorktreeSuggestions()` to apply fuzzy matching when `FuzzyMatching` config is enabled
- [x] 2.6 Update `addBranchSuggestions()` to apply fuzzy matching when enabled
- [x] 2.7 Update `addProjectSuggestions()` to apply fuzzy matching when enabled
- [x] 2.8 Add exclusion pattern filtering helper `matchesExclusionPatterns(name, patterns []string) bool`
- [x] 2.9 Apply exclusion patterns to branch suggestions in `addWorktreeSuggestions()` and `addBranchSuggestions()`
- [x] 2.10 Apply exclusion patterns to project suggestions in `addProjectSuggestions()`
- [x] 2.11 Enhance worktree descriptions with remote tracking info in `addWorktreeSuggestions()`
- [x] 2.12 Enhance branch descriptions with remote info in `addBranchSuggestions()`
- [x] 2.13 Add `IsCurrent` field population in `addWorktreeSuggestions()` when context is worktree
- [x] 2.14 Implement smart sorting: current worktree first, default branch second, alphabetical rest (moved to cmd layer)

## 3. Infrastructure Layer - Status Indicators

- [x] 3.1 Add dirty status check for current worktree in `addWorktreeSuggestions()` using `GetRepositoryStatus()`
- [x] 3.2 Set `IsDirty` field on current worktree suggestion only (not all worktrees)
- [x] 3.3 Ensure status check respects 500ms timeout with graceful degradation (via context and error handling)

## 4. Cmd Layer - Completion Actions

- [x] 4.1 Refactor `suggestionsToCarapaceAction()` to use `carapace.Batch()` for separate project/branch actions
- [x] 4.2 Apply `.Suffix("/")` to project suggestions only in the batch
- [x] 4.3 Apply branch suggestions without suffix in the batch
- [x] 4.4 Apply smart sorting in `suggestionsToCarapaceAction()` using suggestion metadata
- [x] 4.5 Apply visual style hints (dirty indicator) to current worktree suggestion (via description prefix)
- [x] 4.6 Fix `actionBranchesForProject()` to create synthetic context for target project
- [x] 4.7 Remove broken `ProjectName` filter in `actionBranchesForProject()` (suggestions already correct from synthetic context)
- [x] 4.8 Apply exclusion pattern filtering in `suggestionsToCarapaceAction()` using config (done in resolver layer)

## 5. Configuration Loading

- [x] 5.1 Update config loader to parse `completion.exclude_branches` from TOML (field already exists with toml tag)
- [x] 5.2 Update config loader to parse `completion.exclude_projects` from TOML (field already exists with toml tag)
- [x] 5.3 Wire exclusion patterns from config through to completion functions (done via cr.config.Completion.Exclude*)

## 6. Testing

- [x] 6.1 Add unit tests for `fuzzyMatch()` function in `internal/infrastructure/context_resolver_test.go`
- [x] 6.2 Add unit tests for `matchesExclusionPatterns()` helper
- [x] 6.3 Add unit tests for project suggestions from project context in `context_resolver_test.go`
- [x] 6.4 Add unit tests for project suggestions from worktree context in `context_resolver_test.go`
- [x] 6.5 Add unit tests for exclusion pattern filtering in `context_resolver_test.go`
- [x] 6.6 Add unit tests for smart sorting logic in `cmd/suggestions_test.go`
- [x] 6.7 Add integration test for cross-project completion with synthetic context
- [x] 6.8 Add E2E test for progressive project completion (auto-slash suffix)
- [x] 6.9 Add E2E test for fuzzy matching completion
- [x] 6.10 Verify existing completion tests still pass (regression check)

## 7. Documentation and Finalization

- [x] 7.1 Update `openspec/specs/shell-completion/spec.md` with progressive completion requirements
- [x] 7.2 Update `openspec/specs/path-resolution/spec.md` with project suggestions from all contexts
- [x] 7.3 Run `mise run check` to verify all linting and tests pass
- [ ] 7.4 Manual testing of completion in bash/zsh shells (deferred to user testing)
