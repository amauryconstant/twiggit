## Context

The completion system uses Carapace's `ActionMultiParts("/")` for progressive completion of `project/branch` syntax. The current implementation has three bugs and lacks quality-of-life features.

**Current Architecture:**
```
cmd/suggestions.go                    infrastructure/context_resolver.go
┌─────────────────────────┐          ┌──────────────────────────────────┐
│ actionWorktreeTarget()  │          │ GetResolutionSuggestions(ctx,    │
│   └─ ActionMultiParts   │─────────▶│   partial, opts...)              │
│        ├─ case 0:       │          │                                  │
│        │   projectsOr   │          │ └─ getProjectContextSuggestions  │
│        │   Branches()   │          │ └─ getWorktreeContextSuggestions │
│        └─ case 1:       │          │ └─ getOutsideGitContextSuggestions│
│            branchesFor  │          │                                  │
│            Project()    │          └──────────────────────────────────┘
└─────────────────────────┘
```

**Constraints:**
- 500ms timeout for git operations (graceful degradation)
- 5-second cache on completion actions
- No new interface methods if avoidable
- Tests written AFTER implementation

## Goals / Non-Goals

**Goals:**
- Fix 3 bugs: missing project suggestions, broken cross-project completion, missing "/" suffix
- Implement fuzzy matching, smart sorting, enhanced descriptions, status indicators, exclusion patterns
- Maintain existing performance characteristics (timeout, cache)
- Zero breaking changes

**Non-Goals:**
- Status indicators for ALL worktrees (too expensive - requires N git status calls)
- Recent items tracking (requires persistence layer)
- Activity-based sorting (requires extra git operations)

## Decisions

### Decision 1: Project Suggestions from Project/Worktree Context

**Choice:** Add `addProjectSuggestions()` helper to `contextResolver` that calls existing `discoverProjects()` method.

**Rationale:** 
- `getOutsideGitContextSuggestions` already has this logic - reuse pattern
- `discoverProjects()` already exists and handles filesystem scanning
- Filtering by partial match and excluding current project is trivial

**Alternatives Considered:**
- Call `ProjectService.ListProjectSummaries()` from cmd layer - adds cross-layer dependency
- Cache project list separately - adds complexity, existing 5s cache is sufficient

### Decision 2: Fix Cross-Project Branch Completion

**Choice:** Create synthetic `Context` with target project path, pass to `GetCompletionSuggestionsFromContext()`.

```go
targetCtx := &domain.Context{
    Type:        domain.ContextProject,
    ProjectName: projectName,
    Path:        filepath.Join(config.ProjectsDirectory, projectName),
}
suggestions := GetCompletionSuggestionsFromContext(targetCtx, "")
```

**Rationale:**
- No new interface methods needed
- Reuses existing suggestion logic
- `GetResolutionSuggestions` already handles `ContextProject` correctly

**Alternatives Considered:**
- Add `GetSuggestionsForProject(projectName)` to interfaces - more surface area
- Fetch branches directly in cmd layer - bypasses service layer

### Decision 3: Auto-Slash for Project Suggestions

**Choice:** Use `carapace.Batch()` to combine project suggestions (with `.Suffix("/")`) and branch suggestions (no suffix).

```go
func suggestionsToCarapaceAction(suggestions []*domain.ResolutionSuggestion) carapace.Action {
    projects := filterByType(suggestions, domain.PathTypeProject)
    branches := filterByType(suggestions, domain.PathTypeWorktree)
    
    return carapace.Batch(
        carapace.ActionValues(projectValues...).Suffix("/"),
        carapace.ActionValues(branchValues...),
    ).ToA()
}
```

**Rationale:**
- Carapace doesn't support per-value suffixes in single action
- Batch + ToA merges actions cleanly
- Shell receives "project/" and triggers ActionMultiParts case 1

**Alternatives Considered:**
- Add suffix to all suggestions - breaks branch completion
- Custom RawValue construction - more complex, Batch is cleaner

### Decision 4: Fuzzy Matching Implementation

**Choice:** Implement simple case-insensitive subsequence matching.

```go
func fuzzyMatch(pattern, text string) bool {
    pattern = strings.ToLower(pattern)
    text = strings.ToLower(text)
    pi := 0
    for _, c := range text {
        if pi < len(pattern) && byte(c) == pattern[pi] {
            pi++
        }
    }
    return pi == len(pattern)
}
```

**Rationale:**
- Config flag `NavigationConfig.FuzzyMatching` already exists
- Pure string operations - no git calls, no performance impact
- Sufficient for branch/project names (short strings)

**Alternatives Considered:**
- Third-party fuzzy library - overkill for this use case
- Levenshtein distance - more expensive, less intuitive for completion

### Decision 5: Smart Sorting

**Choice:** Sort suggestions in `suggestionsToCarapaceAction()` before creating Carapace action.

Priority order:
1. Current worktree (if context is worktree and matches)
2. "main" / default branch
3. Other branches alphabetically

**Rationale:**
- Sorting is cheap (already have all suggestions in memory)
- Current worktree is most likely target when in worktree context
- Default branch is common navigation target

**Alternatives Considered:**
- Sort in resolver layer - mixes data retrieval with presentation
- Track recent selections - requires persistence, out of scope

### Decision 6: Enhanced Descriptions

**Choice:** Build descriptions in `context_resolver.go` using available `BranchInfo` and `WorktreeInfo` fields.

Format:
- Worktree: `"Worktree • origin/branch • 2 days ago"` or `"Worktree • 2 commits ahead"`
- Branch: `"Branch • origin/branch • create worktree"`

**Rationale:**
- Data already available from `ListBranches()` and `ListWorktrees()`
- No additional git calls
- Description construction is presentation logic but lives in resolver (acceptable for completion)

**Alternatives Considered:**
- Return raw data, format in cmd layer - more complex data flow
- Skip enhanced descriptions - reduces UX value

### Decision 7: Status Indicators (Current Worktree Only)

**Choice:** Check dirty status for current worktree only, use Carapace styles for visual indication.

```go
if worktree.Branch == currentBranch {
    if !repoStatus.IsClean {
        description = "⚠ " + description  // or use style.Yellow
    }
}
```

**Rationale:**
- Single `GetRepositoryStatus()` call is acceptable
- N calls for all worktrees would be too slow
- Current worktree is most relevant for navigation decisions

**Alternatives Considered:**
- Check all worktrees - N git status calls, too slow
- Skip status entirely - loses valuable information

### Decision 8: Exclusion Patterns

**Choice:** Add glob-based exclusion patterns to config, filter in `suggestionsToCarapaceAction()`.

```toml
[completion]
exclude_branches = ["dependabot/*", "renovate/*"]
exclude_projects = ["archive/*"]
```

**Rationale:**
- Pure filtering after fetch - no impact on git operations
- Glob patterns are familiar and flexible
- User-configurable for different project styles

**Alternatives Considered:**
- Regex patterns - more powerful but less familiar
- Age-based filtering - requires parsing dates, more complex

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Status indicator adds 1 git call per completion | Already covered by 500ms timeout; graceful degradation returns empty |
| Fuzzy matching could match too many items | Combine with prefix match first, then fuzzy; respect MaxSuggestions config |
| Exclusion patterns could hide wanted branches | Use glob patterns (not regex), document in config, user controls |
| Enhanced descriptions could be verbose | Keep format concise, truncate dates to relative ("2 days ago") |
| Synthetic context for cross-project could fail silently | If project doesn't exist, git operations fail → empty suggestions (graceful) |

## Migration Plan

No migration required - all changes are additive:
1. New config fields are optional (defaults to empty = no filtering)
2. Existing completion behavior preserved for users without new config
3. Bug fixes apply immediately without configuration

## Open Questions

None - design decisions are complete based on exploration phase.
