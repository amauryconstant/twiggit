//go:build integration
// +build integration

package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestConfigManager_Integration_ConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()

	// Create config directory and file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
worktrees_dir = "/test/worktrees"
default_source_branch = "develop"
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	config, err := manager.Load()
	require.NoError(t, err)

	assert.Equal(t, "/test/projects", config.ProjectsDirectory)
	assert.Equal(t, "/test/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "develop", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_XDGFallback(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()

	// Create config file in .config structure
	configDir := filepath.Join(tempDir, ".config", "twiggit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/fallback/projects"
worktrees_dir = "/fallback/worktrees"
default_source_branch = "main"
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set HOME to temp directory, but not XDG_CONFIG_HOME
	os.Setenv("HOME", tempDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Unsetenv("HOME")

	config, err := manager.Load()
	require.NoError(t, err)

	assert.Equal(t, "/fallback/projects", config.ProjectsDirectory)
	assert.Equal(t, "/fallback/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_Validation(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()

	// Create config directory and invalid config file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "relative/path"
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	_, err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestConfigManager_Integration_MalformedTOML(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()

	// Create config directory and malformed TOML file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
invalid toml syntax here
worktrees_dir = "/test/worktrees"
`

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	_, err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestConfigManager_Integration_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	// Ensure clean environment state
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	// Set XDG_CONFIG_HOME to empty temp directory (no config file)
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	os.Unsetenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a fresh manager after setting environment variable
	manager := NewConfigManager()

	config, err := manager.Load()
	require.NoError(t, err)

	// Should load defaults when no config file exists
	defaultConfig := domain.DefaultConfig()
	assert.Equal(t, defaultConfig.ProjectsDirectory, config.ProjectsDirectory)
	assert.Equal(t, defaultConfig.WorktreesDirectory, config.WorktreesDirectory)
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
}
