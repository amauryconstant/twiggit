package mocks

import "twiggit/internal/domain"

// MockShellInfrastructure is a mock implementation of infrastructure.ShellInfrastructure for testing
type MockShellInfrastructure struct {
	// Mock functions
	GenerateWrapperFunc      func(shellType domain.ShellType) (string, error)
	DetectConfigFileFunc     func(shellType domain.ShellType) (string, error)
	InstallWrapperFunc       func(shellType domain.ShellType, wrapper, configFile string, force bool) error
	ValidateInstallationFunc func(shellType domain.ShellType, configFile string) error
}

// NewMockShellInfrastructure creates a new MockShellInfrastructure for testing
func NewMockShellInfrastructure() *MockShellInfrastructure {
	return &MockShellInfrastructure{}
}

// GenerateWrapper mocks generating shell wrapper functions
func (m *MockShellInfrastructure) GenerateWrapper(shellType domain.ShellType) (string, error) {
	if m.GenerateWrapperFunc != nil {
		return m.GenerateWrapperFunc(shellType)
	}
	return "", nil
}

// DetectConfigFile mocks detecting shell config file location
func (m *MockShellInfrastructure) DetectConfigFile(shellType domain.ShellType) (string, error) {
	if m.DetectConfigFileFunc != nil {
		return m.DetectConfigFileFunc(shellType)
	}
	return "", nil
}

// InstallWrapper mocks installing wrapper to config file
func (m *MockShellInfrastructure) InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error {
	if m.InstallWrapperFunc != nil {
		return m.InstallWrapperFunc(shellType, wrapper, configFile, force)
	}
	return nil
}

// ValidateInstallation mocks validating wrapper installation
func (m *MockShellInfrastructure) ValidateInstallation(shellType domain.ShellType, configFile string) error {
	if m.ValidateInstallationFunc != nil {
		return m.ValidateInstallationFunc(shellType, configFile)
	}
	return nil
}
