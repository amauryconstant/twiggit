package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/output"
)

// NewInitCmd creates a new init command
func NewInitCmd(f *CommandConfig) *cobra.Command {
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

Examples:
  eval "$(twiggit init)"                  # Add to your shell config for instant activation
  twiggit init bash                       # Print bash wrapper to stdout
  twiggit init --install                  # Install to auto-detected config file
  twiggit init bash --install -c ~/.bashrc  # Install to specific config file`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          wrapArgsValidator(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFile != "" && !install {
				return core.NewUsageError("--config requires --install", nil)
			}
			if force && !install {
				return core.NewUsageError("--force requires --install", nil)
			}

			var shellType core.ShellType
			if len(args) > 0 {
				shellType = core.ShellType(args[0])
			}

			if install {
				return runInitInstall(cmd, f, shellType, configFile, force)
			}
			return runInitStdout(cmd, f, shellType)
		},
	}

	cmd.Flags().BoolVarP(&install, "install", "i", false, "install wrapper to shell config file")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "custom config file path (requires --install)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force reinstall even if already installed (requires --install)")

	carapace.Gen(cmd).PositionalCompletion(
		carapace.ActionValues("bash", "zsh", "fish"),
	)

	return cmd
}

// runInitStdout writes the shell wrapper to stdout (default behavior).
// After slice 9 the wrapper comes from core.ShellWrapper rather than
// shellService.GenerateWrapper.
func runInitStdout(cmd *cobra.Command, f *CommandConfig, shellType core.ShellType) error {
	if shellType == "" {
		detected, err := git.DetectShellFromEnv()
		if err != nil {
			return fmt.Errorf("shell auto-detection failed: %w", err)
		}
		shellType = detected
	}

	if !core.IsValidShellType(shellType) {
		err := core.NewOpValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type")
		err.Suggestions = []string{"Supported shells: bash, zsh, fish"}
		return err
	}

	wrapper, err := core.ShellWrapper(shellType)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper: %w", err)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), wrapper)
	return nil
}

// runInitInstall installs the wrapper to a shell config file. After
// slice 9 the orchestration lives in cmd/ with output.InstallWrapper
// providing the file-write side.
func runInitInstall(cmd *cobra.Command, f *CommandConfig, shellType core.ShellType, configFile string, force bool) error {
	if shellType == "" {
		detected, err := git.DetectShellFromEnv()
		if err != nil {
			return fmt.Errorf("shell auto-detection failed: %w", err)
		}
		shellType = detected
	}

	if !core.IsValidShellType(shellType) {
		err := core.NewOpValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type")
		err.Suggestions = []string{"Supported shells: bash, zsh, fish"}
		return err
	}

	if configFile == "" {
		detected, err := output.DetectConfigFile(shellType)
		if err != nil {
			return fmt.Errorf("config file detection failed: %w", err)
		}
		configFile = detected
	}

	// Skip when already installed and not forcing reinstall.
	if !force {
		if err := output.ValidateInstallation(shellType, configFile); err == nil {
			result := &core.SetupShellResult{
				ShellType:   shellType,
				IsInstalled: true,
				IsSkipped:   true,
				ConfigFile:  configFile,
				Message:     "Shell wrapper already installed",
			}
			logv(cmd, 1, "Setting up shell wrapper")
			logv(cmd, 2, "  shell type: %s", result.ShellType)
			logv(cmd, 2, "  config file: %s", result.ConfigFile)
			return displayInitResults(cmd.OutOrStdout(), result)
		}
	}

	wrapper, err := core.ShellWrapper(shellType)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper: %w", err)
	}

	if err := output.InstallWrapper(shellType, wrapper, configFile, force); err != nil {
		if isAlreadyInstalled(err) {
			result := &core.SetupShellResult{
				ShellType:   shellType,
				IsInstalled: true,
				IsSkipped:   true,
				ConfigFile:  configFile,
				Message:     "Shell wrapper already installed",
			}
			return displayInitResults(cmd.OutOrStdout(), result)
		}
		return fmt.Errorf("failed to install wrapper: %w", err)
	}

	result := &core.SetupShellResult{
		ShellType:   shellType,
		IsInstalled: true,
		ConfigFile:  configFile,
		Message:     "Shell wrapper installed successfully",
	}

	logv(cmd, 1, "Setting up shell wrapper")
	logv(cmd, 2, "  shell type: %s", result.ShellType)
	logv(cmd, 2, "  config file: %s", result.ConfigFile)

	return displayInitResults(cmd.OutOrStdout(), result)
}

// isAlreadyInstalled reports whether err is the typed shell-already-installed
// sentinel from output.InstallWrapper.
func isAlreadyInstalled(err error) bool {
	var oe *core.OperationError
	return errors.As(err, &oe) && oe.Op == "shell.already_installed"
}

// displayInitResults outputs installation results (for install mode only)
func displayInitResults(out io.Writer, result *core.SetupShellResult) error {
	if result.IsSkipped {
		_, _ = fmt.Fprintf(out, "Shell wrapper already installed for %s\n", result.ShellType)
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
		_, _ = fmt.Fprintf(out, "Use --force to reinstall\n")
		return nil
	}

	if result.IsInstalled {
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
