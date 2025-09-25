package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaury/twiggit/internal/domain"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
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
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
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

// TestCheckCandidateWorktree_ErrorHandling tests error handling in checkCandidateWorktree
func (s *GitClientContextTestSuite) TestCheckCandidateWorktree_ErrorHandling() {
	ctx := context.Background()

	// Setup main repository
	mainRepoPath, _ := s.setupRepositoryWithInitialCommit("main-repo")

	// Test with a non-existent candidate path
	nonExistentPath := filepath.Join(s.TempDir, "non-existent-candidate")

	// This should return nil, nil (not a git repository)
	worktreeInfo, err := s.Client.checkCandidateWorktree(ctx, nonExistentPath, mainRepoPath)
	s.Require().NoError(err)
	s.Nil(worktreeInfo)
}

// TestCheckCandidateWorktree_GetWorktreeStatusError tests error handling when GetWorktreeStatus fails
func (s *GitClientContextTestSuite) TestCheckCandidateWorktree_GetWorktreeStatusError() {
	ctx := context.Background()

	// Setup main repository
	mainRepoPath, _ := s.setupRepositoryWithInitialCommit("main-repo")

	// Create a candidate directory that is a git repository but has issues
	candidatePath := filepath.Join(s.TempDir, "candidate-repo")
	_, err := git.PlainInit(candidatePath, false)
	s.Require().NoError(err)

	// Add origin remote pointing to main repo to make it pass the hasOriginRemote check
	repo, err := git.PlainOpen(candidatePath)
	s.Require().NoError(err)

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{mainRepoPath},
	})
	s.Require().NoError(err)

	// Now test checkCandidateWorktree - it should fail at GetWorktreeStatus
	// because the candidate repository doesn't have any commits
	worktreeInfo, err := s.Client.checkCandidateWorktree(ctx, candidatePath, mainRepoPath)
	s.Require().Error(err) // Should fail due to no commits in candidate repo
	s.Nil(worktreeInfo)
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

// TestGetAllBranches_ContextCancellation tests context cancellation in GetAllBranches
func (s *GitClientContextTestSuite) TestGetAllBranches_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	branches, err := s.Client.GetAllBranches(ctx, repoPath)
	s.Require().Error(err)
	s.Nil(branches)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestGetAllBranches_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestGetAllBranches_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	branches, err := s.Client.GetAllBranches(ctx, "")
	s.Require().Error(err)
	s.Nil(branches)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	branches, err = s.Client.GetAllBranches(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Nil(branches)
	s.True(domain.IsDomainErrorType(err, domain.ErrGitCommand))
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

// TestGetWorktreeStatus_ContextCancellation tests context cancellation in GetWorktreeStatus
func (s *GitClientContextTestSuite) TestGetWorktreeStatus_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	status, err := s.Client.GetWorktreeStatus(ctx, repoPath)
	s.Require().Error(err)
	s.Nil(status)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestGetWorktreeStatus_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestGetWorktreeStatus_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	status, err := s.Client.GetWorktreeStatus(ctx, "")
	s.Require().Error(err)
	s.Nil(status)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	status, err = s.Client.GetWorktreeStatus(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Nil(status)
	s.True(domain.IsDomainErrorType(err, domain.ErrWorktreeNotFound))
}

// TestGetWorktreeStatus_NonGitRepository tests error handling for non-git repository
func (s *GitClientContextTestSuite) TestGetWorktreeStatus_NonGitRepository() {
	ctx := context.Background()

	// Create a directory that's not a git repository
	nonGitDir := filepath.Join(s.TempDir, "non-git-repo")
	err := os.MkdirAll(nonGitDir, 0755)
	s.Require().NoError(err)

	// This should fail with not a repository error
	status, err := s.Client.GetWorktreeStatus(ctx, nonGitDir)
	s.Require().Error(err)
	s.Nil(status)
	s.True(domain.IsDomainErrorType(err, domain.ErrNotRepository))
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

// TestHasUncommittedChanges_ContextCancellation tests context cancellation in HasUncommittedChanges
func (s *GitClientContextTestSuite) TestHasUncommittedChanges_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should return false (not error) for context cancellation
	hasChanges := s.Client.HasUncommittedChanges(ctx, repoPath)
	s.False(hasChanges)
}

// TestHasUncommittedChanges_EmptyPath tests behavior with empty path
func (s *GitClientContextTestSuite) TestHasUncommittedChanges_EmptyPath() {
	ctx := context.Background()

	// Test with empty path
	hasChanges := s.Client.HasUncommittedChanges(ctx, "")
	s.False(hasChanges)
}

// TestHasUncommittedChanges_NonGitRepository tests behavior with non-git repository
func (s *GitClientContextTestSuite) TestHasUncommittedChanges_NonGitRepository() {
	ctx := context.Background()

	// Create a directory that's not a git repository
	nonGitDir := filepath.Join(s.TempDir, "non-git-repo")
	err := os.MkdirAll(nonGitDir, 0755)
	s.Require().NoError(err)

	// This should return false (not error) for non-git repository
	hasChanges := s.Client.HasUncommittedChanges(ctx, nonGitDir)
	s.False(hasChanges)
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

// TestGetCurrentBranch_ContextCancellation tests context cancellation in GetCurrentBranch
func (s *GitClientContextTestSuite) TestGetCurrentBranch_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	branch, err := s.Client.GetCurrentBranch(ctx, repoPath)
	s.Require().Error(err)
	s.Empty(branch)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestGetCurrentBranch_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestGetCurrentBranch_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	branch, err := s.Client.GetCurrentBranch(ctx, "")
	s.Require().Error(err)
	s.Empty(branch)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	branch, err = s.Client.GetCurrentBranch(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Empty(branch)
	s.True(domain.IsDomainErrorType(err, domain.ErrGitCommand))
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

// TestCreateWorktree_ContextCancellation tests context cancellation in CreateWorktree
func (s *GitClientContextTestSuite) TestCreateWorktree_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	worktreePath := filepath.Join(s.TempDir, "worktree1")

	// This should fail with context canceled error
	err = s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", worktreePath)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestCreateWorktree_InvalidInputs tests error handling for invalid inputs
func (s *GitClientContextTestSuite) TestCreateWorktree_InvalidInputs() {
	ctx := context.Background()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	worktreePath := filepath.Join(s.TempDir, "worktree1")

	// Test with empty repo path
	err = s.Client.CreateWorktree(ctx, "", "feature/test-branch", worktreePath)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with empty branch name
	err = s.Client.CreateWorktree(ctx, mainRepo, "", worktreePath)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))

	// Test with empty target path
	err = s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", "")
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))
}

// TestCreateWorktree_TargetPathExists tests error handling when target path already exists
func (s *GitClientContextTestSuite) TestCreateWorktree_TargetPathExists() {
	ctx := context.Background()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Create a directory that already exists
	worktreePath := filepath.Join(s.TempDir, "existing-worktree")
	err = os.MkdirAll(worktreePath, 0755)
	s.Require().NoError(err)

	// This should fail with worktree exists error
	err = s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", worktreePath)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrWorktreeExists))
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

// TestRemoveWorktree_ContextCancellation tests context cancellation in RemoveWorktree
func (s *GitClientContextTestSuite) TestRemoveWorktree_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	worktreePath := filepath.Join(s.TempDir, "worktree1")

	// This should fail with context canceled error
	err = s.Client.RemoveWorktree(ctx, mainRepo, worktreePath, false)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestRemoveWorktree_NonExistentWorktree tests error handling for non-existent worktree
func (s *GitClientContextTestSuite) TestRemoveWorktree_NonExistentWorktree() {
	ctx := context.Background()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	nonExistentPath := filepath.Join(s.TempDir, "non-existent")

	// This should fail with worktree not found error
	err = s.Client.RemoveWorktree(ctx, mainRepo, nonExistentPath, false)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrWorktreeNotFound))
}

// TestRemoveWorktree_UncommittedChanges tests error handling for worktree with uncommitted changes
func (s *GitClientContextTestSuite) TestRemoveWorktree_UncommittedChanges() {
	ctx := context.Background()

	// Setup main repository with initial commit
	mainRepo, _ := s.setupRepositoryWithInitialCommit("main")

	// Create a worktree first
	worktreePath := filepath.Join(s.TempDir, "worktree1")
	err := s.Client.CreateWorktree(ctx, mainRepo, "feature/test-branch", worktreePath)
	s.Require().NoError(err)

	// Add uncommitted changes to the worktree
	uncommittedFile := filepath.Join(worktreePath, "uncommitted.txt")
	err = os.WriteFile(uncommittedFile, []byte("uncommitted content"), 0644)
	s.Require().NoError(err)

	// This should fail with uncommitted changes error
	err = s.Client.RemoveWorktree(ctx, mainRepo, worktreePath, false)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrUncommittedChanges))

	// Now remove with force flag - should succeed
	err = s.Client.RemoveWorktree(ctx, mainRepo, worktreePath, true)
	s.Require().NoError(err)
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

// TestGetRepositoryRoot_ContextCancellation tests context cancellation in GetRepositoryRoot
func (s *GitClientContextTestSuite) TestGetRepositoryRoot_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	root, err := s.Client.GetRepositoryRoot(ctx, repoPath)
	s.Require().Error(err)
	s.Empty(root)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestGetRepositoryRoot_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestGetRepositoryRoot_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	root, err := s.Client.GetRepositoryRoot(ctx, "")
	s.Require().Error(err)
	s.Empty(root)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	root, err = s.Client.GetRepositoryRoot(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Empty(root)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))
}

// TestGetRepositoryRoot_NonGitRepository tests error handling for non-git repository
func (s *GitClientContextTestSuite) TestGetRepositoryRoot_NonGitRepository() {
	ctx := context.Background()

	// Create a directory that's not a git repository
	nonGitDir := filepath.Join(s.TempDir, "non-git-repo")
	err := os.MkdirAll(nonGitDir, 0755)
	s.Require().NoError(err)

	// This should fail with not a repository error
	root, err := s.Client.GetRepositoryRoot(ctx, nonGitDir)
	s.Require().Error(err)
	s.Empty(root)
	s.True(domain.IsDomainErrorType(err, domain.ErrNotRepository))
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

// TestGetRemoteBranches_ContextCancellation tests context cancellation in GetRemoteBranches
func (s *GitClientContextTestSuite) TestGetRemoteBranches_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	mainRepo := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// This should fail with context canceled error
	_, err = s.Client.GetRemoteBranches(ctx, mainRepo)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestGetRemoteBranches_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestGetRemoteBranches_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	remoteBranches, err := s.Client.GetRemoteBranches(ctx, "")
	s.Require().Error(err)
	s.Nil(remoteBranches)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	remoteBranches, err = s.Client.GetRemoteBranches(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Nil(remoteBranches)
	s.True(domain.IsDomainErrorType(err, domain.ErrGitCommand))
}

// TestGetRemoteBranches_WithRemote tests with actual remote branches
func (s *GitClientContextTestSuite) TestGetRemoteBranches_WithRemote() {
	ctx := context.Background()

	// Setup repository
	mainRepo := filepath.Join(s.TempDir, "main")
	repo, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Add a remote with some branches (simulated)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/user/repo.git"},
	})
	s.Require().NoError(err)

	// Get remote branches - should be empty since we haven't actually fetched any
	remoteBranches, err := s.Client.GetRemoteBranches(ctx, mainRepo)
	s.Require().NoError(err)
	s.Empty(remoteBranches) // No actual remote branches without fetching
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

// TestBranchExists_ContextCancellation tests context cancellation in BranchExists
func (s *GitClientContextTestSuite) TestBranchExists_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// This should return false (not error) for context cancellation
	exists := s.Client.BranchExists(ctx, repoPath, "master")
	s.False(exists)
}

// TestBranchExists_EmptyInputs tests behavior with empty inputs
func (s *GitClientContextTestSuite) TestBranchExists_EmptyInputs() {
	ctx := context.Background()

	repoPath := filepath.Join(s.TempDir, "repo")
	_, err := git.PlainInit(repoPath, false)
	s.Require().NoError(err)

	// Test with empty repo path
	exists := s.Client.BranchExists(ctx, "", "master")
	s.False(exists)

	// Test with empty branch name
	exists = s.Client.BranchExists(ctx, repoPath, "")
	s.False(exists)

	// Test with both empty
	exists = s.Client.BranchExists(ctx, "", "")
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
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestIsMainRepository_NoOriginRemote tests repository without origin remote
func (s *GitClientContextTestSuite) TestIsMainRepository_NoOriginRemote() {
	ctx := context.Background()

	// Setup a repository without origin remote
	mainRepo := filepath.Join(s.TempDir, "no-origin-repo")
	_, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Should be main repository (no origin remote)
	isMain, err := s.Client.IsMainRepository(ctx, mainRepo)
	s.Require().NoError(err)
	s.True(isMain)
}

// TestIsMainRepository_WithRemoteOrigin tests repository with remote origin (GitHub/GitLab)
func (s *GitClientContextTestSuite) TestIsMainRepository_WithRemoteOrigin() {
	ctx := context.Background()

	// Setup a repository with remote origin
	mainRepo := filepath.Join(s.TempDir, "remote-origin-repo")
	repo, err := git.PlainInit(mainRepo, false)
	s.Require().NoError(err)

	// Add a remote origin pointing to GitHub
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/user/repo.git"},
	})
	s.Require().NoError(err)

	// Should be main repository (remote origin)
	isMain, err := s.Client.IsMainRepository(ctx, mainRepo)
	s.Require().NoError(err)
	s.True(isMain)
}

// TestIsMainRepository_WorktreeWithLocalOrigin tests worktree with local origin
func (s *GitClientContextTestSuite) TestIsMainRepository_WorktreeWithLocalOrigin() {
	ctx := context.Background()

	// Setup main repository
	mainRepoPath := filepath.Join(s.TempDir, "main-repo")
	_, err := git.PlainInit(mainRepoPath, false)
	s.Require().NoError(err)

	// Setup worktree repository
	worktreePath := filepath.Join(s.TempDir, "worktree-repo")
	worktreeRepo, err := git.PlainInit(worktreePath, false)
	s.Require().NoError(err)

	// Add origin pointing to main repository (same workspace)
	_, err = worktreeRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{mainRepoPath},
	})
	s.Require().NoError(err)

	// Should NOT be main repository (has local origin in same workspace)
	isMain, err := s.Client.IsMainRepository(ctx, worktreePath)
	s.Require().NoError(err)
	s.False(isMain)
}

// TestIsBareRepository_WithContext tests context support in IsBareRepository
func (s *GitClientContextTestSuite) TestIsBareRepository_WithContext() {
	ctx := context.Background()

	// Setup a bare git repository
	bareRepoDir := filepath.Join(s.TempDir, "bare-repo")
	_, err := git.PlainInit(bareRepoDir, true) // true = bare repository
	s.Require().NoError(err)

	// Setup a regular git repository
	regularRepoDir := filepath.Join(s.TempDir, "regular-repo")
	_, err = git.PlainInit(regularRepoDir, false) // false = regular repository
	s.Require().NoError(err)

	// Test bare repository detection
	isBare, err := s.Client.IsBareRepository(ctx, bareRepoDir)
	s.Require().NoError(err)
	s.True(isBare)

	// Test regular repository detection
	isBare, err = s.Client.IsBareRepository(ctx, regularRepoDir)
	s.Require().NoError(err)
	s.False(isBare)
}

// TestIsBareRepository_ContextCancellation tests context cancellation handling
func (s *GitClientContextTestSuite) TestIsBareRepository_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	bareRepoDir := filepath.Join(s.TempDir, "bare-repo")
	_, err := git.PlainInit(bareRepoDir, true)
	s.Require().NoError(err)

	// This should fail with context canceled error
	_, err = s.Client.IsBareRepository(ctx, bareRepoDir)
	s.Require().Error(err)
	s.True(domain.IsDomainErrorType(err, domain.ErrValidation))
}

// TestIsBareRepository_InvalidPath tests error handling for invalid paths
func (s *GitClientContextTestSuite) TestIsBareRepository_InvalidPath() {
	ctx := context.Background()

	// Test with empty path
	isBare, err := s.Client.IsBareRepository(ctx, "")
	s.Require().Error(err)
	s.False(isBare)
	s.True(domain.IsDomainErrorType(err, domain.ErrInvalidPath))

	// Test with non-existent path
	isBare, err = s.Client.IsBareRepository(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.False(isBare)
	s.True(domain.IsDomainErrorType(err, domain.ErrGitCommand))
}

// TestCheckCandidateWorktree tests the checkCandidateWorktree function
func (s *GitClientContextTestSuite) TestCheckCandidateWorktree() {
	ctx := context.Background()

	// Setup main repository
	mainRepoPath, _ := s.setupRepositoryWithInitialCommit("main-repo")

	// Create a worktree directory that is NOT a valid worktree
	candidatePath := filepath.Join(s.TempDir, "candidate")
	err := os.MkdirAll(candidatePath, 0755)
	s.Require().NoError(err)

	// Test candidate that is not a git repository
	worktreeInfo, err := s.Client.checkCandidateWorktree(ctx, candidatePath, mainRepoPath)
	s.Require().NoError(err)
	s.Nil(worktreeInfo)

	// Test candidate that is a git repository but not a worktree
	_, err = git.PlainInit(candidatePath, false)
	s.Require().NoError(err)

	worktreeInfo, err = s.Client.checkCandidateWorktree(ctx, candidatePath, mainRepoPath)
	s.Require().NoError(err)
	s.Nil(worktreeInfo) // Should be nil because it doesn't have main repo as origin
}

// TestHasOriginRemote tests the hasOriginRemote function
func (s *GitClientContextTestSuite) TestHasOriginRemote() {
	// Setup main repository
	mainRepoPath, _ := s.setupRepositoryWithInitialCommit("main-repo")

	// Create a separate repository
	otherRepoPath := filepath.Join(s.TempDir, "other-repo")
	otherRepo, err := git.PlainInit(otherRepoPath, false)
	s.Require().NoError(err)

	// Test that it doesn't have main repo as origin (since we didn't set it up)
	hasOrigin := s.Client.hasOriginRemote(context.Background(), otherRepoPath, mainRepoPath)
	s.False(hasOrigin)

	// Add a remote to make it a proper worktree setup
	_, err = otherRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{mainRepoPath},
	})
	s.Require().NoError(err)

	// Test again - should still be false because paths don't match exactly
	hasOrigin = s.Client.hasOriginRemote(context.Background(), otherRepoPath, mainRepoPath)
	s.True(hasOrigin) // Should be true now that we added the origin remote
}

// TestGetWorktreeStatusCLI tests the CLI fallback for worktree status
func (s *GitClientContextTestSuite) TestGetWorktreeStatusCLI() {
	ctx := context.Background()

	// Setup a repository with initial commit
	repoPath, _ := s.setupRepositoryWithInitialCommit("cli-test-repo")

	// Test CLI fallback method
	worktreeInfo, err := s.Client.getWorktreeStatusCLI(ctx, repoPath)
	s.Require().NoError(err)
	s.NotNil(worktreeInfo)
	s.Equal(repoPath, worktreeInfo.Path)
	s.Equal("master", worktreeInfo.Branch) // go-git uses "master" as default
	s.NotEmpty(worktreeInfo.Commit)
	s.True(worktreeInfo.Clean)
}

// TestGetWorktreeStatusCLI_ErrorHandling tests error handling in CLI method
func (s *GitClientContextTestSuite) TestGetWorktreeStatusCLI_ErrorHandling() {
	ctx := context.Background()

	// Test with non-existent directory
	worktreeInfo, err := s.Client.getWorktreeStatusCLI(ctx, "/non/existent/path")
	s.Require().Error(err)
	s.Nil(worktreeInfo)
	s.True(domain.IsDomainErrorType(err, domain.ErrGitCommand))
}

func TestGitClientContextTestSuite(t *testing.T) {
	suite.Run(t, new(GitClientContextTestSuite))
}
