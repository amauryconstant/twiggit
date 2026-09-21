package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// hasSuggestion reports whether the supplied ValidationError carries the
// supplied suggestion text. Suggestions are now rendered via the cmd
// formatter (with a "Hint: " prefix) rather than embedded in Error(),
// so tests that previously inspected Error() for emoji-prefixed lines
// now query the Suggestions accessor directly.
func hasSuggestion(err error, needle string) bool {
	ve := &ValidationError{}
	ok := errors.As(err, &ve)
	if !ok {
		return false
	}
	for _, s := range ve.Suggestions() {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

func TestValidateBranchName_EmptyBranch(t *testing.T) {
	result := ValidateBranchName("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")
	assert.True(t, hasSuggestion(result.Error, "Provide a valid branch name"))
}

func TestValidateBranchName_InvalidCharacters(t *testing.T) {
	result := ValidateBranchName("feature@branch#name")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name format is invalid")
	assert.True(t, hasSuggestion(result.Error, "Use only alphanumeric characters"))
}

func TestValidateBranchName_ValidBranch(t *testing.T) {
	result := ValidateBranchName("feature-branch")

	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}

func TestValidateBranchName_WhitespaceOnly(t *testing.T) {
	result := ValidateBranchName("   ")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")
}

func TestValidateProjectName_EmptyProject(t *testing.T) {
	result := ValidateProjectName("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "project name is required")
	assert.True(t, hasSuggestion(result.Error, "Provide a valid project name"))
}

func TestValidateProjectName_InvalidCharacters(t *testing.T) {
	result := ValidateProjectName("project@invalid")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "project name format is invalid")
	assert.True(t, hasSuggestion(result.Error, "Use only alphanumeric characters"))
}

func TestValidateProjectName_ValidProject(t *testing.T) {
	result := ValidateProjectName("my-project")

	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}

func TestValidateProjectName_PathTraversal(t *testing.T) {
	pathTraversalCases := []string{
		"..",
		"../",
		"../etc",
		"..project",
		"project..",
		"project/../other",
		"../project",
	}

	for _, projectName := range pathTraversalCases {
		t.Run(projectName, func(t *testing.T) {
			result := ValidateProjectName(projectName)

			assert.False(t, result.IsSuccess(), "Project name %q should fail validation", projectName)
			assert.Contains(t, result.Error.Error(), "invalid")
		})
	}
}

func TestValidateShellType_EmptyShell(t *testing.T) {
	result := ValidateShellType("")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "shell type is required")
	assert.True(t, hasSuggestion(result.Error, "Provide a valid shell type"))
}

func TestValidateShellType_UnsupportedShell(t *testing.T) {
	result := ValidateShellType("powershell")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "unsupported shell type")
	assert.True(t, hasSuggestion(result.Error, "Supported shells"))
}

func TestValidateShellType_ValidShell(t *testing.T) {
	validShells := []string{"bash", "zsh", "fish"}

	for _, shell := range validShells {
		result := ValidateShellType(shell)
		assert.True(t, result.IsSuccess(), "Shell %s should be valid", shell)
		assert.True(t, result.Value)
	}
}

func TestValidateShellType_WhitespaceHandling(t *testing.T) {
	result := ValidateShellType("  bash  ")

	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "shell type format is invalid")
	assert.True(t, hasSuggestion(result.Error, "leading or trailing whitespace"))
}

func TestValidationPipeline_ComposeValidations(t *testing.T) {
	pipeline := NewValidationPipeline(
		ValidateBranchNameNotEmpty,
		ValidateBranchNameFormat,
	)

	result := pipeline.Validate("valid-branch")
	assert.True(t, result.IsSuccess())

	result = pipeline.Validate("")
	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name is required")

	result = pipeline.Validate("invalid@branch")
	assert.False(t, result.IsSuccess())
	assert.Contains(t, result.Error.Error(), "branch name format is invalid")
}

func TestValidationPipeline_EmptyPipeline(t *testing.T) {
	pipeline := NewValidationPipeline[string]()

	result := pipeline.Validate("any-input")
	assert.True(t, result.IsSuccess())
	assert.True(t, result.Value)
}

func TestValidateBranchName_InvalidTrailingChars(t *testing.T) {
	invalidCases := []string{
		"branch-",
		"branch.",
		"branch_",
		"feature-branch-",
		"develop.",
		"release_",
		"hotfix--",
		"bugfix..",
		"test__",
	}

	for _, branchName := range invalidCases {
		t.Run(branchName, func(t *testing.T) {
			result := ValidateBranchName(branchName)

			assert.False(t, result.IsSuccess())
			assert.True(t, hasSuggestion(result.Error, "cannot end with"))
		})
	}
}

func TestValidateBranchName_ValidEndingChars(t *testing.T) {
	validCases := []string{
		"branch-1",
		"branch_a",
		"branch.b",
		"feature-branch-123",
		"develop_abc",
		"release.v1.0",
		"hotfix-final",
		"bugfix_2.0",
		"test-main",
	}

	for _, branchName := range validCases {
		t.Run(branchName, func(t *testing.T) {
			result := ValidateBranchName(branchName)
			assert.True(t, result.IsSuccess())
			assert.True(t, result.Value)
		})
	}
}

func TestValidateBranchName_ReservedNames(t *testing.T) {
	reservedCases := []struct {
		name       string
		branchName string
	}{
		{"HEAD uppercase", "HEAD"},
		{"HEAD lowercase", "head"},
		{"HEAD mixed", "HeAd"},
		{"main", "main"},
		{"Main", "Main"},
		{"MASTER uppercase", "MASTER"},
		{"master lowercase", "master"},
		{"Master mixed", "Master"},
		{"ORIG_HEAD", "ORIG_HEAD"},
		{"orig_head lowercase", "orig_head"},
		{"FETCH_HEAD", "FETCH_HEAD"},
		{"fetch_head lowercase", "fetch_head"},
		{"MERGE_HEAD", "MERGE_HEAD"},
		{"merge_head lowercase", "merge_head"},
		{"MERGE_STATE", "MERGE_STATE"},
		{"merge_state lowercase", "merge_state"},
		{"CHERRY_PICK_HEAD", "CHERRY_PICK_HEAD"},
		{"cherry_pick_head lowercase", "cherry_pick_head"},
		{"REVERT_HEAD", "REVERT_HEAD"},
		{"revert_head lowercase", "revert_head"},
	}

	for _, tc := range reservedCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateBranchName(tc.branchName)
			assert.False(t, result.IsSuccess(), "Reserved branch name %q should fail validation", tc.branchName)
			assert.True(t, hasSuggestion(result.Error, "reserved branch name"))
		})
	}
}

func TestValidateBranchName_NotReserved(t *testing.T) {
	validCases := []string{
		"main-fix",
		"master-branch",
		"header-branch",
		"main_branch",
		"master_branch",
		"header_branch",
		"main.branch",
		"master.branch",
		"header.branch",
		"main-fix-1",
		"master-branch-2",
		"header-branch-3",
	}

	for _, branchName := range validCases {
		t.Run(branchName, func(t *testing.T) {
			result := ValidateBranchName(branchName)
			assert.True(t, result.IsSuccess(), "Branch name %q should pass validation", branchName)
			assert.True(t, result.Value)
		})
	}
}

func TestValidateBranchName_InvalidLeadingChars(t *testing.T) {
	invalidCases := []string{
		"-branch",
		".branch",
		"-feature-branch",
		".develop",
		"-release",
		".hotfix",
		"-test",
		".bugfix",
		"-1-branch",
		".test-branch",
	}

	for _, branchName := range invalidCases {
		t.Run(branchName, func(t *testing.T) {
			result := ValidateBranchName(branchName)

			assert.False(t, result.IsSuccess(), "Branch name %q should fail validation", branchName)
			assert.True(t, hasSuggestion(result.Error, "cannot start with"))
		})
	}
}

func TestValidateBranchName_ValidLeadingChars(t *testing.T) {
	validCases := []string{
		"branch-1",
		"branch-a",
		"feature.branch",
		"develop_1",
		"release-2.0",
		"hotfix_fix",
		"test-branch",
		"bugfix_123",
		"1-branch",
		"a-branch",
		"branch.1",
		"branch_name",
	}

	for _, branchName := range validCases {
		t.Run(branchName, func(t *testing.T) {
			result := ValidateBranchName(branchName)
			assert.True(t, result.IsSuccess(), "Branch name %q should pass validation", branchName)
			assert.True(t, result.Value)
		})
	}
}
