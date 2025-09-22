package mocks

import (
	"github.com/stretchr/testify/mock"
)

// InfrastructureServiceMock is a mock implementation of infrastructure.InfrastructureService
type InfrastructureServiceMock struct {
	mock.Mock
}

// NewInfrastructureServiceMock creates a new InfrastructureServiceMock
func NewInfrastructureServiceMock() *InfrastructureServiceMock {
	return &InfrastructureServiceMock{}
}

// PathExists mocks the PathExists method
func (m *InfrastructureServiceMock) PathExists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

// PathWritable mocks the PathWritable method
func (m *InfrastructureServiceMock) PathWritable(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

// IsGitRepository mocks the IsGitRepository method
func (m *InfrastructureServiceMock) IsGitRepository(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}
