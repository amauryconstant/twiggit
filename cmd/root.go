package cmd

import (
	"errors"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/application"
	"twiggit/internal/domain"
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
	ContextService    application.ContextService
	ShellService      application.ShellService
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

	// Add persistent verbose flag
	cmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (can be used multiple times: -v, -vv)")

	// Add subcommands
	cmd.AddCommand(NewListCommand(config))
	cmd.AddCommand(NewCreateCommand(config))
	cmd.AddCommand(NewDeleteCommand(config))
	cmd.AddCommand(NewPruneCommand(config))
	cmd.AddCommand(NewCDCommand(config))
	cmd.AddCommand(NewInitCmd(config))
	cmd.AddCommand(NewVersionCommand(config))

	carapace.Gen(cmd)

	// Replace Cobra's default completion command with Carapace's version
	if completionCmd, _, err := cmd.Find([]string{"completion"}); err == nil {
		cmd.RemoveCommand(completionCmd)
	}
	cmd.AddCommand(newCompletionCommand(cmd))

	return cmd
}
