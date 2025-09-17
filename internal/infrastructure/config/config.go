// Package config handles configuration management for twiggit
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config represents the application configuration
type Config struct {
	// Workspace is the root directory containing all projects
	Workspace string `koanf:"workspace"`
	// Project is the currently active project name
	Project string `koanf:"project"`
	// Verbose enables detailed logging
	Verbose bool `koanf:"verbose"`
	// Quiet suppresses non-essential output
	Quiet bool `koanf:"quiet"`
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME") // fallback
	}

	return &Config{
		Workspace: filepath.Join(home, "Workspaces"),
		Project:   "",
		Verbose:   false,
		Quiet:     false,
	}
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig() (*Config, error) {
	config := NewConfig()
	k := koanf.New(".")

	// Load from configuration files (lower priority)
	configPaths := config.getConfigPaths()
	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}
			break // Use first found config file
		}
	}

	// Load from environment variables (higher priority)
	if err := k.Load(env.Provider("TWIGGIT_", ".", func(s string) string {
		// Convert TWIGGIT_WORKSPACE -> workspace (lowercase)
		return strings.ToLower(s[8:]) // Remove "TWIGGIT_" prefix and lowercase
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Unmarshal into config struct
	if err := k.Unmarshal("", config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Workspace == "" {
		return errors.New("workspace path cannot be empty")
	}

	if c.Verbose && c.Quiet {
		return errors.New("verbose and quiet cannot both be enabled")
	}

	return nil
}

// getConfigPaths returns the list of possible configuration file paths
// in order of preference (XDG-compliant)
func (c *Config) getConfigPaths() []string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}

	paths := make([]string, 0, 3)

	// XDG_CONFIG_HOME/twiggit/config.yaml
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, "twiggit", "config.yaml"))
	}

	// ~/.config/twiggit/config.yaml
	paths = append(paths, filepath.Join(home, ".config", "twiggit", "config.yaml"))

	// ~/.twiggit.yaml (legacy)
	paths = append(paths, filepath.Join(home, ".twiggit.yaml"))

	return paths
}
