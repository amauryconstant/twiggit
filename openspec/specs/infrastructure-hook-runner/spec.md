# Capability: Hook Runner

## Purpose

Execute post-create hook commands from `.twiggit.toml` after a
successful worktree creation, with env-var injection, per-command
timeout, and failure collection.

## Requirements

### Requirement: Read hook config

The system SHALL read `.twiggit.toml` at the repository root and
extract `[hooks.post-create].commands`. Missing config file or empty
command list SHALL result in a no-op (`HookResult.Executed = false`).



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Command execution

The system SHALL execute each command in the new worktree directory
(also `cwd`) using `os/exec`. The runner SHALL use
`Config.Shell.HookTimeout` seconds as a per-command timeout.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Environment variables

The system SHALL set the following env vars for each hook command:

| Variable | Value |
|---|---|
| `TWIGGIT_WORKTREE_PATH` | Absolute path to the new worktree |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | New branch name |
| `TWIGGIT_SOURCE_BRANCH` | Branch created from |
| `TWIGGIT_MAIN_REPO_PATH` | Path to the main repository |

#### Scenario: Env vars present

- **WHEN** a hook command is invoked
- **THEN** `os.Getenv("TWIGGIT_WORKTREE_PATH")` SHALL return the
  new worktree's absolute path
- **AND** `os.Getenv("TWIGGIT_BRANCH_NAME")` SHALL return the new
  branch name

### Requirement: Failure collection

The system SHALL continue running remaining commands when one fails
(non-zero exit or timeout). Each failure SHALL be captured in
`HookResult.Failures` with `{Command, ExitCode, Output}`. The runner
SHALL NOT roll back the worktree.

#### Scenario: One failure, run rest

- **WHEN** command 1 exits non-zero but command 2 succeeds
- **THEN** command 2 SHALL still run
- **AND** `HookResult.Failures[0]` SHALL record command 1's exit code
- **AND** `HookResult.Success` SHALL be `false`

### Requirement: Timeout per command

When a command exceeds `HookTimeout` seconds, the system SHALL kill
the process, record a `HookFailure` with the timeout flag, and
continue with the next command.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Compile-time interface check

The runner SHALL satisfy `application.HookRunner` via
`var _ application.HookRunner = (*HookRunnerImpl)(nil)`.



#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
### Requirement: Service contract

The `Run(ctx, *application.HookRunRequest)` method SHALL execute
hooks of the requested `HookType` with the supplied context and
return `*domain.HookResult`. The cmd-layer display contract lives
in `cli-worktree-hooks`; this spec is silent on display.


#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract
