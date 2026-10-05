package core

// RequestWithShellType interface for requests that have a shell type field
type RequestWithShellType interface {
	ShellType() ShellType
}

// ValidateShellTypeRequest validates a request with a shell type field
func ValidateShellTypeRequest(req RequestWithShellType) error {
	if !IsValidShellType(req.ShellType()) {
		return &ValidationError{
			Op:          "ShellValidation",
			Field:       "shellType",
			Value:       string(req.ShellType()),
			Message:     "unsupported shell type",
			Suggestions: []string{"Supported shells: bash, zsh, fish"},
		}
	}
	return nil
}

// SetupShellRequest represents a request to set up shell wrapper.
// The shellType field is unexported; use NewSetupShellRequest to
// construct. Per golang-naming, the ShellType method follows the
// noun-only convention (no Get prefix).
type SetupShellRequest struct {
	shellType ShellType

	// ConfigFile specifies an explicit config file to use (optional)
	ConfigFile string

	// IsForceOverwrite specifies whether to overwrite existing wrapper
	IsForceOverwrite bool
}

// NewSetupShellRequest constructs a SetupShellRequest with the
// supplied shell type. ConfigFile and IsForceOverwrite stay zero; set
// them on the returned struct if needed.
func NewSetupShellRequest(shellType ShellType) *SetupShellRequest {
	return &SetupShellRequest{shellType: shellType}
}

// ShellType returns the shell type for validation
func (r *SetupShellRequest) ShellType() ShellType {
	return r.shellType
}

// Validate validates setup shell request
func (r *SetupShellRequest) Validate() error {
	return ValidateShellTypeRequest(r)
}

// ValidateInstallationRequest represents a request to validate shell installation
type ValidateInstallationRequest struct {
	shellType ShellType

	// ConfigFile specifies an explicit config file to check (optional)
	ConfigFile string
}

// NewValidateInstallationRequest constructs a ValidateInstallationRequest.
func NewValidateInstallationRequest(shellType ShellType) *ValidateInstallationRequest {
	return &ValidateInstallationRequest{shellType: shellType}
}

// ShellType returns the shell type for validation
func (r *ValidateInstallationRequest) ShellType() ShellType {
	return r.shellType
}

// Validate validates the validate installation request
func (r *ValidateInstallationRequest) Validate() error {
	return ValidateShellTypeRequest(r)
}

// GenerateWrapperRequest represents a request to generate a shell wrapper
type GenerateWrapperRequest struct {
	shellType ShellType

	// CustomTemplate allows specifying a custom wrapper template (optional)
	CustomTemplate string
}

// NewGenerateWrapperRequest constructs a GenerateWrapperRequest.
func NewGenerateWrapperRequest(shellType ShellType) *GenerateWrapperRequest {
	return &GenerateWrapperRequest{shellType: shellType}
}

// ShellType returns the shell type for validation
func (r *GenerateWrapperRequest) ShellType() ShellType {
	return r.shellType
}

// Validate validates the generate wrapper request
func (r *GenerateWrapperRequest) Validate() error {
	if !IsValidShellType(r.shellType) {
		return &ValidationError{
			Op:          "GenerateWrapper",
			Field:       "shellType",
			Value:       string(r.shellType),
			Message:     "unsupported shell type",
			Suggestions: []string{"Supported shells: bash, zsh, fish"},
		}
	}
	return nil
}
