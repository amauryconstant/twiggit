## Context

Version information is currently defined in `cmd/version.go` as package-level variables (`Version`, `Commit`, `Date`). This mixes infrastructure-level build metadata with presentation layer code. The germinator pattern establishes a dedicated version package at the infrastructure level for cleaner separation.

Current state:
- Version variables in `cmd/version.go`
- GoReleaser ldflags target `twiggit/cmd.Version`, `twiggit/cmd.Commit`, `twiggit/cmd.Date`
- Local build tasks in `.mise/config.toml` target `twiggit/cmd.version`, `twiggit/cmd.commit`, `twiggit/cmd.date`
- `version` command in cmd layer references these directly

## Goals / Non-Goals

**Goals:**
- Extract version variables to `internal/version` package
- Maintain identical version command output behavior
- Update GoReleaser ldflags to new package path
- Update local build configuration to use new package path

**Non-Goals:**
- Changing version output format
- Adding new version-related features
- Modifying how other commands access version info (none currently do)

## Decisions

1. **Package Location: `internal/version/version.go`**
    - Rationale: Follows germinator pattern; version is infrastructure-level data injected at build time
    - Alternatives: `internal/infrastructure/version/` - rejected as unnecessary nesting for single-file package

2. **Variable Export: Capitalized names (Version, Commit, Date)**
    - Rationale: Required for ldflags injection and external package access
    - Alternatives: Getter functions - rejected as over-engineering for simple data holders

  3. **String() Function Format: `<version> (<full-commit>) <date>`**
      - Rationale: Matches current output format exactly
      - Behavior: Returns `<version> () ` with empty parens and trailing space when commit is empty (dev builds); returns `<version> (<commit>) ` with trailing space when date is empty but commit is present
      - Note: If commit is ≤7 chars, use as-is (no truncation)
      - Note: String() does NOT include "twiggit " prefix - the command layer prepends it

4. **Build Configuration Updates: Both Release and Local**
    - Rationale: Both GoReleaser (for releases) and local build tasks (mise run build) must inject version information correctly
    - GoReleaser: Update ldflags from `twiggit/cmd.*` to `twiggit/internal/version.*`
    - Local builds: Update ldflags in both `build` and `build:local` tasks from `twiggit/cmd.*` to `twiggit/internal/version.*`
    - Note: Note the case difference - GoReleaser uses `Version` (capitalized), local builds use `version` (lowercase) due to Go syntax

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| GoReleaser ldflags path mismatch | Verify with `mise run build` and test version output |
| Local build ldflags path mismatch | Update both build and build:local tasks in .mise/config.toml |
| Breaking existing imports (none exist) | N/A - version package is new |
