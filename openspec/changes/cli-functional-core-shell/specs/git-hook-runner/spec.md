# Spec Delta

## MODIFIED Requirements

### Requirement: HookRunner implementation lives in internal/git

The `HookRunner` implementation SHALL live in `internal/git/hook_runner.go` (migrated from `internal/infrastructure/hook_runner.go`). The runner SHALL use `core.CommandExecutor` from `internal/git/command_executor.go` for `os/exec` invocations. The runner SHALL return `*core.HookResult`.

#### Scenario: HookRunner file path is internal/git
- **WHEN** the runner source file is located
- **THEN** it SHALL be `internal/git/hook_runner.go`

#### Scenario: Run signature returns core.HookResult
- **WHEN** the runner's `Run` method is called with a valid `HookRunRequest`
- **THEN** it SHALL return a `*core.HookResult` (renamed from `*domain.HookResult`)

### Requirement: Interface satisfaction check is local

The runner SHALL satisfy the consumer-side `HookRunner` interface via `var _ git.HookRunner = (*git.HookRunnerImpl)(nil)` declared in `internal/git/hook_runner.go` (the interface is defined next to the implementation in the consumer-side Tier 2 layout).

#### Scenario: Interface satisfaction compiles
- **WHEN** the build runs against the migrated package
- **THEN** the satisfaction check SHALL compile; failure to satisfy the interface SHALL be a build error
