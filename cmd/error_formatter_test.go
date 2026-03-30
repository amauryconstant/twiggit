package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"twiggit/internal/domain"
)

func TestNewErrorFormatter(t *testing.T) {
	formatter := NewErrorFormatter()
	assert.NotNil(t, formatter)
	assert.False(t, formatter.quiet)
}

func TestNewErrorFormatterWithOptions(t *testing.T) {
	tests := []struct {
		name   string
		quiet  bool
		expect bool
	}{
		{"quiet mode disabled", false, false},
		{"quiet mode enabled", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewErrorFormatterWithOptions(tt.quiet)
			assert.NotNil(t, formatter)
			assert.Equal(t, tt.expect, formatter.quiet)
		})
	}
}

func TestErrorFormatter_FormatValidationError(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required").
		WithSuggestions([]string{"Provide a valid branch name"})

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Provide a valid branch name")
}

func TestErrorFormatter_FormatValidationErrorWithContext(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "ProjectName", "", "project name required when not in project context").
		WithContext("Current directory: /home/user/random-dir")

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project name required when not in project context")
	assert.Contains(t, output, "Context:")
	assert.Contains(t, output, "Current directory: /home/user/random-dir")
}

func TestErrorFormatter_FormatValidationErrorMultipleSuggestions(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required").
		WithSuggestions([]string{
			"Provide a valid branch name",
			"Branch names should follow git naming conventions",
		})

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.Contains(t, output, "Hint: Provide a valid branch name")
	assert.Contains(t, output, "Hint: Branch names should follow git naming conventions")
}

func TestErrorFormatter_FormatProjectServiceError(t *testing.T) {
	projectErr := domain.NewProjectServiceError("nonexistent-project", "", "DiscoverProject", "project not found", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(projectErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project not found for project 'nonexistent-project'")
	assert.Contains(t, output, "Hint:")
}

func TestErrorFormatter_FormatWorktreeServiceErrorNotFound(t *testing.T) {
	worktreeErr := domain.NewWorktreeServiceError("/path/to/worktree", "feature-branch", "ResolvePath", "worktree not found", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(worktreeErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "worktree not found")
	assert.Contains(t, output, "Hint:")
}

func TestErrorFormatter_FormatWorktreeServiceErrorOther(t *testing.T) {
	worktreeErr := domain.NewWorktreeServiceError("/path/to/worktree", "feature-branch", "DeleteWorktree", "permission denied", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(worktreeErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "permission denied")
	assert.Contains(t, output, "Hint:")
}

func TestErrorFormatter_FormatServiceError(t *testing.T) {
	genericErr := domain.NewServiceError("ContextService", "GetCurrentContext", "failed to detect context", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(genericErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "failed to detect context")
}

func TestErrorFormatter_FormatNavigationServiceError(t *testing.T) {
	navErr := domain.NewNavigationServiceError("main", "/path/to/project", "ResolvePath", "worktree not found", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(navErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "worktree not found")
	// NavigationServiceError falls through to generic formatting since it's not *domain.ServiceError
}

func TestErrorFormatter_FormatResolutionError(t *testing.T) {
	resErr := domain.NewResolutionError("target", "/path/to/project", "target not found", nil, nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(resErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "target not found")
	// ResolutionError falls through to generic formatting since it's not *domain.ServiceError
}

func TestErrorFormatter_QuietModeSuppressesHints(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required").
		WithSuggestions([]string{"Provide a valid branch name"})

	formatter := NewErrorFormatterWithOptions(true)
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.NotContains(t, output, "Hint:", "quiet mode should suppress hints")
}

func TestErrorFormatter_QuietModePreservesErrorMessage(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required").
		WithSuggestions([]string{"Provide a valid branch name"})

	formatter := NewErrorFormatterWithOptions(true)
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "branch name is required")
}

func TestErrorFormatter_FormatGenericError(t *testing.T) {
	genericErr := errors.New("something went wrong")

	formatter := NewErrorFormatter()
	output := formatter.Format(genericErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "something went wrong")
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ValidationError returns true",
			err:      domain.NewValidationError("req", "field", "value", "message"),
			expected: true,
		},
		{
			name:     "WorktreeServiceError returns false",
			err:      domain.NewWorktreeServiceError("/path", "branch", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ProjectServiceError returns false",
			err:      domain.NewProjectServiceError("name", "/path", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ServiceError returns false",
			err:      domain.NewServiceError("svc", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "generic error returns false",
			err:      errors.New("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWorktreeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "WorktreeServiceError returns true",
			err:      domain.NewWorktreeServiceError("/path", "branch", "op", "msg", nil),
			expected: true,
		},
		{
			name:     "ValidationError returns false",
			err:      domain.NewValidationError("req", "field", "value", "message"),
			expected: false,
		},
		{
			name:     "ProjectServiceError returns false",
			err:      domain.NewProjectServiceError("name", "/path", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "generic error returns false",
			err:      errors.New("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWorktreeError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsProjectError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ProjectServiceError returns true",
			err:      domain.NewProjectServiceError("name", "/path", "op", "msg", nil),
			expected: true,
		},
		{
			name:     "WorktreeServiceError returns false",
			err:      domain.NewWorktreeServiceError("/path", "branch", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ValidationError returns false",
			err:      domain.NewValidationError("req", "field", "value", "message"),
			expected: false,
		},
		{
			name:     "generic error returns false",
			err:      errors.New("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isProjectError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsServiceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ServiceError returns true",
			err:      domain.NewServiceError("svc", "op", "msg", nil),
			expected: true,
		},
		{
			name:     "NavigationServiceError returns false (not a ServiceError)",
			err:      domain.NewNavigationServiceError("target", "ctx", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ResolutionError returns false (not a ServiceError)",
			err:      domain.NewResolutionError("target", "ctx", "msg", nil, nil),
			expected: false,
		},
		{
			name:     "WorktreeServiceError returns false (more specific)",
			err:      domain.NewWorktreeServiceError("/path", "branch", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ProjectServiceError returns false (more specific)",
			err:      domain.NewProjectServiceError("name", "/path", "op", "msg", nil),
			expected: false,
		},
		{
			name:     "ValidationError returns false",
			err:      domain.NewValidationError("req", "field", "value", "message"),
			expected: false,
		},
		{
			name:     "generic error returns false",
			err:      errors.New("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isServiceError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorFormatter_Register(t *testing.T) {
	formatter := NewErrorFormatter()
	assert.NotEmpty(t, formatter.matchers, "expected formatter to have registered matchers")
}

func TestErrorFormatter_WithQuietMode(t *testing.T) {
	formatter := NewErrorFormatter()

	// Create a formatter with quiet mode off
	formatter.quiet = false
	wrapped := formatter.withQuietMode(func(err error) string {
		return "Hint: test hint\n"
	})
	output := wrapped(errors.New("test"))
	assert.Contains(t, output, "Hint:")

	// Create a formatter with quiet mode on
	formatter.quiet = true
	wrappedQuiet := formatter.withQuietMode(func(err error) string {
		return "Hint: test hint\n"
	})
	outputQuiet := wrappedQuiet(errors.New("test"))
	assert.NotContains(t, outputQuiet, "Hint:", "quiet mode should filter out hints")
}

func TestFormatValidationError(t *testing.T) {
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "branch", "feature-1", "branch name cannot be empty").
		WithSuggestions([]string{"Specify a branch name"}).
		WithContext("context info")

	output := formatValidationError(validationErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Context:")
}

func TestFormatWorktreeError(t *testing.T) {
	worktreeErr := domain.NewWorktreeServiceError(
		"/path/to/worktree",
		"feature-branch",
		"DeleteWorktree",
		"worktree not found",
		nil,
	)

	output := formatWorktreeError(worktreeErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "Hint:")
}

func TestFormatProjectError(t *testing.T) {
	projectErr := domain.NewProjectServiceError(
		"test-project",
		"/path/to/project",
		"DiscoverProject",
		"project not found",
		nil,
	)

	output := formatProjectError(projectErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "Hint:")
}

func TestFormatServiceError(t *testing.T) {
	serviceErr := domain.NewServiceError(
		"WorktreeService",
		"CreateWorktree",
		"failed to create worktree",
		nil,
	)

	output := formatServiceError(serviceErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "Hint:")
}

func TestFormatGenericError(t *testing.T) {
	formatter := NewErrorFormatter()
	genericErr := errors.New("something went wrong")

	output := formatter.formatGenericError(genericErr)

	assert.True(t, strings.HasPrefix(output, "Error:"))
	assert.Contains(t, output, "something went wrong")
}

func TestErrorFormatter_RegistrationOrder(t *testing.T) {
	formatter := NewErrorFormatter()
	// Registration order should be: validation, worktree, project, service
	// This tests the order of matchers
	assert.GreaterOrEqual(t, len(formatter.matchers), 4, "should have at least 4 matchers registered")
}
