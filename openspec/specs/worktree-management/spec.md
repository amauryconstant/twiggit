## ADDED Requirements

### Requirement: List Worktrees

The system SHALL list git worktrees for the current project or all projects, displaying branch names, paths, and status indicators.

#### Scenario: List worktrees from project context

- **WHEN** user runs `twiggit list` from within a project directory or worktree
- **THEN** system SHALL list worktrees for the current project only
- **AND** main worktree SHALL be excluded from the list by default
- **AND** output SHALL show branch name, path, and status (modified/detached)
- **AND** "No worktrees found" message SHALL be displayed if no worktrees exist

#### Scenario: List all worktrees with --all flag

- **WHEN** user runs `twiggit list --all`
- **THEN** system SHALL list worktrees from all projects
- **AND** main worktrees SHALL be excluded from the list
- **AND** output SHALL show project context (branch name, path, status) for each worktree

#### Scenario: List worktrees from outside git context without --all

- **WHEN** user runs `twiggit list` from outside any git context without --all flag
- **THEN** system SHALL return error indicating project name is required
- **AND** system SHALL suggest using --all flag or running from within a project context

#### Scenario: Display worktree status indicators

- **WHEN** worktree has uncommitted changes
- **THEN** system SHALL display "(modified)" status indicator
- **AND** worktree path and branch name SHALL still be shown

- **WHEN** worktree HEAD is detached
- **THEN** system SHALL display "(detached)" status indicator
- **AND** worktree path and branch name SHALL still be shown

- **WHEN** worktree is clean and HEAD is attached
- **THEN** system SHALL display no status indicator
- **AND** only branch name and path SHALL be shown

---

### Requirement: Create Worktree

The system SHALL create a new git worktree from a specified source branch, with validation for duplicate worktrees, branch name validation, and context-aware project inference.

#### Scenario: Create worktree with project/branch specification

- **WHEN** user runs `twiggit create myproject/feature-branch`
- **AND** source branch exists (defaults to `main` if not specified)
- **THEN** system SHALL create worktree at `~/Worktrees/myproject/feature-branch/`
- **AND** worktree SHALL be created from specified source branch
- **AND** success message SHALL display worktree path and branch name
- **AND** parent directories SHALL be created automatically if they don't exist

#### Scenario: Infer project from context

- **WHEN** user runs `twiggit create feature-branch` from within a project directory or worktree
- **AND** project is inferred from current context
- **THEN** system SHALL create worktree for the inferred project
- **AND** worktree path SHALL use the inferred project name
- **AND** worktree SHALL be created from default source branch (main) unless --source is specified

#### Scenario: Validate branch name before creation

- **WHEN** user provides an invalid branch name (e.g., contains special characters, too long, empty)
- **THEN** system SHALL return validation error before attempting creation
- **AND** error message SHALL explain why the branch name is invalid
- **AND** system SHALL suggest valid branch name format

#### Scenario: Validate source branch exists

- **WHEN** user specifies a source branch with `--source` flag that does not exist
- **THEN** system SHALL return error indicating source branch does not exist
- **AND** worktree creation SHALL not be attempted
- **AND** error SHALL include the missing branch name

#### Scenario: Return error if outside git context without project specification

- **WHEN** user runs `twiggit create feature-branch` from outside any git context
- **AND** no project is specified in the argument
- **THEN** system SHALL return error "cannot infer project: not in a project context and no project specified"
- **AND** worktree creation SHALL not be attempted

#### Scenario: Create worktree with custom source branch

- **WHEN** user runs `twiggit create myproject/feature-branch --source develop`
- **AND** develop branch exists in the project
- **THEN** system SHALL create worktree from develop branch instead of main
- **AND** worktree path and success message SHALL reflect the feature branch name

---

### Requirement: Delete Worktree

The system SHALL delete a git worktree with safety checks for uncommitted changes, with options to preserve branch and force deletion.

#### Scenario: Delete worktree with safety checks

- **WHEN** user runs `twiggit delete myproject/feature-branch`
- **AND** worktree exists and has no uncommitted changes
- **AND** worktree is not the current active worktree
- **THEN** system SHALL remove the worktree directory
- **AND** system SHALL delete the corresponding git branch
- **AND** success message SHALL display the deleted worktree path
- **AND** current directory SHALL be maintained after deletion

#### Scenario: Abort deletion of worktree with uncommitted changes

- **WHEN** user runs `twiggit delete` on a worktree with uncommitted changes
- **AND** --force flag is not provided
- **THEN** system SHALL return error indicating worktree has uncommitted changes
- **AND** system SHALL suggest using --force to override safety checks
- **AND** worktree SHALL not be deleted

#### Scenario: Force delete worktree with uncommitted changes

- **WHEN** user runs `twiggit delete --force myproject/feature-branch`
- **AND** worktree has uncommitted changes
- **THEN** system SHALL bypass uncommitted changes safety check
- **AND** worktree directory SHALL be removed
- **AND** corresponding git branch SHALL be deleted
- **AND** success message SHALL display the deleted worktree path

#### Scenario: Handle worktree already removed

- **WHEN** user runs `twiggit delete` on a worktree that no longer exists
- **THEN** system SHALL display "Deleted worktree: <path> (already removed)" message
- **AND** command SHALL exit gracefully without error
- **AND** no further deletion attempts SHALL be made

#### Scenario: Delete worktree with --keep-branch flag

- **WHEN** user runs `twiggit delete --keep-branch myproject/feature-branch`
- **THEN** system SHALL remove the worktree directory
- **AND** system SHALL NOT delete the corresponding git branch
- **AND** success message SHALL indicate branch was preserved

#### Scenario: Validate merged-only constraint

- **WHEN** user runs `twiggit delete --merged-only myproject/feature-branch`
- **AND** the specified branch is not merged into the base branch
- **THEN** system SHALL return error indicating branch is not merged
- **AND** system SHALL explain that --merged-only requires the branch to be merged
- **AND** worktree SHALL not be deleted

#### Scenario: Delete worktree with --merged-only flag

- **WHEN** user runs `twiggit delete --merged-only myproject/feature-branch`
- **AND** the specified branch is merged into the base branch
- **THEN** system SHALL remove the worktree directory
- **AND** system SHALL delete the corresponding git branch
- **AND** success message SHALL display the deleted worktree path

#### Scenario: Change directory after deletion with -C flag

- **WHEN** user runs `twiggit delete -C myproject/feature-branch`
- **AND** deletion succeeds
- **THEN** system SHALL output to path to change to (e.g., project main directory)
- **AND** path SHALL be printed to stdout for shell wrapper consumption
- **AND** success message SHALL still be displayed to stdout/stderr

#### Scenario: Change directory after creation with -C flag

- **WHEN** user runs `twiggit create -C myproject/feature-branch`
- **AND** worktree creation succeeds
- **THEN** system SHALL output to path to change to (the newly created worktree directory)
- **AND** path SHALL be printed to stdout for shell wrapper consumption
- **AND** success message SHALL still be displayed to stdout/stderr

---

### Requirement: Prune Worktrees

The system SHALL prune stale worktree references from the git repository using the `git worktree prune` command.

#### Scenario: Prune worktrees successfully

- **WHEN** user runs command to prune worktrees (through any interface)
- **AND** repository is valid
- **THEN** system SHALL execute `git worktree prune`
- **AND** stale worktree references SHALL be removed
- **AND** success message SHALL be displayed if available
- **AND** no error SHALL be returned

#### Scenario: Return error for empty repository path

- **WHEN** pruning with an empty repository path
- **THEN** system SHALL return error indicating repository path cannot be empty
- **AND** pruning SHALL not be attempted

#### Scenario: Return error if prune fails

- **WHEN** git worktree prune command fails
- **THEN** system SHALL return error indicating prune failed
- **AND** error SHALL include stderr output from git command
- **AND** appropriate error details SHALL be provided

---

### Requirement: Check Branch Merged Status

The system SHALL check if a branch is merged into the current branch using `git branch --merged` command.

#### Scenario: Check if branch is merged

- **WHEN** system checks if a branch is merged
- **AND** branch exists in the repository
- **THEN** system SHALL execute `git branch --merged` command
- **AND** output SHALL be parsed for merged branches
- **AND** result SHALL indicate true if branch is found in merged list
- **AND** false SHALL be returned if branch is not in merged list

#### Scenario: Return error for empty repository path

- **WHEN** checking merged status with an empty repository path
- **THEN** system SHALL return error indicating repository path cannot be empty
- **AND** check SHALL not be attempted

#### Scenario: Return error for empty branch name

- **WHEN** checking merged status with an empty branch name
- **THEN** system SHALL return error indicating branch name cannot be empty
- **AND** check SHALL not be attempted

#### Scenario: Return error if git branch command fails

- **WHEN** git branch --merged command fails to execute
- **THEN** system SHALL return error indicating failed to check merged status
- **AND** error SHALL include stderr output from git command
- **AND** appropriate error details SHALL be provided

---

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

---

### Requirement: Handle Duplicate Worktree Creation

The system SHALL handle attempts to create a worktree that already exists, preventing duplicate worktree creation.

#### Scenario: Return error when worktree already exists

- **WHEN** user runs `twiggit create myproject/feature-branch`
- **AND** worktree at `~/Worktrees/myproject/feature-branch/` already exists
- **THEN** system SHALL return error indicating worktree already exists
- **AND** error SHALL include the existing worktree path
- **AND** worktree creation SHALL not be attempted

#### Scenario: Return error when branch exists but worktree doesn't

- **WHEN** user runs `twiggit create myproject/feature-branch`
- **AND** branch exists in repository
- **AND** worktree for that branch does not exist
- **THEN** system SHALL create the worktree
- **AND** worktree path SHALL be created at configured worktree location
- **AND** success message SHALL indicate worktree was created for existing branch
