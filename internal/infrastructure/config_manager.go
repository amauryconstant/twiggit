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
	m.loadDefaults()

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

	// 6. Return a copy to maintain immutability
	return &domain.Config{
		ProjectsDirectory:   config.ProjectsDirectory,
		WorktreesDirectory:  config.WorktreesDirectory,
		DefaultSourceBranch: config.DefaultSourceBranch,
	}, nil
}

// GetConfig returns the loaded configuration (immutable copy)
func (m *koanfConfigManager) GetConfig() *domain.Config {
	if m.config == nil {
		return nil
	}
	// Return a deep copy to maintain immutability
	return &domain.Config{
		ProjectsDirectory:   m.config.ProjectsDirectory,
		WorktreesDirectory:  m.config.WorktreesDirectory,
		DefaultSourceBranch: m.config.DefaultSourceBranch,
	}
}

// loadDefaults loads default configuration values
func (m *koanfConfigManager) loadDefaults() {
	defaults := domain.DefaultConfig()

	// Set defaults using koanf (matching the struct tags)
	if err := m.ko.Set("projects_dir", defaults.ProjectsDirectory); err != nil {
		// Koanf Set only returns error for invalid keys, which shouldn't happen with our keys
		panic(fmt.Errorf("failed to set projects_dir default: %w", err))
	}
	if err := m.ko.Set("worktrees_dir", defaults.WorktreesDirectory); err != nil {
		panic(fmt.Errorf("failed to set worktrees_dir default: %w", err))
	}
	if err := m.ko.Set("default_source_branch", defaults.DefaultSourceBranch); err != nil {
		panic(fmt.Errorf("failed to set default_source_branch default: %w", err))
	}
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
