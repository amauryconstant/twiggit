package core

// SetupShellResult represents the result of a shell setup operation
type SetupShellResult struct {
	// ShellType indicates which shell was set up
	ShellType ShellType

	// IsInstalled indicates whether the wrapper was successfully installed
	IsInstalled bool

	// IsSkipped indicates whether the operation was skipped (already installed)
	IsSkipped bool

	// ConfigFile indicates which config file was used
	ConfigFile string

	// Message contains a human-readable message about the operation
	Message string

	// Warning contains any warnings about the operation
	Warning string
}

// ValidateInstallationResult represents the result of a shell installation validation
type ValidateInstallationResult struct {
	// ShellType indicates which shell was validated
	ShellType ShellType

	// IsInstalled indicates whether the wrapper is installed
	IsInstalled bool

	// ConfigFile indicates which config file contains the wrapper
	ConfigFile string

	// Version indicates the detected wrapper version (if available)
	Version string

	// Message contains a human-readable message about the validation
	Message string
}

// GenerateWrapperResult represents the result of a wrapper generation operation
type GenerateWrapperResult struct {
	// ShellType indicates which shell the wrapper was generated for
	ShellType ShellType

	// WrapperContent contains the generated wrapper content
	WrapperContent string

	// TemplateUsed indicates which template was used
	TemplateUsed string

	// Message contains a human-readable message about the generation
	Message string
}
