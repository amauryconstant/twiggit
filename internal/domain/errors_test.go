package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitWorktreeError_FormatErrorMessage(t *testing.T) {
	t.Run("with branch name and cause", func(t *testing.T) {
		cause := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed to add worktree",
		}
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", cause)
		msg := err.Error()

		assert.Contains(t, msg, "git worktree operation failed")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.Contains(t, msg, "branch: feature-branch")
		assert.Contains(t, msg, "failed")
		assert.Contains(t, msg, "caused by:")
		assert.Contains(t, msg, "git command failed")
	})

	t.Run("with branch name but no cause", func(t *testing.T) {
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", nil)
		msg := err.Error()

		assert.Contains(t, msg, "git worktree operation failed")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.Contains(t, msg, "branch: feature-branch")
		assert.Contains(t, msg, "failed")
		assert.NotContains(t, msg, "caused by:")
	})

	t.Run("without branch name but with cause", func(t *testing.T) {
		cause := NewGitCommandError("git", []string{"worktree", "add"}, 128, "", "fatal: error", "failed", nil)
		err := NewGitWorktreeError("/path/to/worktree", "", "failed", cause)
		msg := err.Error()

		assert.Contains(t, msg, "git worktree operation failed")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.NotContains(t, msg, "branch:")
		assert.Contains(t, msg, "failed")
		assert.Contains(t, msg, "caused by:")
	})

	t.Run("without branch name and without cause", func(t *testing.T) {
		err := NewGitWorktreeError("/path/to/worktree", "", "failed", nil)
		msg := err.Error()

		assert.Contains(t, msg, "git worktree operation failed")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.NotContains(t, msg, "branch:")
		assert.Contains(t, msg, "failed")
		assert.NotContains(t, msg, "caused by:")
	})
}

func TestGitWorktreeError_GetCauseDetails(t *testing.T) {
	t.Run("nil cause returns empty string", func(t *testing.T) {
		err := &GitWorktreeError{Cause: nil}
		details := err.getCauseDetails()
		assert.Empty(t, details)
	})

	t.Run("GitCommandError cause returns formatted error", func(t *testing.T) {
		gitCmdErr := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed",
		}
		err := &GitWorktreeError{Cause: gitCmdErr}
		details := err.getCauseDetails()

		assert.Contains(t, details, "git command failed")
		assert.Contains(t, details, "git")
		assert.Contains(t, details, "worktree add")
		assert.Contains(t, details, "exit code 128")
	})

	t.Run("generic error cause returns error message", func(t *testing.T) {
		genericErr := NewValidationError("request", "field", "value", "validation failed")
		err := &GitWorktreeError{Cause: genericErr}
		details := err.getCauseDetails()

		assert.Contains(t, details, "validation failed")
		assert.Contains(t, details, "field")
		assert.Contains(t, details, "value")
	})
}

func TestGitCommandError_HasUsefulStderr(t *testing.T) {
	t.Run("stderr with useful content returns true", func(t *testing.T) {
		err := &GitCommandError{Stderr: "fatal: Invalid refspec"}
		assert.True(t, err.hasUsefulStderr())
	})

	t.Run("empty stderr returns false", func(t *testing.T) {
		err := &GitCommandError{Stderr: ""}
		assert.False(t, err.hasUsefulStderr())
	})

	t.Run("whitespace-only stderr returns false", func(t *testing.T) {
		err := &GitCommandError{Stderr: "   \n\t  "}
		assert.False(t, err.hasUsefulStderr())
	})

	t.Run("stderr with mixed whitespace and content returns true", func(t *testing.T) {
		err := &GitCommandError{Stderr: "  fatal: error\n"}
		assert.True(t, err.hasUsefulStderr())
	})
}

func TestContainsOnlyWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "spaces only",
			input:    "    ",
			expected: true,
		},
		{
			name:     "tabs only",
			input:    "\t\t\t",
			expected: true,
		},
		{
			name:     "newlines only",
			input:    "\n\n\n",
			expected: true,
		},
		{
			name:     "mixed whitespace",
			input:    " \t\n \t ",
			expected: true,
		},
		{
			name:     "string with content",
			input:    "hello",
			expected: false,
		},
		{
			name:     "string with content and whitespace",
			input:    "  hello world  ",
			expected: false,
		},
		{
			name:     "string with special characters",
			input:    "!@#$%",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsOnlyWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitCommandError_FormatErrorMessage(t *testing.T) {
	t.Run("with useful stderr includes stderr", func(t *testing.T) {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed",
		}
		msg := err.Error()

		assert.Contains(t, msg, "git command failed")
		assert.Contains(t, msg, "git")
		assert.Contains(t, msg, "worktree add")
		assert.Contains(t, msg, "exit code 128")
		assert.Contains(t, msg, "failed")
		assert.Contains(t, msg, "stderr:")
		assert.Contains(t, msg, "fatal: Invalid refspec")
	})

	t.Run("with empty stderr does not include stderr", func(t *testing.T) {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "",
			Message:  "failed",
		}
		msg := err.Error()

		assert.Contains(t, msg, "git command failed")
		assert.NotContains(t, msg, "stderr:")
	})

	t.Run("with whitespace-only stderr does not include stderr", func(t *testing.T) {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "   \n\t  ",
			Message:  "failed",
		}
		msg := err.Error()

		assert.Contains(t, msg, "git command failed")
		assert.NotContains(t, msg, "stderr:")
	})
}

func TestGitWorktreeError_Unwrap(t *testing.T) {
	t.Run("nil cause returns nil", func(t *testing.T) {
		err := &GitWorktreeError{Cause: nil}
		assert.NoError(t, err.Unwrap())
	})

	t.Run("returns cause error", func(t *testing.T) {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitWorktreeError{Cause: cause}
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestGitCommandError_Unwrap(t *testing.T) {
	t.Run("nil cause returns nil", func(t *testing.T) {
		err := &GitCommandError{Cause: nil}
		assert.NoError(t, err.Unwrap())
	})

	t.Run("returns cause error", func(t *testing.T) {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitCommandError{Cause: cause}
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestGitRepositoryError_IsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"not found lowercase", "repository not found", true},
		{"not found uppercase", "REPOSITORY NOT FOUND", true},
		{"does not exist lowercase", "repository does not exist", true},
		{"no such file or directory", "no such file or directory", true},
		{"No Such File Or Directory mixed", "No Such File Or Directory", true},
		{"other error", "permission denied", false},
		{"empty message", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewGitRepositoryError("/path", tt.message, nil)
			assert.Equal(t, tt.expected, err.IsNotFound())
		})
	}
}

func TestGitWorktreeError_IsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"not found lowercase", "worktree not found", true},
		{"not found uppercase", "WORKTREE NOT FOUND", true},
		{"does not exist lowercase", "worktree does not exist", true},
		{"other error", "permission denied", false},
		{"empty message", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewGitWorktreeError("/path", "branch", tt.message, nil)
			assert.Equal(t, tt.expected, err.IsNotFound())
		})
	}
}
