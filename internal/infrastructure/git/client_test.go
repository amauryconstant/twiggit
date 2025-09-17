package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/suite"
)

// GitClientContextTestSuite tests context-aware git operations
type GitClientContextTestSuite struct {
	suite.Suite
	Client  *Client
	TempDir string
	Cleanup func()
}

// SetupTest initializes infrastructure components for each test
func (s *GitClientContextTestSuite) SetupTest() {
	s.Client = NewClient()
	s.TempDir = s.T().TempDir()
	s.Cleanup = func() {
		_ = os.RemoveAll(s.TempDir)
	}
}

// TearDownTest cleans up infrastructure test resources
func (s *GitClientContextTestSuite) TearDownTest() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

// setupRepositoryWithInitialCommit creates a git repository with an initial commit.
// Returns the repository path and the worktree for further operations.
func (s *GitClientContextTestSuite) setupRepositoryWithInitialCommit(repoName string) (string, *git.Worktree) {
	s.T().Helper()
	repoPath := filepath.Join(s.TempDir, repoName)
	repo, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	wt, err := repo.Worktree()
	s.Require().NoError(err)

	// Create a test file
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("initial content"), 0644)
	s.Require().NoError(err)

	// Add and commit the file
	_, err = wt.Add("test.txt")
	s.Require().NoError(err)

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	s.Require().NoError(err)

	return repoPath, wt
}

// createAndCheckoutBranch creates a new branch and checks it out.
// Returns an error if the operation fails.
func (s *GitClientContextTestSuite) createAndCheckoutBranch(wt *git.Worktree, branchName string) {
	s.T().Helper()
	err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})
	s.Require().NoError(err)
}

// TestIsGitRepository_WithContext tests context support in IsGitRepository
func (s *GitClientContextTestSuite) TestIsGitRepository_WithContext() {
	ctx := context.Background()

	// Setup a git repository
	gitDir := filepath.Join(s.TempDir, "git-repo")
	_, err := git.PlainInit(gitDir, false)
	s.Require().NoError(err)

	// This should fail because IsGitRepository doesn't accept context yet
	isRepo, err := s.Client.IsGitRepository(ctx, gitDir)

	s.Require().NoError(err)
	s.True(isRepo)
}

// TestIsGitRepository_ContextCancellation tests context cancellation handling
func (s *GitClientContextTestSuite) TestIsGitRepository_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	gitDir := filepath.Join(s.TempDir, "git-repo")
	_, err := git.PlainInit(gitDir, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	_, err = s.Client.IsGitRepository(ctx, gitDir)
	s.Require().Error(err)
	s.Contains(err.Error(), "context canceled")
}

// TestListWorktrees_ContextTimeout tests context timeout handling
func (s *GitClientContextTestSuite) TestListWorktrees_ContextTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give context time to timeout
	time.Sleep(10 * time.Millisecond)

	gitDir := filepath.Join(s.TempDir, "git-repo")
	_, err := git.PlainInit(gitDir, false)
	s.Require().NoError(err)

	// This should fail with context deadline exceeded
	_, err = s.Client.ListWorktrees(ctx, gitDir)
	s.Require().Error(err)
	s.Contains(err.Error(), "context deadline exceeded")
}

// TestListWorktrees_UsingGoGit tests that ListWorktrees uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestListWorktrees_UsingGoGit() {
	ctx := context.Background()

	// Setup main repository with initial commit
	mainRepo, _ := s.setupRepositoryWithInitialCommit("main")

	// This test expects the implementation to use go-git internally
	worktrees, err := s.Client.ListWorktrees(ctx, mainRepo)
	s.Require().NoError(err)
	s.Len(worktrees, 1) // Should include main worktree
	s.Equal(mainRepo, worktrees[0].Path)
}

// TestGetAllBranches_UsingGoGit tests that GetAllBranches uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestGetAllBranches_UsingGoGit() {
	ctx := context.Background()

	// Setup repository with multiple branches
	mainRepo, wt := s.setupRepositoryWithInitialCommit("main")

	// Create and checkout a new branch
	s.createAndCheckoutBranch(wt, "feature/test")

	// This test expects go-git based branch listing
	branches, err := s.Client.GetAllBranches(ctx, mainRepo)
	s.Require().NoError(err)
	s.Contains(branches, "master") // go-git default branch name
	s.Contains(branches, "feature/test")
}

// TestGetWorktreeStatus_UsingGoGit tests that GetWorktreeStatus uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestGetWorktreeStatus_UsingGoGit() {
	ctx := context.Background()

	// Setup repository
	mainRepo, wt := s.setupRepositoryWithInitialCommit("main")

	// Create another commit to have something to check status against
	testFile := filepath.Join(mainRepo, "status_test.txt")
	err := os.WriteFile(testFile, []byte("status test content"), 0644)
	s.Require().NoError(err)

	// Add and commit the file
	_, err = wt.Add("status_test.txt")
	s.Require().NoError(err)

	commitHash, err := wt.Commit("Status test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	s.Require().NoError(err)

	// This test expects go-git based status checking
	status, err := s.Client.GetWorktreeStatus(ctx, mainRepo)
	s.Require().NoError(err)
	s.Equal(mainRepo, status.Path)
	s.Equal("master", status.Branch) // go-git uses "master" as default
	s.Equal(commitHash.String(), status.Commit)
	s.True(status.Clean) // Should be clean after commit
}

// TestHasUncommittedChanges_UsingGoGit tests that HasUncommittedChanges uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestHasUncommittedChanges_UsingGoGit() {
	ctx := context.Background()

	// Setup repository
	mainRepo := filepath.Join(s.TempDir, "main")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Initially should be clean (no uncommitted changes)
	hasChanges := s.Client.HasUncommittedChanges(ctx, mainRepo)
	s.False(hasChanges)

	// Create an uncommitted file
	testFile := filepath.Join(mainRepo, "uncommitted.txt")
	err = os.WriteFile(testFile, []byte("uncommitted content"), 0644)
	s.Require().NoError(err)

	// Now should have uncommitted changes
	hasChanges = s.Client.HasUncommittedChanges(ctx, mainRepo)
	s.True(hasChanges)
}

// TestGetCurrentBranch_UsingGoGit tests that GetCurrentBranch uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestGetCurrentBranch_UsingGoGit() {
	ctx := context.Background()

	// Setup repository with initial commit
	mainRepo, wt := s.setupRepositoryWithInitialCommit("main")

	// Should be on master branch initially (go-git default)
	branch, err := s.Client.GetCurrentBranch(ctx, mainRepo)
	s.Require().NoError(err)
	s.Equal("master", branch)

	// Create and checkout a new branch
	s.createAndCheckoutBranch(wt, "feature/new-branch")

	// Should now be on the new branch
	branch, err = s.Client.GetCurrentBranch(ctx, mainRepo)
	s.Require().NoError(err)
	s.Equal("feature/new-branch", branch)
}

// TestCreateWorktree_UsingGoGit tests that CreateWorktree uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestCreateWorktree_UsingGoGit() {
	ctx := context.Background()

	// Setup main repository with initial commit
	mainRepo, _ := s.setupRepositoryWithInitialCommit("main")

	// Create a worktree path
	worktreePath := filepath.Join(s.TempDir, "worktree1")

	// This should create a worktree using go-git
	err := s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", worktreePath)
	s.Require().NoError(err)

	// Verify the worktree was created
	s.DirExists(worktreePath)

	// Verify it's a git repository
	isRepo, err := s.Client.IsGitRepository(ctx, worktreePath)
	s.Require().NoError(err)
	s.True(isRepo)
}

// TestRemoveWorktree_UsingGoGit tests that RemoveWorktree uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestRemoveWorktree_UsingGoGit() {
	ctx := context.Background()

	// Setup main repository with initial commit
	mainRepo, _ := s.setupRepositoryWithInitialCommit("main")

	// Create a worktree first
	worktreePath := filepath.Join(s.TempDir, "worktree1")
	err := s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", worktreePath)
	s.Require().NoError(err)

	// Now remove it using go-git
	err = s.Client.RemoveWorktree(ctx, mainRepo, worktreePath, false)
	s.Require().NoError(err)

	// Verify the worktree directory was removed
	s.NoDirExists(worktreePath)
}

// TestGetRepositoryRoot_UsingGoGit tests that GetRepositoryRoot uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestGetRepositoryRoot_UsingGoGit() {
	ctx := context.Background()

	// Setup repository
	mainRepo := filepath.Join(s.TempDir, "main")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Test from within the repository
	root, err := s.Client.GetRepositoryRoot(ctx, mainRepo)
	s.Require().NoError(err)
	s.Equal(mainRepo, root)

	// Test from a subdirectory
	subDir := filepath.Join(mainRepo, "subdir")
	err = os.Mkdir(subDir, 0755)
	s.Require().NoError(err)

	root, err = s.Client.GetRepositoryRoot(ctx, subDir)
	s.Require().NoError(err)
	s.Equal(mainRepo, root)
}

// TestGetRemoteBranches_UsingGoGit tests that GetRemoteBranches uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestGetRemoteBranches_UsingGoGit() {
	ctx := context.Background()

	// Setup repository
	mainRepo := filepath.Join(s.TempDir, "main")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Initially no remote branches (should return empty slice, not error)
	remoteBranches, err := s.Client.GetRemoteBranches(ctx, mainRepo)
	s.Require().NoError(err)
	s.Empty(remoteBranches)
}

// TestBranchExists_UsingGoGit tests that BranchExists uses go-git instead of CLI
func (s *GitClientContextTestSuite) TestBranchExists_UsingGoGit() {
	ctx := context.Background()

	// Setup repository
	mainRepo, _ := s.setupRepositoryWithInitialCommit("main")

	// Main branch should exist (go-git uses "master" by default)
	exists := s.Client.BranchExists(ctx, mainRepo, "master")
	s.True(exists)

	// Non-existent branch should not exist
	exists = s.Client.BranchExists(ctx, mainRepo, "non-existent")
	s.False(exists)
}

// TestIsMainRepository_WithContext tests context support in IsMainRepository
func (s *GitClientContextTestSuite) TestIsMainRepository_WithContext() {
	ctx := context.Background()

	// Setup a main git repository
	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// This should fail because IsMainRepository doesn't accept context yet
	isMain, err := s.Client.IsMainRepository(ctx, mainRepo)

	s.Require().NoError(err)
	s.True(isMain) // Should be main repository
}

// TestIsMainRepository_ContextCancellation tests context cancellation in IsMainRepository
func (s *GitClientContextTestSuite) TestIsMainRepository_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	_, err = s.Client.IsMainRepository(ctx, mainRepo)
	s.Require().Error(err)
	s.Contains(err.Error(), "context canceled")
}

func TestGitClientContextTestSuite(t *testing.T) {
	suite.Run(t, new(GitClientContextTestSuite))
}
