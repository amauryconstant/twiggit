package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceValidationResult_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "empty result should be valid",
			result: ValidationResult{
				Valid:  true,
				Errors: []*DomainError{},
			},
			expected: true,
		},
		{
			name: "result with errors should be invalid",
			result: ValidationResult{
				Valid: false,
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
				},
			},
			expected: false,
		},
		{
			name: "result with multiple errors should be invalid",
			result: ValidationResult{
				Valid: false,
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
					{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
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
		result   ValidationResult
		expected int
	}{
		{
			name: "empty result should have zero errors",
			result: ValidationResult{
				Valid:  true,
				Errors: []*DomainError{},
			},
			expected: 0,
		},
		{
			name: "result with one error should have count of one",
			result: ValidationResult{
				Valid: false,
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
				},
			},
			expected: 1,
		},
		{
			name: "result with multiple errors should have correct count",
			result: ValidationResult{
				Valid: false,
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
					{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
					{Type: ErrWorkspaceProjectAlreadyExists, Message: "project already exists"},
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

// GetErrorsByType method was removed during consolidation, so this test is no longer applicable

func TestWorkspaceValidationResult_GetFirstError(t *testing.T) {
	testCases := []struct {
		name     string
		result   ValidationResult
		expected *DomainError
	}{
		{
			name: "should return nil for empty result",
			result: ValidationResult{
				Errors: []*DomainError{},
			},
			expected: nil,
		},
		{
			name: "should return first error for single error",
			result: ValidationResult{
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
				},
			},
			expected: &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
		},
		{
			name: "should return first error for multiple errors",
			result: ValidationResult{
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "first error"},
					{Type: ErrWorkspaceProjectNotFound, Message: "second error"},
					{Type: ErrWorkspaceProjectAlreadyExists, Message: "third error"},
				},
			},
			expected: &DomainError{Type: ErrWorkspaceInvalidPath, Message: "first error"},
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
		result := ValidationResult{
			Errors: []*DomainError{},
		}

		err := &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}
		result.AddError(err)

		require.Len(t, result.Errors, 1)
		assert.Equal(t, err, result.Errors[0])
	})

	t.Run("should add error to result with existing errors", func(t *testing.T) {
		result := ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
			},
		}

		err := &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}
		result.AddError(err)

		require.Len(t, result.Errors, 2)
		assert.Equal(t, &DomainError{Type: ErrWorkspaceProjectNotFound, Message: "project not found"}, result.Errors[0])
		assert.Equal(t, err, result.Errors[1])
	})

	t.Run("should add multiple errors", func(t *testing.T) {
		result := ValidationResult{
			Errors: []*DomainError{},
		}

		err1 := &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}
		err2 := &DomainError{Type: ErrWorkspaceProjectNotFound, Message: "project not found"}
		err3 := &DomainError{Type: ErrWorkspaceProjectAlreadyExists, Message: "project already exists"}

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
		result1 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
			},
		}

		result2 := &ValidationResult{
			Errors: []*DomainError{},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 1)
		assert.Equal(t, &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}, merged.Errors[0])
	})

	t.Run("should merge two non-empty results", func(t *testing.T) {
		result1 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
			},
		}

		result2 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
			},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 2)
		assert.Equal(t, &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}, merged.Errors[0])
		assert.Equal(t, &DomainError{Type: ErrWorkspaceProjectNotFound, Message: "project not found"}, merged.Errors[1])
	})

	t.Run("should merge results with multiple errors", func(t *testing.T) {
		result1 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceInvalidPath, Message: "invalid path 1"},
				{Type: ErrWorkspaceProjectNotFound, Message: "project not found 1"},
			},
		}

		result2 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceProjectAlreadyExists, Message: "project already exists"},
				{Type: ErrWorkspaceWorktreeNotFound, Message: "worktree not found"},
			},
		}

		merged := result1.Merge(result2)

		require.Len(t, merged.Errors, 4)
		assert.Equal(t, &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path 1"}, merged.Errors[0])
		assert.Equal(t, &DomainError{Type: ErrWorkspaceProjectNotFound, Message: "project not found 1"}, merged.Errors[1])
		assert.Equal(t, &DomainError{Type: ErrWorkspaceProjectAlreadyExists, Message: "project already exists"}, merged.Errors[2])
		assert.Equal(t, &DomainError{Type: ErrWorkspaceWorktreeNotFound, Message: "worktree not found"}, merged.Errors[3])
	})

	t.Run("should not modify original results", func(t *testing.T) {
		result1 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
			},
		}

		result2 := &ValidationResult{
			Errors: []*DomainError{
				{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
			},
		}

		merged := result1.Merge(result2)

		// Original results should be unchanged
		require.Len(t, result1.Errors, 1)
		require.Len(t, result2.Errors, 1)
		assert.Equal(t, &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"}, result1.Errors[0])
		assert.Equal(t, &DomainError{Type: ErrWorkspaceProjectNotFound, Message: "project not found"}, result2.Errors[0])

		// Merged result should have both errors
		require.Len(t, merged.Errors, 2)
	})
}

func TestWorkspaceValidationResult_ToError(t *testing.T) {
	testCases := []struct {
		name          string
		result        ValidationResult
		expectedError error
	}{
		{
			name: "should return nil for valid result",
			result: ValidationResult{
				Errors: []*DomainError{},
			},
			expectedError: nil,
		},
		{
			name: "should return first error for single error",
			result: ValidationResult{
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
				},
			},
			expectedError: &DomainError{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
		},
		{
			name: "should return first error for multiple errors",
			result: ValidationResult{
				Errors: []*DomainError{
					{Type: ErrWorkspaceInvalidPath, Message: "first error"},
					{Type: ErrWorkspaceProjectNotFound, Message: "second error"},
				},
			},
			expectedError: &DomainError{Type: ErrWorkspaceInvalidPath, Message: "first error"},
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
		result := NewWorkspaceValidationResult()
		errors := []*DomainError{
			{Type: ErrWorkspaceInvalidPath, Message: "invalid path"},
			{Type: ErrWorkspaceProjectNotFound, Message: "project not found"},
		}

		// Add errors manually since NewWorkspaceValidationResult is not variadic
		for _, err := range errors {
			result.AddError(err)
		}

		require.NotNil(t, result)
		assert.False(t, result.IsValid())
		assert.Equal(t, 2, result.GetErrorCount())
		assert.Equal(t, errors, result.Errors)
	})
}

func TestValidateWorkspacePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected *ValidationResult
	}{
		{
			name:     "valid absolute path should pass",
			path:     "/home/user/workspace",
			expected: NewWorkspaceValidationResult(),
		},
		{
			name:     "valid relative path should pass",
			path:     "./workspace",
			expected: NewWorkspaceValidationResult(),
		},
		{
			name: "empty path should fail",
			path: "",
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidPath, "workspace path cannot be empty"))
				return result
			}(),
		},
		{
			name: "whitespace-only path should fail",
			path: "   ",
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidPath, "workspace path cannot be empty"))
				return result
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspacePath(tc.path)
			assert.Equal(t, tc.expected.Valid, result.Valid)
			assert.Equal(t, tc.expected.GetErrorCount(), result.GetErrorCount())
			if tc.expected.HasErrors() {
				assert.Equal(t, tc.expected.FirstError().Type, result.FirstError().Type)
				assert.Equal(t, tc.expected.FirstError().Message, result.FirstError().Message)
			}
		})
	}
}

func TestValidateWorkspaceProjectName(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		workspace   *Workspace
		expected    *ValidationResult
	}{
		{
			name:        "valid new project name should pass",
			projectName: "new-project",
			workspace:   &Workspace{Projects: []*Project{}},
			expected:    NewWorkspaceValidationResult(),
		},
		{
			name:        "empty project name should fail",
			projectName: "",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "project name cannot be empty"))
				return result
			}(),
		},
		{
			name:        "whitespace-only project name should fail",
			projectName: "   ",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "project name cannot be empty"))
				return result
			}(),
		},
		{
			name:        "duplicate project name should fail",
			projectName: "existing-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceProjectAlreadyExists, "project 'existing-project' already exists in workspace"))
				return result
			}(),
		},
		{
			name:        "case-sensitive duplicate project name should pass",
			projectName: "Existing-Project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: NewWorkspaceValidationResult(),
		},
		{
			name:        "unique project name should pass",
			projectName: "unique-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
					{Name: "another-project", GitRepo: "/another/repo"},
				},
			},
			expected: NewWorkspaceValidationResult(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceProjectName(tc.projectName, tc.workspace)
			assert.Equal(t, tc.expected.Valid, result.Valid)
			assert.Equal(t, tc.expected.GetErrorCount(), result.GetErrorCount())
			if tc.expected.HasErrors() {
				assert.Equal(t, tc.expected.FirstError().Type, result.FirstError().Type)
				assert.Equal(t, tc.expected.FirstError().Message, result.FirstError().Message)
			}
		})
	}
}

func TestValidateWorkspaceProjectExists(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		workspace   *Workspace
		expected    *ValidationResult
	}{
		{
			name:        "existing project should pass",
			projectName: "existing-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: NewWorkspaceValidationResult(),
		},
		{
			name:        "non-existent project should fail",
			projectName: "non-existent-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceProjectNotFound, "project 'non-existent-project' not found in workspace"))
				return result
			}(),
		},
		{
			name:        "empty project name should fail",
			projectName: "",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "project name cannot be empty"))
				return result
			}(),
		},
		{
			name:        "whitespace-only project name should fail",
			projectName: "   ",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "project name cannot be empty"))
				return result
			}(),
		},
		{
			name:        "nil workspace should fail",
			projectName: "test-project",
			workspace:   nil,
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "workspace cannot be nil"))
				return result
			}(),
		},
		{
			name:        "empty workspace should fail",
			projectName: "test-project",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceProjectNotFound, "project 'test-project' not found in workspace"))
				return result
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceProjectExists(tc.projectName, tc.workspace)
			assert.Equal(t, tc.expected.Valid, result.Valid)
			assert.Equal(t, tc.expected.GetErrorCount(), result.GetErrorCount())
			if tc.expected.HasErrors() {
				assert.Equal(t, tc.expected.FirstError().Type, result.FirstError().Type)
				assert.Equal(t, tc.expected.FirstError().Message, result.FirstError().Message)
			}
		})
	}
}

func TestValidateWorkspaceCreation(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected *ValidationResult
	}{
		{
			name:     "valid path should pass",
			path:     "/home/user/workspace",
			expected: NewWorkspaceValidationResult(),
		},
		{
			name: "empty path should fail",
			path: "",
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidPath, "workspace path cannot be empty"))
				return result
			}(),
		},
		{
			name: "whitespace-only path should fail",
			path: "   ",
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidPath, "workspace path cannot be empty"))
				return result
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceCreation(tc.path)
			assert.Equal(t, tc.expected.Valid, result.Valid)
			assert.Equal(t, tc.expected.GetErrorCount(), result.GetErrorCount())
			if tc.expected.HasErrors() {
				assert.Equal(t, tc.expected.FirstError().Type, result.FirstError().Type)
				assert.Equal(t, tc.expected.FirstError().Message, result.FirstError().Message)
			}
		})
	}
}

func TestValidateWorkspaceHealth(t *testing.T) {
	testCases := []struct {
		name      string
		workspace *Workspace
		expected  *ValidationResult
	}{
		{
			name:      "valid workspace should pass",
			workspace: &Workspace{Path: "/valid/path"},
			expected:  NewWorkspaceValidationResult(),
		},
		{
			name:      "invalid path should pass in domain validation (infrastructure handles path validation)",
			workspace: &Workspace{Path: "/invalid/path"},
			expected:  NewWorkspaceValidationResult(), // Domain validation only checks for empty path
		},
		{
			name:      "nil workspace should fail",
			workspace: nil,
			expected: func() *ValidationResult {
				result := NewWorkspaceValidationResult()
				result.AddError(NewWorkspaceError(ErrWorkspaceInvalidConfiguration, "workspace cannot be nil"))
				return result
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceHealth(tc.workspace)
			assert.Equal(t, tc.expected.Valid, result.Valid)
			assert.Equal(t, tc.expected.GetErrorCount(), result.GetErrorCount())
			if tc.expected.HasErrors() {
				assert.Equal(t, tc.expected.FirstError().Type, result.FirstError().Type)
				assert.Equal(t, tc.expected.FirstError().Message, result.FirstError().Message)
			}
		})
	}
}
