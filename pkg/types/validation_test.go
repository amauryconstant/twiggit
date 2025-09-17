package types

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationResult_AddError(t *testing.T) {
	result := NewValidationResult()
	assert.True(t, result.Valid)

	err := NewWorktreeError(ErrInvalidBranchName, "test error", "")
	result.AddError(err)

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, err, result.Errors[0])
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := NewValidationResult()
	result.AddWarning("test warning")

	assert.True(t, result.Valid) // Warnings don't affect validity
	assert.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings, "test warning")
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := NewValidationResult()
	assert.False(t, result.HasErrors())

	result.AddError(NewWorktreeError(ErrValidation, "test", ""))
	assert.True(t, result.HasErrors())
}

func TestValidationResult_FirstError(t *testing.T) {
	result := NewValidationResult()
	assert.Nil(t, result.FirstError())

	err1 := NewWorktreeError(ErrValidation, "first error", "")
	err2 := NewWorktreeError(ErrValidation, "second error", "")

	result.AddError(err1)
	result.AddError(err2)

	assert.Equal(t, err1, result.FirstError())
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expectValid bool
		errorType   ErrorType
		description string
	}{
		{"valid simple branch", "feature", true, ErrUnknown, ""},
		{"valid with numbers", "feature123", true, ErrUnknown, ""},
		{"valid with dash", "feature-branch", true, ErrUnknown, ""},
		{"valid with slash", "feature/auth", true, ErrUnknown, ""},
		{"valid with dots", "v1.2.3", true, ErrUnknown, ""},

		{"empty branch name", "", false, ErrInvalidBranchName, "branch name cannot be empty"},
		{"branch with space", "feature branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with tab", "feature\tbranch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch starting with dash", "-feature", false, ErrInvalidBranchName, "branch name cannot start with a hyphen"},
		{"branch ending with .lock", "feature.lock", false, ErrInvalidBranchName, "branch name cannot end with '.lock'"},
		{"branch with double dots", "feature..branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with double slash", "feature//branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with tilde", "feature~branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with caret", "feature^branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with colon", "feature:branch", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with question", "feature?", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with asterisk", "feature*", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with bracket", "feature[1]", false, ErrInvalidBranchName, "branch name contains invalid character"},
		{"branch with backslash", "feature\\branch", false, ErrInvalidBranchName, "branch name contains invalid character"},

		{"too long branch", strings.Repeat("a", MaxBranchNameLength+1), false, ErrInvalidBranchName, "branch name too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateBranchName(tt.branchName)

			assert.Equal(t, tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				assert.True(t, result.HasErrors(), "Should have errors when invalid")
				assert.True(t, IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					assert.Contains(t, result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectValid bool
		errorType   ErrorType
		description string
	}{
		{"valid absolute path", "/tmp/test", true, ErrUnknown, ""},
		{"valid absolute path with subdirs", "/home/user/projects/test", true, ErrUnknown, ""},

		{"empty path", "", false, ErrInvalidPath, "path cannot be empty"},
		{"relative path", "relative/path", false, ErrInvalidPath, "path must be absolute"},
		{"path with null character", "/tmp/test\x00", false, ErrInvalidPath, "path contains null character"},
		{"too long path", "/" + strings.Repeat("a", 4097), false, ErrInvalidPath, "path too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePath(tt.path)

			assert.Equal(t, tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				assert.True(t, result.HasErrors(), "Should have errors when invalid")
				assert.True(t, IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					assert.Contains(t, result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func TestValidatePathWritable(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "twiggit-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		setupPath   func() string
		expectValid bool
		errorType   ErrorType
	}{
		{
			name: "valid writable path",
			setupPath: func() string {
				return filepath.Join(tempDir, "new-path")
			},
			expectValid: true,
		},
		{
			name: "path already exists",
			setupPath: func() string {
				existingPath := filepath.Join(tempDir, "existing")
				require.NoError(t, os.MkdirAll(existingPath, 0755))
				return existingPath
			},
			expectValid: false,
			errorType:   ErrPathNotWritable,
		},
		{
			name: "parent directory doesn't exist",
			setupPath: func() string {
				return filepath.Join(tempDir, "non-existent", "path")
			},
			expectValid: false,
			errorType:   ErrPathNotWritable,
		},
		{
			name: "invalid path format",
			setupPath: func() string {
				return "relative/path"
			},
			expectValid: false,
			errorType:   ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath()
			result := ValidatePathWritable(path)

			assert.Equal(t, tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				assert.True(t, result.HasErrors(), "Should have errors when invalid")
				assert.True(t, IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
			}
		})
	}
}

func TestValidateWorktreeCreation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "twiggit-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		branchName  string
		targetPath  string
		expectValid bool
	}{
		{
			name:        "valid branch and path",
			branchName:  "feature-branch",
			targetPath:  filepath.Join(tempDir, "new-worktree"),
			expectValid: true,
		},
		{
			name:        "invalid branch name",
			branchName:  "invalid branch name",
			targetPath:  filepath.Join(tempDir, "valid-path"),
			expectValid: false,
		},
		{
			name:        "invalid path",
			branchName:  "valid-branch",
			targetPath:  "relative/path",
			expectValid: false,
		},
		{
			name:        "both invalid",
			branchName:  "invalid branch",
			targetPath:  "relative/path",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateWorktreeCreation(tt.branchName, tt.targetPath)

			assert.Equal(t, tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				assert.True(t, result.HasErrors(), "Should have errors when invalid")
			}
		})
	}
}

func TestValidateBranchName_EdgeCases(t *testing.T) {
	// Test UTF-8 handling
	t.Run("valid UTF-8 characters", func(t *testing.T) {
		result := ValidateBranchName("feature-🚀")
		// This should be valid as it contains valid UTF-8 but may fail regex
		// The exact behavior depends on git's rules for Unicode in branch names
		assert.False(t, result.Valid) // Based on our regex, this should be invalid
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		invalidUTF8 := "feature-" + string([]byte{0xff, 0xfe})
		result := ValidateBranchName(invalidUTF8)
		assert.False(t, result.Valid)
		assert.True(t, IsErrorType(result.FirstError(), ErrInvalidBranchName))
		assert.Contains(t, result.FirstError().Error(), "invalid UTF-8")
	})

	t.Run("single character branch", func(t *testing.T) {
		result := ValidateBranchName("a")
		assert.True(t, result.Valid)
	})

	t.Run("branch with all allowed characters", func(t *testing.T) {
		result := ValidateBranchName("abc123._-/xyz")
		assert.True(t, result.Valid)
	})
}
