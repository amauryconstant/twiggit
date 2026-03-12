package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/domain"
)

// NewInitCmd creates a new init command
func NewInitCmd(config *CommandConfig) *cobra.Command {
	var install, force bool
	var configFile string

	cmd := &cobra.Command{
		Use:   "init [shell]",
		Short: "Generate or install shell wrapper",
		Long: `Generate shell wrapper functions that intercept 'twiggit cd' calls
and enable seamless directory navigation between worktrees and projects.

The wrapper provides:
- Automatic directory change on 'twiggit cd'
- Escape hatch with 'builtin cd' for shell built-in
- Pass-through for all other commands

Supported shells: bash, zsh, fish

Usage:
  twiggit init                    # Print wrapper to stdout (eval-safe)
  twiggit init bash               # Print bash wrapper to stdout
  twiggit init --install          # Install to auto-detected config file
  twiggit init bash --install     # Install bash wrapper to auto-detected config
  twiggit init bash --install -c ~/.bashrc  # Install to specific config file

Examples:
  # Add to your shell config for instant activation:
  eval "$(twiggit init)"

  # Or install permanently to your shell config:
  twiggit init --install

Flags:
  -i, --install    Install wrapper to shell config file
  -c, --config     Custom config file path (requires --install)
  -f, --force      Force reinstall even if already installed (requires --install)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate flag combinations
			if configFile != "" && !install {
				return errors.New("--config requires --install")
			}
			if force && !install {
				return errors.New("--force requires --install")
			}

			// Parse shell type from positional argument
			var shellType domain.ShellType
			if len(args) > 0 {
				shellType = domain.ShellType(args[0])
			}

			if install {
				return runInitInstall(cmd, config, shellType, configFile, force)
			}
			return runInitStdout(cmd, config, shellType)
		},
	}

	cmd.Flags().BoolVarP(&install, "install", "i", false, "install wrapper to shell config file")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "custom config file path (requires --install)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force reinstall even if already installed (requires --install)")

	// Shell completion for positional [shell] argument
	carapace.Gen(cmd).PositionalCompletion(
		carapace.ActionValues("bash", "zsh", "fish"),
	)

	return cmd
}

// runInitStdout outputs the shell wrapper to stdout (default behavior)
func runInitStdout(cmd *cobra.Command, config *CommandConfig, shellType domain.ShellType) error {
	// Auto-detect shell if not specified
	if shellType == "" {
		var err error
		shellType, err = domain.DetectShellFromEnv()
		if err != nil {
			return fmt.Errorf("shell auto-detection failed: %w", err)
		}
	}

	// Validate shell type
	if !domain.IsValidShellType(shellType) {
		return domain.NewValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type").
			WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
	}

	// Generate wrapper
	request := &domain.GenerateWrapperRequest{
		ShellType: shellType,
	}

	result, err := config.Services.ShellService.GenerateWrapper(context.Background(), request)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper: %w", err)
	}

	// Output wrapper to stdout (no metadata, eval-safe)
	_, _ = fmt.Fprint(cmd.OutOrStdout(), result.WrapperContent)

	return nil
}

// runInitInstall installs the wrapper to a shell config file
func runInitInstall(cmd *cobra.Command, config *CommandConfig, shellType domain.ShellType, configFile string, force bool) error {
	// Auto-detect shell if not specified
	if shellType == "" {
		var err error
		shellType, err = domain.DetectShellFromEnv()
		if err != nil {
			return fmt.Errorf("shell auto-detection failed: %w", err)
		}
	}

	// Validate shell type
	if !domain.IsValidShellType(shellType) {
		return domain.NewValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type").
			WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
	}

	request := &domain.SetupShellRequest{
		ShellType:      shellType,
		ForceOverwrite: force,
		ConfigFile:     configFile,
	}

	result, err := config.Services.ShellService.SetupShell(context.Background(), request)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	logv(cmd, 1, "Setting up shell wrapper")
	logv(cmd, 2, "  shell type: %s", result.ShellType)
	logv(cmd, 2, "  config file: %s", result.ConfigFile)

	return displayInitResults(cmd.OutOrStdout(), result)
}

// displayInitResults outputs installation results (for install mode only)
func displayInitResults(out io.Writer, result *domain.SetupShellResult) error {
	if result.Skipped {
		_, _ = fmt.Fprintf(out, "Shell wrapper already installed for %s\n", result.ShellType)
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
		_, _ = fmt.Fprintf(out, "Use --force to reinstall\n")
		return nil
	}

	if result.Installed {
		_, _ = fmt.Fprintf(out, "Shell wrapper installed for %s\n", result.ShellType)
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
		if _, err := os.Stat(result.ConfigFile); err == nil {
			_, _ = fmt.Fprintf(out, "\nTo activate the wrapper:\n")
			_, _ = fmt.Fprintf(out, "  1. Restart your shell, or\n")
			_, _ = fmt.Fprintf(out, "  2. Run: source %s\n", result.ConfigFile)
		}
		_, _ = fmt.Fprintf(out, "\nUsage:\n")
		_, _ = fmt.Fprintf(out, "  twiggit cd <branch>     # Change to worktree\n")
		_, _ = fmt.Fprintf(out, "  builtin cd <path>       # Use shell built-in cd\n")
	}

	return nil
}
