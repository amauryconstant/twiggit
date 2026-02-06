//go:build test

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"twiggit/internal/domain"
)

func TestErrorFormatter_FormatValidationError_PlainTextFormatting(t *testing.T) {
	// RED: Test that will fail initially
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "", "branch name is required").
		WithSuggestions([]string{"Provide a valid branch name"})

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	// These assertions should fail initially since we haven't implemented the formatter
	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name is required")
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Provide a valid branch name")
}

func TestErrorFormatter_FormatValidationError_InvalidBranchName(t *testing.T) {
	// RED: Test invalid branch name formatting
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "BranchName", "invalid@branch", "branch name format is invalid").
		WithSuggestions([]string{"Use only alphanumeric characters, dots, hyphens, and underscores"})

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "branch name format is invalid")
	assert.Contains(t, output, "Hint:")
	assert.Contains(t, output, "Use only alphanumeric characters, dots, hyphens, and underscores")
}

func TestErrorFormatter_FormatProjectNotFoundError(t *testing.T) {
	// RED: Test project not found error formatting
	projectErr := domain.NewProjectServiceError("nonexistent-project", "", "DiscoverProject", "project not found", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(projectErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project 'nonexistent-project' not found")
}

func TestErrorFormatter_FormatWorktreeNotFoundError(t *testing.T) {
	// RED: Test worktree not found error formatting
	worktreeErr := domain.NewWorktreeServiceError("/path/to/worktree", "feature-branch", "ResolvePath", "worktree not found", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(worktreeErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "worktree 'feature-branch' not found")
}

func TestErrorFormatter_FormatGenericError(t *testing.T) {
	// RED: Test generic error formatting
	genericErr := domain.NewServiceError("ContextService", "GetCurrentContext", "failed to detect context", nil)

	formatter := NewErrorFormatter()
	output := formatter.Format(genericErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "failed to detect context")
}

func TestErrorFormatter_MultipleSuggestions(t *testing.T) {
	// RED: Test multiple suggestions formatting
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

func TestErrorFormatter_ContextInformation(t *testing.T) {
	// RED: Test context information in errors
	validationErr := domain.NewValidationError("CreateWorktreeRequest", "ProjectName", "", "project name required when not in project context").
		WithContext("Current directory: /home/user/random-dir")

	formatter := NewErrorFormatter()
	output := formatter.Format(validationErr)

	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "project name required when not in project context")
	assert.Contains(t, output, "Current directory: /home/user/random-dir")
}
