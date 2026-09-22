# Spec Delta

## Purpose

Documents the context-detection priority chain (worktree > project > outside git) for the Tier 2 layout. Path-traversal protection uses `core.NormalizePath` and `core.IsPathUnder` from `internal/core/`. The legacy `infrastructure-context-resolver` spec is left intact per the deferred-migration non-goal in the proposal.

## ADDED Requirements

### Requirement: Context-detection priority chain

The system SHALL detect context with the following priority (first match wins):

1. **Worktree** — CWD is under `<core.Config.WorktreesDirectory>/<project>/<branch>/` and contains a `.git` file (worktree pointer).
2. **Project** — CWD is under `<core.Config.ProjectsDirectory>/<project>/` and contains a `.git` directory.
3. **Outside git** — none of the above.

#### Scenario: Worktree context wins over project context
- **WHEN** CWD is `$HOME/Worktrees/myapp/feat/foo` with a `.git` file pointing to the main repo
- **THEN** the resolved context is `core.ContextWorktree`

#### Scenario: Project context when not in a worktree
- **WHEN** CWD is `$HOME/Projects/myapp` with a `.git` directory
- **THEN** the resolved context is `core.ContextProject`

#### Scenario: Outside git when no patterns match
- **WHEN** CWD is `/tmp/random` with no `.git` file or directory
- **THEN** the resolved context is `core.ContextOutsideGit`

### Requirement: Path-utils dependency moved to internal/core

When checking whether `dir` is under `core.Config.ProjectsDirectory` or `core.Config.WorktreesDirectory`, the resolver SHALL resolve symlinks (via `core.NormalizePath` and `core.IsPathUnder` from `internal/core/`) before comparing, to prevent symlink-based path-traversal bypasses.

#### Scenario: Path-utils helpers are imported from internal/core
- **WHEN** the resolver package's import list is read
- **THEN** the import SHALL be `twiggit/internal/core` (not `twiggit/internal/infrastructure` or `twiggit/internal/domain`)

#### Scenario: Symlink-resolved path comparison
- **WHEN** `dir = /tmp/sneaky` is a symlink to `/tmp/real` and `/tmp/real` is inside `core.Config.ProjectsDirectory`
- **THEN** the resolver classifies `dir` as `core.ContextProject` after symlink resolution
