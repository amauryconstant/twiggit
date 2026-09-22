package mocks

import (
	"twiggit/internal/application"
	"twiggit/internal/core"

	"github.com/stretchr/testify/mock"
)

var _ application.ContextDetector = (*MockContextDetector)(nil)

// MockContextDetector is a mock implementation of application.ContextDetector
type MockContextDetector struct {
	mock.Mock
}

// NewMockContextDetector creates a new MockContextDetector
func NewMockContextDetector() *MockContextDetector {
	return &MockContextDetector{}
}

// DetectContext provides a mock function with given fields: dir
func (m *MockContextDetector) DetectContext(dir string) (*core.Context, error) {
	args := m.Called(dir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Context), args.Error(1)
}
