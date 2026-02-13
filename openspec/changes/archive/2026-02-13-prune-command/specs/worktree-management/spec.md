## ADDED Requirements

### Requirement: Prune Merged Worktrees

The system SHALL provide a prune command for context-aware bulk deletion of merged worktrees with protected branch safety and optional branch deletion.

#### Scenario: Prune merged worktrees in current project

- **WHEN** user runs `twiggit prune` from within a project directory or worktree
- **AND** merged worktrees exist for the current project
- **THEN** system SHALL delete all merged worktrees in the current project
- **AND** main worktree SHALL be excluded from deletion
- **AND** protected branches SHALL be excluded from deletion
- **AND** branches SHALL NOT be deleted by default
- **AND** summary SHALL display number of worktrees deleted

#### Scenario: Prune merged worktrees with branch deletion

- **WHEN** user runs `twiggit prune --delete-branches` from within a project directory
- **AND** merged worktrees exist
- **AND** branches are not protected
- **THEN** system SHALL delete all merged worktrees
- **AND** system SHALL delete corresponding git branches
- **AND** summary SHALL display number of worktrees and branches deleted

#### Scenario: Prune merged worktrees across all projects

- **WHEN** user runs `twiggit prune --all`
- **AND** merged worktrees exist across multiple projects
- **THEN** system SHALL list merged worktrees to be deleted across all projects
- **AND** system SHALL prompt user for confirmation
- **AND** upon confirmation, system SHALL delete all merged worktrees
- **AND** summary SHALL display number of worktrees deleted across all projects

#### Scenario: Dry run prune operation

- **WHEN** user runs `twiggit prune --dry-run`
- **THEN** system SHALL display list of worktrees that would be deleted
- **AND** no worktrees SHALL be deleted
- **AND** system SHALL display number of worktrees to be pruned

#### Scenario: Prune with force flag

- **WHEN** user runs `twiggit prune --force`
- **AND** some worktrees have uncommitted changes
- **THEN** system SHALL bypass uncommitted changes safety check
- **AND** system SHALL delete worktrees regardless of uncommitted changes
- **AND** summary SHALL indicate worktrees with uncommitted changes were force deleted

#### Scenario: Protect protected branches from pruning

- **WHEN** user runs `twiggit prune --delete-branches`
- **AND** merged worktree is on a protected branch (main, master, develop, staging, production)
- **THEN** system SHALL skip the worktree
- **AND** system SHALL display "Skipping protected branch: <branch>"
- **AND** system SHALL return error if all worktrees are protected

#### Scenario: Context-aware project inference

- **WHEN** user runs `twiggit prune` from within a project worktree
- **THEN** system SHALL infer project from current worktree
- **AND** system SHALL only prune worktrees for that project
- **AND** system SHALL not affect other projects

#### Scenario: Change directory after prune
- **WHEN** user runs `twiggit prune myproject/feature-branch` (single worktree)
- **AND** deletion succeeds
- **THEN** system SHALL output project main directory path
- **AND** path SHALL be printed to stdout for shell wrapper consumption
- **AND** user SHALL be navigated to project directory

#### Scenario: Do not change directory after bulk prune
- **WHEN** user runs `twiggit prune --all` (bulk operation)
- **THEN** system SHALL NOT output path for shell wrapper
- **AND** current directory SHALL be maintained
