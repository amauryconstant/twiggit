// Package cmd contains the CLI command definitions for twiggit
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/amaury/twiggit/internal/di"
	"github.com/spf13/cobra"
)

// ShellType represents the type of shell being used
type ShellType string

const (
	// ShellBash represents the bash shell
	ShellBash ShellType = "bash"
	// ShellZsh represents the zsh shell
	ShellZsh ShellType = "zsh"
	// ShellFish represents the fish shell
	ShellFish ShellType = "fish"
	// ShellUnknown represents an unknown shell type
	ShellUnknown ShellType = "unknown"
)

// ShellInfo contains information about the detected shell
type ShellInfo struct {
	Type        ShellType
	Path        string
	ConfigFiles []string
}

// NewSetupShellCmd creates the setup-shell command
func NewSetupShellCmd(container *di.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-shell",
		Short: "Setup shell integration for twiggit",
		Long: `Setup shell integration for twiggit.

This command detects your current shell and provides the necessary
wrapper functions to enable seamless directory switching with twiggit.

The wrapper functions will:
- Intercept 'twiggit cd' calls and change directories using 'builtin cd'
- Provide an escape hatch with 'builtin cd' for the shell's built-in command
- Pass through all other 'twiggit' commands unchanged
- Show appropriate warnings about shell built-in override

Examples:
  twiggit setup-shell              # Detect current shell and show setup instructions
  twiggit setup-shell --shell bash # Force bash shell setup
  twiggit setup-shell --shell zsh  # Force zsh shell setup
  twiggit setup-shell --shell fish # Force fish shell setup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupShellCommand(cmd, args, container)
		},
	}

	// Add flags
	cmd.Flags().String("shell", "", "Force specific shell type (bash, zsh, fish)")

	return cmd
}

// runSetupShellCommand implements the setup-shell command functionality
func runSetupShellCommand(cmd *cobra.Command, _ []string, _ *di.Container) error {
	// Get shell type from flag or auto-detect
	shellFlag, _ := cmd.Flags().GetString("shell")
	var shellInfo *ShellInfo
	var err error

	if shellFlag != "" {
		// Use specified shell type
		shellType := parseShellType(shellFlag)
		if shellType == ShellUnknown {
			return fmt.Errorf("unsupported shell type: %s. Supported shells: bash, zsh, fish", shellFlag)
		}
		shellInfo = &ShellInfo{
			Type:        shellType,
			Path:        getDefaultShellPath(shellType),
			ConfigFiles: getConfigFilesForShell(shellType),
		}
	} else {
		// Auto-detect shell
		shellInfo, err = detectShell()
		if err != nil {
			return fmt.Errorf("failed to detect shell: %w", err)
		}
	}

	// Generate and display setup instructions
	return displaySetupInstructions(shellInfo)
}

// detectShell automatically detects the current shell
func detectShell() (*ShellInfo, error) {
	// Try multiple detection methods

	// Method 1: Check SHELL environment variable
	if shellPath := os.Getenv("SHELL"); shellPath != "" {
		shellType := getShellTypeFromPath(shellPath)
		if shellType != ShellUnknown {
			return &ShellInfo{
				Type:        shellType,
				Path:        shellPath,
				ConfigFiles: getConfigFilesForShell(shellType),
			}, nil
		}
	}

	// Method 2: Check process name (works better on some systems)
	if procName := os.Getenv("0"); procName != "" {
		shellType := getShellTypeFromPath(procName)
		if shellType != ShellUnknown {
			return &ShellInfo{
				Type:        shellType,
				Path:        procName,
				ConfigFiles: getConfigFilesForShell(shellType),
			}, nil
		}
	}

	// Method 3: Try to get parent process name (Unix-like systems)
	if runtime.GOOS != "windows" {
		if parentShell, err := getParentProcessShell(); err == nil && parentShell != "" {
			shellType := getShellTypeFromPath(parentShell)
			if shellType != ShellUnknown {
				return &ShellInfo{
					Type:        shellType,
					Path:        parentShell,
					ConfigFiles: getConfigFilesForShell(shellType),
				}, nil
			}
		}
	}

	// Method 4: Try common shell paths
	commonShells := []string{
		"/bin/bash", "/usr/bin/bash", "/usr/local/bin/bash",
		"/bin/zsh", "/usr/bin/zsh", "/usr/local/bin/zsh",
		"/bin/fish", "/usr/bin/fish", "/usr/local/bin/fish",
	}

	for _, shellPath := range commonShells {
		if _, err := os.Stat(shellPath); err == nil {
			shellType := getShellTypeFromPath(shellPath)
			if shellType != ShellUnknown {
				return &ShellInfo{
					Type:        shellType,
					Path:        shellPath,
					ConfigFiles: getConfigFilesForShell(shellType),
				}, nil
			}
		}
	}

	return nil, errors.New("could not detect shell type. Please specify shell type with --shell flag")
}

// getShellTypeFromPath determines shell type from executable path
func getShellTypeFromPath(path string) ShellType {
	base := filepath.Base(path)
	switch {
	case strings.Contains(base, "bash"):
		return ShellBash
	case strings.Contains(base, "zsh"):
		return ShellZsh
	case strings.Contains(base, "fish"):
		return ShellFish
	default:
		return ShellUnknown
	}
}

// parseShellType converts string to ShellType
func parseShellType(shell string) ShellType {
	switch strings.ToLower(shell) {
	case "bash":
		return ShellBash
	case "zsh":
		return ShellZsh
	case "fish":
		return ShellFish
	default:
		return ShellUnknown
	}
}

// getDefaultShellPath returns the default path for a shell type
func getDefaultShellPath(shellType ShellType) string {
	switch shellType {
	case ShellBash:
		return "/bin/bash"
	case ShellZsh:
		return "/bin/zsh"
	case ShellFish:
		return "/usr/bin/fish"
	default:
		return ""
	}
}

// getConfigFilesForShell returns configuration files for a shell type
func getConfigFilesForShell(shellType ShellType) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	switch shellType {
	case ShellBash:
		return []string{
			filepath.Join(homeDir, ".bashrc"),
			filepath.Join(homeDir, ".bash_profile"),
			filepath.Join(homeDir, ".profile"),
		}
	case ShellZsh:
		return []string{
			filepath.Join(homeDir, ".zshrc"),
			filepath.Join(homeDir, ".zprofile"),
		}
	case ShellFish:
		return []string{
			filepath.Join(homeDir, ".config", "fish", "config.fish"),
		}
	default:
		return []string{}
	}
}

// getParentProcessShell gets the parent process shell (Unix-like systems)
func getParentProcessShell() (string, error) {
	// Try to get parent process name using ps command
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(os.Getppid()))
	case "darwin":
		cmd = exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(os.Getppid()))
	default:
		return "", errors.New("unsupported platform for parent process detection")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute ps command: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// displaySetupInstructions shows the user how to set up shell integration
func displaySetupInstructions(shellInfo *ShellInfo) error {
	fmt.Printf("🔧 Shell Integration Setup\n")
	fmt.Printf("============================\n\n")

	fmt.Printf("Detected shell: %s (%s)\n\n", shellInfo.Type, shellInfo.Path)

	// Show wrapper function
	fmt.Printf("📋 Add the following function to your shell configuration:\n")
	fmt.Printf("--------------------------------------------------------\n\n")

	wrapperFunc := generateWrapperFunction(shellInfo.Type)
	fmt.Printf("%s\n\n", wrapperFunc)

	// Show configuration file suggestions
	fmt.Printf("📁 Configuration files to modify:\n")
	fmt.Printf("==================================\n")

	for i, configFile := range shellInfo.ConfigFiles {
		exists := ""
		if _, err := os.Stat(strings.Replace(configFile, "~", os.Getenv("HOME"), 1)); err == nil {
			exists = " ✓ (exists)"
		}
		fmt.Printf("%d. %s%s\n", i+1, configFile, exists)
	}

	fmt.Printf("\n")

	// Show instructions
	fmt.Printf("📖 Setup Instructions:\n")
	fmt.Printf("=======================\n")
	fmt.Printf("1. Choose one of the configuration files listed above\n")
	fmt.Printf("2. Add the wrapper function shown above to that file\n")
	fmt.Printf("3. Reload your shell or source the configuration file:\n")

	switch shellInfo.Type {
	case ShellBash:
		fmt.Printf("   source ~/.bashrc  # or your chosen file\n")
	case ShellZsh:
		fmt.Printf("   source ~/.zshrc   # or your chosen file\n")
	case ShellFish:
		fmt.Printf("   source ~/.config/fish/config.fish  # or your chosen file\n")
	}

	fmt.Printf("\n")
	fmt.Printf("✅ Verification:\n")
	fmt.Printf("================\n")
	fmt.Printf("After setup, you can test with:\n")
	fmt.Printf("  twiggit cd --help\n\n")

	fmt.Printf("⚠️  Important Notes:\n")
	fmt.Printf("====================\n")
	fmt.Printf("• The wrapper function overrides the shell's built-in 'cd' for twiggit commands\n")
	fmt.Printf("• Use 'builtin cd' to access the shell's original cd command\n")
	fmt.Printf("• All other twiggit commands (list, create, delete) work normally\n")
	fmt.Printf("• The wrapper only affects 'twiggit cd' commands\n")

	// Add zfunctions note for zsh users
	if shellInfo.Type == ShellZsh {
		fmt.Printf("\n💡 Advanced users: For zsh functions integration, see documentation.\n")
	}

	return nil
}

// generateWrapperFunction creates the shell-specific wrapper function
func generateWrapperFunction(shellType ShellType) string {
	switch shellType {
	case ShellBash:
		return `# twiggit shell integration wrapper for bash
twiggit() {
    if [[ "$1" == "cd" ]]; then
        # Handle twiggit cd command
        local target_path
        target_path=$(command twiggit cd "${@:2}" 2>&1)
        local exit_code=$?
        if [[ $exit_code -eq 0 ]]; then
            builtin cd "$target_path"
        else
            echo "$target_path" >&2
            return $exit_code
        fi
    else
        # Pass through all other twiggit commands
        command twiggit "$@"
    fi
}

# Warning: This wrapper function overrides the shell's built-in 'cd' for twiggit commands.
# Use 'builtin cd' to access the shell's original cd command.`

	case ShellZsh:
		return `# twiggit shell integration wrapper for zsh
twiggit() {
    if [[ "$1" == "cd" ]]; then
        # Handle twiggit cd command
        local target_path
        target_path=$(command twiggit cd "${@:2}" 2>&1)
        local exit_code=$?
        if [[ $exit_code -eq 0 ]]; then
            builtin cd "$target_path"
        else
            echo "$target_path" >&2
            return $exit_code
        fi
    else
        # Pass through all other twiggit commands
        command twiggit "$@"
    fi
}

# Warning: This wrapper function overrides the shell's built-in 'cd' for twiggit commands.
# Use 'builtin cd' to access the shell's original cd command.`

	case ShellFish:
		return `# twiggit shell integration wrapper for fish
function twiggit
    if test "$argv[1]" = "cd"
        # Handle twiggit cd command
        set target_path (command twiggit cd $argv[2..-1] 2>&1)
        set exit_code $status
        if test $exit_code -eq 0
            builtin cd $target_path
        else
            echo $target_path >&2
            return $exit_code
        end
    else
        # Pass through all other twiggit commands
        command twiggit $argv
    end
end

# Warning: This wrapper function overrides the shell's built-in 'cd' for twiggit commands.
# Use 'builtin cd' to access the shell's original cd command.`

	default:
		return "# Unsupported shell type"
	}
}
