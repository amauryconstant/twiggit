package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"twiggit/internal/domain"
)

// NewInitCmd creates a new init command
func NewInitCmd(config *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [config-file]",
		Short: "Install shell wrapper",
		Long: `Install shell wrapper functions that intercept 'twiggit cd' calls
and enable seamless directory navigation between worktrees and projects.

The wrapper provides:
- Automatic directory change on 'twiggit cd'
- Escape hatch with 'builtin cd' for shell built-in
- Pass-through for all other commands

Supported shells: bash, zsh, fish

Usage:
  twiggit init                    # Auto-detect shell and config file
  twiggit init ~/.bashrc          # Install to specific config file
  twiggit init --shell=zsh        # Explicit shell, auto-detect config file
  twiggit init ~/.config/my-zsh --shell=zsh  # Explicit config and shell`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := ""
			if len(args) > 0 {
				configFile = args[0]
			}
			return runInit(cmd, config, configFile)
		},
	}

	cmd.Flags().String("shell", "", "shell type (bash|zsh|fish) [optional, inferred from config file]")
	cmd.Flags().Bool("force", false, "force reinstall even if already installed")
	cmd.Flags().Bool("dry-run", false, "show what would be done without making changes")
	cmd.Flags().Bool("check", false, "check if wrapper is installed")

	return cmd
}

func runInit(cmd *cobra.Command, config *CommandConfig, configFile string) error {
	check, _ := cmd.Flags().GetBool("check")

	if check {
		return runInitCheck(cmd, config, configFile)
	}

	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	shellTypeStr, _ := cmd.Flags().GetString("shell")

	var shellType domain.ShellType
	if shellTypeStr != "" {
		shellType = domain.ShellType(shellTypeStr)
	}

	request := &domain.SetupShellRequest{
		ShellType:  shellType,
		Force:      force,
		DryRun:     dryRun,
		ConfigFile: configFile,
	}

	result, err := config.Services.ShellService.SetupShell(context.Background(), request)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	return displayInitResults(cmd.OutOrStdout(), result, dryRun)
}

func runInitCheck(cmd *cobra.Command, config *CommandConfig, configFile string) error {
	shellTypeStr, _ := cmd.Flags().GetString("shell")

	var shellType domain.ShellType
	if shellTypeStr != "" {
		shellType = domain.ShellType(shellTypeStr)
	}

	request := &domain.ValidateInstallationRequest{
		ShellType:  shellType,
		ConfigFile: configFile,
	}

	result, err := config.Services.ShellService.ValidateInstallation(context.Background(), request)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	out := cmd.OutOrStdout()
	if result.Installed {
		_, _ = fmt.Fprintf(out, "Shell wrapper is installed\n")
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
	} else {
		_, _ = fmt.Fprintf(out, "Shell wrapper not installed\n")
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
	}

	return nil
}

func displayInitResults(out io.Writer, result *domain.SetupShellResult, dryRun bool) error {
	if result.Skipped {
		_, _ = fmt.Fprintf(out, "Shell wrapper already installed for %s\n", result.ShellType)
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
		_, _ = fmt.Fprintf(out, "Use --force to reinstall\n")
		return nil
	}

	if dryRun {
		_, _ = fmt.Fprintf(out, "Would install wrapper for %s:\n", result.ShellType)
		_, _ = fmt.Fprintf(out, "Config file: %s\n", result.ConfigFile)
		_, _ = fmt.Fprintf(out, "Wrapper function:\n%s\n", result.WrapperContent)
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
