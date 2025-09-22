package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ProjectTestSuite provides hybrid suite setup for project tests
type ProjectTestSuite struct {
	suite.Suite
	Project   *Project
	Workspace *Workspace
}

// SetupTest initializes domain objects for each test
func (s *ProjectTestSuite) SetupTest() {
	var err error
	s.Project, err = NewProject("test-project", "/repo/path")
	s.Require().NoError(err)

	s.Workspace, err = NewWorkspace("/test/workspace")
	s.Require().NoError(err)
}

// TestProject_NewProject tests project creation with table-driven approach
func (s *ProjectTestSuite) TestProject_NewProject() {
	testCases := []struct {
		name         string
		projectName  string
		gitRepo      string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project",
			projectName: "my-project",
			gitRepo:     "/path/to/repo",
			expectError: false,
		},
		{
			name:         "empty project name",
			projectName:  "",
			gitRepo:      "/path/to/repo",
			expectError:  true,
			errorMessage: "project name cannot be empty",
		},
		{
			name:         "empty git repo path",
			projectName:  "my-project",
			gitRepo:      "",
			expectError:  true,
			errorMessage: "git repository path cannot be empty",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			project, err := NewProject(tt.projectName, tt.gitRepo)

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
				s.Nil(project)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(project)
				s.Equal(tt.projectName, project.Name)
				s.Equal(tt.gitRepo, project.GitRepo)
				s.Empty(project.Worktrees)
			}
		})
	}
}

// TestProject_AddWorktree tests adding worktrees with table-driven approach
func (s *ProjectTestSuite) TestProject_AddWorktree() {
	testCases := []struct {
		name           string
		setupWorktrees func() []*Worktree
		addWorktree    *Worktree
		expectError    bool
		errorMessage   string
		finalCount     int
	}{
		{
			name: "should add first worktree",
			setupWorktrees: func() []*Worktree {
				return []*Worktree{}
			},
			addWorktree: func() *Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return worktree
			}(),
			expectError: false,
			finalCount:  1,
		},
		{
			name: "should add second worktree",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			addWorktree: func() *Worktree {
				worktree, err := NewWorktree("/another/path", "main")
				s.Require().NoError(err)
				return worktree
			}(),
			expectError: false,
			finalCount:  2,
		},
		{
			name: "should reject duplicate worktree path",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			addWorktree: func() *Worktree {
				worktree, err := NewWorktree("/worktree/path", "different-branch")
				s.Require().NoError(err)
				return worktree
			}(),
			expectError:  true,
			errorMessage: "worktree already exists at path",
			finalCount:   1,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup project with initial worktrees
			s.Project.Worktrees = tt.setupWorktrees()

			// Add worktree
			err := s.Project.AddWorktree(tt.addWorktree)

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
			} else {
				s.Require().NoError(err)
			}

			// Verify final state
			s.Len(s.Project.Worktrees, tt.finalCount)
		})
	}
}

// TestProject_RemoveWorktree tests removing worktrees with table-driven approach
func (s *ProjectTestSuite) TestProject_RemoveWorktree() {
	testCases := []struct {
		name           string
		setupWorktrees func() []*Worktree
		removePath     string
		expectError    bool
		errorMessage   string
		finalCount     int
	}{
		{
			name: "should remove existing worktree",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			removePath:  "/worktree/path",
			expectError: false,
			finalCount:  0,
		},
		{
			name: "should handle non-existent worktree",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			removePath:   "/nonexistent/path",
			expectError:  true,
			errorMessage: "worktree not found at path",
			finalCount:   1,
		},
		{
			name: "should handle empty worktree list",
			setupWorktrees: func() []*Worktree {
				return []*Worktree{}
			},
			removePath:   "/nonexistent/path",
			expectError:  true,
			errorMessage: "worktree not found at path",
			finalCount:   0,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup project with initial worktrees
			s.Project.Worktrees = tt.setupWorktrees()

			// Remove worktree
			err := s.Project.RemoveWorktree(tt.removePath)

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
			} else {
				s.Require().NoError(err)
			}

			// Verify final state
			s.Len(s.Project.Worktrees, tt.finalCount)
		})
	}
}

// TestProject_GetWorktree tests getting worktrees with table-driven approach
func (s *ProjectTestSuite) TestProject_GetWorktree() {
	testCases := []struct {
		name             string
		setupWorktrees   func() []*Worktree
		getPath          string
		expectError      bool
		errorMessage     string
		expectedWorktree *Worktree
	}{
		{
			name: "should get existing worktree",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			getPath:     "/worktree/path",
			expectError: false,
			expectedWorktree: func() *Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return worktree
			}(),
		},
		{
			name: "should handle non-existent worktree",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/worktree/path", "feature-branch")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			getPath:          "/nonexistent/path",
			expectError:      true,
			errorMessage:     "worktree not found at path",
			expectedWorktree: nil,
		},
		{
			name: "should handle empty worktree list",
			setupWorktrees: func() []*Worktree {
				return []*Worktree{}
			},
			getPath:          "/nonexistent/path",
			expectError:      true,
			errorMessage:     "worktree not found at path",
			expectedWorktree: nil,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup project with initial worktrees
			s.Project.Worktrees = tt.setupWorktrees()

			// Get worktree
			found, err := s.Project.GetWorktree(tt.getPath)

			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
				s.Nil(found)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(found)
				s.Equal(tt.expectedWorktree.Path, found.Path)
				s.Equal(tt.expectedWorktree.Branch, found.Branch)
			}
		})
	}
}

// TestProject_ListBranches tests branch listing with table-driven approach
func (s *ProjectTestSuite) TestProject_ListBranches() {
	testCases := []struct {
		name             string
		setupWorktrees   func() []*Worktree
		expectedCount    int
		expectedBranches []string
	}{
		{
			name: "should handle empty project",
			setupWorktrees: func() []*Worktree {
				return []*Worktree{}
			},
			expectedCount:    0,
			expectedBranches: []string{},
		},
		{
			name: "should list single branch",
			setupWorktrees: func() []*Worktree {
				worktree, err := NewWorktree("/path1", "main")
				s.Require().NoError(err)
				return []*Worktree{worktree}
			},
			expectedCount:    1,
			expectedBranches: []string{"main"},
		},
		{
			name: "should list multiple branches with deduplication",
			setupWorktrees: func() []*Worktree {
				worktree1, err := NewWorktree("/path1", "main")
				s.Require().NoError(err)
				worktree2, err := NewWorktree("/path2", "feature-1")
				s.Require().NoError(err)
				worktree3, err := NewWorktree("/path3", "feature-2")
				s.Require().NoError(err)
				worktree4, err := NewWorktree("/path4", "main") // Duplicate branch
				s.Require().NoError(err)
				return []*Worktree{worktree1, worktree2, worktree3, worktree4}
			},
			expectedCount:    3,
			expectedBranches: []string{"main", "feature-1", "feature-2"},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup project with initial worktrees
			s.Project.Worktrees = tt.setupWorktrees()

			// List branches
			branches := s.Project.ListBranches()

			// Verify results
			s.Len(branches, tt.expectedCount)
			for _, expectedBranch := range tt.expectedBranches {
				s.Contains(branches, expectedBranch)
			}
		})
	}
}

func (s *ProjectTestSuite) TestProject_CanAddWorktree_Pure() {
	// Arrange - deterministic, no I/O
	project := &Project{
		Name:      "test-project",
		GitRepo:   "/path/to/repo",
		Worktrees: []*Worktree{},
	}

	existingWorktree := &Worktree{Path: "/existing/path", Branch: "main"}
	project.Worktrees = append(project.Worktrees, existingWorktree)

	newWorktree := &Worktree{Path: "/new/path", Branch: "feature"}
	duplicateWorktree := &Worktree{Path: "/existing/path", Branch: "feature"}

	// Act & Assert - pure business logic
	s.Require().NoError(project.CanAddWorktree(newWorktree))
	s.Require().Error(project.CanAddWorktree(duplicateWorktree))
}

func (s *ProjectTestSuite) TestProject_HasWorktreeOnBranch_Pure() {
	project := &Project{
		Name: "test-project",
		Worktrees: []*Worktree{
			{Branch: "main", Path: "/main/path"},
			{Branch: "feature", Path: "/feature/path"},
		},
	}

	s.True(project.HasWorktreeOnBranch("main"))
	s.True(project.HasWorktreeOnBranch("feature"))
	s.False(project.HasWorktreeOnBranch("nonexistent"))
}

func (s *ProjectTestSuite) TestProject_AddWorktree_Pure() {
	project := &Project{
		Name:      "test-project",
		GitRepo:   "/path/to/repo",
		Worktrees: []*Worktree{},
	}

	worktree := &Worktree{Path: "/new/path", Branch: "feature"}

	// Act
	err := project.AddWorktree(worktree)

	// Assert
	s.Require().NoError(err)
	s.Len(project.Worktrees, 1)
	s.Equal(worktree, project.Worktrees[0])
}

func (s *ProjectTestSuite) TestProject_EnhancedFeatures() {
	testCases := []struct {
		name         string
		setupProject func() *Project
		testFunc     func(*Project)
		expectError  bool
		errorMessage string
	}{
		{
			name: "should support project metadata",
			setupProject: func() *Project {
				project, err := NewProject("test-project", "/repo/path")
				s.Require().NoError(err)
				return project
			},
			testFunc: func(project *Project) {
				// This should fail initially - we need to add metadata support
				project.SetMetadata("description", "A test project")
				project.SetMetadata("owner", "team-a")
				project.SetMetadata("created-at", "2023-01-01")

				value, exists := project.GetMetadata("description")
				s.True(exists)
				s.Equal("A test project", value)

				value, exists = project.GetMetadata("owner")
				s.True(exists)
				s.Equal("team-a", value)

				_, exists = project.GetMetadata("non-existent")
				s.False(exists)
			},
			expectError: false,
		},
		{
			name: "should provide worktree statistics",
			setupProject: func() *Project {
				project, err := NewProject("test-project", "/repo/path")
				s.Require().NoError(err)

				// Add some worktrees
				worktree1, _ := NewWorktree("/path1", "main")
				worktree2, _ := NewWorktree("/path2", "feature-1")
				worktree3, _ := NewWorktree("/path3", "feature-2")

				s.Require().NoError(project.AddWorktree(worktree1))
				s.Require().NoError(project.AddWorktree(worktree2))
				s.Require().NoError(project.AddWorktree(worktree3))
				return project
			},
			testFunc: func(project *Project) {
				// This should fail initially - we need to add statistics
				stats := project.GetWorktreeStatistics()
				s.NotNil(stats)
				s.Equal(3, stats.TotalCount)
				s.Equal(3, stats.UnknownCount) // All start as unknown
				s.Equal(0, stats.CleanCount)
				s.Equal(0, stats.DirtyCount)
				s.Len(stats.Branches, 3)
			},
			expectError: false,
		},
		{
			name: "should provide project health check",
			setupProject: func() *Project {
				project, err := NewProject("test-project", "/repo/path")
				s.Require().NoError(err)
				return project
			},
			testFunc: func(project *Project) {
				// This should fail initially - we need to add health check
				health := project.GetHealth()
				s.NotNil(health)
				s.Equal("unhealthy", health.Status)
				s.Contains(health.Issues, "git repository not validated")
				s.Equal(0, health.WorktreeCount)
			},
			expectError: false,
		},
		{
			name: "should support project configuration",
			setupProject: func() *Project {
				project, err := NewProject("test-project", "/repo/path")
				s.Require().NoError(err)
				return project
			},
			testFunc: func(project *Project) {
				// This should fail initially - we need to add configuration support
				project.SetConfig("max-worktrees", 10)
				project.SetConfig("auto-cleanup", true)
				project.SetConfig("default-branch", "main")

				value, exists := project.GetConfig("max-worktrees")
				s.True(exists)
				s.Equal(10, value)

				value, exists = project.GetConfig("auto-cleanup")
				s.True(exists)
				s.Equal(true, value)

				value, exists = project.GetConfig("default-branch")
				s.True(exists)
				s.Equal("main", value)

				_, exists = project.GetConfig("non-existent")
				s.False(exists)
			},
			expectError: false,
		},
		{
			name: "should support worktree filtering",
			setupProject: func() *Project {
				project, err := NewProject("test-project", "/repo/path")
				s.Require().NoError(err)

				// Add worktrees with different branches
				worktree1, _ := NewWorktree("/path1", "main")
				worktree2, _ := NewWorktree("/path2", "feature-1")
				worktree3, _ := NewWorktree("/path3", "feature-2")
				worktree4, _ := NewWorktree("/path4", "main")

				s.Require().NoError(project.AddWorktree(worktree1))
				s.Require().NoError(project.AddWorktree(worktree2))
				s.Require().NoError(project.AddWorktree(worktree3))
				s.Require().NoError(project.AddWorktree(worktree4))
				return project
			},
			testFunc: func(project *Project) {
				// This should fail initially - we need to add filtering
				mainWorktrees := project.GetWorktreesByBranch("main")
				s.Len(mainWorktrees, 2)

				featureWorktrees := project.GetWorktreesByBranch("feature-1")
				s.Len(featureWorktrees, 1)

				cleanWorktrees := project.GetWorktreesByStatus(StatusClean)
				s.Empty(cleanWorktrees) // None are clean yet
			},
			expectError: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			project := tt.setupProject()
			tt.testFunc(project)
		})
	}
}

// Test suite entry point
func TestProjectSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
