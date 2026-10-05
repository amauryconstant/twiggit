package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResultOr_Success(t *testing.T) {
	is := assert.New(t)
	r := NewResultOr("value", nil)
	is.Equal("value", r.Value)
	is.True(r.Success, "Success must be true when err is nil")
	is.True(r.IsSuccess())
	is.False(r.IsError())
	is.NoError(r.Error)
}

func TestNewResultOr_Failure(t *testing.T) {
	is := assert.New(t)
	boom := errors.New("boom")
	r := NewResultOr("value", boom)
	is.Equal("value", r.Value)
	is.False(r.Success, "Success must be false when err is non-nil")
	is.False(r.IsSuccess())
	is.True(r.IsError())
	is.ErrorIs(r.Error, boom)
}

func TestNewResultOr_DefendsAgainstAliasing(t *testing.T) {
	is := assert.New(t)
	src := []string{"a", "b"}
	r := NewResultOr(src, nil)
	r.Value[0] = "MUTATED"
	is.Equal("a", src[0], "slice backing array must be cloned at construction")
}

func TestResult_ZeroIsFailure(t *testing.T) {
	is := assert.New(t)
	var r Result[int]
	is.False(r.IsSuccess())
	is.True(r.IsError())
	is.Equal(0, r.Value)
	is.NoError(r.Error)
}

func TestNotFoundError_Is_EntityMembershipTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		entity     string
		wantSentin []error
		wantNot    []error
	}{
		{
			"project entity → ErrProjectNotFound", "project",
			[]error{ErrProjectNotFound},
			[]error{ErrWorktreeNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"worktree entity → ErrWorktreeNotFound", "worktree",
			[]error{ErrWorktreeNotFound},
			[]error{ErrProjectNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"resolution entity → ErrResolutionNotFound", "resolution",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"navigation entity → ErrResolutionNotFound", "navigation",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"git.repository entity → ErrGitRepoNotFound", "git.repository",
			[]error{ErrGitRepoNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound},
		},
		{
			"repo entity → ErrGitRepoNotFound", "repo",
			[]error{ErrGitRepoNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound},
		},
		{
			"unknown entity matches no sentinel", "alien",
			nil,
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := &NotFoundError{Entity: tt.entity, Name: "x"}
			for _, s := range tt.wantSentin {
				assert.ErrorIs(t, e, s, "entity=%q must walk to %s", tt.entity, sentinelName(s))
			}
			for _, s := range tt.wantNot {
				assert.NotErrorIs(t, e, s, "entity=%q must NOT walk to %s", tt.entity, sentinelName(s))
			}
		})
	}
}

func TestOperationError_Is_ExtendedPrefixCoverage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		op         string
		wantSentin []error
		wantNot    []error
	}{
		{
			"project.op → ErrProjectNotFound", "project.discover",
			[]error{ErrProjectNotFound},
			[]error{ErrWorktreeNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"project (bare) → ErrProjectNotFound", "project",
			[]error{ErrProjectNotFound},
			[]error{ErrWorktreeNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"worktree.op → ErrWorktreeNotFound", "worktree.delete",
			[]error{ErrWorktreeNotFound},
			[]error{ErrProjectNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"worktree (bare) → ErrWorktreeNotFound", "worktree",
			[]error{ErrWorktreeNotFound},
			[]error{ErrProjectNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"git.worktree.* still matches ErrWorktreeNotFound", "git.worktree.create",
			[]error{ErrWorktreeNotFound},
			[]error{ErrProjectNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
		{
			"navigation.* → ErrResolutionNotFound", "navigation.list",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"navigation (bare) → ErrResolutionNotFound", "navigation",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"resolution.* → ErrResolutionNotFound", "resolution.find",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"resolution (bare) → ErrResolutionNotFound", "resolution",
			[]error{ErrResolutionNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrGitRepoNotFound},
		},
		{
			"git.repository.* → ErrGitRepoNotFound", "git.repository.open",
			[]error{ErrGitRepoNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound},
		},
		{
			"git.repo.* → ErrGitRepoNotFound", "git.repo.clone",
			[]error{ErrGitRepoNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound},
		},
		{
			"repo (bare) → ErrGitRepoNotFound", "repo",
			[]error{ErrGitRepoNotFound},
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound},
		},
		{
			"unrelated op matches nothing", "shell.wrapper.installation",
			nil,
			[]error{ErrProjectNotFound, ErrWorktreeNotFound, ErrResolutionNotFound, ErrGitRepoNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := &OperationError{Op: tt.op, Message: "msg"}
			for _, s := range tt.wantSentin {
				assert.ErrorIs(t, e, s, "op=%q must walk to %s", tt.op, sentinelName(s))
			}
			for _, s := range tt.wantNot {
				assert.NotErrorIs(t, e, s, "op=%q must NOT walk to %s", tt.op, sentinelName(s))
			}
		})
	}
}

func TestOperationError_Is_WalksCauseFirst(t *testing.T) {
	t.Parallel()
	t.Run("Cause containing sentinel reaches it", func(t *testing.T) {
		t.Parallel()
		wrapped := errors.New("context probe failed")
		e := &OperationError{Op: "context.detect", Cause: wrapped}
		assert.ErrorIs(t, e, wrapped, "Op=Cause must walk first regardless of Op prefix")
	})
}

func TestUsageError_Unwrap_WalksToCause(t *testing.T) {
	t.Parallel()
	t.Run("nil cause returns nil from Unwrap", func(t *testing.T) {
		t.Parallel()
		u := NewUsageError("missing flag", nil)
		assert.NoError(t, u.Unwrap())
	})
	t.Run("non-nil cause reaches the sentinel", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("parser: bad arg")
		u := NewUsageError("wrapped", sentinel)
		assert.ErrorIs(t, u, sentinel, "errors.Is must walk through hidden cause")
	})
	t.Run("UsageWrap preserves cause chain", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("cobra: --foo")
		wrapped := UsageWrap(sentinel)
		require.Error(t, wrapped)
		var target *UsageError
		require.ErrorAs(t, wrapped, &target)
		assert.ErrorIs(t, target, sentinel)
	})
}

func TestValidationError_Unwrap_WalksToCause(t *testing.T) {
	t.Parallel()
	t.Run("nil cause returns nil", func(t *testing.T) {
		t.Parallel()
		v := NewValidationError("branch", "x", "msg")
		assert.NoError(t, v.Unwrap())
	})
	t.Run("cause field set by NewWorktreeServiceError", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("disk full")
		v := NewWorktreeServiceError("/p", "b", "op", "msg", sentinel)
		assert.ErrorIs(t, v, sentinel, "errors.Is must walk through the hidden cause")
	})
}

// sentinelName returns a stable identifier for the sentinel so test
// failure messages stay readable.
func sentinelName(s error) string {
	switch {
	case errors.Is(s, ErrProjectNotFound):
		return "ErrProjectNotFound"
	case errors.Is(s, ErrWorktreeNotFound):
		return "ErrWorktreeNotFound"
	case errors.Is(s, ErrResolutionNotFound):
		return "ErrResolutionNotFound"
	case errors.Is(s, ErrGitRepoNotFound):
		return "ErrGitRepoNotFound"
	default:
		return "unknown"
	}
}
