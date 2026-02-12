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
