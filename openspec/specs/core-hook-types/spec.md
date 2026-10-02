# Capability: Hook Types

## Purpose

Defines the hook-domain value objects used by the post-create hook runner. Types live in `internal/core/hook_types.go`; the runner that consumes them lives in `internal/git/hook_runner.go` and is owned by `git-hook-runner`.

## Requirements

### Requirement: HookConfig

`core.HookConfig` SHALL carry a list of `HookDefinition` entries grouped by `HookType` (currently only `HookTypePostCreate`). The config loads from `[hooks.post-create]` in `.twiggit.toml`.

#### Scenario: HookConfig decodes TOML block

- **WHEN** `.twiggit.toml` contains:
  ```toml
  [hooks.post-create]
  commands = ["mise trust", "npm install"]
  ```
- **THEN** `HookConfig.PostCreate` SHALL be `[]core.HookDefinition{{Command: "mise trust"}, {Command: "npm install"}}`

### Requirement: HookDefinition

`core.HookDefinition` SHALL be a struct with `Command string` (the shell command to execute), `WorkingDirectory string` (optional; defaults to the new worktree path), and `TimeoutSeconds int` (optional; defaults to `Config.Shell.HookTimeout`).

#### Scenario: Default working directory and timeout

- **WHEN** a `HookDefinition` is constructed with only `Command = "mise trust"`
- **THEN** `WorkingDirectory == ""` and `TimeoutSeconds == 0`
- **AND** the runner SHALL fall back to the new worktree path for `WorkingDirectory`
- **AND** SHALL fall back to `Config.Shell.HookTimeout` for the per-command deadline

### Requirement: HookType enum

`core.HookType` SHALL be an `int` type with an explicit `HookTypeUnknown` sentinel at `iota` position 0. Currently only `HookTypePostCreate` is defined; future hook lifecycle stages (pre-create, post-delete, post-prune) will add new variants without renumbering.

#### Scenario: HookType unknown sentinel

- **WHEN** `var t core.HookType` is declared
- **THEN** `t == core.HookTypeUnknown` (iota 0)

### Requirement: HookFailure

`core.HookFailure` SHALL carry `Command string`, `ExitCode int`, `Output string`, and `TimedOut bool`. The runner populates one `HookFailure` per non-zero exit. `TimedOut == true` distinguishes a process killed by `context.WithTimeout` from a process that exited non-zero on its own.

#### Scenario: Timed-out hook

- **WHEN** a hook command exceeds `Config.Shell.HookTimeout`
- **THEN** the runner SHALL record `HookFailure{TimedOut: true}` for that command
- **AND** SHALL continue with the next hook

### Requirement: HookResult uses HasExecuted and IsSuccessful

`core.HookResult` SHALL carry `HasExecuted bool` (true when at least one hook ran), `IsSuccessful bool` (true when every hook exited 0 or no hooks ran), and `Failures []core.HookFailure` (empty when no failures). The cmd layer reads these fields directly without string-keyed access.

#### Scenario: No hooks configured

- **WHEN** `.twiggit.toml` has no `post-create` block
- **THEN** `HookResult{HasExecuted: false, IsSuccessful: true, Failures: nil}` SHALL be returned

#### Scenario: One of three hooks fails

- **WHEN** `mise trust` exits 1, `npm install` exits 0, `bundle install` exits 0
- **THEN** `HookResult.HasExecuted == true`, `IsSuccessful == false`, `len(Failures) == 1`, `Failures[0].Command == "mise trust"`