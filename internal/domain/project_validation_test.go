package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ProjectValidationTestSuite provides test setup for project validation tests
type ProjectValidationTestSuite struct {
	suite.Suite
}

func TestProjectValidationSuite(t *testing.T) {
	suite.Run(t, new(ProjectValidationTestSuite))
}

func (s *ProjectValidationTestSuite) TestProjectValidationResult_AddError() {
	result := NewProjectValidationResult()
	s.True(result.Valid)

	err := NewProjectError(ErrInvalidProjectName, "test error", "")
	result.AddError(err)

	s.False(result.Valid)
	s.Len(result.Errors, 1)
	s.Equal(err, result.Errors[0])
}

func (s *ProjectValidationTestSuite) TestProjectValidationResult_AddWarning() {
	result := NewProjectValidationResult()
	result.AddWarning("test warning")

	s.True(result.Valid) // Warnings don't affect validity
	s.Len(result.Warnings, 1)
	s.Contains(result.Warnings, "test warning")
}

func (s *ProjectValidationTestSuite) TestProjectValidationResult_HasErrors() {
	result := NewProjectValidationResult()
	s.False(result.HasErrors())

	result.AddError(NewProjectError(ErrProjectValidation, "test", ""))
	s.True(result.HasErrors())
}

func (s *ProjectValidationTestSuite) TestProjectValidationResult_FirstError() {
	result := NewProjectValidationResult()
	s.Nil(result.FirstError())

	err := NewProjectError(ErrInvalidProjectName, "first error", "")
	result.AddError(err)
	s.Equal(err, result.FirstError())
}

func (s *ProjectValidationTestSuite) TestNewProjectValidationResult() {
	result := NewProjectValidationResult()

	s.True(result.Valid)
	s.Empty(result.Errors)
	s.Empty(result.Warnings)
}

func (s *ProjectValidationTestSuite) TestValidateProjectName() {
	tests := []struct {
		name           string
		projectName    string
		expectValid    bool
		expectError    bool
		errorType      DomainErrorType
		expectedErrMsg string
	}{
		{
			name:        "ValidName",
			projectName: "my-project",
			expectValid: true,
			expectError: false,
		},
		{
			name:           "EmptyName",
			projectName:    "",
			expectValid:    false,
			expectError:    true,
			errorType:      ErrInvalidProjectName,
			expectedErrMsg: "project name cannot be empty",
		},
		{
			name:           "WhitespaceOnly",
			projectName:    "   ",
			expectValid:    false,
			expectError:    true,
			errorType:      ErrInvalidProjectName,
			expectedErrMsg: "project name cannot be empty",
		},
		{
			name:        "ValidWithNumbers",
			projectName: "project-123",
			expectValid: true,
			expectError: false,
		},
		{
			name:        "ValidWithUnderscore",
			projectName: "my_project",
			expectValid: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ValidateProjectName(tt.projectName)

			s.Equal(tt.expectValid, result.Valid)

			if tt.expectError {
				s.True(result.HasErrors())
				s.Len(result.Errors, 1)

				err := result.Errors[0]
				s.Equal(tt.errorType, err.Type)
				s.Contains(err.Message, tt.expectedErrMsg)
			} else {
				s.False(result.HasErrors())
				s.Empty(result.Errors)
			}
		})
	}
}

func (s *ProjectValidationTestSuite) TestValidateGitRepoPath() {
	tests := []struct {
		name           string
		gitRepoPath    string
		expectValid    bool
		expectError    bool
		errorType      DomainErrorType
		expectedErrMsg string
	}{
		{
			name:        "ValidPath",
			gitRepoPath: "/valid/repo/path",
			expectValid: true,
			expectError: false,
		},
		{
			name:           "EmptyPath",
			gitRepoPath:    "",
			expectValid:    false,
			expectError:    true,
			errorType:      ErrInvalidGitRepoPath,
			expectedErrMsg: "git repository path cannot be empty",
		},
		{
			name:           "WhitespaceOnly",
			gitRepoPath:    "   ",
			expectValid:    false,
			expectError:    true,
			errorType:      ErrInvalidGitRepoPath,
			expectedErrMsg: "git repository path cannot be empty",
		},
		{
			name:        "ValidRelativePath",
			gitRepoPath: "relative/path",
			expectValid: true,
			expectError: false,
		},
		{
			name:        "ValidHomePath",
			gitRepoPath: "~/project",
			expectValid: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ValidateGitRepoPath(tt.gitRepoPath)

			s.Equal(tt.expectValid, result.Valid)

			if tt.expectError {
				s.True(result.HasErrors())
				s.Len(result.Errors, 1)

				err := result.Errors[0]
				s.Equal(tt.errorType, err.Type)
				s.Contains(err.Message, tt.expectedErrMsg)
			} else {
				s.False(result.HasErrors())
				s.Empty(result.Errors)
			}
		})
	}
}

func (s *ProjectValidationTestSuite) TestValidateProjectCreation() {
	tests := []struct {
		name         string
		projectName  string
		gitRepoPath  string
		expectValid  bool
		expectErrors int
	}{
		{
			name:         "ValidProject",
			projectName:  "my-project",
			gitRepoPath:  "/valid/repo",
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name:         "InvalidName",
			projectName:  "",
			gitRepoPath:  "/valid/repo",
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name:         "InvalidPath",
			projectName:  "my-project",
			gitRepoPath:  "",
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name:         "BothInvalid",
			projectName:  "",
			gitRepoPath:  "",
			expectValid:  false,
			expectErrors: 2,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ValidateProjectCreation(tt.projectName, tt.gitRepoPath)

			s.Equal(tt.expectValid, result.Valid)
			s.Len(result.Errors, tt.expectErrors)
		})
	}
}
