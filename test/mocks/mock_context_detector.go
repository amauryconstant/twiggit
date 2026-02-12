package mocks

import (
	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockContextDetector is a mock implementation of domain.ContextDetector
type MockContextDetector struct {
	mock.Mock
}

// NewMockContextDetector creates a new MockContextDetector
func NewMockContextDetector() *MockContextDetector {
	return &MockContextDetector{}
}

// DetectContext provides a mock function with given fields: dir
func (m *MockContextDetector) DetectContext(dir string) (*domain.Context, error) {
	args := m.Called(dir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Context), args.Error(1)
}
