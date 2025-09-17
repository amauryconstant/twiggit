package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// MiseTestDir creates a directory with mise configuration files for testing
type MiseTestDir struct {
	Path    string
	cleanup func()
}

// Cleanup removes the test directory
func (m *MiseTestDir) Cleanup() {
	if m.cleanup != nil {
		m.cleanup()
	}
}

// NewMiseTestDir creates a test directory with mise configuration
func NewMiseTestDir(t *testing.T, pattern string) *MiseTestDir {
	t.Helper()

	tempDir, cleanup := TempDir(t, pattern)

	return &MiseTestDir{
		Path:    tempDir,
		cleanup: cleanup,
	}
}

// WithLocalConfig adds a .mise.local.toml file to the directory
func (m *MiseTestDir) WithLocalConfig(t *testing.T) *MiseTestDir {
	t.Helper()

	miseFile := filepath.Join(m.Path, ".mise.local.toml")
	content := `[tools]
node = "20.0.0"
python = "3.11"

[env]
NODE_ENV = "development"
DEBUG = "true"
`
	require.NoError(t, os.WriteFile(miseFile, []byte(content), 0644))

	return m
}

// WithMiseConfig adds a mise/config.local.toml file to the directory
func (m *MiseTestDir) WithMiseConfig(t *testing.T) *MiseTestDir {
	t.Helper()

	miseDir := filepath.Join(m.Path, "mise")
	MustMkdirAll(t, miseDir, 0755)

	configFile := filepath.Join(miseDir, "config.local.toml")
	content := `[tools]
go = "1.21.0"
rust = "stable"

[env]
RUST_BACKTRACE = "1"
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	return m
}

// WithBothConfigs adds both .mise.local.toml and mise/config.local.toml
func (m *MiseTestDir) WithBothConfigs(t *testing.T) *MiseTestDir {
	t.Helper()

	return m.WithLocalConfig(t).WithMiseConfig(t)
}

// ExpectedFiles returns the list of files that should exist
func (m *MiseTestDir) ExpectedFiles() []string {
	var files []string

	localConfig := filepath.Join(m.Path, ".mise.local.toml")
	if _, err := os.Stat(localConfig); err == nil {
		files = append(files, ".mise.local.toml")
	}

	miseConfig := filepath.Join(m.Path, "mise", "config.local.toml")
	if _, err := os.Stat(miseConfig); err == nil {
		files = append(files, "mise/config.local.toml")
	}

	return files
}
