# Capability: Worktree Hooks

## Purpose

Execute user-configured post-create hook commands after a successful
worktree creation, with environment-variable injection and failure
collection. Hook configuration lives in `.twiggit.toml` at the
repository root.

## Requirements

### Requirement: Post-create hook configuration

The system SHALL read `[hooks.post-create].commands` from
`.twiggit.toml` at the repository root and SHALL execute each command
after a successful worktree create.

#### Scenario: Hooks configured

- **WHEN** `.twiggit.toml` defines
  ```toml
  [hooks.post-create]
  commands = ["mise trust", "npm install"]
  ```
- **AND** user runs `twiggit create myproject/feature`
- **THEN** system SHALL execute `mise trust` and `npm install` in the
  new worktree after git operations complete

#### Scenario: No hooks configured

- **WHEN** `.twiggit.toml` does not exist or has no `post-create` block
- **THEN** system SHALL skip hook execution
- **AND** `HookResult.Executed` SHALL be `false`
- **AND** `HookResult.Failures` SHALL be empty

### Requirement: Environment variables

The system SHALL set the following environment variables in the hook
process:

| Variable | Value |
|---|---|
| `TWIGGIT_WORKTREE_PATH` | Absolute path to the new worktree (also cwd) |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | New branch name |
| `TWIGGIT_SOURCE_BRANCH` | Branch created from |
| `TWIGGIT_MAIN_REPO_PATH` | Path to the main repository |

#### Scenario: Env vars present

- **WHEN** a hook command runs
- **THEN** `os.Getenv("TWIGGIT_WORKTREE_PATH")` SHALL return the
  new worktree's absolute path

### Requirement: Failure collection

The system SHALL run all hook commands even if a previous one fails.
Failures SHALL be collected into `HookResult.Failures` and SHALL NOT
roll back the worktree.

#### Scenario: One command fails

- **WHEN** `mise trust` exits non-zero but `npm install` succeeds
- **THEN** system SHALL continue running `npm install`
- **AND** `HookResult.Success` SHALL be `false`
- **AND** `HookResult.Failures[0]` SHALL record the failed command,
  exit code, and output

#### Scenario: All commands fail

- **WHEN** every hook command exits non-zero
- **THEN** system SHALL still return success for the worktree itself
- **AND** `HookResult.Success` SHALL be `false`
- **AND** each failure SHALL be in `HookResult.Failures`

### Requirement: Cmd-layer display of hook warnings

The cmd layer SHALL display a warning section after `twiggit create`
when hooks failed, indicating the worktree is ready but setup may be
incomplete. The hook runner itself SHALL NOT print; the cmd layer
owns display.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
