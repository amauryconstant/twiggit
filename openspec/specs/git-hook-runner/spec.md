# Capability: Git Hook Runner

## Purpose

Documents the `HookRunner` implementation for the Tier 2 layout. Lives in `internal/git/hook_runner.go`; uses `git.CommandExecutor` for `os/exec` invocations; returns `*core.HookResult`. The consumer-side `HookRunner` interface lives in `internal/cmdutil/hook_runner_iface.go` per the interfaces-where-consumed rule; the implementation satisfies it via a compile-time check.

## Requirements

### Requirement: HookRunner interface lives in internal/cmdutil

The `HookRunner` consumer-side interface SHALL be declared in `internal/cmdutil/hook_runner_iface.go`:

```go
type HookRunner interface {
    Run(ctx context.Context, req *core.HookRunRequest) (*core.HookResult, error)
}
```

The interface sits with the consumer (cmd/) because each cmd/<command>.go consumes the runner; the previous `application.HookRunner` location is removed.

#### Scenario: Interface file path is internal/cmdutil

- **WHEN** the consumer-side interface source file is located
- **THEN** it SHALL be `internal/cmdutil/hook_runner_iface.go`

#### Scenario: Compile-time satisfaction check compiles

- **WHEN** the build runs against the migrated package
- **THEN** the `var _ cmdutil.HookRunner = (*git.HookRunner)(nil)` declaration SHALL compile

### Requirement: HookRunner implementation lives in internal/git

The `HookRunner` implementation SHALL live in `internal/git/hook_runner.go`. The runner SHALL use `git.CommandExecutor` from `internal/git/command_executor.go` for `os/exec` invocations. The runner SHALL return a `*core.HookResult` (renamed from `*domain.HookResult`).

#### Scenario: Implementation file path is internal/git

- **WHEN** the runner implementation source file is located
- **THEN** it SHALL be `internal/git/hook_runner.go`

#### Scenario: Run signature returns *core.HookResult

- **WHEN** the runner's `Run` method is called with a valid `*core.HookRunRequest`
- **THEN** it SHALL return a `*core.HookResult`

#### Scenario: Run accepts *core.HookRunRequest

- **WHEN** the runner signature is read
- **THEN** the input parameter SHALL be `*core.HookRunRequest`

### Requirement: Read context file from .twiggit.toml

The system SHALL read `.twiggit.toml` at the repository root and extract either `[hooks.post-create].command` (a single command) or `[[hooks.post-create]]` blocks (a list of `core.HookDefinition` entries, each with its own `command`, optional `working_directory`, and optional `timeout_seconds`). Missing config file or empty command list SHALL result in a no-op (`HookResult.HasExecuted = false`). Per-command timeout falls back to `Config.Shell.HookTimeout` when the definition's `timeout_seconds` is zero.

#### Scenario: Missing config is a no-op

- **WHEN** `.twiggit.toml` does not exist
- **THEN** the runner returns `&core.HookResult{HasExecuted: false}` with no error

#### Scenario: Per-definition timeout

- **WHEN** a `HookDefinition` carries `timeout_seconds = 5` and the command exceeds 5 seconds
- **THEN** the runner kills the process, records a `HookFailure{TimedOut: true}` for that definition, and continues with the next definition

#### Scenario: Failures do not roll back the worktree

- **WHEN** one definition in `[[hooks.post-create]]` exits non-zero
- **THEN** the runner captures the failure in `HookResult.Failures` and continues running remaining definitions

### Requirement: Hook runner uses typed Command enum

The runner SHALL dispatch shell invocations through `git.CommandExecutor` with the typed constant `git.CmdSh` (not a free-form string), passing the user-authored script as the `args...` tail. The boundary is type-safe: a non-allow-listed program name is unrepresentable.

#### Scenario: Hook command is invoked via CmdSh

- **WHEN** the runner executes a hook definition
- **THEN** it SHALL call `executor.ExecuteWithTimeout(ctx, workDir, git.CmdSh, timeout, "-c", script)`
- **AND** the reader confirms `git.CmdSh` is the typed enum constant, not the literal `"sh"`

### Requirement: Env vars injected per command

The runner SHALL set the following env vars for each hook command:

| Variable | Value |
|---|---|
| `TWIGGIT_WORKTREE_PATH` | Absolute path to the new worktree |
| `TWIGGIT_PROJECT_NAME` | Project identifier |
| `TWIGGIT_BRANCH_NAME` | New branch name |
| `TWIGGIT_SOURCE_BRANCH` | Branch created from |
| `TWIGGIT_MAIN_REPO_PATH` | Path to the main repository |

#### Scenario: Env vars set for each command

- **WHEN** the runner executes a hook command
- **THEN** every env var listed above is set in the command's process environment
