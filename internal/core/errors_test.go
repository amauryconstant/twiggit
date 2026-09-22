package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitWorktreeError_FormatErrorMessage(t *testing.T) {
	t.Run("with branch name and cause", func(t *testing.T) {
		cause := errors.New("fatal: Invalid refspec")
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", cause)
		msg := err.Error()

		assert.Contains(t, msg, "git.worktree")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.Contains(t, msg, "branch:feature-branch")
		assert.Contains(t, msg, "failed")
		assert.Contains(t, msg, "fatal: Invalid refspec")
	})

	t.Run("without cause", func(t *testing.T) {
		err := NewGitWorktreeError("/path/to/worktree", "feature-branch", "failed", nil)
		msg := err.Error()

		assert.Contains(t, msg, "git.worktree")
		assert.Contains(t, msg, "/path/to/worktree")
		assert.Contains(t, msg, "branch:feature-branch")
		assert.NotContains(t, msg, "caused by:")
	})
}

func TestGitCommandError_FormatErrorMessage(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		err := NewGitCommandError("git", []string{"worktree", "add"}, 128, "", "fatal: error", "failed", nil)
		msg := err.Error()

		assert.Contains(t, msg, "git.command")
		assert.Contains(t, msg, "git")
		assert.Contains(t, msg, "worktree add")
		assert.Contains(t, msg, "exit 128")
		assert.Contains(t, msg, "failed")
	})
}

func TestGitWorktreeError_SentinelWalk(t *testing.T) {
	err := NewGitWorktreeError("/wt", "b", "missing", nil)
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
	assert.NotErrorIs(t, err, ErrGitRepoNotFound)
	assert.NotErrorIs(t, err, ErrProjectNotFound)
}

func TestGitRepositoryError_SentinelWalk(t *testing.T) {
	err := NewGitRepositoryError("/p", "msg", nil)
	assert.ErrorIs(t, err, ErrGitRepoNotFound)
	assert.NotErrorIs(t, err, ErrWorktreeNotFound)
}

func TestOperationError_Unwrap(t *testing.T) {
	cause := errors.New("underlying")
	op := NewGitRepositoryError("/p", "msg", cause)
	assert.Equal(t, cause, op.Unwrap())
}

// TestErrSentinels_WalkThroughWraps asserts each demoted constructor
// walks to its canonical NotFound sentinel via errors.Is.
func TestErrSentinels_WalkThroughWraps(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		ownSentinel   error
		otherA        error
		otherB        error
	}{
		{
			name:        "NewGitRepositoryError → ErrGitRepoNotFound",
			err:         NewGitRepositoryError("/p", "msg", nil),
			ownSentinel: ErrGitRepoNotFound,
			otherA:      ErrWorktreeNotFound,
			otherB:      ErrProjectNotFound,
		},
		{
			name:        "NewGitWorktreeError → ErrWorktreeNotFound",
			err:         NewGitWorktreeError("/p", "b", "msg", nil),
			ownSentinel: ErrWorktreeNotFound,
			otherA:      ErrGitRepoNotFound,
			otherB:      ErrProjectNotFound,
		},
		{
			name:        "NewWorktreeServiceError → ErrWorktreeNotFound",
			err:         NewWorktreeServiceError("/p", "b", "op", "msg", nil),
			ownSentinel: ErrWorktreeNotFound,
			otherA:      ErrProjectNotFound,
			otherB:      ErrResolutionNotFound,
		},
		{
			name:        "NewProjectServiceError → ErrProjectNotFound",
			err:         NewProjectServiceError("name", "/p", "op", "msg", nil),
			ownSentinel: ErrProjectNotFound,
			otherA:      ErrWorktreeNotFound,
			otherB:      ErrResolutionNotFound,
		},
		{
			name:        "NewNavigationServiceError → ErrResolutionNotFound",
			err:         NewNavigationServiceError("t", "ctx", "op", "msg", nil),
			ownSentinel: ErrResolutionNotFound,
			otherA:      ErrWorktreeNotFound,
			otherB:      ErrProjectNotFound,
		},
		{
			name:        "NewResolutionError → ErrResolutionNotFound",
			err:         NewResolutionError("t", "ctx", "msg", nil, nil),
			ownSentinel: ErrResolutionNotFound,
			otherA:      ErrWorktreeNotFound,
			otherB:      ErrProjectNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.err, tt.ownSentinel, "should match own sentinel")
			assert.NotErrorIs(t, tt.err, tt.otherA, "should not match unrelated sentinel")
			assert.NotErrorIs(t, tt.err, tt.otherB, "should not match unrelated sentinel")
		})
	}
}
