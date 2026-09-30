//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"twiggit/internal/config"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_Integration_ConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	manager := config.NewManager()

	// Create config directory and file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
worktrees_dir = "/test/worktrees"
default_source_branch = "develop"
`

	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	config, err := manager.Load()
	require.NoError(t, err)

	assert.Equal(t, "/test/projects", config.ProjectsDirectory)
	assert.Equal(t, "/test/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "develop", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_XDGFallback(t *testing.T) {
	tempDir := t.TempDir()
	manager := config.NewManager()

	// Create config file in .config structure
	configDir := filepath.Join(tempDir, ".config", "twiggit")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/fallback/projects"
worktrees_dir = "/fallback/worktrees"
default_source_branch = "main"
`

	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set HOME to temp directory, but not XDG_CONFIG_HOME
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	config, err := manager.Load()
	require.NoError(t, err)

	assert.Equal(t, "/fallback/projects", config.ProjectsDirectory)
	assert.Equal(t, "/fallback/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_Validation(t *testing.T) {
	tempDir := t.TempDir()
	manager := config.NewManager()

	// Create config directory and invalid config file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "relative/path"
`

	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	_, err = manager.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestConfigManager_Integration_MalformedTOML(t *testing.T) {
	tempDir := t.TempDir()
	manager := config.NewManager()

	// Create config directory and malformed TOML file
	configDir := filepath.Join(tempDir, "twiggit")
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
invalid toml syntax here
worktrees_dir = "/test/worktrees"
`

	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Set XDG_CONFIG_HOME to temp directory
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	_, err = manager.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestConfigManager_Integration_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()

	// Ensure clean environment state. t.Setenv auto-restores after the test.
	// Set XDG_CONFIG_HOME to empty temp directory (no config file).
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("HOME", "")

	// Create a fresh manager after setting environment variable
	manager := config.NewManager()

	config, err := manager.Load()
	require.NoError(t, err)

	// Should load defaults when no config file exists
	defaultConfig := core.DefaultConfig()
	assert.Equal(t, defaultConfig.ProjectsDirectory, config.ProjectsDirectory)
	assert.Equal(t, defaultConfig.WorktreesDirectory, config.WorktreesDirectory)
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
}
