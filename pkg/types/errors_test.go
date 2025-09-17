package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorType_String(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errType.String())
		})
	}
}

func TestWorktreeError_Error(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestWorktreeError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &WorktreeError{
		Type:    ErrGitCommand,
		Message: "git command failed",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
}

func TestWorktreeError_WithSuggestion(t *testing.T) {
	err := &WorktreeError{
		Type:        ErrInvalidBranchName,
		Message:     "invalid branch name",
		Suggestions: []string{},
	}

	result := err.WithSuggestion("Use alphanumeric characters only")

	assert.Equal(t, err, result) // Should return the same instance
	assert.Contains(t, err.Suggestions, "Use alphanumeric characters only")
}

func TestWorktreeError_WithCode(t *testing.T) {
	err := &WorktreeError{
		Type:    ErrValidation,
		Message: "validation failed",
	}

	result := err.WithCode("VALIDATION_001")

	assert.Equal(t, err, result) // Should return the same instance
	assert.Equal(t, "VALIDATION_001", err.Code)
}

func TestNewWorktreeError(t *testing.T) {
	err := NewWorktreeError(ErrWorktreeNotFound, "worktree not found", "/tmp/path")

	assert.Equal(t, ErrWorktreeNotFound, err.Type)
	assert.Equal(t, "worktree not found", err.Message)
	assert.Equal(t, "/tmp/path", err.Path)
	assert.Empty(t, err.Suggestions)
	assert.Nil(t, err.Cause)
}

func TestWrapError(t *testing.T) {
	cause := errors.New("filesystem error")
	err := WrapError(ErrPathNotWritable, "cannot write to path", "/tmp/path", cause)

	assert.Equal(t, ErrPathNotWritable, err.Type)
	assert.Equal(t, "cannot write to path", err.Message)
	assert.Equal(t, "/tmp/path", err.Path)
	assert.Equal(t, cause, err.Cause)
	assert.Empty(t, err.Suggestions)
}

func TestIsErrorType(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorType(tt.err, tt.errType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorktreeError_Chaining(t *testing.T) {
	// Test method chaining
	err := NewWorktreeError(ErrInvalidBranchName, "invalid name", "/path").
		WithSuggestion("Use valid characters").
		WithSuggestion("Avoid special characters").
		WithCode("BRANCH_001")

	assert.Equal(t, ErrInvalidBranchName, err.Type)
	assert.Equal(t, "invalid name", err.Message)
	assert.Equal(t, "/path", err.Path)
	assert.Equal(t, "BRANCH_001", err.Code)
	assert.Len(t, err.Suggestions, 2)
	assert.Contains(t, err.Suggestions, "Use valid characters")
	assert.Contains(t, err.Suggestions, "Avoid special characters")
}
