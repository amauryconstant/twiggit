package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"twiggit/internal/domain"
)

type ConfigManagerTestSuite struct {
	suite.Suite
	manager     domain.ConfigManager
	tempDir     string
	originalXDG string
}

func (s *ConfigManagerTestSuite) SetupTest() {
	s.originalXDG = os.Getenv("XDG_CONFIG_HOME")
	s.tempDir = s.T().TempDir()
	os.Setenv("XDG_CONFIG_HOME", s.tempDir)
	s.manager = NewConfigManager()
}

func (s *ConfigManagerTestSuite) TearDownTest() {
	if s.originalXDG != "" {
		os.Setenv("XDG_CONFIG_HOME", s.originalXDG)
	} else {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
}

func TestConfigManager(t *testing.T) {
	suite.Run(t, new(ConfigManagerTestSuite))
}

func (s *ConfigManagerTestSuite) TestLoadDefaults() {
	config, err := s.manager.Load()
	s.Require().NoError(err)
	s.Require().NotNil(config)

	defaultConfig := domain.DefaultConfig()

	s.Contains(config.ProjectsDirectory, "Projects", "ProjectsDirectory should contain 'Projects'")
	s.Contains(config.WorktreesDirectory, "Worktrees", "WorktreesDirectory should contain 'Worktrees'")

	s.Equal(defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	s.Equal(defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	s.Equal(defaultConfig.Git.CacheEnabled, config.Git.CacheEnabled)

	s.Equal(defaultConfig.ContextDetection.CacheTTL, config.ContextDetection.CacheTTL)
}

func (s *ConfigManagerTestSuite) TestGetConfigImmutable() {
	config, err := s.manager.Load()
	s.Require().NoError(err)

	config.ProjectsDirectory = "/modified/path"

	newConfig := s.manager.GetConfig()
	s.NotEqual("/modified/path", newConfig.ProjectsDirectory)
}

func (s *ConfigManagerTestSuite) TestGetConfigDeepCopy() {
	config, err := s.manager.Load()
	s.Require().NoError(err)

	originalProjectsDir := config.ProjectsDirectory
	originalWorktreesDir := config.WorktreesDirectory
	originalSourceBranch := config.DefaultSourceBranch

	config.ProjectsDirectory = "/modified/path"
	config.WorktreesDirectory = "/another/path"
	config.DefaultSourceBranch = "modified"

	originalConfig := s.manager.GetConfig()
	s.Equal(originalProjectsDir, originalConfig.ProjectsDirectory)
	s.Equal(originalWorktreesDir, originalConfig.WorktreesDirectory)
	s.Equal(originalSourceBranch, originalConfig.DefaultSourceBranch)
}

func (s *ConfigManagerTestSuite) TestResolveConfigPath() {
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
		s.Run(tc.name, func() {
			path := resolveConfigPath(tc.xdgConfigHome, tc.homeDir)
			s.Equal(tc.expectedPath, path)
		})
	}
}

func (s *ConfigManagerTestSuite) TestBuildDefaultConfig() {
	config := buildDefaultConfig()

	s.Require().NotNil(config)

	expectedConfig := domain.DefaultConfig()

	s.Contains(config.ProjectsDirectory, "Projects")
	s.Contains(config.WorktreesDirectory, "Worktrees")
	s.Equal(expectedConfig.DefaultSourceBranch, config.DefaultSourceBranch)

	s.Equal(expectedConfig.ContextDetection.CacheTTL, config.ContextDetection.CacheTTL)
	s.Equal(expectedConfig.Git.CLITimeout, config.Git.CLITimeout)

	originalProjectsDir := config.ProjectsDirectory
	config.ProjectsDirectory = "/modified"
	newConfig := buildDefaultConfig()
	s.NotEqual("/modified", newConfig.ProjectsDirectory)
	s.Equal(originalProjectsDir, newConfig.ProjectsDirectory)
}

func (s *ConfigManagerTestSuite) TestConfigFileExists() {
	exists := configFileExists("/nonexistent/path/config.toml")
	s.False(exists)

	goModPath := "../../go.mod"
	absPath, err := filepath.Abs(goModPath)
	s.Require().NoError(err)
	exists = configFileExists(absPath)
	s.True(exists)
}

func (s *ConfigManagerTestSuite) TestValidateConfig() {
	validConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	err := validateConfig(validConfig)
	s.Require().NoError(err)

	invalidConfig := &domain.Config{
		ProjectsDirectory:   "",
		WorktreesDirectory:  "",
		DefaultSourceBranch: "",
	}

	err = validateConfig(invalidConfig)
	s.Error(err)
}

func (s *ConfigManagerTestSuite) TestCopyConfig() {
	originalConfig := &domain.Config{
		ProjectsDirectory:   "/home/user/Projects",
		WorktreesDirectory:  "/home/user/Worktrees",
		DefaultSourceBranch: "main",
	}

	copiedConfig := copyConfig(originalConfig)

	s.Require().NotNil(copiedConfig)
	s.Equal(originalConfig.ProjectsDirectory, copiedConfig.ProjectsDirectory)
	s.Equal(originalConfig.WorktreesDirectory, copiedConfig.WorktreesDirectory)
	s.Equal(originalConfig.DefaultSourceBranch, copiedConfig.DefaultSourceBranch)

	copiedConfig.ProjectsDirectory = "/modified/path"
	s.NotEqual("/modified/path", originalConfig.ProjectsDirectory)
	s.Equal("/home/user/Projects", originalConfig.ProjectsDirectory)
}

func (s *ConfigManagerTestSuite) TestLoadDefaultsErrorHandling() {
	manager := NewConfigManager()
	config, err := manager.Load()

	s.Require().NoError(err, "Load() should succeed with valid default keys")
	s.Require().NotNil(config, "Config should be loaded successfully")

	defaultConfig := domain.DefaultConfig()
	s.Equal(defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
	s.Equal(defaultConfig.Git.CLITimeout, config.Git.CLITimeout)
	s.Equal(defaultConfig.Git.CacheEnabled, config.Git.CacheEnabled)
}

func (s *ConfigManagerTestSuite) TestExpandConfigPath() {
	// Save original HOME and restore after test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set up test environment
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
		s.Run(tc.name, func() {
			if tc.setupEnv != nil {
				tc.setupEnv()
			}
			result := expandConfigPath(tc.input)
			s.Equal(tc.expected, result)
		})
	}
}

func (s *ConfigManagerTestSuite) TestExpandConfigPathFallbacks() {
	// Test fallback behavior when os.UserHomeDir would fail
	// We can't easily mock os.UserHomeDir, but we can test the logic indirectly
	// by ensuring the function doesn't panic and returns something reasonable

	s.Run("handles paths gracefully", func() {
		// Test with a simple tilde path - this should always work
		result := expandConfigPath("~/test")
		s.NotEmpty(result)
		s.NotContains(result, "~", "Tilde should be expanded")
	})
}

func (s *ConfigManagerTestSuite) TestNormalizeConfigPaths() {
	// Save original HOME and restore after test
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
		s.Run(tc.name, func() {
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

			s.Equal(tc.expectedProjects, config.ProjectsDirectory)
			s.Equal(tc.expectedWorktrees, config.WorktreesDirectory)
			s.Equal(tc.expectedBackupDir, config.Shell.Wrapper.BackupDir)
		})
	}
}

func (s *ConfigManagerTestSuite) TestLoadWithEnvVarExpansion() {
	// Save original environment
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
	}()

	// Set up test environment
	os.Setenv("HOME", s.tempDir)
	os.Setenv("XDG_CONFIG_HOME", s.tempDir)
	os.Setenv("TWIGGIT_TEST_PROJECTS", "/custom/projects")
	os.Setenv("TWIGGIT_TEST_WORKTREES", "/custom/worktrees")

	// Create config file with environment variables
	configDir := filepath.Join(s.tempDir, "twiggit")
	s.Require().NoError(os.MkdirAll(configDir, 0755))

	configContent := `
projects_dir = "$TWIGGIT_TEST_PROJECTS"
worktrees_dir = "${TWIGGIT_TEST_WORKTREES}"

[shell.wrapper]
backup_dir = "~/backups"
`
	configPath := filepath.Join(configDir, "config.toml")
	s.Require().NoError(os.WriteFile(configPath, []byte(configContent), 0644))

	// Load config
	manager := NewConfigManager()
	config, err := manager.Load()

	s.Require().NoError(err)
	s.Require().NotNil(config)

	// Verify expansion occurred
	s.Equal("/custom/projects", config.ProjectsDirectory)
	s.Equal("/custom/worktrees", config.WorktreesDirectory)
	s.Equal(filepath.Join(s.tempDir, "backups"), config.Shell.Wrapper.BackupDir)
}
