# Capability: Hook Runner

## Purpose

Execute post-create hook commands from `.twiggit.toml` after a
successful worktree creation, with env-var injection, per-command
timeout, and failure collection.

## Requirements

### Requirement: Read hook config

The system SHALL read `.twiggit.toml` at the repository root and extract `[hooks.post-create].commands`. Missing config file or empty command list SHALL result in a no-op (`HookResult.Executed = false`). The runner lives in `internal/git/hook_runner.go` (migrated from `internal/infrastructure/hook_runner.go`) and returns `*core.HookResult` (the previous hook-result type under the `domain` package is renamed to `core.HookResult`).

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Command execution

The system SHALL execute each command in the new worktree directory (also `cwd`) using `os/exec`. The runner SHALL use `Config.Shell.HookTimeout` seconds as a per-command timeout. The runner uses `git.CommandExecutor` from `internal/git/command_executor.go` for `os/exec` invocations.

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
- **THEN** `os.Getenv("TWIGGIT_WORKTREE_PATH")` SHALL return the new worktree's absolute path
- **AND** `os.Getenv("TWIGGIT_BRANCH_NAME")` SHALL return the new branch name

### Requirement: Failure collection

The system SHALL continue running remaining commands when one fails (non-zero exit or timeout). Each failure SHALL be captured in `HookResult.Failures` with `{Command, ExitCode, Output}`. The runner SHALL NOT roll back the worktree.

#### Scenario: One failure, run rest

- **WHEN** command 1 exits non-zero but command 2 succeeds
- **THEN** command 2 SHALL still run
- **AND** `HookResult.Failures[0]` SHALL record command 1's exit code
- **AND** `HookResult.Success` SHALL be `false`

### Requirement: Timeout per command

When a command exceeds `HookTimeout` seconds, the system SHALL kill the process, record a `HookFailure` with the timeout flag, and continue with the next command.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Compile-time interface check

The runner SHALL satisfy `cmdutil.HookRunner` via `var _ cmdutil.HookRunner = (*HookRunnerImpl)(nil)`. The `cmdutil.HookRunner` interface is declared in `internal/cmdutil/hook_runner_iface.go` as the consumer-side interface (per the `golang-cli-architecture` rule "Interfaces where consumed"). The previous `application.HookRunner` interface is removed.

#### Scenario: Compile-time satisfaction check compiles

- **WHEN** the build runs against the migrated package
- **THEN** the `var _ cmdutil.HookRunner = (*git.HookRunnerImpl)(nil)` declaration SHALL compile
- **AND** failure to satisfy the interface SHALL be a build error

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

### Requirement: Service contract

The `Run(ctx, *core.HookRunRequest) (*core.HookResult, error)` method SHALL execute hooks of the requested `HookType` with the supplied context and return `*core.HookResult` plus a non-nil `error` when any hook command fails (non-zero exit, timeout, or filesystem error reading `.twiggit.toml`). The cmd-layer display contract lives in `cli-worktree-hooks`; this spec is silent on display. The previous hook-run-request type under the `application` package and the previous hook-result type under the `domain` package are removed; both live in the `core` package now.

#### Scenario: Definition holds

- **WHEN** the surface described above is exercised
- **THEN** it SHALL match the documented shape exactly
- **AND** the implementation SHALL compile against the contract

#### Scenario: Error return on hook failure

- **WHEN** at least one hook command exits non-zero
- **THEN** `Run` SHALL return a non-nil `error`
- **AND** the `*core.HookResult` SHALL still report `Success = false` and `Failures` per the Failure collection requirement
