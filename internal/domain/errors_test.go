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
		err := &GitWorktreeError{Err: nil}
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
		err := &GitWorktreeError{Err: gitCmdErr}
		details := err.getCauseDetails()

		assert.Contains(t, details, "git command failed")
		assert.Contains(t, details, "git")
		assert.Contains(t, details, "worktree add")
		assert.Contains(t, details, "exit code 128")
	})

	t.Run("generic error cause returns error message", func(t *testing.T) {
		genericErr := NewValidationError("request", "field", "value", "validation failed")
		err := &GitWorktreeError{Err: genericErr}
		details := err.getCauseDetails()

		assert.Contains(t, details, "validation failed")
		assert.Contains(t, details, "field")
		assert.Contains(t, details, "value")
	})
}

func TestGitCommandError_HasUsefulStderr(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected bool
	}{
		{"stderr with useful content", "fatal: Invalid refspec", true},
		{"empty stderr", "", false},
		{"whitespace-only stderr", "   \n\t  ", false},
		{"stderr with mixed whitespace and content", "  fatal: error\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &GitCommandError{Stderr: tt.stderr}
			assert.Equal(t, tt.expected, err.hasUsefulStderr())
		})
	}
}

func TestContainsOnlyWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"spaces only", "    ", true},
		{"tabs only", "\t\t\t", true},
		{"newlines only", "\n\n\n", true},
		{"mixed whitespace", " \t\n \t ", true},
		{"string with content", "hello", false},
		{"string with content and whitespace", "  hello world  ", false},
		{"string with special characters", "!@#$%", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, containsOnlyWhitespace(tt.input))
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
		err := &GitWorktreeError{Err: nil}
		assert.NoError(t, err.Unwrap())
	})

	t.Run("returns cause error", func(t *testing.T) {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitWorktreeError{Err: cause}
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestGitCommandError_Unwrap(t *testing.T) {
	t.Run("nil cause returns nil", func(t *testing.T) {
		err := &GitCommandError{Err: nil}
		assert.NoError(t, err.Unwrap())
	})

	t.Run("returns cause error", func(t *testing.T) {
		cause := NewValidationError("request", "field", "value", "error")
		err := &GitCommandError{Err: cause}
		assert.Equal(t, cause, err.Unwrap())
	})
}

// TestErrSentinels_WalkThroughWraps asserts that each NotFound sentinel
// is reachable via errors.Is walks through its corresponding wrapper
// type, and that unrelated sentinels do not match.
func TestErrSentinels_WalkThroughWraps(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		ownSentinel    error
		otherSentinel  error
		otherSentinel2 error
	}{
		{
			name:           "GitRepositoryError matches ErrGitRepoNotFound only",
			err:            NewGitRepositoryError("/p", "msg", nil),
			ownSentinel:    ErrGitRepoNotFound,
			otherSentinel:  ErrWorktreeNotFound,
			otherSentinel2: ErrProjectNotFound,
		},
		{
			name:           "GitWorktreeError matches ErrWorktreeNotFound only",
			err:            NewGitWorktreeError("/p", "b", "msg", nil),
			ownSentinel:    ErrWorktreeNotFound,
			otherSentinel:  ErrGitRepoNotFound,
			otherSentinel2: ErrProjectNotFound,
		},
		{
			name:           "WorktreeServiceError matches ErrWorktreeNotFound only",
			err:            NewWorktreeServiceError("/p", "b", "op", "msg", nil),
			ownSentinel:    ErrWorktreeNotFound,
			otherSentinel:  ErrProjectNotFound,
			otherSentinel2: ErrResolutionNotFound,
		},
		{
			name:           "ProjectServiceError matches ErrProjectNotFound only",
			err:            NewProjectServiceError("name", "/p", "op", "msg", nil),
			ownSentinel:    ErrProjectNotFound,
			otherSentinel:  ErrWorktreeNotFound,
			otherSentinel2: ErrResolutionNotFound,
		},
		{
			name:           "NavigationServiceError matches ErrResolutionNotFound only",
			err:            NewNavigationServiceError("t", "ctx", "op", "msg", nil),
			ownSentinel:    ErrResolutionNotFound,
			otherSentinel:  ErrWorktreeNotFound,
			otherSentinel2: ErrProjectNotFound,
		},
		{
			name:           "ResolutionError matches ErrResolutionNotFound only",
			err:            NewResolutionError("t", "ctx", "msg", nil, nil),
			ownSentinel:    ErrResolutionNotFound,
			otherSentinel:  ErrWorktreeNotFound,
			otherSentinel2: ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.err, tt.ownSentinel, "should match own sentinel")
			assert.NotErrorIs(t, tt.err, tt.otherSentinel, "should not match unrelated sentinel")
			assert.NotErrorIs(t, tt.err, tt.otherSentinel2, "should not match unrelated sentinel")
		})
	}
}
