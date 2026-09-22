package mocks

import (
	"twiggit/internal/application"
	"twiggit/internal/core"

	"github.com/stretchr/testify/mock"
)

var _ application.ShellInfrastructure = (*MockShellInfrastructure)(nil)

// MockShellInfrastructure is a mock implementation of application.ShellInfrastructure for testing
type MockShellInfrastructure struct {
	mock.Mock
}

// NewMockShellInfrastructure creates a new MockShellInfrastructure for testing
func NewMockShellInfrastructure() *MockShellInfrastructure {
	return &MockShellInfrastructure{}
}

// GenerateWrapper mocks generating shell wrapper functions
func (m *MockShellInfrastructure) GenerateWrapper(shellType core.ShellType) (string, error) {
	args := m.Called(shellType)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// DetectConfigFile mocks detecting shell config file location
func (m *MockShellInfrastructure) DetectConfigFile(shellType core.ShellType) (string, error) {
	args := m.Called(shellType)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// InstallWrapper mocks installing wrapper to config file
func (m *MockShellInfrastructure) InstallWrapper(shellType core.ShellType, wrapper, configFile string, force bool) error {
	args := m.Called(shellType, wrapper, configFile, force)
	return args.Error(0)
}

// ValidateInstallation mocks validating wrapper installation
func (m *MockShellInfrastructure) ValidateInstallation(shellType core.ShellType, configFile string) error {
	args := m.Called(shellType, configFile)
	return args.Error(0)
}

// ComposeWrapper mocks composing custom template with placeholders
func (m *MockShellInfrastructure) ComposeWrapper(template string, shellType core.ShellType) string {
	args := m.Called(template, shellType)
	return args.String(0)
}
