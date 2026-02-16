## Purpose

Resolve target identifiers (branches, projects, worktrees) to concrete filesystem paths based on current context, enabling the `cd` command to navigate between worktrees and projects. Also provides progressive completion for cross-project references and filtering for existing worktrees.

## Requirements

### Requirement: Path Resolution

The system SHALL resolve target identifiers (branches, projects, worktrees) to concrete filesystem paths based on current context, enabling the `cd` command to navigate between worktrees and projects.

#### Scenario: Resolve worktree from branch name in project context

- **WHEN** user runs `twiggit cd feature-branch` from within a project directory
- **AND** feature branch worktree exists for the current project
- **THEN** system SHALL resolve path to `~/Worktrees/<current-project>/feature-branch/`
- **AND** resolved path type SHALL be `PathTypeWorktree`
- **AND** absolute path SHALL be output to stdout

#### Scenario: Resolve project main directory from project context

- **WHEN** user runs `twiggit cd main` from within a project directory or worktree
- **THEN** system SHALL resolve path to the main project directory
- **AND** resolved path type SHALL be `PathTypeProject`
- **AND** absolute path SHALL be output to stdout

#### Scenario: Resolve worktree from branch name in worktree context

- **WHEN** user runs `twiggit cd different-branch` from within a worktree
- **AND** different branch worktree exists for the same project
- **THEN** system SHALL resolve path to `~/Worktrees/<current-project>/different-branch/`
- **AND** resolved path type SHALL be `PathTypeWorktree`
- **AND** absolute path SHALL be output to stdout

#### Scenario: Resolve cross-project worktree

- **WHEN** user runs `twiggit cd otherproject/feature-branch`
- **AND** specified worktree exists
- **THEN** system SHALL resolve path to `~/Worktrees/otherproject/feature-branch/`
- **AND** resolved path type SHALL be `PathTypeWorktree`
- **AND** absolute path SHALL be output to stdout

#### Scenario: Resolve cross-project main directory

- **WHEN** user runs `twiggit cd otherproject` from any context
- **AND** other project exists
- **THEN** system SHALL resolve path to `~/Projects/otherproject/`
- **AND** resolved path type SHALL be `PathTypeProject`
- **AND** absolute path SHALL be output to stdout

#### Scenario: Return error for invalid target

- **WHEN** user runs `twiggit cd` with a target that cannot be resolved
- **THEN** system SHALL return error indicating invalid target format
- **OR** error SHALL indicate worktree or project not found
- **AND** error SHALL be descriptive of what went wrong

#### Scenario: Return error for empty target outside git context

- **WHEN** user runs `twiggit cd` with no arguments
- **AND** current context is `ContextOutsideGit`
- **THEN** system SHALL return error "no target specified and no default worktree in context"
- **AND** no further path resolution SHALL be attempted

#### Scenario: Use default target in project context

- **WHEN** user runs `twiggit cd` with no arguments from within a project directory
- **THEN** system SHALL use "main" as the default target
- **AND** system SHALL resolve path to the main project directory
- **AND** absolute path SHALL be output to stdout

#### Scenario: Use default target in worktree context

- **WHEN** user runs `twiggit cd` with no arguments from within a worktree
- **THEN** system SHALL use the current branch name as default target
- **AND** system SHALL navigate to the same worktree (idempotent operation)
- **AND** absolute path SHALL be output to stdout

#### Scenario: Validate resolved path exists

- **WHEN** system resolves a target to a path
- **THEN** system SHALL validate that the resolved path exists
- **AND** system SHALL validate that the path is accessible
- **AND** system SHALL validate that the path is a directory
- **AND** if validation fails, error SHALL be returned with appropriate message

#### Scenario: Return error for nonexistent resolved path

- **WHEN** system resolves a target to a path that does not exist
- **THEN** system SHALL return error indicating path does not exist
- **AND** error SHALL include the path that was resolved
- **AND** error type (worktree vs project) SHALL be indicated in the message

#### Scenario: Provide completion suggestions

- **WHEN** user requests navigation suggestions (for shell completion)
- **THEN** system SHALL provide list of possible targets based on current context
- **AND** suggestions SHALL include project names and branch names
- **AND** suggestions SHALL be limited to configured maximum (if configured)
- **AND** suggestions SHALL be filtered to be relevant to current project if in project context

#### Scenario: Resolve relative path to absolute path

- **WHEN** system validates a path that is relative
- **THEN** system SHALL convert to path to absolute
- **AND** absolute path SHALL be used for all further operations
- **AND** system SHALL handle both Unix and Windows path formats correctly

#### Scenario: Reject path traversal sequences

- **WHEN** user provides a target containing path traversal sequences (e.g., `..`, `../`, `./.`, `./`)
- **THEN** system SHALL reject the identifier
- **AND** error SHALL indicate "project or branch name contains path traversal sequences"
- **AND** resolution SHALL not proceed
- **AND** system SHALL protect against accessing files outside intended directories

#### Scenario: Validate path is under configured directories

- **WHEN** system resolves a worktree path
- **THEN** system SHALL validate path is under configured worktrees directory
- **AND** if path is outside worktrees directory, error SHALL be returned
- **AND** error SHALL indicate "worktree path is outside configured worktrees directory"
- **AND** same validation SHALL apply for project paths against projects directory

### Requirement: Progressive Cross-Project Completion

The system SHALL support progressive completion for cross-project references when the partial input contains a `/`.

#### Scenario: Complete branches after project prefix

- **WHEN** user requests completion with partial input matching `<project>/`
- **THEN** system SHALL detect the project name before the `/`
- **AND** system SHALL fetch branches from that project's repository
- **AND** suggestions SHALL be formatted as `<project>/<branch>`
- **AND** suggestions SHALL include descriptions for each branch

#### Scenario: Complete branches with partial branch name

- **WHEN** user requests completion with partial input matching `<project>/fea`
- **THEN** system SHALL suggest only branches starting with "fea" from that project
- **AND** suggestions SHALL be formatted as `<project>/feature-*`

#### Scenario: Handle nonexistent project in cross-project completion

- **WHEN** user requests completion with partial input matching `<nonexistent-project>/`
- **AND** the project does not exist in configured projects directory
- **THEN** system SHALL return empty suggestions
- **AND** system SHALL NOT return an error (graceful degradation)

#### Scenario: Handle slow or unreachable project repository

- **WHEN** user requests completion with partial input matching `<project>/`
- **AND** the project's repository cannot be accessed within 500ms timeout
- **THEN** system SHALL gracefully degrade by returning empty suggestions for that project
- **AND** system SHALL NOT return an error

### Requirement: Existing Worktree Filter

The system SHALL support filtering suggestions to include only materialized worktrees.

#### Scenario: Filter to existing worktrees only

- **WHEN** completion is requested with existing-worktree-only option enabled
- **THEN** system SHALL return only branches that have materialized worktrees
- **AND** system SHALL NOT include branches without worktrees
- **AND** system SHALL NOT include the project root ("main") as a worktree

#### Scenario: Combine filter with context-aware completion

- **WHEN** completion is requested from project context with existing-worktree-only option
- **THEN** system SHALL return only existing worktrees for the current project
- **AND** suggestions SHALL be branch names without project prefix

#### Scenario: Combine filter with cross-project completion

- **WHEN** completion is requested with partial `<project>/` and existing-worktree-only option
- **THEN** system SHALL return only existing worktrees for the specified project
- **AND** suggestions SHALL be formatted as `<project>/<branch>`
