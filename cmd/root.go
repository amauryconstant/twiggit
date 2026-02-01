// Package cmd contains the CLI commands for twiggit.
package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
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

// NewRootCommand creates a new root command with the given configuration
func NewRootCommand(config *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twiggit",
		Short: "A pragmatic tool for managing git worktrees",
		Long: `twiggit is a pragmatic tool for managing git worktrees with a focus on rebase workflows.
It provides context-aware operations for creating, listing, navigating, and deleting worktrees
across multiple projects.`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if config == nil || config.Config == nil {
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
