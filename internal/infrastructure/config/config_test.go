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
		"TWIGGIT_WORKSPACES_PATH",
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
		"TWIGGIT_WORKSPACES_PATH",
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
	s.Equal(filepath.Join(os.Getenv("HOME"), "Workspaces"), config.WorkspacesPath)
	s.Empty(config.Project)
	s.False(config.Verbose)
	s.False(config.Quiet)
}

// TestConfig_LoadFromEnvironment tests loading config from environment variables
func (s *ConfigTestSuite) TestConfig_LoadFromEnvironment() {
	// Set environment variables
	s.Require().NoError(os.Setenv("TWIGGIT_WORKSPACES_PATH", "/custom/workspace"))
	s.Require().NoError(os.Setenv("TWIGGIT_PROJECT", "my-project"))
	s.Require().NoError(os.Setenv("TWIGGIT_VERBOSE", "true"))
	s.Require().NoError(os.Setenv("TWIGGIT_QUIET", "false"))

	config, err := LoadConfig()
	s.Require().NoError(err)

	s.Equal("/custom/workspace", config.WorkspacesPath)
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
	configFile := filepath.Join(configDir, "config.toml")

	configContent := `workspaces_path = "/file/workspace"
project = "file-project"
verbose = true
quiet = false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Set XDG_CONFIG_HOME to temp directory
	s.Require().NoError(os.Setenv("XDG_CONFIG_HOME", tempDir))

	config, err := LoadConfig()
	s.Require().NoError(err)

	s.Equal("/file/workspace", config.WorkspacesPath)
	s.Equal("file-project", config.Project)
	s.True(config.Verbose)
	s.False(config.Quiet)
	s.Empty(config.DefaultSourceBranch) // Default value
}

// TestConfig_EnvironmentOverridesFile tests that environment variables override file values
func (s *ConfigTestSuite) TestConfig_EnvironmentOverridesFile() {
	// Create temporary config file
	tempDir := s.T().TempDir()
	configFile := filepath.Join(tempDir, "twiggit", "config.toml")
	s.Require().NoError(os.MkdirAll(filepath.Dir(configFile), 0755))

	configContent := `
workspace = "/file/workspace"
project = "file-project"
verbose = false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	s.Require().NoError(err)

	// Set environment variables to override file values
	s.Require().NoError(os.Setenv("XDG_CONFIG_HOME", tempDir))
	s.Require().NoError(os.Setenv("TWIGGIT_WORKSPACES_PATH", "/env/workspace"))
	s.Require().NoError(os.Setenv("TWIGGIT_VERBOSE", "true"))

	config, err := LoadConfig()
	s.Require().NoError(err)

	// Environment should override file
	s.Equal("/env/workspace", config.WorkspacesPath)
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
				WorkspacesPath: "/valid/path",
				Project:        "valid-project",
			},
			expectError: false,
		},
		{
			name: "empty workspace",
			config: &Config{
				WorkspacesPath: "",
				Project:        "valid-project",
			},
			expectError: true,
			errorMsg:    "workspaces path cannot be empty",
		},
		{
			name: "verbose and quiet both true",
			config: &Config{
				WorkspacesPath: "/valid/path",
				Verbose:        true,
				Quiet:          true,
			},
			expectError: true,
			errorMsg:    "verbose and quiet cannot both be enabled",
		},
		{
			name: "invalid default source branch name",
			config: &Config{
				WorkspacesPath:      "/valid/path",
				DefaultSourceBranch: "invalid branch name",
			},
			expectError: true,
			errorMsg:    "invalid default source branch name",
		},
		{
			name: "valid default source branch name",
			config: &Config{
				WorkspacesPath:      "/valid/path",
				DefaultSourceBranch: "develop",
			},
			expectError: false,
		},
		{
			name: "invalid default source branch name",
			config: &Config{
				WorkspacesPath:      "/valid/path",
				DefaultSourceBranch: "invalid branch name",
			},
			expectError: true,
			errorMsg:    "invalid default source branch name",
		},
		{
			name: "valid default source branch name",
			config: &Config{
				WorkspacesPath:      "/valid/path",
				DefaultSourceBranch: "develop",
			},
			expectError: false,
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
	homeConfig := filepath.Join(os.Getenv("HOME"), ".config", "twiggit", "config.toml")
	s.Contains(paths, homeConfig)

	// Should include legacy home path
	legacyConfig := filepath.Join(os.Getenv("HOME"), ".twiggit.toml")
	s.Contains(paths, legacyConfig)

	// Should have at least these 2 paths
	s.GreaterOrEqual(len(paths), 2)
}

// BenchmarkConfig_LoadConfig_NoFile benchmarks configuration loading when no config file exists
func BenchmarkConfig_LoadConfig_NoFile(b *testing.B) {
	// Save original environment
	originalXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")

	// Create temp directory and ensure no config files exist
	tempDir := b.TempDir()
	os.Setenv("HOME", tempDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	defer func() {
		if originalXdgConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXdgConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatalf("LoadConfig failed: %v", err)
		}
	}
}

// BenchmarkConfig_LoadConfig_WithFile benchmarks configuration loading with a TOML config file
func BenchmarkConfig_LoadConfig_WithFile(b *testing.B) {
	// Save original environment
	originalXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")

	// Create temp directory with config file
	tempDir := b.TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	os.MkdirAll(configDir, 0755)

	configContent := `# Twiggit configuration
projects_path = "/home/user/Projects"
workspaces_path = "/home/user/Workspaces"
project = "my-project"
default_source_branch = "develop"
verbose = false
quiet = false
`
	configFile := filepath.Join(configDir, "config.toml")
	os.WriteFile(configFile, []byte(configContent), 0644)

	os.Setenv("XDG_CONFIG_HOME", tempDir)
	os.Setenv("HOME", tempDir)

	defer func() {
		if originalXdgConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXdgConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatalf("LoadConfig failed: %v", err)
		}
	}
}

// BenchmarkConfig_LoadConfig_WithEnvironment benchmarks configuration loading with environment variables
func BenchmarkConfig_LoadConfig_WithEnvironment(b *testing.B) {
	// Save original environment
	originalXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	originalTwiggitProjectsPath := os.Getenv("TWIGGIT_PROJECTS_PATH")
	originalTwiggitWorkspacesPath := os.Getenv("TWIGGIT_WORKSPACES_PATH")
	originalTwiggitProject := os.Getenv("TWIGGIT_PROJECT")
	originalTwiggitDefaultSourceBranch := os.Getenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
	originalTwiggitVerbose := os.Getenv("TWIGGIT_VERBOSE")

	// Create temp directory and ensure no config files exist
	tempDir := b.TempDir()
	os.Setenv("HOME", tempDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	// Set environment variables
	os.Setenv("TWIGGIT_PROJECTS_PATH", "/env/projects")
	os.Setenv("TWIGGIT_WORKSPACES_PATH", "/env/workspaces")
	os.Setenv("TWIGGIT_PROJECT", "env-project")
	os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", "main")
	os.Setenv("TWIGGIT_VERBOSE", "true")

	defer func() {
		if originalXdgConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXdgConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalTwiggitProjectsPath != "" {
			os.Setenv("TWIGGIT_PROJECTS_PATH", originalTwiggitProjectsPath)
		} else {
			os.Unsetenv("TWIGGIT_PROJECTS_PATH")
		}
		if originalTwiggitWorkspacesPath != "" {
			os.Setenv("TWIGGIT_WORKSPACES_PATH", originalTwiggitWorkspacesPath)
		} else {
			os.Unsetenv("TWIGGIT_WORKSPACES_PATH")
		}
		if originalTwiggitProject != "" {
			os.Setenv("TWIGGIT_PROJECT", originalTwiggitProject)
		} else {
			os.Unsetenv("TWIGGIT_PROJECT")
		}
		if originalTwiggitDefaultSourceBranch != "" {
			os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", originalTwiggitDefaultSourceBranch)
		} else {
			os.Unsetenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
		}
		if originalTwiggitVerbose != "" {
			os.Setenv("TWIGGIT_VERBOSE", originalTwiggitVerbose)
		} else {
			os.Unsetenv("TWIGGIT_VERBOSE")
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatalf("LoadConfig failed: %v", err)
		}
	}
}

// BenchmarkConfig_LoadConfig_WithFileAndEnvironment benchmarks configuration loading with both file and environment
func BenchmarkConfig_LoadConfig_WithFileAndEnvironment(b *testing.B) {
	// Save original environment
	originalXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	originalTwiggitProjectsPath := os.Getenv("TWIGGIT_PROJECTS_PATH")
	originalTwiggitWorkspacesPath := os.Getenv("TWIGGIT_WORKSPACES_PATH")
	originalTwiggitProject := os.Getenv("TWIGGIT_PROJECT")
	originalTwiggitDefaultSourceBranch := os.Getenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
	originalTwiggitVerbose := os.Getenv("TWIGGIT_VERBOSE")

	// Create temp directory with config file
	tempDir := b.TempDir()
	configDir := filepath.Join(tempDir, "twiggit")
	os.MkdirAll(configDir, 0755)

	configContent := `# Twiggit configuration
projects_path = "/file/projects"
workspaces_path = "/file/workspaces"
project = "file-project"
default_source_branch = "develop"
verbose = false
quiet = false
`
	configFile := filepath.Join(configDir, "config.toml")
	os.WriteFile(configFile, []byte(configContent), 0644)

	os.Setenv("XDG_CONFIG_HOME", tempDir)
	os.Setenv("HOME", tempDir)

	// Set environment variables that should override file
	os.Setenv("TWIGGIT_PROJECTS_PATH", "/env/projects")
	os.Setenv("TWIGGIT_WORKSPACES_PATH", "/env/workspaces")
	os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", "main")
	os.Setenv("TWIGGIT_VERBOSE", "true")

	defer func() {
		if originalXdgConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXdgConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalTwiggitProjectsPath != "" {
			os.Setenv("TWIGGIT_PROJECTS_PATH", originalTwiggitProjectsPath)
		} else {
			os.Unsetenv("TWIGGIT_PROJECTS_PATH")
		}
		if originalTwiggitWorkspacesPath != "" {
			os.Setenv("TWIGGIT_WORKSPACES_PATH", originalTwiggitWorkspacesPath)
		} else {
			os.Unsetenv("TWIGGIT_WORKSPACES_PATH")
		}
		if originalTwiggitProject != "" {
			os.Setenv("TWIGGIT_PROJECT", originalTwiggitProject)
		} else {
			os.Unsetenv("TWIGGIT_PROJECT")
		}
		if originalTwiggitDefaultSourceBranch != "" {
			os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", originalTwiggitDefaultSourceBranch)
		} else {
			os.Unsetenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
		}
		if originalTwiggitVerbose != "" {
			os.Setenv("TWIGGIT_VERBOSE", originalTwiggitVerbose)
		} else {
			os.Unsetenv("TWIGGIT_VERBOSE")
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig()
		if err != nil {
			b.Fatalf("LoadConfig failed: %v", err)
		}
	}
}

// BenchmarkConfig_NewConfig benchmarks config creation with default values
func BenchmarkConfig_NewConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewConfig()
	}
}

// BenchmarkConfig_Validate benchmarks config validation
func BenchmarkConfig_Validate(b *testing.B) {
	config := &Config{
		ProjectsPath:        "/valid/projects/path",
		WorkspacesPath:      "/valid/workspaces/path",
		Project:             "valid-project",
		DefaultSourceBranch: "develop",
		Verbose:             false,
		Quiet:               false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := config.Validate()
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}

// BenchmarkConfig_getConfigPaths benchmarks config path resolution
func BenchmarkConfig_getConfigPaths(b *testing.B) {
	config := &Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.getConfigPaths()
	}
}

// TestConfigSuite runs the config test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
