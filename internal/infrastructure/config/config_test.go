package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite provides hybrid suite setup for config tests
type ConfigTestSuite struct {
	suite.Suite
	originalEnv map[string]string
}

// SetupTest saves original environment variables for each test
func (s *ConfigTestSuite) SetupTest() {
	s.originalEnv = make(map[string]string)
	envVars := []string{
		"TWIGGIT_WORKSPACE",
		"TWIGGIT_PROJECT",
		"TWIGGIT_VERBOSE",
		"TWIGGIT_QUIET",
		"XDG_CONFIG_HOME",
		"HOME",
	}

	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			s.originalEnv[env] = value
		}
	}
}

// TearDownTest restores original environment variables
func (s *ConfigTestSuite) TearDownTest() {
	for env := range s.originalEnv {
		_ = os.Setenv(env, s.originalEnv[env])
	}

	// Unset any env vars that weren't originally set
	envVars := []string{
		"TWIGGIT_WORKSPACE",
		"TWIGGIT_PROJECT",
		"TWIGGIT_VERBOSE",
		"TWIGGIT_QUIET",
		"XDG_CONFIG_HOME",
	}

	for _, env := range envVars {
		if _, exists := s.originalEnv[env]; !exists {
			_ = os.Unsetenv(env)
		}
	}
}

// TestConfig_NewConfig tests config creation with default values
func (s *ConfigTestSuite) TestConfig_NewConfig() {
	config := NewConfig()
	s.Require().NotNil(config)

	// Test default values
	s.Equal(filepath.Join(os.Getenv("HOME"), "Workspaces"), config.Workspace)
	s.Empty(config.Project)
	s.False(config.Verbose)
	s.False(config.Quiet)
}

// TestConfig_LoadFromEnvironment tests loading config from environment variables
func (s *ConfigTestSuite) TestConfig_LoadFromEnvironment() {
	// Set environment variables
	s.Require().NoError(os.Setenv("TWIGGIT_WORKSPACE", "/custom/workspace"))
	s.Require().NoError(os.Setenv("TWIGGIT_PROJECT", "my-project"))
	s.Require().NoError(os.Setenv("TWIGGIT_VERBOSE", "true"))
	s.Require().NoError(os.Setenv("TWIGGIT_QUIET", "false"))

	config, err := LoadConfig()
	s.Require().NoError(err)

	s.Equal("/custom/workspace", config.Workspace)
	s.Equal("my-project", config.Project)
	s.True(config.Verbose)
	s.False(config.Quiet)
}

// TestConfig_LoadFromFile tests loading config from file
func (s *ConfigTestSuite) TestConfig_LoadFromFile() {
	// Create temporary config file in the expected XDG structure
	tempDir := s.T().TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	s.Require().NoError(os.MkdirAll(configDir, 0755))
	configFile := filepath.Join(configDir, "config.yaml")

	configContent := `workspace: /file/workspace
project: file-project
verbose: true
quiet: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Set XDG_CONFIG_HOME to temp directory
	s.Require().NoError(os.Setenv("XDG_CONFIG_HOME", tempDir))

	config, err := LoadConfig()
	s.Require().NoError(err)

	s.Equal("/file/workspace", config.Workspace)
	s.Equal("file-project", config.Project)
	s.True(config.Verbose)
	s.False(config.Quiet)
}

// TestConfig_EnvironmentOverridesFile tests that environment variables override file values
func (s *ConfigTestSuite) TestConfig_EnvironmentOverridesFile() {
	// Create temporary config file
	tempDir := s.T().TempDir()
	configFile := filepath.Join(tempDir, "twiggit", "config.yaml")
	s.Require().NoError(os.MkdirAll(filepath.Dir(configFile), 0755))

	configContent := `
workspace: /file/workspace
project: file-project
verbose: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Set environment variables to override file values
	s.Require().NoError(os.Setenv("XDG_CONFIG_HOME", tempDir))
	s.Require().NoError(os.Setenv("TWIGGIT_WORKSPACE", "/env/workspace"))
	s.Require().NoError(os.Setenv("TWIGGIT_VERBOSE", "true"))

	config, err := LoadConfig()
	s.Require().NoError(err)

	// Environment should override file
	s.Equal("/env/workspace", config.Workspace)
	s.True(config.Verbose)
	// File value should be used where env not set
	s.Equal("file-project", config.Project)
}

// TestConfig_Validate tests config validation with table-driven approach
func (s *ConfigTestSuite) TestConfig_Validate() {
	testCases := []struct {
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
			errorMsg:    "workspaces path cannot be empty",
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

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			err := tt.config.Validate()

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestConfig_XDGConfigPaths tests XDG config path resolution
func (s *ConfigTestSuite) TestConfig_XDGConfigPaths() {
	config := &Config{}
	paths := config.getConfigPaths()

	// Should include XDG_CONFIG_HOME path
	homeConfig := filepath.Join(os.Getenv("HOME"), ".config", "twiggit", "config.yaml")
	s.Contains(paths, homeConfig)

	// Should include legacy home path
	legacyConfig := filepath.Join(os.Getenv("HOME"), ".twiggit.yaml")
	s.Contains(paths, legacyConfig)

	// Should have at least these 2 paths
	s.GreaterOrEqual(len(paths), 2)
}

// TestConfigSuite runs the config test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
