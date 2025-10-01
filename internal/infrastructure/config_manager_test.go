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
