package shell

import "twiggit/internal/domain"

// ShellService defines the shell infrastructure service interface
type ShellService interface {
	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(shellType domain.ShellType) (string, error)

	// DetectConfigFile detects the appropriate config file for the shell type
	DetectConfigFile(shellType domain.ShellType) (string, error)

	// InstallWrapper installs the wrapper to the shell config file
	InstallWrapper(shellType domain.ShellType, wrapper string) error

	// ValidateInstallation validates whether the wrapper is installed
	ValidateInstallation(shellType domain.ShellType) error
}
