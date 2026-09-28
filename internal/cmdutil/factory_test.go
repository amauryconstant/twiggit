package cmdutil_test

import (
	"errors"
	"log/slog"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubGitClient returns a *git.Client the Factory can cache. Tests that
// need a wired Factory but do not exercise real git behaviour pass
// this through WithGitClientFactory.
func stubGitClient() *git.Client {
	return &git.Client{}
}

// TestNewFactory_ReturnsNonNilWithSystemIOStreams covers the happy
// construction path: every eager field populated, every lazy field
// non-nil so the first call sites do not panic. AppVersion defaults
// empty when no WithVersion option is supplied; the wired variant is
// covered by TestNewFactory_WithVersionSetsAppVersion.
func TestNewFactory_ReturnsNonNilWithSystemIOStreams(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	require.NotNil(t, f)
	require.NotNil(t, f.IOStreams)
	require.NotNil(t, f.Config)
	require.NotNil(t, f.GitClient)
	require.NotNil(t, f.Logger)
	assert.Empty(t, f.AppVersion, "AppVersion defaults empty without WithVersion")
	assert.NotEmpty(t, f.Executable)
}

// TestNewFactory_WithVersionSetsAppVersion covers the WithVersion
// functional option. main.go uses this to wire version.Version into
// the factory at composition time.
func TestNewFactory_WithVersionSetsAppVersion(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory(cmdutil.WithVersion("test"))

	assert.Equal(t, "test", f.AppVersion)
}

// TestFactory_DefaultLazyFieldsReturnSentinels covers the no-arg
// construction path: lazy fields return the package sentinels so
// Init() surfaces a joined error before any command body runs.
// main.go MUST wire WithConfigLoader and WithGitClientFactory or
// every command will hit a sentinel on first lazy access.
func TestFactory_DefaultLazyFieldsReturnSentinels(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	_, err := f.Config()
	assert.ErrorIs(t, err, cmdutil.ErrNoConfigLoader)

	_, err = f.GitClient()
	assert.ErrorIs(t, err, cmdutil.ErrNoGitClient)
}

// TestFactory_DefaultInitJoinsSentinels covers Init's contract when
// no options are wired: every lazy field surfaces its sentinel and
// Init returns a joined error so callers see the full diagnostic in
// one pass.
func TestFactory_DefaultInitJoinsSentinels(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()

	err := f.Init()
	require.Error(t, err)
	assert.ErrorIs(t, err, cmdutil.ErrNoConfigLoader)
	assert.ErrorIs(t, err, cmdutil.ErrNoGitClient)
}

// TestFactory_ConfigIsCachedAcrossCalls covers the sync.OnceValue
// guarantee under the wired path: WithConfigLoader wraps the supplied
// constructor once, and two calls share the same pointer.
func TestFactory_ConfigIsCachedAcrossCalls(t *testing.T) {
	t.Parallel()

	var calls int
	stubCfg := &core.Config{DefaultSourceBranch: "stub"}
	f := cmdutil.NewFactory(cmdutil.WithConfigLoader(func() (*core.Config, error) {
		calls++
		return stubCfg, nil
	}))

	first, err := f.Config()
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := f.Config()
	require.NoError(t, err)
	assert.Same(t, first, second, "config must be cached across calls")
	assert.Equal(t, 1, calls, "loader must run exactly once across two calls")
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

	stubCfg := &core.Config{DefaultSourceBranch: "stub"}
	f.Config = func() (*core.Config, error) { return stubCfg, nil }

	got, err := f.Config()
	require.NoError(t, err)
	assert.Same(t, stubCfg, got)
}

// TestFactory_InitTouchesEveryLazyField covers the "fail fast at
// startup" contract under the wired path. The default-path sentinel
// behaviour is covered by TestFactory_DefaultInitJoinsSentinels.
func TestFactory_InitTouchesEveryLazyField(t *testing.T) {
	t.Parallel()

	t.Run("happy path with wiring", func(t *testing.T) {
		t.Parallel()
		f := cmdutil.NewFactory(
			cmdutil.WithConfigLoader(func() (*core.Config, error) {
				return &core.Config{}, nil
			}),
			cmdutil.WithGitClientFactory(func() (cmdutil.Client, error) {
				return stubGitClient(), nil
			}),
		)
		assert.NoError(t, f.Init())
	})

	t.Run("config failure surfaces", func(t *testing.T) {
		t.Parallel()
		f := cmdutil.NewFactory(
			cmdutil.WithConfigLoader(func() (*core.Config, error) { return nil, errors.New("config broken") }),
			cmdutil.WithGitClientFactory(func() (cmdutil.Client, error) { return stubGitClient(), nil }),
		)

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

	stub := stubGitClient()
	f := cmdutil.NewFactory(
		cmdutil.WithConfigLoader(func() (*core.Config, error) { return &core.Config{}, nil }),
		cmdutil.WithGitClientFactory(func() (cmdutil.Client, error) { return stub, nil }),
	)
	require.NoError(t, f.Init())

	gc, err := f.GitClient()
	require.NoError(t, err)
	require.NotNil(t, gc)
	gcConcrete, ok := gc.(*git.Client)
	require.True(t, ok, "GitClient must be *git.Client under the wired path")

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

	assert.Same(t, gcConcrete, ro, "RepoOpener must return same instance as GitClient")
	assert.Same(t, gcConcrete, br, "BranchReader must return same instance as GitClient")
	assert.Same(t, gcConcrete, rr, "RepositoryReader must return same instance as GitClient")
	assert.Same(t, gcConcrete, rm, "RemoteReader must return same instance as GitClient")
	assert.Same(t, gcConcrete, wt, "WorktreeWriter must return same instance as GitClient")
	assert.Same(t, gcConcrete, bw, "BranchWriter must return same instance as GitClient")
}

// TestFactory_InitTouchesEveryPerRoleField covers core-git R11.S3:
// Init must invoke each per-role field so construction failures
// surface during startup rather than at first lazy access.
func TestFactory_InitTouchesEveryPerRoleField(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory(
		cmdutil.WithConfigLoader(func() (*core.Config, error) { return &core.Config{}, nil }),
		cmdutil.WithGitClientFactory(func() (cmdutil.Client, error) { return stubGitClient(), nil }),
	)

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

	f := cmdutil.NewFactory(
		cmdutil.WithConfigLoader(func() (*core.Config, error) { return &core.Config{}, nil }),
	)

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
	f.GitClient = func() (cmdutil.Client, error) { return nil, sentinel }

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
