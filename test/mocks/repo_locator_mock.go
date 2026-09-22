package mocks

import (
	"twiggit/internal/application"
	"twiggit/internal/domain"

	"github.com/stretchr/testify/mock"
)

var _ application.RepoLocator = (*MockRepoLocator)(nil)

// MockRepoLocator is a mock implementation of application.RepoLocator
type MockRepoLocator struct {
	mock.Mock
}

// NewMockRepoLocator creates a new MockRepoLocator
func NewMockRepoLocator() *MockRepoLocator {
	return &MockRepoLocator{}
}

// FindGitRepositories provides a mock function with given fields: dir
func (m *MockRepoLocator) FindGitRepositories(dir string) ([]domain.GitDir, error) {
	args := m.Called(dir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.GitDir), args.Error(1)
}
