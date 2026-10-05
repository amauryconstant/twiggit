package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Command is the typed set of process names the CommandExecutor
// accepts. It exists so non-allow-listed invocations (anything that is
// not the git CLI or /bin/sh) are unrepresentable at the type level:
// production callers pass CmdGit or CmdSh, and the executor
// runtime-validates the value so the boundary stays tight even if a
// future caller forgets the convention.
type Command string

const (
	// CmdGit is the git CLI binary; the writer places real git
	// subcommand arguments after it. Args are internal-only (derived
	// from CLI flags, not free-form user text) but the typed constant
	// documents the contract at the call site.
	CmdGit Command = "git"

	// CmdSh is the POSIX shell binary; the hook runner uses it as
	// `sh -c <user-script>` where <user-script> is read from
	// .twiggit.toml. This constant names the interpreter only; the
	// script body itself is treated as user input.
	CmdSh Command = "sh"
)

// String returns the underlying process name. Satisfies fmt.Stringer so
// error messages and logs print "git" / "sh" rather than the typed
// representation.
func (c Command) String() string {
	return string(c)
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	ExitCode int           // Process exit code
	Stdout   string        // Standard output
	Stderr   string        // Standard error
	Duration time.Duration // Command execution duration
	Err      error         // Raw execution error (e.g. *exec.ExitError); preserved so callers can wrap without losing the chain
}

// CommandExecutor defines the interface for executing external commands.
// The cmd argument is the typed Command enum (not a free-form string)
// so callers cannot accidentally dispatch an arbitrary process name;
// the concrete commandExecutor runtime-validates against the enum and
// rejects unknown values before invoking os/exec.
type CommandExecutor interface {
	// Execute executes a command in the specified directory
	Execute(ctx context.Context, dir string, cmd Command, args ...string) (*CommandResult, error)

	// ExecuteWithTimeout executes a command with a specific timeout
	ExecuteWithTimeout(ctx context.Context, dir string, cmd Command, timeout time.Duration, args ...string) (*CommandResult, error)
}

// commandExecutor implements CommandExecutor using os/exec
type commandExecutor struct {
	defaultTimeout time.Duration
}

// NewCommandExecutor creates a new CommandExecutor backed by os/exec
func NewCommandExecutor(defaultTimeout time.Duration) CommandExecutor {
	return &commandExecutor{
		defaultTimeout: defaultTimeout,
	}
}

// Execute executes a command in the specified directory
func (ce *commandExecutor) Execute(ctx context.Context, dir string, cmd Command, args ...string) (*CommandResult, error) {
	return ce.ExecuteWithTimeout(ctx, dir, cmd, ce.defaultTimeout, args...)
}

// ExecuteWithTimeout executes a command with a specific timeout. Stdout
// and stderr are captured into separate buffers via cmd.Stdout /
// cmd.Stderr so the CommandResult carries the precise split the
// process emitted (no substring classification — git routes its
// errors to stderr natively, so the classifier was redundant with
// the kernel's own fd split).
//
// cmd MUST be one of the allow-listed Command constants (CmdGit or
// CmdSh). The dispatch below uses string literals for the
// exec.CommandContext program argument so the program name is a
// compile-time constant; the typed Command enum plus the runtime
// default case keep the boundary tight against future regressions.
func (ce *commandExecutor) ExecuteWithTimeout(ctx context.Context, dir string, cmd Command, timeout time.Duration, args ...string) (*CommandResult, error) {
	start := time.Now()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var command *exec.Cmd
	switch cmd {
	case CmdGit:
		// #nosec G204 -- program name is the constant "git"; args are
		// the internal subcommand list (worktree, branch, …) plus CLI-derived
		// flags, none of which are arbitrary user-supplied text.
		command = exec.CommandContext(timeoutCtx, "git", args...)
	case CmdSh:
		// #nosec G204 -- sh is invoked as `sh -c <user-script>` where
		// <user-script> is read from .twiggit.toml and forwarded to
		// sh as the user-authored hook payload; the script body is the
		// user input we deliberately pass through.
		command = exec.CommandContext(timeoutCtx, "sh", args...)
	default:
		return nil, fmt.Errorf("command_executor: refusing to execute non-allow-listed command %q (allowed: %s, %s)", cmd, CmdGit, CmdSh)
	}
	if dir != "" {
		command.Dir = dir
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = &stdoutBuf
	command.Stderr = &stderrBuf

	err := command.Run()
	duration := time.Since(start)

	result := &CommandResult{
		ExitCode: 0,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
		Err:      err,
	}
	if code, found := extractExitCode(err); found {
		result.ExitCode = code
	}

	if err != nil {
		if result.ExitCode == 0 {
			return nil, NewCommandError("execute", fmt.Sprintf("failed to execute command: %v", err), err)
		}
		return result, NewCommandError("non-zero-exit", "command exited with non-zero status", result.Err)
	}

	return result, nil
}

// extractExitCode extracts the exit code from an error if it's an exec.ExitError
func extractExitCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}

	exitError := &exec.ExitError{}
	if errors.As(err, &exitError) {
		return exitError.ExitCode(), true
	}

	return 0, false
}
