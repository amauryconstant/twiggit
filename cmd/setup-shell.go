package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"twiggit/internal/domain"
)

// NewSetupShellCmd creates a new setup-shell command
func NewSetupShellCmd(config *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-shell",
		Short: "Install shell wrapper for directory navigation",
		Long: `Install shell wrapper functions that intercept 'twiggit cd' calls
and enable seamless directory navigation between worktrees and projects.

The wrapper provides:
- Automatic directory change on 'twiggit cd'
- Escape hatch with 'builtin cd' for shell built-in
- Pass-through for all other commands

Supported shells: bash, zsh, fish

Usage: twiggit setup-shell --shell=bash`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSetupShell(cmd, config)
		},
	}

	cmd.Flags().String("shell", "", "shell type (bash|zsh|fish) [required]")
	cmd.Flags().Bool("force", false, "force reinstall even if already installed")
	cmd.Flags().Bool("dry-run", false, "show what would be done without making changes")

	_ = cmd.MarkFlagRequired("shell")

	return cmd
}

func runSetupShell(cmd *cobra.Command, config *CommandConfig) error {
	shellTypeStr, _ := cmd.Flags().GetString("shell")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Validate shell type
	shellType := domain.ShellType(shellTypeStr)
	if !isValidShellType(shellType) {
		return fmt.Errorf("unsupported shell type: %s (supported: bash, zsh, fish)", shellTypeStr)
	}

	// Create request
	request := &domain.SetupShellRequest{
		ShellType: shellType,
		Force:     force,
		DryRun:    dryRun,
	}

	// Execute service
	result, err := config.Services.ShellService.SetupShell(context.Background(), request)
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Display results
	return displaySetupResults(cmd.OutOrStdout(), result, dryRun)
}

func displaySetupResults(out io.Writer, result *domain.SetupShellResult, dryRun bool) error {
	if result.Skipped {
		fmt.Fprintf(out, "✓ Shell wrapper already installed for %s\n", result.ShellType)
		fmt.Fprintf(out, "Use --force to reinstall\n")
		return nil
	}

	if dryRun {
		fmt.Fprintf(out, "Would install wrapper for %s:\n", result.ShellType)
		fmt.Fprintf(out, "Wrapper function:\n%s\n", result.WrapperContent)
		return nil
	}

	if result.Installed {
		fmt.Fprintf(out, "✓ Shell wrapper installed for %s\n", result.ShellType)
		fmt.Fprintf(out, "✓ %s\n", result.Message)

		fmt.Fprintf(out, "\nTo activate the wrapper:\n")
		fmt.Fprintf(out, "  1. Restart your shell, or\n")
		fmt.Fprintf(out, "  2. Run: source ~/.bashrc (or ~/.zshrc, etc.)\n")
		fmt.Fprintf(out, "\nUsage:\n")
		fmt.Fprintf(out, "  twiggit cd <branch>     # Change to worktree\n")
		fmt.Fprintf(out, "  builtin cd <path>       # Use shell built-in cd\n")
	}

	return nil
}

// isValidShellType checks if the shell type is supported
func isValidShellType(shellType domain.ShellType) bool {
	switch shellType {
	case domain.ShellBash, domain.ShellZsh, domain.ShellFish:
		return true
	default:
		return false
	}
}
