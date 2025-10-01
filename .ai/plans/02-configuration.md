# Configuration System Implementation Plan

## Overview

This plan implements a simple Koanf-based configuration management system for twiggit with defaults loading, config file loading, validation, and immutable configuration after loading. The system follows the established foundation layer and integrates with the existing project structure.

## Context from Documentation

> **From implementation.md**: "Configuration SHALL be loaded using Koanf with TOML support. Configuration validation SHALL occur during startup."

> **From technology.md**: "Koanf SHALL load configuration in simple order: defaults → config file."

> **From implementation.md**: "Location: XDG Base Directory specification SHALL be followed for config folders (`$HOME/.config/twiggit/config.toml` or `$XDG_CONFIG_HOME/twiggit/config.toml`). Format: TOML format SHALL be supported exclusively."

## Implementation Steps

### Step 1: Define Configuration Structure and Interfaces

#### 1.1 Create Configuration Domain Types

**File**: `internal/domain/config.go`

```go
package domain

import (
	"fmt"
	"os"
	"path/filepath"
)

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Config represents the complete application configuration
type Config struct {
	// Directory paths
	ProjectsDirectory string `toml:"projects_dir" koanf:"projects_dir"`
	WorktreesDirectory string `toml:"worktrees_dir" koanf:"worktrees_dir"`
	
	// Default principal branch
	DefaultSourceBranch string `toml:"default_source_branch" koanf:"default_source_branch"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		home = "."
	}
	return &Config{
		ProjectsDirectory:  filepath.Join(home, "Projects"),
		WorktreesDirectory: filepath.Join(home, "Worktrees"),
		DefaultSourceBranch: "main",
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var errors []error
	
	// Validate projects directory
	if !filepath.IsAbs(c.ProjectsDirectory) {
		errors = append(errors, errors.New("projects_directory must be absolute path"))
	}
	
	// Validate worktrees directory
	if !filepath.IsAbs(c.WorktreesDirectory) {
		errors = append(errors, errors.New("worktrees_directory must be absolute path"))
	}
	
	// Validate default source branch
	if c.DefaultSourceBranch == "" {
		errors = append(errors, errors.New("default_source_branch cannot be empty"))
	}
	
	if len(errors) > 0 {
		return errors.New("config validation failed")
	}
	
	return nil
}
```

#### 1.2 Define ConfigManager Interface

**File**: `internal/domain/config.go`

```go
// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from defaults and config file
	Load() (*Config, error)
	
	// GetConfig returns the loaded configuration (immutable after Load)
	GetConfig() *Config
}
```

### Step 2: Create Infrastructure Layer

#### 2.1 Create Infrastructure Directory

First, create the infrastructure directory that doesn't exist yet:

```bash
mkdir -p internal/infrastructure
```

#### 2.2 Create Infrastructure Interfaces

**File**: `internal/infrastructure/interfaces.go`

```go
package infrastructure

import "twiggit/internal/domain"
}

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from defaults and config file
	Load() (*domain.Config, error)
	
	// GetConfig returns the loaded configuration (immutable after Load)
	GetConfig() *domain.Config
}
```

#### 2.3 Create Concrete Implementation

**File**: `internal/infrastructure/config_manager.go`

```go
package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/v2/parsers/toml"
	"github.com/knadh/koanf/v2/providers/file"
	
	"twiggit/internal/domain"
)

type koanfConfigManager struct {
	ko    *koanf.Koanf
	config *domain.Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() ConfigManager {
	return &koanfConfigManager{
		ko: koanf.New("."),
	}
}

// Load loads configuration from defaults and config file
func (m *koanfConfigManager) Load() (*domain.Config, error) {
	// 1. Load defaults
	if err := m.loadDefaults(); err != nil {
		return nil, fmt.Errorf("load config: failed to load defaults: %w", err)
	}
	
	// 2. Load config file
	if err := m.loadConfigFile(); err != nil {
		return nil, fmt.Errorf("load config: failed to load config file: %w", err)
	}
	
	// 3. Unmarshal to config object
	config := &domain.Config{}
	if err := m.ko.Unmarshal("", config); err != nil {
		return nil, fmt.Errorf("load config: failed to unmarshal configuration: %w", err)
	}
	
	// 4. Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("load config: validation failed: %w", err)
	}
	
	// 5. Store immutable config
	m.config = config
	
	return config, nil
}

// GetConfig returns the loaded configuration (immutable copy)
func (m *koanfConfigManager) GetConfig() *domain.Config {
	if m.config == nil {
		return nil
	}
	// Return a deep copy to maintain immutability
	return &domain.Config{
		ProjectsDirectory:  m.config.ProjectsDirectory,
		WorktreesDirectory: m.config.WorktreesDirectory,
		DefaultSourceBranch: m.config.DefaultSourceBranch,
	}
}

// loadDefaults loads default configuration values
func (m *koanfConfigManager) loadDefaults() error {
	defaults := domain.DefaultConfig()
	
	// Set defaults using koanf
	m.ko.Set("projects_dir", defaults.ProjectsDirectory)
	m.ko.Set("worktrees_dir", defaults.WorktreesDirectory)
	m.ko.Set("default_source_branch", defaults.DefaultSourceBranch)
	
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
		return fmt.Errorf("load config: failed to parse config file %s: %w", configPath, err)
	}
	
	return nil
}

// getConfigFilePath returns the path to the configuration file following XDG Base Directory specification
func (m *koanfConfigManager) getConfigFilePath() string {
	// Check XDG_CONFIG_HOME first
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "twiggit", "config.toml")
	}
	
	// Fallback to $HOME/.config
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

**File**: `internal/infrastructure/config_manager_test.go`

```go
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
```

#### 3.2 Integration Tests for Configuration Loading

**File**: `internal/infrastructure/integration_test.go`

```go
//go:build integration
// +build integration

package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"twiggit/internal/domain"
)

func TestConfigManager_Integration_ConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()
	
	// Create config file
	configPath := filepath.Join(tempDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
worktrees_dir = "/test/worktrees"
default_source_branch = "develop"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	config, err := manager.Load()
	require.NoError(t, err)
	
	assert.Equal(t, "/test/projects", config.ProjectsDirectory)
	assert.Equal(t, "/test/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "develop", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_XDGFallback(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()
	
	// Create config file in .config structure
	configDir := filepath.Join(tempDir, ".config", "twiggit")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	configPath := filepath.Join(configDir, "config.toml")
	configContent := `
projects_dir = "/fallback/projects"
worktrees_dir = "/fallback/worktrees"
default_source_branch = "main"
`
	
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Set HOME to temp directory, but not XDG_CONFIG_HOME
	os.Setenv("HOME", tempDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Unsetenv("HOME")
	
	config, err := manager.Load()
	require.NoError(t, err)
	
	assert.Equal(t, "/fallback/projects", config.ProjectsDirectory)
	assert.Equal(t, "/fallback/worktrees", config.WorktreesDirectory)
	assert.Equal(t, "main", config.DefaultSourceBranch)
}

func TestConfigManager_Integration_Validation(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()
	
	// Create invalid config file
	configPath := filepath.Join(tempDir, "config.toml")
	configContent := `
projects_dir = "relative/path"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	_, err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestConfigManager_Integration_MalformedTOML(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()
	
	// Create malformed TOML file
	configPath := filepath.Join(tempDir, "config.toml")
	configContent := `
projects_dir = "/test/projects"
invalid toml syntax here
worktrees_dir = "/test/worktrees"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	_, err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestConfigManager_Integration_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewConfigManager()
	
	// Set XDG_CONFIG_HOME to empty temp directory (no config file)
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	config, err := manager.Load()
	require.NoError(t, err)
	
	// Should load defaults when no config file exists
	defaultConfig := domain.DefaultConfig()
	assert.Equal(t, defaultConfig.ProjectsDirectory, config.ProjectsDirectory)
	assert.Equal(t, defaultConfig.WorktreesDirectory, config.WorktreesDirectory)
	assert.Equal(t, defaultConfig.DefaultSourceBranch, config.DefaultSourceBranch)
}
```

### Step 4: Add Configuration to Root Command

#### 4.1 Update Root Command to Use Configuration

**File**: `cmd/root.go` (modify existing file)

```go
package cmd

import (
	"fmt"
	
	"github.com/spf13/cobra"
	
	"twiggit/internal/infrastructure"
)

var (
	configManager infrastructure.ConfigManager
	appConfig     *Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "twiggit",
	Short: "A pragmatic tool for managing git worktrees",
	Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees 
across multiple projects.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration (simple approach - no flags)
		var err error
		appConfig, err = configManager.Load()
		if err != nil {
			return fmt.Errorf("cmd: failed to load configuration: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Initialize configuration manager
	configManager = infrastructure.NewConfigManager()
	
	return rootCmd.Execute()
}

// GetConfig returns the global application configuration
func GetConfig() *Config {
	return appConfig
}
```

#### 4.2 Add Global Config Access to Domain

**File**: `internal/domain/config.go` (add to existing file)

```go
// Global configuration accessor (simple approach for Phase 2)
// Note: This will be replaced with dependency injection in later phases

var globalConfig *Config

// SetGlobalConfig sets the global configuration (called from main)
func SetGlobalConfig(config *Config) {
	globalConfig = config
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *Config {
	if globalConfig == nil {
		return DefaultConfig()
	}
	// Return a copy to maintain immutability
	return &Config{
		ProjectsDirectory:  globalConfig.ProjectsDirectory,
		WorktreesDirectory: globalConfig.WorktreesDirectory,
		DefaultSourceBranch: globalConfig.DefaultSourceBranch,
	}
}
```

#### 4.3 Update Main Entry Point

**File**: `main.go` (modify existing file)

```go
package main

import (
	"twiggit/cmd"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

func main() {
	// Initialize and load configuration
	configManager := infrastructure.NewConfigManager()
	config, err := configManager.Load()
	if err != nil {
		// For now, panic on config errors - this will be improved in CLI phase
		panic(err)
	}
	
	// Set global configuration for simple access
	domain.SetGlobalConfig(config)
	
	// Execute CLI
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
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
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/worktrees",
			DefaultSourceBranch: "main",
		}
		
		err := config.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("invalid projects directory", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "relative/path",
			WorktreesDirectory: "/valid/worktrees",
			DefaultSourceBranch: "main",
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
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
	})
	
	t.Run("empty default source branch", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "/valid/projects",
			WorktreesDirectory: "/valid/worktrees",
			DefaultSourceBranch: "",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
	})
	
	t.Run("multiple validation errors", func(t *testing.T) {
		config := &Config{
			ProjectsDirectory:  "relative/path",
			WorktreesDirectory: "another/relative/path",
			DefaultSourceBranch: "",
		}
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
		// Should contain all validation errors
		assert.Contains(t, err.Error(), "projects_directory must be absolute path")
		assert.Contains(t, err.Error(), "worktrees_directory must be absolute path")
		assert.Contains(t, err.Error(), "default_source_branch cannot be empty")
	})
}
```

### Step 5: Update Dependencies

#### 5.1 Add Koanf Dependency

**File**: `go.mod` (add to existing dependencies)

```bash
go get github.com/knadh/koanf/v2@v2.1.1
```

This will add the unified koanf v2 dependency with all necessary sub-packages.

### Step 6: Implementation Checklist

#### 6.1 Core Implementation
- [ ] Create `internal/infrastructure/` directory
- [ ] Define Config domain type with validation in `internal/domain/config.go`
- [ ] Create ConfigManager interface in domain layer
- [ ] Implement Koanf-based ConfigManager in `internal/infrastructure/config_manager.go`
- [ ] Add simple loading (defaults → config file)
- [ ] Implement proper XDG Base Directory support
- [ ] Implement configuration validation
- [ ] Ensure configuration immutability after loading

#### 6.2 Testing
- [ ] Unit tests for ConfigManager using Testify (mocked file operations)
- [ ] Unit tests for configuration validation
- [ ] Integration tests for configuration loading with real files
- [ ] Tests for XDG Base Directory behavior
- [ ] Tests for essential error scenarios (malformed TOML, permission errors)
- [ ] Tests for immutability and deep copy behavior

#### 6.3 Integration
- [ ] Update root command to use configuration
- [ ] Add global config access to domain package
- [ ] Update main.go to load configuration early
- [ ] Ensure configuration is loaded before command execution
- [ ] Add proper error handling for configuration failures

#### 6.4 Dependencies
- [ ] Add unified koanf v2 dependency
- [ ] Verify all imports are correct
- [ ] Ensure project compiles with new dependencies

#### 6.5 Quality
- [ ] Code follows Go best practices
- [ ] All functions have godoc comments
- [ ] Error messages follow existing foundation pattern (errors.New)
- [ ] Configuration loading is optimized for simplicity
- [ ] All linting checks pass
- [ ] Test coverage reaches >80% with `go test -cover ./...`

### Step 7: Success Criteria

1. **Configuration Loading**: Configuration loads correctly from defaults and config file
2. **XDG Compliance**: Proper XDG Base Directory specification implementation
3. **Validation**: All configuration values are properly validated with clear error messages
4. **Immutability**: Configuration cannot be modified after loading (true deep copy)
5. **Test Coverage**: >80% test coverage for configuration system measured with `go test -cover ./...`
6. **Error Handling**: Clear, actionable error messages following existing foundation pattern
7. **Integration**: Simple global config access working with existing main() and CLI structure
8. **Essential Error Scenarios**: Robust handling of malformed TOML, missing files, and validation errors

## Architecture Decisions

### Interface Placement
- **Phase 2**: `ConfigManager` interface in `internal/domain/config.go` for immediate functionality
- **Future**: Interface will remain in domain layer, dependency injection will be added later

### Configuration Loading Strategy
- **Simple Approach**: Load defaults first, then config file if it exists
- **XDG Compliance**: Check `XDG_CONFIG_HOME` first, fallback to `$HOME/.config/`
- **Benefits**: Simple, reliable, follows standards, easy to understand and test

### Project Structure
- **Infrastructure Layer**: Created `internal/infrastructure/` for concrete implementation
- **Domain Layer**: Interface and Config struct in `internal/domain/config.go`
- **Rationale**: Clean separation of concerns, follows established patterns

### Testing Strategy
- **Unit Tests**: Testify with mocked file operations for business logic
- **Integration Tests**: Real file operations with build tags for file system behavior
- **Focus**: Essential error scenarios and XDG compliance, comprehensive coverage

### Global Access Pattern
- **Phase 2**: Simple global config accessor in domain package
- **Future**: Will be replaced with dependency injection when services layer is added
- **Benefits: Immediate functionality, clear migration path

## Next Steps

After implementing the configuration system:

1. Implement the context detection system (Phase 3)
2. Create the core git worktree management functionality (Phase 4)
3. Implement CLI commands with configuration integration (Phase 6)
4. Add shell integration and completion (Phase 7)
5. Implement comprehensive testing suite (Phase 8)

This configuration system provides a solid foundation for the rest of the twiggit application, ensuring consistent behavior across different environments and use cases while following established architectural patterns and maintaining simplicity for Phase 2 scope.

## Next Steps

After implementing the configuration system:

1. Implement the context detection system (Phase 3)
2. Create the core git worktree management functionality (Phase 4)
3. Implement CLI commands with configuration integration (Phase 6)
4. Add shell integration and completion (Phase 7)
5. Implement comprehensive testing suite (Phase 8)

This configuration system provides a solid foundation for the rest of the twiggit application, ensuring consistent behavior across different environments and use cases while following established architectural patterns.
