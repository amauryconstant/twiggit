package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func TestErrors(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

func (s *ErrorsTestSuite) TestGitWorktreeError_FormatErrorMessage() {
	s.Run("with branch name and cause", func() {
		cause := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed to add worktree",
		}
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", cause)
		msg := err.Error()

		s.Contains(msg, "git worktree operation failed")
		s.Contains(msg, "/path/to/worktree")
		s.Contains(msg, "branch: feature-branch")
		s.Contains(msg, "failed")
		s.Contains(msg, "caused by:")
		s.Contains(msg, "git command failed")
	})

	s.Run("with branch name but no cause", func() {
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", nil)
		msg := err.Error()

		s.Contains(msg, "git worktree operation failed")
		s.Contains(msg, "/path/to/worktree")
		s.Contains(msg, "branch: feature-branch")
		s.Contains(msg, "failed")
		s.NotContains(msg, "caused by:")
	})

	s.Run("without branch name but with cause", func() {
		cause := NewGitCommandError("git", []string{"worktree", "add"}, 128, "", "fatal: error", "failed", nil)
		err := NewGitWorktreeError("/path/to/worktree", "", "failed", cause)
		msg := err.Error()

		s.Contains(msg, "git worktree operation failed")
		s.Contains(msg, "/path/to/worktree")
		s.NotContains(msg, "branch:")
		s.Contains(msg, "failed")
		s.Contains(msg, "caused by:")
	})

	s.Run("without branch name and without cause", func() {
		err := NewGitWorktreeError("/path/to/worktree", "", "failed", nil)
		msg := err.Error()

		s.Contains(msg, "git worktree operation failed")
		s.Contains(msg, "/path/to/worktree")
		s.NotContains(msg, "branch:")
		s.Contains(msg, "failed")
		s.NotContains(msg, "caused by:")
	})
}

func (s *ErrorsTestSuite) TestGitWorktreeError_GetCauseDetails() {
	s.Run("nil cause returns empty string", func() {
		err := &GitWorktreeError{Cause: nil}
		details := err.getCauseDetails()
		s.Empty(details)
	})

	s.Run("GitCommandError cause returns formatted error", func() {
		gitCmdErr := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed",
		}
		err := &GitWorktreeError{Cause: gitCmdErr}
		details := err.getCauseDetails()

		s.Contains(details, "git command failed")
		s.Contains(details, "git")
		s.Contains(details, "worktree add")
		s.Contains(details, "exit code 128")
	})

	s.Run("generic error cause returns error message", func() {
		genericErr := NewValidationError("request", "field", "value", "validation failed")
		err := &GitWorktreeError{Cause: genericErr}
		details := err.getCauseDetails()

		s.Contains(details, "validation failed")
		s.Contains(details, "field")
		s.Contains(details, "value")
	})
}

func (s *ErrorsTestSuite) TestGitCommandError_HasUsefulStderr() {
	s.Run("stderr with useful content returns true", func() {
		err := &GitCommandError{Stderr: "fatal: Invalid refspec"}
		s.True(err.hasUsefulStderr())
	})

	s.Run("empty stderr returns false", func() {
		err := &GitCommandError{Stderr: ""}
		s.False(err.hasUsefulStderr())
	})

	s.Run("whitespace-only stderr returns false", func() {
		err := &GitCommandError{Stderr: "   \n\t  "}
		s.False(err.hasUsefulStderr())
	})

	s.Run("stderr with mixed whitespace and content returns true", func() {
		err := &GitCommandError{Stderr: "  fatal: error\n"}
		s.True(err.hasUsefulStderr())
	})
}

func (s *ErrorsTestSuite) TestContainsOnlyWhitespace() {
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
		s.Run(tt.name, func() {
			result := containsOnlyWhitespace(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ErrorsTestSuite) TestGitCommandError_FormatErrorMessage() {
	s.Run("with useful stderr includes stderr", func() {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "fatal: Invalid refspec",
			Message:  "failed",
		}
		msg := err.Error()

		s.Contains(msg, "git command failed")
		s.Contains(msg, "git")
		s.Contains(msg, "worktree add")
		s.Contains(msg, "exit code 128")
		s.Contains(msg, "failed")
		s.Contains(msg, "stderr:")
		s.Contains(msg, "fatal: Invalid refspec")
	})

	s.Run("with empty stderr does not include stderr", func() {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "",
			Message:  "failed",
		}
		msg := err.Error()

		s.Contains(msg, "git command failed")
		s.NotContains(msg, "stderr:")
	})

	s.Run("with whitespace-only stderr does not include stderr", func() {
		err := &GitCommandError{
			Command:  "git",
			Args:     []string{"worktree", "add"},
			ExitCode: 128,
			Stderr:   "   \n\t  ",
			Message:  "failed",
		}
		msg := err.Error()

		s.Contains(msg, "git command failed")
		s.NotContains(msg, "stderr:")
	})
}

func (s *ErrorsTestSuite) TestGitWorktreeError_Unwrap() {
	s.Run("nil cause returns nil", func() {
		err := &GitWorktreeError{Cause: nil}
		s.NoError(err.Unwrap())
	})

	s.Run("returns cause error", func() {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitWorktreeError{Cause: cause}
		s.Equal(cause, err.Unwrap())
	})
}

func (s *ErrorsTestSuite) TestGitCommandError_Unwrap() {
	s.Run("nil cause returns nil", func() {
		err := &GitCommandError{Cause: nil}
		s.NoError(err.Unwrap())
	})

	s.Run("returns cause error", func() {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitCommandError{Cause: cause}
		s.Equal(cause, err.Unwrap())
	})
}

func (s *ErrorsTestSuite) TestGitRepositoryError_IsNotFound() {
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
		s.Run(tt.name, func() {
			err := NewGitRepositoryError("/path", tt.message, nil)
			s.Equal(tt.expected, err.IsNotFound())
		})
	}
}

func (s *ErrorsTestSuite) TestGitWorktreeError_IsNotFound() {
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
		s.Run(tt.name, func() {
			err := NewGitWorktreeError("/path", "branch", tt.message, nil)
			s.Equal(tt.expected, err.IsNotFound())
		})
	}
}
