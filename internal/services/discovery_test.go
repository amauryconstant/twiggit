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
}

// SetupTest initializes infrastructure components for each test
func (s *DiscoveryServiceTestSuite) SetupTest() {
	s.GitClient = &mocks.GitClientMock{}
	s.Service = NewDiscoveryService(s.GitClient)
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

	service := NewDiscoveryService(gitClient)

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
			workspacePath: "/tmp/test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				// Setup directory structure in test
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree2"), 0755))

				// Mock git repository detection for project directory (main repository) and worktree paths
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree1")).Return(true, nil)
				m.On("IsGitRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree2")).Return(true, nil)

				// Mock bare repository detection (return false for all - these are not bare repos)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1")).Return(false, nil)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree1")).Return(false, nil)
				m.On("IsBareRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree2")).Return(false, nil)

				// Mock worktree status calls for analysis
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1")).Return(
					&domain.WorktreeInfo{
						Path:       filepath.Join(workspacePath, "project1"),
						Branch:     "main",
						Commit:     "main123",
						Clean:      true,
						CommitTime: time.Now(),
					}, nil)
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree1")).Return(
					&domain.WorktreeInfo{
						Path:       filepath.Join(workspacePath, "project1", "worktree1"),
						Branch:     "main",
						Commit:     "abc123",
						Clean:      true,
						CommitTime: time.Now(),
					}, nil)
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1", "worktree2")).Return(
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
			workspacePath: "/tmp/empty-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				s.Require().NoError(os.MkdirAll(workspacePath, 0755))
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return empty list for non-existent workspace",
			workspacePath: "/tmp/non-existent",
			setupMocks:    func(m *mocks.GitClientMock, workspacePath string) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			service := NewDiscoveryService(mockGit)

			// Cleanup
			defer func() { _ = os.RemoveAll(tt.workspacePath) }()

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
			path: "/tmp/test-worktree",
			setupMocks: func(m *mocks.GitClientMock) {
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), "/tmp/test-worktree").Return(
					&domain.WorktreeInfo{
						Path:       "/tmp/test-worktree",
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
				m.On("GetWorktreeStatus", mock.AnythingOfType("context.backgroundCtx"), "/invalid/path").Return(
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
			service := NewDiscoveryService(mockGit)
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
			workspacePath: "/tmp/test-workspace",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				// Create test directory structure
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project2"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "not-a-project"), 0755))

				// Mock main repository detection
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "project2")).Return(true, nil)
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "not-a-project")).Return(false, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "should handle workspace with no git repositories",
			workspacePath: "/tmp/no-repos",
			setupMocks: func(m *mocks.GitClientMock, workspacePath string) {
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "regular-dir"), 0755))
				m.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), filepath.Join(workspacePath, "regular-dir")).Return(false, nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return empty list for non-existent workspace",
			workspacePath: "/tmp/non-existent-projects",
			setupMocks:    func(m *mocks.GitClientMock, workspacePath string) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &mocks.GitClientMock{}
			service := NewDiscoveryService(mockGit)

			// Cleanup
			defer func() { _ = os.RemoveAll(tt.workspacePath) }()

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
		service := NewDiscoveryService(mockGit)
		service.SetConcurrency(4) // Test with 4 workers

		workspacePath := "/tmp/perf-test"
		defer func() { _ = os.RemoveAll(workspacePath) }()

		// Create multiple project directories
		projectCount := 10
		for i := 0; i < projectCount; i++ {
			projectPath := filepath.Join(workspacePath, fmt.Sprintf("project%d", i))
			s.Require().NoError(os.MkdirAll(projectPath, 0755))

			// Mock each as main repository
			mockGit.On("IsMainRepository", mock.AnythingOfType("context.backgroundCtx"), projectPath).Return(true, nil)
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
