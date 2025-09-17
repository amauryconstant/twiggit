// Package fixtures contains test case data for twiggit
package fixtures

import (
	"errors"
	"reflect"
	"strings"
)

// Define error types locally to avoid import cycles
var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrInvalidProjectName   = errors.New("invalid project name")
	ErrInvalidPath          = errors.New("invalid path")
	ErrWorktreeCorrupted    = errors.New("worktree corrupted")
	ErrInvalidWorktreeName  = errors.New("invalid worktree name")
	ErrBranchNotFound       = errors.New("branch not found")
	ErrInvalidBranchName    = errors.New("invalid branch name")
	ErrInvalidTestCase      = errors.New("invalid test case")
	ErrInvalidTestCaseName  = errors.New("invalid test case name")
)

// ProjectWorktreeTestCase represents a test case for project and worktree operations
type ProjectWorktreeTestCase struct {
	Name           string
	ProjectPath    string
	WorktreePath   string
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() error
	CleanupFunc    func() error
}

// ProjectTestCase represents a test case for project-specific operations
type ProjectTestCase struct {
	Name           string
	ProjectName    string
	BasePath       string
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (any, error)
	CleanupFunc    func() error
}

// WorkspaceTestCase represents a test case for workspace operations
type WorkspaceTestCase struct {
	Name           string
	WorkspacePath  string
	Projects       []string
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (any, error)
	CleanupFunc    func() error
}

// WorktreeDomainTestCase represents a test case for worktree domain operations
type WorktreeDomainTestCase struct {
	Name           string
	WorktreeName   string
	ProjectPath    string
	BranchName     string
	ExpectedError  error
	ExpectedResult any
	SetupFunc      func() (any, error)
	CleanupFunc    func() error
}

// GetProjectWorktreeTestCases returns a comprehensive set of test cases for project and worktree operations
func GetProjectWorktreeTestCases() []ProjectWorktreeTestCase {
	return []ProjectWorktreeTestCase{
		{
			Name:          "Valid project with single worktree",
			ProjectPath:   "/tmp/test-project",
			WorktreePath:  "/tmp/test-project/worktrees/feature-1",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"worktreeCount": 1,
				"isValid":       true,
			},
		},
		{
			Name:          "Valid project with multiple worktrees",
			ProjectPath:   "/tmp/test-project-multi",
			WorktreePath:  "/tmp/test-project-multi/worktrees/feature-1",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"worktreeCount": 3,
				"isValid":       true,
			},
		},
		{
			Name:           "Invalid project path",
			ProjectPath:    "/nonexistent/path",
			WorktreePath:   "/nonexistent/path/worktrees/feature-1",
			ExpectedError:  ErrProjectNotFound,
			ExpectedResult: nil,
		},
		{
			Name:          "Project without worktrees",
			ProjectPath:   "/tmp/test-project-no-worktrees",
			WorktreePath:  "",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"worktreeCount": 0,
				"isValid":       true,
			},
		},
		{
			Name:           "Project with corrupted worktree",
			ProjectPath:    "/tmp/test-project-corrupted",
			WorktreePath:   "/tmp/test-project-corrupted/worktrees/corrupted",
			ExpectedError:  ErrWorktreeCorrupted,
			ExpectedResult: nil,
		},
	}
}

// GetProjectTestCases returns test cases for project operations
func GetProjectTestCases() []ProjectTestCase {
	return []ProjectTestCase{
		{
			Name:          "Create new project",
			ProjectName:   "test-project",
			BasePath:      "/tmp",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"name":  "test-project",
				"path":  "/tmp/test-project",
				"valid": true,
			},
		},
		{
			Name:           "Create project with existing name",
			ProjectName:    "existing-project",
			BasePath:       "/tmp",
			ExpectedError:  ErrProjectAlreadyExists,
			ExpectedResult: nil,
		},
		{
			Name:           "Create project with invalid name",
			ProjectName:    "invalid/name",
			BasePath:       "/tmp",
			ExpectedError:  ErrInvalidProjectName,
			ExpectedResult: nil,
		},
		{
			Name:           "Create project in non-existent directory",
			ProjectName:    "test-project",
			BasePath:       "/nonexistent/path",
			ExpectedError:  ErrInvalidPath,
			ExpectedResult: nil,
		},
	}
}

// GetWorkspaceTestCases returns test cases for workspace operations
func GetWorkspaceTestCases() []WorkspaceTestCase {
	return []WorkspaceTestCase{
		{
			Name:          "Create workspace with single project",
			WorkspacePath: "/tmp/test-workspace",
			Projects:      []string{"project1"},
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"projectCount": 1,
				"isValid":      true,
			},
		},
		{
			Name:          "Create workspace with multiple projects",
			WorkspacePath: "/tmp/test-workspace-multi",
			Projects:      []string{"project1", "project2", "project3"},
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"projectCount": 3,
				"isValid":      true,
			},
		},
		{
			Name:          "Create workspace with no projects",
			WorkspacePath: "/tmp/test-workspace-empty",
			Projects:      []string{},
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"projectCount": 0,
				"isValid":      true,
			},
		},
		{
			Name:           "Create workspace with invalid path",
			WorkspacePath:  "/invalid/path/workspace",
			Projects:       []string{"project1"},
			ExpectedError:  ErrInvalidPath,
			ExpectedResult: nil,
		},
	}
}

// GetWorktreeDomainTestCases returns test cases for worktree domain operations
func GetWorktreeDomainTestCases() []WorktreeDomainTestCase {
	return []WorktreeDomainTestCase{
		{
			Name:          "Create worktree from main branch",
			WorktreeName:  "feature-1",
			ProjectPath:   "/tmp/test-project",
			BranchName:    "main",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"name":   "feature-1",
				"branch": "main",
				"valid":  true,
			},
		},
		{
			Name:          "Create worktree from feature branch",
			WorktreeName:  "feature-2",
			ProjectPath:   "/tmp/test-project",
			BranchName:    "develop",
			ExpectedError: nil,
			ExpectedResult: map[string]interface{}{
				"name":   "feature-2",
				"branch": "develop",
				"valid":  true,
			},
		},
		{
			Name:           "Create worktree with invalid name",
			WorktreeName:   "invalid/name",
			ProjectPath:    "/tmp/test-project",
			BranchName:     "main",
			ExpectedError:  ErrInvalidWorktreeName,
			ExpectedResult: nil,
		},
		{
			Name:           "Create worktree from non-existent branch",
			WorktreeName:   "feature-3",
			ProjectPath:    "/tmp/test-project",
			BranchName:     "non-existent-branch",
			ExpectedError:  ErrBranchNotFound,
			ExpectedResult: nil,
		},
		{
			Name:           "Create worktree in non-existent project",
			WorktreeName:   "feature-4",
			ProjectPath:    "/nonexistent/project",
			BranchName:     "main",
			ExpectedError:  ErrProjectNotFound,
			ExpectedResult: nil,
		},
	}
}

// GetValidationTestCases returns test cases for validation operations
func GetValidationTestCases() []struct {
	Name        string
	Input       any
	Expected    bool
	ErrorType   error
	Description string
} {
	return []struct {
		Name        string
		Input       any
		Expected    bool
		ErrorType   error
		Description string
	}{
		{
			Name:        "Valid project name",
			Input:       "valid-project",
			Expected:    true,
			ErrorType:   nil,
			Description: "Should accept valid project names",
		},
		{
			Name:        "Invalid project name with slash",
			Input:       "invalid/project",
			Expected:    false,
			ErrorType:   ErrInvalidProjectName,
			Description: "Should reject project names with slashes",
		},
		{
			Name:        "Invalid project name with spaces",
			Input:       "invalid project",
			Expected:    false,
			ErrorType:   ErrInvalidProjectName,
			Description: "Should reject project names with spaces",
		},
		{
			Name:        "Empty project name",
			Input:       "",
			Expected:    false,
			ErrorType:   ErrInvalidProjectName,
			Description: "Should reject empty project names",
		},
		{
			Name:        "Valid worktree name",
			Input:       "valid-worktree",
			Expected:    true,
			ErrorType:   nil,
			Description: "Should accept valid worktree names",
		},
		{
			Name:        "Invalid worktree name with slash",
			Input:       "invalid/worktree",
			Expected:    false,
			ErrorType:   ErrInvalidWorktreeName,
			Description: "Should reject worktree names with slashes",
		},
		{
			Name:        "Valid branch name",
			Input:       "valid-branch",
			Expected:    true,
			ErrorType:   nil,
			Description: "Should accept valid branch names",
		},
		{
			Name:        "Invalid branch name with spaces",
			Input:       "invalid branch",
			Expected:    false,
			ErrorType:   ErrInvalidBranchName,
			Description: "Should reject branch names with spaces",
		},
	}
}

// Helper functions for test case manipulation

// FilterTestCases filters test cases by name pattern
func FilterTestCases[T any](cases []T, namePattern string) []T {
	var result []T
	for _, tc := range cases {
		// Use reflection to get the Name field
		v := reflectValue(tc)
		if nameField := v.FieldByName("Name"); nameField.IsValid() {
			if strings.Contains(nameField.String(), namePattern) {
				result = append(result, tc)
			}
		}
	}
	return result
}

// GetTestCasesByErrorType filters test cases by expected error type
func GetTestCasesByErrorType[T any](cases []T, errorType error) []T {
	var result []T
	for _, tc := range cases {
		v := reflectValue(tc)
		if errorField := v.FieldByName("ExpectedError"); errorField.IsValid() {
			if err, ok := errorField.Interface().(error); ok && errors.Is(err, errorType) {
				result = append(result, tc)
			}
		}
	}
	return result
}

// GetTestCasesByResult filters test cases by expected result type
func GetTestCasesByResult[T any](cases []T, resultType any) []T {
	var result []T
	for _, tc := range cases {
		v := reflectValue(tc)
		if resultField := v.FieldByName("ExpectedResult"); resultField.IsValid() {
			if resultField.Interface() == resultType {
				result = append(result, tc)
			}
		}
	}
	return result
}

// reflectValue is a helper to get reflect.Value from any type
func reflectValue(v any) reflect.Value {
	if val, ok := v.(reflect.Value); ok {
		return val
	}
	return reflect.ValueOf(v)
}

// GetTestNames returns all test case names
func GetTestNames[T any](cases []T) []string {
	var names []string
	for _, tc := range cases {
		v := reflectValue(tc)
		if nameField := v.FieldByName("Name"); nameField.IsValid() {
			names = append(names, nameField.String())
		}
	}
	return names
}

// ValidateTestCase validates that a test case has required fields
func ValidateTestCase(tc any) error {
	v := reflect.ValueOf(tc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ErrInvalidTestCase
	}

	// Check for required Name field
	nameField := v.FieldByName("Name")
	if !nameField.IsValid() || nameField.Kind() != reflect.String || nameField.String() == "" {
		return ErrInvalidTestCaseName
	}

	return nil
}

// SanitizeTestName sanitizes test names for use in test runners
func SanitizeTestName(name string) string {
	// Replace spaces and special characters with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")

	// Remove consecutive underscores
	for strings.Contains(sanitized, "__") {
		sanitized = strings.ReplaceAll(sanitized, "__", "_")
	}

	// Trim leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")

	return sanitized
}

// GetTestCaseDescription returns a human-readable description of a test case
func GetTestCaseDescription(tc any) string {
	v := reflectValue(tc)

	name := "Unknown"
	if nameField := v.FieldByName("Name"); nameField.IsValid() {
		name = nameField.String()
	}

	description := name

	// Add error expectation if present
	if errorField := v.FieldByName("ExpectedError"); errorField.IsValid() {
		if err, ok := errorField.Interface().(error); ok && err != nil {
			description += " (should fail with " + err.Error() + ")"
		}
	}

	// Add result expectation if present
	if resultField := v.FieldByName("ExpectedResult"); resultField.IsValid() && !resultField.IsNil() {
		description += " (should succeed)"
	}

	return description
}
