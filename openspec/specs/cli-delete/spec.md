# Capability: Delete Worktree

## Purpose

Remove a worktree and (by default) its branch, with safety checks for
uncommitted changes, the current worktree, and the `--merged-only`
constraint. Alias: `rm`. Use `--keep-branch` to preserve the branch.

## Requirements

### Requirement: Delete with safety checks

The system SHALL delete a worktree and its branch by default, with
safety checks for uncommitted changes and the current worktree.

#### Scenario: Standard delete

- **WHEN** user runs `twiggit delete myproject/feature`
- **AND** worktree exists with no uncommitted changes
- **AND** worktree is not the user's current directory
- **THEN** system SHALL remove the worktree directory
- **AND** SHALL delete the corresponding git branch
- **AND** success message SHALL display the deleted worktree path

#### Scenario: Already removed is idempotent

- **WHEN** user runs `twiggit delete myproject/feature` and the worktree
  no longer exists
- **THEN** system SHALL emit "Deleted worktree: <path> (already removed)"
- **AND** SHALL exit with status 0

### Requirement: Refuse dirty worktree without `--force`

The system SHALL refuse to delete a worktree with uncommitted changes
unless `--force`/`-f` is set.

#### Scenario: Dirty, no force

- **WHEN** user runs `twiggit delete myproject/feature`
- **AND** worktree has uncommitted changes
- **AND** `--force` is NOT passed
- **THEN** system SHALL return error naming `--force` as the override
- **AND** SHALL NOT delete

#### Scenario: Force delete

- **WHEN** user runs `twiggit delete --force myproject/feature`
- **AND** worktree has uncommitted changes
- **THEN** system SHALL bypass the safety check
- **AND** SHALL remove the worktree and branch
- **AND** success message SHALL display the deleted path

### Requirement: `--keep-branch`

The system SHALL accept `--keep-branch` to remove the worktree without
deleting its branch.

#### Scenario: Keep branch

- **WHEN** user runs `twiggit delete --keep-branch myproject/feature`
- **THEN** system SHALL remove the worktree directory
- **AND** SHALL NOT delete the corresponding branch
- **AND** success message SHALL indicate the branch was preserved

### Requirement: `--merged-only`

The system SHALL accept `-m`/`--merged-only` to require the branch to
be merged into the base branch before deletion.

#### Scenario: Not merged, --merged-only set

- **WHEN** user runs `twiggit delete --merged-only myproject/feature`
- **AND** the branch is not merged into the base branch
- **THEN** system SHALL return error explaining the `--merged-only` constraint
- **AND** SHALL NOT delete

#### Scenario: Merged, --merged-only set

- **WHEN** user runs `twiggit delete --merged-only myproject/feature`
- **AND** the branch is merged
- **THEN** system SHALL delete the worktree and branch

### Requirement: Refuse to delete current worktree

The system SHALL refuse to delete the worktree the user is currently
inside.

#### Scenario: Inside target worktree

- **WHEN** user runs `twiggit delete myproject/feature` from inside
  `myproject/feature`
- **THEN** system SHALL return error naming the constraint
- **AND** SHALL suggest changing directory first

### Requirement: `-C` / `--cd` navigation

When `-C`/`--cd` is passed, the system SHALL print the project main
directory path to stdout on success, so the shell wrapper can navigate.

#### Scenario: `-C` from current worktree

- **WHEN** user runs `twiggit delete -C myproject/feature` from inside
  that worktree
- **AND** deletion succeeds
- **THEN** system SHALL print the project main directory path to stdout

#### Scenario: `-C` from elsewhere

- **WHEN** user runs `twiggit delete -C myproject/feature` from outside
  the target worktree
- **THEN** system SHALL NOT print a navigation path to stdout
- **AND** success message SHALL still go to stderr

### Requirement: Branch deletion routed through git client

The system SHALL use the git client's `DeleteBranch` operation (CLI-backed
per `infrastructure-git-client` routing) when removing a branch, so that
branches currently checked out by other worktrees are handled correctly
by `git branch -d`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: `rm` alias

The system SHALL accept `twiggit rm myproject/feature` as an alias for
`twiggit delete myproject/feature`. See `cli-command-options-pattern`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
