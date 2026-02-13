## ADDED Requirements

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
