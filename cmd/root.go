package cmd

import (
	"fmt"
	"os"

	"github.com/amaury/twiggit/internal/di"
	"github.com/amaury/twiggit/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for twiggit
func NewRootCmd(container *di.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "twiggit",
		Short: "Simple Git worktree and project management",
		Long: `twiggit is a fast and intuitive tool for managing Git worktrees and projects.

It provides simple commands to switch between projects and worktrees,
create new worktrees, list existing ones, and clean up when done.

Perfect for developers who work with multiple branches across different projects.`,
		Version: version.Version(),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	// Add subcommands (ordered by usage frequency)
	rootCmd.AddCommand(NewSwitchCmd(container))
	rootCmd.AddCommand(NewListCmd(container))
	rootCmd.AddCommand(NewCreateCmd(container))
	rootCmd.AddCommand(NewDeleteCmd(container))

	// Set up custom error handling for all commands
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		formattedErr := FormatDomainError(err)
		fmt.Printf("%s\n", formattedErr.Error())
		os.Exit(1)
		return nil // This won't be reached due to os.Exit
	})

	// Wrap all subcommands with error formatting
	wrapCommandsWithErrorHandling(rootCmd, container)

	return rootCmd
}

// wrapCommandsWithErrorHandling wraps all subcommands with error formatting
func wrapCommandsWithErrorHandling(rootCmd *cobra.Command, _ *di.Container) {
	for _, cmd := range rootCmd.Commands() {
		// Silence Cobra's default error printing
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		originalRunE := cmd.RunE
		if originalRunE != nil {
			// Store the original RunE and set up PreRunE for validation
			originalPreRunE := cmd.PreRunE
			cmd.PreRunE = func(c *cobra.Command, args []string) error {
				// Validate arguments and handle errors
				if err := validateCommandArguments(c, args); err != nil {
					formattedErr := FormatDomainError(err)
					fmt.Printf("%s\n", formattedErr.Error())
					os.Exit(1)
				}

				// Call original PreRunE if it exists
				if originalPreRunE != nil {
					return originalPreRunE(c, args)
				}
				return nil
			}

			cmd.RunE = func(c *cobra.Command, args []string) error {
				err := originalRunE(c, args)
				if err != nil {
					formattedErr := FormatDomainError(err)
					fmt.Printf("%s\n", formattedErr.Error())
					// Exit with error code but don't return error to avoid Cobra's default error printing
					os.Exit(1)
				}
				return nil
			}
		}
	}
}

// validateCommandArguments validates arguments for specific commands
func validateCommandArguments(cmd *cobra.Command, args []string) error {
	switch cmd.Name() {
	case "create", "switch":
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}
	case "delete", "list":
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
		}
	}
	return nil
}
