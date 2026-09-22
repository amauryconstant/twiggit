# Spec Delta

## MODIFIED Requirements

### Requirement: Context-detection priority chain

The system SHALL detect context with the following priority (first match wins):
1. **Worktree** — CWD is under `<core.Config.WorktreesDirectory>/<project>/<branch>/` and contains a `.git` file (worktree pointer).
2. **Project** — CWD is under `<core.Config.ProjectsDirectory>/<project>/` and contains a `.git` directory.
3. **Outside git** — none of the above.

#### Scenario: Priority chain matches the documented shape
- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Path-utils dependency moved to internal/core

When checking whether `dir` is under `core.Config.ProjectsDirectory` or `core.Config.WorktreesDirectory`, the resolver SHALL resolve symlinks (via `core.NormalizePath` and `core.IsPathUnder` from `internal/core/`) before comparing, to prevent symlink-based path-traversal bypasses.

#### Scenario: Path-utils helpers are imported from internal/core
- **WHEN** the resolver package's import list is read
- **THEN** the import SHALL be `twiggit/internal/core` (not `twiggit/internal/infrastructure` or `twiggit/internal/domain`)
