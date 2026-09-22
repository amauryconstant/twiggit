# Spec Delta

## Purpose

Documents the `HookRunner` implementation for the Tier 2 layout. Lives in `internal/git/hook_runner.go`; uses `git.CommandExecutor` for `os/exec` invocations; returns `*core.HookResult`. The consumer-side `HookRunner` interface lives in `internal/cmdutil/hook_runner_iface.go` per the interfaces-where-consumed rule; the implementation satisfies it via a compile-time check. The legacy `infrastructure-hook-runner` spec is left intact per the deferred-migration non-goal in the proposal.

## ADDED Requirements

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
- **THEN** the `var _ cmdutil.HookRunner = (*git.HookRunnerImpl)(nil)` declaration SHALL compile

### Requirement: HookRunner implementation lives in internal/git

The `HookRunnerImpl` implementation SHALL live in `internal/git/hook_runner.go` (migrated from `internal/infrastructure/hook_runner.go`). The runner SHALL use `git.CommandExecutor` from `internal/git/command_executor.go` for `os/exec` invocations. The runner SHALL return a `*core.HookResult` (renamed from `*domain.HookResult`).

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

The system SHALL read `.twiggit.toml` at the repository root and extract `[hooks.post-create].commands`. Missing config file or empty command list SHALL result in a no-op (`HookResult.Executed = false`). Per-command timeout uses `Config.Shell.HookTimeout` seconds.

#### Scenario: Missing config is a no-op
- **WHEN** `.twiggit.toml` does not exist
- **THEN** the runner returns `&core.HookResult{Executed: false}` with no error

#### Scenario: Per-command timeout
- **WHEN** a command exceeds `Config.Shell.HookTimeout` seconds
- **THEN** the runner kills the process, records a failure with the timeout flag, and continues with the next command

#### Scenario: Failures do not roll back the worktree
- **WHEN** one command in `[hooks.post-create].commands` exits non-zero
- **THEN** the runner captures the failure in `HookResult.Failures` and continues running remaining commands

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
