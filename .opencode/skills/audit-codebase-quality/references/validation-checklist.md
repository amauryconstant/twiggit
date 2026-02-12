# Finding Validation Checklist

All findings from ALL audit areas MUST meet these criteria. This is the single source of truth for finding validation.

## Required Criteria (ALL must be met)

A finding MUST have:
- ✗ **Problem exists** - Something is broken, missing, or needs improvement
- ✗ **Actionable** - Can be fixed with specific steps or code change
- ✗ **Impact measurable** - Why this matters (security, performance, maintainability, etc.)
- ✗ **Severity justified** - Reason for assigned severity is clear

## Finding Exclusion Rules (Do NOT report as findings)

These are **NOT findings** and should be excluded from main audit report:

❌ "No action needed" → Positive observation
❌ "Follows conventions correctly" → Positive observation
❌ "Structure is appropriate" → Positive observation
❌ "All tests pass" → Positive observation
❌ "No circular dependencies" → Positive observation
❌ "No issues found" → Positive observation
❌ "Proper implementation" → Positive observation
❌ "Correct pattern" → Positive observation

## Validation Examples

### VALID Finding (Report)

```
- Severity: HIGH
- Problem: Missing slice pre-allocation causes O(n²) performance
- Action: Add capacity parameter to make()
- Impact: Significant performance degradation in hot path
- Justification: Affects filterSuggestions() called frequently
```

### INVALID (Positive Observation - Exclude)

```
- Severity: LOW
- Observation: Package structure follows clean architecture
- Problem: None
- Action: No action needed
→ This is a positive observation, NOT a finding
```

### INVALID (Not Actionable - Exclude)

```
- Severity: MEDIUM
- Problem: Code style could be improved
- Action: "Consider refactoring" (vague, not actionable)
- Impact: Subjective, not measurable
→ Too vague to be a finding
```

### INVALID (No Problem - Exclude)

```
- Severity: LOW
- Observation: Domain layer has no dependencies on other internal packages
- Action: No action needed
→ Correct architecture, NOT a finding
```

## Usage in Audit Areas

Each audit area in `areas.md` includes: "See [validation-checklist.md](validation-checklist.md)."
