package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func setupConfigManagerTest(t *testing.T) (domain.ConfigManager, string, string) {
	t.Helper()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	manager := NewConfigManager()
	t.Cleanup(func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	})
	return manager, tempDir, originalXDG
}

func TestConfigManager_LoadDefaults(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	defaultConfig := domain.DefaultConfig()

	assert.Contains(t, config.ProjectsDirectory, "Projects", "ProjectsDirectory should contain 'Projects'")
	assert.Contains(t, config.WorktreesDirectory, "Worktrees", "WorktreesDirectory should contain 'Worktrees'")

	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	assert.Equal(t, defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	assert.Equal(t, defaultConfig.Git.CacheEnabled, config.Git.CacheEnabled)

	assert.Equal(t, defaultConfig.ContextDetection.CacheTTL, config.ContextDetection.CacheTTL)
}

func TestConfigManager_GetConfigImmutable(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)

	config.ProjectsDirectory = "/modified/path"

	newConfig := manager.GetConfig()
	assert.NotEqual(t, "/modified/path", newConfig.ProjectsDirectory)
}

func TestConfigManager_GetConfigDeepCopy(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)

	originalProjectsDir := config.ProjectsDirectory
	originalWorktreesDir := config.WorktreesDirectory
	originalSourceBranch := config.DefaultSourceBranch

	config.ProjectsDirectory = "/modified/path"
	config.WorktreesDirectory = "/another/path"
	config.DefaultSourceBranch = "modified"

	originalConfig := manager.GetConfig()
	assert.Equal(t, originalProjectsDir, originalConfig.ProjectsDirectory)
	assert.Equal(t, originalWorktreesDir, originalConfig.WorktreesDirectory)
	assert.Equal(t, originalSourceBranch, originalConfig.DefaultSourceBranch)
}

func TestConfigManager_ResolveConfigPath(t *testing.T) {
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := resolveConfigPath(tc.xdgConfigHome, tc.homeDir)
			assert.Equal(t, tc.expectedPath, path)
		})
	}
}

func TestConfigManager_BuildDefaultConfig(t *testing.T) {
	config := buildDefaultConfig()

	require.NotNil(t, config)

	expectedConfig := domain.DefaultConfig()

	assert.Contains(t, config.ProjectsDirectory, "Projects")
	assert.Contains(t, config.WorktreesDirectory, "Worktrees")
	assert.Equal(t, expectedConfig.DefaultSourceBranch, config.DefaultSourceBranch)

	assert.Equal(t, expectedConfig.ContextDetection.CacheTTL, config.ContextDetection.CacheTTL)
	assert.Equal(t, expectedConfig.Git.CLITimeout, config.Git.CLITimeout)

	originalProjectsDir := config.ProjectsDirectory
	config.ProjectsDirectory = "/modified"
	newConfig := buildDefaultConfig()
	assert.NotEqual(t, "/modified", newConfig.ProjectsDirectory)
	assert.Equal(t, originalProjectsDir, newConfig.ProjectsDirectory)
}

func TestConfigManager_ConfigFileExists(t *testing.T) {
	exists := configFileExists("/nonexistent/path/config.toml")
	assert.False(t, exists)

	goModPath := "../../go.mod"
	absPath, err := filepath.Abs(goModPath)
	require.NoError(t, err)
	exists = configFileExists(absPath)
	assert.True(t, exists)
}

func TestConfigManager_ValidateConfig(t *testing.T) {
	validConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	err := validateConfig(validConfig)
	require.NoError(t, err)

	invalidConfig := &domain.Config{
		ProjectsDirectory:   "",
		WorktreesDirectory:  "",
		DefaultSourceBranch: "",
	}

	err = validateConfig(invalidConfig)
	assert.Error(t, err)
}

func TestConfigManager_CopyConfig(t *testing.T) {
	originalConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	copiedConfig := copyConfig(originalConfig)

	require.NotNil(t, copiedConfig)
	assert.Equal(t, originalConfig.ProjectsDirectory, copiedConfig.ProjectsDirectory)
	assert.Equal(t, originalConfig.WorktreesDirectory, copiedConfig.WorktreesDirectory)
	assert.Equal(t, originalConfig.DefaultSourceBranch, copiedConfig.DefaultSourceBranch)

	copiedConfig.ProjectsDirectory = "/modified/path"
	assert.NotEqual(t, "/modified/path", originalConfig.ProjectsDirectory)
	assert.Equal(t, "/home/user/Projects", originalConfig.ProjectsDirectory)
}

func TestConfigManager_LoadDefaultsErrorHandling(t *testing.T) {
	manager := NewConfigManager()
	config, err := manager.Load()

	require.NoError(t, err, "Load() should succeed with valid default keys")
	require.NotNil(t, config, "Config should be loaded successfully")

	defaultConfig := domain.DefaultConfig()
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	assert.Equal(t, defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	assert.Equal(t, defaultConfig.Git.CacheEnabled, config.Git.CacheEnabled)
}

func TestConfigManager_ExpandConfigPath(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Setenv("HOME", "/home/testuser")
	os.Setenv("TEST_VAR", "/custom/path")

	tests := []struct {
		name     string
		input    string
		setupEnv func()
		expected string
	}{
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde expansion",
			input:    "~/Projects",
			expected: "/home/testuser/Projects",
		},
		{
			name:     "tilde alone",
			input:    "~",
			expected: "/home/testuser",
		},
		{
			name:     "dollar sign variable",
			input:    "$HOME/Projects",
			expected: "/home/testuser/Projects",
		},
		{
			name:     "curly brace variable",
			input:    "${HOME}/Worktrees",
			expected: "/home/testuser/Worktrees",
		},
		{
			name:     "custom env variable",
			input:    "$TEST_VAR/subdir",
			expected: "/custom/path/subdir",
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path/Projects",
			expected: "/absolute/path/Projects",
		},
		{
			name:     "undefined env var expands to empty",
			input:    "$UNDEFINED_VAR/Projects",
			expected: "/Projects",
		},
		{
			name:     "mixed variables in path",
			input:    "$HOME/${TEST_VAR}/mixed",
			expected: "/home/testuser//custom/path/mixed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupEnv != nil {
				tc.setupEnv()
			}
			result := expandConfigPath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestConfigManager_ExpandConfigPathFallbacks(t *testing.T) {
	t.Run("handles paths gracefully", func(t *testing.T) {
		result := expandConfigPath("~/test")
		assert.NotEmpty(t, result)
		assert.NotContains(t, result, "~", "Tilde should be expanded")
	})
}

func TestConfigManager_NormalizeConfigPaths(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", "/home/testuser")

	tests := []struct {
		name              string
		projectsDir       string
		worktreesDir      string
		backupDir         string
		expectedProjects  string
		expectedWorktrees string
		expectedBackupDir string
	}{
		{
			name:              "expand all three path fields",
			projectsDir:       "$HOME/Projects",
			worktreesDir:      "${HOME}/Worktrees",
			backupDir:         "~/.config/twiggit/backups",
			expectedProjects:  "/home/testuser/Projects",
			expectedWorktrees: "/home/testuser/Worktrees",
			expectedBackupDir: "/home/testuser/.config/twiggit/backups",
		},
		{
			name:              "absolute paths unchanged",
			projectsDir:       "/absolute/projects",
			worktreesDir:      "/absolute/worktrees",
			backupDir:         "/absolute/backups",
			expectedProjects:  "/absolute/projects",
			expectedWorktrees: "/absolute/worktrees",
			expectedBackupDir: "/absolute/backups",
		},
		{
			name:              "empty paths remain empty",
			projectsDir:       "",
			worktreesDir:      "",
			backupDir:         "",
			expectedProjects:  "",
			expectedWorktrees: "",
			expectedBackupDir: "",
		},
		{
			name:              "mixed absolute and variable paths",
			projectsDir:       "/absolute/projects",
			worktreesDir:      "$HOME/Worktrees",
			backupDir:         "~/.backups",
			expectedProjects:  "/absolute/projects",
			expectedWorktrees: "/home/testuser/Worktrees",
			expectedBackupDir: "/home/testuser/.backups",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &domain.Config{
				ProjectsDirectory:  tc.projectsDir,
				WorktreesDirectory: tc.worktreesDir,
				Shell: domain.ShellConfig{
					Wrapper: domain.ShellWrapperConfig{
						BackupDir: tc.backupDir,
					},
				},
			}

			normalizeConfigPaths(config)

			assert.Equal(t, tc.expectedProjects, config.ProjectsDirectory)
			assert.Equal(t, tc.expectedWorktrees, config.WorktreesDirectory)
			assert.Equal(t, tc.expectedBackupDir, config.Shell.Wrapper.BackupDir)
		})
	}
}

func TestConfigManager_LoadWithEnvVarExpansion(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
	}()

	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	os.Setenv("TWIGGIT_TEST_PROJECTS", "/custom/projects")
	os.Setenv("TWIGGIT_TEST_WORKTREES", "/custom/worktrees")

	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `
projects_dir = "$TWIGGIT_TEST_PROJECTS"
worktrees_dir = "${TWIGGIT_TEST_WORKTREES}"

[shell.wrapper]
backup_dir = "~/backups"
`
	configPath := filepath.Join(configDir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	manager := NewConfigManager()
	config, err := manager.Load()

	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "/custom/projects", config.ProjectsDirectory)
	assert.Equal(t, "/custom/worktrees", config.WorktreesDirectory)
	assert.Equal(t, filepath.Join(tempDir, "backups"), config.Shell.Wrapper.BackupDir)
}
