package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/suite"
)

// GitClientTestSuite provides hybrid suite setup for git client tests
type GitClientTestSuite struct {
	suite.Suite
	Client  *Client
	TempDir string
	Cleanup func()
}

// SetupTest initializes infrastructure components for each test
func (s *GitClientTestSuite) SetupTest() {
	s.Client = NewClient()
	s.TempDir = s.T().TempDir()
	s.Cleanup = func() {
		_ = os.RemoveAll(s.TempDir)
	}
}

// TearDownTest cleans up infrastructure test resources
func (s *GitClientTestSuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

// TestGitClient_NewClient tests client creation
func (s *GitClientTestSuite) TestGitClient_NewClient() {
	client := NewClient()
	s.Require().NotNil(client)
}

// TestGitClient_IsGitRepository tests repository validation with table-driven approach
func (s *GitClientTestSuite) TestGitClient_IsGitRepository() {
	testCases := []struct {
		name         string
		setup        func() string
		expectRepo   bool
		expectError  bool
		errorMessage string
	}{
		{
			name: "should return true for valid git repository",
			setup: func() string {
				gitDir := filepath.Join(s.TempDir, "git-repo")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)
				return gitDir
			},
			expectRepo:  true,
			expectError: false,
		},
		{
			name: "should return false for non-git directory",
			setup: func() string {
				nonGitDir := filepath.Join(s.TempDir, "non-git")
				err := os.Mkdir(nonGitDir, 0755)
				s.Require().NoError(err)
				return nonGitDir
			},
			expectRepo:  false,
			expectError: false,
		},
		{
			name: "should return error for non-existent path",
			setup: func() string {
				return filepath.Join(s.TempDir, "does-not-exist")
			},
			expectRepo:   false,
			expectError:  true,
			errorMessage: "does not exist",
		},
		{
			name: "should return error for empty path",
			setup: func() string {
				return ""
			},
			expectRepo:   false,
			expectError:  true,
			errorMessage: "path cannot be empty",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			path := tt.setup()
			isRepo, err := s.Client.IsGitRepository(path)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
			} else {
				s.Require().NoError(err)
			}

			s.Equal(tt.expectRepo, isRepo)
		})
	}
}

// TestGitClient_IsMainRepository tests main repository detection with table-driven approach
func (s *GitClientTestSuite) TestGitClient_IsMainRepository() {
	testCases := []struct {
		name         string
		setup        func() (string, func())
		expectMain   bool
		expectError  bool
		errorMessage string
	}{
		{
			name: "should return true for main repository",
			setup: func() (string, func()) {
				gitDir := filepath.Join(s.TempDir, "main-repo")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)
				return gitDir, func() {}
			},
			expectMain:  true,
			expectError: false,
		},
		{
			name: "should return false for worktree",
			setup: func() (string, func()) {
				// Create main repository
				mainDir := filepath.Join(s.TempDir, "main-for-worktree")
				_, err := git.PlainInit(mainDir, false)
				s.Require().NoError(err)

				// Create a worktree
				worktreeDir := filepath.Join(s.TempDir, "test-worktree")
				err = s.Client.CreateWorktree(mainDir, "new-branch", worktreeDir)
				s.Require().NoError(err)

				return worktreeDir, func() {
					_ = os.RemoveAll(worktreeDir)
				}
			},
			expectMain:  false,
			expectError: false,
		},
		{
			name: "should return false for non-git directory",
			setup: func() (string, func()) {
				nonGitDir := filepath.Join(s.TempDir, "not-git")
				err := os.MkdirAll(nonGitDir, 0755)
				s.Require().NoError(err)
				return nonGitDir, func() {
					_ = os.RemoveAll(nonGitDir)
				}
			},
			expectMain:  false,
			expectError: false,
		},
		{
			name: "should return error for empty path",
			setup: func() (string, func()) {
				return "", func() {}
			},
			expectMain:   false,
			expectError:  true,
			errorMessage: "path cannot be empty",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			path, cleanup := tt.setup()
			defer cleanup()

			isMain, err := s.Client.IsMainRepository(path)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
			} else {
				s.Require().NoError(err)
			}

			s.Equal(tt.expectMain, isMain)
		})
	}
}

// TestGitClient_ListWorktrees tests worktree listing with table-driven approach
func (s *GitClientTestSuite) TestGitClient_ListWorktrees() {
	testCases := []struct {
		name         string
		setup        func() (string, *Client)
		expectError  bool
		expectCount  int
		errorMessage string
	}{
		{
			name: "should return main repository for repository with no worktrees",
			setup: func() (string, *Client) {
				gitDir := filepath.Join(s.TempDir, "main-repo")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)
				return gitDir, s.Client
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "should return error for non-git directory",
			setup: func() (string, *Client) {
				nonGitDir := filepath.Join(s.TempDir, "non-git")
				err := os.Mkdir(nonGitDir, 0755)
				s.Require().NoError(err)
				return nonGitDir, s.Client
			},
			expectError: true,
			expectCount: 0,
		},
		{
			name: "should return error for empty repository path",
			setup: func() (string, *Client) {
				return "", s.Client
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			path, client := tt.setup()

			worktrees, err := client.ListWorktrees(path)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
				s.Nil(worktrees)
			} else {
				s.Require().NoError(err)
				s.NotNil(worktrees)
				s.Len(worktrees, tt.expectCount)
				if tt.expectCount > 0 {
					s.Equal(path, worktrees[0].Path)
					s.NotEmpty(worktrees[0].Branch)
					s.NotEmpty(worktrees[0].Commit)
				}
			}
		})
	}
}

// TestGitClient_CreateWorktree tests worktree creation with table-driven approach
func (s *GitClientTestSuite) TestGitClient_CreateWorktree() {
	testCases := []struct {
		name         string
		setup        func() (string, string, string)
		expectError  bool
		errorMessage string
	}{
		{
			name: "should create worktree from existing branch",
			setup: func() (string, string, string) {
				// Create a git repository with initial commit
				gitDir := filepath.Join(s.TempDir, "main-repo-create")
				repo, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				// Create initial commit
				wt, err := repo.Worktree()
				s.Require().NoError(err)

				// Create a test file
				testFile := filepath.Join(gitDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0644)
				s.Require().NoError(err)

				_, err = wt.Add("test.txt")
				s.Require().NoError(err)

				_, err = wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
				})
				s.Require().NoError(err)

				// Get current branch name (could be master or main depending on git config)
				head, err := repo.Head()
				s.Require().NoError(err)
				branchName := head.Name().Short()

				// Create a new branch for the worktree
				branchRef := plumbing.ReferenceName("refs/heads/" + branchName + "-worktree")
				err = wt.Checkout(&git.CheckoutOptions{
					Branch: branchRef,
					Create: true,
				})
				s.Require().NoError(err)

				// Switch back to main branch
				mainRef := plumbing.ReferenceName("refs/heads/" + branchName)
				err = wt.Checkout(&git.CheckoutOptions{
					Branch: mainRef,
				})
				s.Require().NoError(err)

				// Create worktree from the new branch
				worktreePath := filepath.Join(s.TempDir, "worktree-1")
				return gitDir, branchName + "-worktree", worktreePath
			},
			expectError: false,
		},
		{
			name: "should return error for empty repository path",
			setup: func() (string, string, string) {
				worktreePath := filepath.Join(s.TempDir, "worktree-1")
				return "", "main", worktreePath
			},
			expectError: true,
		},
		{
			name: "should return error for empty branch name",
			setup: func() (string, string, string) {
				gitDir := filepath.Join(s.TempDir, "main-repo-empty-branch")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				worktreePath := filepath.Join(s.TempDir, "worktree-empty-branch")
				return gitDir, "", worktreePath
			},
			expectError: true,
		},
		{
			name: "should return error for empty target path",
			setup: func() (string, string, string) {
				gitDir := filepath.Join(s.TempDir, "main-repo-empty-target")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				return gitDir, "main", ""
			},
			expectError: true,
		},
		{
			name: "should return error for existing target path",
			setup: func() (string, string, string) {
				gitDir := filepath.Join(s.TempDir, "main-repo-existing-target")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				worktreePath := filepath.Join(s.TempDir, "existing-dir")
				err = os.Mkdir(worktreePath, 0755)
				s.Require().NoError(err)

				return gitDir, "main", worktreePath
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			repoPath, branch, targetPath := tt.setup()

			err := s.Client.CreateWorktree(repoPath, branch, targetPath)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				// Verify worktree was created
				s.DirExists(targetPath)
			}
		})
	}
}

// TestGitClient_GetWorktreeStatus tests worktree status retrieval with table-driven approach
func (s *GitClientTestSuite) TestGitClient_GetWorktreeStatus() {
	testCases := []struct {
		name         string
		setup        func() string
		expectError  bool
		expectClean  bool
		errorMessage string
	}{
		{
			name: "should return status for clean worktree",
			setup: func() string {
				// Create a git repository
				gitDir := filepath.Join(s.TempDir, "main-repo")
				repo, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				// Create initial commit
				wt, err := repo.Worktree()
				s.Require().NoError(err)

				testFile := filepath.Join(gitDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0644)
				s.Require().NoError(err)

				_, err = wt.Add("test.txt")
				s.Require().NoError(err)

				_, err = wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
				})
				s.Require().NoError(err)

				return gitDir
			},
			expectError: false,
			expectClean: true,
		},
		{
			name: "should return error for non-existent worktree path",
			setup: func() string {
				return filepath.Join(s.TempDir, "does-not-exist")
			},
			expectError: true,
		},
		{
			name: "should return error for empty worktree path",
			setup: func() string {
				return ""
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			path := tt.setup()

			status, err := s.Client.GetWorktreeStatus(path)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
				s.Nil(status)
			} else {
				s.Require().NoError(err)
				s.NotNil(status)
				s.Equal(tt.expectClean, status.Clean)
				s.Equal(path, status.Path)
				s.NotEmpty(status.Branch)
				s.NotEmpty(status.Commit)
			}
		})
	}
}

// TestGitClient_RemoveWorktree tests worktree removal with table-driven approach
func (s *GitClientTestSuite) TestGitClient_RemoveWorktree() {
	testCases := []struct {
		name         string
		setup        func() (string, string)
		expectError  bool
		errorMessage string
	}{
		{
			name: "should remove existing worktree",
			setup: func() (string, string) {
				// Create a git repository
				gitDir := filepath.Join(s.TempDir, "main-repo-remove")
				repo, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				// Create initial commit
				wt, err := repo.Worktree()
				s.Require().NoError(err)

				testFile := filepath.Join(gitDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0644)
				s.Require().NoError(err)

				_, err = wt.Add("test.txt")
				s.Require().NoError(err)

				_, err = wt.Commit("Initial commit", &git.CommitOptions{
					Author: &object.Signature{Name: "Test Author", Email: "test@example.com"},
				})
				s.Require().NoError(err)

				// Get current branch name
				head, err := repo.Head()
				s.Require().NoError(err)
				branchName := head.Name().Short()

				// Create a new branch for the worktree
				branchRef := plumbing.ReferenceName("refs/heads/" + branchName + "-remove")
				err = wt.Checkout(&git.CheckoutOptions{
					Branch: branchRef,
					Create: true,
				})
				s.Require().NoError(err)

				// Switch back to main branch
				mainRef := plumbing.ReferenceName("refs/heads/" + branchName)
				err = wt.Checkout(&git.CheckoutOptions{
					Branch: mainRef,
				})
				s.Require().NoError(err)

				// Create worktree from the new branch
				worktreePath := filepath.Join(s.TempDir, "worktree-remove")
				err = s.Client.CreateWorktree(gitDir, branchName+"-remove", worktreePath)
				s.Require().NoError(err)

				return gitDir, worktreePath
			},
			expectError: false,
		},
		{
			name: "should return error for empty repository path",
			setup: func() (string, string) {
				worktreePath := filepath.Join(s.TempDir, "worktree-1")
				return "", worktreePath
			},
			expectError: true,
		},
		{
			name: "should return error for empty worktree path",
			setup: func() (string, string) {
				gitDir := filepath.Join(s.TempDir, "main-repo-empty-worktree")
				_, err := git.PlainInit(gitDir, false)
				s.Require().NoError(err)

				return gitDir, ""
			},
			expectError: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			repoPath, worktreePath := tt.setup()

			err := s.Client.RemoveWorktree(repoPath, worktreePath, false)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMessage != "" {
					s.Contains(err.Error(), tt.errorMessage)
				}
			} else {
				s.Require().NoError(err)
				// Verify worktree directory was removed (git worktree remove removes the directory by default)
				s.NoDirExists(worktreePath)
			}
		})
	}
}

// TestWorktreeInfo_Validation tests worktree info validation with table-driven approach
func (s *GitClientTestSuite) TestWorktreeInfo_Validation() {
	testCases := []struct {
		name        string
		worktree    domain.WorktreeInfo
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid worktree info",
			worktree: domain.WorktreeInfo{
				Path:   "/valid/path",
				Branch: "main",
				Commit: "abc123",
				Clean:  true,
			},
			expectError: false,
		},
		{
			name: "empty path",
			worktree: domain.WorktreeInfo{
				Path:   "",
				Branch: "main",
				Commit: "abc123",
			},
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name: "empty branch",
			worktree: domain.WorktreeInfo{
				Path:   "/valid/path",
				Branch: "",
				Commit: "abc123",
			},
			expectError: true,
			errorMsg:    "branch cannot be empty",
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			err := tt.worktree.Validate()

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorMsg != "" {
					s.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestGitClientSuite runs the git client test suite
func TestGitClientSuite(t *testing.T) {
	suite.Run(t, new(GitClientTestSuite))
}
