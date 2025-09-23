package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateWorkspacePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected WorkspaceValidationResult
	}{
		{
			name: "valid absolute path should pass",
			path: "/home/user/workspace",
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
		{
			name: "valid relative path should pass",
			path: "./workspace",
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
		{
			name: "empty path should fail",
			path: "",
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "workspace path cannot be empty"},
				},
			},
		},
		{
			name: "whitespace-only path should fail",
			path: "   ",
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidPath, Message: "workspace path cannot be empty"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspacePath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateWorkspaceProjectName(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		workspace   *Workspace
		expected    WorkspaceValidationResult
	}{
		{
			name:        "valid new project name should pass",
			projectName: "new-project",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
		{
			name:        "empty project name should fail",
			projectName: "",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidConfiguration, Message: "project name cannot be empty"},
				},
			},
		},
		{
			name:        "whitespace-only project name should fail",
			projectName: "   ",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidConfiguration, Message: "project name cannot be empty"},
				},
			},
		},
		{
			name:        "duplicate project name should fail",
			projectName: "existing-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorProjectAlreadyExists, Message: "project 'existing-project' already exists in workspace"},
				},
			},
		},
		{
			name:        "case-sensitive duplicate project name should pass",
			projectName: "Existing-Project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
		{
			name:        "unique project name should pass",
			projectName: "unique-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
					{Name: "another-project", GitRepo: "/another/repo"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceProjectName(tc.projectName, tc.workspace)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateWorkspaceProjectExists(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		workspace   *Workspace
		expected    WorkspaceValidationResult
	}{
		{
			name:        "existing project should pass",
			projectName: "existing-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{},
			},
		},
		{
			name:        "non-existent project should fail",
			projectName: "non-existent-project",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorProjectNotFound, Message: "project 'non-existent-project' not found in workspace"},
				},
			},
		},
		{
			name:        "empty project name should fail",
			projectName: "",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidConfiguration, Message: "project name cannot be empty"},
				},
			},
		},
		{
			name:        "whitespace-only project name should fail",
			projectName: "   ",
			workspace: &Workspace{
				Projects: []*Project{
					{Name: "existing-project", GitRepo: "/repo/path"},
				},
			},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidConfiguration, Message: "project name cannot be empty"},
				},
			},
		},
		{
			name:        "nil workspace should fail",
			projectName: "test-project",
			workspace:   nil,
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorInvalidConfiguration, Message: "workspace cannot be nil"},
				},
			},
		},
		{
			name:        "empty workspace should fail",
			projectName: "test-project",
			workspace:   &Workspace{Projects: []*Project{}},
			expected: WorkspaceValidationResult{
				Errors: []WorkspaceError{
					{Type: WorkspaceErrorProjectNotFound, Message: "project 'test-project' not found in workspace"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceProjectExists(tc.projectName, tc.workspace)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateWorkspaceCreation(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected WorkspaceValidationResult
	}{
		{
			name:     "valid path should pass",
			path:     "/home/user/workspace",
			expected: NewWorkspaceValidationResult(),
		},
		{
			name: "empty path should fail",
			path: "",
			expected: NewWorkspaceValidationResult(
				WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "workspace path cannot be empty"},
			),
		},
		{
			name: "whitespace-only path should fail",
			path: "   ",
			expected: NewWorkspaceValidationResult(
				WorkspaceError{Type: WorkspaceErrorInvalidPath, Message: "workspace path cannot be empty"},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceCreation(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateWorkspaceHealth(t *testing.T) {
	// Create a mock path validator for testing
	mockPathValidator := &MockPathValidator{
		IsValidWorkspacePathFunc: func(path string) bool {
			return path == "/valid/path"
		},
	}

	testCases := []struct {
		name      string
		workspace *Workspace
		validator PathValidator
		expected  WorkspaceValidationResult
	}{
		{
			name:      "valid workspace should pass",
			workspace: &Workspace{Path: "/valid/path"},
			validator: mockPathValidator,
			expected:  NewWorkspaceValidationResult(),
		},
		{
			name:      "invalid path should fail",
			workspace: &Workspace{Path: "/invalid/path"},
			validator: mockPathValidator,
			expected: NewWorkspaceValidationResult(
				WorkspaceError{Type: WorkspaceErrorValidationFailed, Message: "workspace path '/invalid/path' is not valid"},
			),
		},
		{
			name:      "nil workspace should fail",
			workspace: nil,
			validator: mockPathValidator,
			expected: NewWorkspaceValidationResult(
				WorkspaceError{Type: WorkspaceErrorInvalidConfiguration, Message: "workspace cannot be nil"},
			),
		},
		{
			name:      "nil validator should fail",
			workspace: &Workspace{Path: "/valid/path"},
			validator: nil,
			expected: NewWorkspaceValidationResult(
				WorkspaceError{Type: WorkspaceErrorInvalidConfiguration, Message: "path validator cannot be nil"},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateWorkspaceHealth(tc.workspace, tc.validator)
			assert.Equal(t, tc.expected, result)
		})
	}
}
