# Capability: Worktree Pruning

## Purpose

The worktree pruning capability provides functionality for identifying, listing, and deleting merged worktrees with safety checks, protected branch handling, and bulk operation support.

## ADDED Requirements

### Requirement: List Merged Worktrees for Pruning

The system SHALL identify and list merged worktrees for pruning, displaying branch names, paths, and merge status.

#### Scenario: List merged worktrees in current project

- **WHEN** user runs `twiggit prune --dry-run` from within a project directory or worktree
- **THEN** system SHALL list all merged worktrees for the current project
- **AND** main worktree SHALL be excluded from the list
- **AND** protected branches SHALL be excluded from the list
- **AND** output SHALL show branch name, path, and merge status
- **AND** "No merged worktrees found" message SHALL be displayed if no merged worktrees exist

#### Scenario: List merged worktrees across all projects

- **WHEN** user runs `twiggit prune --all --dry-run`
- **THEN** system SHALL list merged worktrees from all projects
- **AND** main worktrees SHALL be excluded
- **AND** protected branches SHALL be excluded
- **AND** output SHALL show project context (branch name, path, status) for each worktree

#### Scenario: List merged worktrees from outside git context

- **WHEN** user runs `twiggit prune --dry-run` from outside any git context without --all flag
- **THEN** system SHALL return error indicating project name is required
- **AND** system SHALL suggest using --all flag or running from within a project context

#### Scenario: Display protected branch warning

- **WHEN** merged worktree is on a protected branch (main, master, develop, staging, production by default)
- **THEN** system SHALL not include the worktree in the list
- **AND** system SHALL display message "Skipping protected branch: <branch>"

---

### Requirement: Delete Merged Worktrees

The system SHALL delete merged worktrees with optional branch deletion, protected branch safety checks, and user confirmation for bulk operations.

#### Scenario: Delete single merged worktree

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** the worktree branch is merged into the base branch
- **AND** worktree has no uncommitted changes
- **AND** branch is not protected
- **THEN** system SHALL remove the worktree directory
- **AND** system SHALL NOT delete the branch by default
- **AND** success message SHALL display the deleted worktree path
- **AND** current directory SHALL be maintained after deletion

#### Scenario: Delete merged worktree with branch deletion

- **WHEN** user runs `twiggit prune --delete-branches myproject/feature-branch`
- **AND** the worktree branch is merged
- **AND** branch is not protected
- **THEN** system SHALL remove the worktree directory
- **AND** system SHALL delete the corresponding git branch
- **AND** success message SHALL indicate both worktree and branch were deleted

#### Scenario: Abort pruning of unmerged worktree

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** the specified branch is not merged into the base branch
- **THEN** system SHALL display message "Skipping unmerged worktree: <branch>"
- **AND** worktree SHALL NOT be deleted

#### Scenario: Delete merged worktree with uncommitted changes using --force

- **WHEN** user runs `twiggit prune --force myproject/feature-branch`
- **AND** the worktree branch is merged
- **AND** worktree has uncommitted changes
- **AND** branch is not protected
- **THEN** system SHALL bypass uncommitted changes safety check
- **AND** worktree directory SHALL be removed
- **AND** branch deletion SHALL follow --delete-branches flag behavior
- **AND** success message SHALL display the deleted worktree path

#### Scenario: Protect protected branches from deletion

- **WHEN** user runs `twiggit prune --delete-branches myproject/main`
- **AND** main is a protected branch (default: main, master, develop, staging, production)
- **THEN** system SHALL return error
- **AND** error message SHALL indicate "Cannot delete protected branch: main"
- **AND** worktree and branch SHALL NOT be deleted

#### Scenario: Bulk delete merged worktrees with confirmation

- **WHEN** user runs `twiggit prune --all --delete-branches`
- **AND** merged worktrees are found across projects
- **THEN** system SHALL display list of worktrees to be deleted
- **AND** system SHALL prompt user for confirmation
- **AND** system SHALL ask "Delete X merged worktrees? (y/n)"
- **AND** upon confirmation, system SHALL delete all listed worktrees
- **AND** upon cancellation, system SHALL abort without deletion

#### Scenario: Skip user confirmation with --force flag in bulk mode

- **WHEN** user runs `twiggit prune --all --delete-branches --force`
- **AND** merged worktrees are found across projects
- **THEN** system SHALL NOT prompt for confirmation
- **AND** system SHALL immediately delete all merged worktrees
- **AND** success message SHALL indicate number of worktrees deleted

#### Scenario: Navigate to project directory after single-worktree prune

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** deletion succeeds
- **THEN** system SHALL output path to change to (project main directory)
- **AND** path SHALL be printed to stdout for shell wrapper consumption
- **AND** success message SHALL still be displayed to stdout/stderr

#### Scenario: Do not change directory after bulk prune

- **WHEN** user runs `twiggit prune --all --delete-branches`
- **THEN** system SHALL NOT output path for shell wrapper
- **AND** current directory SHALL be maintained

---

### Requirement: Validate Prune Operation Constraints

The system SHALL validate prune operation constraints including worktree status, merge status, and protected branch configuration.

#### Scenario: Validate worktree exists before pruning

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** worktree does not exist
- **THEN** system SHALL return error indicating worktree not found
- **AND** error SHALL include the worktree path
- **AND** pruning SHALL not be attempted

#### Scenario: Validate merged status before pruning

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** the specified branch is not merged into the base branch
- **THEN** system SHALL skip that worktree
- **AND** system SHALL display message "Skipping unmerged worktree: <branch>"
- **AND** no error SHALL be returned

#### Scenario: Validate branch is not protected before deletion

- **WHEN** user runs `twiggit prune --delete-branches myproject/develop`
- **AND** develop is in the protected branches list
- **THEN** system SHALL return error
- **AND** error message SHALL list all protected branches
- **AND** branch SHALL NOT be deleted

#### Scenario: Validate worktree is not the current active worktree

- **WHEN** user runs `twiggit prune` from within the worktree being pruned
- **THEN** system SHALL return error indicating cannot prune current worktree
- **AND** error SHALL suggest changing directory first
- **AND** worktree SHALL not be deleted

#### Scenario: Detect uncommitted changes in worktree

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** worktree has uncommitted changes
- **AND** --force flag is not provided
- **THEN** system SHALL return error indicating worktree has uncommitted changes
- **AND** system SHALL suggest using --force to override safety check
- **AND** worktree SHALL not be deleted

---

### Requirement: Configure Protected Branches

The system SHALL allow configuration of protected branches to prevent accidental deletion of critical branches.

#### Scenario: Load default protected branches

- **WHEN** system initializes protected branch configuration
- **THEN** system SHALL load default protected branches: main, master, develop, staging, production
- **AND** default branches SHALL be used if configuration file does not specify protected branches

#### Scenario: Load custom protected branches from config

- **WHEN** user configures custom protected branches in twiggit config
- **THEN** system SHALL load protected branches from configuration file
- **AND** configured branches SHALL replace default protected branches
- **AND** protected branch check SHALL use configured list

#### Scenario: Validate protected branch configuration

- **WHEN** user provides invalid protected branch configuration (non-existent branches)
- **THEN** system SHALL return validation error
- **AND** error SHALL indicate which branches are invalid
- **AND** system SHALL not apply invalid configuration

---

### Requirement: Handle Prune Operation Errors

The system SHALL handle prune operation errors gracefully with clear error messages and appropriate error handling.

#### Scenario: Return error if prune operation fails

- **WHEN** worktree directory removal fails during prune
- **THEN** system SHALL return error indicating prune failed
- **AND** error SHALL include the worktree path and failure reason
- **AND** system SHALL attempt to continue with other worktrees in bulk mode
- **AND** summary SHALL indicate which worktrees failed to delete

#### Scenario: Handle branch deletion failure

- **WHEN** branch deletion fails during prune with --delete-branches
- **THEN** system SHALL return error indicating branch deletion failed
- **AND** error SHALL include branch name and failure reason
- **AND** worktree directory SHALL remain deleted (rollback not attempted)
- **AND** user SHALL be notified of partial success

#### Scenario: Provide summary of prune operation

- **WHEN** prune operation completes (successful or partial)
- **THEN** system SHALL display summary of operation
- **AND** summary SHALL include number of worktrees deleted
- **AND** summary SHALL include number of worktrees skipped
- **AND** summary SHALL include number of branches deleted (if applicable)
- **AND** summary SHALL include any errors encountered

#### Scenario: Return error for empty repository path

- **WHEN** pruning with an empty repository path
- **THEN** system SHALL return error indicating repository path cannot be empty
- **AND** pruning SHALL not be attempted
