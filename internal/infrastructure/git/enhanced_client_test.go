package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/test/helpers"
	"github.com/stretchr/testify/suite"
)

// EnhancedGitClientTestSuite provides hybrid suite setup for enhanced git client tests
type EnhancedGitClientTestSuite struct {
	suite.Suite
	Client  *Client
	TempDir string
	Cleanup func()
}

// SetupTest initializes infrastructure components for each test
func (s *EnhancedGitClientTestSuite) SetupTest() {
	s.Client = NewClient()
	s.TempDir = s.T().TempDir()
	s.Cleanup = func() {
		_ = os.RemoveAll(s.TempDir)
	}
}

// TearDownTest cleans up infrastructure test resources
func (s *EnhancedGitClientTestSuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

// Helper function to create a test git repository
func (s *EnhancedGitClientTestSuite) createTestGitRepo() string {
	repo := helpers.NewGitRepo(s.T(), "twiggit-git-test-*")
	s.T().Cleanup(repo.Cleanup)
	return repo.Path
}

// Helper function to create test branches
func (s *EnhancedGitClientTestSuite) createTestBranches(repoPath string, branches []string) {
	for _, branch := range branches {
		cmd := exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = repoPath
		s.Require().NoError(cmd.Run(), "Failed to create branch %s", branch)

		// Make a small change and commit
		testFile := filepath.Join(repoPath, branch+".txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		s.Require().NoError(err)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = repoPath
		s.Require().NoError(cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "Add "+branch+".txt")
		cmd.Dir = repoPath
		s.Require().NoError(cmd.Run())
	}

	// Switch back to main branch
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// Try master if main doesn't exist
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		s.Require().NoError(cmd.Run())
	}
}

// TestEnhancedGitClient_GetRepositoryRoot tests repository root detection with table-driven approach
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_GetRepositoryRoot() {
	testCases := []struct {
		name        string
		setupPath   func() string
		expectError bool
	}{
		{
			name: "should return root for repository root",
			setupPath: func() string {
				return s.createTestGitRepo()
			},
			expectError: false,
		},
		{
			name: "should return root for subdirectory in repository",
			setupPath: func() string {
				repoPath := s.createTestGitRepo()
				subDir := filepath.Join(repoPath, "subdir", "nested")
				s.Require().NoError(os.MkdirAll(subDir, 0755))
				return subDir
			},
			expectError: false,
		},
		{
			name: "should return error for non-repository path",
			setupPath: func() string {
				tempDir, _ := os.MkdirTemp("", "non-repo-*")
				return tempDir
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			path := tt.setupPath()
			defer func() { _ = os.RemoveAll(path) }()

			root, err := s.Client.GetRepositoryRoot(path)

			if tt.expectError {
				s.Error(err)
				s.Empty(root)
			} else {
				s.NoError(err)
				s.NotEmpty(root)
				// Verify the root is actually a git repository
				isRepo, _ := s.Client.IsGitRepository(root)
				s.True(isRepo)
			}
		})
	}
}

// TestEnhancedGitClient_GetCurrentBranch tests current branch detection with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_GetCurrentBranch() {
	s.Run("should return current branch name", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		branch, err := s.Client.GetCurrentBranch(repoPath)

		s.NoError(err)
		// Should be either "main" or "master" depending on git version
		s.Contains([]string{"main", "master"}, branch)
	})

	s.Run("should return error for non-repository", func() {
		tempDir, cleanup := helpers.TempDir(s.T(), "non-repo-*")
		defer cleanup()

		_, err := s.Client.GetCurrentBranch(tempDir)
		s.Error(err)
	})

	s.Run("should return error for empty path", func() {
		_, err := s.Client.GetCurrentBranch("")
		s.Error(err)
	})
}

// TestEnhancedGitClient_GetAllBranches tests branch listing with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_GetAllBranches() {
	s.Run("should return all local branches", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branches
		testBranches := []string{"feature-1", "feature-2", "bugfix"}
		s.createTestBranches(repoPath, testBranches)

		branches, err := s.Client.GetAllBranches(repoPath)

		s.NoError(err)
		s.GreaterOrEqual(len(branches), 4) // main/master + 3 test branches

		// Check that our test branches are included
		branchMap := make(map[string]bool)
		for _, branch := range branches {
			branchMap[branch] = true
		}

		for _, testBranch := range testBranches {
			s.True(branchMap[testBranch], "Branch %s should be in the list", testBranch)
		}
	})

	s.Run("should return error for non-repository", func() {
		tempDir, cleanup := helpers.TempDir(s.T(), "non-repo-*")
		defer cleanup()

		_, err := s.Client.GetAllBranches(tempDir)
		s.Error(err)
	})
}

// TestEnhancedGitClient_GetRemoteBranches tests remote branch listing with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_GetRemoteBranches() {
	s.Run("should return empty list for repository with no remotes", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		branches, err := s.Client.GetRemoteBranches(repoPath)

		s.NoError(err)
		s.Empty(branches)
	})

	s.Run("should return error for non-repository", func() {
		tempDir, cleanup := helpers.TempDir(s.T(), "non-repo-*")
		defer cleanup()

		_, err := s.Client.GetRemoteBranches(tempDir)
		s.Error(err)
	})
}

// TestEnhancedGitClient_BranchExists tests branch existence checking with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_BranchExists() {
	s.Run("should return true for existing branch", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branch
		s.createTestBranches(repoPath, []string{"test-branch"})

		exists := s.Client.BranchExists(repoPath, "test-branch")
		s.True(exists)
	})

	s.Run("should return false for non-existing branch", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		exists := s.Client.BranchExists(repoPath, "non-existing-branch")
		s.False(exists)
	})

	s.Run("should return false for non-repository", func() {
		tempDir, cleanup := helpers.TempDir(s.T(), "non-repo-*")
		defer cleanup()

		exists := s.Client.BranchExists(tempDir, "any-branch")
		s.False(exists)
	})
}

// TestEnhancedGitClient_HasUncommittedChanges tests uncommitted changes detection with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_HasUncommittedChanges() {
	s.Run("should return false for clean repository", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		hasChanges := s.Client.HasUncommittedChanges(repoPath)
		s.False(hasChanges)
	})

	s.Run("should return true for repository with uncommitted changes", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Make uncommitted changes
		testFile := filepath.Join(repoPath, "new-file.txt")
		err := os.WriteFile(testFile, []byte("new content"), 0644)
		s.Require().NoError(err)

		hasChanges := s.Client.HasUncommittedChanges(repoPath)
		s.True(hasChanges)
	})

	s.Run("should return true for repository with modified files", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Modify existing file
		testFile := filepath.Join(repoPath, "README.md")
		err := os.WriteFile(testFile, []byte("# Modified README\n"), 0644)
		s.Require().NoError(err)

		hasChanges := s.Client.HasUncommittedChanges(repoPath)
		s.True(hasChanges)
	})

	s.Run("should return false for non-repository", func() {
		tempDir, cleanup := helpers.TempDir(s.T(), "non-repo-*")
		defer cleanup()

		hasChanges := s.Client.HasUncommittedChanges(tempDir)
		s.False(hasChanges)
	})
}

// TestEnhancedGitClient_Integration tests complete repository analysis with sub-tests
func (s *EnhancedGitClientTestSuite) TestEnhancedGitClient_Integration() {
	s.Run("should work together for complete repository analysis", func() {
		repoPath := s.createTestGitRepo()
		defer func() { _ = os.RemoveAll(repoPath) }()

		// Create test branches
		testBranches := []string{"feature-a", "feature-b"}
		s.createTestBranches(repoPath, testBranches)

		// Test repository detection
		isRepo, err := s.Client.IsGitRepository(repoPath)
		s.Require().NoError(err)
		s.True(isRepo)

		// Test repository root
		root, err := s.Client.GetRepositoryRoot(repoPath)
		s.NoError(err)
		s.Equal(repoPath, root)

		// Test current branch
		currentBranch, err := s.Client.GetCurrentBranch(repoPath)
		s.NoError(err)
		s.NotEmpty(currentBranch)

		// Test all branches
		allBranches, err := s.Client.GetAllBranches(repoPath)
		s.NoError(err)
		s.GreaterOrEqual(len(allBranches), 3) // main/master + 2 test branches

		// Test branch existence
		for _, branch := range testBranches {
			exists := s.Client.BranchExists(repoPath, branch)
			s.True(exists, "Branch %s should exist", branch)
		}

		// Test uncommitted changes (should be clean)
		hasChanges := s.Client.HasUncommittedChanges(repoPath)
		s.False(hasChanges)

		// Create worktree and test it
		worktreePath := filepath.Join(repoPath, "worktrees", "feature-a-wt")
		err = s.Client.CreateWorktree(repoPath, "feature-a", worktreePath)
		s.NoError(err)

		// Verify worktree was created
		worktrees, err := s.Client.ListWorktrees(repoPath)
		s.NoError(err)
		s.GreaterOrEqual(len(worktrees), 2) // main repo + new worktree

		// Clean up
		err = s.Client.RemoveWorktree(repoPath, worktreePath, false)
		s.NoError(err)
	})
}

// TestEnhancedGitClientSuite runs the enhanced git client test suite
func TestEnhancedGitClientSuite(t *testing.T) {
	suite.Run(t, new(EnhancedGitClientTestSuite))
}
