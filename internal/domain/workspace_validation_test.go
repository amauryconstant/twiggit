package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceValidationResult_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		result   WorkspaceValidationResult
		expected bool
	}{
		{
			name: "empty result should be valid",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
			expected: true,
		},
		{
			name: "result with errors should be invalid",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				},
			},
			expected: false,
		},
		{
			name: "result with multiple errors should be invalid",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
					{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.IsValid()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceValidationResult_GetErrorCount(t *testing.T) {
	testCases := []struct {
		name     string
		result   WorkspaceValidationResult
		expected int
	}{
		{
			name: "empty result should have zero errors",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
			expected: 0,
		},
		{
			name: "result with one error should have count of one",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				},
			},
			expected: 1,
		},
		{
			name: "result with multiple errors should have correct count",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
					{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
					{Type: WorkspaceErrorProjectAlreadyExists, Message: "project already exists"},
				},
			},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.GetErrorCount()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceValidationResult_GetErrorsByType(t *testing.T) {
	testCases := []struct {
		name      string
		result    WorkspaceValidationResult
		errorType WorkspaceErrorType
		expected  []WorkspaceError
	}{
		{
			name: "should return empty slice for no matching errors",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				},
			},
			errorType: WorkspaceErrorProjectNotFound,
			expected:  []WorkspaceError{},
		},
		{
			name: "should return single matching error",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
					{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
				},
			},
			errorType: WorkspaceErrorProjectNotFound,
			expected: []WorkspaceError{
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
			},
		},
		{
			name: "should return multiple matching errors",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path 1"},
					{Type: WorkspaceErrorProjectNotFound, Message: "project not found 1"},
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path 2"},
					{Type: WorkspaceErrorProjectNotFound, Message: "project not found 2"},
				},
			},
			errorType: WorkspaceErrorProjectNotFound,
			expected: []WorkspaceError{
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found 1"},
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found 2"},
			},
		},
		{
			name: "should return all errors when type matches all",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
					{Type: WorkspaceErrorInvalidPath, Message: "another invalid path"},
				},
			},
			errorType: WorkspaceErrorInvalidPath,
			expected: []WorkspaceError{
				{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				{Type: WorkspaceErrorInvalidPath, Message: "another invalid path"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.GetErrorsByType(tc.errorType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceValidationResult_GetFirstError(t *testing.T) {
	testCases := []struct {
		name     string
		result   WorkspaceValidationResult
		expected *WorkspaceError
	}{
		{
			name: "should return nil for empty result",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
			expected: nil,
		},
		{
			name: "should return first error for single error",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				},
			},
			expected: &WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
		},
		{
			name: "should return first error for multiple errors",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "first error"},
					{Type: WorkspaceErrorProjectNotFound, Message: "second error"},
					{Type: WorkspaceErrorProjectAlreadyExists, Message: "third error"},
				},
			},
			expected: &WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "first error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.GetFirstError()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWorkspaceValidationResult_AddError(t *testing.T) {
	t.Run("should add error to empty result", func(t *testing.T) {
		result := WorkspaceValidationResult{
			Errors: []WorkspaceError{},
		}

		err := WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}
		result.AddError(err)

		require.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
	})

	t.Run("should add error to result with existing errors", func(t *testing.T) {
		result := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
			},
		}

		err := WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}
		result.AddError(err)

		require.Len(t, result.Errors, 2)
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorProjectNotFound, Message: "project not found"}, result.Errors[0])
		assert.Equal(t, err, result.Errors[1])
	})

	t.Run("should add multiple errors", func(t *testing.T) {
		result := WorkspaceValidationResult{
			Errors: []WorkspaceError{},
		}

		err1 := WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}
		err2 := WorkspaceError{Type: WorkspaceErrorProjectNotFound, Message: "project not found"}
		err3 := WorkspaceError{Type: WorkspaceErrorProjectAlreadyExists, Message: "project already exists"}

		result.AddError(err1)
		result.AddError(err2)
		result.AddError(err3)

		require.Len(t, result.Errors, 3)
		assert.Equal(t, err1, result.Errors[0])
		assert.Equal(t, err2, result.Errors[1])
		assert.Equal(t, err3, result.Errors[2])
	})
}

func TestWorkspaceValidationResult_Merge(t *testing.T) {
	t.Run("should merge with empty result", func(t *testing.T) {
		result1 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
			},
		}

		result2 := WorkspaceValidationResult{
			Errors: []WorkspaceError{},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 1)
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}, merged.Errors[0])
	})

	t.Run("should merge two non-empty results", func(t *testing.T) {
		result1 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
			},
		}

		result2 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
			},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 2)
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}, merged.Errors[0])
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorProjectNotFound, Message: "project not found"}, merged.Errors[1])
	})

	t.Run("should merge results with multiple errors", func(t *testing.T) {
		result1 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorInvalidPath, Message: "invalid path 1"},
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found 1"},
			},
		}

		result2 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorProjectAlreadyExists, Message: "project already exists"},
				{Type: WorkspaceErrorWorktreeNotFound, Message: "worktree not found"},
			},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 4)
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path 1"}, merged.Errors[0])
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorProjectNotFound, Message: "project not found 1"}, merged.Errors[1])
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorProjectAlreadyExists, Message: "project already exists"}, merged.Errors[2])
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorWorktreeNotFound, Message: "worktree not found"}, merged.Errors[3])
	})

	t.Run("should not modify original results", func(t *testing.T) {
		result1 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
			},
		}

		result2 := WorkspaceValidationResult{
			Errors: []WorkspaceError{
				{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
			},
		}

		merged := result1.Merge(result2)

		// Original results should be unchanged
		require.Len(t, result1.Errors, 1)
		require.Len(t, result2.Errors, 1)
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"}, result1.Errors[0])
		assert.Equal(t, WorkspaceError{Type: WorkspaceErrorProjectNotFound, Message: "project not found"}, result2.Errors[0])

		// Merged result should have both errors
		require.Len(t, merged.Errors, 2)
	})
}

func TestWorkspaceValidationResult_ToError(t *testing.T) {
	testCases := []struct {
		name          string
		result        WorkspaceValidationResult
		expectedError error
	}{
		{
			name: "should return nil for valid result",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
			expectedError: nil,
		},
		{
			name: "should return first error for single error",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
				},
			},
			expectedError: &WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
		},
		{
			name: "should return first error for multiple errors",
			result: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "first error"},
					{Type: WorkspaceErrorProjectNotFound, Message: "second error"},
				},
			},
			expectedError: &WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "first error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.ToError()
			assert.Equal(t, tc.expectedError, result)
		})
	}
}

func TestNewWorkspaceValidationResult(t *testing.T) {
	t.Run("should create empty result", func(t *testing.T) {
		result := NewWorkspaceValidationResult()

		require.NotNil(t, result)
		assert.True(t, result.IsValid())
		assert.Equal(t, 0, result.GetErrorCount())
		assert.Empty(t, result.Errors)
	})

	t.Run("should create result with initial errors", func(t *testing.T) {
		errors := []WorkspaceError{
			{Type: WorkspaceErrorInvalidPath, Message: "invalid path"},
			{Type: WorkspaceErrorProjectNotFound, Message: "project not found"},
		}

		result := NewWorkspaceValidationResult(errors...)

		require.NotNil(t, result)
		assert.False(t, result.IsValid())
		assert.Equal(t, 2, result.GetErrorCount())
		assert.Equal(t, errors, result.Errors)
	})
}
