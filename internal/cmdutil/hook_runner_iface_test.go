package cmdutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
)

// HookRunnerFunc adapts a plain function to the HookRunner interface.
// Useful for tests and for one-off inline runners in command factories.
type HookRunnerFunc func(ctx context.Context, req *core.HookRunRequest) (*core.HookResult, error)

// Run implements HookRunner.
func (f HookRunnerFunc) Run(ctx context.Context, req *core.HookRunRequest) (*core.HookResult, error) {
	return f(ctx, req)
}

// Compile-time guarantee that HookRunnerFunc satisfies HookRunner.
var _ HookRunner = HookRunnerFunc(nil)

func TestHookRunnerFunc_ImplementsInterface(t *testing.T) {
	t.Parallel()

	want := &core.HookResult{
		HookType:     core.HookPostCreate,
		HasExecuted:  true,
		IsSuccessful: true,
	}

	var runner HookRunner = HookRunnerFunc(func(_ context.Context, req *core.HookRunRequest) (*core.HookResult, error) {
		// Echo the request fields through to confirm the impl saw them.
		assert.Equal(t, core.HookPostCreate, req.HookType)
		assert.Equal(t, "/tmp/worktrees/feature", req.WorktreePath)
		assert.Equal(t, "twiggit", req.ProjectName)
		assert.Equal(t, "feature", req.BranchName)
		assert.Equal(t, "main", req.SourceBranch)
		assert.Equal(t, "/tmp/repo", req.MainRepoPath)
		assert.Equal(t, "/tmp/repo/.twiggit.yaml", req.ConfigFilePath)
		return want, nil
	})

	req := &core.HookRunRequest{
		HookType:       core.HookPostCreate,
		WorktreePath:   "/tmp/worktrees/feature",
		ProjectName:    "twiggit",
		BranchName:     "feature",
		SourceBranch:   "main",
		MainRepoPath:   "/tmp/repo",
		ConfigFilePath: "/tmp/repo/.twiggit.yaml",
	}

	result, err := runner.Run(t.Context(), req)
	require.NoError(t, err)
	require.Same(t, want, result)
	assert.Equal(t, core.HookPostCreate, result.HookType)
	assert.True(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
}

func TestHookRunnerFunc_NoOpReturnsNilResult(t *testing.T) {
	t.Parallel()

	runner := HookRunnerFunc(func(_ context.Context, _ *core.HookRunRequest) (*core.HookResult, error) {
		return nil, nil
	})

	result, err := runner.Run(t.Context(), &core.HookRunRequest{HookType: core.HookPostCreate})
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestHookRunnerFunc_PropagatesError(t *testing.T) {
	t.Parallel()

	boom := assert.AnError
	runner := HookRunnerFunc(func(_ context.Context, _ *core.HookRunRequest) (*core.HookResult, error) {
		return nil, boom
	})

	result, err := runner.Run(t.Context(), &core.HookRunRequest{})
	require.ErrorIs(t, err, boom)
	assert.Nil(t, result)
}
