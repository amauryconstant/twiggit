## 1. Domain Layer - Configuration and Types

- [ ] 1.1 Add exclusion pattern fields to `CompletionConfig` in `internal/domain/config.go`
- [ ] 1.2 Add `IsCurrent`, `IsDirty`, `Remote`, `StyleHint` fields to `ResolutionSuggestion` in `internal/domain/context.go`
- [ ] 1.3 Add default exclusion pattern config (empty slices) in `DefaultConfig()`

## 2. Infrastructure Layer - Context Resolver Enhancements

- [ ] 2.1 Add `addProjectSuggestions()` helper method to `contextResolver` in `internal/infrastructure/context_resolver.go`
- [ ] 2.2 Call `addProjectSuggestions()` from `getProjectContextSuggestions()` (exclude current project)
- [ ] 2.3 Call `addProjectSuggestions()` from `getWorktreeContextSuggestions()` (exclude current project)
- [ ] 2.4 Implement `fuzzyMatch()` function for subsequence matching in `context_resolver.go`
- [ ] 2.5 Update `addWorktreeSuggestions()` to apply fuzzy matching when `FuzzyMatching` config is enabled
- [ ] 2.6 Update `addBranchSuggestions()` to apply fuzzy matching when enabled
- [ ] 2.7 Update `addProjectSuggestions()` to apply fuzzy matching when enabled
- [ ] 2.8 Add exclusion pattern filtering helper `matchesExclusionPatterns(name, patterns []string) bool`
- [ ] 2.9 Apply exclusion patterns to branch suggestions in `addWorktreeSuggestions()` and `addBranchSuggestions()`
- [ ] 2.10 Apply exclusion patterns to project suggestions in `addProjectSuggestions()`
- [ ] 2.11 Enhance worktree descriptions with remote tracking info in `addWorktreeSuggestions()`
- [ ] 2.12 Enhance branch descriptions with remote info in `addBranchSuggestions()`
- [ ] 2.13 Add `IsCurrent` field population in `addWorktreeSuggestions()` when context is worktree
- [ ] 2.14 Implement smart sorting: current worktree first, default branch second, alphabetical rest

## 3. Infrastructure Layer - Status Indicators

- [ ] 3.1 Add dirty status check for current worktree in `getWorktreeContextSuggestions()` using `GetRepositoryStatus()`
- [ ] 3.2 Set `IsDirty` field on current worktree suggestion only (not all worktrees)
- [ ] 3.3 Ensure status check respects 500ms timeout with graceful degradation

## 4. Cmd Layer - Completion Actions

- [ ] 4.1 Refactor `suggestionsToCarapaceAction()` to use `carapace.Batch()` for separate project/branch actions
- [ ] 4.2 Apply `.Suffix("/")` to project suggestions only in the batch
- [ ] 4.3 Apply branch suggestions without suffix in the batch
- [ ] 4.4 Apply smart sorting in `suggestionsToCarapaceAction()` using suggestion metadata
- [ ] 4.5 Apply visual style hints (dirty indicator) to current worktree suggestion
- [ ] 4.6 Fix `actionBranchesForProject()` to create synthetic context for target project
- [ ] 4.7 Remove broken `ProjectName` filter in `actionBranchesForProject()` (suggestions already correct from synthetic context)
- [ ] 4.8 Apply exclusion pattern filtering in `suggestionsToCarapaceAction()` using config

## 5. Configuration Loading

- [ ] 5.1 Update config loader to parse `completion.exclude_branches` from TOML
- [ ] 5.2 Update config loader to parse `completion.exclude_projects` from TOML
- [ ] 5.3 Wire exclusion patterns from config through to completion functions

## 6. Testing

- [ ] 6.1 Add unit tests for `fuzzyMatch()` function in `internal/infrastructure/context_resolver_test.go`
- [ ] 6.2 Add unit tests for `matchesExclusionPatterns()` helper
- [ ] 6.3 Add unit tests for project suggestions from project context in `context_resolver_test.go`
- [ ] 6.4 Add unit tests for project suggestions from worktree context in `context_resolver_test.go`
- [ ] 6.5 Add unit tests for exclusion pattern filtering in `context_resolver_test.go`
- [ ] 6.6 Add unit tests for smart sorting logic in `context_resolver_test.go`
- [ ] 6.7 Add integration test for cross-project completion with synthetic context
- [ ] 6.8 Add E2E test for progressive project completion (auto-slash suffix)
- [ ] 6.9 Add E2E test for fuzzy matching completion
- [ ] 6.10 Verify existing completion tests still pass (regression check)

## 7. Documentation and Finalization

- [ ] 7.1 Update `openspec/specs/shell-completion/spec.md` with progressive completion requirements
- [ ] 7.2 Update `openspec/specs/path-resolution/spec.md` with project suggestions from all contexts
- [ ] 7.3 Run `mise run check` to verify all linting and tests pass
- [ ] 7.4 Manual testing of completion in bash/zsh shells
