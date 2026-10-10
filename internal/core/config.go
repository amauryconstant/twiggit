package core

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

	// IsGitValidationEnabled reports whether context detection validates git repositories.
	IsGitValidationEnabled bool `toml:"enable_git_validation" koanf:"enable_git_validation"`
}

// GitConfig represents git operations specific configuration
type GitConfig struct {
	// Timeout for CLI git operations in seconds
	CLITimeout int `toml:"cli_timeout" koanf:"cli_timeout"`

	// IsCacheEnabled reports whether caching is enabled for git operations.
	IsCacheEnabled bool `toml:"cache_enabled" koanf:"cache_enabled"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	IsCacheEnabled         bool          `toml:"cache_enabled" koanf:"cache_enabled"`
	CacheTTL               time.Duration `toml:"cache_ttl" koanf:"cache_ttl"`
	IsConcurrentOpsEnabled bool          `toml:"concurrent_operations" koanf:"concurrent_operations"`
	MaxConcurrent          int           `toml:"max_concurrent" koanf:"max_concurrent"`
}

// ValidationConfig holds validation-specific configuration
type ValidationConfig struct {
	IsStrictBranchNames    bool     `toml:"strict_branch_names" koanf:"strict_branch_names"`
	IsRequireCleanWorktree bool     `toml:"require_clean_worktree" koanf:"require_clean_worktree"`
	IsAllowForceDelete     bool     `toml:"allow_force_delete" koanf:"allow_force_delete"`
	ProtectedBranches      []string `toml:"protected_branches" koanf:"protected_branches"`
}

// NavigationConfig holds navigation-specific configuration
type NavigationConfig struct {
	IsSuggestionsEnabled bool `toml:"enable_suggestions" koanf:"enable_suggestions"`
	MaxSuggestions       int  `toml:"max_suggestions" koanf:"max_suggestions"`
	IsFuzzyMatching      bool `toml:"fuzzy_matching" koanf:"fuzzy_matching"`
}

// ShellWrapperConfig represents shell wrapper specific configuration
type ShellWrapperConfig struct {
	// IsEnabled reports whether shell wrapper functionality is on.
	IsEnabled bool `toml:"enabled" koanf:"enabled"`

	// IsAutoDetect reports whether shell type should be auto-detected.
	IsAutoDetect bool `toml:"auto_detect" koanf:"auto_detect"`

	// Default shell type if auto-detection fails
	DefaultShell string `toml:"default_shell" koanf:"default_shell"`

	// IsBackupEnabled reports whether backups of existing configuration files are kept.
	IsBackupEnabled bool `toml:"backup_enabled" koanf:"backup_enabled"`

	// Backup directory for configuration file backups
	BackupDir string `toml:"backup_dir" koanf:"backup_dir"`
}

// RebaseConfig holds knobs that govern the rebase walk. FetchOnAll is
// reserved for a follow-up: a per-project fetch opt-in during --all
// fan-out. ConflictPolicy is reserved for a follow-up: a switch that
// moves past stop-on-first conflict to a more permissive policy.
type RebaseConfig struct {
	FetchOnAll     bool   `toml:"fetch_on_all" koanf:"fetch_on_all"`
	ConflictPolicy string `toml:"conflict_policy" koanf:"conflict_policy"`
}

// SyncConfig holds knobs that govern the sync walk. DefaultRemote
// resolves the remote when --remote is absent. PruneRemoteRefs asks
// git fetch to drop refs that no longer exist upstream. RebaseAfterSync
// triggers the rebase walk after a successful fetch.
type SyncConfig struct {
	DefaultRemote   string `toml:"default_remote" koanf:"default_remote"`
	PruneRemoteRefs bool   `toml:"prune_remote_refs" koanf:"prune_remote_refs"`
	RebaseAfterSync bool   `toml:"rebase_after_sync" koanf:"rebase_after_sync"`
}

// StatusConfig holds knobs that govern the `twiggit status` command
// and the stale heuristic. StaleBehind and StaleDays are independent:
// either trips the IsStale column. A value of 0 disables the
// corresponding half of the heuristic. The TOML key prefix is
// `[status]`.
type StatusConfig struct {
	StaleBehind int `toml:"stale_behind" koanf:"stale_behind"`
	StaleDays   int `toml:"stale_days" koanf:"stale_days"`
}

// CompletionConfig represents shell completion specific configuration
type CompletionConfig struct {
	// Timeout for completion operations
	Timeout string `toml:"timeout" koanf:"timeout"`

	// ExcludeBranches contains glob patterns for branches to exclude from suggestions
	ExcludeBranches []string `toml:"exclude_branches" koanf:"exclude_branches"`

	// ExcludeProjects contains glob patterns for projects to exclude from suggestions
	ExcludeProjects []string `toml:"exclude_projects" koanf:"exclude_projects"`
}

// ShellConfig represents shell integration specific configuration
type ShellConfig struct {
	// Shell wrapper configuration
	Wrapper ShellWrapperConfig `toml:"wrapper" koanf:"wrapper"`

	// IsEnabled reports whether shell integration features are on.
	IsEnabled bool `toml:"enabled" koanf:"enabled"`

	// Timeout for shell operations in seconds
	Timeout int `toml:"timeout" koanf:"timeout"`

	// HookTimeout is the timeout for hook execution in seconds
	HookTimeout int `toml:"hook_timeout" koanf:"hook_timeout"`
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

	// Completion settings
	Completion CompletionConfig `toml:"completion" koanf:"completion"`

	// Rebase settings
	Rebase RebaseConfig `toml:"rebase" koanf:"rebase"`

	// Sync settings
	Sync SyncConfig `toml:"sync" koanf:"sync"`

	// Status settings (stale heuristic thresholds for `twiggit status`).
	Status StatusConfig `toml:"status" koanf:"status"`

	// IsColorEnabled reports whether ANSI color output is enabled.
	// It is set by config.Manager from NO_COLOR at load time (default true).
	IsColorEnabled bool `toml:"-" koanf:"-"`
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
			CacheTTL:               "5m",
			GitOperationTimeout:    "30s",
			IsGitValidationEnabled: true,
		},
		Git: GitConfig{
			CLITimeout:     30,
			IsCacheEnabled: true,
		},
		Services: ServiceConfig{
			IsCacheEnabled:         true,
			CacheTTL:               5 * time.Minute,
			IsConcurrentOpsEnabled: true,
			MaxConcurrent:          4,
		},
		Validation: ValidationConfig{
			IsStrictBranchNames:    true,
			IsRequireCleanWorktree: true,
			IsAllowForceDelete:     false,
			ProtectedBranches:      []string{"main", "master", "develop", "staging", "production"},
		},
		Navigation: NavigationConfig{
			IsSuggestionsEnabled: true,
			MaxSuggestions:       10,
			IsFuzzyMatching:      false,
		},
		Shell: ShellConfig{
			IsEnabled:   true,
			Timeout:     30,
			HookTimeout: 30,
			Wrapper: ShellWrapperConfig{
				IsEnabled:       true,
				IsAutoDetect:    true,
				DefaultShell:    "bash",
				IsBackupEnabled: true,
				BackupDir:       "~/.config/twiggit/backups",
			},
		},
		Completion: CompletionConfig{
			Timeout:         "500ms",
			ExcludeBranches: []string{},
			ExcludeProjects: []string{},
		},
		Rebase: RebaseConfig{
			FetchOnAll:     false,
			ConflictPolicy: "stop",
		},
		Sync: SyncConfig{
			DefaultRemote:   "origin",
			PruneRemoteRefs: true,
			RebaseAfterSync: false,
		},
		Status: StatusConfig{
			StaleBehind: 20,
			StaleDays:   30,
		},
		IsColorEnabled: true,
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

	// Validate status stale thresholds. Zero is a valid value (it
	// disables that half of the stale heuristic); negative values are
	// not — the heuristic is a count / day count, not a sign. No upper
	// bound: a user may set a high threshold to silence the column.
	if c.Status.StaleBehind < 0 {
		validationErrors = append(validationErrors, "status.stale_behind cannot be negative")
	}
	if c.Status.StaleDays < 0 {
		validationErrors = append(validationErrors, "status.stale_days cannot be negative")
	}

	if len(validationErrors) > 0 {
		return &ValidationError{
			Op:          "Config.Validate",
			Field:       "validation",
			Message:     "config validation failed",
			Suggestions: validationErrors,
		}
	}

	return nil
}
