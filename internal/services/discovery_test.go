package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockGitClient for testing
type MockGitClient struct {
	mock.Mock
}

// DiscoveryServiceTestSuite provides hybrid suite setup for discovery service tests
type DiscoveryServiceTestSuite struct {
	suite.Suite
	GitClient *MockGitClient
	Service   *DiscoveryService
	TempDir   string
	Cleanup   func()
}

// SetupTest initializes infrastructure components for each test
func (s *DiscoveryServiceTestSuite) SetupTest() {
	s.GitClient = &MockGitClient{}
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

func (m *MockGitClient) IsGitRepository(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockGitClient) IsMainRepository(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockGitClient) GetRepositoryRoot(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockGitClient) ListWorktrees(repoPath string) ([]*domain.WorktreeInfo, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]*domain.WorktreeInfo), args.Error(1)
}

func (m *MockGitClient) CreateWorktree(repoPath, branch, targetPath string) error {
	args := m.Called(repoPath, branch, targetPath)
	return args.Error(0)
}

func (m *MockGitClient) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	args := m.Called(repoPath, worktreePath, force)
	return args.Error(0)
}

func (m *MockGitClient) GetWorktreeStatus(worktreePath string) (*domain.WorktreeInfo, error) {
	args := m.Called(worktreePath)
	return args.Get(0).(*domain.WorktreeInfo), args.Error(1)
}

func (m *MockGitClient) GetCurrentBranch(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockGitClient) GetAllBranches(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitClient) GetRemoteBranches(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitClient) BranchExists(repoPath, branch string) bool {
	args := m.Called(repoPath, branch)
	return args.Bool(0)
}

func (m *MockGitClient) HasUncommittedChanges(repoPath string) bool {
	args := m.Called(repoPath)
	return args.Bool(0)
}

// TestDiscoveryService_NewDiscoveryService tests service creation
func (s *DiscoveryServiceTestSuite) TestDiscoveryService_NewDiscoveryService() {
	gitClient := &MockGitClient{}

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
		setupMocks    func(*MockGitClient, string)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should discover worktrees in workspace directory",
			workspacePath: "/tmp/test-workspace",
			setupMocks: func(m *MockGitClient, workspacePath string) {
				// Setup directory structure in test
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree2"), 0755))

				// Mock git repository detection
				m.On("IsGitRepository", filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("ListWorktrees", filepath.Join(workspacePath, "project1")).Return([]*domain.WorktreeInfo{
					{Path: filepath.Join(workspacePath, "project1", "worktree1"), Branch: "main", Commit: "abc123", Clean: true},
					{Path: filepath.Join(workspacePath, "project1", "worktree2"), Branch: "feature", Commit: "def456", Clean: false},
				}, nil)

				// Mock worktree status calls for analysis
				m.On("GetWorktreeStatus", filepath.Join(workspacePath, "project1", "worktree1")).Return(
					&domain.WorktreeInfo{
						Path:   filepath.Join(workspacePath, "project1", "worktree1"),
						Branch: "main",
						Commit: "abc123",
						Clean:  true,
					}, nil)
				m.On("GetWorktreeStatus", filepath.Join(workspacePath, "project1", "worktree2")).Return(
					&domain.WorktreeInfo{
						Path:   filepath.Join(workspacePath, "project1", "worktree2"),
						Branch: "feature",
						Commit: "def456",
						Clean:  false,
					}, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "should handle empty workspace gracefully",
			workspacePath: "/tmp/empty-workspace",
			setupMocks: func(m *MockGitClient, workspacePath string) {
				s.Require().NoError(os.MkdirAll(workspacePath, 0755))
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "should return error for non-existent workspace",
			workspacePath: "/tmp/non-existent",
			setupMocks:    func(m *MockGitClient, workspacePath string) {},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &MockGitClient{}
			service := NewDiscoveryService(mockGit)

			// Cleanup
			defer func() { _ = os.RemoveAll(tt.workspacePath) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath)

			// Test
			worktrees, err := service.DiscoverWorktrees(tt.workspacePath)

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
		setupMocks  func(*MockGitClient)
		expectError bool
	}{
		{
			name: "should return detailed worktree information",
			path: "/tmp/test-worktree",
			setupMocks: func(m *MockGitClient) {
				m.On("GetWorktreeStatus", "/tmp/test-worktree").Return(
					&domain.WorktreeInfo{
						Path:   "/tmp/test-worktree",
						Branch: "feature-branch",
						Commit: "abc123456",
						Clean:  true,
					}, nil)
			},
			expectError: false,
		},
		{
			name: "should handle invalid worktree paths",
			path: "/invalid/path",
			setupMocks: func(m *MockGitClient) {
				m.On("GetWorktreeStatus", "/invalid/path").Return(
					(*domain.WorktreeInfo)(nil),
					errors.New("mock error"))
			},
			expectError: true,
		},
		{
			name: "should return error for empty path",
			path: "",
			setupMocks: func(m *MockGitClient) {
				// No mocks needed for empty path
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &MockGitClient{}
			service := NewDiscoveryService(mockGit)
			tt.setupMocks(mockGit)

			// Test
			worktree, err := service.AnalyzeWorktree(tt.path)

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
		setupMocks    func(*MockGitClient, string)
		expectedCount int
		expectError   bool
	}{
		{
			name:          "should find all git repositories in workspace",
			workspacePath: "/tmp/test-workspace",
			setupMocks: func(m *MockGitClient, workspacePath string) {
				// Create test directory structure
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "project2"), 0755))
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "not-a-project"), 0755))

				// Mock main repository detection
				m.On("IsMainRepository", filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("IsMainRepository", filepath.Join(workspacePath, "project2")).Return(true, nil)
				m.On("IsMainRepository", filepath.Join(workspacePath, "not-a-project")).Return(false, nil)

				// Mock worktree listing for projects
				m.On("ListWorktrees", filepath.Join(workspacePath, "project1")).Return([]*domain.WorktreeInfo{
					{Path: filepath.Join(workspacePath, "project1"), Branch: "main", Commit: "abc123", Clean: true},
				}, nil)
				m.On("ListWorktrees", filepath.Join(workspacePath, "project2")).Return([]*domain.WorktreeInfo{
					{Path: filepath.Join(workspacePath, "project2"), Branch: "main", Commit: "def456", Clean: true},
				}, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "should handle workspace with no git repositories",
			workspacePath: "/tmp/no-repos",
			setupMocks: func(m *MockGitClient, workspacePath string) {
				s.Require().NoError(os.MkdirAll(filepath.Join(workspacePath, "regular-dir"), 0755))
				m.On("IsMainRepository", filepath.Join(workspacePath, "regular-dir")).Return(false, nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			// Setup
			mockGit := &MockGitClient{}
			service := NewDiscoveryService(mockGit)

			// Cleanup
			defer func() { _ = os.RemoveAll(tt.workspacePath) }()

			// Setup mocks
			tt.setupMocks(mockGit, tt.workspacePath)

			// Test
			projects, err := service.DiscoverProjects(tt.workspacePath)

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
		mockGit := &MockGitClient{}
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
			mockGit.On("IsMainRepository", projectPath).Return(true, nil)
			mockGit.On("ListWorktrees", projectPath).Return([]*domain.WorktreeInfo{
				{Path: projectPath, Branch: "main", Commit: "abc123", Clean: true},
			}, nil)
		}

		// Test with timing
		start := time.Now()
		projects, err := service.DiscoverProjects(workspacePath)
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
