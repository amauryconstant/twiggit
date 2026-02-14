package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"twiggit/internal/domain"
)

// Pure functions extracted from ConfigManager

// resolveConfigPath returns the path to the configuration file following XDG Base Directory specification
func resolveConfigPath(xdgConfigHome, homeDir string) string {
	// Check XDG_CONFIG_HOME first
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "twiggit", "config.toml")
	}

	// Fallback to $HOME/.config
	if homeDir != "" {
		return filepath.Join(homeDir, ".config", "twiggit", "config.toml")
	}

	// Fallback to current directory if home directory can't be determined
	return "config.toml"
}

// buildDefaultConfig creates a new default configuration
func buildDefaultConfig() *domain.Config {
	return domain.DefaultConfig()
}

// configFileExists checks if a file exists at the given path
func configFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// validateConfig validates a configuration object
func validateConfig(config *domain.Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

// copyConfig creates a deep copy of a configuration object
func copyConfig(config *domain.Config) *domain.Config {
	return &domain.Config{
		ProjectsDirectory:   config.ProjectsDirectory,
		WorktreesDirectory:  config.WorktreesDirectory,
		DefaultSourceBranch: config.DefaultSourceBranch,
		ContextDetection:    config.ContextDetection,
		Git:                 config.Git,
		Services:            config.Services,
		Validation:          config.Validation,
		Navigation:          config.Navigation,
		Shell:               config.Shell,
	}
}

type koanfConfigManager struct {
	ko     *koanf.Koanf
	config *domain.Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() domain.ConfigManager {
	return &koanfConfigManager{
		ko: koanf.New("."),
	}
}

// Load loads configuration from defaults and config file
func (m *koanfConfigManager) Load() (*domain.Config, error) {
	// 1. Load defaults
	if err := m.loadDefaults(); err != nil {
		return nil, domain.NewConfigError("", "failed to load default configuration", err)
	}

	// 2. Load config file using pure function for existence check
	configPath := m.getConfigFilePath()
	if configFileExists(configPath) {
		// Load TOML file
		if err := m.ko.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return nil, domain.NewConfigError(configPath, "failed to parse config file", err)
		}
	}

	// 3. Unmarshal to config object
	config := &domain.Config{}
	if err := m.ko.Unmarshal("", config); err != nil {
		return nil, domain.NewConfigError(configPath, "failed to unmarshal configuration", err)
	}

	// 4. Validate configuration using pure function
	if err := validateConfig(config); err != nil {
		return nil, domain.NewConfigError(configPath, "validation failed", err)
	}

	// 5. Store immutable config
	m.config = config

	// 6. Return a copy using pure function to maintain immutability
	return copyConfig(config), nil
}

// GetConfig returns the loaded configuration (immutable copy)
func (m *koanfConfigManager) GetConfig() *domain.Config {
	if m.config == nil {
		return nil
	}
	// Return a deep copy using pure function to maintain immutability
	return copyConfig(m.config)
}

// loadDefaults loads default configuration values
func (m *koanfConfigManager) loadDefaults() error {
	defaults := buildDefaultConfig()

	// Set defaults using koanf (matching the struct tags)
	if err := m.ko.Set("projects_dir", defaults.ProjectsDirectory); err != nil {
		return fmt.Errorf("failed to set projects_dir default: %w", err)
	}
	if err := m.ko.Set("worktrees_dir", defaults.WorktreesDirectory); err != nil {
		return fmt.Errorf("failed to set worktrees_dir default: %w", err)
	}
	if err := m.ko.Set("default_source_branch", defaults.DefaultSourceBranch); err != nil {
		return fmt.Errorf("failed to set default_source_branch default: %w", err)
	}

	// Set nested structure defaults
	if err := m.ko.Set("context_detection.cache_ttl", defaults.ContextDetection.CacheTTL); err != nil {
		return fmt.Errorf("failed to set context_detection.cache_ttl default: %w", err)
	}
	if err := m.ko.Set("context_detection.git_operation_timeout", defaults.ContextDetection.GitOperationTimeout); err != nil {
		return fmt.Errorf("failed to set context_detection.git_operation_timeout default: %w", err)
	}
	if err := m.ko.Set("context_detection.enable_git_validation", defaults.ContextDetection.EnableGitValidation); err != nil {
		return fmt.Errorf("failed to set context_detection.enable_git_validation default: %w", err)
	}

	if err := m.ko.Set("git.cli_timeout", defaults.Git.CLITimeout); err != nil {
		return fmt.Errorf("failed to set git.cli_timeout default: %w", err)
	}
	if err := m.ko.Set("git.cache_enabled", defaults.Git.CacheEnabled); err != nil {
		return fmt.Errorf("failed to set git.cache_enabled default: %w", err)
	}
	return nil
}

// getConfigFilePath returns the path to the configuration file following XDG Base Directory specification
func (m *koanfConfigManager) getConfigFilePath() string {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	home, _ := os.UserHomeDir()
	return resolveConfigPath(xdgHome, home)
}
