# Dead-Code Report

Fields or types defined in the codebase that have no current
producer or consumer. Each entry lists the location, the evidence
(no call site / no read site), and an action menu.

---

## `domain.ResolutionSuggestion.Remote`

- **Defined:** `internal/domain/context.go:88` — `Remote string` field
  on `domain.ResolutionSuggestion`
- **Doc comment:** "Remote contains remote tracking info (e.g.,
  'origin/branch')"
- **Producers:** none. `rg "ResolutionSuggestion{" internal/`
  shows constructors set only `IsCurrent` and `IsDirty`. The
  `ListBranches` impl in `gogit_client.go` populates
  `domain.BranchInfo.Remote` (different struct) but never assigns
  the same value to `ResolutionSuggestion.Remote`.
- **Consumers:** none. `rg "\.Remote\b" internal/ cmd/ test/` shows
  every match on the `domain.ResolutionSuggestion.Remote` field is
  the definition itself.
- **Likely intent:** populate from `BranchInfo.Remote` when building
  suggestions for `--output=json` consumers or for completion
  descriptions showing upstream tracking info.
- **Action menu:**
  - (a) Populate the field from `BranchInfo.Remote` in the resolver,
    document in `domain-context-types`, and surface via JSON output.
  - (b) Delete the field. Update `domain-context-types` to reflect
    the smaller surface. (Lowest risk.)

## `domain.ResolutionSuggestion.StyleHint`

- **Defined:** `internal/domain/context.go:91` — `StyleHint string`
  field on `domain.ResolutionSuggestion`
- **Doc comment:** "StyleHint provides styling information for
  display"
- **Producers:** none. No constructor or builder sets this field.
- **Consumers:** none. `rg "\.StyleHint" internal/ cmd/ test/`
  shows only the definition.
- **Likely intent:** allow the resolver to tag suggestions with
  styling metadata (e.g., `"warning"`, `"info"`, `"default"`) that
  downstream Carapace or JSON consumers can use to colorize output.
- **Action menu:**
  - (a) Decide on the StyleHint vocabulary and populate it from the
    resolver.
  - (b) Delete the field. (Lowest risk.)

## Resolution

These fields pre-date the introduction of `BranchInfo.Remote` and
the JSON-output feature in `cli-output-formats`. The current JSON
output (`twiggit list -o json`) does not include them, so deleting
them has no consumer-visible effect today.

**Recommendation:** open a separate refactor change to either
populate the fields (option a) or delete them (option b). Don't
block the spec restructure on this decision.
