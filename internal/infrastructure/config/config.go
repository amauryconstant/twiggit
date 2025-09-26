// Package config handles configuration management for twiggit
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"

	"github.com/amaury/twiggit/internal/infrastructure"
)

// Config represents the application configuration
type Config struct {
	// ProjectsPath is the directory containing git project repositories
	ProjectsPath string `koanf:"projects_path"`
	// WorkspacesPath is the directory containing worktree checkouts
	WorkspacesPath string `koanf:"workspaces_path"`
	// Project is the currently active project name
	Project string `koanf:"project"`
	// DefaultSourceBranch is the default source branch for creating worktrees
	DefaultSourceBranch string `koanf:"default_source_branch"`
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
		ProjectsPath:        filepath.Join(home, "Projects"),
		WorkspacesPath:      filepath.Join(home, "Workspaces"),
		Project:             "",
		DefaultSourceBranch: "", // Empty by default, will fallback to "main" in CLI
		Verbose:             false,
		Quiet:               false,
	}
}

// fileProvider implements a koanf provider that uses our FileSystem interface
type fileProvider struct {
	fs   infrastructure.FileSystem
	path string
}

// NewFileProvider creates a new file provider that uses our FileSystem interface
func NewFileProvider(fs infrastructure.FileSystem, path string) koanf.Provider {
	return &fileProvider{
		fs:   fs,
		path: path,
	}
}

// ReadBytes reads the configuration file bytes
func (p *fileProvider) ReadBytes() ([]byte, error) {
	data, err := p.fs.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return data, nil
}

// Read returns the raw configuration data
func (p *fileProvider) Read() (map[string]interface{}, error) {
	// Return empty map to let koanf handle the parsing with the TOML parser
	// The TOML parser will use the ReadBytes method internally
	return map[string]interface{}{}, nil
}

// Option is a functional option for configuring Config loading
type Option func(*configLoader)

// WithFileSystem sets a custom filesystem for config loading
func WithFileSystem(fs infrastructure.FileSystem) Option {
	return func(cl *configLoader) {
		cl.fileSystem = fs
	}
}

// configLoader handles configuration loading with configurable dependencies
type configLoader struct {
	fileSystem infrastructure.FileSystem
}

// LoadConfig loads configuration from files and environment variables with optional configuration
func LoadConfig(opts ...Option) (*Config, error) {
	loader := &configLoader{
		fileSystem: infrastructure.NewOSFileSystem(),
	}

	// Apply options
	for _, opt := range opts {
		opt(loader)
	}

	return loader.load()
}

// load handles the actual configuration loading logic
func (cl *configLoader) load() (*Config, error) {
	config := NewConfig()
	k := koanf.New(".")

	// Load from configuration files (lower priority)
	configPaths := config.getConfigPaths()
	for _, configPath := range configPaths {
		if _, err := cl.fileSystem.Stat(configPath); err == nil {
			if err := k.Load(NewFileProvider(cl.fileSystem, configPath), toml.Parser()); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}
			break // Use first found config file
		}
	}

	// Load from environment variables (higher priority)
	if err := k.Load(env.Provider("TWIGGIT_", ".", func(s string) string {
		// Convert TWIGGIT_WORKSPACES_PATH -> workspaces_path (lowercase)
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
	// Ensure ProjectsPath is set (fallback to sibling of WorkspacesPath)
	c.ensureProjectsPathFallback()

	if c.WorkspacesPath == "" {
		return errors.New("workspaces path cannot be empty")
	}

	if c.ProjectsPath == "" {
		return errors.New("projects path cannot be empty")
	}

	if c.Verbose && c.Quiet {
		return errors.New("verbose and quiet cannot both be enabled")
	}

	// Validate default_source_branch if provided
	if c.DefaultSourceBranch != "" {
		// Git branch names should not contain spaces, special characters, or start with -
		if !isValidBranchName(c.DefaultSourceBranch) {
			return fmt.Errorf("invalid default source branch name: %s", c.DefaultSourceBranch)
		}
	}

	return nil
}

// ensureProjectsPathFallback ensures ProjectsPath is set (fallback to sibling of WorkspacesPath)
func (c *Config) ensureProjectsPathFallback() {
	// Ensure ProjectsPath is set (fallback to sibling of WorkspacesPath)
	if c.ProjectsPath == "" {
		parent := filepath.Dir(c.WorkspacesPath)
		c.ProjectsPath = filepath.Join(parent, "Projects")
	}
}

// getConfigPaths returns the list of possible configuration file paths
// in order of preference (XDG-compliant)
func (c *Config) getConfigPaths() []string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}

	paths := make([]string, 0, 3)

	// XDG_CONFIG_HOME/twiggit/config.toml
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		paths = append(paths, filepath.Join(xdgConfigHome, "twiggit", "config.toml"))
	}

	// ~/.config/twiggit/config.toml
	paths = append(paths, filepath.Join(home, ".config", "twiggit", "config.toml"))

	// ~/.twiggit.toml (legacy)
	paths = append(paths, filepath.Join(home, ".twiggit.toml"))

	return paths
}

// isValidBranchName checks if a branch name follows Git branch naming rules
func isValidBranchName(name string) bool {
	// Git branch names cannot:
	// - Start with a dot
	// - Start with a dash
	// - Contain spaces
	// - Contain consecutive dots
	// - End with a dot or slash
	// - Contain invalid characters: ~^:?*[
	// - Be empty
	// - Be "HEAD" (reserved)

	if name == "" || name == "HEAD" {
		return false
	}

	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return false
	}

	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, "/") {
		return false
	}

	if strings.Contains(name, " ") || strings.Contains(name, "..") {
		return false
	}

	// Check for invalid Git characters
	invalidChars := "~^:?*[@#"
	for _, char := range name {
		if strings.ContainsRune(invalidChars, char) {
			return false
		}
	}

	return true
}
