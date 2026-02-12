package domain

// RequestWithShellType interface for requests that have a shell type field
type RequestWithShellType interface {
	GetShellType() ShellType
}

// ValidateShellTypeRequest validates a request with a shell type field
func ValidateShellTypeRequest(req RequestWithShellType) error {
	if !IsValidShellType(req.GetShellType()) {
		return NewValidationError("ShellValidation", "shellType", string(req.GetShellType()), "unsupported shell type").
			WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
	}
	return nil
}

// SetupShellRequest represents a request to set up shell wrapper
type SetupShellRequest struct {
	// ShellType specifies shell type for wrapper setup
	ShellType ShellType

	// ConfigFile specifies an explicit config file to use (optional)
	ConfigFile string

	// ForceOverwrite specifies whether to overwrite existing wrapper
	ForceOverwrite bool

	// DryRun specifies whether to perform a dry run without making changes
	DryRun bool
}

// GetShellType returns the shell type for validation
func (r *SetupShellRequest) GetShellType() ShellType {
	return r.ShellType
}

// ValidateShellSetupRequest validates setup shell request
func (r *SetupShellRequest) ValidateShellSetupRequest() error {
	return ValidateShellTypeRequest(r)
}

// ValidateInstallationRequest represents a request to validate shell installation
type ValidateInstallationRequest struct {
	// ShellType specifies shell type to validate
	ShellType ShellType

	// ConfigFile specifies an explicit config file to check (optional)
	ConfigFile string
}

// GetShellType returns the shell type for validation
func (r *ValidateInstallationRequest) GetShellType() ShellType {
	return r.ShellType
}

// ValidateValidateInstallationRequest validates the validate installation request
func (r *ValidateInstallationRequest) ValidateValidateInstallationRequest() error {
	return ValidateShellTypeRequest(r)
}

// GenerateWrapperRequest represents a request to generate a shell wrapper
type GenerateWrapperRequest struct {
	// ShellType specifies shell type for wrapper generation
	ShellType ShellType

	// CustomTemplate allows specifying a custom wrapper template (optional)
	CustomTemplate string
}

// GetShellType returns the shell type for validation
func (r *GenerateWrapperRequest) GetShellType() ShellType {
	return r.ShellType
}

// ValidateGenerateWrapperRequest validates the generate wrapper request
func (r *GenerateWrapperRequest) ValidateGenerateWrapperRequest() error {
	if !IsValidShellType(r.ShellType) {
		return NewValidationError("GenerateWrapper", "shellType", string(r.ShellType), "unsupported shell type").
			WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
	}
	return nil
}
