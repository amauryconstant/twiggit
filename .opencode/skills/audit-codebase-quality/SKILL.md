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
- **Language-specific patterns**: See [go-patterns.md](references/go-patterns.md) for Go-specific heuristics

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
- **go-patterns**: Go-specific patterns (see [go-patterns.md](references/go-patterns.md))
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
| **CRITICAL** | Architectural violations that break design principles | Interface in wrong layer, circular dependencies, dead code in production path, security vulnerabilities |
| **HIGH** | Issues that significantly impact maintainability | Large duplications (>50 lines), naming conflicts, unused interfaces, architectural inconsistencies |
| **MEDIUM** | Issues affecting code quality but not blocking | Test gaps, error handling inconsistencies, missing documentation, consolidation opportunities |
| **LOW** | Cosmetic or optional improvements | Style violations, minor inconsistencies, nice-to-haves, documentation typos |

## Language-Specific Heuristics

See [go-patterns.md](references/go-patterns.md) for Go-specific audit patterns including:
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

#### Finding Summary Template

```markdown
# Audit Report: [Project Name]

## Executive Summary

**Audit Date**: [Date]
**Auditor**: AI Agent
**Codebase Scope**: [directories audited]
**Total Findings**: [number] (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)

[1-2 paragraph overview of key issues found]

## Findings by Severity

### CRITICAL

[Critical finding 1]

- **Severity**: CRITICAL
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Related Locations**: [other files]
- **Description**: [Clear description]
- **Recommendation**: [Specific fix]

[Additional critical findings...]

### HIGH

[High priority finding 1]

- **Severity**: HIGH
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Related Locations**: [other files]
- **Description**: [Clear description]
- **Recommendation**: [Specific fix]

[Additional high findings...]

### MEDIUM

[Medium priority finding 1]

- **Severity**: MEDIUM
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Description**: [Clear description]
- **Recommendation**: [Specific fix or improvement]

[Additional medium findings...]

### LOW

[Low priority finding]

- **Severity**: LOW
- **Area**: [audit area]
- **Primary Location**: [file:line]
- **Description**: [Clear description]
- **Recommendation**: [Optional improvement or style fix]

[Additional low findings...]

## Recommendations

1. [Critical recommendation]
2. [High priority recommendation]
[Additional recommendations...]

## Appendix

### Full Findings List

| ID | Severity | Area | Location | Pattern | Description |
|-----|-----------|------|----------|---------|-------------|
| 1 | CRITICAL | [area] | [file:line] | [pattern] | [description] |

[Complete findings table...]
```

#### Simple Finding Example

```markdown
## Finding

- **Severity**: HIGH
- **Area**: duplicate-code
- **Location**: internal/domain/shell.go:104-159
- **Pattern**: shell_wrapper_templates
- **Description**: Duplicate shell wrapper templates (bash, zsh, fish) defined in both domain/shell.go and infrastructure/shell_infra.go
- **Recommendation**: Consolidate to infrastructure/shell_infra.go, remove from domain/shell.go (eliminates ~90 lines)
```

### Severity Guidelines
- **CRITICAL**: [specific conditions]
- **HIGH**: [specific conditions]
- **MEDIUM**: [specific conditions]
- **LOW**: [specific conditions]

### Language-Specific Notes
See [go-patterns.md](references/go-patterns.md) for language-specific heuristics.

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
- **go-patterns.md**: Go-specific heuristics and patterns
- **deduplication.md**: Strategy for merging overlapping findings from multiple areas

### assets/

Files used in output (templates, examples, etc.).

**Current:** Contains example_asset.txt from initialization (can be deleted if not needed)

**Appropriate for:**
- Report templates (markdown, HTML)
- Example code snippets for recommendations
- Documentation of common patterns

## Workflow

1. **Discovery**: Identify all source directories to audit using auto-discovery patterns
2. **Configuration**: Load audit areas from audit-areas.md, apply depth settings, include/exclude patterns
3. **Language detection**: Determine programming language(s) in codebase for language-specific heuristics
4. **Parallel execution**: Launch agents for each audit area using `task` tool with `explore` subagent type
5. **Deduplication**: Merge overlapping findings using fingerprint-based matching (file:line + issue_type + pattern)
6. **Prioritization**: Sort by severity and impact
7. **Reporting**: Generate comprehensive markdown report using template pattern from output-patterns.md

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
