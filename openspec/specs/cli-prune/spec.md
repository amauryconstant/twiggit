# Capability: Prune Worktrees

## Purpose

Bulk-deletion of merged worktrees for post-merge cleanup, with optional
branch deletion, protected-branch safety, dry-run preview, and bulk
confirmation. Single-worktree prune emits a navigation path for the shell
wrapper to consume.

## Requirements

### Requirement: Prune merged worktrees in current project

The system SHALL delete all merged worktrees in the current project while
excluding the main worktree and any branch in `ValidationConfig.ProtectedBranches`.

#### Scenario: Prune in current project

- **WHEN** user runs `twiggit prune` from within a project directory or worktree
- **AND** merged worktrees exist for the current project
- **THEN** system SHALL delete all merged worktrees
- **AND** main worktree SHALL be excluded
- **AND** protected branches SHALL be excluded
- **AND** branches SHALL NOT be deleted (use `--delete-branches` for that)
- **AND** summary SHALL show deleted count and skipped count

#### Scenario: No merged worktrees

- **WHEN** user runs `twiggit prune` and no merged worktrees exist
- **THEN** system SHALL display "No merged worktrees found"
- **AND** exit cleanly with status 0

### Requirement: Delete branches with `--delete-branches`

The system SHALL delete the corresponding git branch after removing a
merged worktree when `--delete-branches` is set, subject to the protected
branches list.

#### Scenario: Prune with branch deletion

- **WHEN** user runs `twiggit prune --delete-branches` from within a project
- **AND** merged worktrees exist on non-protected branches
- **THEN** system SHALL delete worktrees
- **AND** system SHALL delete corresponding branches
- **AND** summary SHALL include branch deletion count

#### Scenario: Skip protected branch

- **WHEN** user runs `twiggit prune --delete-branches myproject/main`
- **AND** `main` is in `ValidationConfig.ProtectedBranches` (default:
  `main`, `master`, `develop`, `staging`, `production`)
- **THEN** system SHALL skip that worktree
- **AND** SHALL emit "Skipping protected branch: main"
- **THEN** SHALL return error if all targets are protected

### Requirement: Bulk prune across projects with confirmation

The system SHALL prune merged worktrees across all projects with preview,
confirmation, and short-circuit flags.

#### Scenario: Preview all projects

- **WHEN** user runs `twiggit prune --all --dry-run`
- **THEN** system SHALL list merged worktrees from all projects
- **AND** main worktrees and protected branches SHALL be excluded
- **AND** no deletion SHALL occur

#### Scenario: Bulk prune with confirmation

- **WHEN** user runs `twiggit prune --all`
- **AND** `--force`, `--yes`, and `--dry-run` are NOT set
- **THEN** system SHALL run a dry-run preview to stderr
- **AND** system SHALL prompt "Continue? (y/n)" on stdin
- **AND** on `y`/`yes`, system SHALL proceed with the real prune
- **AND** on any other input or EOF, system SHALL abort without deleting

#### Scenario: Bulk prune with `--yes`

- **WHEN** user runs `twiggit prune --all --yes` (alias `-y`)
- **THEN** system SHALL skip the confirmation prompt
- **AND** system SHALL still apply uncommitted-changes safety checks
- **AND** proceed with the prune

#### Scenario: Bulk prune with `--force`

- **WHEN** user runs `twiggit prune --all --force` (alias `-f`)
- **THEN** system SHALL skip the confirmation prompt
- **AND** system SHALL bypass uncommitted-changes safety checks
- **AND** proceed with the prune

### Requirement: Short flags

The system SHALL expose the following short flags alongside their long
forms: `-f` (`--force`), `-y` (`--yes`), `-d` (`--delete-branches`),
`-a` (`--all`), `-n` (`--dry-run`).

#### Scenario: Short flags accepted

- **WHEN** user runs `twiggit prune -f -y -d -a -n`
- **THEN** each short flag SHALL be honored identically to its long form

### Requirement: Single-worktree prune navigation

The system SHALL output the project main directory path to stdout after a
successful single-worktree prune, so the shell wrapper can navigate the user.

#### Scenario: Navigate after single-worktree prune

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** deletion succeeds
- **THEN** system SHALL print the project main directory path to stdout
- **AND** success and summary messages SHALL go to stderr

#### Scenario: No navigation after bulk prune

- **WHEN** user runs `twiggit prune --all`
- **THEN** system SHALL NOT print a path to stdout
- **AND** the user's current directory SHALL be preserved

### Requirement: Unmerged worktree skip

The system SHALL skip worktrees whose branch is not merged into the base
branch and report them, rather than erroring.

#### Scenario: Skip unmerged target

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** the branch is not merged into the base branch
- **THEN** system SHALL emit "Skipping unmerged worktree: feature-branch"
- **AND** SHALL NOT delete the worktree
- **AND** SHALL exit cleanly

### Requirement: Progress reporting in bulk mode

The system SHALL report progress to stderr during bulk prune operations
(`--all` or no specific target), and SHALL suppress it under
`--quiet`/`-q`. See `cli-quiet-mode`.

#### Scenario: Bulk progress

- **WHEN** user runs `twiggit prune --all` without `--quiet`
- **THEN** "Pruning merged worktrees..." SHALL appear on stderr before
  work begins
- **AND** "Prune complete" SHALL appear on stderr when finished

#### Scenario: Bulk progress suppressed

- **WHEN** user runs `twiggit prune --all --quiet`
- **THEN** progress messages SHALL NOT appear on stderr
- **AND** summary output SHALL also be suppressed

### Requirement: Uncommitted changes safety

The system SHALL refuse to delete a worktree with uncommitted changes
unless `--force` is set, and SHALL suggest `--force` in the error message.

#### Scenario: Block prune with dirty worktree

- **WHEN** user runs `twiggit prune myproject/feature-branch`
- **AND** the worktree has uncommitted changes
- **AND** `--force` is NOT set
- **THEN** system SHALL return an error naming `--force` as the override
- **AND** SHALL NOT delete the worktree

### Requirement: Refuse to prune the current worktree

The system SHALL refuse to prune the worktree the user is currently
inside, with an actionable hint.

#### Scenario: Self-prune blocked

- **WHEN** user runs `twiggit prune` from inside the worktree being pruned
- **THEN** system SHALL return an error explaining the constraint
- **AND** SHALL suggest changing directory first

### Requirement: Configurable protected branches

The system SHALL load `ValidationConfig.ProtectedBranches` from
configuration (default: `main`, `master`, `develop`, `staging`,
`production`) and SHALL use this list for prune safety checks.
See `domain-config-types`.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
