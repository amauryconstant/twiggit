package domain

import (
	"errors"
	"os"
	"path/filepath"
)

// Config represents the complete application configuration
type Config struct {
	// Directory paths
	ProjectsDirectory  string `toml:"projects_dir" koanf:"projects_dir"`
	WorktreesDirectory string `toml:"worktrees_dir" koanf:"worktrees_dir"`

	// Default principal branch
	DefaultSourceBranch string `toml:"default_source_branch" koanf:"default_source_branch"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory can't be determined
		if cwd, err := os.Getwd(); err == nil {
			home = cwd
		} else {
			// Last resort - use /tmp as fallback
			home = "/tmp"
		}
	}
	return &Config{
		ProjectsDirectory:   filepath.Join(home, "Projects"),
		WorktreesDirectory:  filepath.Join(home, "Worktrees"),
		DefaultSourceBranch: "main",
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var validationErrors []error

	// Validate projects directory
	if !filepath.IsAbs(c.ProjectsDirectory) {
		validationErrors = append(validationErrors, errors.New("projects_directory must be absolute path"))
	}

	// Validate worktrees directory
	if !filepath.IsAbs(c.WorktreesDirectory) {
		validationErrors = append(validationErrors, errors.New("worktrees_directory must be absolute path"))
	}

	// Validate default source branch
	if c.DefaultSourceBranch == "" {
		validationErrors = append(validationErrors, errors.New("default_source_branch cannot be empty"))
	}

	if len(validationErrors) > 0 {
		var errorMsg string
		for _, err := range validationErrors {
			errorMsg += err.Error() + "; "
		}
		return errors.New("config validation failed: " + errorMsg)
	}

	return nil
}

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from defaults and config file
	Load() (*Config, error)

	// GetConfig returns the loaded configuration (immutable after Load)
	GetConfig() *Config
}

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
		ProjectsDirectory:   globalConfig.ProjectsDirectory,
		WorktreesDirectory:  globalConfig.WorktreesDirectory,
		DefaultSourceBranch: globalConfig.DefaultSourceBranch,
	}
}
