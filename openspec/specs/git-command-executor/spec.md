# Capability: Git Command Executor

## Purpose

Defines the I/O adapter that wraps `os/exec` invocations for the write-side git client (`*git.cliClient`) and the hook runner (`*git.HookRunner`). Lives in `internal/git/command_executor.go`. Every `Execute` call SHALL execute under a caller-supplied `context.Context` deadline. Executor failures SHALL be wrapped via `git.NewCommandError(args, msg, exitErr)`.

## Requirements

### Requirement: CommandExecutor interface

`git.CommandExecutor` SHALL be a consumer-side interface declared where consumed (currently in `internal/git/command_executor.go`); the interface SHALL have exactly one method: `Execute(ctx context.Context, args []string, dir string, env []string) (git.CommandResult, error)`. Production code uses the concrete `git.NewCommandExecutor()`; tests use `git.NewMockCommandExecutor()`.

#### Scenario: Execute accepts context, args, dir, env

- **WHEN** a caller invokes `executor.Execute(ctx, []string{"git", "status"}, "/tmp/repo", nil)`
- **THEN** the executor SHALL run `git status` in `/tmp/repo` with the inherited process environment
- **AND** SHALL honor `ctx` cancellation: when `ctx.Err() != nil`, the running `*exec.Cmd` SHALL be killed via `Process.Kill()`

### Requirement: CommandResult carries stdout, stderr, exit code

`git.CommandResult` SHALL carry `Stdout []byte`, `Stderr []byte`, and `ExitCode int`. The `ExitCode` SHALL be `-1` when the command was killed by context cancellation (no exit code produced). Both `Stdout` and `Stderr` SHALL be defensive copies (`slices.Clone`) of the underlying buffer's contents.

#### Scenario: Successful command

- **WHEN** `git status --porcelain` exits 0 with stdout `""` and stderr `""`
- **THEN** `CommandResult{Stdout: nil, Stderr: nil, ExitCode: 0}` SHALL be returned with no error

#### Scenario: Non-zero exit

- **WHEN** `git checkout missing-branch` exits 1 with stderr `"error: pathspec` missing entry`..."
- **THEN** `CommandResult{Stderr: []byte("error: pathspec..."), ExitCode: 1}` SHALL be returned along with a `*core.OperationError` wrapping the exit error (via `git.NewCommandError`)

#### Scenario: Context cancellation kills the process

- **WHEN** `ctx` is cancelled mid-execution
- **THEN** the underlying `*exec.Cmd` SHALL be killed
- **AND** `CommandResult.ExitCode == -1`
- **AND** the returned error SHALL be a `*core.OperationError` with `Op == "git.command.cancel"`

### Requirement: NewCommandExecutor returns the production concrete

`git.NewCommandExecutor() *git.commandExecutor` SHALL return the production executor wired to `os/exec`. No public constructor SHALL exist for a one-off `CommandExecutor` value; the production wiring is the only path.

#### Scenario: NewCommandExecutor returns non-nil

- **WHEN** a caller invokes `e := git.NewCommandExecutor()`
- **THEN** `e` SHALL be non-nil and SHALL satisfy the `git.CommandExecutor` interface

### Requirement: NewMockCommandExecutor returns the test mock

`git.NewMockCommandExecutor() *git.MockCommandExecutor` SHALL return a testify-mock-backed executor. The mock SHALL record `Execute(ctx, args, dir, env)` invocations via `.On("Execute", ...)`. Every mock-based test SHALL end with `mock.AssertExpectations(t)`.

#### Scenario: Mock records Execute calls

- **WHEN** a test sets `mockCmdExecutor.On("Execute", mock.Anything, []string{"git", "status"}, "/tmp", mock.Anything).Return(git.CommandResult{ExitCode: 0}, nil)`
- **AND** a test calls `executor.Execute(ctx, []string{"git", "status"}, "/tmp", nil)`
- **THEN** the recorded call SHALL match the expectation
- **AND** `mockCmdExecutor.AssertExpectations(t)` SHALL pass