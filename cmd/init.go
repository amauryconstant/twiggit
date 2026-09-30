package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
	"twiggit/internal/output"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// InitOptions captures every input to the init run paths.
type InitOptions struct {
	IO            *iostreams.IOStreams
	Config        func() (*core.Config, error)
	GitClient     func() (cmdutil.Client, error)
	Ctx           context.Context
	GlobalOptions *cmdutil.GlobalOptions

	// Per-command fields.
	ShellType  core.ShellType
	Install    bool
	Force      bool
	ConfigFile string
}

// NewCmdInit creates a new init command.
//
// runF is the optional override used by tests; pass nil to install
// the default init body.
func NewCmdInit(f *cmdutil.Factory, runF func(*InitOptions) error) *cobra.Command {
	opts := &InitOptions{
		IO:            f.IOStreams,
		Config:        f.Config,
		GitClient:     f.GitClient,
		Ctx:           f.Context,
		GlobalOptions: f.GlobalOptions,
	}

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
		RunE: func(_ *cobra.Command, args []string) error {
			if opts.ConfigFile != "" && !opts.Install {
				return core.NewUsageError("--config requires --install", nil)
			}
			if opts.Force && !opts.Install {
				return core.NewUsageError("--force requires --install", nil)
			}

			if len(args) > 0 {
				opts.ShellType = core.ShellType(args[0])
			}

			if runF != nil {
				return runF(opts)
			}
			return runInit(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Install, "install", "i", false, "install wrapper to shell config file")
	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "", "custom config file path (requires --install)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "force reinstall even if already installed (requires --install)")

	carapace.Gen(cmd).PositionalCompletion(
		carapace.ActionValues("bash", "zsh", "fish"),
	)

	return cmd
}

// runInit dispatches to either stdout-print or file-install mode
// based on the --install flag.
func runInit(opts *InitOptions) error {
	if opts.Install {
		return runInitInstall(opts)
	}
	return runInitStdout(opts)
}

// runInitStdout writes the shell wrapper to stdout (default behavior).
// After slice 9 the wrapper comes from core.ShellWrapper rather than
// shellService.GenerateWrapper.
func runInitStdout(opts *InitOptions) error {
	shellType, err := resolveShellType(opts)
	if err != nil {
		return err
	}

	if !core.IsValidShellType(shellType) {
		verr := core.NewOpValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type")
		verr.Suggestions = []string{"Supported shells: bash, zsh, fish"}
		return verr
	}

	wrapper, err := core.ShellWrapper(shellType)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper: %w", err)
	}

	_, _ = fmt.Fprint(writeOrIgnore(opts.IO.Out), wrapper)
	return nil
}

// runInitInstall installs the wrapper to a shell config file. After
// slice 9 the orchestration lives in cmd/ with output.InstallWrapper
// providing the file-write side.
func runInitInstall(opts *InitOptions) error {
	shellType, err := resolveShellType(opts)
	if err != nil {
		return err
	}

	if !core.IsValidShellType(shellType) {
		verr := core.NewOpValidationError("ShellInit", "shellType", string(shellType), "unsupported shell type")
		verr.Suggestions = []string{"Supported shells: bash, zsh, fish"}
		return verr
	}

	configFile := opts.ConfigFile
	if configFile == "" {
		detected, err := output.DetectConfigFile(shellType)
		if err != nil {
			return fmt.Errorf("config file detection failed: %w", err)
		}
		configFile = detected
	}

	// Skip when already installed and not forcing reinstall.
	if !opts.Force {
		if err := output.ValidateInstallation(shellType, configFile); err == nil {
			result := &core.SetupShellResult{
				ShellType:   shellType,
				IsInstalled: true,
				IsSkipped:   true,
				ConfigFile:  configFile,
				Message:     "Shell wrapper already installed",
			}
			verbosef(opts.IO, "Setting up shell wrapper")
			verbosef(opts.IO, "shell type: %s", result.ShellType)
			verbosef(opts.IO, "config file: %s", result.ConfigFile)
			return displayInitResults(opts.IO.Out, result)
		}
	}

	wrapper, err := core.ShellWrapper(shellType)
	if err != nil {
		return fmt.Errorf("failed to generate wrapper: %w", err)
	}

	if err := output.InstallWrapper(shellType, wrapper, configFile, opts.Force); err != nil {
		if isAlreadyInstalled(err) {
			result := &core.SetupShellResult{
				ShellType:   shellType,
				IsInstalled: true,
				IsSkipped:   true,
				ConfigFile:  configFile,
				Message:     "Shell wrapper already installed",
			}
			return displayInitResults(opts.IO.Out, result)
		}
		return fmt.Errorf("failed to install wrapper: %w", err)
	}

	result := &core.SetupShellResult{
		ShellType:   shellType,
		IsInstalled: true,
		ConfigFile:  configFile,
		Message:     "Shell wrapper installed successfully",
	}

	verbosef(opts.IO, "Setting up shell wrapper")
	verbosef(opts.IO, "shell type: %s", result.ShellType)
	verbosef(opts.IO, "config file: %s", result.ConfigFile)

	return displayInitResults(opts.IO.Out, result)
}

// resolveShellType returns the user-supplied shell type or auto-
// detects from $SHELL when empty.
func resolveShellType(opts *InitOptions) (core.ShellType, error) {
	if opts.ShellType != "" {
		return opts.ShellType, nil
	}
	detected, err := git.DetectShellFromEnv()
	if err != nil {
		return "", fmt.Errorf("shell auto-detection failed: %w", err)
	}
	return detected, nil
}

// isAlreadyInstalled reports whether err is the typed shell-already-installed
// sentinel from output.InstallWrapper. Matches via errors.Is per
// domain-typed-errors req 8.
func isAlreadyInstalled(err error) bool {
	return errors.Is(err, core.ErrShellAlreadyInstalled)
}

// displayInitResults outputs installation results (for install mode only)
func displayInitResults(out io.Writer, result *core.SetupShellResult) error {
	out = writeOrIgnore(out)
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
