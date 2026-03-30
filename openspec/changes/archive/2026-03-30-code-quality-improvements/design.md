## Context

Twiggit follows DDD-light architecture with clear layer separation (cmd → application → service → infrastructure → domain). The project conventions in AGENTS.md specify:
- ValidationError returned directly without wrapping
- Service errors wrapped with domain.New*ServiceError()
- IsNotFound() method for error categorization

Current issues identified:
- 10+ locations return fmt.Errorf instead of proper domain errors
- ProjectServiceError, NavigationServiceError, ResolutionError lack IsNotFound() methods
- Duplicate code in shell auto-detection, navigation resolution, path validation
- PruneMergedWorktrees modifies shared result without synchronization
- CLI timeout hardcoded instead of using config
- Hook timeout hardcoded instead of configurable

## Goals / Non-Goals

**Goals:**
- Consistent error handling following project conventions
- Nil-safe operations preventing panics
- No duplicate code patterns
- Race-condition-free concurrent operations
- CLI flag consistency and better UX

**Non-Goals:**
- No new user-facing features
- No breaking changes to existing APIs
- No changes to error message content
- No refactoring of service layer beyond identified issues

## Decisions

### Decision 1: Error Return Style

**Choice**: ValidationError will be returned directly from cmd layer without wrapping  
**Rationale**: Per AGENTS.md convention, ValidationError already contains context and should not be wrapped. Wrapping obscures the error type making errors.As() checks fail.  
**Alternatives Considered**: Wrapping all errors with fmt.Errorf - rejected as it violates existing convention and breaks error categorization

### Decision 2: IsNotFound() Implementation

**Choice**: Add IsNotFound() method to ProjectServiceError, NavigationServiceError, ResolutionError  
**Rationale**: Enables consistent error categorization via errors.As() without string matching. Consistent with existing WorktreeServiceError.IsNotFound().  
**Alternatives Considered**: Using error string matching - rejected as fragile and error-prone

### Decision 3: Concurrency Protection

**Choice**: Use mutex to protect result struct modifications in PruneMergedWorktrees  
**Rationale**: go-git repository operations are thread-safe but the result aggregation is not. Simple mutex is appropriate here.  
**Alternatives Considered**: Channel-based result aggregation - rejected as overkill for this use case

### Decision 4: Code Deduplication Approach

**Choice**: Extract to package-level helper functions, not new interfaces  
**Rationale**: Keeps existing interface contracts, minimal refactoring, improves maintainability  
**Alternatives Considered**: New utility package - rejected as over-architecture for simple helpers

### Decision 5: CLI Timeout Configuration

**Choice**: Route CLI timeout through config.Git.CLITimeout with fallback  
**Rationale**: Makes timeout configurable without breaking existing behavior  
**Alternatives Considered**: Only use hardcoded default - rejected as inflexible

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking error handling during refactor | Verify existing error messages preserved via tests |
| Concurrency fix introduces new bugs | Run race detector, add integration test |
| Duplicate code extraction changes behavior | Ensure shared functions have identical behavior |

## Open Questions

None - all technical decisions have clear rationale and approach.
