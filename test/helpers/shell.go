package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ShellTestHelper provides functional shell command testing utilities
type ShellTestHelper struct {
	t           *testing.T
	command     string
	args        []string
	workingDir  string
	environment map[string]string
	timeout     int
}

// NewShellTestHelper creates a new ShellTestHelper instance
func NewShellTestHelper(t *testing.T) *ShellTestHelper {
	t.Helper()
	return &ShellTestHelper{
		t:           t,
		environment: make(map[string]string),
		timeout:     30, // Default 30 second timeout
	}
}

// WithCommand sets the command for functional composition
func (h *ShellTestHelper) WithCommand(command string) *ShellTestHelper {
	h.command = command
	return h
}

// WithArgs sets the arguments for functional composition
func (h *ShellTestHelper) WithArgs(args ...string) *ShellTestHelper {
	h.args = args
	return h
}

// WithWorkingDirectory sets the working directory for command execution
func (h *ShellTestHelper) WithWorkingDirectory(dir string) *ShellTestHelper {
	h.workingDir = dir
	return h
}

// WithEnvironment sets an environment variable for command execution
func (h *ShellTestHelper) WithEnvironment(key, value string) *ShellTestHelper {
	h.environment[key] = value
	return h
}

// WithTimeout sets the timeout for command execution in seconds
func (h *ShellTestHelper) WithTimeout(seconds int) *ShellTestHelper {
	h.timeout = seconds
	return h
}

// ExecuteCommand executes a shell command with the configured options
func (h *ShellTestHelper) ExecuteCommand(command string, args ...string) (string, error) {
	// Use provided command/args or fall back to configured ones
	cmdToRun := command
	cmdArgs := args

	if h.command != "" {
		cmdToRun = h.command
	}
	if len(h.args) > 0 {
		cmdArgs = h.args
	}

	cmd := exec.Command(cmdToRun, cmdArgs...)

	// Set working directory if specified
	if h.workingDir != "" {
		cmd.Dir = h.workingDir
	}

	// Set environment variables
	if len(h.environment) > 0 {
		env := os.Environ()
		for key, value := range h.environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Execute command and capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ExecuteCommandWithOutput executes a command and returns both stdout and stderr
func (h *ShellTestHelper) ExecuteCommandWithOutput(command string, args ...string) (stdout, stderr string, err error) {
	// Use provided command/args or fall back to configured ones
	cmdToRun := command
	cmdArgs := args

	if h.command != "" {
		cmdToRun = h.command
	}
	if len(h.args) > 0 {
		cmdArgs = h.args
	}

	cmd := exec.Command(cmdToRun, cmdArgs...)

	// Set working directory if specified
	if h.workingDir != "" {
		cmd.Dir = h.workingDir
	}

	// Set environment variables
	if len(h.environment) > 0 {
		env := os.Environ()
		for key, value := range h.environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), err
}

// CommandExists checks if a command exists in the system PATH
func (h *ShellTestHelper) CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// GetWorkingDirectory returns the current working directory
func (h *ShellTestHelper) GetWorkingDirectory() string {
	pwd, err := os.Getwd()
	if err != nil {
		h.t.Fatalf("Failed to get working directory: %v", err)
	}
	return pwd
}

// Reset resets the helper configuration for reuse
func (h *ShellTestHelper) Reset() *ShellTestHelper {
	h.command = ""
	h.args = nil
	h.workingDir = ""
	h.environment = make(map[string]string)
	h.timeout = 30
	return h
}
