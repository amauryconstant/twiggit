// Package cmd contains the CLI commands for twiggit.
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

var (
	appConfig *domain.Config
)

// CommandConfig holds the configuration for CLI commands
type CommandConfig struct {
	Services *ServiceContainer
	Config   *domain.Config
}

// ServiceContainer holds all service dependencies for commands
type ServiceContainer struct {
	WorktreeService   application.WorktreeService
	ProjectService    application.ProjectService
	NavigationService application.NavigationService
	ContextService    domain.ContextService
	ShellService      application.ShellService
	GitClient         infrastructure.GitClient
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "twiggit",
	Short: "A pragmatic tool for managing git worktrees",
	Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees
across multiple projects.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		// Get global configuration (loaded in main)
		appConfig = domain.GetGlobalConfig()
		if appConfig == nil {
			return errors.New("cmd: configuration not loaded")
		}
		return nil
	},
}

// Note: Commands are now initialized in main.go via NewRootCommand to ensure proper service injection

// Execute runs the root command.
// Note: This function is deprecated. Use NewRootCommand() instead for proper service initialization.
func Execute() error {
	// Try to get global config, but don't fail if it's not set
	config := domain.GetGlobalConfig()
	if config == nil {
		config = domain.DefaultConfig()
	}

	// Create a basic command config for backward compatibility
	commandConfig := &CommandConfig{
		Config:   config,
		Services: &ServiceContainer{}, // Empty services - will cause errors but won't crash
	}

	// Use NewRootCommand for proper initialization
	cmd := NewRootCommand(commandConfig)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}

// GetConfig returns the global application configuration
func GetConfig() *domain.Config {
	return appConfig
}

// NewRootCommand creates a new root command with the given configuration
func NewRootCommand(config *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twiggit",
		Short: "A pragmatic tool for managing git worktrees",
		Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees
across multiple projects.`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if config != nil && config.Config != nil {
				appConfig = config.Config
			} else {
				appConfig = domain.GetGlobalConfig()
			}
			if appConfig == nil {
				return errors.New("cmd: configuration not loaded")
			}
			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(config))
	cmd.AddCommand(NewCreateCommand(config))
	cmd.AddCommand(NewDeleteCommand(config))
	cmd.AddCommand(NewCDCommand(config))
	cmd.AddCommand(NewSetupShellCmd(config))
	cmd.AddCommand(NewVersionCommand(config))

	return cmd
}
