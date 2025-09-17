package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockGitClient for testing
type MockGitClient struct {
	mock.Mock
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

func (m *MockGitClient) ListWorktrees(repoPath string) ([]*types.WorktreeInfo, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]*types.WorktreeInfo), args.Error(1)
}

func (m *MockGitClient) CreateWorktree(repoPath, branch, targetPath string) error {
	args := m.Called(repoPath, branch, targetPath)
	return args.Error(0)
}

func (m *MockGitClient) RemoveWorktree(repoPath, worktreePath string, force bool) error {
	args := m.Called(repoPath, worktreePath, force)
	return args.Error(0)
}

func (m *MockGitClient) GetWorktreeStatus(worktreePath string) (*types.WorktreeInfo, error) {
	args := m.Called(worktreePath)
	return args.Get(0).(*types.WorktreeInfo), args.Error(1)
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

func TestDiscoveryService_NewDiscoveryService(t *testing.T) {
	gitClient := &MockGitClient{}

	service := NewDiscoveryService(gitClient)

	assert.NotNil(t, service)
	assert.Equal(t, gitClient, service.gitClient)
	assert.Equal(t, defaultConcurrency, service.concurrency)
}

func TestDiscoveryService_DiscoverWorktrees(t *testing.T) {
	tests := []struct {
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
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree1"), 0755))
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "project1", "worktree2"), 0755))

				// Mock git repository detection
				m.On("IsGitRepository", filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("ListWorktrees", filepath.Join(workspacePath, "project1")).Return([]*types.WorktreeInfo{
					{Path: filepath.Join(workspacePath, "project1", "worktree1"), Branch: "main", Commit: "abc123", Clean: true},
					{Path: filepath.Join(workspacePath, "project1", "worktree2"), Branch: "feature", Commit: "def456", Clean: false},
				}, nil)

				// Mock worktree status calls for analysis
				m.On("GetWorktreeStatus", filepath.Join(workspacePath, "project1", "worktree1")).Return(
					&types.WorktreeInfo{
						Path:   filepath.Join(workspacePath, "project1", "worktree1"),
						Branch: "main",
						Commit: "abc123",
						Clean:  true,
					}, nil)
				m.On("GetWorktreeStatus", filepath.Join(workspacePath, "project1", "worktree2")).Return(
					&types.WorktreeInfo{
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
				require.NoError(t, os.MkdirAll(workspacePath, 0755))
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, worktrees, tt.expectedCount)
			}

			mockGit.AssertExpectations(t)
		})
	}
}

func TestDiscoveryService_AnalyzeWorktree(t *testing.T) {
	tests := []struct {
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
					&types.WorktreeInfo{
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
					(*types.WorktreeInfo)(nil),
					assert.AnError)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockGit := &MockGitClient{}
			service := NewDiscoveryService(mockGit)
			tt.setupMocks(mockGit)

			// Test
			worktree, err := service.AnalyzeWorktree(tt.path)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, worktree)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, worktree)
				assert.Equal(t, tt.path, worktree.Path)
			}

			mockGit.AssertExpectations(t)
		})
	}
}

func TestDiscoveryService_DiscoverProjects(t *testing.T) {
	tests := []struct {
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
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "project1"), 0755))
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "project2"), 0755))
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "not-a-project"), 0755))

				// Mock main repository detection
				m.On("IsMainRepository", filepath.Join(workspacePath, "project1")).Return(true, nil)
				m.On("IsMainRepository", filepath.Join(workspacePath, "project2")).Return(true, nil)
				m.On("IsMainRepository", filepath.Join(workspacePath, "not-a-project")).Return(false, nil)

				// Mock worktree listing for projects
				m.On("ListWorktrees", filepath.Join(workspacePath, "project1")).Return([]*types.WorktreeInfo{
					{Path: filepath.Join(workspacePath, "project1"), Branch: "main", Commit: "abc123", Clean: true},
				}, nil)
				m.On("ListWorktrees", filepath.Join(workspacePath, "project2")).Return([]*types.WorktreeInfo{
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
				require.NoError(t, os.MkdirAll(filepath.Join(workspacePath, "regular-dir"), 0755))
				m.On("IsMainRepository", filepath.Join(workspacePath, "regular-dir")).Return(false, nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, projects, tt.expectedCount)
			}

			mockGit.AssertExpectations(t)
		})
	}
}

func TestDiscoveryService_Performance(t *testing.T) {
	t.Run("should handle concurrent discovery efficiently", func(t *testing.T) {
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
			require.NoError(t, os.MkdirAll(projectPath, 0755))

			// Mock each as main repository
			mockGit.On("IsMainRepository", projectPath).Return(true, nil)
			mockGit.On("ListWorktrees", projectPath).Return([]*types.WorktreeInfo{
				{Path: projectPath, Branch: "main", Commit: "abc123", Clean: true},
			}, nil)
		}

		// Test with timing
		start := time.Now()
		projects, err := service.DiscoverProjects(workspacePath)
		duration := time.Since(start)

		// Assert
		require.NoError(t, err)
		assert.Len(t, projects, projectCount)
		// Should complete quickly with concurrency
		assert.Less(t, duration, time.Second, "Discovery should complete quickly with concurrent processing")

		mockGit.AssertExpectations(t)
	})
}
