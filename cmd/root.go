// Package cmd contains the CLI commands for twiggit.
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"twiggit/internal/domain"
)

var (
	appConfig *domain.Config
)

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

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}

// GetConfig returns the global application configuration
func GetConfig() *domain.Config {
	return appConfig
}
