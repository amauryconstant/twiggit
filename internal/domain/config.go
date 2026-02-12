package domain

import (
	"os"
	"path/filepath"
	"time"
)

// ContextDetectionConfig represents context detection specific configuration
type ContextDetectionConfig struct {
	// Cache TTL for context detection results
	CacheTTL string `toml:"cache_ttl" koanf:"cache_ttl"`

	// Timeout for git operations during context detection
	GitOperationTimeout string `toml:"git_operation_timeout" koanf:"git_operation_timeout"`

	// Enable git repository validation during context detection
	EnableGitValidation bool `toml:"enable_git_validation" koanf:"enable_git_validation"`
}

// GitConfig represents git operations specific configuration
type GitConfig struct {
	// Timeout for CLI git operations in seconds
	CLITimeout int `toml:"cli_timeout" koanf:"cli_timeout"`

	// Enable caching for git operations
	CacheEnabled bool `toml:"cache_enabled" koanf:"cache_enabled"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	CacheEnabled  bool          `toml:"cache_enabled" koanf:"cache_enabled"`
	CacheTTL      time.Duration `toml:"cache_ttl" koanf:"cache_ttl"`
	ConcurrentOps bool          `toml:"concurrent_operations" koanf:"concurrent_operations"`
	MaxConcurrent int           `toml:"max_concurrent" koanf:"max_concurrent"`
}

// ValidationConfig holds validation-specific configuration
type ValidationConfig struct {
	StrictBranchNames    bool `toml:"strict_branch_names" koanf:"strict_branch_names"`
	RequireCleanWorktree bool `toml:"require_clean_worktree" koanf:"require_clean_worktree"`
	AllowForceDelete     bool `toml:"allow_force_delete" koanf:"allow_force_delete"`
}

// NavigationConfig holds navigation-specific configuration
type NavigationConfig struct {
	EnableSuggestions bool `toml:"enable_suggestions" koanf:"enable_suggestions"`
	MaxSuggestions    int  `toml:"max_suggestions" koanf:"max_suggestions"`
	FuzzyMatching     bool `toml:"fuzzy_matching" koanf:"fuzzy_matching"`
}

// ShellConfigFiles represents shell-specific configuration file paths
type ShellConfigFiles struct {
	Bash string `toml:"bash" koanf:"bash"`
	Zsh  string `toml:"zsh" koanf:"zsh"`
	Fish string `toml:"fish" koanf:"fish"`
}

// ShellWrapperConfig represents shell wrapper specific configuration
type ShellWrapperConfig struct {
	// Enable shell wrapper functionality
	Enabled bool `toml:"enabled" koanf:"enabled"`

	// Auto-detect shell type
	AutoDetect bool `toml:"auto_detect" koanf:"auto_detect"`

	// Default shell type if auto-detection fails
	DefaultShell string `toml:"default_shell" koanf:"default_shell"`

	// Configuration file paths for each shell
	ConfigFiles ShellConfigFiles `toml:"config_files" koanf:"config_files"`

	// Enable backup of existing configuration files
	BackupEnabled bool `toml:"backup_enabled" koanf:"backup_enabled"`

	// Backup directory for configuration file backups
	BackupDir string `toml:"backup_dir" koanf:"backup_dir"`
}

// ShellConfig represents shell integration specific configuration
type ShellConfig struct {
	// Shell wrapper configuration
	Wrapper ShellWrapperConfig `toml:"wrapper" koanf:"wrapper"`

	// Enable shell integration features
	Enabled bool `toml:"enabled" koanf:"enabled"`

	// Timeout for shell operations in seconds
	Timeout int `toml:"timeout" koanf:"timeout"`
}

// Config represents the complete application configuration
type Config struct {
	// Directory paths
	ProjectsDirectory  string `toml:"projects_dir" koanf:"projects_dir"`
	WorktreesDirectory string `toml:"worktrees_dir" koanf:"worktrees_dir"`

	// Default principal branch
	DefaultSourceBranch string `toml:"default_source_branch" koanf:"default_source_branch"`

	// Context detection settings
	ContextDetection ContextDetectionConfig `toml:"context_detection" koanf:"context_detection"`

	// Git operations settings
	Git GitConfig `toml:"git" koanf:"git"`

	// Service settings
	Services ServiceConfig `toml:"services" koanf:"services"`

	// Validation settings
	Validation ValidationConfig `toml:"validation" koanf:"validation"`

	// Navigation settings
	Navigation NavigationConfig `toml:"navigation" koanf:"navigation"`

	// Shell integration settings
	Shell ShellConfig `toml:"shell" koanf:"shell"`
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
		ContextDetection: ContextDetectionConfig{
			CacheTTL:            "5m",
			GitOperationTimeout: "30s",
			EnableGitValidation: true,
		},
		Git: GitConfig{
			CLITimeout:   30,
			CacheEnabled: true,
		},
		Services: ServiceConfig{
			CacheEnabled:  true,
			CacheTTL:      5 * time.Minute,
			ConcurrentOps: true,
			MaxConcurrent: 4,
		},
		Validation: ValidationConfig{
			StrictBranchNames:    true,
			RequireCleanWorktree: true,
			AllowForceDelete:     false,
		},
		Navigation: NavigationConfig{
			EnableSuggestions: true,
			MaxSuggestions:    10,
			FuzzyMatching:     false,
		},
		Shell: ShellConfig{
			Enabled: true,
			Timeout: 30,
			Wrapper: ShellWrapperConfig{
				Enabled:      true,
				AutoDetect:   true,
				DefaultShell: "bash",
				ConfigFiles: ShellConfigFiles{
					Bash: "~/.bashrc",
					Zsh:  "~/.zshrc",
					Fish: "~/.config/fish/config.fish",
				},
				BackupEnabled: true,
				BackupDir:     "~/.config/twiggit/backups",
			},
		},
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var validationErrors []string

	// Validate projects directory
	if !filepath.IsAbs(c.ProjectsDirectory) {
		validationErrors = append(validationErrors, "projects_directory must be absolute path")
	}

	// Validate worktrees directory
	if !filepath.IsAbs(c.WorktreesDirectory) {
		validationErrors = append(validationErrors, "worktrees_directory must be absolute path")
	}

	// Validate default source branch
	if c.DefaultSourceBranch == "" {
		validationErrors = append(validationErrors, "default_source_branch cannot be empty")
	}

	if len(validationErrors) > 0 {
		return NewValidationError("Config.Validate", "validation", "", "config validation failed").
			WithSuggestions(validationErrors)
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
