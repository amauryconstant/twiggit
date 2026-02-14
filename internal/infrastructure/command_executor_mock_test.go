package infrastructure

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	mock.Mock
}

// NewMockCommandExecutor creates a new MockCommandExecutor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{}
}

// Execute executes the mock command
func (m *MockCommandExecutor) Execute(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
	resultArgs := m.Called(ctx, dir, cmd, args)
	if resultArgs.Get(0) == nil {
		return nil, resultArgs.Error(1) //nolint:wrapcheck
	}
	return resultArgs.Get(0).(*CommandResult), resultArgs.Error(1) //nolint:wrapcheck
}

// ExecuteWithTimeout executes the mock command with timeout
func (m *MockCommandExecutor) ExecuteWithTimeout(ctx context.Context, dir, cmd string, timeout time.Duration, args ...string) (*CommandResult, error) {
	resultArgs := m.Called(ctx, dir, cmd, timeout, args)
	if resultArgs.Get(0) == nil {
		return nil, resultArgs.Error(1) //nolint:wrapcheck
	}
	return resultArgs.Get(0).(*CommandResult), resultArgs.Error(1) //nolint:wrapcheck
}
