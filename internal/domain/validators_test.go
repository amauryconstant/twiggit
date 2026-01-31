package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBranchName_EmptyBranch(t *testing.T) {
	// RED: Test that will fail initially
	result := ValidateBranchName("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")
	assert.Contains(t, result.Error.Error(), "💡 Provide a valid branch name")
}

func TestValidateBranchName_InvalidCharacters(t *testing.T) {
	// RED: Test invalid branch name characters
	result := ValidateBranchName("feature@branch#name")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name format is invalid")
	assert.Contains(t, result.Error.Error(), "💡 Use only alphanumeric characters, dots, hyphens, and underscores")
}

func TestValidateBranchName_ValidBranch(t *testing.T) {
	// RED: Test valid branch name (this should pass once implemented)
	result := ValidateBranchName("feature-branch")

	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}

func TestValidateBranchName_WhitespaceOnly(t *testing.T) {
	// RED: Test whitespace-only branch name
	result := ValidateBranchName("   ")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")
}

func TestValidateProjectName_EmptyProject(t *testing.T) {
	// RED: Test empty project name
	result := ValidateProjectName("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "project name is required")
	assert.Contains(t, result.Error.Error(), "💡 Provide a valid project name")
}

func TestValidateProjectName_InvalidCharacters(t *testing.T) {
	// RED: Test invalid project name characters
	result := ValidateProjectName("project@invalid")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "project name format is invalid")
	assert.Contains(t, result.Error.Error(), "💡 Use only alphanumeric characters, hyphens, and underscores")
}

func TestValidateProjectName_ValidProject(t *testing.T) {
	// RED: Test valid project name
	result := ValidateProjectName("my-project")

	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}

func TestValidateShellType_EmptyShell(t *testing.T) {
	// RED: Test empty shell type
	result := ValidateShellType("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "shell type is required")
	assert.Contains(t, result.Error.Error(), "💡 Provide a valid shell type (bash, zsh, fish)")
}

func TestValidateShellType_UnsupportedShell(t *testing.T) {
	// RED: Test unsupported shell type
	result := ValidateShellType("powershell")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "unsupported shell type")
	assert.Contains(t, result.Error.Error(), "💡 Supported shells: bash, zsh, fish")
}

func TestValidateShellType_ValidShell(t *testing.T) {
	// RED: Test valid shell types
	validShells := []string{"bash", "zsh", "fish"}

	for _, shell := range validShells {
		result := ValidateShellType(shell)
		assert.True(t, result.IsSuccess(), "Shell %s should be valid", shell)
		assert.True(t, result.Value)
	}
}

func TestValidateShellType_WhitespaceHandling(t *testing.T) {
	// RED: Test shell type with extra whitespace
	result := ValidateShellType("  bash  ")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "shell type format is invalid")
	assert.Contains(t, result.Error.Error(), "💡 Shell type should not contain leading or trailing whitespace")
}

func TestValidationPipeline_ComposeValidations(t *testing.T) {
	// RED: Test validation pipeline composition
	pipeline := NewValidationPipeline(
		ValidateBranchNameNotEmpty,
		ValidateBranchNameFormat,
	)

	// Test valid input
	result := pipeline.Validate("valid-branch")
	assert.True(t, result.IsSuccess())

	// Test invalid input (should fail on first validation)
	result = pipeline.Validate("")
	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")

	// Test invalid format (should fail on second validation)
	result = pipeline.Validate("invalid@branch")
	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name format is invalid")
}

func TestValidationPipeline_EmptyPipeline(t *testing.T) {
	// RED: Test empty validation pipeline
	pipeline := NewValidationPipeline[string]()

	result := pipeline.Validate("any-input")
	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}
