package mocks

import (
	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockShellInfrastructure is a mock implementation of infrastructure.ShellInfrastructure for testing
type MockShellInfrastructure struct {
	mock.Mock
}

// NewMockShellInfrastructure creates a new MockShellInfrastructure for testing
func NewMockShellInfrastructure() *MockShellInfrastructure {
	return &MockShellInfrastructure{}
}

// GenerateWrapper mocks generating shell wrapper functions
func (m *MockShellInfrastructure) GenerateWrapper(shellType domain.ShellType) (string, error) {
	args := m.Called(shellType)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// DetectConfigFile mocks detecting shell config file location
func (m *MockShellInfrastructure) DetectConfigFile(shellType domain.ShellType) (string, error) {
	args := m.Called(shellType)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// InstallWrapper mocks installing wrapper to config file
func (m *MockShellInfrastructure) InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error {
	args := m.Called(shellType, wrapper, configFile, force)
	return args.Error(0)
}

// ValidateInstallation mocks validating wrapper installation
func (m *MockShellInfrastructure) ValidateInstallation(shellType domain.ShellType, configFile string) error {
	args := m.Called(shellType, configFile)
	return args.Error(0)
}
