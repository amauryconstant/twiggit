package domain

// SetupShellResult represents the result of a shell setup operation
type SetupShellResult struct {
	// ShellType indicates which shell was set up
	ShellType ShellType

	// Installed indicates whether the wrapper was successfully installed
	Installed bool

	// Skipped indicates whether the operation was skipped (already installed)
	Skipped bool

	// DryRun indicates whether this was a dry run operation
	DryRun bool

	// WrapperContent contains the generated wrapper content (for dry runs)
	WrapperContent string

	// ConfigFile indicates which config file was used or would be used
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

	// Installed indicates whether the wrapper is installed
	Installed bool

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
