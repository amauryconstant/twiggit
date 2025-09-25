//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amaury/twiggit/internal/infrastructure/config"
)

// TestTOMLConfigurationLoading tests TOML configuration file loading and validation
func TestTOMLConfigurationLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("load valid TOML configuration with default_source_branch", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-config-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a valid TOML configuration file
		configContent := `# Twiggit configuration
projects_path = "/home/user/Projects"
workspaces_path = "/home/user/Workspaces"
project = "my-project"
default_source_branch = "develop"
verbose = false
quiet = false
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Load configuration
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify configuration values
		assert.Equal(t, "/home/user/Projects", cfg.ProjectsPath)
		assert.Equal(t, "/home/user/Workspaces", cfg.WorkspacesPath)
		assert.Equal(t, "my-project", cfg.Project)
		assert.Equal(t, "develop", cfg.DefaultSourceBranch)
		assert.False(t, cfg.Verbose)
		assert.False(t, cfg.Quiet)
	})

	t.Run("load TOML configuration with minimal settings", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-minimal-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a minimal TOML configuration file
		configContent := `default_source_branch = "main"
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Load configuration
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify default values are used for unspecified settings
		assert.NotEmpty(t, cfg.ProjectsPath)   // Should have default value
		assert.NotEmpty(t, cfg.WorkspacesPath) // Should have default value
		assert.Equal(t, "", cfg.Project)       // Default is empty
		assert.Equal(t, "main", cfg.DefaultSourceBranch)
		assert.False(t, cfg.Verbose)
		assert.False(t, cfg.Quiet)
	})

	t.Run("load TOML configuration without default_source_branch", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-no-default-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create TOML configuration without default_source_branch
		configContent := `projects_path = "/custom/projects"
workspaces_path = "/custom/workspaces"
verbose = true
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Load configuration
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify configuration values
		assert.Equal(t, "/custom/projects", cfg.ProjectsPath)
		assert.Equal(t, "/custom/workspaces", cfg.WorkspacesPath)
		assert.Equal(t, "", cfg.Project)
		assert.Equal(t, "", cfg.DefaultSourceBranch) // Should be empty when not specified
		assert.True(t, cfg.Verbose)
		assert.False(t, cfg.Quiet)
	})
}

// TestXDGCompliance tests XDG Base Directory Specification compliance for TOML configuration
func TestXDGCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("XDG_CONFIG_HOME takes highest priority", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-xdg-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create XDG_CONFIG_HOME directory structure
		xdgConfigDir := filepath.Join(tempDir, "xdg_config")
		xdgTwiggitDir := filepath.Join(xdgConfigDir, "twiggit")
		err = os.MkdirAll(xdgTwiggitDir, 0755)
		require.NoError(t, err)

		// Create ~/.config directory structure
		homeConfigDir := filepath.Join(tempDir, "home_config")
		homeTwiggitDir := filepath.Join(homeConfigDir, ".config", "twiggit")
		err = os.MkdirAll(homeTwiggitDir, 0755)
		require.NoError(t, err)

		// Create legacy ~/.twiggit.toml directory
		homeDir := filepath.Join(tempDir, "home")
		err = os.MkdirAll(homeDir, 0755)
		require.NoError(t, err)

		// Create different config files in each location
		xdgConfigContent := `default_source_branch = "xdg_priority"
projects_path = "/xdg/projects"
workspaces_path = "/xdg/workspaces"
`
		xdgConfigPath := filepath.Join(xdgTwiggitDir, "config.toml")
		err = os.WriteFile(xdgConfigPath, []byte(xdgConfigContent), 0644)
		require.NoError(t, err)

		homeConfigContent := `default_source_branch = "home_config_priority"
projects_path = "/home/config/projects"
workspaces_path = "/home/config/workspaces"
`
		homeConfigPath := filepath.Join(homeTwiggitDir, "config.toml")
		err = os.WriteFile(homeConfigPath, []byte(homeConfigContent), 0644)
		require.NoError(t, err)

		legacyConfigContent := `default_source_branch = "legacy_priority"
projects_path = "/legacy/projects"
workspaces_path = "/legacy/workspaces"
`
		legacyConfigPath := filepath.Join(homeDir, ".twiggit.toml")
		err = os.WriteFile(legacyConfigPath, []byte(legacyConfigContent), 0644)
		require.NoError(t, err)

		// Set environment variables
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		oldHome := os.Getenv("HOME")

		os.Setenv("XDG_CONFIG_HOME", xdgConfigDir)
		os.Setenv("HOME", homeDir)

		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
			if oldHome != "" {
				os.Setenv("HOME", oldHome)
			} else {
				os.Unsetenv("HOME")
			}
		}()

		// Load configuration - should use XDG_CONFIG_HOME
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify XDG_CONFIG_HOME config is used (highest priority)
		assert.Equal(t, "xdg_priority", cfg.DefaultSourceBranch)
		assert.Equal(t, "/xdg/projects", cfg.ProjectsPath)
		assert.Equal(t, "/xdg/workspaces", cfg.WorkspacesPath)
	})

	t.Run("fallback to ~/.config/twiggit/config.toml when XDG_CONFIG_HOME not set", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-xdg-fallback-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create ~/.config directory structure
		homeDir := filepath.Join(tempDir, "home")
		homeTwiggitDir := filepath.Join(homeDir, ".config", "twiggit")
		err = os.MkdirAll(homeTwiggitDir, 0755)
		require.NoError(t, err)

		// Create legacy ~/.twiggit.toml directory
		legacyConfigContent := `default_source_branch = "legacy_should_not_be_used"
projects_path = "/legacy/projects"
workspaces_path = "/legacy/workspaces"
`
		legacyConfigPath := filepath.Join(homeDir, ".twiggit.toml")
		err = os.WriteFile(legacyConfigPath, []byte(legacyConfigContent), 0644)
		require.NoError(t, err)

		// Create ~/.config/twiggit/config.toml
		homeConfigContent := `default_source_branch = "home_config_fallback"
projects_path = "/home/config/fallback/projects"
workspaces_path = "/home/config/fallback/workspaces"
`
		homeConfigPath := filepath.Join(homeTwiggitDir, "config.toml")
		err = os.WriteFile(homeConfigPath, []byte(homeConfigContent), 0644)
		require.NoError(t, err)

		// Set HOME but not XDG_CONFIG_HOME
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		oldHome := os.Getenv("HOME")

		// Unset XDG_CONFIG_HOME to force fallback
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Setenv("HOME", homeDir)

		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			}
			if oldHome != "" {
				os.Setenv("HOME", oldHome)
			} else {
				os.Unsetenv("HOME")
			}
		}()

		// Load configuration - should use ~/.config/twiggit/config.toml
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify ~/.config/twiggit/config.toml config is used (fallback priority)
		assert.Equal(t, "home_config_fallback", cfg.DefaultSourceBranch)
		assert.Equal(t, "/home/config/fallback/projects", cfg.ProjectsPath)
		assert.Equal(t, "/home/config/fallback/workspaces", cfg.WorkspacesPath)
	})

	t.Run("fallback to ~/.twiggit.toml when no other configs found", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-xdg-legacy-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create home directory
		homeDir := filepath.Join(tempDir, "home")
		err = os.MkdirAll(homeDir, 0755)
		require.NoError(t, err)

		// Create only legacy ~/.twiggit.toml
		legacyConfigContent := `default_source_branch = "legacy_fallback"
projects_path = "/legacy/fallback/projects"
workspaces_path = "/legacy/fallback/workspaces"
verbose = true
`
		legacyConfigPath := filepath.Join(homeDir, ".twiggit.toml")
		err = os.WriteFile(legacyConfigPath, []byte(legacyConfigContent), 0644)
		require.NoError(t, err)

		// Set HOME but not XDG_CONFIG_HOME, and don't create ~/.config/twiggit/
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		oldHome := os.Getenv("HOME")

		os.Unsetenv("XDG_CONFIG_HOME")
		os.Setenv("HOME", homeDir)

		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			}
			if oldHome != "" {
				os.Setenv("HOME", oldHome)
			} else {
				os.Unsetenv("HOME")
			}
		}()

		// Load configuration - should use ~/.twiggit.toml
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify ~/.twiggit.toml config is used (legacy fallback)
		assert.Equal(t, "legacy_fallback", cfg.DefaultSourceBranch)
		assert.Equal(t, "/legacy/fallback/projects", cfg.ProjectsPath)
		assert.Equal(t, "/legacy/fallback/workspaces", cfg.WorkspacesPath)
		assert.True(t, cfg.Verbose)
	})

	t.Run("use defaults when no config files found", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-xdg-defaults-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create empty home directory (no config files)
		homeDir := filepath.Join(tempDir, "home")
		err = os.MkdirAll(homeDir, 0755)
		require.NoError(t, err)

		// Set HOME but not XDG_CONFIG_HOME, and don't create any config files
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		oldHome := os.Getenv("HOME")

		os.Unsetenv("XDG_CONFIG_HOME")
		os.Setenv("HOME", homeDir)

		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			}
			if oldHome != "" {
				os.Setenv("HOME", oldHome)
			} else {
				os.Unsetenv("HOME")
			}
		}()

		// Load configuration - should use defaults
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify default values are used
		assert.Contains(t, cfg.ProjectsPath, "Projects")     // Default path contains "Projects"
		assert.Contains(t, cfg.WorkspacesPath, "Workspaces") // Default path contains "Workspaces"
		assert.Equal(t, "", cfg.Project)                     // Default is empty
		assert.Equal(t, "", cfg.DefaultSourceBranch)         // Default is empty
		assert.False(t, cfg.Verbose)                         // Default is false
		assert.False(t, cfg.Quiet)                           // Default is false
	})
}

// TestTOMLConfigurationValidation tests TOML configuration validation
func TestTOMLConfigurationValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("reject TOML with invalid default_source_branch", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-invalid-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create TOML configuration with invalid branch name
		configContent := `default_source_branch = "invalid@branch#name"
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Load configuration - should fail validation
		_, err = config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid default source branch name")
		assert.Contains(t, err.Error(), "invalid@branch#name")
	})

	t.Run("reject TOML with both verbose and quiet enabled", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-verbose-quiet-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create TOML configuration with both verbose and quiet enabled
		configContent := `verbose = true
quiet = true
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Load configuration - should fail validation
		_, err = config.LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verbose and quiet cannot both be enabled")
	})

	t.Run("accept TOML with valid special branch names", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-valid-branch-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create TOML configuration with valid branch names containing allowed special characters
		validBranchNames := []string{
			"feature/branch-name",
			"hotfix/123-bug-fix",
			"release/v2.1.0",
			"develop",
			"main",
			"master",
			"feature_branch",
			"feature-branch",
		}

		for _, branchName := range validBranchNames {
			configContent := `default_source_branch = "` + branchName + `"`
			configPath := filepath.Join(tempDir, "config.toml")
			err = os.WriteFile(configPath, []byte(configContent), 0644)
			require.NoError(t, err)

			// Set XDG_CONFIG_HOME to point to our temp directory
			oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)
			defer func() {
				if oldXdgConfigHome != "" {
					os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
				} else {
					os.Unsetenv("XDG_CONFIG_HOME")
				}
			}()

			// Create the twiggit subdirectory
			configDir := filepath.Join(tempDir, "twiggit")
			err = os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			// Move the config file to the correct location
			correctConfigPath := filepath.Join(configDir, "config.toml")
			err = os.Rename(configPath, correctConfigPath)
			require.NoError(t, err)

			// Load configuration - should succeed
			cfg, err := config.LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, branchName, cfg.DefaultSourceBranch)

			// Clean up for next iteration
			os.RemoveAll(configDir)
		}
	})
}

// TestTOMLConfigurationPriority tests configuration loading priority
func TestTOMLConfigurationPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("environment variables override TOML configuration", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "twiggit-toml-priority-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create TOML configuration
		configContent := `default_source_branch = "develop"
projects_path = "/from/toml/projects"
workspaces_path = "/from/toml/workspaces"
verbose = false
quiet = false
`
		configPath := filepath.Join(tempDir, "config.toml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set XDG_CONFIG_HOME to point to our temp directory
		oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer func() {
			if oldXdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Create the twiggit subdirectory
		configDir := filepath.Join(tempDir, "twiggit")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Move the config file to the correct location
		correctConfigPath := filepath.Join(configDir, "config.toml")
		err = os.Rename(configPath, correctConfigPath)
		require.NoError(t, err)

		// Set environment variables that should override TOML
		oldTwiggitProjectsPath := os.Getenv("TWIGGIT_PROJECTS_PATH")
		oldTwiggitWorkspacesPath := os.Getenv("TWIGGIT_WORKSPACES_PATH")
		oldTwiggitDefaultSourceBranch := os.Getenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
		oldTwiggitVerbose := os.Getenv("TWIGGIT_VERBOSE")

		os.Setenv("TWIGGIT_PROJECTS_PATH", "/from/env/projects")
		os.Setenv("TWIGGIT_WORKSPACES_PATH", "/from/env/workspaces")
		os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", "main")
		os.Setenv("TWIGGIT_VERBOSE", "true")

		defer func() {
			if oldTwiggitProjectsPath != "" {
				os.Setenv("TWIGGIT_PROJECTS_PATH", oldTwiggitProjectsPath)
			} else {
				os.Unsetenv("TWIGGIT_PROJECTS_PATH")
			}
			if oldTwiggitWorkspacesPath != "" {
				os.Setenv("TWIGGIT_WORKSPACES_PATH", oldTwiggitWorkspacesPath)
			} else {
				os.Unsetenv("TWIGGIT_WORKSPACES_PATH")
			}
			if oldTwiggitDefaultSourceBranch != "" {
				os.Setenv("TWIGGIT_DEFAULT_SOURCE_BRANCH", oldTwiggitDefaultSourceBranch)
			} else {
				os.Unsetenv("TWIGGIT_DEFAULT_SOURCE_BRANCH")
			}
			if oldTwiggitVerbose != "" {
				os.Setenv("TWIGGIT_VERBOSE", oldTwiggitVerbose)
			} else {
				os.Unsetenv("TWIGGIT_VERBOSE")
			}
		}()

		// Load configuration
		cfg, err := config.LoadConfig()
		require.NoError(t, err)

		// Verify environment variables override TOML values
		assert.Equal(t, "/from/env/projects", cfg.ProjectsPath)
		assert.Equal(t, "/from/env/workspaces", cfg.WorkspacesPath)
		assert.Equal(t, "main", cfg.DefaultSourceBranch)
		assert.True(t, cfg.Verbose)
		assert.False(t, cfg.Quiet) // Not set in env, should use TOML value
	})
}
