package git

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestExitError creates an ExitError for testing
func createTestExitError(exitCode int) *exec.ExitError {
	// Create a command that fails with desired exit code
	cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", exitCode))
	err := cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return exitErr
		}
	}

	// Fallback - this shouldn't happen in normal testing
	return &exec.ExitError{}
}

// TestExtractExitCode tests pure function for exit code extraction
func TestExtractExitCode(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		expectedCode  int
		expectedFound bool
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedCode:  0,
			expectedFound: false,
		},
		{
			name:          "exec.ExitError with exit code 1",
			err:           createTestExitError(1),
			expectedCode:  1,
			expectedFound: true,
		},
		{
			name:          "exec.ExitError with exit code 127",
			err:           createTestExitError(127),
			expectedCode:  127,
			expectedFound: true,
		},
		{
			name:          "generic error",
			err:           assert.AnError,
			expectedCode:  0,
			expectedFound: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			code, found := extractExitCode(tt.err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

// TestExecuteWithTimeout_Integration tests refactored method end-to-end
func TestExecuteWithTimeout_Integration(t *testing.T) {
	executor := NewCommandExecutor(5 * time.Second)

	t.Run("successful command", func(t *testing.T) {
		ctx := t.Context()
		result, err := executor.ExecuteWithTimeout(ctx, "", CmdGit, 1*time.Second, "--version")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "git version")
		assert.Empty(t, result.Stderr)
		assert.Greater(t, result.Duration, time.Duration(0))
	})

	t.Run("command failure", func(t *testing.T) {
		ctx := t.Context()
		result, err := executor.ExecuteWithTimeout(ctx, "", CmdGit, 1*time.Second, "--bogus-flag-xyz")

		require.Error(t, err) // Error expected for non-zero exit code
		assert.NotEqual(t, 0, result.ExitCode)
		assert.Contains(t, err.Error(), "command exited with non-zero status")
	})

	t.Run("non-allow-listed command rejected", func(t *testing.T) {
		ctx := t.Context()
		_, err := executor.ExecuteWithTimeout(ctx, "", Command("rm"), 1*time.Second, "-rf", "/tmp")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-allow-listed command")
		assert.Contains(t, err.Error(), `"rm"`)
	})
}
