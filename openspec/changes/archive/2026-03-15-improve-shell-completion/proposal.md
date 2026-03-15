## Why

Shell completion currently has three bugs that break expected workflows: (1) projects aren't suggested from project/worktree context, (2) cross-project branch completion (`project/<tab>`) returns empty results, and (3) selecting a project doesn't auto-append "/" to trigger branch completion. Additionally, the completion system lacks quality-of-life features users expect: fuzzy matching, smart sorting, enriched descriptions, status indicators, and exclusion patterns for noisy branches.

## What Changes

**Bug Fixes:**
- Add project name suggestions from project/worktree context (spec-compliant)
- Fix cross-project branch completion by creating synthetic context for target project
- Add "/" suffix to project suggestions for progressive completion

**Enhancements:**
- Implement fuzzy matching for completion filtering (config flag exists but unimplemented)
- Smart sorting: current worktree first, then default branch, then alphabetical
- Enhanced descriptions with remote tracking info and relative dates
- Status indicators showing dirty state for current worktree only
- Exclusion patterns to filter out noisy branches/projects via config

## Capabilities

### New Capabilities

- `completion-filtering`: Pattern-based exclusion of branches and projects from suggestions, plus fuzzy matching for partial input
- `completion-enrichment`: Enhanced descriptions, status indicators, and smart sorting for completion suggestions

### Modified Capabilities

- `shell-completion`: Adding requirements for progressive project completion (auto-slash suffix)
- `path-resolution`: Clarifying that project suggestions SHALL be provided from all contexts, not just outside-git

## Impact

**Files Modified:**
- `internal/infrastructure/context_resolver.go` - Add project suggestions, enhanced descriptions, smart sorting
- `cmd/suggestions.go` - Fix cross-project completion, add "/" suffix for projects, apply styles
- `internal/domain/config.go` - Add exclusion pattern configuration
- `internal/domain/context.go` - Add fields to `ResolutionSuggestion` for status/style hints

**No Breaking Changes** - All changes are additive or bug fixes.

**Performance Considerations:**
- Status indicators limited to current worktree only (1 git status call)
- Existing 5-second cache and 500ms timeout remain in place
- Fuzzy matching is pure string operations (no git calls)
