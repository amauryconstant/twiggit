---
name: audit-codebase-quality
description: Comprehensive code quality audit for consistency, coherency, and de-duplication. Use when reviewing any codebase for architectural issues, duplicate code patterns, test gaps, documentation accuracy, and quality standards. Supports language-agnostic auditing with auto-discovery of source directories.
---

# Codebase Quality Audit

## Overview

Comprehensive code quality audit tool that identifies inconsistencies, duplications, and architectural issues across any codebase. Uses parallel processing to audit multiple quality areas simultaneously, then deduplicates and consolidates findings into a prioritized action plan.

Supports **any codebase** (language-agnostic) with intelligent auto-discovery of source directories, extensible audit areas, and language-specific heuristics.

## When to Use This Skill

Use when you need to review codebase quality across:

- **Package structure**: Naming conventions, directory organization, layer separation, architectural violations
- **Duplicate code**: Repeated logic, similar functions, consolidation opportunities, copied code
- **Interface compliance**: Implementation completeness, signature mismatches, unused interfaces, method gaps
- **Test patterns**: Test organization, coverage gaps, mock usage consistency, test naming
- **Documentation accuracy**: AGENTS.md vs actual code, missing types/methods, incorrect signatures
- **Import patterns**: Import ordering, circular dependencies, unused imports, external dependencies
- **Error handling**: Wrapping patterns, error type usage, message consistency, chain breaks
- **Mock centralization**: Inline vs centralized mocks, duplicate mock implementations, missing mocks
- **Language-specific patterns**: See [go-patterns/](references/go-patterns/) directory for Go-specific heuristics

## Audit Process

This skill uses **parallel processing workflow** - multiple independent audit areas execute simultaneously, then findings are deduplicated and consolidated using fingerprint-based matching.

### Discovery Phase

Auto-discovers source directories using these patterns:
- `internal/`, `cmd/`, `pkg/`, `src/`, `lib/`
- `api/`, `app/`, `server/`, `client/`
- Language-specific: `*.go`, `*.py`, `*.ts/tsx`, `*.java`, `*.rs`, `*.cs`

Automatically skips standard build artifacts and dependencies:
- `.git/`, `vendor/`, `node_modules/`, `__pycache__/`, `target/`
- `bin/`, `build/`, `dist/`, `coverage/`, `*.egg-info/`

### Audit Areas

See [audit-areas.md](references/audit-areas.md) for configurable audit categories.

Default areas (extensible - add more by editing audit-areas.md):
- **package-structure**: Naming conventions, directory organization, layer separation, architectural violations
- **duplicate-code**: Repeated logic, similar functions, consolidation opportunities, copied code
- **interface-compliance**: Implementation completeness, signature mismatches, unused interfaces, method gaps
- **test-patterns**: Test organization, coverage gaps, mock usage consistency, test naming conventions
- **documentation-accuracy**: AGENTS.md vs actual code, missing types/methods, incorrect signatures
- **import-consistency**: Import ordering, circular dependencies, unused imports, external dependencies
- **error-handling**: Wrapping patterns, error type usage, message consistency, error chain breaks
- **mock-centralization**: Inline vs centralized mocks, duplicate mock implementations, missing mock implementations
- **go-patterns**: Go-specific patterns (see [go-patterns/](references/go-patterns/) directory)
- **security**: Common security vulnerabilities, secret handling, input validation
- **performance**: Common performance anti-patterns, inefficient algorithms, resource leaks
- **dead-code**: Unused code, unreachable code, commented-out production code

### Parallel Execution

Each audit area runs independently using parallel agent tasks. After completion:

1. **Deduplicate findings** - See [deduplication.md](references/deduplication.md) for fingerprint-based matching strategy
2. **Priority resolution**: When multiple areas report same issue, keep highest severity or most detailed
3. **Consolidate report**: Generate single comprehensive markdown with de-duplicated findings

### Output Format

Generates single comprehensive markdown report with:

1. **Executive Summary** - High-level overview, metrics, total findings
2. **Findings by Severity**
   - **CRITICAL**: Architectural violations, security issues, dead code in production
   - **HIGH**: Large duplications (>50 lines), naming conflicts, unused interfaces, architectural violations
   - **MEDIUM**: Test gaps, error handling inconsistencies, consolidation needs, documentation gaps
   - **LOW**: Cosmetic issues, nice-to-haves, style improvements, minor inconsistencies
3. **Recommendations** - Prioritized actionable items with severity
4. **Appendix** - Full findings list with file locations and line numbers

See [deduplication.md](references/deduplication.md) for fingerprint-based merging strategy.

## Depth Level Specifications

Each depth level controls detail granularity and example complexity:

| Depth Level | Description | Code Examples | Report Length |
|-------------|-------------|---------------|---------------|
| **Minimal** | One-line summary | None | Short (50-100 lines) |
| **Standard** | Full descriptions with brief snippets | 1-2 line snippets | Medium (200-400 lines) |
| **Detailed** | Complete analysis with function examples | 5-20 line functions | Long (600-1000 lines) |
| **Maximum** | Maximum detail with before/after blocks | 20+ line comparisons | Very long (1500+ lines) |

### Example Granularity by Depth Level

**Finding: Missing Slice Pre-allocation (HIGH)**

**Minimal depth output:**
```
Missing slice capacity pre-allocation in filterSuggestions() causes O(n²) reallocations.
```

**Standard depth output:**
```
Missing slice capacity pre-allocation in filterSuggestions() causes O(n²) reallocations.
Issue: result := make([]string, 0) lacks capacity parameter.
```

**Detailed depth output:**
```
Missing slice capacity pre-allocation in filterSuggestions() causes O(n²) reallocations.

func filterSuggestions(suggestions []string, partial string) []string {
    result := make([]string, 0) // No capacity - repeated allocations
    for _, suggestion := range suggestions {
        if strings.HasPrefix(suggestion, partial) {
            result = append(result, suggestion)
        }
    }
    return result
}
```

**Maximum depth output:**
```
Missing slice capacity pre-allocation in filterSuggestions() causes O(n²) reallocations.

**BAD CODE:**
func filterSuggestions(suggestions []string, partial string) []string {
    result := make([]string, 0) // No capacity - repeated allocations
    for _, suggestion := range suggestions {
        if strings.HasPrefix(suggestion, partial) {
            result = append(result, suggestion)
        }
    }
    return result
}

**GOOD CODE:**
func filterSuggestions(suggestions []string, partial string) []string {
    result := make([]string, 0, len(suggestions)) // Pre-allocate capacity
    for _, suggestion := range suggestions {
        if strings.HasPrefix(suggestion, partial) {
            result = append(result, suggestion)
        }
    }
    return result
}

**Impact:** Eliminates O(n²) allocation overhead in hot path called frequently.
```

## User Interaction (Required Before Execution)

This skill ALWAYS requires user configuration before audit execution. No defaults are inferred.

### Step 1: Select Depth Level

**Minimal** - One line per finding, no code examples
- Brief description only
- No code snippets
- Quick scan of problems

**Standard** - Detailed findings with brief examples
- Full descriptions
- 1-2 line code snippets
- Recommended for routine audits

**Detailed** - Full analysis with code snippets
- Complete descriptions
- Full function examples (5-20 lines)
- Before/after comparisons where useful

**Maximum** - Maximum detail with extensive examples
- Comprehensive descriptions
- Complete before/after code blocks (20+ lines)
- Extensive context and rationale
- Only use for deep-dive analysis

### Step 2: Select Priority Level

**Critical only** - Only CRITICAL severity findings
**Critical + High** - CRITICAL and HIGH severity findings
**Critical + High + Medium** - CRITICAL, HIGH, and MEDIUM severity findings
**All findings** - Include all findings including LOW severity

### Step 3: Positive Observations

**No (default)** - Exclude positive observations (e.g., "no issues found", "follows conventions")
**Yes** - Include positive observations in separate document

---

## Usage

```bash
# Audit entire codebase (auto-discovery)
audit-codebase-quality

# Audit specific directory
audit-codebase-quality /path/to/project

# Audit with maximum depth (default: 3)
audit-codebase-quality --max-depth 5

# Audit specific language
audit-codebase-quality --lang go

# Detailed mode (file-by-file findings, useful for large codebases)
audit-codebase-quality --detailed

# Include/exclude patterns
audit-codebase-quality --include '*/internal/*' --exclude '*/vendor/*'
```

## Finding Severity Standards

| Severity | Definition | Examples |
|----------|-------------|-----------|
| **CRITICAL** | Architectural violations, memory leaks, security vulnerabilities, string-based error detection | Memory leaks (unbounded caches), path traversal bypasses, `strings.Contains(err.Error())` instead of `errors.As()`, circular dependencies, interface in wrong layer |
| **HIGH** | Issues that significantly impact maintainability | Large duplications (>50 lines), naming conflicts, unused interfaces, critical files >100 lines without tests, cmd layer importing infrastructure, layer dependency violations |
| **MEDIUM** | Issues affecting code quality but not blocking | Test gaps (<100 lines files), error handling inconsistencies, missing documentation for 1-3 error types, consolidation opportunities |
| **LOW** | Cosmetic or optional improvements - MUST be actual problems | Style violations, minor inconsistencies, nice-to-haves, documentation typos. **NOTE: "No issues found" or "Follows conventions" are NOT findings and are excluded.** |

## Language-Specific Heuristics

See [go-patterns/](references/go-patterns/) directory for Go-specific audit patterns including:
- Package naming conventions
- File organization patterns
- Interface placement rules
- Error handling idioms
- Testing framework patterns

For other languages, add reference files following this pattern:
- `references/python-patterns.md`
- `references/javascript-patterns.md`
- `references/rust-patterns.md`

## Best Practices

1. **Focus on actionable findings**: Each finding should suggest concrete fix or improvement
2. **Provide context**: Include file paths and line numbers for all issues
3. **Prioritize ruthlessly**: Not everything needs fixing - triage by impact and effort
4. **Cross-reference**: Related findings should reference each other
5. **Architecture over style**: Naming conventions matter, but consistency and architecture matter more
6. **Document assumptions**: When using heuristics, explain reasoning clearly

## Real-World Examples

These are patterns successfully identified during actual codebase audits that demonstrate the skill's effectiveness:

### Memory Leak Detection (CRITICAL)

**Finding:** Unbounded repository cache in GoGitClient
```go
// ISSUE: internal/infrastructure/gogit_client.go:16-29
type GoGitClientImpl struct {
    cache        map[string]*git.Repository // Unbounded - grows forever
    cacheEnabled bool
}

// RECOMMENDATION: Use bounded LRU cache
import "github.com/hashicorp/golang-lru/v2"

type GoGitClientImpl struct {
    cache        *lru.Cache[string, *git.Repository] // Bounded cache
    cacheEnabled bool
}
```
**Impact:** Prevents gradual memory degradation as users navigate between repositories.

### String-Based Error Detection (CRITICAL)

**Finding:** Using `strings.Contains(err.Error())` instead of `errors.As()`
```go
// ISSUE: cmd/delete.go:86-90
errStr := err.Error()
if strings.Contains(errStr, "worktree not found") ||
    strings.Contains(errStr, "invalid git repository") ||
    strings.Contains(errStr, "repository does not exist") {
    return fmt.Errorf("worktree not found: %s", worktreePath)
}

// RECOMMENDATION: Use type-based error detection
var worktreeErr *domain.WorktreeServiceError
if errors.As(err, &worktreeErr) {
    return fmt.Errorf("worktree not found: %s", worktreeErr.WorktreePath)
}
```
**Impact:** Makes error handling fragile - breaks if error messages change.

### Path Traversal Vulnerability (HIGH)

**Finding:** Weak path traversal detection doesn't catch all vectors
```go
// ISSUE: internal/infrastructure/context_resolver.go:42-44
func containsPathTraversal(s string) bool {
    return strings.Contains(s, "..") || strings.Contains(s, string(filepath.Separator)+".")
}

// RECOMMENDATION: Clean path and detect changes
func containsPathTraversal(s string) bool {
    cleaned := filepath.Clean(s)
    if cleaned != s {
        return true  // Traversal detected
    }
    return strings.Contains(s, "%2e%2e") || strings.Contains(s, "%2E%2E")
}
```
**Impact:** Prevents bypass via URL-encoded sequences like `..//` or `%2e%2e`.

### Mock Duplication (CRITICAL)

**Finding:** MockCLIClient defined in two places
- Inline version: `internal/infrastructure/cli_client.go:324-377` (54 lines)
- Centralized version: `test/mocks/git_service_mock.go:85-126` (42 lines)

**Recommendation:** Remove inline version, standardize on centralized mock.

**Impact:** Eliminates maintenance burden and ensures consistent test behavior.

### Duplicate Code Consolidation (HIGH)

**Finding:** Shell wrapper templates duplicated across bash/zsh/fish (~125 lines)
```go
// ISSUE: Three nearly identical functions
func bashWrapperTemplate() string { /* 40 lines */ }
func zshWrapperTemplate() string { /* 40 lines */ }
func fishWrapperTemplate() string { /* 38 lines */ }

// RECOMMENDATION: Template-based approach
const baseWrapperTemplate = `### BEGIN TWIGGIT WRAPPER
# Twiggit {{SHELL_TYPE}} wrapper
twiggit() {{SHELL_FUNCTION_DEF}} {
    {{SHELL_CASE_BEGIN}} "$1" {{SHELL_CASE_END}}
        cd) {{SHELL_ACTION_PREFIX}}cd "$@"
            shift
            ;;
        create) {{SHELL_ACTION_PREFIX}}twiggit create "$@"
            shift
            ;;
        ...
    {{SHELL_CASE_CLOSE}}
}
### END TWIGGIT WRAPPER`
```
**Impact:** Reduces 125 lines to ~60 lines (52% reduction).

### Missing Critical Tests (CRITICAL)

**Finding:** Files >100 lines with no test coverage
- `internal/infrastructure/git_client.go` (136 lines) - CompositeGitClient routing logic
- `cmd/error_handler.go` (137 lines) - CLI error handling infrastructure

**Recommendation:** Create comprehensive test files covering all functions.

### N+1 Performance Issue (HIGH)

**Finding:** O(n*m) filesystem queries
```go
// ISSUE: Lists all projects, then queries worktrees for each
func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
    projects, err := s.projectService.ListProjects(ctx) // O(n)
    for _, project := range projects {
        worktrees, err := s.gitService.ListWorktrees(ctx, project.GitRepoPath) // O(m) per project
    }
}

// RECOMMENDATION: Parse path directly
func (s *worktreeService) findProjectByWorktree(ctx context.Context, worktreePath string) (*domain.ProjectInfo, error) {
    relPath, _ := filepath.Rel(s.config.WorktreesDirectory, worktreePath)
    parts := strings.Split(relPath, string(filepath.Separator))
    projectName := parts[0]
    return s.projectService.DiscoverProject(ctx, projectName, nil)
}
```
**Impact:** Reduces O(n*m) to O(1) for this common operation.

### Missing Error Type Documentation (CRITICAL)

**Finding:** 5 major error types completely undocumented
- `GitCommandError` - Git CLI command execution failures
- `ServiceError` - General service operation errors
- `ShellError` - Shell service errors with context
- `ResolutionError` - Path resolution errors with suggestions
- `ConflictError` - Operation conflict errors

**Recommendation:** Add complete documentation to `internal/domain/AGENTS.md`.

### Layer Dependency Violation (HIGH)

**Finding:** cmd layer imports infrastructure directly
```go
// ISSUE: cmd/root.go:27
import (
    "myapp/internal/infrastructure"
)

type ServiceContainer struct {
    GitClient infrastructure.GitClient  // Should use application layer
}

// RECOMMENDATION: Go through application services
type ServiceContainer struct {
    WorktreeService application.WorktreeService  // Clean architecture
}
```
**Impact:** Violates clean architecture principles, creates tight coupling.

### O(n²) Slice Operations (HIGH)

**Finding:** Missing slice pre-allocation causing repeated reallocations
```go
// ISSUE: No capacity pre-allocated
func filterSuggestions(suggestions []string, partial string) []string {
    result := make([]string, 0) // No capacity - repeated allocations
    for _, suggestion := range suggestions {
        if strings.HasPrefix(suggestion, partial) {
            result = append(result, suggestion)
        }
    }
    return result
}

// RECOMMENDATION: Pre-allocate capacity
func filterSuggestions(suggestions []string, partial string) []string {
    result := make([]string, 0, len(suggestions)) // Full capacity
    for _, suggestion := range suggestions {
        if strings.HasPrefix(suggestion, partial) {
            result = append(result, suggestion)
        }
    }
    return result
}
```
**Impact:** Eliminates O(n²) allocation overhead.

## Extending This Skill

### Adding New Audit Areas

1. Create new section in `references/audit-areas.md`:
   ```markdown
   ## Your New Area

   Check for [specific concerns or patterns].

   ### What to Look For
   - Pattern 1
   - Pattern 2
   - Pattern 3

    ### Examples

#### Audit Report Template

```markdown
# Audit Report: [Project Name]

## Executive Summary

**Audit Date**: [Date]
**Auditor**: AI Agent
**Codebase Scope**: [directories audited]
**Total Findings**: [number] (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)
**Depth Level**: [Minimal/Standard/Detailed/Maximum]
**Priority Filter**: [user selection]

[1-2 paragraph overview of key issues found]

## Findings by Severity

### CRITICAL

[Critical finding 1]

- **Severity**: CRITICAL
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Related Locations**: [other files]
- **Description**: [Clear description]
- **Recommendation**: [Specific fix with code example based on depth level]

[Additional critical findings...]

### HIGH

[High priority finding 1]

- **Severity**: HIGH
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Related Locations**: [other files]
- **Description**: [Clear description]
- **Recommendation**: [Specific fix with code example based on depth level]

[Additional high findings...]

### MEDIUM

[Similar structure...]

### LOW (if user selected "All findings")

[Similar structure...]

## Recommendations

1. [Critical recommendation with priority]
2. [High priority recommendation]
[Additional recommendations sorted by severity and impact...]

## Appendix

### Full Findings List

| ID | Severity | Area | Location | Pattern | Description |
|-----|-----------|------|----------|---------|-------------|
| 1 | CRITICAL | [area] | [file:line] | [pattern] | [description] |

[Complete findings table...]
```

#### Positive Observations Template (separate document, generated only if user selects "Yes")

```markdown
# Positive Observations: [Project Name]

**Audit Date**: [Date]
**Auditor**: AI Agent (Codebase Quality Audit)
**Codebase Scope**: [directories audited]
**Depth Level**: [Minimal/Standard/Detailed/Maximum]

## Areas with No Issues

### [Audit Area Name]

[Positive observations for this area - what's working well, following conventions]

[Additional audit areas...]

## Architecture Strengths

[Summary of positive architectural patterns, clean separation, good practices observed]

## Summary

[Optional: 1-2 paragraphs highlighting what's working particularly well]

## Recommendation

No changes needed in areas listed above. Focus on addressing findings in main audit report.
```

**Note**: Positive observations document is only generated when user explicitly selects "Yes" to "Include positive observations?" question.

### Severity Guidelines
- **CRITICAL**: Architectural violations, memory leaks, security vulnerabilities, string-based error detection
- **HIGH**: Large duplications (>50 lines), naming conflicts, unused interfaces, layer dependency violations
- **MEDIUM**: Test gaps, error handling inconsistencies, consolidation opportunities
- **LOW**: Cosmetic or optional improvements (must be actual problems, not positive observations)

### Language-Specific Notes
See [go-patterns/](references/go-patterns/) directory for Go-specific heuristics.

2. Implement audit logic in your workflow (consider parallel execution pattern)

### Adding Language Support

1. Create `references/<language>-patterns.md`:
   - Common naming conventions for that language
   - File organization patterns
   - Error handling idioms
   - Testing framework patterns
   - Build/dependency conventions
   - Common anti-patterns

2. Update `references/audit-areas.md` to reference language-specific patterns

3. Update discovery logic to recognize language file patterns

## Resources

### scripts/

Executable code for audit automation and tooling.

**Current:** Contains example.py from initialization (can be deleted or customized)

**Appropriate for:**
- Auto-discovery scripts (directory scanning, language detection)
- Deduplication utilities (fingerprint generation, matching logic)
- Report generation scripts (markdown formatting, statistics)

### references/

Documentation and reference material loaded as needed during audit.

- **audit-areas.md**: Configurable audit categories and checklists
- **validation-checklist.md**: Finding validation criteria (excluded from audit-areas.md)
- **go-patterns/**: Go-specific heuristics and patterns (directory with focused files)
  - **go-basics.md**: Package naming, error handling, interface design
  - **go-testing.md**: Testing patterns
  - **go-concurrency.md**: Context usage and concurrency
  - **go-configuration.md**: Configuration and dependency management
  - **go-performance.md**: Cache, slice, path validation, filesystem operations
  - **go-anti-patterns.md**: Common anti-patterns and audit-specific patterns
- **deduplication.md**: Strategy for merging overlapping findings from multiple areas
- **positive-observations-template.md**: Template for positive observations document (generated only when requested)

### assets/

Files used in output (templates, examples, etc.).

**Current:** Contains example_asset.txt from initialization (can be deleted if not needed)

**Appropriate for:**
- Report templates (markdown, HTML)
- Example code snippets for recommendations
- Documentation of common patterns

## Workflow

0. **User Configuration** (REQUIRED - No defaults)
   - Prompt user for depth level: Minimal/Standard/Detailed/Maximum
   - Prompt user for priority level: Critical/Critical+High/Critical+High+Medium/All
   - Prompt user for positive observations: No (default) / Yes
   - Store configuration for report generation

1. **Discovery**: Identify all source directories to audit using auto-discovery patterns
2. **Configuration**: Load audit areas from audit-areas.md, apply depth settings, include/exclude patterns
3. **Language detection**: Determine programming language(s) in codebase for language-specific heuristics
4. **Parallel execution**: Launch agents for each audit area using `task` tool with `explore` subagent type
5. **Finding Validation**: Filter out positive observations using validation checklists from each audit area
6. **Deduplication**: Merge overlapping findings using fingerprint-based matching (file:line + issue_type + pattern)
7. **Prioritization**: Sort by severity and impact, apply user priority filter
8. **Reporting**: Generate markdown report using user-selected depth level template
9. **Positive Observations** (optional): Generate separate document if user selected "Yes"

## Quality Standards

### Finding Quality

Each finding should include:
- **Clear description**: What the issue is
- **Evidence**: File paths and line numbers
- **Severity justification**: Why this severity was chosen
- **Recommendation**: Specific actionable fix or improvement
- **Impact estimate**: Lines of code affected, or estimated fix effort

### Report Quality

- **Accurate**: No false positives, verified against actual code
- **Complete**: All discovered issues reported
- **Prioritized**: Most critical issues first
- **Actionable**: Clear next steps with effort estimates
- **Context-aware**: Consider language, project size, team conventions
