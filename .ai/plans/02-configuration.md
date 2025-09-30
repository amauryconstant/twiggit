# Configuration System Implementation Plan

## Overview

This plan implements the Koanf-based configuration management system for twiggit with priority loading, validation, and immutable configuration after loading. The system follows the established foundation layer and integrates with the existing project structure.

## Context from Documentation

> **From implementation.md**: "Configuration SHALL be loaded using Koanf with TOML support. Environment variables SHALL override config file values. Command flags SHALL override all other configuration sources. Configuration validation SHALL occur during startup."

> **From technology.md**: "Koanf SHALL load configuration in priority order: defaults → config file → environment variables → command flags."

> **From implementation.md**: "Location: XDG Base Directory specification SHALL be followed for config folders (`$HOME/.config/twiggit/config.toml`). Format: TOML format SHALL be supported exclusively."

## Implementation Steps

### Step 1: Define Configuration Structure and Interfaces

#### 1.1 Create Configuration Domain Types

**File**: `internal/domain/config.go`

```go
package domain

import (
	"path/filepath"
)

// Config represents the complete application configuration
type Config struct {
	// Directory paths
	ProjectsDirectory string `toml:"projects_dir" koanf:"projects_dir"`
	WorktreesDirectory string `toml:"worktrees_dir" koanf:"worktrees_dir"`
	
	// Default behavior
	DefaultSourceBranch string `toml:"default_source_branch" koanf:"default_source_branch"`
	
	// Git implementation
	GitImplementation string `toml:"git_implementation" koanf:"git_implementation"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ProjectsDirectory:  filepath.Join(home, "Projects"),
		WorktreesDirectory: filepath.Join(home, "Workspaces"),
		DefaultSourceBranch: "main",
		GitImplementation: "go-git",
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var errors []error
	
	// Validate projects directory
	if !filepath.IsAbs(c.ProjectsDirectory) {
		errors = append(errors, fmt.Errorf("projects_directory must be absolute path: %s", c.ProjectsDirectory))
	}
	
	// Validate worktrees directory
	if !filepath.IsAbs(c.WorktreesDirectory) {
		errors = append(errors, fmt.Errorf("worktrees_directory must be absolute path: %s", c.WorktreesDirectory))
	}
	
	// Validate default source branch
	if c.DefaultSourceBranch == "" {
		errors = append(errors, fmt.Errorf("default_source_branch cannot be empty"))
	}
	
	// Validate git implementation
	validGitImpls := []string{"go-git", "system-git"}
	if !contains(validGitImpls, c.GitImplementation) {
		errors = append(errors, fmt.Errorf("git_implementation must be one of: %v", validGitImpls))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}
	
	return nil
}
```

#### 1.2 Define ConfigManager Interface

**File**: `internal/infrastructure/interfaces.go` (add to existing file)

```go
// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from all sources in priority order
	Load() (*domain.Config, error)
	
	// GetConfig returns the loaded configuration (immutable after Load)
	GetConfig() *domain.Config
	
	// ValidateConfig validates a configuration object
	ValidateConfig(config *domain.Config) error
}
```

### Step 2: Implement Koanf-based Configuration Manager

#### 2.1 Create Concrete Implementation

**File**: `internal/infrastructure/config/manager.go`

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/v2/parsers/toml"
	"github.com/knadh/koanf/v2/providers/env"
	"github.com/knadh/koanf/v2/providers/file"
	
	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure"
)

type koanfConfigManager struct {
	ko    *koanf.Koanf
	config *domain.Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() infrastructure.ConfigManager {
	return &koanfConfigManager{
		ko: koanf.New("."),
	}
}

// Load loads configuration from all sources in priority order
func (m *koanfConfigManager) Load() (*domain.Config, error) {
	// 1. Load defaults
	if err := m.loadDefaults(); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}
	
	// 2. Load config file
	if err := m.loadConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	
	// 3. Load environment variables
	if err := m.loadEnvironmentVariables(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}
	
	// 4. Unmarshal to config object
	config := &domain.Config{}
	if err := m.ko.Unmarshal("", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}
	
	// 5. Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// 6. Store immutable config
	m.config = config
	
	return config, nil
}

// GetConfig returns the loaded configuration
func (m *koanfConfigManager) GetConfig() *domain.Config {
	if m.config == nil {
		return nil
	}
	// Return a copy to maintain immutability
	configCopy := *m.config
	return &configCopy
}

// ValidateConfig validates a configuration object
func (m *koanfConfigManager) ValidateConfig(config *domain.Config) error {
	return config.Validate()
}

// loadDefaults loads default configuration values
func (m *koanfConfigManager) loadDefaults() error {
	defaults := domain.DefaultConfig()
	
	// Set defaults using koanf
	m.ko.Set("projects_dir", defaults.ProjectsDirectory)
	m.ko.Set("worktrees_dir", defaults.WorktreesDirectory)
	m.ko.Set("default_source_branch", defaults.DefaultSourceBranch)
	m.ko.Set("git_implementation", defaults.GitImplementation)
	
	return nil
}

// loadConfigFile loads configuration from TOML file
func (m *koanfConfigManager) loadConfigFile() error {
	configPath := m.getConfigFilePath()
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, that's okay
		return nil
	}
	
	// Load TOML file
	if err := m.ko.Load(file.Provider(configPath), toml.Parser()); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}
	
	return nil
}

// loadEnvironmentVariables loads configuration from environment variables
func (m *koanfConfigManager) loadEnvironmentVariables() error {
	// Load environment variables with TWIGGIT_ prefix
	return m.ko.Load(env.Provider("TWIGGIT_", ".", func(s string) string {
		// Convert TWIGGIT_PROJECTS_DIR to projects_dir
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "TWIGGIT_")), "_", "_")
	}), nil)
}

// getConfigFilePath returns the path to the configuration file
func (m *koanfConfigManager) getConfigFilePath() string {
	// Follow XDG Base Directory specification
	// $HOME/.config/twiggit/config.toml
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		return "config.toml"
	}
	
	return filepath.Join(home, ".config", "twiggit", "config.toml")
}
```

### Step 3: Create Configuration Tests

#### 3.1 Unit Tests for ConfigManager

**File**: `internal/infrastructure/config/manager_test.go`

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/amaury/twiggit/internal/domain"
)

func TestConfigManager_Load_Defaults(t *testing.T) {
	manager := NewConfigManager()
	
	config, err := manager.Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Verify defaults are loaded
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, "Projects"), config.ProjectsDirectory)
	assert.Equal(t, filepath.Join(home, "Workspaces"), config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
	assert.Equal(t, "go-git", config.GitImplementation)
}

func TestConfigManager_Load_ConfigFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.toml")
	
	configContent := `
projects_dir = "/custom/projects"
worktrees_dir = "/custom/workspaces"
default_source_branch = "develop"
git_implementation = "system-git"
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Create manager with mocked config path
	manager := &koanfConfigManager{
		ko: koanf.New("."),
	}
	
	// Mock the config file path
	manager.ko.Set("projects_dir", "/custom/projects")
	manager.ko.Set("worktrees_dir", "/custom/workspaces")
	manager.ko.Set("default_source_branch", "develop")
	manager.ko.Set("git_implementation", "system-git")
	
	config := &domain.Config{}
	err = manager.ko.Unmarshal("", config)
	require.NoError(t, err)
	
	// Verify config file values are loaded
	assert.Equal(t, "/custom/projects", config.ProjectsDirectory)
	assert.Equal(t, "/custom/workspaces", config.WorktreesDirectory)
	assert.Equal(t, "develop", config.DefaultSourceBranch)
	assert.Equal(t, "system-git", config.GitImplementation)
}

func TestConfigManager_Load_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("TWIGGIT_PROJECTS_DIR", "/env/projects")
	os.Setenv("TWIGGIT_WORKTREES_DIR", "/env/workspaces")
	os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", "env-main")
	defer func() {
		os.Unsetenv("TWIGGIT_PROJECTS_DIR")
		os.Unsetenv("TWIGGIT_WORKTREES_DIR")
		os.Unsetenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
	}()
	
	manager := NewConfigManager()
	
	config, err := manager.Load()
	require.NoError(t, err)
	
	// Verify environment variables override defaults
	assert.Equal(t, "/env/projects", config.ProjectsDirectory)
	assert.Equal(t, "/env/workspaces", config.WorktreesDirectory)
	assert.Equal(t, "env-main", config.DefaultSourceBranch)
}

func TestConfigManager_ValidateConfig(t *testing.T) {
	manager := NewConfigManager()
	
	t.Run("valid config", func(t *testing.T) {
		config := &domain.Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "go-git",
		}
		
		err := manager.ValidateConfig(config)
		assert.NoError(t, err)
	})
	
	t.Run("invalid relative paths", func(t *testing.T) {
		config := &domain.Config{
			ProjectsDirectory:  "relative/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "go-git",
		}
		
		err := manager.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
	})
	
	t.Run("invalid git implementation", func(t *testing.T) {
		config := &domain.Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "invalid-git",
		}
		
		err := manager.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git_implementation must be one of")
	})
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
```

#### 3.2 Integration Tests for Configuration Loading

**File**: `internal/infrastructure/config/integration_test.go`

```go
//go:build integration
// +build integration

package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	
	"github.com/amaury/twiggit/internal/domain"
)

var _ = ginkgo.Describe("Configuration Integration", func() {
	var (
		manager infrastructure.ConfigManager
		tempDir string
	)
	
	ginkgo.BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "twiggit-config-test")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		
		manager = NewConfigManager()
	})
	
	ginkgo.AfterEach(func() {
		os.RemoveAll(tempDir)
	})
	
	ginkgo.Context("with config file", func() {
		ginkgo.It("loads configuration from file", func() {
			configPath := filepath.Join(tempDir, "config.toml")
			configContent := `
projects_dir = "/test/projects"
worktrees_dir = "/test/workspaces"
default_source_branch = "develop"
git_implementation = "system-git"
`
			
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			
			// Set XDG_CONFIG_HOME to temp directory
			os.Setenv("XDG_CONFIG_HOME", tempDir)
			defer os.Unsetenv("XDG_CONFIG_HOME")
			
			config, err := manager.Load()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(config.ProjectsDirectory).To(gomega.Equal("/test/projects"))
			gomega.Expect(config.WorktreesDirectory).To(gomega.Equal("/test/workspaces"))
			gomega.Expect(config.DefaultSourceBranch).To(gomega.Equal("develop"))
			gomega.Expect(config.GitImplementation).To(gomega.Equal("system-git"))
		})
	})
	
	ginkgo.Context("with environment variables", func() {
		ginkgo.It("overrides config file with environment variables", func() {
			// Create config file
			configPath := filepath.Join(tempDir, "config.toml")
			configContent := `
projects_dir = "/file/projects"
default_source_branch = "file-main"
`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			
			// Set environment variables
			os.Setenv("TWIGGIT_PROJECTS_DIR", "/env/projects")
			os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", "env-main")
			defer func() {
				os.Unsetenv("TWIGGIT_PROJECTS_DIR")
				os.Unsetenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
			}()
			
			config, err := manager.Load()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			
			// Environment variables should override config file
			gomega.Expect(config.ProjectsDirectory).To(gomega.Equal("/env/projects"))
			gomega.Expect(config.DefaultSourceBranch).To(gomega.Equal("env-main"))
		})
	})
	
	ginkgo.Context("validation", func() {
		ginkgo.It("validates configuration on load", func() {
			// Create invalid config file
			configPath := filepath.Join(tempDir, "config.toml")
			configContent := `
projects_dir = "relative/path"
git_implementation = "invalid"
`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			
			_, err = manager.Load()
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("configuration validation failed"))
		})
	})
})

func TestConfigurationIntegration(t *testing.T) {
	ginkgo.RunSpecs(t, "Configuration Integration Suite")
}
```

### Step 4: Add Configuration to Root Command

#### 4.1 Update Root Command to Use Configuration

**File**: `cmd/root.go` (modify existing file)

```go
package cmd

import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
	
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
)

var (
	configManager infrastructure.ConfigManager
	appConfig     *domain.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "twiggit",
	Short: "A pragmatic tool for managing git worktrees",
	Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees 
across multiple projects.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration before running any command
		var err error
		appConfig, err = configManager.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Initialize configuration manager
	configManager = config.NewConfigManager()
	
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("projects-dir", "", "Projects directory (overrides config)")
	rootCmd.PersistentFlags().String("worktrees-dir", "", "Worktrees directory (overrides config)")
	rootCmd.PersistentFlags().String("default-source-branch", "", "Default source branch for create command")
	rootCmd.PersistentFlags().String("git-implementation", "", "Git implementation to use (go-git, system-git)")
	
	// Bind flags to viper/koanf (will be processed after config loading)
	// This allows command flags to override all other configuration sources
}
```

### Step 5: Add Configuration Validation Tests

#### 5.1 Configuration Validation Tests

**File**: `internal/domain/config_test.go`

```go
package domain

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	require.NotNil(t, config)
	assert.NotEmpty(t, config.ProjectsDirectory)
	assert.NotEmpty(t, config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
	assert.Equal(t, "go-git", config.GitImplementation)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "go-git",
		}
		
		err := config.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("invalid projects directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "relative/path",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "go-git",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
	})
	
	t.Run("invalid worktrees directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "relative/path",
			DefaultSourceBranch: "main",
			GitImplementation: "go-git",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
	})
	
	t.Run("empty default source branch", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "",
			GitImplementation: "go-git",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
	})
	
	t.Run("invalid git implementation", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/workspaces",
			DefaultSourceBranch: "main",
			GitImplementation: "invalid-git",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git_implementation must be one of")
	})
	
	t.Run("multiple validation errors", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "relative/path",
			WorktreesDirectory: "another/relative/path",
			DefaultSourceBranch: "",
			GitImplementation: "invalid-git",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
		// Should contain all validation errors
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
		assert.Contains(t, err.Error(), "git_implementation must be one of")
	})
}
```

### Step 6: Add Helper Functions

#### 6.1 Utility Functions

**File**: `internal/domain/config.go` (add to existing file)

```go
// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
```

### Step 7: Update Dependencies

#### 7.1 Add Koanf Dependency

**File**: `go.mod` (add to existing dependencies)

```go
require (
	github.com/knadh/koanf/v2 v2.1.1
	github.com/knadh/koanf/parsers/toml v2.1.1
	github.com/knadh/koanf/providers/env v2.1.1
	github.com/knadh/koanf/providers/file v2.1.1
)
```

### Step 8: Documentation and Examples

#### 8.1 Configuration Documentation

**File**: `docs/configuration.md`

```markdown
# Configuration

twiggit uses a hierarchical configuration system with the following priority order:

1. **Defaults** - Built-in default values
2. **Config File** - TOML configuration file
3. **Environment Variables** - TWIGGIT_ prefixed variables
4. **Command Flags** - Command-line flags (highest priority)

## Configuration File

The configuration file is located at `$HOME/.config/twiggit/config.toml` following the XDG Base Directory specification.

### Example Configuration

```toml
# Directory paths
projects_dir = "/custom/path/to/projects"
worktrees_dir = "/custom/path/to/workspaces"

# Default behavior
default_source_branch = "develop"

# Git implementation (go-git or system-git)
git_implementation = "go-git"
```

## Environment Variables

All configuration options can be overridden using environment variables with the `TWIGGIT_` prefix:

- `TWIGGIT_PROJECTS_DIR` - Override projects directory
- `TWIGGIT_WORKTREES_DIR` - Override worktrees directory
- `TWIGGIT_DEFAULT_SOURCE_BRANCH` - Override default source branch
- `TWIGGIT_GIT_IMPLEMENTATION` - Override git implementation

Example:
```bash
export TWIGGIT_PROJECTS_DIR="/my/projects"
export TWIGGIT_DEFAULT_SOURCE_BRANCH="develop"
```

## Command Flags

Command flags provide the highest priority and override all other configuration sources:

```bash
twiggit --projects-dir="/my/projects" --default-source-branch="develop" list
```

## Validation

All configuration values are validated on startup:

- Directory paths must be absolute
- Default source branch cannot be empty
- Git implementation must be one of: go-git, system-git
```

## Implementation Checklist

### Core Implementation
- [ ] Define Config domain type with validation
- [ ] Create ConfigManager interface
- [ ] Implement Koanf-based ConfigManager
- [ ] Add priority loading (defaults → file → env → flags)
- [ ] Implement configuration validation
- [ ] Ensure configuration immutability after loading

### Testing
- [ ] Unit tests for ConfigManager
- [ ] Unit tests for configuration validation
- [ ] Integration tests for configuration loading
- [ ] Tests for priority override behavior
- [ ] Tests for error handling and validation

### Integration
- [ ] Update root command to use configuration
- [ ] Add command flags for configuration override
- [ ] Ensure configuration is loaded before command execution
- [ ] Add proper error handling for configuration failures

### Documentation
- [ ] Configuration documentation with examples
- [ ] Environment variable reference
- [ ] Command flag documentation
- [ ] Validation rules documentation

### Quality
- [ ] Code follows Go best practices
- [ ] All functions have godoc comments
- [ ] Error messages are actionable and clear
- [ ] Configuration loading is optimized for performance
- [ ] All linting checks pass

## Success Criteria

1. **Configuration Loading**: Configuration loads correctly from all sources in priority order
2. **Validation**: All configuration values are properly validated with clear error messages
3. **Immutability**: Configuration cannot be modified after loading
4. **Performance**: Configuration loading completes in <10ms
5. **Test Coverage**: >90% test coverage for configuration system
6. **Error Handling**: Clear, actionable error messages for all failure scenarios
7. **Documentation**: Complete documentation with examples and validation rules

## Next Steps

After implementing the configuration system:

1. Implement the context detection system
2. Create the core git worktree management functionality
3. Implement CLI commands with configuration integration
4. Add shell integration and completion
5. Implement comprehensive testing suite

This configuration system provides a solid foundation for the rest of the twiggit application, ensuring consistent behavior across different environments and use cases.