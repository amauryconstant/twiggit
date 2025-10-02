package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CommandExecutorTestSuite struct {
	suite.Suite
	executor *DefaultCommandExecutor
}

func TestCommandExecutorSuite(t *testing.T) {
	suite.Run(t, new(CommandExecutorTestSuite))
}

func (s *CommandExecutorTestSuite) SetupTest() {
	s.executor = NewDefaultCommandExecutor(5 * time.Second)
}

// TestIsErrorLine tests the pure function for error line detection
func (s *CommandExecutorTestSuite) TestIsErrorLine() {
	testCases := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "error prefix lowercase",
			line:     "error: something went wrong",
			expected: true,
		},
		{
			name:     "error prefix uppercase",
			line:     "ERROR: something went wrong",
			expected: true,
		},
		{
			name:     "error prefix mixed case",
			line:     "Error: something went wrong",
			expected: true,
		},
		{
			name:     "fatal prefix lowercase",
			line:     "fatal: something went wrong",
			expected: true,
		},
		{
			name:     "fatal prefix uppercase",
			line:     "FATAL: something went wrong",
			expected: true,
		},
		{
			name:     "warning prefix lowercase",
			line:     "warning: something went wrong",
			expected: true,
		},
		{
			name:     "warning prefix uppercase",
			line:     "WARNING: something went wrong",
			expected: true,
		},
		{
			name:     "normal output line",
			line:     "commit abc1234",
			expected: false,
		},
		{
			name:     "line containing error but not prefix",
			line:     "this line contains error but not as prefix",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "whitespace only",
			line:     "   ",
			expected: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			result := isErrorLine(tt.line)
			s.Equal(tt.expected, result)
		})
	}
}

// TestClassifyLines tests the pure function for line classification
func (s *CommandExecutorTestSuite) TestClassifyLines() {
	testCases := []struct {
		name           string
		lines          []string
		expectedStdout []string
		expectedStderr []string
	}{
		{
			name:           "all normal lines",
			lines:          []string{"commit abc1234", "branch main", "status clean"},
			expectedStdout: []string{"commit abc1234", "branch main", "status clean"},
			expectedStderr: []string{},
		},
		{
			name:           "all error lines",
			lines:          []string{"error: something failed", "fatal: repository not found", "warning: old version"},
			expectedStdout: []string{},
			expectedStderr: []string{"error: something failed", "fatal: repository not found", "warning: old version"},
		},
		{
			name:           "mixed lines",
			lines:          []string{"commit abc1234", "error: conflict detected", "branch main", "warning: detached HEAD"},
			expectedStdout: []string{"commit abc1234", "branch main"},
			expectedStderr: []string{"error: conflict detected", "warning: detached HEAD"},
		},
		{
			name:           "empty lines",
			lines:          []string{"", "   ", "commit abc1234"},
			expectedStdout: []string{"", "   ", "commit abc1234"},
			expectedStderr: []string{},
		},
		{
			name:           "case insensitive error detection",
			lines:          []string{"ERROR: uppercase", "Error: mixed case", "warning: WARNING"},
			expectedStdout: []string{},
			expectedStderr: []string{"ERROR: uppercase", "Error: mixed case", "warning: WARNING"},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			stdout, stderr := classifyLines(tt.lines)
			s.Equal(tt.expectedStdout, stdout)
			s.Equal(tt.expectedStderr, stderr)
		})
	}
}

// createTestExitError creates an ExitError for testing
func createTestExitError(exitCode int) *exec.ExitError {
	// Create a command that fails with the desired exit code
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

// TestExtractExitCode tests the pure function for exit code extraction
func (s *CommandExecutorTestSuite) TestExtractExitCode() {
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
		s.Run(tt.name, func() {
			code, found := extractExitCode(tt.err)
			s.Equal(tt.expectedCode, code)
			s.Equal(tt.expectedFound, found)
		})
	}
}

// TestCreateCommandResult tests the pure function for result creation
func (s *CommandExecutorTestSuite) TestCreateCommandResult() {
	testCases := []struct {
		name         string
		cmd          string
		args         []string
		output       []byte
		err          error
		duration     time.Duration
		expectedExit int
		expectedOut  string
		expectedErr  string
	}{
		{
			name:         "successful command",
			cmd:          "git",
			args:         []string{"status"},
			output:       []byte("On branch main\nnothing to commit"),
			err:          nil,
			duration:     100 * time.Millisecond,
			expectedExit: 0,
			expectedOut:  "On branch main\nnothing to commit",
			expectedErr:  "",
		},
		{
			name:         "command with error output",
			cmd:          "git",
			args:         []string{"status"},
			output:       []byte("error: not a git repository"),
			err:          createTestExitError(128),
			duration:     50 * time.Millisecond,
			expectedExit: 128,
			expectedOut:  "",
			expectedErr:  "error: not a git repository",
		},
		{
			name:         "mixed output",
			cmd:          "git",
			args:         []string{"status"},
			output:       []byte("On branch main\nerror: some files are modified\nwarning: detached HEAD"),
			err:          createTestExitError(1),
			duration:     75 * time.Millisecond,
			expectedExit: 1,
			expectedOut:  "On branch main",
			expectedErr:  "error: some files are modified\nwarning: detached HEAD",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			result := createCommandResult(tt.cmd, tt.args, tt.output, tt.err, tt.duration)

			s.Equal(tt.expectedExit, result.ExitCode)
			s.Equal(tt.expectedOut, result.Stdout)
			s.Equal(tt.expectedErr, result.Stderr)
			s.Equal(tt.duration, result.Duration)
		})
	}
}

// TestExecuteWithTimeout_Integration tests the refactored method end-to-end
func (s *CommandExecutorTestSuite) TestExecuteWithTimeout_Integration() {
	s.Run("successful command", func() {
		ctx := context.Background()
		result, err := s.executor.ExecuteWithTimeout(ctx, "", "echo", 1*time.Second, "hello world")

		s.Require().NoError(err)
		s.Equal(0, result.ExitCode)
		s.Equal("hello world\n", result.Stdout)
		s.Empty(result.Stderr)
		s.Greater(result.Duration, time.Duration(0))
	})

	s.Run("command failure", func() {
		ctx := context.Background()
		result, err := s.executor.ExecuteWithTimeout(ctx, "", "false", 1*time.Second)

		s.Require().Error(err) // Error expected for non-zero exit code
		s.Equal(1, result.ExitCode)
		s.Empty(result.Stdout)
		s.Empty(result.Stderr)
		s.Contains(err.Error(), "command exited with non-zero status")
	})

	s.Run("command not found", func() {
		ctx := context.Background()
		_, err := s.executor.ExecuteWithTimeout(ctx, "", "nonexistent-command-12345", 1*time.Second)

		s.Require().Error(err)
		s.Contains(err.Error(), "failed to execute command")
	})
}
