package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// WorktreeValidationTestSuite provides test setup for worktree validation tests
type WorktreeValidationTestSuite struct {
	suite.Suite
}

func TestWorktreeValidationSuite(t *testing.T) {
	suite.Run(t, new(WorktreeValidationTestSuite))
}

func (s *WorktreeValidationTestSuite) TestValidateBranchName() {
	tests := []struct {
		name        string
		branchName  string
		expectValid bool
		errorType   DomainErrorType
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
				s.True(IsDomainErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					s.Contains(result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func (s *WorktreeValidationTestSuite) TestValidatePath() {
	tests := []struct {
		name        string
		path        string
		expectValid bool
		errorType   DomainErrorType
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
				s.True(IsDomainErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
				if tt.description != "" {
					s.Contains(result.FirstError().Error(), tt.description, "Error should contain expected description")
				}
			}
		})
	}
}

func (s *WorktreeValidationTestSuite) TestValidatePathWritable() {
	// Test only pure business logic (path format validation)
	// Filesystem operations are now handled by the service layer
	tests := []struct {
		name        string
		path        string
		expectValid bool
		errorType   DomainErrorType
	}{
		{
			name:        "valid path format",
			path:        "/valid/path",
			expectValid: true,
		},
		{
			name:        "empty path",
			path:        "",
			expectValid: false,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "relative path",
			path:        "relative/path",
			expectValid: false,
			errorType:   ErrInvalidPath,
		},
		{
			name:        "path with null character",
			path:        "/valid\x00/path",
			expectValid: false,
			errorType:   ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ValidatePathWritable(tt.path)

			s.Equal(tt.expectValid, result.Valid, "Validation result should match expected")

			if !tt.expectValid {
				s.True(result.HasErrors(), "Should have errors when invalid")
				s.True(IsDomainErrorType(result.FirstError(), tt.errorType), "Should have correct error type")
			}
		})
	}
}

func (s *WorktreeValidationTestSuite) TestValidateWorktreeCreation() {
	// Test only pure business logic (format validation)
	// Filesystem operations are now handled by the service layer
	tests := []struct {
		name        string
		branchName  string
		targetPath  string
		expectValid bool
	}{
		{
			name:        "valid branch and path",
			branchName:  "feature-branch",
			targetPath:  "/valid/path",
			expectValid: true,
		},
		{
			name:        "invalid branch name",
			branchName:  "invalid branch name",
			targetPath:  "/valid/path",
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

func (s *WorktreeValidationTestSuite) TestValidateBranchName_EdgeCases() {
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
		s.True(IsDomainErrorType(result.FirstError(), ErrInvalidBranchName))
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
