package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/internal/infrastructure/validation"
	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// DiscoveryServiceTestSuite provides hybrid suite setup for discovery service tests
type DiscoveryServiceTestSuite struct {
	suite.Suite
	GitClient *mocks.GitClientMock
	Service   *DiscoveryService
	TempDir   string
	Cleanup   func()
	Config    *config.Config
}

// SetupTest initializes infrastructure components for each test
func (s *DiscoveryServiceTestSuite) SetupTest() {
	s.GitClient = &mocks.GitClientMock{}

	// Create test config with mock git client
	s.Config = &config.Config{Workspace: s.T().TempDir()}
	testFileSystem := os.DirFS("/tmp")

	s.Service = NewDiscoveryService(s.GitClient, s.Config, testFileSystem, nil)
	s.TempDir = s.T().TempDir()
	s.Cleanup = func() {
		_ = os.RemoveAll(s.TempDir)
	}
}

// TearDownTest cleans up infrastructure test resources
func (s *DiscoveryServiceTestSuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

// TestDiscoveryService_NewDiscoveryService tests service creation
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_NewDiscoveryService() {
	gitClient := &mocks.GitClientMock{}

	testConfig := &config.Config{Workspace: s.T().TempDir()}
	testFileSystem := os.DirFS("/tmp")

	service := NewDiscoveryService(gitClient, testConfig, testFileSystem, nil)

	s.NotNil(service)
	s.Equal(gitClient, service.gitClient)
	s.Equal(defaultConcurrency, service.concurrency)
}

// TestDiscoveryService_DiscoverWorktrees tests worktree discovery with table-driven approach
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_DiscoverWorktrees() {
	testCases := []struct {
		name          string
		workspacePath string
		setupMocks    func(*mocks.GitClientMock, string)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should discover worktrees in workspace directory",
			workspacePath: "test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				// Setup directory structure in test - use absolute path for creation, relative for FileSystem
				absWorkspacePath := filepath.Join("/tmp", workspacePath)
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "project1", "worktree1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "project1", "worktree2"), 0755))

				// Mock git repository detection for project directory (main repository) and worktree paths
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree1"
				})).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree2"
				})).Return(true, nil)

				// Mock bare repository detection (return false for all - these are not bare repos)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(false, nil)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree1"
				})).Return(false, nil)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree2"
				})).Return(false, nil)

				// Mock worktree status calls for analysis
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(
					&domain.WorktreeInfo{
						Path:       filepath.Join(workspacePath, "project1"),
						Branch:     "main",
						Commit:     "main123",
						Clean:      true,
						CommitTime: time.Now(),
					}, nil)
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree1"
				})).Return(
					&domain.WorktreeInfo{
						Path:       filepath.Join(workspacePath, "project1", "worktree1"),
						Branch:     "main",
						Commit:     "abc123",
						Clean:      true,
						CommitTime: time.Now(),
					}, nil)
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "worktree2"
				})).Return(
					&domain.WorktreeInfo{
						Path:       filepath.Join(workspacePath, "project1", "worktree2"),
						Branch:     "feature",
						Commit:     "def456",
						Clean:      false,
						CommitTime: time.Now(),
					}, nil)
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "should handle empty workspace gracefully",
			workspacePath: "empty-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				absWorkspacePath := filepath.Join("/tmp", workspacePath)
				s.Require().NoError(os.MkdirAll(absWorkspacePath, 0755))
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return empty list for non-existent workspace",
			workspacePath: "non-existent",
			setupMocks:    func(m *mocks.GitClientMock, workspacePath string) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{Workspace: s.T().TempDir()}
			testFileSystem := os.DirFS("/tmp")
			pathValidator := validation.NewPathValidator()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem, pathValidator)

			// Cleanup
			defer func() { _ = os.RemoveAll(filepath.Join("/tmp", tt.workspacePath)) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath)

			// Test
			ctx := context.Background()
			worktrees, err := service.DiscoverWorktrees(ctx, tt.workspacePath)

			// Assert
			if tt.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Len(worktrees, tt.expectedCount)
			}

			mockGit.AssertExpectations(s.T())
		})
	}
}

// TestDiscoveryService_AnalyzeWorktree tests worktree analysis with table-driven approach
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_AnalyzeWorktree() {
	testCases := []struct {
		name        string
		path        string
		setupMocks  func(*mocks.GitClientMock)
		expectError bool
	}{
		{
			name: "should return detailed worktree information",
			path: "test-worktree",
			setupMocks: func(m *mocks.GitClientMock) {
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					// The path should be converted to an absolute path by convertToAbsolutePath
					return filepath.IsAbs(path) && filepath.Base(path) == "test-worktree"
				})).Return(
					&domain.WorktreeInfo{
						Path:       "test-worktree",
						Branch:     "feature-branch",
						Commit:     "abc123456",
						Clean:      true,
						CommitTime: time.Now(),
					}, nil)
			},
			expectError: false,
		},
		{
			name: "should handle invalid worktree paths",
			path: "/invalid/path",
			setupMocks: func(m *mocks.GitClientMock) {
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					// The path should remain absolute since it's already absolute, but may be prefixed with temp dir
					return filepath.IsAbs(path) && filepath.Base(path) == "path"
				})).Return(
					(*domain.WorktreeInfo)(nil),
					errors.New("mock error"))
			},
			expectError: true,
		},
		{
			name: "should return error for empty path",
			path: "",
			setupMocks: func(m *mocks.GitClientMock) {
				// No mocks needed for empty path
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{Workspace: s.T().TempDir()}
			testFileSystem := os.DirFS("/tmp")
			pathValidator := validation.NewPathValidator()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem, pathValidator)
			tt.setupMocks(mockGit)

			// Test
			ctx := context.Background()
			worktree, err := service.AnalyzeWorktree(ctx, tt.path)

			// Assert
			if tt.expectError {
				s.Require().Error(err)
				s.Nil(worktree)
			} else {
				s.Require().NoError(err)
				s.NotNil(worktree)
				s.Equal(tt.path, worktree.Path)
			}

			mockGit.AssertExpectations(s.T())
		})
	}
}

// TestDiscoveryService_DiscoverProjects tests project discovery with table-driven approach
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_DiscoverProjects() {
	testCases := []struct {
		name          string
		workspacePath string
		setupMocks    func(*mocks.GitClientMock, string)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should find all git repositories in workspace",
			workspacePath: "test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				// Create test directory structure - use absolute path for creation, relative for FileSystem
				absWorkspacePath := filepath.Join("/tmp", workspacePath)
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "project2"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "not-a-project"), 0755))

				// Mock main repository detection
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project2"
				})).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "not-a-project"
				})).Return(false, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "should handle workspace with no git repositories",
			workspacePath: "no-repos",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				absWorkspacePath := filepath.Join("/tmp", workspacePath)
				s.Require().NoError(os.MkdirAll(filepath.Join(absWorkspacePath, "regular-dir"), 0755))
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "regular-dir"
				})).Return(false, nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return empty list for non-existent workspace",
			workspacePath: "non-existent-projects",
			setupMocks:    func(m *mocks.GitClientMock, workspacePath string) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{Workspace: s.T().TempDir()}
			testFileSystem := os.DirFS("/tmp")
			pathValidator := validation.NewPathValidator()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem, pathValidator)

			// Cleanup
			defer func() { _ = os.RemoveAll(filepath.Join("/tmp", tt.workspacePath)) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath)

			// Test
			ctx := context.Background()
			projects, err := service.DiscoverProjects(ctx, tt.workspacePath)

			// Assert
			if tt.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Len(projects, tt.expectedCount)
			}

			mockGit.AssertExpectations(s.T())
		})
	}
}

// TestDiscoveryService_Performance tests performance with sub-tests
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_Performance() {
	s.Run("should handle concurrent discovery efficiently", func() {
		// Setup
		mockGit := &mocks.GitClientMock{}
		testConfig := &config.Config{Workspace: s.T().TempDir()}
		testFileSystem := os.DirFS("/tmp")
		pathValidator := validation.NewPathValidator()
		service := NewDiscoveryService(mockGit, testConfig, testFileSystem, pathValidator)
		service.SetConcurrency(4) // Test with 4 workers

		workspacePath := "perf-test"
		defer func() { _ = os.RemoveAll(filepath.Join("/tmp", workspacePath)) }()

		// Create multiple project directories
		projectCount := 10
		for i := 0; i < projectCount; i++ {
			projectPath := filepath.Join(workspacePath, fmt.Sprintf("project%d", i))
			absProjectPath := filepath.Join("/tmp", projectPath)
			s.Require().NoError(os.MkdirAll(absProjectPath, 0755))

			// Mock each as main repository
			mockGit.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
				return filepath.IsAbs(path) && filepath.Base(path) == fmt.Sprintf("project%d", i)
			})).Return(true, nil)
		}

		// Test with timing
		start := time.Now()
		ctx := context.Background()
		projects, err := service.DiscoverProjects(ctx, workspacePath)
		duration := time.Since(start)

		// Assert
		s.Require().NoError(err)
		s.Len(projects, projectCount)
		// Should complete quickly with concurrency
		s.Less(duration, time.Second, "Discovery should complete quickly with concurrent processing")

		mockGit.AssertExpectations(s.T())
	})
}

// TestDiscoveryService_DiscoverProjects_WithPureDomain tests integration with pure domain entities
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_DiscoverProjects_WithPureDomain() {
	// Test that DiscoveryService properly creates pure domain entities
	// without I/O operations in the domain layer

	testCases := []struct {
		name           string
		projectsPath   string
		setupMocks     func(*mocks.GitClientMock, *mocks.InfrastructureServiceMock, string)
		expectError    bool
		expectedErrMsg string
		expectedCount  int
	}{
		{
			name:         "should create pure Project entities without I/O in domain",
			projectsPath: "test-projects",
			setupMocks: func(gitMock *mocks.GitClientMock, infraMock *mocks.InfrastructureServiceMock, projectsPath string) {
				// Create test directory structure - use absolute path for creation, relative for FileSystem
				absProjectsPath := filepath.Join("/tmp", projectsPath)
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "project2"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "not-a-repo"), 0755))

				// Infrastructure validation: check if projects path exists
				infraMock.On("PathExists", projectsPath).Return(true)

				// Mock git repository checks
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(true, nil)
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project2"
				})).Return(true, nil)
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "not-a-repo"
				})).Return(false, nil)
			},
			expectError:   false,
			expectedCount: 2, // project1 and project2, not not-a-repo
		},
		{
			name:         "should return empty list for non-existent projects path",
			projectsPath: "nonexistent-projects",
			setupMocks: func(gitMock *mocks.GitClientMock, infraMock *mocks.InfrastructureServiceMock, projectsPath string) {
				// Infrastructure validation: path doesn't exist
				infraMock.On("PathExists", projectsPath).Return(false)
				// No git calls should be made since infrastructure validation fails
			},
			expectError:   false,
			expectedCount: 0, // Should return empty list, not error
		},
		{
			name:         "should filter out non-git repositories using infrastructure",
			projectsPath: "mixed-projects",
			setupMocks: func(gitMock *mocks.GitClientMock, infraMock *mocks.InfrastructureServiceMock, projectsPath string) {
				// Create test directory structure
				absProjectsPath := filepath.Join("/tmp", projectsPath)
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "valid-project"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "another-valid"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "invalid-project"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(absProjectsPath, "error-project"), 0755))

				// Infrastructure validation: path exists
				infraMock.On("PathExists", projectsPath).Return(true)

				// Mock git checks - only some are valid repositories
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "valid-project"
				})).Return(true, nil)
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "another-valid"
				})).Return(true, nil)
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "invalid-project"
				})).Return(false, nil)
				gitMock.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "error-project"
				})).Return(false, errors.New("git error"))
			},
			expectError:   false,
			expectedCount: 2, // Only valid projects should be included
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Test that DiscoveryService properly uses InfrastructureService for path validation
			// while still using deps.ReadDir() for directory operations

			// Setup infrastructure service mock
			infraMock := &mocks.InfrastructureServiceMock{}

			// Setup test dependencies with infrastructure service
			testConfig := &config.Config{Workspace: s.T().TempDir()}
			testFileSystem := os.DirFS("/tmp")
			pathValidator := validation.NewPathValidator()

			// Create service with infrastructure integration
			s.Service = NewDiscoveryServiceWithInfra(s.GitClient, testConfig, testFileSystem, pathValidator, infraMock)

			// Setup mocks
			tt.setupMocks(s.GitClient, infraMock, tt.projectsPath)

			ctx := context.Background()
			projects, err := s.Service.DiscoverProjects(ctx, tt.projectsPath)

			if tt.expectError {
				s.Require().Error(err)
				if tt.expectedErrMsg != "" {
					s.Contains(err.Error(), tt.expectedErrMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Len(projects, tt.expectedCount)

				// Verify that created projects are pure domain entities
				for _, project := range projects {
					s.NotNil(project)
					s.NotEmpty(project.Name)
					s.NotEmpty(project.GitRepo)
					// Project should be a pure domain entity with no I/O operations
					s.NotNil(project.Worktrees)
					s.NotNil(project.Metadata)
				}
			}

			s.GitClient.AssertExpectations(s.T())
			infraMock.AssertExpectations(s.T())
		})
	}
}

// TestDiscoveryService_AnalyzeWorktree_WithPureDomain tests worktree analysis with pure domain entities
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_AnalyzeWorktree_WithPureDomain() {
	// Test that AnalyzeWorktree creates pure domain entities
	// with deterministic timestamps and no I/O operations

	// Setup: Mock git client to return worktree status with specific commit time
	expectedCommitTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	s.GitClient.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
		return filepath.IsAbs(path) && filepath.Base(path) == "test-worktree"
	})).Return(
		&domain.WorktreeInfo{
			Path:       "test-worktree",
			Branch:     "feature-branch",
			Commit:     "abc123456",
			Clean:      true,
			CommitTime: expectedCommitTime,
		}, nil)

	// Execute: Call AnalyzeWorktree
	worktree, err := s.Service.AnalyzeWorktree(context.Background(), "test-worktree")

	// Verify
	s.Require().NoError(err, "AnalyzeWorktree should not return an error")
	s.Require().NotNil(worktree, "Worktree should not be nil")

	// 1. Verify that returned worktree is a pure domain entity
	s.Equal("test-worktree", worktree.Path, "Worktree path should match")
	s.Equal("feature-branch", worktree.Branch, "Worktree branch should match")
	s.Equal("abc123456", worktree.Commit, "Worktree commit should match")
	s.Equal(domain.StatusClean, worktree.Status, "Worktree status should be clean")

	// 2. Verify that LastUpdated is set from commit time, not time.Now()
	s.Equal(expectedCommitTime, worktree.LastUpdated, "LastUpdated should be set from commit time")
	s.NotEqual(time.Now(), worktree.LastUpdated, "LastUpdated should not be current time")

	// 3. Verify that the worktree is a proper domain entity with validation
	s.True(worktree.IsClean(), "Worktree should be clean")
	s.NotEmpty(worktree.Path, "Worktree should have a valid path")
	s.NotEmpty(worktree.Branch, "Worktree should have a valid branch")

	// 4. Verify no I/O operations are called in domain entity creation
	// (This is verified by the fact that we only mocked GetWorktreeStatus once)
	s.GitClient.AssertExpectations(s.T())
}

// TestDiscoveryService_ConvertToWorktree_WithPureDomain tests conversion to pure domain entities
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_ConvertToWorktree_WithPureDomain() {
	// Test that convertToWorktree creates pure domain entities
	// with proper status mapping and deterministic timestamps

	testCases := []struct {
		name           string
		worktreeInfo   *domain.WorktreeInfo
		expectError    bool
		expectedStatus domain.WorktreeStatus
	}{
		{
			name: "should create clean worktree with commit timestamp",
			worktreeInfo: &domain.WorktreeInfo{
				Path:       "/test/worktree",
				Branch:     "main",
				Commit:     "abc123",
				Clean:      true,
				CommitTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectError:    false,
			expectedStatus: domain.StatusClean,
		},
		{
			name: "should create dirty worktree with commit timestamp",
			worktreeInfo: &domain.WorktreeInfo{
				Path:       "/test/worktree",
				Branch:     "feature",
				Commit:     "def456",
				Clean:      false,
				CommitTime: time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC),
			},
			expectError:    false,
			expectedStatus: domain.StatusDirty,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			testConfig := &config.Config{Workspace: s.T().TempDir()}
			testFileSystem := os.DirFS("/tmp")
			pathValidator := validation.NewPathValidator()
			s.Service = NewDiscoveryService(s.GitClient, testConfig, testFileSystem, pathValidator)

			worktree, err := s.Service.convertToWorktree(tt.worktreeInfo)

			if tt.expectError {
				s.Require().Error(err)
				s.Nil(worktree)
			} else {
				s.Require().NoError(err)
				s.NotNil(worktree)

				// Verify pure domain entity properties
				s.Equal(tt.worktreeInfo.Path, worktree.Path)
				s.Equal(tt.worktreeInfo.Branch, worktree.Branch)
				s.Equal(tt.expectedStatus, worktree.Status)
				s.Equal(tt.worktreeInfo.CommitTime, worktree.LastUpdated)
				s.Equal(tt.worktreeInfo.Commit, worktree.Commit)

				// Verify that timestamp is deterministic (from commit, not current time)
				s.NotEqual(time.Now(), worktree.LastUpdated)
				s.Equal(tt.worktreeInfo.CommitTime, worktree.LastUpdated)
			}
		})
	}
}

// TestDiscoveryServiceSuite runs the discovery service test suite
func TestDiscoveryServiceSuite(t *testing.T) {
	suite.Run(t, new(DiscoveryServiceTestSuite))
}
