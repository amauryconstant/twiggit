//go:build concurrent
// +build concurrent

package concurrent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"twiggit/internal/core"
	"twiggit/internal/git"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

// ConcurrentTestSuite provides concurrent operation testing
type ConcurrentTestSuite struct {
	suite.Suite
	tempDir     string
	config      *core.Config
	gitExecutor git.CommandExecutor
}

func TestConcurrentSuite(t *testing.T) {
	suite.Run(t, new(ConcurrentTestSuite))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
	os.Exit(m.Run())
}

func (s *ConcurrentTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()

	// Create directories
	projectsDir := filepath.Join(s.tempDir, "projects")
	worktreesDir := filepath.Join(s.tempDir, "worktrees")
	s.Require().NoError(os.MkdirAll(projectsDir, 0o755))
	s.Require().NoError(os.MkdirAll(worktreesDir, 0o755))

	// Create config
	s.config = &core.Config{
		ProjectsDirectory:   projectsDir,
		WorktreesDirectory:  worktreesDir,
		DefaultSourceBranch: "main",
	}

	// Initialize infrastructure
	s.gitExecutor = git.NewCommandExecutor(30 * time.Second)
}

// createTestProject creates a test project with the given name
func (s *ConcurrentTestSuite) createTestProject(name string) string {
	projectPath := filepath.Join(s.config.ProjectsDirectory, name)

	// Use shell commands to create a proper git repo
	cmd := exec.Command("git", "init", projectPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to init git repo: %v\nOutput: %s", err, string(output))
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@twiggit.dev")
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to configure git email: %v\nOutput: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to configure git name: %v\nOutput: %s", err, string(output))
	}

	// Create initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit", "--allow-empty")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_EMAIL=test@twiggit.dev",
		"GIT_AUTHOR_NAME=Test User",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to create initial commit: %v\nOutput: %s", err, string(output))
	}

	// Rename default branch to main
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to rename branch to main: %v\nOutput: %s", err, string(output))
	}

	return projectPath
}

// TestConcurrentListOperations tests concurrent list operations on same project
func (s *ConcurrentTestSuite) TestConcurrentListOperations() {
	projectPath := s.createTestProject("test-project")

	// Create some worktrees first using git CLI
	for i := range 3 {
		branchName := fmt.Sprintf("feature-%d", i)
		worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branchName)

		// Create branch
		cmd := exec.Command("git", "branch", branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create branch: %v\nOutput: %s", err, string(output))
		}

		// Create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create worktree: %v\nOutput: %s", err, string(output))
		}
	}

	// Run concurrent list operations using git CLI
	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	for range 10 {
		wg.Go(func() {
			cmd := exec.Command("git", "worktree", "list", "--porcelain")
			cmd.Dir = projectPath
			output, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- fmt.Errorf("list failed: %w: %s", err, string(output))
				return
			}
			// Verify we got valid output
			if len(output) == 0 {
				errChan <- fmt.Errorf("empty output from worktree list")
			}
		})
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		s.Assert().NoError(err, "Concurrent list operation failed")
	}
}

// TestConcurrentCreateOperations tests concurrent create operations on different worktrees
func (s *ConcurrentTestSuite) TestConcurrentCreateOperations() {
	projectPath := s.createTestProject("test-project")

	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	for i := range 5 {
		wg.Go(func() {
			branchName := fmt.Sprintf("concurrent-feature-%d", i)
			worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branchName)

			// Create branch
			cmd := exec.Command("git", "branch", branchName)
			cmd.Dir = projectPath
			if output, err := cmd.CombinedOutput(); err != nil {
				errChan <- fmt.Errorf("branch create failed: %w: %s", err, string(output))
				return
			}

			// Create worktree
			cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
			cmd.Dir = projectPath
			if output, err := cmd.CombinedOutput(); err != nil {
				errChan <- fmt.Errorf("worktree create failed: %w: %s", err, string(output))
				return
			}
		})
	}

	wg.Wait()
	close(errChan)

	// Check for errors - all creates should succeed
	errorCount := 0
	for err := range errChan {
		s.T().Logf("Create error: %v", err)
		errorCount++
	}

	// All 5 should succeed since we're creating different branches
	s.Assert().Equal(0, errorCount, "All create operations should succeed")
}

// TestConcurrentDeleteOperations tests concurrent delete operations on different worktrees
func (s *ConcurrentTestSuite) TestConcurrentDeleteOperations() {
	projectPath := s.createTestProject("test-project")

	// Create worktrees to delete
	branches := []string{"delete-1", "delete-2", "delete-3"}
	for _, branch := range branches {
		worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branch)

		// Create branch
		cmd := exec.Command("git", "branch", branch)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create branch: %v\nOutput: %s", err, string(output))
		}

		// Create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create worktree: %v\nOutput: %s", err, string(output))
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	for _, branch := range branches {
		wg.Go(func() {
			worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branch)

			cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
			cmd.Dir = projectPath
			if output, err := cmd.CombinedOutput(); err != nil {
				errChan <- fmt.Errorf("delete failed: %w: %s", err, string(output))
			}
		})
	}

	wg.Wait()
	close(errChan)

	// Check for errors - all deletes should succeed
	for err := range errChan {
		s.Assert().NoError(err, "Delete operation should succeed")
	}
}

// TestConcurrentCreateDeleteOperations tests create and delete different worktrees concurrently
func (s *ConcurrentTestSuite) TestConcurrentCreateDeleteOperations() {
	projectPath := s.createTestProject("test-project")

	// Create a worktree to delete
	worktreeToDelete := "to-delete"
	deletePath := filepath.Join(s.config.WorktreesDirectory, "test-project", worktreeToDelete)

	// Create branch
	cmd := exec.Command("git", "branch", worktreeToDelete)
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to create branch: %v\nOutput: %s", err, string(output))
	}

	// Create worktree
	cmd = exec.Command("git", "worktree", "add", deletePath, worktreeToDelete)
	cmd.Dir = projectPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.T().Fatalf("Failed to create worktree: %v\nOutput: %s", err, string(output))
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Create goroutine
	wg.Go(func() {
		branchName := "concurrent-create"
		worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branchName)

		// Create branch
		cmd := exec.Command("git", "branch", branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			errChan <- fmt.Errorf("branch create failed: %w: %s", err, string(output))
			return
		}

		// Create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			errChan <- fmt.Errorf("worktree create failed: %w: %s", err, string(output))
		}
	})

	// Delete goroutine
	wg.Go(func() {
		cmd := exec.Command("git", "worktree", "remove", "--force", deletePath)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			errChan <- fmt.Errorf("delete failed: %w: %s", err, string(output))
		}
	})

	wg.Wait()
	close(errChan)

	// Check for errors - both operations should succeed
	for err := range errChan {
		s.Assert().NoError(err, "Concurrent create/delete should succeed")
	}
}

// TestConcurrentPruneWhileList tests prune while listing operations
func (s *ConcurrentTestSuite) TestConcurrentPruneWhileList() {
	projectPath := s.createTestProject("test-project")

	// Create and merge some worktrees
	for i := range 3 {
		branchName := fmt.Sprintf("prunable-%d", i)
		worktreePath := filepath.Join(s.config.WorktreesDirectory, "test-project", branchName)

		// Create branch
		cmd := exec.Command("git", "branch", branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create branch: %v\nOutput: %s", err, string(output))
		}

		// Create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = projectPath
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Fatalf("Failed to create worktree: %v\nOutput: %s", err, string(output))
		}

		// Merge the branch using git command
		cmd = exec.Command("git", "merge", branchName, "--no-edit")
		cmd.Dir = projectPath
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_EMAIL=test@twiggit.dev",
			"GIT_AUTHOR_NAME=Test User",
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			s.T().Logf("Merge output: %s", string(output))
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 6)

	// List operations
	for range 3 {
		wg.Go(func() {
			cmd := exec.Command("git", "worktree", "list", "--porcelain")
			cmd.Dir = projectPath
			output, err := cmd.CombinedOutput()
			if err != nil {
				errChan <- fmt.Errorf("list failed: %w: %s", err, string(output))
				return
			}
			// Verify we got valid output
			if len(output) == 0 {
				errChan <- fmt.Errorf("empty output from worktree list")
			}
		})
	}

	// Prune operations
	for range 3 {
		wg.Go(func() {
			cmd := exec.Command("git", "worktree", "prune")
			cmd.Dir = projectPath
			if output, err := cmd.CombinedOutput(); err != nil {
				errChan <- fmt.Errorf("prune failed: %w: %s", err, string(output))
			}
		})
	}

	wg.Wait()
	close(errChan)

	// Check for errors - operations should not race
	for err := range errChan {
		s.Assert().NoError(err, "Concurrent prune/list should not race")
	}
}
