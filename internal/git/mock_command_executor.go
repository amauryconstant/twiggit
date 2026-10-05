package git

import (
	"context"
	"fmt"
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
func (m *MockCommandExecutor) Execute(ctx context.Context, dir string, cmd Command, args ...string) (*CommandResult, error) {
	resultArgs := m.Called(ctx, dir, cmd, args)
	if mockErr := resultArgs.Error(1); mockErr != nil {
		return nil, fmt.Errorf("mock execute: %w", mockErr)
	}
	raw := resultArgs.Get(0)
	if raw == nil {
		return nil, nil
	}
	result, ok := raw.(*CommandResult)
	if !ok {
		return nil, fmt.Errorf("mock execute: unexpected return type %T", raw)
	}
	return result, nil
}

// ExecuteWithTimeout executes the mock command with timeout
func (m *MockCommandExecutor) ExecuteWithTimeout(ctx context.Context, dir string, cmd Command, timeout time.Duration, args ...string) (*CommandResult, error) {
	resultArgs := m.Called(ctx, dir, cmd, timeout, args)
	if mockErr := resultArgs.Error(1); mockErr != nil {
		return nil, fmt.Errorf("mock execute with timeout: %w", mockErr)
	}
	raw := resultArgs.Get(0)
	if raw == nil {
		return nil, nil
	}
	result, ok := raw.(*CommandResult)
	if !ok {
		return nil, fmt.Errorf("mock execute with timeout: unexpected return type %T", raw)
	}
	return result, nil
}
