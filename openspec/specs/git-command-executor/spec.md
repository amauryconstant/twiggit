# Capability: Git Command Executor

## Purpose

Defines the I/O adapter that wraps `os/exec` invocations for the write-side git client (`*git.cliClient`) and the hook runner (`*git.HookRunner`). Lives in `internal/git/command_executor.go`. Every `Execute` call SHALL execute under a caller-supplied `context.Context` deadline. Executor failures SHALL be wrapped via `git.NewCommandError(args, msg, exitErr)`.

## Requirements

### Requirement: CommandExecutor interface

`git.CommandExecutor` SHALL be a consumer-side interface declared where consumed (currently in `internal/git/command_executor.go`); the interface SHALL have two methods: `Execute(ctx context.Context, dir string, cmd Command, args ...string) (*git.CommandResult, error)` and `ExecuteWithTimeout(ctx context.Context, dir string, cmd Command, timeout time.Duration, args ...string) (*git.CommandResult, error)`. Production code uses the concrete `git.NewCommandExecutor(timeout)`; tests use `git.NewMockCommandExecutor()`.

#### Scenario: Execute accepts context, dir, command, args

- **WHEN** a caller invokes `executor.Execute(ctx, "/tmp/repo", git.CmdGit, "status")`
- **THEN** the executor SHALL run `git status` in `/tmp/repo` with the inherited process environment
- **AND** SHALL honor `ctx` cancellation: when `ctx.Err() != nil`, the running `*exec.Cmd` SHALL be killed via `Process.Kill()`

### Requirement: Typed Command enum for allow-list

The first argument to `Execute*` SHALL be a `git.Command` (typed string enum) restricted to `git.CmdGit` ("git") and `git.CmdSh` ("sh"). The executor SHALL refuse any other `Command` value at runtime with `command_executor: refusing to execute non-allow-listed command "<value>" (allowed: git, sh)`; this makes non-allow-listed process names unrepresentable at the type level and runtime-checked before `os/exec` is invoked.

#### Scenario: Caller passes an allow-listed Command

- **WHEN** a caller invokes `executor.Execute(ctx, "/tmp", git.CmdGit, "status")`
- **THEN** the executor SHALL dispatch to `exec.CommandContext(ctx, "git", "status")`

#### Scenario: Caller passes an unknown Command value

- **WHEN** a caller invokes `executor.Execute(ctx, "/tmp", git.Command("rm"), "-rf", "/")`
- **THEN** the executor SHALL return `nil, error` without invoking `os/exec`
- **AND** the error message SHALL name the rejected value and the allow-list

### Requirement: CommandResult carries stdout, stderr, exit code

`git.CommandResult` SHALL carry `Stdout string`, `Stderr string`, `ExitCode int`, `Duration time.Duration`, and `Err error`. The `ExitCode` SHALL be `-1` when the command was killed by context cancellation (no exit code produced). Stdout and stderr SHALL be captured into separate buffers via `cmd.Stdout` / `cmd.Stderr` (no substring classification — the kernel fd split is authoritative).

#### Scenario: Successful command

- **WHEN** `git status --porcelain` exits 0 with stdout `""` and stderr `""`
- **THEN** `CommandResult{Stdout: "", Stderr: "", ExitCode: 0, Err: nil}` SHALL be returned with no error

#### Scenario: Non-zero exit

- **WHEN** `git checkout missing-branch` exits 1 with stderr `"error: pathspec..."`
- **THEN** `CommandResult{Stderr: "error: pathspec...", ExitCode: 1}` SHALL be returned along with a `*git.CommandError` wrapping the exit error (via `git.NewCommandError`)

#### Scenario: Context cancellation kills the process

- **WHEN** `ctx` is cancelled mid-execution
- **THEN** the underlying `*exec.Cmd` SHALL be killed
- **AND** `CommandResult.ExitCode == -1`
- **AND** the returned error SHALL be a `*git.CommandError` describing the cancellation

### Requirement: NewCommandExecutor returns the production concrete

`git.NewCommandExecutor(defaultTimeout time.Duration) CommandExecutor` SHALL return the production executor wired to `os/exec`. No public constructor SHALL exist for a one-off `CommandExecutor` value; the production wiring is the only path.

#### Scenario: NewCommandExecutor returns non-nil

- **WHEN** a caller invokes `e := git.NewCommandExecutor(30 * time.Second)`
- **THEN** `e` SHALL be non-nil and SHALL satisfy the `git.CommandExecutor` interface

### Requirement: NewMockCommandExecutor returns the test mock

`git.NewMockCommandExecutor() *git.MockCommandExecutor` SHALL return a testify-mock-backed executor. The mock SHALL record `Execute(ctx, dir, cmd, args...)` invocations via `.On("Execute", ...)`. Every mock-based test SHALL end with `mock.AssertExpectations(t)`.

#### Scenario: Mock records Execute calls

- **WHEN** a test sets `mockCmdExecutor.On("Execute", mock.Anything, "/tmp", git.CmdGit, []string{"status"}...).Return(&git.CommandResult{ExitCode: 0}, nil)`
- **AND** a test calls `executor.Execute(ctx, "/tmp", git.CmdGit, "status")`
- **THEN** the recorded call SHALL match the expectation
- **AND** `mockCmdExecutor.AssertExpectations(t)` SHALL pass