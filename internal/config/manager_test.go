package config

import (
	"os"
	"path/filepath"
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsetEnv removes key from the process environment and auto-restores
// its prior state on test cleanup. Needed for tests that assert on
// "variable unset" semantics (e.g. NO_COLOR lookup vs. get), which
// t.Setenv(key, "") cannot express because the variable remains
// present (just empty) in os.LookupEnv.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	orig, had := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

func setupConfigManagerTest(t *testing.T) (*Manager, string, string) {
	t.Helper()
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	manager := NewManager()
	return manager, tempDir, originalXDG
}

func TestConfigManager_LoadDefaults(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	defaultConfig := core.DefaultConfig()

	assert.Contains(t, config.ProjectsDirectory, "Projects", "ProjectsDirectory should contain 'Projects'")
	assert.Contains(t, config.WorktreesDirectory, "Worktrees", "WorktreesDirectory should contain 'Worktrees'")

	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	assert.Equal(t, defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	assert.Equal(t, defaultConfig.Git.IsCacheEnabled, config.Git.IsCacheEnabled)

	assert.Equal(t, defaultConfig.ContextDetection.CacheTTL, config.ContextDetection.CacheTTL)
}

func TestConfigManager_GetConfigImmutable(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)

	config.ProjectsDirectory = "/modified/path"

	newConfig := manager.Config()
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

	originalConfig := manager.Config()
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

	expectedConfig := core.DefaultConfig()

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
	validConfig := &core.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	err := validateConfig(validConfig)
	require.NoError(t, err)

	invalidConfig := &core.Config{
		ProjectsDirectory:   "",
		WorktreesDirectory:  "",
		DefaultSourceBranch: "",
	}

	err = validateConfig(invalidConfig)
	assert.Error(t, err)
}

func TestConfigManager_CopyConfig(t *testing.T) {
	originalConfig := &core.Config{
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
	manager := NewManager()
	config, err := manager.Load()

	require.NoError(t, err, "Load() should succeed with valid default keys")
	require.NotNil(t, config, "Config should be loaded successfully")

	defaultConfig := core.DefaultConfig()
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	assert.Equal(t, defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	assert.Equal(t, defaultConfig.Git.IsCacheEnabled, config.Git.IsCacheEnabled)
}

func TestConfigManager_ExpandConfigPath(t *testing.T) {
	t.Setenv("HOME", "/home/testuser")
	t.Setenv("TEST_VAR", "/custom/path")

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
	t.Setenv("HOME", "/home/testuser")

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
			config := &core.Config{
				ProjectsDirectory:  tc.projectsDir,
				WorktreesDirectory: tc.worktreesDir,
				Shell: core.ShellConfig{
					Wrapper: core.ShellWrapperConfig{
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
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("TWIGGIT_TEST_PROJECTS", "/custom/projects")
	t.Setenv("TWIGGIT_TEST_WORKTREES", "/custom/worktrees")

	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configContent := `
projects_dir = "$TWIGGIT_TEST_PROJECTS"
worktrees_dir = "${TWIGGIT_TEST_WORKTREES}"

[shell.wrapper]
backup_dir = "~/backups"
`
	configPath := filepath.Join(configDir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

	manager := NewManager()
	config, err := manager.Load()

	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "/custom/projects", config.ProjectsDirectory)
	assert.Equal(t, "/custom/worktrees", config.WorktreesDirectory)
	assert.Equal(t, filepath.Join(tempDir, "backups"), config.Shell.Wrapper.BackupDir)
}

func TestConfigManager_NoColorRespected(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.True(t, config.IsColorEnabled, "ColorEnabled should default to true when NO_COLOR is unset")

	t.Setenv("NO_COLOR", "1")
	config, err = manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.False(t, config.IsColorEnabled, "ColorEnabled should be false when NO_COLOR is set")

	unsetEnv(t, "NO_COLOR")
	config, err = manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.True(t, config.IsColorEnabled, "ColorEnabled should return to true when NO_COLOR is unset")
}

func TestConfigManager_ColorEnabledPropagatesToCopy(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	t.Setenv("NO_COLOR", "1")
	loaded, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.False(t, loaded.IsColorEnabled)

	cached := manager.Config()
	require.NotNil(t, cached)
	assert.False(t, cached.IsColorEnabled, "Config copy should preserve ColorEnabled=false")

	unsetEnv(t, "NO_COLOR")
	reloaded, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	assert.True(t, reloaded.IsColorEnabled, "ColorEnabled should reset when NO_COLOR is unset")
}

func TestConfig_RebaseSync_Defaults(t *testing.T) {
	manager, _, _ := setupConfigManagerTest(t)
	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	defaults := core.DefaultConfig()
	assert.Equal(t, defaults.Rebase.FetchOnAll, config.Rebase.FetchOnAll)
	assert.Equal(t, defaults.Rebase.ConflictPolicy, config.Rebase.ConflictPolicy)
	assert.Equal(t, defaults.Sync.DefaultRemote, config.Sync.DefaultRemote)
	assert.Equal(t, defaults.Sync.PruneRemoteRefs, config.Sync.PruneRemoteRefs)
	assert.Equal(t, defaults.Sync.RebaseAfterSync, config.Sync.RebaseAfterSync)
}

func TestConfig_RebaseSync_FileOverridesDefaults(t *testing.T) {
	manager, tempDir, _ := setupConfigManagerTest(t)
	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configContent := `
[rebase]
fetch_on_all = true
conflict_policy = "continue"

[sync]
default_remote = "upstream"
prune_remote_refs = false
rebase_after_sync = true
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configContent), 0o644))

	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.True(t, config.Rebase.FetchOnAll)
	assert.Equal(t, "continue", config.Rebase.ConflictPolicy)
	assert.Equal(t, "upstream", config.Sync.DefaultRemote)
	assert.False(t, config.Sync.PruneRemoteRefs)
	assert.True(t, config.Sync.RebaseAfterSync)
}

func TestConfig_RebaseSync_EnvOverridesFile(t *testing.T) {
	manager, tempDir, _ := setupConfigManagerTest(t)
	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configContent := `
[rebase]
fetch_on_all = false

[sync]
rebase_after_sync = false
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.toml"), []byte(configContent), 0o644))

	t.Setenv("TWIGGIT_REBASE__FETCH_ON_ALL", "true")
	t.Setenv("TWIGGIT_SYNC__REBASE_AFTER_SYNC", "true")

	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.True(t, config.Rebase.FetchOnAll, "env must override file fetch_on_all=false")
	assert.True(t, config.Sync.RebaseAfterSync, "env must override file rebase_after_sync=false")
}

func TestConfigManager_LoadWrapsKoanfErrors(t *testing.T) {
	manager, tempDir, _ := setupConfigManagerTest(t)

	configDir := filepath.Join(tempDir, "twiggit")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	configPath := filepath.Join(configDir, "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("this is not [ valid toml"), 0o644))

	_, err := manager.Load()
	require.Error(t, err)

	var oe *core.OperationError
	require.ErrorAs(t, err, &oe)
	assert.Equal(t, "config.load", oe.Op)
	assert.Equal(t, configPath, oe.Entity)
}
