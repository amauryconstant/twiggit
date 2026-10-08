package cmd

import (
	"fmt"
	"slices"
	"strings"
	"twiggit/internal/core"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var supportedCompletionShells = []string{
	"bash", "zsh", "fish", "powershell",
	"elvish", "nushell", "oil", "tcsh", "xonsh", "cmd-clink",
}

func validateSupportedShell(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if slices.Contains(supportedCompletionShells, args[0]) {
		return nil
	}
	return core.NewValidationError(
		"shell", args[0],
		fmt.Sprintf("unsupported shell %q (supported: %s)",
			args[0], strings.Join(supportedCompletionShells, ", ")),
	)
}

func newCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "completion",
		Short:         "Generate the autocompletion script for the specified shell",
		Long:          `Generate the autocompletion script for twiggit for the specified shell. See each sub-command's help for details on how to use the generated script.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          validateSupportedShell,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	for _, shell := range supportedCompletionShells {
		cmd.AddCommand(newCompletionShellCommand(rootCmd, shell))
	}

	return cmd
}

func newCompletionShellCommand(rootCmd *cobra.Command, shell string) *cobra.Command {
	return &cobra.Command{
		Use:   shell,
		Short: "Generate the autocompletion script for " + shell,
		Long: `Generate the autocompletion script for ` + shell + `.

To load completions:

` + getShellInstructions(shell),
		Example:               "  source <(twiggit completion " + shell + ")",
		SilenceUsage:          true,
		SilenceErrors:         true,
		Args:                  wrapArgsValidator(cobra.NoArgs),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			snippet, err := carapace.Gen(rootCmd).Snippet(shell)
			if err != nil {
				return fmt.Errorf("completion: %w", err)
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), snippet); err != nil {
				return fmt.Errorf("completion: %w", err)
			}
			return nil
		},
	}
}

func getShellInstructions(shell string) string {
	switch shell {
	case "bash":
		return `Bash:
  # Load for current session
  source <(twiggit completion bash)

  # Or add to ~/.bashrc for persistence
  echo 'source <(twiggit completion bash)' >> ~/.bashrc

  # Linux may require installation of bash-completion package`
	case "zsh":
		return `Zsh:
  # Load for current session
  source <(twiggit completion zsh)

  # Or add to ~/.zshrc for persistence
  echo 'source <(twiggit completion zsh)' >> ~/.zshrc`
	case "fish":
		return `Fish:
  # Load for current session
  twiggit completion fish | source

  # Or save to completions directory for persistence
  twiggit completion fish > ~/.config/fish/completions/twiggit.fish`
	case "powershell":
		return `PowerShell:
  # Load for current session
  twiggit completion powershell | Out-String | Invoke-Expression

  # Or add to PowerShell profile for persistence
  twiggit completion powershell >> $PROFILE`
	default:
		return `  source <(twiggit completion ` + shell + `)`
	}
}
