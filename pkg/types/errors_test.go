package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ErrorsTestSuite provides test setup for error type tests
type ErrorsTestSuite struct {
	suite.Suite
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}

func (s *ErrorsTestSuite) TestErrorType_String() {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"NotRepository", ErrNotRepository, "not a git repository"},
		{"CurrentDirectory", ErrCurrentDirectory, "current directory operation error"},
		{"UncommittedChanges", ErrUncommittedChanges, "uncommitted changes detected"},
		{"WorktreeExists", ErrWorktreeExists, "worktree already exists"},
		{"WorktreeNotFound", ErrWorktreeNotFound, "worktree not found"},
		{"InvalidBranchName", ErrInvalidBranchName, "invalid branch name"},
		{"InvalidPath", ErrInvalidPath, "invalid path"},
		{"PathNotWritable", ErrPathNotWritable, "path not writable"},
		{"GitCommand", ErrGitCommand, "git command failed"},
		{"Validation", ErrValidation, "validation error"},
		{"Unknown", ErrUnknown, "unknown error"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.errType.String())
		})
	}
}

func (s *ErrorsTestSuite) TestWorktreeError_Error() {
	tests := []struct {
		name     string
		err      *WorktreeError
		expected string
	}{
		{
			name: "error with path",
			err: &WorktreeError{
				Type:    ErrWorktreeExists,
				Message: "worktree already exists at location",
				Path:    "/tmp/test-path",
			},
			expected: "worktree already exists: worktree already exists at location (path: /tmp/test-path)",
		},
		{
			name: "error without path",
			err: &WorktreeError{
				Type:    ErrInvalidBranchName,
				Message: "branch name is invalid",
				Path:    "",
			},
			expected: "invalid branch name: branch name is invalid",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.err.Error())
		})
	}
}

func (s *ErrorsTestSuite) TestWorktreeError_Unwrap() {
	cause := errors.New("underlying error")
	err := &WorktreeError{
		Type:    ErrGitCommand,
		Message: "git command failed",
		Cause:   cause,
	}

	s.Equal(cause, err.Unwrap())
}

func (s *ErrorsTestSuite) TestWorktreeError_WithSuggestion() {
	err := &WorktreeError{
		Type:        ErrInvalidBranchName,
		Message:     "invalid branch name",
		Suggestions: []string{},
	}

	result := err.WithSuggestion("Use alphanumeric characters only")

	s.Equal(err, result) // Should return the same instance
	s.Contains(err.Suggestions, "Use alphanumeric characters only")
}

func (s *ErrorsTestSuite) TestWorktreeError_WithCode() {
	err := &WorktreeError{
		Type:    ErrValidation,
		Message: "validation failed",
	}

	result := err.WithCode("VALIDATION_001")

	s.Equal(err, result) // Should return the same instance
	s.Equal("VALIDATION_001", err.Code)
}

func (s *ErrorsTestSuite) TestNewWorktreeError() {
	err := NewWorktreeError(ErrWorktreeNotFound, "worktree not found", "/tmp/path")

	s.Equal(ErrWorktreeNotFound, err.Type)
	s.Equal("worktree not found", err.Message)
	s.Equal("/tmp/path", err.Path)
	s.Empty(err.Suggestions)
	s.NoError(err.Cause)
}

func (s *ErrorsTestSuite) TestWrapError() {
	cause := errors.New("filesystem error")
	err := WrapError(ErrPathNotWritable, "cannot write to path", "/tmp/path", cause)

	s.Equal(ErrPathNotWritable, err.Type)
	s.Equal("cannot write to path", err.Message)
	s.Equal("/tmp/path", err.Path)
	s.Equal(cause, err.Cause)
	s.Empty(err.Suggestions)
}

func (s *ErrorsTestSuite) TestIsErrorType() {
	worktreeErr := &WorktreeError{
		Type:    ErrWorktreeExists,
		Message: "test error",
	}

	otherErr := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		errType  ErrorType
		expected bool
	}{
		{"matching worktree error", worktreeErr, ErrWorktreeExists, true},
		{"non-matching worktree error", worktreeErr, ErrWorktreeNotFound, false},
		{"regular error", otherErr, ErrWorktreeExists, false},
		{"nil error", nil, ErrWorktreeExists, false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := IsErrorType(tt.err, tt.errType)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ErrorsTestSuite) TestWorktreeError_Chaining() {
	// Test method chaining
	err := NewWorktreeError(ErrInvalidBranchName, "invalid name", "/path").
		WithSuggestion("Use valid characters").
		WithSuggestion("Avoid special characters").
		WithCode("BRANCH_001")

	s.Equal(ErrInvalidBranchName, err.Type)
	s.Equal("invalid name", err.Message)
	s.Equal("/path", err.Path)
	s.Equal("BRANCH_001", err.Code)
	s.Len(err.Suggestions, 2)
	s.Contains(err.Suggestions, "Use valid characters")
	s.Contains(err.Suggestions, "Avoid special characters")
}
