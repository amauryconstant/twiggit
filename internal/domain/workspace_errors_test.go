package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceErrorType_String(t *testing.T) {
	testCases := []struct {
		name     string
		errType  WorkspaceErrorType
		expected string
	}{
		{
			name:     "WorkspaceErrorInvalidPath",
			errType:  WorkspaceErrorInvalidPath,
			expected: "WorkspaceErrorInvalidPath",
		},
		{
			name:     "WorkspaceErrorProjectNotFound",
			errType:  WorkspaceErrorProjectNotFound,
			expected: "WorkspaceErrorProjectNotFound",
		},
		{
			name:     "WorkspaceErrorProjectAlreadyExists",
			errType:  WorkspaceErrorProjectAlreadyExists,
			expected: "WorkspaceErrorProjectAlreadyExists",
		},
		{
			name:     "WorkspaceErrorWorktreeNotFound",
			errType:  WorkspaceErrorWorktreeNotFound,
			expected: "WorkspaceErrorWorktreeNotFound",
		},
		{
			name:     "WorkspaceErrorInvalidConfiguration",
			errType:  WorkspaceErrorInvalidConfiguration,
			expected: "WorkspaceErrorInvalidConfiguration",
		},
		{
			name:     "WorkspaceErrorDiscoveryFailed",
			errType:  WorkspaceErrorDiscoveryFailed,
			expected: "WorkspaceErrorDiscoveryFailed",
		},
		{
			name:     "WorkspaceErrorValidationFailed",
			errType:  WorkspaceErrorValidationFailed,
			expected: "WorkspaceErrorValidationFailed",
		},
		{
			name:     "unknown error type",
			errType:  WorkspaceErrorType(999),
			expected: "UnknownWorkspaceError",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.errType.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNewWorkspaceError(t *testing.T) {
	testCases := []struct {
		name        string
		errType     WorkspaceErrorType
		message     string
		expectedMsg string
	}{
		{
			name:        "invalid path error",
			errType:     WorkspaceErrorInvalidPath,
			message:     "workspace path cannot be empty",
			expectedMsg: "workspace path cannot be empty",
		},
		{
			name:        "project not found error",
			errType:     WorkspaceErrorProjectNotFound,
			message:     "project 'test-project' not found",
			expectedMsg: "project 'test-project' not found",
		},
		{
			name:        "project already exists error",
			errType:     WorkspaceErrorProjectAlreadyExists,
			message:     "project 'test-project' already exists",
			expectedMsg: "project 'test-project' already exists",
		},
		{
			name:        "worktree not found error",
			errType:     WorkspaceErrorWorktreeNotFound,
			message:     "worktree not found at path '/test/path'",
			expectedMsg: "worktree not found at path '/test/path'",
		},
		{
			name:        "invalid configuration error",
			errType:     WorkspaceErrorInvalidConfiguration,
			message:     "invalid configuration key: 'invalid-key'",
			expectedMsg: "invalid configuration key: 'invalid-key'",
		},
		{
			name:        "discovery failed error",
			errType:     WorkspaceErrorDiscoveryFailed,
			message:     "failed to discover projects: permission denied",
			expectedMsg: "failed to discover projects: permission denied",
		},
		{
			name:        "validation failed error",
			errType:     WorkspaceErrorValidationFailed,
			message:     "workspace validation failed: path is invalid",
			expectedMsg: "workspace validation failed: path is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewWorkspaceError(tc.errType, tc.message)

			// Check that it's a WorkspaceError
			var workspaceErr *WorkspaceError
			require.ErrorAs(t, err, &workspaceErr, "error should be of type *WorkspaceError")

			// Check error type and message
			assert.Equal(t, tc.errType, workspaceErr.Type)
			assert.Equal(t, tc.message, workspaceErr.Message)
			assert.Equal(t, tc.expectedMsg, err.Error())
		})
	}
}

func TestWorkspaceError_Error(t *testing.T) {
	testCases := []struct {
		name     string
		errType  WorkspaceErrorType
		message  string
		expected string
	}{
		{
			name:     "invalid path error",
			errType:  WorkspaceErrorInvalidPath,
			message:  "workspace path cannot be empty",
			expected: "workspace path cannot be empty",
		},
		{
			name:     "project not found error",
			errType:  WorkspaceErrorProjectNotFound,
			message:  "project 'test' not found",
			expected: "project 'test' not found",
		},
		{
			name:     "project already exists error",
			errType:  WorkspaceErrorProjectAlreadyExists,
			message:  "project 'test' already exists",
			expected: "project 'test' already exists",
		},
		{
			name:     "worktree not found error",
			errType:  WorkspaceErrorWorktreeNotFound,
			message:  "worktree not found",
			expected: "worktree not found",
		},
		{
			name:     "invalid configuration error",
			errType:  WorkspaceErrorInvalidConfiguration,
			message:  "invalid config",
			expected: "invalid config",
		},
		{
			name:     "discovery failed error",
			errType:  WorkspaceErrorDiscoveryFailed,
			message:  "discovery failed",
			expected: "discovery failed",
		},
		{
			name:     "validation failed error",
			errType:  WorkspaceErrorValidationFailed,
			message:  "validation failed",
			expected: "validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &WorkspaceError{
				Type:    tc.errType,
				Message: tc.message,
			}

			result := err.Error()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceError_Is(t *testing.T) {
	testCases := []struct {
		name       string
		errType    WorkspaceErrorType
		targetType WorkspaceErrorType
		expected   bool
	}{
		{
			name:       "same type",
			errType:    WorkspaceErrorInvalidPath,
			targetType: WorkspaceErrorInvalidPath,
			expected:   true,
		},
		{
			name:       "different type",
			errType:    WorkspaceErrorInvalidPath,
			targetType: WorkspaceErrorProjectNotFound,
			expected:   false,
		},
		{
			name:       "project not found matches itself",
			errType:    WorkspaceErrorProjectNotFound,
			targetType: WorkspaceErrorProjectNotFound,
			expected:   true,
		},
		{
			name:       "project already exists matches itself",
			errType:    WorkspaceErrorProjectAlreadyExists,
			targetType: WorkspaceErrorProjectAlreadyExists,
			expected:   true,
		},
		{
			name:       "worktree not found matches itself",
			errType:    WorkspaceErrorWorktreeNotFound,
			targetType: WorkspaceErrorWorktreeNotFound,
			expected:   true,
		},
		{
			name:       "invalid configuration matches itself",
			errType:    WorkspaceErrorInvalidConfiguration,
			targetType: WorkspaceErrorInvalidConfiguration,
			expected:   true,
		},
		{
			name:       "discovery failed matches itself",
			errType:    WorkspaceErrorDiscoveryFailed,
			targetType: WorkspaceErrorDiscoveryFailed,
			expected:   true,
		},
		{
			name:       "validation failed matches itself",
			errType:    WorkspaceErrorValidationFailed,
			targetType: WorkspaceErrorValidationFailed,
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &WorkspaceError{
				Type:    tc.errType,
				Message: "test message",
			}

			result := IsWorkspaceErrorType(err, tc.targetType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceError_Unwrap(t *testing.T) {
	t.Run("with nil underlying error", func(t *testing.T) {
		err := &WorkspaceError{
			Type:       WorkspaceErrorInvalidPath,
			Message:    "test message",
			Underlying: nil,
		}

		result := err.Unwrap()
		assert.NoError(t, result)
	})

	t.Run("with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		err := &WorkspaceError{
			Type:       WorkspaceErrorInvalidPath,
			Message:    "test message",
			Underlying: underlyingErr,
		}

		result := err.Unwrap()
		assert.Same(t, underlyingErr, result)
	})
}

func TestWorkspaceError_As(t *testing.T) {
	t.Run("should match WorkspaceError type", func(t *testing.T) {
		workspaceErr := NewWorkspaceError(WorkspaceErrorInvalidPath, "test message")

		var target *WorkspaceError
		result := errors.As(workspaceErr, &target)

		assert.True(t, result)
		assert.Same(t, workspaceErr, target)
		assert.Equal(t, WorkspaceErrorInvalidPath, target.Type)
		assert.Equal(t, "test message", target.Message)
	})

	t.Run("should not match non-WorkspaceError type", func(t *testing.T) {
		workspaceErr := NewWorkspaceError(WorkspaceErrorInvalidPath, "test message")

		var target *ProjectError
		result := errors.As(workspaceErr, &target)

		assert.False(t, result)
		assert.Nil(t, target)
	})
}

func TestWorkspaceError_Wrapping(t *testing.T) {
	t.Run("should wrap underlying error", func(t *testing.T) {
		underlyingErr := errors.New("underlying error")
		workspaceErr := NewWorkspaceError(WorkspaceErrorInvalidPath, "workspace error")
		workspaceErr.Underlying = underlyingErr

		// Test that the error can be unwrapped
		unwrapped := errors.Unwrap(workspaceErr)
		assert.Same(t, underlyingErr, unwrapped)

		// Test error message includes underlying error
		expectedMsg := "workspace error: underlying error"
		assert.Equal(t, expectedMsg, workspaceErr.Error())
	})
}

func TestWorkspaceError_Constants(t *testing.T) {
	// Test that all error type constants have reasonable values
	assert.Equal(t, WorkspaceErrorInvalidPath, WorkspaceErrorType(1))
	assert.Equal(t, WorkspaceErrorProjectNotFound, WorkspaceErrorType(2))
	assert.Equal(t, WorkspaceErrorProjectAlreadyExists, WorkspaceErrorType(3))
	assert.Equal(t, WorkspaceErrorWorktreeNotFound, WorkspaceErrorType(4))
	assert.Equal(t, WorkspaceErrorInvalidConfiguration, WorkspaceErrorType(5))
	assert.Equal(t, WorkspaceErrorDiscoveryFailed, WorkspaceErrorType(6))
	assert.Equal(t, WorkspaceErrorValidationFailed, WorkspaceErrorType(7))
}

func TestWorkspaceError_ErrorInterface(t *testing.T) {
	t.Run("should implement error interface", func(t *testing.T) {
		err := NewWorkspaceError(WorkspaceErrorInvalidPath, "test message")

		// This should compile and run without issues
		var errorInterface error = err
		require.Error(t, errorInterface)
		assert.Equal(t, "test message", errorInterface.Error())
	})
}
