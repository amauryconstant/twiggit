# Deduplication Strategy

When multiple audit areas report findings for the same issues, use this fingerprint-based matching strategy to consolidate into a single prioritized list.

## Fingerprint-Based Matching

Each finding has a unique fingerprint composed of:

```
file_path + line_number + issue_type + pattern
```

### Example Fingerprints

```
internal/services/worktree_service.go:40:validation + error_wrapping
internal/infrastructure/context_detector.go:55:validation + error_context_loss
internal/services/project_service.go:82:validation + error_type_mismatch
```

When two findings have matching fingerprints, they're considered duplicates:

1. **Exact match**: Same file, line, issue type, and pattern
2. **Pattern match**: Same file, issue type, and pattern type (different lines or values)
3. **Cross-area match**: Same issue type and pattern across different audit areas

## Resolution Strategy

When duplicates are detected, resolve using these rules:

### Rule 1: Priority Resolution

**Keep the finding with highest severity.**

| Finding A | Finding B | Result |
|----------|-----------|--------|
| CRITICAL | HIGH | Keep A |
| HIGH | MEDIUM | Keep A |
| MEDIUM | LOW | Keep A |
| CRITICAL | CRITICAL | Keep most detailed (or merge details) |
| HIGH | HIGH | Merge details from both |

### Rule 2: Detail Enrichment

When severities are equal, merge details:

- **Keep the most detailed finding**: If Finding A has more context/examples, keep it
- **Add missing context**: If Finding B has unique details, append to Finding A
- **Cross-reference findings**: List both as references (e.g., "Also reported in audit area X")

### Rule 3: Cross-Area Priority

When different audit areas report same issue:

- **Prefer architectural issues** over test or documentation issues
- **Prefer duplicate code** over minor style issues
- **Prefer security issues** over all others
- **Audit area priority order**: package-structure > duplicate-code > interface-compliance > test-patterns > documentation > import > error-handling > mock > security > performance > dead-code

### Rule 4: Location Consolidation

When same issue occurs in multiple files:

- **Primary location**: File with most impact or most occurrences
- **Related locations**: List other files where issue occurs
- **Pattern summary**: If issue follows a pattern, describe it generically

### Manual Review

Automated deduplication should be followed by manual review to catch:

- **False negatives**: Duplicates not detected (e.g., different file paths but same issue)
- **False positives**: Different issues incorrectly marked as duplicates
- **Contextual differences**: Same pattern but different meaning or impact

### Edge Cases to Consider

When matching fingerprints, handle these special cases:

1. **Findings without line numbers**:
   - Package-level or directory-level issues
   - Example: Package naming violations (entire package)
   - Fingerprint: `<package_path>:<empty>:naming:pattern_name`

2. **Findings with multiple line numbers**:
   - Issues that occur in many places in a file
   - Example: Repeated validation pattern
   - Fingerprint: `<file_path>:<line_start>-<line_end>:error_handling:pattern`
   - Keep earliest occurrence, mark others as duplicates with references

3. **Findings with fuzzy or partial matches**:
   - Similar but not identical patterns
   - Use pattern matching over exact string comparison
   - Fingerprint: `<file_path>:<line_number>:issue_type:pattern:<match_score>`

4. **Cross-area priority conflicts**:
   - If same issue has different severity in different areas, determine based on:
     - Which area has greater overall impact on codebase?
     - Which finding has more detailed evidence or examples?
   - Document priority resolution reasoning

5. **Pattern variations**:
   - Different patterns that represent same underlying issue
   - Merge into single finding with consolidated pattern description
   - Example: `mixed_wrapping_styles` from one area and `error_type_mismatch` from another

6. **Location-based consolidation**:
   - When same issue occurs in multiple files within one area
   - Select primary location (most occurrences or highest impact)
   - List all related locations as secondary references
   - Example: Path validation repeated 5 times in `context_resolver.go`

## Common Cross-Area Duplicates

### Error Handling Inconsistencies

Same error handling issue can be reported by multiple areas:

- **Import consistency**: Mixed import styles or circular dependencies
- **Error handling**: Inconsistent wrapping patterns, message casing
- **Test patterns**: Mock-related issues
- **Documentation**: Missing error types in docs

**Example:**
```
Area A (error-handling): Inconsistent error wrapping - fmt.Errorf vs domain types
Area B (import-consistency): Mixed import ordering - stdlib not first

→ Merge: "Inconsistent patterns" severity MEDIUM, consolidate examples
```

### Naming Inconsistencies

Same naming issue can be reported by multiple areas:

- **Package structure**: Naming convention violations
- **Test patterns**: Mock naming inconsistencies
- **Interface compliance**: Unused interfaces (also dead code)

**Example:**
```
Area A (package-structure): services package uses plural (should be service)
Area B (test-patterns): MockWorktreeService duplicated

→ Merge: "Naming inconsistencies" severity HIGH, list both issues
```

### Interface and Implementation Issues

Same interface/implementation problem reported by multiple areas:

- **Interface compliance**: Missing methods, unused interfaces
- **Documentation accuracy**: Incorrect signatures in docs
- **Dead code**: Unused interfaces, unused functions

**Example:**
```
Area A (interface-compliance): GitRepositoryInterface never implemented
Area B (documentation-accuracy): GitRepositoryInterface documented but missing methods
Area C (dead-code): GitRepositoryInterface is dead code

→ Merge: "GitRepositoryInterface issues" severity HIGH, consolidate details
```

## Merging Process

### Step 1: Collect All Findings

Gather findings from all parallel audit agents into a list:

```json
[
  {
    "audit_area": "package-structure",
    "file_path": "internal/services",
    "line_number": "N/A",
    "issue_type": "naming",
    "pattern": "plural_package_name",
    "severity": "MEDIUM",
    "description": "services package uses plural naming while all other packages use singular"
  },
  {
    "audit_area": "error-handling",
    "file_path": "internal/services/worktree_service.go",
    "line_number": "40",
    "issue_type": "error_wrapping",
    "pattern": "mixed_wrapping_styles",
    "severity": "HIGH",
    "description": "Mixed fmt.Errorf and domain.NewWorktreeServiceError() usage"
  }
]
```

### Step 2: Generate Fingerprints

For each finding, create fingerprint:

```
fingerprint = file_path.replace(/internal\/.*/, "") + ":" + line_number + ":" + issue_type + ":" + pattern
```

Example:
```
services/worktree_service.go:40:error_wrapping:mixed_wrapping_styles
context_detector.go:55:error_context_loss:context_loss
```

### Step 3: Detect Duplicates

Compare fingerprints using these matching rules:

**Exact match:**
```
fingerprint_A == fingerprint_B
```

**Pattern match:**
```
file_path_A == file_path_B &&
issue_type_A == issue_type_B &&
pattern_A == pattern_B
```

**Cross-area pattern match:**
```
issue_type_A == issue_type_B &&
pattern_A == pattern_B
```

### Step 4: Resolve Conflicts

For each duplicate group, apply resolution strategy:

1. **Sort by severity** (CRITICAL > HIGH > MEDIUM > LOW)
2. **Apply Rule 1** (keep highest severity)
3. **Apply Rule 2** (enrich with details)
4. **Apply Rule 3** (cross-area priority)
5. **Apply Rule 4** (location consolidation)

### Step 5: Generate Consolidated Report

Create final findings list with:

- **Merged findings** (deduplicated, enriched)
- **Cross-references** (point to related findings)
- **Primary location** (where issue occurs most)
- **Related locations** (other files with same issue)

## Examples

### Example 1: Simple Duplicate

**Finding A (from package-structure):**
- File: `internal/services`
- Issue: "services package uses plural naming"
- Severity: MEDIUM

**Finding B (from test-patterns):**
- File: `test/mocks/services/worktree_service_mock.go`
- Issue: "MockWorktreeService duplicated in test/mocks/services/"
- Severity: CRITICAL

**Fingerprints:**
- A: `services:naming:plural_package_name:MEDIUM`
- B: `test/mocks/services/worktree_service_mock.go:mocks:duplicate_mock:CRITICAL`

**Resolution:**
- Keep Finding B (CRITICAL > MEDIUM)
- Add reference: "Related: services package naming issue reported in package-structure audit"

### Example 2: Cross-Area Duplicate

**Finding A (from error-handling):**
- File: `internal/services/worktree_service.go:40`
- Issue: "Mixed error wrapping - fmt.Errorf vs domain types"
- Severity: HIGH

**Finding B (from import-consistency):**
- File: `internal/infrastructure/context_detector.go:55`
- Issue: "Loss of error context - no %w wrapping"
- Severity: HIGH

**Fingerprints:**
- A: `worktree_service.go:40:error_handling:mixed_wrapping:HIGH`
- B: `context_detector.go:55:error_handling:error_context_loss:HIGH`

**Resolution:**
- Severities equal (both HIGH)
- Enrich: Merge details from both findings
- Primary location: `internal/services/worktree_service.go` (more impact)
- Related locations: `internal/infrastructure/context_detector.go:55`

**Consolidated finding:**
```
**File**: internal/services/worktree_service.go (primary)
**Related**: internal/infrastructure/context_detector.go:55

**Issue**: Inconsistent error wrapping patterns
- WorktreeService mixes fmt.Errorf and domain.NewWorktreeServiceError()
- ContextDetector loses error context by not using %w

**Severity**: HIGH
**Pattern**: Mixed error handling across layers

Both issues demonstrate lack of consistent error handling strategy.
```

### Example 3: Pattern Consolidation

**Finding A (from duplicate-code):**
- File: `internal/infrastructure/context_resolver.go:133-136`
- Issue: "Path validation error handling repeated"
- Severity: MEDIUM

**Finding B (from duplicate-code):**
- File: `internal/infrastructure/context_resolver.go:240-243`
- Issue: "Path validation error handling repeated"
- Severity: MEDIUM

**Finding C (from duplicate-code):**
- File: `internal/infrastructure/context_resolver.go:272-275`
- Issue: "Path validation error handling repeated"
- Severity: MEDIUM

**Finding D (from duplicate-code):**
- File: `internal/infrastructure/context_resolver.go:339-342`
- Issue: "Path validation error handling repeated"
- Severity: MEDIUM

**Finding E (from duplicate-code):**
- File: `internal/infrastructure/context_resolver.go:449-452`
- Issue: "Path validation error handling repeated"
- Severity: MEDIUM

**Fingerprints:**
- All: `context_resolver.go:error_handling:repeated_validation:MEDIUM`

**Resolution:**
- Same file, same issue type, same pattern → **Pattern match**
- Consolidate into single finding:
  - File: `internal/infrastructure/context_resolver.go`
  - Issue: "Path validation error handling pattern repeated 5 times (lines 133-136, 240-243, 272-275, 339-342, 449-452)"
  - Severity: MEDIUM
  - Recommendation: "Extract into helper function"

## Implementation Notes

### Automation

Deduplication should be automated in the audit workflow:

1. Store all findings in structured format (JSON or internal data structure)
2. Generate fingerprints for each finding
3. Detect duplicates using matching rules
4. Apply resolution strategy programmatically
5. Generate consolidated findings list

### Manual Review

Automated deduplication should be followed by manual review to catch:

- **False negatives**: Duplicates not detected (e.g., different file paths but same issue)
- **False positives**: Different issues incorrectly marked as duplicates
- **Contextual differences**: Same pattern but different meaning or impact

### Reporting

In final report, indicate when findings were merged:

```
## Findings by Severity

### HIGH

**Duplicate Error Handling Patterns**
- Pattern: Inconsistent error wrapping and context loss across layers
- Primary: internal/services/worktree_service.go:40
- Related: internal/infrastructure/context_detector.go:55
- **Consolidated from**: error-handling, import-consistency audits

[Description of consolidated issue...]

[Recommendations...]
```

This makes it clear to readers that multiple raw findings were merged into a single actionable item.
