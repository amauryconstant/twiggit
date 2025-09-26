package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFileSystemChecker is a mock implementation of FileSystemChecker for testing
type MockFileSystemChecker struct {
	// ExistsFunc allows controlling the Exists behavior
	ExistsFunc func(path string) bool
}

// Exists implements the FileSystemChecker interface
func (m *MockFileSystemChecker) Exists(path string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(path)
	}
	return false
}

func TestContextDetector_Detect(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *ContextDetector
		currentDir    string
		expectedType  ContextType
		expectedPath  string
		expectedError bool
	}{
		{
			name: "detect worktree context",
			setup: func() *ContextDetector {
				mockFS := &MockFileSystemChecker{
					ExistsFunc: func(path string) bool {
						return path == "/home/amaury/Projects/twiggit"
					},
				}
				return NewContextDetector("/home/amaury/Workspaces", "/home/amaury/Projects", mockFS)
			},
			currentDir:    "/home/amaury/Workspaces/twiggit/shell-integration",
			expectedType:  ContextWorktree,
			expectedPath:  "/home/amaury/Projects/twiggit",
			expectedError: false,
		},
		{
			name: "detect project context",
			setup: func() *ContextDetector {
				mockFS := &MockFileSystemChecker{
					ExistsFunc: func(path string) bool {
						return true
					},
				}
				return NewContextDetector("/home/amaury/Workspaces", "/home/amaury/Projects", mockFS)
			},
			currentDir:    "/home/amaury/Projects/twiggit",
			expectedType:  ContextProject,
			expectedPath:  "/home/amaury/Projects/twiggit",
			expectedError: false,
		},
		{
			name: "detect outside git context",
			setup: func() *ContextDetector {
				mockFS := &MockFileSystemChecker{
					ExistsFunc: func(path string) bool {
						return false
					},
				}
				return NewContextDetector("/home/amaury/Workspaces", "/home/amaury/Projects", mockFS)
			},
			currentDir:    "/home/amaury/Documents",
			expectedType:  ContextOutsideGit,
			expectedPath:  "",
			expectedError: false,
		},
		{
			name: "error on empty current directory",
			setup: func() *ContextDetector {
				mockFS := &MockFileSystemChecker{
					ExistsFunc: func(path string) bool {
						return false
					},
				}
				return NewContextDetector("", "", mockFS)
			},
			currentDir:    "",
			expectedType:  ContextOutsideGit,
			expectedPath:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := tt.setup()

			context, err := detector.Detect(tt.currentDir)

			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, context.Type)
			assert.Equal(t, tt.expectedPath, context.ProjectPath)
		})
	}
}

func TestContextResolver_Resolve(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *ContextResolver
		context       *Context
		target        string
		expectedType  string
		expectedPath  string
		expectedError bool
	}{
		{
			name: "resolve main from worktree context",
			setup: func() *ContextResolver {
				return NewContextResolver("/home/amaury/Workspaces", "/home/amaury/Projects")
			},
			context: &Context{
				Type:        ContextWorktree,
				CurrentPath: "/home/amaury/Workspaces/twiggit/shell-integration",
				ProjectPath: "/home/amaury/Projects/twiggit",
				ProjectName: "twiggit",
				BranchName:  "shell-integration",
			},
			target:        "main",
			expectedType:  "project",
			expectedPath:  "/home/amaury/Projects/twiggit",
			expectedError: false,
		},
		{
			name: "resolve project from outside context",
			setup: func() *ContextResolver {
				return NewContextResolver("/home/amaury/Workspaces", "/home/amaury/Projects")
			},
			context: &Context{
				Type:        ContextOutsideGit,
				CurrentPath: "",
				ProjectPath: "",
				ProjectName: "",
				BranchName:  "",
			},
			target:        "twiggit",
			expectedType:  "project",
			expectedPath:  "/home/amaury/Projects/twiggit",
			expectedError: false,
		},
		{
			name: "resolve cross-project worktree",
			setup: func() *ContextResolver {
				return NewContextResolver("/home/amaury/Workspaces", "/home/amaury/Projects")
			},
			context: &Context{
				Type:        ContextOutsideGit,
				CurrentPath: "",
				ProjectPath: "",
				ProjectName: "",
				BranchName:  "",
			},
			target:        "twiggit/shell-integration",
			expectedType:  "worktree",
			expectedPath:  "/home/amaury/Workspaces/twiggit/shell-integration",
			expectedError: false,
		},
		{
			name: "error on empty target",
			setup: func() *ContextResolver {
				return NewContextResolver("", "")
			},
			context: &Context{
				Type:        ContextProject,
				CurrentPath: "/home/amaury/Projects/twiggit",
				ProjectPath: "/home/amaury/Projects/twiggit",
				ProjectName: "twiggit",
				BranchName:  "",
			},
			target:        "",
			expectedType:  "",
			expectedPath:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := tt.setup()

			resolution, err := resolver.Resolve(tt.target, tt.context)

			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, resolution.TargetType)
			assert.Equal(t, tt.expectedPath, resolution.TargetPath)
		})
	}
}

func TestContext_String(t *testing.T) {
	tests := []struct {
		name     string
		context  *Context
		expected string
	}{
		{
			name: "project context string",
			context: &Context{
				Type:        ContextProject,
				ProjectName: "twiggit",
				ProjectPath: "/home/amaury/Projects/twiggit",
			},
			expected: "Project(project=twiggit, path=/home/amaury/Projects/twiggit)",
		},
		{
			name: "worktree context string",
			context: &Context{
				Type:         ContextWorktree,
				ProjectName:  "twiggit",
				BranchName:   "shell-integration",
				WorktreePath: "/home/amaury/Workspaces/twiggit/shell-integration",
			},
			expected: "Worktree(project=twiggit, branch=shell-integration, path=/home/amaury/Workspaces/twiggit/shell-integration)",
		},
		{
			name:     "outside git context string",
			context:  &Context{Type: ContextOutsideGit},
			expected: "OutsideGit",
		},
		{
			name:     "unknown context string",
			context:  &Context{Type: ContextUnknown},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.context.String())
		})
	}
}

func TestContext_IsInGitContext(t *testing.T) {
	tests := []struct {
		name     string
		context  *Context
		expected bool
	}{
		{
			name:     "project context is in git",
			context:  &Context{Type: ContextProject},
			expected: true,
		},
		{
			name:     "worktree context is in git",
			context:  &Context{Type: ContextWorktree},
			expected: true,
		},
		{
			name:     "outside git context is not in git",
			context:  &Context{Type: ContextOutsideGit},
			expected: false,
		},
		{
			name:     "unknown context is not in git",
			context:  &Context{Type: ContextUnknown},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.context.IsInGitContext())
		})
	}
}

func TestContext_IsInProjectContext(t *testing.T) {
	tests := []struct {
		name     string
		context  *Context
		expected bool
	}{
		{
			name:     "project context is in project",
			context:  &Context{Type: ContextProject},
			expected: true,
		},
		{
			name:     "worktree context is not in project",
			context:  &Context{Type: ContextWorktree},
			expected: false,
		},
		{
			name:     "outside git context is not in project",
			context:  &Context{Type: ContextOutsideGit},
			expected: false,
		},
		{
			name:     "unknown context is not in project",
			context:  &Context{Type: ContextUnknown},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.context.IsInProjectContext())
		})
	}
}

func TestContext_IsInWorktreeContext(t *testing.T) {
	tests := []struct {
		name     string
		context  *Context
		expected bool
	}{
		{
			name:     "project context is not in worktree",
			context:  &Context{Type: ContextProject},
			expected: false,
		},
		{
			name:     "worktree context is in worktree",
			context:  &Context{Type: ContextWorktree},
			expected: true,
		},
		{
			name:     "outside git context is not in worktree",
			context:  &Context{Type: ContextOutsideGit},
			expected: false,
		},
		{
			name:     "unknown context is not in worktree",
			context:  &Context{Type: ContextUnknown},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.context.IsInWorktreeContext())
		})
	}
}

func TestContextDetector_WithInjectedFileSystemChecker(t *testing.T) {
	// RED: This test WILL fail until we update the constructor to accept FileSystemChecker
	mockFS := &MockFileSystemChecker{
		ExistsFunc: func(path string) bool {
			return true
		},
	}

	// This constructor call SHOULD accept FileSystemChecker parameter but currently doesn't
	detector := NewContextDetector("/workspaces", "/projects", mockFS)

	require.NotNil(t, detector, "ContextDetector should be created with injected FileSystemChecker")

	// Verify the injected checker is used
	context, err := detector.Detect("/workspaces/test-project/feature-branch")
	require.NoError(t, err)
	assert.Equal(t, ContextWorktree, context.Type)
}
