package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_NewConfig(t *testing.T) {
	config := NewConfig()
	require.NotNil(t, config)

	// Test default values
	assert.Equal(t, filepath.Join(os.Getenv("HOME"), "Workspaces"), config.Workspace)
	assert.Equal(t, "", config.Project)
	assert.False(t, config.Verbose)
	assert.False(t, config.Quiet)
}

func TestConfig_LoadFromEnvironment(t *testing.T) {
	// Set environment variables
	require.NoError(t, os.Setenv("TWIGGIT_WORKSPACE", "/custom/workspace"))
	require.NoError(t, os.Setenv("TWIGGIT_PROJECT", "my-project"))
	require.NoError(t, os.Setenv("TWIGGIT_VERBOSE", "true"))
	require.NoError(t, os.Setenv("TWIGGIT_QUIET", "false"))

	defer func() {
		_ = os.Unsetenv("TWIGGIT_WORKSPACE")
		_ = os.Unsetenv("TWIGGIT_PROJECT")
		_ = os.Unsetenv("TWIGGIT_VERBOSE")
		_ = os.Unsetenv("TWIGGIT_QUIET")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "/custom/workspace", config.Workspace)
	assert.Equal(t, "my-project", config.Project)
	assert.True(t, config.Verbose)
	assert.False(t, config.Quiet)
}

func TestConfig_LoadFromFile(t *testing.T) {
	// Create temporary config file in the expected XDG structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	configFile := filepath.Join(configDir, "config.yaml")

	configContent := `workspace: /file/workspace
project: file-project
verbose: true
quiet: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", tempDir))
	defer func() {
		if originalXDG == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "/file/workspace", config.Workspace)
	assert.Equal(t, "file-project", config.Project)
	assert.True(t, config.Verbose)
	assert.False(t, config.Quiet)
}

func TestConfig_EnvironmentOverridesFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "twiggit", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configFile), 0755))

	configContent := `
workspace: /file/workspace
project: file-project
verbose: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override file values
	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", tempDir))
	require.NoError(t, os.Setenv("TWIGGIT_WORKSPACE", "/env/workspace"))
	require.NoError(t, os.Setenv("TWIGGIT_VERBOSE", "true"))

	defer func() {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("TWIGGIT_WORKSPACE")
		_ = os.Unsetenv("TWIGGIT_VERBOSE")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	// Environment should override file
	assert.Equal(t, "/env/workspace", config.Workspace)
	assert.True(t, config.Verbose)
	// File value should be used where env not set
	assert.Equal(t, "file-project", config.Project)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Workspace: "/valid/path",
				Project:   "valid-project",
			},
			expectError: false,
		},
		{
			name: "empty workspace",
			config: &Config{
				Workspace: "",
				Project:   "valid-project",
			},
			expectError: true,
			errorMsg:    "workspace path cannot be empty",
		},
		{
			name: "verbose and quiet both true",
			config: &Config{
				Workspace: "/valid/path",
				Verbose:   true,
				Quiet:     true,
			},
			expectError: true,
			errorMsg:    "verbose and quiet cannot both be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_XDGConfigPaths(t *testing.T) {
	config := &Config{}
	paths := config.getConfigPaths()

	// Should include XDG_CONFIG_HOME path
	homeConfig := filepath.Join(os.Getenv("HOME"), ".config", "twiggit", "config.yaml")
	assert.Contains(t, paths, homeConfig)

	// Should include legacy home path
	legacyConfig := filepath.Join(os.Getenv("HOME"), ".twiggit.yaml")
	assert.Contains(t, paths, legacyConfig)

	// Should have at least these 2 paths
	assert.GreaterOrEqual(t, len(paths), 2)
}
