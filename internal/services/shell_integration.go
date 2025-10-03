package services

import "twiggit/internal/domain"

// ShellIntegration defines the interface for shell infrastructure operations
// This bridges the service layer with the infrastructure layer
type ShellIntegration interface {
	// GenerateWrapper generates a shell wrapper for the specified shell type
	GenerateWrapper(shellType domain.ShellType) (string, error)

	// DetectConfigFile detects the appropriate config file for the shell type
	DetectConfigFile(shellType domain.ShellType) (string, error)

	// InstallWrapper installs the wrapper to the shell config file
	InstallWrapper(shellType domain.ShellType, wrapper string) error

	// ValidateInstallation validates whether the wrapper is installed
	ValidateInstallation(shellType domain.ShellType) error
}
