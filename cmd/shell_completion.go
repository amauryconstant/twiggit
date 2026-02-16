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
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(c.OutOrStdout())
			case "zsh":
				rootCmd.GenZshCompletion(c.OutOrStdout())
			case "fish":
				rootCmd.GenFishCompletion(c.OutOrStdout(), true)
			case "powershell":
				rootCmd.GenPowerShellCompletionWithDesc(c.OutOrStdout())
			}
		},
	}
}
