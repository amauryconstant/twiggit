package cmd

import (
	"github.com/spf13/cobra"
)

// NewCompletionCommand creates and returns the completion command
func NewCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate the autocompletion script for the specified shell",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Run: func(c *cobra.Command, args []string) {
			var _ error // Ignore return values from completion generation
			switch args[0] {
			case "bash":
				_ = rootCmd.GenBashCompletion(c.OutOrStdout())
			case "zsh":
				_ = rootCmd.GenZshCompletion(c.OutOrStdout())
			case "fish":
				_ = rootCmd.GenFishCompletion(c.OutOrStdout(), true)
			case "powershell":
				_ = rootCmd.GenPowerShellCompletionWithDesc(c.OutOrStdout())
			}
		},
	}
}
