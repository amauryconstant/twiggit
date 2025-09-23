package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ProjectErrorsTestSuite provides test setup for project error type tests
type ProjectErrorsTestSuite struct {
	suite.Suite
}

func TestProjectErrorsSuite(t *testing.T) {
	suite.Run(t, new(ProjectErrorsTestSuite))
}

func (s *ProjectErrorsTestSuite) TestProjectErrorType_String() {
	tests := []struct {
		name     string
		errType  ProjectErrorType
		expected string
	}{
		{"InvalidProjectName", ErrInvalidProjectName, "invalid project name"},
		{"InvalidGitRepoPath", ErrInvalidGitRepoPath, "invalid git repository path"},
		{"ProjectNotFound", ErrProjectNotFound, "project not found"},
		{"ProjectAlreadyExists", ErrProjectAlreadyExists, "project already exists"},
		{"ProjectValidation", ErrProjectValidation, "project validation error"},
		{"Unknown", ErrUnknownProjectError, "unknown project error"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.errType.String())
		})
	}
}

func (s *ProjectErrorsTestSuite) TestProjectError_Error() {
	tests := []struct {
		name     string
		err      *ProjectError
		expected string
	}{
		{
			name: "WithPath",
			err: NewProjectError(
				ErrInvalidGitRepoPath,
				"git repository path is invalid",
				"/invalid/path",
			),
			expected: "invalid git repository path: git repository path is invalid (path: /invalid/path)",
		},
		{
			name: "WithoutPath",
			err: NewProjectError(
				ErrInvalidProjectName,
				"project name cannot be empty",
				"",
			),
			expected: "invalid project name: project name cannot be empty",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
		})
	}
}

func (s *ProjectErrorsTestSuite) TestProjectError_Unwrap() {
	cause := errors.New("root cause")
	err := WrapProjectError(
		ErrInvalidGitRepoPath,
		"git repository validation failed",
		"/repo/path",
		cause,
	)

	s.Equal(cause, err.Unwrap())
}

func (s *ProjectErrorsTestSuite) TestProjectError_WithSuggestion() {
	err := NewProjectError(
		ErrInvalidProjectName,
		"project name is invalid",
		"",
	)

	err = err.WithSuggestion("Use alphanumeric characters only")
	err = err.WithSuggestion("Avoid special characters")

	s.Len(err.Suggestions, 2)
	s.Contains(err.Suggestions, "Use alphanumeric characters only")
	s.Contains(err.Suggestions, "Avoid special characters")
}

func (s *ProjectErrorsTestSuite) TestProjectError_WithCode() {
	err := NewProjectError(
		ErrInvalidProjectName,
		"project name is invalid",
		"",
	)

	err = err.WithCode("PROJ_001")

	s.Equal("PROJ_001", err.Code)
}

func (s *ProjectErrorsTestSuite) TestNewProjectError() {
	err := NewProjectError(
		ErrInvalidProjectName,
		"test message",
		"/test/path",
	)

	s.Equal(ErrInvalidProjectName, err.Type)
	s.Equal("test message", err.Message)
	s.Equal("/test/path", err.Path)
	s.Empty(err.Suggestions)
	s.Empty(err.Code)
	s.NoError(err.Cause)
}

func (s *ProjectErrorsTestSuite) TestWrapProjectError() {
	cause := errors.New("root cause")
	err := WrapProjectError(
		ErrInvalidGitRepoPath,
		"validation failed",
		"/repo/path",
		cause,
	)

	s.Equal(ErrInvalidGitRepoPath, err.Type)
	s.Equal("validation failed", err.Message)
	s.Equal("/repo/path", err.Path)
	s.Equal(cause, err.Cause)
	s.Empty(err.Suggestions)
	s.Empty(err.Code)
}

func (s *ProjectErrorsTestSuite) TestIsProjectErrorType() {
	// Test with ProjectError
	projectErr := NewProjectError(ErrProjectNotFound, "not found", "")
	s.True(IsProjectErrorType(projectErr, ErrProjectNotFound))
	s.False(IsProjectErrorType(projectErr, ErrInvalidProjectName))

	// Test with standard error
	standardErr := errors.New("standard error")
	s.False(IsProjectErrorType(standardErr, ErrProjectNotFound))

	// Test with nil
	s.False(IsProjectErrorType(nil, ErrProjectNotFound))
}
