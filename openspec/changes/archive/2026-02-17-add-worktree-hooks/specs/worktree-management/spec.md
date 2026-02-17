## MODIFIED Requirements

### Requirement: Create Worktree

The system SHALL create a new git worktree from a specified source branch, with validation for duplicate worktrees, branch name validation, context-aware project inference, and optional post-create hook execution.

#### Scenario: Create worktree with project/branch specification

- **WHEN** user runs `twiggit create myproject/feature-branch`
- **AND** source branch exists (defaults to `main` if not specified)
- **THEN** system SHALL create worktree at `~/Worktrees/myproject/feature-branch/`
- **AND** worktree SHALL be created from specified source branch
- **AND** success message SHALL display worktree path and branch name
- **AND** parent directories SHALL be created automatically if they don't exist
- **AND** system SHALL return `CreateWorktreeResult` containing worktree info and hook result

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

#### Scenario: Execute post-create hooks after successful creation

- **WHEN** worktree creation succeeds
- **AND** `.twiggit.toml` exists with `[hooks.post-create]` configuration
- **THEN** system SHALL execute configured hook commands
- **AND** `CreateWorktreeResult.HookResult` SHALL contain execution results
- **AND** hook failures SHALL be reported in `HookResult.Failures`

#### Scenario: Return hook result even when no hooks configured

- **WHEN** worktree creation succeeds
- **AND** no `.twiggit.toml` exists or no hooks configured
- **THEN** `CreateWorktreeResult.HookResult` SHALL be nil
- **AND** worktree info SHALL still be returned in `CreateWorktreeResult.Worktree`

#### Scenario: Display hook warnings after creation

- **WHEN** worktree creation succeeds
- **AND** post-create hooks ran with failures
- **THEN** system SHALL display worktree creation success message
- **AND** system SHALL display warning section with hook failure details
- **AND** warning SHALL indicate worktree is ready but setup may be incomplete

#### Scenario: Change directory after creation with -C flag

- **WHEN** user runs `twiggit create -C myproject/feature-branch`
- **AND** worktree creation succeeds
- **THEN** system SHALL output path to change to (the newly created worktree directory)
- **AND** path SHALL be printed to stdout for shell wrapper consumption
- **AND** success message SHALL still be displayed to stdout/stderr
