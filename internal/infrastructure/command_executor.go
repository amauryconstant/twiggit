package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"twiggit/internal/domain"
)

// CommandResult represents the result of executing a command
type CommandResult struct {
	ExitCode int           // Process exit code
	Stdout   string        // Standard output
	Stderr   string        // Standard error
	Duration time.Duration // Command execution duration
}

// CommandExecutor defines the interface for executing external commands
type CommandExecutor interface {
	// Execute executes a command in the specified directory
	Execute(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error)

	// ExecuteWithTimeout executes a command with a specific timeout
	ExecuteWithTimeout(ctx context.Context, dir, cmd string, timeout time.Duration, args ...string) (*CommandResult, error)
}

// DefaultCommandExecutor implements CommandExecutor using os/exec
type DefaultCommandExecutor struct {
	defaultTimeout time.Duration
}

// NewDefaultCommandExecutor creates a new DefaultCommandExecutor
func NewDefaultCommandExecutor(defaultTimeout time.Duration) *DefaultCommandExecutor {
	return &DefaultCommandExecutor{
		defaultTimeout: defaultTimeout,
	}
}

// Execute executes a command in the specified directory
func (e *DefaultCommandExecutor) Execute(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
	return e.ExecuteWithTimeout(ctx, dir, cmd, e.defaultTimeout, args...)
}

// ExecuteWithTimeout executes a command with a specific timeout
func (e *DefaultCommandExecutor) ExecuteWithTimeout(ctx context.Context, dir, cmd string, timeout time.Duration, args ...string) (*CommandResult, error) {
	start := time.Now()

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare the command
	command := exec.CommandContext(timeoutCtx, cmd, args...)
	if dir != "" {
		command.Dir = dir
	}

	// Execute the command
	output, err := command.CombinedOutput()
	duration := time.Since(start)

	// Create result using pure function
	result := createCommandResult(cmd, args, output, err, duration)

	// Check if command failed to start (e.g., command not found)
	if err != nil {
		if _, found := extractExitCode(err); !found {
			return nil, domain.NewGitCommandError(cmd, args, -1, result.Stdout, result.Stderr,
				fmt.Sprintf("failed to execute command: %v", err), err)
		}
	}

	// For non-zero exit codes, return the result with an error (original behavior)
	if result.ExitCode != 0 {
		return result, domain.NewGitCommandError(cmd, args, result.ExitCode, result.Stdout, result.Stderr,
			"command exited with non-zero status", nil)
	}

	return result, nil
}

// isErrorLine determines if a line represents an error, fatal, or warning message
func isErrorLine(line string) bool {
	lowerLine := strings.ToLower(line)
	return strings.Contains(lowerLine, "error:") ||
		strings.Contains(lowerLine, "fatal:") ||
		strings.Contains(lowerLine, "warning:")
}

// classifyLines separates lines into stdout and stderr based on error patterns
func classifyLines(lines []string) (stdout, stderr []string) {
	stdout = make([]string, 0)
	stderr = make([]string, 0)
	for _, line := range lines {
		if isErrorLine(line) {
			stderr = append(stderr, line)
		} else {
			stdout = append(stdout, line)
		}
	}
	return stdout, stderr
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

// createCommandResult creates a CommandResult from execution parameters
func createCommandResult(_ string, _ []string, output []byte, err error, duration time.Duration) *CommandResult {
	exitCode := 0
	if code, found := extractExitCode(err); found {
		exitCode = code
	}

	// Parse output into stdout and stderr
	outputStr := string(output)
	if strings.Contains(strings.ToLower(outputStr), "error:") ||
		strings.Contains(strings.ToLower(outputStr), "fatal:") ||
		strings.Contains(strings.ToLower(outputStr), "warning:") {
		lines := strings.Split(outputStr, "\n")
		stdoutLines, stderrLines := classifyLines(lines)
		return &CommandResult{
			ExitCode: exitCode,
			Stdout:   strings.Join(stdoutLines, "\n"),
			Stderr:   strings.Join(stderrLines, "\n"),
			Duration: duration,
		}
	}

	return &CommandResult{
		ExitCode: exitCode,
		Stdout:   outputStr,
		Stderr:   "",
		Duration: duration,
	}
}
