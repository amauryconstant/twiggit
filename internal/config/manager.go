package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"twiggit/internal/core"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// Manager is the canonical configuration manager. cmdutil.Factory.Config
// constructs it via NewManager and calls Load to retrieve *core.Config.
type Manager = koanfConfigManager

// Pure functions extracted from ConfigManager

// expandConfigPath expands environment variables and tilde in a path string.
// It handles $VAR, ${VAR}, and ~ syntax.
func expandConfigPath(path string) string {
	if path == "" {
		return path
	}

	// Handle tilde expansion first
	if rest, ok := strings.CutPrefix(path, "~"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to $HOME env var
			home = os.Getenv("HOME")
			if home == "" {
				// Last resort fallback
				home = "/tmp"
			}
		}
		return filepath.Join(home, rest)
	}

	// Handle $VAR and ${VAR} expansion
	return os.ExpandEnv(path)
}

// normalizeConfigPaths expands environment variables and tilde in all path fields of the config.
func normalizeConfigPaths(config *core.Config) {
	config.ProjectsDirectory = expandConfigPath(config.ProjectsDirectory)
	config.WorktreesDirectory = expandConfigPath(config.WorktreesDirectory)
	config.Shell.Wrapper.BackupDir = expandConfigPath(config.Shell.Wrapper.BackupDir)
}

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
func buildDefaultConfig() *core.Config {
	return core.DefaultConfig()
}

// configFileExists checks if a file exists at the given path
func configFileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// validateConfig validates a configuration object
func validateConfig(config *core.Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

// copyConfig creates a deep copy of a configuration object
func copyConfig(config *core.Config) *core.Config {
	validation := config.Validation
	validation.ProtectedBranches = slices.Clone(config.Validation.ProtectedBranches)
	return &core.Config{
		ProjectsDirectory:   config.ProjectsDirectory,
		WorktreesDirectory:  config.WorktreesDirectory,
		DefaultSourceBranch: config.DefaultSourceBranch,
		ContextDetection:    config.ContextDetection,
		Git:                 config.Git,
		Services:            config.Services,
		Validation:          validation,
		Navigation:          config.Navigation,
		Shell:               config.Shell,
		Completion:          config.Completion,
		Rebase:              config.Rebase,
		Sync:                config.Sync,
		Status:              config.Status,
		IsColorEnabled:      config.IsColorEnabled,
	}
}

// osRootProvider reads config bytes via os.Root so the load path is
// constrained to the resolved config directory tree. Falls back to the
// standard file provider when the root cannot be opened (e.g. relative
// paths whose parent does not exist as a directory).
type osRootProvider struct {
	path string
	root *os.Root
	base string
}

// newOsRootProvider wraps the given file path in an os.Root bound to its
// parent directory. Subsequent ReadBytes calls go through the bounded
// root so a tampered XDG_CONFIG_HOME cannot redirect the read outside
// the resolved subtree.
func newOsRootProvider(path string) (*osRootProvider, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path: %w", err)
	}
	dir := filepath.Dir(absPath)
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open config root: %w", err)
	}
	return &osRootProvider{
		path: absPath,
		root: root,
		base: filepath.Base(absPath),
	}, nil
}

// ReadBytes satisfies koanf.Provider.
func (p *osRootProvider) ReadBytes() ([]byte, error) {
	defer func() { _ = p.root.Close() }()
	data, err := p.root.ReadFile(p.base)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return data, nil
}

// Read returns an explicit error; koanf uses ReadBytes when a Parser is supplied.
func (p *osRootProvider) Read() (map[string]any, error) {
	return nil, errors.New("osRootProvider: use ReadBytes")
}

type koanfConfigManager struct {
	ko     *koanf.Koanf
	config *core.Config
}

// NewManager creates a new configuration manager backed by koanf.
func NewManager() *Manager {
	return &koanfConfigManager{
		ko: koanf.New("."),
	}
}

// Load loads configuration from defaults and config file.
func (m *koanfConfigManager) Load() (*core.Config, error) {
	// 1. Load defaults
	if err := m.loadDefaults(); err != nil {
		return nil, &core.OperationError{
			Op:      "config.load",
			Message: "failed to load default configuration",
			Cause:   err,
		}
	}

	// 2. Load config file using os.Root-bounded reader
	configPath := m.getConfigFilePath()
	if configPath != "" && configFileExists(configPath) {
		provider, err := newOsRootProvider(configPath)
		if err != nil {
			return nil, &core.OperationError{
				Op:      "config.load",
				Entity:  configPath,
				Message: "failed to open config directory",
				Cause:   err,
			}
		}
		if err := m.ko.Load(provider, toml.Parser()); err != nil {
			return nil, &core.OperationError{
				Op:      "config.load",
				Entity:  configPath,
				Message: "failed to parse config file",
				Cause:   err,
			}
		}
	}

	// 2b. Load env overrides (TWIGGIT_<SECTION>__<KEY>) so test/CI
	// harnesses can flip a knob without touching the config file.
	// Koanf's env provider is configured with the TWIGGIT_ prefix; the
	// transform converts REBASE__FETCH_ON_ALL -> rebase.fetch_on_all so
	// the override lands on the same key the TOML file uses.
	if err := m.ko.Load(env.Provider(".", env.Opt{
		Prefix: "TWIGGIT_",
		TransformFunc: func(k, v string) (string, any) {
			key := strings.ToLower(strings.TrimPrefix(k, "TWIGGIT_"))
			key = strings.ReplaceAll(key, "__", ".")
			return key, v
		},
	}), nil); err != nil {
		return nil, &core.OperationError{
			Op:      "config.load",
			Message: "failed to load env overrides",
			Cause:   err,
		}
	}

	// 3. Unmarshal to config object
	config := &core.Config{}
	if err := m.ko.Unmarshal("", config); err != nil {
		return nil, &core.OperationError{
			Op:      "config.load",
			Entity:  configPath,
			Message: "failed to unmarshal configuration",
			Cause:   err,
		}
	}

	// 4. Normalize paths (expand environment variables and tilde)
	normalizeConfigPaths(config)

	// 5. NO_COLOR handling: presence of NO_COLOR in env disables color output.
	// Default is true (color on); set false when NO_COLOR is set in env (any value).
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		config.IsColorEnabled = false
	} else {
		config.IsColorEnabled = true
	}

	// 6. Validate configuration using pure function
	if err := validateConfig(config); err != nil {
		return nil, &core.OperationError{
			Op:      "config.load",
			Entity:  configPath,
			Message: "validation failed",
			Cause:   err,
		}
	}

	// 7. Store immutable config
	m.config = config

	// 8. Return a copy using pure function to maintain immutability
	return copyConfig(config), nil
}

// Config returns the loaded configuration (immutable copy)
func (m *koanfConfigManager) Config() *core.Config {
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
	if err := m.ko.Set("context_detection.enable_git_validation", defaults.ContextDetection.IsGitValidationEnabled); err != nil {
		return fmt.Errorf("failed to set context_detection.enable_git_validation default: %w", err)
	}

	if err := m.ko.Set("git.cli_timeout", defaults.Git.CLITimeout); err != nil {
		return fmt.Errorf("failed to set git.cli_timeout default: %w", err)
	}
	if err := m.ko.Set("git.cache_enabled", defaults.Git.IsCacheEnabled); err != nil {
		return fmt.Errorf("failed to set git.cache_enabled default: %w", err)
	}
	if err := m.ko.Set("completion.timeout", defaults.Completion.Timeout); err != nil {
		return fmt.Errorf("failed to set completion.timeout default: %w", err)
	}
	if err := m.ko.Set("rebase.fetch_on_all", defaults.Rebase.FetchOnAll); err != nil {
		return fmt.Errorf("failed to set rebase.fetch_on_all default: %w", err)
	}
	if err := m.ko.Set("rebase.conflict_policy", defaults.Rebase.ConflictPolicy); err != nil {
		return fmt.Errorf("failed to set rebase.conflict_policy default: %w", err)
	}
	if err := m.ko.Set("sync.default_remote", defaults.Sync.DefaultRemote); err != nil {
		return fmt.Errorf("failed to set sync.default_remote default: %w", err)
	}
	if err := m.ko.Set("sync.prune_remote_refs", defaults.Sync.PruneRemoteRefs); err != nil {
		return fmt.Errorf("failed to set sync.prune_remote_refs default: %w", err)
	}
	if err := m.ko.Set("sync.rebase_after_sync", defaults.Sync.RebaseAfterSync); err != nil {
		return fmt.Errorf("failed to set sync.rebase_after_sync default: %w", err)
	}
	if err := m.ko.Set("status.stale_behind", defaults.Status.StaleBehind); err != nil {
		return fmt.Errorf("failed to set status.stale_behind default: %w", err)
	}
	if err := m.ko.Set("status.stale_days", defaults.Status.StaleDays); err != nil {
		return fmt.Errorf("failed to set status.stale_days default: %w", err)
	}
	return nil
}

// getConfigFilePath returns the path to the configuration file following XDG Base Directory specification
func (m *koanfConfigManager) getConfigFilePath() string {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	home, _ := os.UserHomeDir()
	return resolveConfigPath(xdgHome, home)
}
