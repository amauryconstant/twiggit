package types

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ValidationTestSuite provides test setup for validation tests
type ValidationTestSuite struct {
	suite.Suite
}

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

func (s *ValidationTestSuite) TestValidationResult_AddError() {
	result := NewValidationResult()
	s.True(result.Valid)

	err := NewWorktreeError(ErrInvalidBranchName, "test error", "")
	result.AddError(err)

	s.False(result.Valid)
	s.Len(result.Errors, 1)
	s.Equal(err, result.Errors[0])
}

func (s *ValidationTestSuite) TestValidationResult_AddWarning() {
	result := NewValidationResult()
	result.AddWarning("test warning")

	s.True(result.Valid) // Warnings don't affect validity
	s.Len(result.Warnings, 1)
	s.Contains(result.Warnings, "test warning")
}

func (s *ValidationTestSuite) TestValidationResult_HasErrors() {
	result := NewValidationResult()
	s.False(result.HasErrors())

	result.AddError(NewWorktreeError(ErrValidation, "test", ""))
	s.True(result.HasErrors())
}

func (s *ValidationTestSuite) TestValidationResult_FirstError() {
	result := NewValidationResult()
	s.NoError(result.FirstError())

	err1 := NewWorktreeError(ErrValidation, "first error", "")
	err2 := NewWorktreeError(ErrValidation, "second error", "")

	result.AddError(err1)
	result.AddError(err2)

	s.Equal(err1, result.FirstError())
}

func (s *ValidationTestSuite) TestValidateBranchName() {
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
		s.Run(tt.name, func() {
			result := ValidateBranchName(tt.branchName)

			s.Equal(tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				s.True(result.HasErrors(), "Should have errors when invalid")
				s.True(IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					s.Contains(result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidatePath() {
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
		s.Run(tt.name, func() {
			result := ValidatePath(tt.path)

			s.Equal(tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				s.True(result.HasErrors(), "Should have errors when invalid")
				s.True(IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					s.Contains(result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidatePathWritable() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "twiggit-test-*")
	s.Require().NoError(err)
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
				s.Require().NoError(os.MkdirAll(existingPath, 0755))
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
		s.Run(tt.name, func() {
			path := tt.setupPath()
			result := ValidatePathWritable(path)

			s.Equal(tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				s.True(result.HasErrors(), "Should have errors when invalid")
				s.True(IsErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidateWorktreeCreation() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "twiggit-test-*")
	s.Require().NoError(err)
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
		s.Run(tt.name, func() {
			result := ValidateWorktreeCreation(tt.branchName, tt.targetPath)

			s.Equal(tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				s.True(result.HasErrors(), "Should have errors when invalid")
			}
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_EdgeCases() {
	// Test UTF-8 handling
	s.Run("valid UTF-8 characters", func() {
		result := ValidateBranchName("feature-🚀")
		// This should be valid as it contains valid UTF-8 but may fail regex
		// The exact behavior depends on git's rules for Unicode in branch names
		s.False(result.Valid) // Based on our regex, this should be invalid
	})

	s.Run("invalid UTF-8", func() {
		invalidUTF8 := "feature-" + string([]byte{0xff, 0xfe})
		result := ValidateBranchName(invalidUTF8)
		s.False(result.Valid)
		s.True(IsErrorType(result.FirstError(), ErrInvalidBranchName))
		s.Contains(result.FirstError().Error(), "invalid UTF-8")
	})

	s.Run("single character branch", func() {
		result := ValidateBranchName("a")
		s.True(result.Valid)
	})

	s.Run("branch with all allowed characters", func() {
		result := ValidateBranchName("abc123._-/xyz")
		s.True(result.Valid)
	})
}
