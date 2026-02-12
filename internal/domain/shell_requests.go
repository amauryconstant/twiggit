package domain

// SetupShellRequest represents a request to set up shell integration
type SetupShellRequest struct {
	// ShellType specifies the shell type to set up
	ShellType ShellType

	// Force indicates whether to force reinstall even if already installed
	Force bool

	// DryRun indicates whether to show what would be done without making changes
	DryRun bool

	// ConfigFile specifies an explicit config file to use (optional)
	ConfigFile string
}

// ValidateShellSetupRequest validates the setup shell request
func (r *SetupShellRequest) ValidateShellSetupRequest() error {
	if !IsValidShellType(r.ShellType) {
		return NewShellError(ErrInvalidShellType, string(r.ShellType), "shell type validation failed")
	}

	return nil
}

// ValidateInstallationRequest represents a request to validate shell installation
type ValidateInstallationRequest struct {
	// ShellType specifies the shell type to validate
	ShellType ShellType

	// ConfigFile specifies an explicit config file to check (optional)
	ConfigFile string
}

// ValidateValidateInstallationRequest validates the validate installation request
func (r *ValidateInstallationRequest) ValidateValidateInstallationRequest() error {
	if !IsValidShellType(r.ShellType) {
		return NewShellError(ErrInvalidShellType, string(r.ShellType), "shell type validation failed")
	}

	return nil
}

// GenerateWrapperRequest represents a request to generate a shell wrapper
type GenerateWrapperRequest struct {
	// ShellType specifies the shell type for wrapper generation
	ShellType ShellType

	// CustomTemplate allows specifying a custom wrapper template (optional)
	CustomTemplate string
}

// ValidateGenerateWrapperRequest validates the generate wrapper request
func (r *GenerateWrapperRequest) ValidateGenerateWrapperRequest() error {
	if !IsValidShellType(r.ShellType) {
		return NewShellError(ErrInvalidShellType, string(r.ShellType), "unsupported shell type")
	}

	return nil
}
