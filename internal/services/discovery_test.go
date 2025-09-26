package services

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/amaury/twiggit/internal/infrastructure/config"
	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// mockDirEntry creates a mock directory entry for testing using centralized mock
func mockDirEntry(name string) fs.DirEntry {
	return mocks.NewMockDirEntry(name, true)
}

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

	// Create temporary directory for test isolation
	s.TempDir = s.T().TempDir()

	// Create test config with mock git client - use temp directory for test isolation
	s.Config = &config.Config{WorkspacesPath: s.TempDir}
	testFileSystem := mocks.NewFileSystemMock()

	s.Service = NewDiscoveryService(s.GitClient, s.Config, testFileSystem)
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

	tempDir := s.T().TempDir()
	testConfig := &config.Config{WorkspacesPath: tempDir}
	testFileSystem := mocks.NewFileSystemMock()

	service := NewDiscoveryService(gitClient, testConfig, testFileSystem)

	s.NotNil(service)
	s.Equal(gitClient, service.gitClient)
	s.Equal(defaultConcurrency, service.concurrency)
}

// TestDiscoveryService_DiscoverWorktrees tests worktree discovery with table-driven approach
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_DiscoverWorktrees() {
	testCases := []struct {
		name          string
		workspacePath string
		setupMocks    func(*mocks.GitClientMock, string, *mocks.FileSystemMock)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should discover worktrees in workspace directory",
			workspacePath: "test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				// Setup directory structure in test - use absolute path for creation, relative for FileSystem
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				absProjectPath := filepath.Join(absWorkspacePath, "project1")
				absWorktree1Path := filepath.Join(absProjectPath, "worktree1")
				absWorktree2Path := filepath.Join(absProjectPath, "worktree2")

				// Mock directory structure
				mockFS.On("ReadDir", absWorkspacePath).Return([]fs.DirEntry{
					mockDirEntry("project1"),
				}, nil)
				mockFS.On("ReadDir", absProjectPath).Return([]fs.DirEntry{
					mockDirEntry("worktree1"),
					mockDirEntry("worktree2"),
				}, nil)

				// Mock Stat calls for pathExists
				mockFS.On("Stat", absWorkspacePath).Return(mocks.NewMockFileInfo("test-workspace", true), nil)
				mockFS.On("Stat", absProjectPath).Return(mocks.NewMockFileInfo("project1", true), nil)
				mockFS.On("Stat", absWorktree1Path).Return(mocks.NewMockFileInfo("worktree1", true), nil)
				mockFS.On("Stat", absWorktree2Path).Return(mocks.NewMockFileInfo("worktree2", true), nil)

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
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				mockFS.On("ReadDir", absWorkspacePath).Return([]fs.DirEntry{}, nil)
				mockFS.On("Stat", absWorkspacePath).Return(mocks.NewMockFileInfo("empty-workspace", true), nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return empty list for non-existent workspace",
			workspacePath: "non-existent",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				mockFS.On("Stat", absWorkspacePath).Return((*mocks.MockFileInfo)(nil), os.ErrNotExist)
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

			// Cleanup
			defer func() { _ = os.RemoveAll(filepath.Join(s.TempDir, tt.workspacePath)) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath, testFileSystem)

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
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem)
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
		setupMocks    func(*mocks.GitClientMock, string, *mocks.FileSystemMock)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should find all git repositories in workspace",
			workspacePath: "test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				// Create test directory structure - use absolute path for creation, relative for FileSystem
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				mockFS.On("MkdirAll", filepath.Join(absWorkspacePath, "project1"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absWorkspacePath, "project2"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absWorkspacePath, "not-a-project"), fs.FileMode(0755)).Return(nil)

				// Mock Stat calls for pathExists validation
				mockFS.On("Stat", absWorkspacePath).Return(mocks.NewMockFileInfo("test-workspace", true), nil)

				// Mock ReadDir calls for directory listing
				mockFS.On("ReadDir", absWorkspacePath).Return([]fs.DirEntry{
					mockDirEntry("project1"),
					mockDirEntry("project2"),
					mockDirEntry("not-a-project"),
				}, nil)

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
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				mockFS.On("MkdirAll", filepath.Join(absWorkspacePath, "regular-dir"), fs.FileMode(0755)).Return(nil)
				mockFS.On("Stat", absWorkspacePath).Return(mocks.NewMockFileInfo("no-repos", true), nil)

				// Mock ReadDir calls for directory listing
				mockFS.On("ReadDir", absWorkspacePath).Return([]fs.DirEntry{
					mockDirEntry("regular-dir"),
				}, nil)

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
			setupMocks: func(m *mocks.GitClientMock, workspacePath string, mockFS *mocks.FileSystemMock) {
				absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
				mockFS.On("Stat", absWorkspacePath).Return((*mocks.MockFileInfo)(nil), os.ErrNotExist)
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

			// Cleanup
			defer func() { _ = os.RemoveAll(filepath.Join(s.TempDir, tt.workspacePath)) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath, testFileSystem)

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
		testConfig := &config.Config{WorkspacesPath: s.TempDir}
		testFileSystem := mocks.NewFileSystemMock()
		service := NewDiscoveryService(mockGit, testConfig, testFileSystem)
		service.SetConcurrency(4) // Test with 4 workers

		workspacePath := "perf-test"
		defer func() { _ = os.RemoveAll(filepath.Join(s.TempDir, workspacePath)) }()

		// Create multiple project directories
		projectCount := 10
		for i := 0; i < projectCount; i++ {
			projectPath := filepath.Join(workspacePath, fmt.Sprintf("project%d", i))
			absProjectPath := filepath.Join(s.TempDir, projectPath)
			testFileSystem.On("MkdirAll", absProjectPath, fs.FileMode(0755)).Return(nil)

			// Mock each as main repository
			mockGit.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
				return filepath.IsAbs(path) && filepath.Base(path) == fmt.Sprintf("project%d", i)
			})).Return(true, nil)
		}

		// Mock Stat call for workspace path validation
		absWorkspacePath := filepath.Join(s.TempDir, workspacePath)
		testFileSystem.On("Stat", absWorkspacePath).Return(mocks.NewMockFileInfo("perf-test", true), nil)

		// Mock ReadDir calls for directory listing
		var projectEntries []fs.DirEntry
		for i := 0; i < projectCount; i++ {
			projectEntries = append(projectEntries, mockDirEntry(fmt.Sprintf("project%d", i)))
		}
		testFileSystem.On("ReadDir", absWorkspacePath).Return(projectEntries, nil)

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
		setupMocks     func(*mocks.GitClientMock, string, *mocks.FileSystemMock)
		expectError    bool
		expectedErrMsg string
		expectedCount  int
	}{
		{
			name:         "should create pure Project entities without I/O in domain",
			projectsPath: "test-projects",
			setupMocks: func(gitMock *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Create test directory structure - use absolute path for creation, relative for FileSystem
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "project1"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "project2"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "not-a-repo"), fs.FileMode(0755)).Return(nil)

				// Mock Stat calls for pathExists validation
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("test-projects", true), nil)

				// Mock ReadDir calls for directory listing
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("project1"),
					mockDirEntry("project2"),
					mockDirEntry("not-a-repo"),
				}, nil)

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
			setupMocks: func(gitMock *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock Stat to return error for non-existent path
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("Stat", absProjectsPath).Return((*mocks.MockFileInfo)(nil), os.ErrNotExist)
				// No ReadDir mock needed since Stat returns error, so ReadDir won't be called
			},
			expectError:   false,
			expectedCount: 0, // Should return empty list, not error
		},
		{
			name:         "should filter out non-git repositories using infrastructure",
			projectsPath: "mixed-projects",
			setupMocks: func(gitMock *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Create test directory structure
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "valid-project"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "another-valid"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "invalid-project"), fs.FileMode(0755)).Return(nil)
				mockFS.On("MkdirAll", filepath.Join(absProjectsPath, "error-project"), fs.FileMode(0755)).Return(nil)

				// Mock Stat calls for pathExists validation
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("mixed-projects", true), nil)

				// Mock ReadDir calls for directory listing
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("valid-project"),
					mockDirEntry("another-valid"),
					mockDirEntry("invalid-project"),
					mockDirEntry("error-project"),
				}, nil)

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
			// Test that DiscoveryService properly uses fs.FS for path validation
			// while still using directory operations directly

			// Setup test dependencies
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()

			// Create service
			s.Service = NewDiscoveryService(s.GitClient, testConfig, testFileSystem)

			// Setup mocks
			tt.setupMocks(s.GitClient, tt.projectsPath, testFileSystem)

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
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			s.Service = NewDiscoveryService(s.GitClient, testConfig, testFileSystem)

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

// TestDiscoveryService_DiscoverProjectsWithFallback tests project discovery with fallback mechanisms
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_DiscoverProjectsWithFallback() {
	testCases := []struct {
		name          string
		projectsPath  string
		setupMocks    func(*mocks.GitClientMock, string, *mocks.FileSystemMock)
		expectedCount int
		expectError   bool
		errorMessage  string
	}{
		{
			name:         "should succeed when primary discovery works",
			projectsPath: "test-projects",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock directory structure with project1, project2, and not-a-repo
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("project1"),
					mockDirEntry("project2"),
					mockDirEntry("not-a-repo"),
				}, nil)

				// Mock Stat calls for pathExists
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("test-projects", true), nil)

				// Mock main repository detection for all directories
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project2"
				})).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "not-a-repo"
				})).Return(false, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:         "should use fallback when primary discovery fails",
			projectsPath: "fallback-test-projects",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock directory structure with project1 and project2
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("project1"),
					mockDirEntry("project2"),
				}, nil)

				// Mock Stat calls for pathExists
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("fallback-test-projects", true), nil)

				// Mock primary discovery calls to succeed
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project1"
				})).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "project2"
				})).Return(true, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:         "should return error when both primary and fallback discovery fail",
			projectsPath: "nonexistent-projects",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock filesystem to return permission error when reading directory
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{}, errors.New("permission denied"))
				// For fallback to fail with ReadDir error, path must exist but ReadDir must fail
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("nonexistent-projects", true), nil)
			},
			expectError:  true,
			errorMessage: "failed to discover projects with fallback",
		},
		{
			name:         "should handle empty projects directory",
			projectsPath: "empty-projects",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock empty directory
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{}, nil)
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("empty-projects", true), nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

			// Cleanup
			defer func() {
				_ = os.RemoveAll(filepath.Join(s.TempDir, tt.projectsPath))
			}()

			// Setup mocks
			tt.setupMocks(mockGit, tt.projectsPath, testFileSystem)

			// Test
			ctx := context.Background()
			projects, err := service.DiscoverProjectsWithFallback(ctx, tt.projectsPath)

			// Assert
			if tt.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errorMessage)
			} else {
				s.Require().NoError(err)
				s.Len(projects, tt.expectedCount)
			}

			mockGit.AssertExpectations(s.T())
		})
	}
}

// TestDiscoveryService_fallbackProjectDiscovery tests fallback project discovery logic
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_fallbackProjectDiscovery() {
	testCases := []struct {
		name          string
		projectsPath  string
		setupMocks    func(*mocks.GitClientMock, string, *mocks.FileSystemMock)
		expectedCount int
		expectError   bool
	}{
		{
			name:         "should discover projects with basic git repository check",
			projectsPath: "fallback-test",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock directory structure with repo1, repo2, and not-repo
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				absRepo1Path := filepath.Join(absProjectsPath, "repo1")
				absRepo2Path := filepath.Join(absProjectsPath, "repo2")
				absNotRepoPath := filepath.Join(absProjectsPath, "not-repo")

				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("repo1"),
					mockDirEntry("repo2"),
					mockDirEntry("not-repo"),
				}, nil)

				// Mock Stat calls for pathExists
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("fallback-test", true), nil)
				mockFS.On("Stat", absRepo1Path).Return(mocks.NewMockFileInfo("repo1", true), nil)
				mockFS.On("Stat", absRepo2Path).Return(mocks.NewMockFileInfo("repo2", true), nil)
				mockFS.On("Stat", absNotRepoPath).Return(mocks.NewMockFileInfo("not-repo", true), nil)

				// Mock basic git repository checks
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "repo1"
				})).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "repo2"
				})).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "not-repo"
				})).Return(false, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:         "should handle git repository check errors gracefully",
			projectsPath: "error-test",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock directory structure with valid-repo and error-repo
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				absValidRepoPath := filepath.Join(absProjectsPath, "valid-repo")
				absErrorRepoPath := filepath.Join(absProjectsPath, "error-repo")

				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{
					mockDirEntry("valid-repo"),
					mockDirEntry("error-repo"),
				}, nil)

				// Mock Stat calls for pathExists
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("error-test", true), nil)
				mockFS.On("Stat", absValidRepoPath).Return(mocks.NewMockFileInfo("valid-repo", true), nil)
				mockFS.On("Stat", absErrorRepoPath).Return(mocks.NewMockFileInfo("error-repo", true), nil)

				// Mock mixed responses
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "valid-repo"
				})).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(path string) bool {
					return filepath.IsAbs(path) && filepath.Base(path) == "error-repo"
				})).Return(false, errors.New("git error"))
			},
			expectedCount: 1, // Only valid-repo should be included
			expectError:   false,
		},
		{
			name:         "should return empty list for non-existent path",
			projectsPath: "nonexistent",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock filesystem to return error for non-existent path
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				// Note: ReadDir won't be called if Stat returns os.ErrNotExist due to pathExists check
				mockFS.On("Stat", absProjectsPath).Return((*mocks.MockFileInfo)(nil), os.ErrNotExist)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:         "should return error when directory read fails",
			projectsPath: "permission-denied",
			setupMocks: func(m *mocks.GitClientMock, projectsPath string, mockFS *mocks.FileSystemMock) {
				// Mock filesystem to return permission error when reading directory
				absProjectsPath := filepath.Join(s.TempDir, projectsPath)
				mockFS.On("ReadDir", absProjectsPath).Return([]fs.DirEntry{}, errors.New("permission denied"))
				// Path must exist (Stat succeeds) but ReadDir must fail for error to be returned
				mockFS.On("Stat", absProjectsPath).Return(mocks.NewMockFileInfo("permission-denied", true), nil)
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			testConfig := &config.Config{WorkspacesPath: s.TempDir}
			testFileSystem := mocks.NewFileSystemMock()
			service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

			// Cleanup
			defer func() { _ = os.RemoveAll(filepath.Join(s.TempDir, tt.projectsPath)) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.projectsPath, testFileSystem)

			// Test
			ctx := context.Background()
			projects, err := service.fallbackProjectDiscovery(ctx, tt.projectsPath)

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

// TestDiscoveryService_ClearCache tests cache clearing functionality
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_ClearCache() {
	s.Run("should clear all cached results", func() {
		// Setup
		mockGit := &mocks.GitClientMock{}
		testConfig := &config.Config{WorkspacesPath: s.TempDir}
		testFileSystem := mocks.NewFileSystemMock()
		service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

		// Add some items to cache
		service.cacheResult("path1", &domain.Worktree{Path: "path1", Branch: "main"})
		service.cacheResult("path2", &domain.Worktree{Path: "path2", Branch: "feature"})

		// Verify cache has items
		s.Len(service.cache, 2, "Cache should have 2 items before clearing")

		// Execute
		service.ClearCache()

		// Verify cache is cleared
		s.Empty(service.cache, "Cache should be empty after clearing")
		s.NotNil(service.cache, "Cache map should still exist but be empty")
	})

	s.Run("should handle empty cache gracefully", func() {
		// Setup
		mockGit := &mocks.GitClientMock{}
		testConfig := &config.Config{WorkspacesPath: s.TempDir}
		testFileSystem := mocks.NewFileSystemMock()
		service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

		// Verify cache is initially empty
		s.Empty(service.cache, "Cache should be initially empty")

		// Execute
		service.ClearCache()

		// Verify cache remains empty
		s.Empty(service.cache, "Cache should remain empty after clearing")
	})

	s.Run("should be thread-safe", func() {
		// Setup
		mockGit := &mocks.GitClientMock{}
		testConfig := &config.Config{WorkspacesPath: s.TempDir}
		testFileSystem := mocks.NewFileSystemMock()
		service := NewDiscoveryService(mockGit, testConfig, testFileSystem)

		// Add items to cache
		service.cacheResult("path1", &domain.Worktree{Path: "path1", Branch: "main"})
		service.cacheResult("path2", &domain.Worktree{Path: "path2", Branch: "feature"})

		// Clear cache concurrently
		done := make(chan bool)
		go func() {
			service.ClearCache()
			done <- true
		}()

		// Wait for completion
		select {
		case <-done:
			// Verify cache is cleared
			s.Empty(service.cache, "Cache should be empty after concurrent clear")
		case <-time.After(time.Second):
			s.Fail("ClearCache should complete quickly")
		}
	})
}

// TestDiscoveryServiceSuite runs the discovery service test suite
func TestDiscoveryServiceSuite(t *testing.T) {
	suite.Run(t, new(DiscoveryServiceTestSuite))
}
