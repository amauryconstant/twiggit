//go:build integration
// +build integration

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainConfigLoadFailure tests that main handles config load failures gracefully
func TestMainConfigLoadFailure_InvalidYAML(t *testing.T) {
	// Create a temp directory with invalid config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Write invalid YAML to config file
	configPath := filepath.Join(configDir, "config.toml")
	invalidYAML := `
projects_dir = "/tmp/projects
# Missing closing quote and invalid structure
worktrees_dir = /tmp/worktrees
`
	require.NoError(t, os.WriteFile(configPath, []byte(invalidYAML), 0644))

	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute with invalid config
	cmd := exec.Command(binaryPath, "--help")
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+tempDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Should fail with non-zero exit code
	require.Error(t, err, "Should fail with invalid config")
	// Exit code should be non-zero (may be 1 for config error)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		assert.NotEqual(t, 0, exitErr.ExitCode(), "Exit code should be non-zero for config error")
	}
}

// TestMainConfigLoadFailure_MissingDirectory tests config load when directory cannot be created
func TestMainConfigLoadFailure_MissingDirectory(t *testing.T) {
	// This test verifies graceful handling when XDG_CONFIG_HOME points to non-existent location
	// The config manager should fall back to defaults gracefully

	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute with non-existent config home (will use defaults)
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent")

	cmd := exec.Command(binaryPath, "--help")
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+nonExistentPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Help command should succeed even with missing config dir (uses defaults)
	require.NoError(t, err, "Help should work with defaults")
	assert.Contains(t, stdout.String(), "twiggit")
}

// TestMainSuccessfulExecution tests successful execution path with valid config
func TestMainSuccessfulExecution_ValidConfig(t *testing.T) {
	// Create a temp directory with valid config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Create valid config
	projectsDir := filepath.Join(tempDir, "projects")
	worktreesDir := filepath.Join(tempDir, "worktrees")
	require.NoError(t, os.MkdirAll(projectsDir, 0755))
	require.NoError(t, os.MkdirAll(worktreesDir, 0755))

	configContent := fmt.Sprintf(`
projects_dir = "%s"
worktrees_dir = "%s"
default_source_branch = "main"

[context_detection]
cache_ttl = "5m"
git_operation_timeout = "30s"
enable_git_validation = true

[git]
cli_timeout = 30
cache_enabled = true

[completion]
timeout = "500ms"
`, projectsDir, worktreesDir)

	configPath := filepath.Join(configDir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute version command (simple command that should work)
	cmd := exec.Command(binaryPath, "version")
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+tempDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "Version command should succeed with valid config")
	assert.Contains(t, stdout.String(), "twiggit")
}

// TestMainHelpCommand tests help command execution
func TestMainHelpCommand(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute help command
	cmd := exec.Command(binaryPath, "--help")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "Help command should succeed")
	output := stdout.String()
	assert.Contains(t, output, "twiggit")
	assert.Contains(t, output, "worktree")
	assert.Contains(t, output, "Usage:")
}

// TestMainCommandExecutionFailure tests command execution failure with appropriate exit codes
func TestMainCommandExecutionFailure_InvalidCommand(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute with unknown command
	cmd := exec.Command(binaryPath, "nonexistent-command")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err, "Should fail with unknown command")

	// Check exit code
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		assert.Equal(t, 1, exitErr.ExitCode(), "Should exit with code 1 for unknown command")
	}
	assert.Contains(t, stderr.String(), "unknown command")
}

// TestMainCommandExecutionFailure_InvalidArguments tests command with invalid arguments
func TestMainCommandExecutionFailure_InvalidArguments(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)

	// Execute create without required arguments
	cmd := exec.Command(binaryPath, "create")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err, "Should fail without arguments")

	// Check exit code
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Usage errors should exit with code 2
		assert.Equal(t, 2, exitErr.ExitCode(), "Should exit with code 2 for usage error")
	}
	// Cobra uses "accepts" not "required" for argument count errors
	assert.Contains(t, stderr.String(), "accepts")
}

// buildTestBinary builds the test binary and returns its path
func buildTestBinary(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")

	// Normalize path to project root
	if strings.HasSuffix(cwd, "test/integration") {
		cwd = filepath.Dir(filepath.Dir(cwd))
	}

	binDir := filepath.Join(cwd, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755), "Failed to create bin directory")

	binaryPath := filepath.Join(binDir, "twiggit-test")

	// Build the binary with integration tag
	cmd := exec.Command("go", "build", "-tags=integration", "-o", binaryPath, "main.go")
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build test binary: %s", string(output))

	return binaryPath
}
