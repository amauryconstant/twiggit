package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MiseIntegrationMock is a mock implementation of domain.MiseIntegration
type MiseIntegrationMock struct {
	mock.Mock
}

// NewMiseIntegrationMock creates a new MiseIntegrationMock
func NewMiseIntegrationMock() *MiseIntegrationMock {
	return &MiseIntegrationMock{}
}

// SetupWorktree mocks the SetupWorktree method
func (m *MiseIntegrationMock) SetupWorktree(sourceRepoPath, worktreePath string) error {
	args := m.Called(sourceRepoPath, worktreePath)
	return args.Error(0)
}

// IsAvailable mocks the IsAvailable method
func (m *MiseIntegrationMock) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

// DetectConfigFiles mocks the DetectConfigFiles method
func (m *MiseIntegrationMock) DetectConfigFiles(repoPath string) []string {
	args := m.Called(repoPath)
	return args.Get(0).([]string)
}

// CopyConfigFiles mocks the CopyConfigFiles method
func (m *MiseIntegrationMock) CopyConfigFiles(sourceDir, targetDir string, configFiles []string) error {
	args := m.Called(sourceDir, targetDir, configFiles)
	return args.Error(0)
}

// TrustDirectory mocks the TrustDirectory method
func (m *MiseIntegrationMock) TrustDirectory(dirPath string) error {
	args := m.Called(dirPath)
	return args.Error(0)
}

// Disable mocks the Disable method
func (m *MiseIntegrationMock) Disable() {
	m.Called()
}

// Enable mocks the Enable method
func (m *MiseIntegrationMock) Enable() {
	m.Called()
}

// IsEnabled mocks the IsEnabled method
func (m *MiseIntegrationMock) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// SetExecPath mocks the SetExecPath method
func (m *MiseIntegrationMock) SetExecPath(path string) {
	m.Called(path)
}
