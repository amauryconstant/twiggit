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
	"github.com/amaury/twiggit/internal/infrastructure"
	"github.com/amaury/twiggit/internal/infrastructure/config"
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
	Deps      *infrastructure.Deps
}

// SetupTest initializes infrastructure components for each test
func (s *DiscoveryServiceTestSuite) SetupTest() {
	s.GitClient = &mocks.GitClientMock{}

	// Create test deps with mock git client
	s.Deps = &infrastructure.Deps{
		GitClient:  s.GitClient,
		Config:     &config.Config{Workspace: s.T().TempDir()},
		FileSystem: os.DirFS("/tmp"),
	}

	s.Service = NewDiscoveryService(s.Deps)
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

	deps := &infrastructure.Deps{
		GitClient:  gitClient,
		Config:     &config.Config{Workspace: s.T().TempDir()},
		FileSystem: os.DirFS("/tmp"),
	}

	service := NewDiscoveryService(deps)

	s.NotNil(service)
	s.Equal(gitClient, service.deps.GitClient)
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
			deps := &infrastructure.Deps{
				GitClient:  mockGit,
				Config:     &config.Config{Workspace: s.T().TempDir()},
				FileSystem: os.DirFS("/tmp"),
			}
			service := NewDiscoveryService(deps)

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
			deps := &infrastructure.Deps{
				GitClient:  mockGit,
				Config:     &config.Config{Workspace: s.T().TempDir()},
				FileSystem: os.DirFS("/tmp"),
			}
			service := NewDiscoveryService(deps)
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
			deps := &infrastructure.Deps{
				GitClient:  mockGit,
				Config:     &config.Config{Workspace: s.T().TempDir()},
				FileSystem: os.DirFS("/tmp"),
			}
			service := NewDiscoveryService(deps)

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
		deps := &infrastructure.Deps{
			GitClient:  mockGit,
			Config:     &config.Config{Workspace: s.T().TempDir()},
			FileSystem: os.DirFS("/tmp"),
		}
		service := NewDiscoveryService(deps)
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

// TestDiscoveryServiceSuite runs the discovery service test suite
func TestDiscoveryServiceSuite(t *testing.T) {
	suite.Run(t, new(DiscoveryServiceTestSuite))
}
