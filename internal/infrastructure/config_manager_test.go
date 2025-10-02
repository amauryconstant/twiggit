package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func TestConfigManager_Load_Defaults(t *testing.T) {
	manager := NewConfigManager()

	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify defaults are loaded
	defaultConfig := domain.DefaultConfig()
	assert.Equal(t, defaultConfig.ProjectsDirectory, config.ProjectsDirectory)
	assert.Equal(t, defaultConfig.WorktreesDirectory, config.WorktreesDirectory)
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
}

func TestConfigManager_GetConfig_Immutable(t *testing.T) {
	manager := NewConfigManager()

	config, err := manager.Load()
	require.NoError(t, err)

	// Try to modify returned config
	config.ProjectsDirectory = "/modified/path"

	// Get config again - should not be modified
	newConfig := manager.GetConfig()
	assert.NotEqual(t, "/modified/path", newConfig.ProjectsDirectory)
}

func TestConfigManager_GetConfig_DeepCopy(t *testing.T) {
	manager := NewConfigManager()

	config, err := manager.Load()
	require.NoError(t, err)

	// Modify the returned config
	config.ProjectsDirectory = "/modified/path"
	config.WorktreesDirectory = "/another/path"
	config.DefaultSourceBranch = "modified"

	// Get config again - should be original values
	originalConfig := manager.GetConfig()
	defaultConfig := domain.DefaultConfig()
	assert.Equal(t, defaultConfig.ProjectsDirectory, originalConfig.ProjectsDirectory)
	assert.Equal(t, defaultConfig.WorktreesDirectory, originalConfig.WorktreesDirectory)
	assert.Equal(t, defaultConfig.DefaultSourceBranch, originalConfig.DefaultSourceBranch)
}

// Pure function tests for extracted functions

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name          string
		xdgConfigHome string
		homeDir       string
		expectedPath  string
	}{
		{
			name:          "XDG_CONFIG_HOME takes precedence",
			xdgConfigHome: "/custom/config",
			homeDir:       "/home/user",
			expectedPath:  "/custom/config/twiggit/config.toml",
		},
		{
			name:          "fallback to HOME/.config when XDG not set",
			xdgConfigHome: "",
			homeDir:       "/home/user",
			expectedPath:  "/home/user/.config/twiggit/config.toml",
		},
		{
			name:          "fallback to current directory when HOME unavailable",
			xdgConfigHome: "",
			homeDir:       "",
			expectedPath:  "config.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := resolveConfigPath(tt.xdgConfigHome, tt.homeDir)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestBuildDefaultConfig(t *testing.T) {
	config := buildDefaultConfig()

	require.NotNil(t, config)

	// Verify it matches domain.DefaultConfig()
	expectedConfig := domain.DefaultConfig()
	assert.Equal(t, expectedConfig.ProjectsDirectory, config.ProjectsDirectory)
	assert.Equal(t, expectedConfig.WorktreesDirectory, config.WorktreesDirectory)
	assert.Equal(t, expectedConfig.DefaultSourceBranch, config.DefaultSourceBranch)

	// Verify immutability - modifying returned config shouldn't affect future calls
	config.ProjectsDirectory = "/modified"
	newConfig := buildDefaultConfig()
	assert.NotEqual(t, "/modified", newConfig.ProjectsDirectory)
}

func TestConfigFileExists(t *testing.T) {
	// Test with a file that doesn't exist
	exists := configFileExists("/nonexistent/path/config.toml")
	assert.False(t, exists)

	// Test with absolute path to go.mod (should exist in this project)
	exists = configFileExists("/home/amaury/Projects/twiggit/go.mod")
	assert.True(t, exists)
}

func TestValidateConfig(t *testing.T) {
	// Test with valid config
	validConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	err := validateConfig(validConfig)
	require.NoError(t, err)

	// Test with invalid config (empty directories)
	invalidConfig := &domain.Config{
		ProjectsDirectory:   "",
		WorktreesDirectory:  "",
		DefaultSourceBranch: "",
	}

	err = validateConfig(invalidConfig)
	assert.Error(t, err)
}

func TestCopyConfig(t *testing.T) {
	originalConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	// Copy the config
	copiedConfig := copyConfig(originalConfig)

	require.NotNil(t, copiedConfig)
	assert.Equal(t, originalConfig.ProjectsDirectory, copiedConfig.ProjectsDirectory)
	assert.Equal(t, originalConfig.WorktreesDirectory, copiedConfig.WorktreesDirectory)
	assert.Equal(t, originalConfig.DefaultSourceBranch, copiedConfig.DefaultSourceBranch)

	// Verify immutability - modifying copy shouldn't affect original
	copiedConfig.ProjectsDirectory = "/modified/path"
	assert.NotEqual(t, "/modified/path", originalConfig.ProjectsDirectory)
	assert.Equal(t, "/home/user/Projects", originalConfig.ProjectsDirectory)
}
