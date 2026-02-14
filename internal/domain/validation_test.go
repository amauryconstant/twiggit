package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidationTestSuite struct {
	suite.Suite
}

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

func (s *ValidationTestSuite) TestValidateBranchName_EmptyBranch() {
	result := ValidateBranchName("")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "branch name is required")
	s.Contains(result.Error.Error(), "💡 Provide a valid branch name")
}

func (s *ValidationTestSuite) TestValidateBranchName_InvalidCharacters() {
	result := ValidateBranchName("feature@branch#name")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "branch name format is invalid")
	s.Contains(result.Error.Error(), "💡 Use only alphanumeric characters, dots, hyphens, and underscores")
}

func (s *ValidationTestSuite) TestValidateBranchName_ValidBranch() {
	result := ValidateBranchName("feature-branch")

	s.True(result.IsSuccess())
	s.True(result.Value)
}

func (s *ValidationTestSuite) TestValidateBranchName_WhitespaceOnly() {
	result := ValidateBranchName("   ")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "branch name is required")
}

func (s *ValidationTestSuite) TestValidateProjectName_EmptyProject() {
	result := ValidateProjectName("")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "project name is required")
	s.Contains(result.Error.Error(), "💡 Provide a valid project name")
}

func (s *ValidationTestSuite) TestValidateProjectName_InvalidCharacters() {
	result := ValidateProjectName("project@invalid")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "project name format is invalid")
	s.Contains(result.Error.Error(), "💡 Use only alphanumeric characters, hyphens, and underscores")
}

func (s *ValidationTestSuite) TestValidateProjectName_ValidProject() {
	result := ValidateProjectName("my-project")

	s.True(result.IsSuccess())
	s.True(result.Value)
}

func (s *ValidationTestSuite) TestValidateProjectName_PathTraversal() {
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
		s.Run(projectName, func() {
			result := ValidateProjectName(projectName)

			s.False(result.IsSuccess(), "Project name %q should fail validation", projectName)
			s.Contains(result.Error.Error(), "invalid")
		})
	}
}

func (s *ValidationTestSuite) TestValidateShellType_EmptyShell() {
	result := ValidateShellType("")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "shell type is required")
	s.Contains(result.Error.Error(), "💡 Provide a valid shell type (bash, zsh, fish)")
}

func (s *ValidationTestSuite) TestValidateShellType_UnsupportedShell() {
	result := ValidateShellType("powershell")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "unsupported shell type")
	s.Contains(result.Error.Error(), "💡 Supported shells: bash, zsh, fish")
}

func (s *ValidationTestSuite) TestValidateShellType_ValidShell() {
	validShells := []string{"bash", "zsh", "fish"}

	for _, shell := range validShells {
		result := ValidateShellType(shell)
		s.True(result.IsSuccess(), "Shell %s should be valid", shell)
		s.True(result.Value)
	}
}

func (s *ValidationTestSuite) TestValidateShellType_WhitespaceHandling() {
	result := ValidateShellType("  bash  ")

	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "shell type format is invalid")
	s.Contains(result.Error.Error(), "💡 Shell type should not contain leading or trailing whitespace")
}

func (s *ValidationTestSuite) TestValidationPipeline_ComposeValidations() {
	pipeline := NewValidationPipeline(
		ValidateBranchNameNotEmpty,
		ValidateBranchNameFormat,
	)

	// Test valid input
	result := pipeline.Validate("valid-branch")
	s.True(result.IsSuccess())

	// Test invalid input (should fail on first validation)
	result = pipeline.Validate("")
	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "branch name is required")

	// Test invalid format (should fail on second validation)
	result = pipeline.Validate("invalid@branch")
	s.False(result.IsSuccess())
	s.Contains(result.Error.Error(), "branch name format is invalid")
}

func (s *ValidationTestSuite) TestValidationPipeline_EmptyPipeline() {
	pipeline := NewValidationPipeline[string]()

	result := pipeline.Validate("any-input")
	s.True(result.IsSuccess())
	s.True(result.Value)
}

func (s *ValidationTestSuite) TestValidateBranchName_InvalidTrailingChars() {
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
		s.Run(branchName, func() {
			result := ValidateBranchName(branchName)

			s.False(result.IsSuccess())
			s.Contains(result.Error.Error(), "cannot end with")
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_ValidEndingChars() {
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
		s.Run(branchName, func() {
			result := ValidateBranchName(branchName)

			s.True(result.IsSuccess())
			s.True(result.Value)
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_ReservedNames() {
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
		s.Run(tc.name, func() {
			result := ValidateBranchName(tc.branchName)

			s.False(result.IsSuccess(), "Reserved branch name %q should fail validation", tc.branchName)
			s.Contains(result.Error.Error(), "reserved branch name")
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_NotReserved() {
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
		s.Run(branchName, func() {
			result := ValidateBranchName(branchName)

			s.True(result.IsSuccess(), "Branch name %q should pass validation", branchName)
			s.True(result.Value)
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_InvalidLeadingChars() {
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
		s.Run(branchName, func() {
			result := ValidateBranchName(branchName)

			s.False(result.IsSuccess(), "Branch name %q should fail validation", branchName)
			s.Contains(result.Error.Error(), "cannot start with")
		})
	}
}

func (s *ValidationTestSuite) TestValidateBranchName_ValidLeadingChars() {
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
		s.Run(branchName, func() {
			result := ValidateBranchName(branchName)

			s.True(result.IsSuccess(), "Branch name %q should pass validation", branchName)
			s.True(result.Value)
		})
	}
}
