## Purpose

Define hook execution behavior for post-worktree-creation operations, allowing users to run custom setup commands automatically after creating a new worktree.

## Requirements

### Requirement: Post-Create Hook Execution

The system SHALL execute user-defined commands after successful worktree creation when a `.twiggit.toml` configuration file exists at the repository root with a `hooks.post-create` section.

#### Scenario: Execute post-create hooks after worktree creation

- **WHEN** user creates a worktree with `twiggit create myproject/feature-branch`
- **AND** `.twiggit.toml` exists at the repository root with `[hooks.post-create]` section
- **THEN** system SHALL execute each command in `commands` array sequentially
- **AND** commands SHALL execute in the new worktree directory
- **AND** environment variables SHALL be set with execution context

#### Scenario: Skip hooks when no configuration file exists

- **WHEN** user creates a worktree
- **AND** no `.twiggit.toml` file exists at the repository root
- **THEN** system SHALL skip hook execution
- **AND** worktree creation SHALL complete normally
- **AND** no error or warning SHALL be displayed

#### Scenario: Skip hooks when hooks section is empty

- **WHEN** user creates a worktree
- **AND** `.twiggit.toml` exists but has no `[hooks.post-create]` section
- **THEN** system SHALL skip hook execution
- **AND** worktree creation SHALL complete normally

---

### Requirement: Hook Execution Environment

The system SHALL provide context about the worktree operation via environment variables during hook command execution.

#### Scenario: Set environment variables for hook execution

- **WHEN** system executes post-create hook commands
- **THEN** following environment variables SHALL be set:
  - `TWIGGIT_WORKTREE_PATH`: Absolute path to the new worktree
  - `TWIGGIT_PROJECT_NAME`: Project identifier
  - `TWIGGIT_BRANCH_NAME`: Name of the new branch
  - `TWIGGIT_SOURCE_BRANCH`: Branch the worktree was created from
  - `TWIGGIT_MAIN_REPO_PATH`: Absolute path to the main repository

#### Scenario: Execute commands in worktree directory

- **WHEN** system executes post-create hook commands
- **THEN** current working directory SHALL be the new worktree path
- **AND** commands SHALL have access to files in the worktree

---

### Requirement: Hook Failure Handling

The system SHALL report hook command failures as warnings without rolling back the worktree creation.

#### Scenario: Continue after hook command failure

- **WHEN** a post-create hook command exits with non-zero status
- **THEN** system SHALL continue executing remaining commands
- **AND** system SHALL collect all failures
- **AND** worktree SHALL remain created

#### Scenario: Display hook failure warnings

- **WHEN** one or more post-create hook commands fail
- **THEN** system SHALL display warning message after worktree creation success
- **AND** warning SHALL include failed command, exit code, and output
- **AND** warning SHALL indicate worktree is ready but setup may be incomplete

#### Scenario: Report all failures in result

- **WHEN** post-create hooks complete with any failures
- **THEN** result SHALL include `HookResult` with `Success: false`
- **AND** `HookResult.Failures` SHALL contain each failed command's details
- **AND** each failure SHALL include command string, exit code, and combined output

#### Scenario: Report success when all hooks pass

- **WHEN** all post-create hook commands exit with zero status
- **THEN** result SHALL include `HookResult` with `Success: true`
- **AND** `HookResult.Failures` SHALL be empty
- **AND** no warning SHALL be displayed

---

### Requirement: Hook Configuration Format

The system SHALL parse hook configuration from `.twiggit.toml` at the repository root using TOML format.

#### Scenario: Parse valid hook configuration

- **WHEN** `.twiggit.toml` contains valid TOML with hooks section
- **THEN** system SHALL parse `[hooks.post-create]` section
- **AND** system SHALL read `commands` array as list of strings

#### Scenario: Handle malformed configuration file

- **WHEN** `.twiggit.toml` exists but contains invalid TOML syntax
- **THEN** system SHALL skip hook execution
- **AND** system SHALL log warning about parse failure
- **AND** worktree creation SHALL complete normally

#### Scenario: Handle missing commands array

- **WHEN** `[hooks.post-create]` section exists but `commands` field is missing or not an array
- **THEN** system SHALL treat as no commands configured
- **AND** worktree creation SHALL complete normally
