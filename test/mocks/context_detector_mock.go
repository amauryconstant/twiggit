package mocks

import (
	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
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
