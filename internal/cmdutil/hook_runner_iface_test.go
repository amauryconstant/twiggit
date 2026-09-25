package cmdutil

import (
	"context"
	"testing"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestGitHookRunner_SatisfiesHookRunner pins the consumer-side
// interface declaration: the production *git.HookRunner type (in
// internal/git/hook_runner.go) must satisfy cmdutil.HookRunner so
// cmd/ can consume it through Factory.GitHookRunner / the run-time
// interface. The assertion lives in cmdutil because git cannot import
// cmdutil (cycle: cmdutil → git for Factory).
func TestGitHookRunner_SatisfiesHookRunner(t *testing.T) {
	t.Parallel()

	runner := git.NewHookRunner(git.NewCommandExecutor(0))

	// Compile-time + runtime interface assertion. If git.HookRunner
	// ever drifts away from the Run signature, this assignment fails
	// to compile.
	var iface HookRunner = runner

	assert.NotNil(t, iface)
	assert.Equal(t, "*git.HookRunner", typeName(iface))
}

// typeName is a tiny helper to keep the assertion message informative
// without depending on reflect in the wider codebase.
func typeName(v any) string {
	if v == nil {
		return "<nil>"
	}
	return runtimeTypeName(v)
}

// runtimeTypeName extracts the dynamic type name via a type switch on
// the concrete *git.HookRunner shape we just assigned. Avoids
// importing reflect for one assertion.
func runtimeTypeName(v any) string {
	switch v.(type) {
	case *git.HookRunner:
		return "*git.HookRunner"
	default:
		return "<unknown>"
	}
}
