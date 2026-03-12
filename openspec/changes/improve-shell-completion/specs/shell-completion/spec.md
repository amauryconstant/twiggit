## Purpose

Shell completion support for twiggit CLI via Carapace, providing context-aware tab completion for commands, arguments, and flags across multiple shell environments.

## MODIFIED Requirements

### Requirement: Command Argument Completion

The system SHALL provide context-aware tab completion for command arguments via Carapace's `PositionalCompletion`.

#### Scenario: Complete cd command from project context

- **WHEN** user presses tab after `twiggit cd ` from within a project directory
- **THEN** system SHALL suggest branch names for current project
- **AND** system SHALL include "main" as a suggestion
- **AND** suggestions SHALL include descriptions (e.g., "Project root directory", "Worktree for branch X")
- **AND** system SHALL suggest other project names for cross-project navigation

#### Scenario: Complete cd command from worktree context

- **WHEN** user presses tab after `twiggit cd ` from within a worktree
- **THEN** system SHALL suggest sibling worktree branch names
- **AND** system SHALL include "main" as a suggestion
- **AND** system SHALL suggest other project names for cross-project navigation

#### Scenario: Complete cd command from outside git context

- **WHEN** user presses tab after `twiggit cd ` from outside any git repository
- **THEN** system SHALL suggest project names from configured projects directory

#### Scenario: Complete create command argument

- **WHEN** user presses tab after `twiggit create ` from within a project directory
- **THEN** system SHALL suggest branch names (indicating worktree creation)
- **AND** system SHALL NOT include branches that already have worktrees (unless showing all)

#### Scenario: Complete delete command argument

- **WHEN** user presses tab after `twiggit delete ` from within a project directory
- **THEN** system SHALL suggest ONLY existing worktrees
- **AND** system SHALL NOT suggest branches without materialized worktrees

### Requirement: Progressive Project Completion

The system SHALL support progressive completion for project/branch syntax by automatically appending "/" when a project is selected.

#### Scenario: Project suggestion includes slash suffix

- **WHEN** completion suggests a project name
- **THEN** suggestion SHALL include "/" as a suffix
- **AND** accepting the suggestion SHALL result in input like "projectname/"

#### Scenario: Slash suffix triggers branch completion

- **WHEN** user accepts a project suggestion with "/" suffix
- **AND** presses tab again
- **THEN** system SHALL provide branch suggestions for that project

#### Scenario: Branch suggestions have no suffix

- **WHEN** completion suggests a branch name
- **THEN** suggestion SHALL NOT include "/" suffix
- **AND** accepting the suggestion SHALL complete the input without additional characters
