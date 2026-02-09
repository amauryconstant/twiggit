//go:build e2e
// +build e2e

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pelletier/go-toml"
	"twiggit/internal/domain"
)

// ConfigHelper provides configuration management utilities for E2E tests
type ConfigHelper struct {
	tempDir       string
	configDir     string
	configPath    string
	projectsDir   string
	worktreesDir  string
	defaultBranch string
	content       strings.Builder
}

// NewConfigHelper creates a new ConfigHelper instance
func NewConfigHelper() *ConfigHelper {
	tempDir := GinkgoT().TempDir()
	configDir := filepath.Join(tempDir, "twiggit")

	return &ConfigHelper{
		tempDir:       tempDir,
		configDir:     configDir,
		configPath:    filepath.Join(configDir, "config.toml"),
		projectsDir:   filepath.Join(tempDir, "projects"),
		worktreesDir:  filepath.Join(tempDir, "worktrees"),
		defaultBranch: "main",
	}
}

// WithTempDir sets the temp directory for this ConfigHelper
func (c *ConfigHelper) WithTempDir(tempDir string) *ConfigHelper {
	c.tempDir = tempDir
	c.configDir = filepath.Join(tempDir, "twiggit")
	c.configPath = filepath.Join(c.configDir, "config.toml")
	c.projectsDir = filepath.Join(tempDir, "projects")
	c.worktreesDir = filepath.Join(tempDir, "worktrees")
	return c
}

// WithProjectsDir sets the projects directory
func (c *ConfigHelper) WithProjectsDir(path string) *ConfigHelper {
	c.projectsDir = path
	return c
}

// WithWorktreesDir sets the worktrees directory
func (c *ConfigHelper) WithWorktreesDir(path string) *ConfigHelper {
	c.worktreesDir = path
	return c
}

// WithDefaultSourceBranch sets the default source branch
func (c *ConfigHelper) WithDefaultSourceBranch(branch string) *ConfigHelper {
	c.defaultBranch = branch
	return c
}

// WithCustomConfig adds custom configuration content
func (c *ConfigHelper) WithCustomConfig(content string) *ConfigHelper {
	c.content.WriteString(content)
	c.content.WriteString("\n")
	return c
}

// Build creates the configuration file and returns the config directory path
func (c *ConfigHelper) Build() string {
	// Create config directory
	err := os.MkdirAll(c.configDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	// Create projects and worktrees directories
	err = os.MkdirAll(c.projectsDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	err = os.MkdirAll(c.worktreesDir, 0755)
	Expect(err).NotTo(HaveOccurred())

	// Build configuration content
	var configContent strings.Builder

	// Basic configuration
	configContent.WriteString(fmt.Sprintf(`# Test configuration for E2E tests
projects_dir = "%s"
worktrees_dir = "%s"
default_source_branch = "%s"

`, c.projectsDir, c.worktreesDir, c.defaultBranch))

	// Add custom content if provided
	if c.content.Len() > 0 {
		configContent.WriteString(c.content.String())
	}

	// Add context detection configuration
	configContent.WriteString(`
[context_detection]
cache_ttl = "1m"
git_operation_timeout = "10s"
enable_git_validation = true

[git]
cli_timeout = 10
cache_enabled = false

[services]
cache_enabled = false
cache_ttl = "1m"
concurrent_ops = false
max_concurrent = 1
`)

	// Write configuration file
	err = os.WriteFile(c.configPath, []byte(configContent.String()), 0644)
	Expect(err).NotTo(HaveOccurred())

	// Validate TOML syntax
	if _, err := toml.LoadFile(c.configPath); err != nil {
		GinkgoT().Fatalf("Invalid TOML configuration: %v", err)
	}

	// Validate against domain.Config struct
	config := &domain.Config{
		ProjectsDirectory:   c.projectsDir,
		WorktreesDirectory:  c.worktreesDir,
		DefaultSourceBranch: c.defaultBranch,
		ContextDetection: domain.ContextDetectionConfig{
			CacheTTL:            "1m",
			GitOperationTimeout: "10s",
			EnableGitValidation: true,
		},
		Git: domain.GitConfig{
			CLITimeout:   10,
			CacheEnabled: false,
		},
		Services: domain.ServiceConfig{
			CacheEnabled:  false,
			CacheTTL:      time.Minute,
			ConcurrentOps: false,
			MaxConcurrent: 1,
		},
	}

	if err := config.Validate(); err != nil {
		GinkgoT().Fatalf("Config validation failed: %v", err)
	}

	return c.tempDir
}

// GetConfigPath returns the path to the configuration file
func (c *ConfigHelper) GetConfigPath() string {
	return c.configPath
}

// GetConfigDir returns the config directory path (for XDG_CONFIG_HOME)
func (c *ConfigHelper) GetConfigDir() string {
	return c.tempDir
}

// GetProjectsDir returns the projects directory path
func (c *ConfigHelper) GetProjectsDir() string {
	return c.projectsDir
}

// GetWorktreesDir returns the worktrees directory path
func (c *ConfigHelper) GetWorktreesDir() string {
	return c.worktreesDir
}

func (c *ConfigHelper) GetTempDir() string {
	return c.tempDir
}
