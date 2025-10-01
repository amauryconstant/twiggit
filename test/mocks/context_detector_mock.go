package mocks

import (
	"github.com/stretchr/testify/mock"
	"twiggit/internal/domain"
)

// ContextDetectorMock is a mock implementation of domain.ContextDetector
type ContextDetectorMock struct {
	mock.Mock
}

// DetectContext provides a mock function with given fields: dir
func (m *ContextDetectorMock) DetectContext(dir string) (*domain.Context, error) {
	args := m.Called(dir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Context), args.Error(1)
}

// InvalidateCacheForRepo provides a mock function with given fields: repoPath
func (m *ContextDetectorMock) InvalidateCacheForRepo(repoPath string) {
	m.Called(repoPath)
}

// ClearCache provides a mock function
func (m *ContextDetectorMock) ClearCache() {
	m.Called()
}
