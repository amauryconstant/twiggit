package cmdutil_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
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
		f := cmdutil.NewFactory()
		assert.NoError(t, f.Init())
	})

	t.Run("config failure surfaces", func(t *testing.T) {
		f := cmdutil.NewFactory()
		f.Config = func() (*core.Config, error) { return nil, errors.New("config broken") }

		err := f.Init()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config broken")
	})
}
