## Purpose

Enhanced descriptions, status indicators, and smart sorting for completion suggestions to help users quickly identify relevant targets.

## Requirements

### Requirement: Smart Sorting of Suggestions

The system SHALL sort completion suggestions using a priority-based algorithm that surfaces the most relevant targets first.

#### Scenario: Current worktree appears first

- **WHEN** user requests completion from a worktree context
- **AND** the current worktree's branch is among the suggestions
- **THEN** current worktree branch SHALL appear first in the suggestion list

#### Scenario: Default branch appears second

- **WHEN** user requests completion from any context
- **AND** the repository has a default branch (main, master, or configured default)
- **THEN** default branch SHALL appear immediately after current worktree (if applicable)
- **AND** default branch SHALL appear first if not in worktree context

#### Scenario: Remaining branches sorted alphabetically

- **WHEN** current worktree and default branch have been positioned
- **THEN** remaining suggestions SHALL be sorted alphabetically

#### Scenario: Sorting applies to all suggestion types

- **WHEN** suggestions include both projects and branches
- **THEN** sorting SHALL apply within each suggestion type separately
- **AND** projects and branches SHALL NOT be intermixed in sorting

### Requirement: Enhanced Branch Descriptions

The system SHALL provide enriched descriptions for branch suggestions including remote tracking information and activity context.

#### Scenario: Branch with remote tracking info

- **WHEN** completion suggests a branch with a remote tracking branch
- **THEN** description SHALL include remote name (e.g., "origin/feature-1")

#### Scenario: Branch without remote

- **WHEN** completion suggests a local-only branch
- **THEN** description SHALL indicate "local only" or equivalent

#### Scenario: Worktree description with path hint

- **WHEN** completion suggests a branch that has a worktree
- **THEN** description SHALL indicate worktree status (e.g., "Worktree • 2 commits ahead")

#### Scenario: Unmaterialized branch description

- **WHEN** completion suggests a branch without a worktree
- **THEN** description SHALL indicate "create worktree" or equivalent

### Requirement: Status Indicator for Current Worktree

The system SHALL display a visual status indicator when the current worktree has uncommitted changes.

#### Scenario: Dirty worktree indicator

- **WHEN** user requests completion from a worktree with uncommitted changes
- **AND** the current worktree's branch appears in suggestions
- **THEN** suggestion SHALL include a dirty status indicator

#### Scenario: Clean worktree no indicator

- **WHEN** user requests completion from a worktree with clean working directory
- **THEN** no status indicator SHALL be shown

#### Scenario: Status indicator limited to current worktree

- **WHEN** suggestions include multiple worktree branches
- **THEN** status indicator SHALL only be shown for the CURRENT worktree
- **AND** other worktree branches SHALL NOT have status indicators
- **AND** this limitation exists to avoid N git status calls

#### Scenario: Status check timeout

- **WHEN** git status operation exceeds 500ms timeout
- **THEN** system SHALL return suggestions without status indicator
- **AND** completion SHALL not fail due to timeout

### Requirement: Project Descriptions

The system SHALL provide meaningful descriptions for project suggestions.

#### Scenario: Project description from outside git

- **WHEN** completion suggests a project from outside git context
- **THEN** description SHALL be "Project directory"

#### Scenario: Project description from project context

- **WHEN** completion suggests another project from project/worktree context
- **THEN** description SHALL indicate cross-project navigation (e.g., "Navigate to project")
