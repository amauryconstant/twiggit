package cmdutil_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
)

// TestNewFactory_ReturnsNonNilWithSystemIOStreams covers the happy
// construction path: every eager field populated, every lazy field
// non-nil so the first call sites do not panic.
func TestNewFactory_ReturnsNonNilWithSystemIOStreams(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	require.NotNil(t, f)
	require.NotNil(t, f.IOStreams)
	require.NotNil(t, f.Config)
	require.NotNil(t, f.GitClient)
	require.NotNil(t, f.Logger)
	assert.NotEmpty(t, f.AppVersion)
	assert.NotEmpty(t, f.Executable)
}

// TestFactory_ConfigIsCachedAcrossCalls covers the sync.OnceValue
// guarantee: two calls share the same pointer so consumers can rely on
// the cache rather than re-reading the config file.
func TestFactory_ConfigIsCachedAcrossCalls(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	first, err := f.Config()
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := f.Config()
	require.NoError(t, err)
	assert.Same(t, first, second, "config must be cached across calls")
}

// TestFactory_LoggerIsCachedAcrossCalls confirms the Logger field also
// returns the same slog.Logger pointer on repeated calls.
func TestFactory_LoggerIsCachedAcrossCalls(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	first := f.Logger()
	second := f.Logger()

	require.NotNil(t, first)
	assert.Same(t, first, second)
	assert.IsType(t, &slog.Logger{}, first)
}

// TestFactory_LazyFieldsAreReplaceable covers the test seam: callers
// can swap a field with a stub before the first invocation. The
// Factory must respect the swapped function and never touch the
// underlying sync.OnceValue again.
func TestFactory_LazyFieldsAreReplaceable(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	// Swap Config with a deterministic stub before the first call.
	stubCfg := &core.Config{DefaultSourceBranch: "stub"}
	f.Config = func() (*core.Config, error) { return stubCfg, nil }

	got, err := f.Config()
	require.NoError(t, err)
	assert.Same(t, stubCfg, got)
}

// TestFactory_InitTouchesEveryLazyField covers the "fail fast at
// startup" contract: Init returns nil when the underlying singletons
// can be constructed, joining all errors otherwise.
func TestFactory_InitTouchesEveryLazyField(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		f := cmdutil.NewFactory()
		assert.NoError(t, f.Init())
	})

	t.Run("config failure surfaces", func(t *testing.T) {
		t.Parallel()
		f := cmdutil.NewFactory()
		f.Config = func() (*core.Config, error) { return nil, errors.New("config broken") }

		err := f.Init()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config broken")
	})
}

// TestFactory_PerRoleFieldsReturnSameCompositeInstance pins the
// "per-role fields re-use the composite's cached concrete" contract
// from core-git R11.S1 / git-client R2.S2. Every per-role lazy field
// must return the same underlying *git.Client pointer as f.GitClient().
func TestFactory_PerRoleFieldsReturnSameCompositeInstance(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	require.NoError(t, f.Init())

	gc, err := f.GitClient()
	require.NoError(t, err)
	require.NotNil(t, gc)

	ro, err := f.RepoOpener()
	require.NoError(t, err)
	br, err := f.BranchReader()
	require.NoError(t, err)
	rr, err := f.RepositoryReader()
	require.NoError(t, err)
	rm, err := f.RemoteReader()
	require.NoError(t, err)
	wt, err := f.WorktreeWriter()
	require.NoError(t, err)
	bw, err := f.BranchWriter()
	require.NoError(t, err)

	assert.Same(t, gc, ro, "RepoOpener must return same instance as GitClient")
	assert.Same(t, gc, br, "BranchReader must return same instance as GitClient")
	assert.Same(t, gc, rr, "RepositoryReader must return same instance as GitClient")
	assert.Same(t, gc, rm, "RemoteReader must return same instance as GitClient")
	assert.Same(t, gc, wt, "WorktreeWriter must return same instance as GitClient")
	assert.Same(t, gc, bw, "BranchWriter must return same instance as GitClient")
}

// TestFactory_InitTouchesEveryPerRoleField covers core-git R11.S3:
// Init must invoke each per-role field so construction failures
// surface during startup rather than at first lazy access.
func TestFactory_InitTouchesEveryPerRoleField(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	sentinel := errors.New("branch reader init sentinel")
	f.BranchReader = func() (core.BranchReader, error) { return nil, sentinel }

	err := f.Init()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel,
		"Init must touch BranchReader and surface its error via errors.Join")
}

// TestFactory_PerRoleErrorPropagation covers core-git R12.S1:
// when the underlying GitClient returns a *git.ExternalError, the
// per-role field must propagate the error unchanged — no wrap, no
// log, no mutation. The single-handling rule requires the error to
// reach the command layer once for the boundary formatter.
func TestFactory_PerRoleErrorPropagation(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	sentinel := &git.ExternalError{
		Tool:      "git",
		Operation: "open",
		Message:   "boom",
		Cause:     nil,
		OperationError: &core.OperationError{
			Op:      "git.repository",
			Message: "boom",
		},
	}
	f.GitClient = func() (*git.Client, error) { return nil, sentinel }

	_, err := f.RepoOpener()
	require.Error(t, err)

	var oe *core.OperationError
	require.ErrorAs(t, err, &oe,
		"errors.As must walk to embedded *core.OperationError")
	assert.Equal(t, "git.repository", oe.Op,
		"propagated error must preserve the original Op value")
	assert.Same(t, sentinel, err,
		"Factory must not wrap the GitClient error")
}
