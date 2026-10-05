package cmd

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestFactory assembles a minimal *cmdutil.Factory pointing at
// scratch temp dirs. Cached lazy fields (Config, GitClient) are
// pinned so the cobra RunE path can read them. GlobalOptions is
// left nil so NewRootCommand installs a fresh struct (which the
// flag-bound persistent flags mutate in place).
func newTestFactory(t *testing.T) *cmdutil.Factory {
	t.Helper()

	ios, _, _, _ := iostreams.Test()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()
	f := &cmdutil.Factory{
		IOStreams:  ios,
		Context:    t.Context(),
		AppVersion: "test",
		Executable: "twiggit-test",
	}
	f.Config = func() (*core.Config, error) { return cfg, nil }
	gitClient, err := git.NewClient()
	require.NoError(t, err)
	f.GitClient = func() (cmdutil.Client, error) { return gitClient, nil }
	f.Logger = func() *slog.Logger {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return f
}

// initTestOpts constructs InitOptions with the supplied shell type,
// install-mode flag, and config file path. Returns the options and
// the iostreams.Test IOS so tests can read ios.Stdout().
func initTestOpts(t *testing.T, shell core.ShellType, install, force bool, configFile string) (*InitOptions, *iostreams.IOStreams) {
	t.Helper()

	ios, _, _, _ := iostreams.Test()

	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = t.TempDir()
	cfg.WorktreesDirectory = t.TempDir()

	gitClient, err := git.NewClient()
	require.NoError(t, err)

	opts := &InitOptions{
		IO:            ios,
		Config:        func() (*core.Config, error) { return cfg, nil },
		GitClient:     func() (cmdutil.Client, error) { return gitClient, nil },
		Ctx:           t.Context(),
		GlobalOptions: &cmdutil.GlobalOptions{},
		ShellType:     shell,
		IsInstall:     install,
		IsForce:       force,
		ConfigFile:    configFile,
	}

	return opts, ios
}

// iosStdout extracts the stdout buffer content from an iostreams.Test
// IOS. Returns empty string if the buffer isn't a *bytes.Buffer.
func iosStdout(t *testing.T, ios *iostreams.IOStreams) string {
	t.Helper()
	buf, ok := ios.Out.(*bytes.Buffer)
	if !ok {
		return ""
	}
	return buf.String()
}

// TestInit_StdoutPrintsWrapper pins the default (no --install) flow:
// runInit writes the bash wrapper to stdout so `eval` works. Test
// asserts a known wrapper substring is present.
func TestInit_StdoutPrintsWrapper(t *testing.T) {
	opts, ios := initTestOpts(t, core.ShellBash, false, false, "")

	require.NoError(t, runInit(opts))

	written := iosStdout(t, ios)
	assert.NotEmpty(t, written, "init must write the wrapper to stdout")
	// The bash wrapper always opens with a function declaration.
	assert.Contains(t, written, "twiggit()")
}

// TestInit_InvalidShellTypeReturnsValidationError confirms runInit
// surfaces a *core.ValidationError when the supplied shell type is
// not in the supported set, rather than silently passing it through.
func TestInit_InvalidShellTypeReturnsValidationError(t *testing.T) {
	opts, _ := initTestOpts(t, core.ShellType("powershell"), false, false, "")

	err := runInit(opts)
	require.Error(t, err)

	var ve *core.ValidationError
	require.ErrorAs(t, err, &ve, "invalid shell type must yield *core.ValidationError")
	assert.Equal(t, "shellType", ve.Field)
	assert.Equal(t, string(core.ShellType("powershell")), ve.Value)
}

// TestNewCmdInit_ConfigRequiresInstall pins the cobra RunE guard for
// `--config` without `--install`. The check lives in NewCmdInit,
// not in runInit, so this test exercises the full cobra pipeline.
func TestNewCmdInit_ConfigRequiresInstall(t *testing.T) {
	f := newTestFactory(t)
	cmd := NewCmdInit(f, nil)
	cmd.SetArgs([]string{"bash", "--config", "/tmp/bashrc"})

	err := cmd.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "--config without --install must yield *core.UsageError")
	assert.Contains(t, ue.Message, "--config")
}

// TestNewCmdInit_ForceRequiresInstall pins the symmetric cobra guard
// for `--force` without `--install`.
func TestNewCmdInit_ForceRequiresInstall(t *testing.T) {
	f := newTestFactory(t)
	cmd := NewCmdInit(f, nil)
	cmd.SetArgs([]string{"bash", "--force"})

	err := cmd.Execute()
	require.Error(t, err)
	var ue *core.UsageError
	require.ErrorAs(t, err, &ue, "--force without --install must yield *core.UsageError")
	assert.Contains(t, ue.Message, "--force")
}

// keep strings import referenced for future tests without extra noise.
var _ = strings.Contains
