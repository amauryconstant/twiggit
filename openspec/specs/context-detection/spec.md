## ADDED Requirements

### Requirement: Context Detection

The system SHALL automatically detect the user's current git context (project, worktree, or outside git) to enable context-aware command behavior. Detection SHALL follow a priority order and SHALL cache results for performance.

#### Scenario: Detect project context from directory with .git

- **WHEN** user runs any command from a directory containing `.git/` folder
- **THEN** system SHALL detect `ContextProject` context type
- **AND** project name SHALL be extracted from the directory name
- **AND** context path SHALL be set to the directory containing `.git/`
- **AND** detection SHALL traverse up the directory tree until finding `.git/` or reaching filesystem root

#### Scenario: Detect worktree context from worktree directory pattern

- **WHEN** user runs any command from a directory matching `~/Worktrees/<project>/<branch>/` pattern
- **AND** the directory contains a `.git` file (not directory) with "gitdir:" content
- **THEN** system SHALL detect `ContextWorktree` context type
- **AND** project name SHALL be extracted from the first level of the worktree path
- **AND** branch name SHALL be extracted from the second level of the worktree path
- **AND** worktree detection SHALL take priority over project detection when both patterns match

#### Scenario: Detect outside git context

- **WHEN** user runs any command from a directory with no `.git/` directory or worktree pattern match
- **THEN** system SHALL detect `ContextOutsideGit` context type
- **AND** explanation SHALL indicate "Not in a git repository or worktree"
- **AND** context SHALL be treated as outside of any git workspace

#### Scenario: Return error for invalid directory

- **WHEN** user runs any command with an empty directory path
- **THEN** system SHALL return error with message "empty directory path"
- **AND** error SHALL be wrapped in context detection error

#### Scenario: Return error for nonexistent directory

- **WHEN** user runs any command with a directory path that does not exist
- **THEN** system SHALL return error with message "directory does not exist"
- **AND** error SHALL be wrapped in context detection error

#### Scenario: Validate worktree before returning worktree context

- **WHEN** directory path matches worktree pattern but `.git` file is not present
- **OR** `.git` is a directory instead of a file
- **OR** `.git` file content does not contain "gitdir:"
- **THEN** system SHALL NOT return `ContextWorktree`
- **AND** system SHALL fall through to project context detection or outside git context

#### Scenario: Cache detection results with normalized paths

- **WHEN** user runs multiple commands from the same directory
- **OR** user uses different path formats for the same location (e.g., absolute path vs. `./path`)
- **THEN** system SHALL use normalized, symlink-resolved paths as cache keys
- **AND** subsequent detections SHALL return cached context object
- **AND** cache SHALL persist across multiple detections for performance

#### Scenario: Traverse nested directories to find project context

- **WHEN** user runs command from a nested subdirectory within a project
- **THEN** system SHALL traverse up the directory tree until finding `.git/` directory
- **AND** project name SHALL be extracted from the directory containing `.git/`
- **AND** traversal SHALL stop at filesystem root if no `.git/` is found

#### Scenario: Handle cross-platform path handling

- **WHEN** system runs on Windows with Windows-style paths
- **OR** system runs on Unix with Unix-style paths
- **THEN** path operations SHALL use platform-appropriate separators
- **AND** context detection SHALL work correctly on both platforms
