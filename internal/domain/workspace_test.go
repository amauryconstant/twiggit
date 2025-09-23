package domain

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

// DomainTestSuite provides common setup for domain layer tests
type DomainTestSuite struct {
	suite.Suite
	Project   *Project
	Workspace *Workspace
}

// SetupTest initializes domain objects for each test
func (s *DomainTestSuite) SetupTest() {
	var err error
	s.Project, err = NewProject("test-project", "/repo/path")
	s.Require().NoError(err)

	s.Workspace, err = NewWorkspace("/test/workspace")
	s.Require().NoError(err)
}

// TableTestRunner provides a generic way to run table-driven tests within testify suites
type TableTestRunner[T any] struct {
	suite *suite.Suite
	tests []T
	run   func(T)
}

// NewTableTestRunner creates a new table test runner for the given test cases
func NewTableTestRunner[T any](s *suite.Suite, tests []T, run func(T)) *TableTestRunner[T] {
	return &TableTestRunner[T]{
		suite: s,
		tests: tests,
		run:   run,
	}
}

// Run executes all table-driven tests using the suite's Run method
func (r *TableTestRunner[T]) Run() {
	for _, tt := range r.tests {
		testName := r.getTestName(tt)
		r.suite.Run(testName, func() {
			r.run(tt)
		})
	}
}

// getTestName extracts a meaningful test name from the test case
func (r *TableTestRunner[T]) getTestName(testCase T) string {
	// Try to get name from struct field
	v := reflect.ValueOf(testCase)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("Name"); field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}

	// Fallback to type name and index
	return fmt.Sprintf("%T", testCase)
}

// WorkspaceTestSuite provides hybrid suite setup for workspace tests
type WorkspaceTestSuite struct {
	DomainTestSuite
}

// TestWorkspace_NewWorkspace tests workspace creation with table-driven approach
func (s *WorkspaceTestSuite) TestWorkspace_NewWorkspace() {
	testCases := []struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid workspace",
			path:        "/home/user/workspace",
			expectError: false,
		},
		{
			name:         "empty path",
			path:         "",
			expectError:  true,
			errorMessage: "workspace path cannot be empty",
		},
	}

	NewTableTestRunner(&s.Suite, testCases, func(tt struct {
		name         string
		path         string
		expectError  bool
		errorMessage string
	}) {
		workspace, err := NewWorkspace(tt.path)

		if tt.expectError {
			s.Require().Error(err)
			s.Contains(err.Error(), tt.errorMessage)
			s.Nil(workspace)
		} else {
			s.Require().NoError(err)
			s.Require().NotNil(workspace)
			s.Equal(tt.path, workspace.Path)
			s.Empty(workspace.Projects)
		}
	}).Run()
}

// TestWorkspace_AddProject tests project addition to workspace
func (s *WorkspaceTestSuite) TestWorkspace_AddProject() {
	workspace, err := NewWorkspace("/workspace")
	s.Require().NoError(err)

	project, err := NewProject("test-project", "/repo/path")
	s.Require().NoError(err)

	// Add first project
	err = workspace.AddProject(project)
	s.Require().NoError(err)
	s.Len(workspace.Projects, 1)
	s.Equal(project, workspace.Projects[0])

	// Try to add duplicate project
	duplicateProject, err := NewProject("test-project", "/different/repo")
	s.Require().NoError(err)

	err = workspace.AddProject(duplicateProject)
	s.Require().Error(err)
	s.Equal("project 'test-project' already exists in workspace", err.Error())
	s.Len(workspace.Projects, 1) // Should not be added
}

// TestWorkspace_RemoveProject tests project removal from workspace
func (s *WorkspaceTestSuite) TestWorkspace_RemoveProject() {
	workspace, err := NewWorkspace("/workspace")
	s.Require().NoError(err)

	project, err := NewProject("test-project", "/repo/path")
	s.Require().NoError(err)

	err = workspace.AddProject(project)
	s.Require().NoError(err)

	// Remove existing project
	err = workspace.RemoveProject("test-project")
	s.Require().NoError(err)
	s.Empty(workspace.Projects)

	// Try to remove non-existent project
	err = workspace.RemoveProject("nonexistent-project")
	s.Require().Error(err)
	s.Equal("project 'nonexistent-project' not found in workspace", err.Error())
}

// TestWorkspace_GetProject tests project retrieval from workspace
func (s *WorkspaceTestSuite) TestWorkspace_GetProject() {
	workspace, err := NewWorkspace("/workspace")
	s.Require().NoError(err)

	project, err := NewProject("test-project", "/repo/path")
	s.Require().NoError(err)

	err = workspace.AddProject(project)
	s.Require().NoError(err)

	// Get existing project
	found, err := workspace.GetProject("test-project")
	s.Require().NoError(err)
	s.Equal(project, found)

	// Get non-existent project
	_, err = workspace.GetProject("nonexistent-project")
	s.Require().Error(err)
	s.Equal("project 'nonexistent-project' not found in workspace", err.Error())
}

// TestWorkspace_ListAllWorktrees tests listing all worktrees in workspace
func (s *WorkspaceTestSuite) TestWorkspace_ListAllWorktrees() {
	workspace, err := NewWorkspace("/workspace")
	s.Require().NoError(err)

	// Empty workspace
	worktrees := workspace.ListAllWorktrees()
	s.Empty(worktrees)

	// Add projects with worktrees
	project1, _ := NewProject("project1", "/repo1")
	project2, _ := NewProject("project2", "/repo2")

	worktree1, _ := NewWorktree("/path1", "main")
	worktree2, _ := NewWorktree("/path2", "feature")
	worktree3, _ := NewWorktree("/path3", "develop")

	s.Require().NoError(project1.AddWorktree(worktree1))
	s.Require().NoError(project1.AddWorktree(worktree2))
	s.Require().NoError(project2.AddWorktree(worktree3))

	s.Require().NoError(workspace.AddProject(project1))
	s.Require().NoError(workspace.AddProject(project2))

	worktrees = workspace.ListAllWorktrees()
	s.Len(worktrees, 3)
	s.Contains(worktrees, worktree1)
	s.Contains(worktrees, worktree2)
	s.Contains(worktrees, worktree3)
}

// TestWorkspace_GetWorktreeByPath tests worktree retrieval by path
func (s *WorkspaceTestSuite) TestWorkspace_GetWorktreeByPath() {
	workspace, err := NewWorkspace("/workspace")
	s.Require().NoError(err)

	project, _ := NewProject("project1", "/repo1")
	worktree, _ := NewWorktree("/worktree/path", "main")
	s.Require().NoError(project.AddWorktree(worktree))
	s.Require().NoError(workspace.AddProject(project))

	// Find existing worktree
	found, err := workspace.GetWorktreeByPath("/worktree/path")
	s.Require().NoError(err)
	s.Equal(worktree, found)

	// Try to find non-existent worktree
	_, err = workspace.GetWorktreeByPath("/nonexistent/path")
	s.Require().Error(err)
	s.Contains(err.Error(), "worktree not found")
}

func (s *WorkspaceTestSuite) TestWorkspace_GetStatistics_Pure() {
	workspace := &Workspace{
		Path: "/workspace",
		Projects: []*Project{
			{
				Name:    "project1",
				GitRepo: "/repo1",
				Worktrees: []*Worktree{
					{Path: "/path1", Branch: "main"},
					{Path: "/path2", Branch: "feature"},
				},
			},
			{
				Name:    "project2",
				GitRepo: "/repo2",
				Worktrees: []*Worktree{
					{Path: "/path3", Branch: "develop"},
				},
			},
		},
	}

	stats := workspace.GetStatistics()
	s.NotNil(stats)
	s.Equal(2, stats.ProjectCount)
	s.Equal(3, stats.TotalWorktreeCount)
}

func (s *WorkspaceTestSuite) TestWorkspace_GetProject_Pure() {
	workspace := &Workspace{
		Path: "/workspace",
		Projects: []*Project{
			{Name: "project1", GitRepo: "/repo1"},
			{Name: "project2", GitRepo: "/repo2"},
		},
	}

	// Test existing project
	project, err := workspace.GetProject("project1")
	s.Require().NoError(err)
	s.Equal("project1", project.Name)

	// Test non-existent project
	_, err = workspace.GetProject("nonexistent")
	s.Require().Error(err)
	s.Equal("project 'nonexistent' not found in workspace", err.Error())
}

func (s *WorkspaceTestSuite) TestWorkspace_EnhancedFeatures() {
	s.Run("should support workspace configuration", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// This should fail initially - we need to add configuration support
		workspace.SetConfig("scan-depth", 3)
		workspace.SetConfig("exclude-patterns", []string{".git", "node_modules"})
		workspace.SetConfig("auto-discover", true)

		value, exists := workspace.GetConfig("scan-depth")
		s.True(exists)
		s.Equal(3, value)

		value, exists = workspace.GetConfig("auto-discover")
		s.True(exists)
		s.Equal(true, value)

		patterns, exists := workspace.GetConfig("exclude-patterns")
		s.True(exists)
		s.Equal([]string{".git", "node_modules"}, patterns)

		_, exists = workspace.GetConfig("non-existent")
		s.False(exists)
	})

	s.Run("should provide workspace statistics", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// Add projects with worktrees
		project1, _ := NewProject("project1", "/repo1")
		project2, _ := NewProject("project2", "/repo2")

		worktree1, _ := NewWorktree("/path1", "main")
		worktree2, _ := NewWorktree("/path2", "feature")
		worktree3, _ := NewWorktree("/path3", "develop")

		s.Require().NoError(project1.AddWorktree(worktree1))
		s.Require().NoError(project1.AddWorktree(worktree2))
		s.Require().NoError(project2.AddWorktree(worktree3))

		s.Require().NoError(workspace.AddProject(project1))
		s.Require().NoError(workspace.AddProject(project2))

		// This should fail initially - we need to add statistics
		stats := workspace.GetStatistics()
		s.NotNil(stats)
		s.Equal(2, stats.ProjectCount)
		s.Equal(3, stats.TotalWorktreeCount)
		s.Equal(3, stats.UnknownWorktreeCount)
		s.Equal(0, stats.CleanWorktreeCount)
		s.Equal(0, stats.DirtyWorktreeCount)
		s.Len(stats.AllBranches, 3)
	})

	s.Run("should support workspace configuration", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// This should fail initially - we need to add configuration support
		workspace.SetConfig("scan-depth", 3)
		workspace.SetConfig("exclude-patterns", []string{".git", "node_modules"})
		workspace.SetConfig("auto-discover", true)

		value, exists := workspace.GetConfig("scan-depth")
		s.True(exists)
		s.Equal(3, value)

		value, exists = workspace.GetConfig("auto-discover")
		s.True(exists)
		s.Equal(true, value)

		patterns, exists := workspace.GetConfig("exclude-patterns")
		s.True(exists)
		s.Equal([]string{".git", "node_modules"}, patterns)

		_, exists = workspace.GetConfig("non-existent")
		s.False(exists)
	})

	s.Run("should provide workspace health check", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// This should pass now - domain validation only checks basic rules
		health := workspace.GetHealth()
		s.NotNil(health)
		s.Equal("healthy", health.Status) // Changed: workspace with valid path is now healthy in domain layer
		s.Empty(health.Issues)            // No domain-level validation issues
		s.Equal(0, health.ProjectCount)
		s.Equal(0, health.WorktreeCount)
	})

	s.Run("should support project discovery", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// This should fail initially - we need to add project discovery
		discovered, err := workspace.DiscoverProjects()
		s.Require().NoError(err) // Should not fail for minimal implementation
		s.Empty(discovered)
	})

	s.Run("should support workspace metadata", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// This should fail initially - we need to add metadata support
		workspace.SetMetadata("created-at", "2023-01-01")
		workspace.SetMetadata("last-scanned", "2023-01-02")
		workspace.SetMetadata("version", "1.0.0")

		value, exists := workspace.GetMetadata("created-at")
		s.True(exists)
		s.Equal("2023-01-01", value)

		value, exists = workspace.GetMetadata("version")
		s.True(exists)
		s.Equal("1.0.0", value)

		_, exists = workspace.GetMetadata("non-existent")
		s.False(exists)
	})

	s.Run("should support worktree search and filtering", func() {
		workspace, err := NewWorkspace("/workspace")
		s.Require().NoError(err)

		// Add projects with worktrees
		project1, _ := NewProject("project1", "/repo1")
		project2, _ := NewProject("project2", "/repo2")

		worktree1, _ := NewWorktree("/path1", "main")
		worktree2, _ := NewWorktree("/path2", "feature-1")
		worktree3, _ := NewWorktree("/path3", "feature-2")

		s.Require().NoError(project1.AddWorktree(worktree1))
		s.Require().NoError(project1.AddWorktree(worktree2))
		s.Require().NoError(project2.AddWorktree(worktree3))

		s.Require().NoError(workspace.AddProject(project1))
		s.Require().NoError(workspace.AddProject(project2))

		// This should fail initially - we need to add search functionality
		mainWorktrees := workspace.FindWorktreesByBranch("main")
		s.Len(mainWorktrees, 1)

		featureWorktrees := workspace.FindWorktreesByBranchPattern("feature-*")
		s.Len(featureWorktrees, 2)

		project1Worktrees := workspace.FindWorktreesByProject("project1")
		s.Len(project1Worktrees, 2)

		cleanWorktrees := workspace.FindWorktreesByStatus(StatusClean)
		s.Empty(cleanWorktrees) // None are clean yet
	})
}

// TestWorkspaceSuite runs the workspace test suite
func TestWorkspaceSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceTestSuite))
}
