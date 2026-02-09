package infrastructure

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"twiggit/internal/domain"
)

type CLIClientTestSuite struct {
	suite.Suite
}

func TestCLIClientSuite(t *testing.T) {
	suite.Run(t, new(CLIClientTestSuite))
}

// TestParseWorktreeLine tests the pure function for parsing worktree lines
func (s *CLIClientTestSuite) TestParseWorktreeLine() {
	testCases := []struct {
		name           string
		line           string
		expectedResult *domain.WorktreeInfo
	}{
		{
			name:           "worktree line",
			line:           "worktree /path/to/worktree",
			expectedResult: &domain.WorktreeInfo{Path: "/path/to/worktree"},
		},
		{
			name:           "HEAD line",
			line:           "HEAD abc1234",
			expectedResult: nil,
		},
		{
			name:           "branch line",
			line:           "branch refs/heads/main",
			expectedResult: nil,
		},
		{
			name:           "detached line",
			line:           "detached",
			expectedResult: nil,
		},
		{
			name:           "empty line",
			line:           "",
			expectedResult: nil,
		},
		{
			name:           "unrelated line",
			line:           "some other content",
			expectedResult: nil,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			result := parseWorktreeLine(tt.line)
			s.Equal(tt.expectedResult, result)
		})
	}
}

// TestBuildWorktreeAddArgs tests the pure function for building worktree add arguments
func (s *CLIClientTestSuite) TestBuildWorktreeAddArgs() {
	testCases := []struct {
		name         string
		branchExists bool
		branchName   string
		worktreePath string
		sourceBranch string
		expectedArgs []string
	}{
		{
			name:         "new branch with source",
			branchExists: false,
			branchName:   "feature",
			worktreePath: "/path/to/worktree",
			sourceBranch: "main",
			expectedArgs: []string{"worktree", "add", "-b", "feature", "/path/to/worktree", "main"},
		},
		{
			name:         "new branch without source",
			branchExists: false,
			branchName:   "feature",
			worktreePath: "/path/to/worktree",
			sourceBranch: "",
			expectedArgs: []string{"worktree", "add", "-b", "feature", "/path/to/worktree"},
		},
		{
			name:         "existing branch",
			branchExists: true,
			branchName:   "existing",
			worktreePath: "/path/to/worktree",
			sourceBranch: "main",
			expectedArgs: []string{"worktree", "add", "/path/to/worktree", "existing"},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			args := buildWorktreeAddArgs(tt.branchExists, tt.branchName, tt.worktreePath, tt.sourceBranch)
			s.Equal(tt.expectedArgs, args)
		})
	}
}

// TestBuildWorktreeRemoveArgs tests the pure function for building worktree remove arguments
func (s *CLIClientTestSuite) TestBuildWorktreeRemoveArgs() {
	testCases := []struct {
		name         string
		worktreePath string
		force        bool
		expectedArgs []string
	}{
		{
			name:         "remove without force",
			worktreePath: "/path/to/worktree",
			force:        false,
			expectedArgs: []string{"worktree", "remove", "/path/to/worktree"},
		},
		{
			name:         "remove with force",
			worktreePath: "/path/to/worktree",
			force:        true,
			expectedArgs: []string{"worktree", "remove", "--force", "/path/to/worktree"},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			args := buildWorktreeRemoveArgs(tt.worktreePath, tt.force)
			s.Equal(tt.expectedArgs, args)
		})
	}
}

func TestCLIClient_CreateWorktree(t *testing.T) {
	worktreeDir := t.TempDir()
	callCount := 0
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)

		callCount++
		if callCount == 1 {
			assert.Equal(t, []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"}, args)
			return &CommandResult{ExitCode: 1, Stdout: ""}, nil
		} else {
			assert.Equal(t, []string{"worktree", "add", "-b", "feature", worktreeDir, "main"}, args)
			if err := os.MkdirAll(worktreeDir, 0755); err != nil {
				return nil, err
			}
			return &CommandResult{ExitCode: 0, Stdout: ""}, nil
		}
	})
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(context.Background(), "/test/repo", "feature", "main", worktreeDir)
	assert.NoError(t, err)
}

func TestCLIClient_CreateWorktree_WithExistingBranch(t *testing.T) {
	worktreeDir := t.TempDir()
	callCount := 0
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)

		callCount++
		if callCount == 1 {
			assert.Equal(t, []string{"show-ref", "--verify", "--quiet", "refs/heads/existing-branch"}, args)
			return &CommandResult{ExitCode: 0, Stdout: ""}, nil
		} else {
			assert.Equal(t, []string{"worktree", "add", worktreeDir, "existing-branch"}, args)
			if err := os.MkdirAll(worktreeDir, 0755); err != nil {
				return nil, err
			}
			return &CommandResult{ExitCode: 0, Stdout: ""}, nil
		}
	})
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(context.Background(), "/test/repo", "existing-branch", "", worktreeDir)
	assert.NoError(t, err)
}

func TestCLIClient_CreateWorktree_Failure(t *testing.T) {
	callCount := 0
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)

		callCount++
		if callCount == 1 {
			// First call: check if branch exists (it doesn't)
			assert.Equal(t, []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"}, args)
			return &CommandResult{ExitCode: 1, Stdout: ""}, nil
		} else {
			// Second call: worktree add fails
			assert.Equal(t, []string{"worktree", "add", "-b", "feature", "/path/to/worktree", "main"}, args)
			return &CommandResult{ExitCode: 1, Stderr: "fatal: Invalid path"}, nil
		}
	})
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(context.Background(), "/test/repo", "feature", "main", "/path/to/worktree")
	require.Error(t, err)
	var worktreeErr *domain.GitWorktreeError
	require.ErrorAs(t, err, &worktreeErr)
}

func TestCLIClient_DeleteWorktree(t *testing.T) {
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)
		assert.Equal(t, []string{"worktree", "remove", "/path/to/worktree"}, args)
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	})
	client := NewCLIClient(mockExecutor)

	err := client.DeleteWorktree(context.Background(), "/test/repo", "/path/to/worktree", false)
	assert.NoError(t, err)
}

func TestCLIClient_DeleteWorktree_WithForce(t *testing.T) {
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)
		assert.Equal(t, []string{"worktree", "remove", "--force", "/path/to/worktree"}, args)
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	})
	client := NewCLIClient(mockExecutor)

	err := client.DeleteWorktree(context.Background(), "/test/repo", "/path/to/worktree", true)
	assert.NoError(t, err)
}

func TestCLIClient_ListWorktrees(t *testing.T) {
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)
		assert.Equal(t, []string{"worktree", "list", "--porcelain"}, args)

		// Mock git worktree list output
		mockOutput := `worktree /path/to/repo
HEAD abcdef1
branch refs/heads/main
worktree /path/to/worktree1
HEAD bcdef2a
branch refs/heads/feature-branch
worktree /path/to/worktree2
HEAD cdef3ab
detached`

		return &CommandResult{ExitCode: 0, Stdout: mockOutput}, nil
	})
	client := NewCLIClient(mockExecutor)

	worktrees, err := client.ListWorktrees(context.Background(), "/test/repo")
	require.NoError(t, err)
	assert.Len(t, worktrees, 3)

	// Check main worktree
	mainWorktree := findWorktree(worktrees, "/path/to/repo")
	require.NotNil(t, mainWorktree)
	assert.Equal(t, "main", mainWorktree.Branch)
	assert.Equal(t, "abcdef1", mainWorktree.Commit)
	assert.False(t, mainWorktree.IsDetached)

	// Check feature worktree
	featureWorktree := findWorktree(worktrees, "/path/to/worktree1")
	require.NotNil(t, featureWorktree)
	assert.Equal(t, "feature-branch", featureWorktree.Branch)
	assert.Equal(t, "bcdef2a", featureWorktree.Commit)
	assert.False(t, featureWorktree.IsDetached)

	// Check detached worktree
	detachedWorktree := findWorktree(worktrees, "/path/to/worktree2")
	require.NotNil(t, detachedWorktree)
	assert.Equal(t, "cdef3ab", detachedWorktree.Commit)
	assert.True(t, detachedWorktree.IsDetached)
}

func TestCLIClient_PruneWorktrees(t *testing.T) {
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)
		assert.Equal(t, []string{"worktree", "prune"}, args)
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	})
	client := NewCLIClient(mockExecutor)

	err := client.PruneWorktrees(context.Background(), "/test/repo")
	assert.NoError(t, err)
}

func TestCLIClient_IsBranchMerged(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		output     string
		expected   bool
		expectErr  bool
	}{
		{
			name:       "branch is merged",
			branchName: "feature-branch",
			output:     "  main\n* feature-branch\n  develop\n",
			expected:   true,
			expectErr:  false,
		},
		{
			name:       "branch is not merged",
			branchName: "feature-branch",
			output:     "  main\n  develop\n",
			expected:   false,
			expectErr:  false,
		},
		{
			name:       "branch with asterisk is merged",
			branchName: "main",
			output:     "* main\n  develop\n",
			expected:   true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
				assert.Equal(t, "/test/repo", dir)
				assert.Equal(t, "git", cmd)
				assert.Equal(t, []string{"branch", "--merged"}, args)
				return &CommandResult{ExitCode: 0, Stdout: tt.output}, nil
			})
			client := NewCLIClient(mockExecutor)

			isMerged, err := client.IsBranchMerged(context.Background(), "/test/repo", tt.branchName)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, isMerged)
			}
		})
	}
}

func TestCLIClient_Timeout(t *testing.T) {
	// Create mock command executor
	mockExecutor := NewMockCommandExecutor(func(ctx context.Context, dir, cmd string, args ...string) (*CommandResult, error) {
		assert.Equal(t, "/test/repo", dir)
		assert.Equal(t, "git", cmd)
		assert.Equal(t, []string{"worktree", "list", "--porcelain"}, args)
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	})
	client := NewCLIClient(mockExecutor, 5)

	_, err := client.ListWorktrees(context.Background(), "/test/repo")
	assert.NoError(t, err)
}

func TestCLIClient_ParseWorktreeList(t *testing.T) {
	client := NewCLIClient(nil)

	// Test parsing porcelain output
	output := `worktree /path/to/repo
HEAD abcdef1
branch refs/heads/main
worktree /path/to/worktree1
HEAD bcdef2a
branch refs/heads/feature-branch
worktree /path/to/worktree2
HEAD cdef3ab
detached`

	worktrees, err := client.parseWorktreeList(output)
	require.NoError(t, err)
	assert.Len(t, worktrees, 3)

	// Verify parsed data
	assert.Equal(t, "/path/to/repo", worktrees[0].Path)
	assert.Equal(t, "main", worktrees[0].Branch)
	assert.Equal(t, "abcdef1", worktrees[0].Commit)
	assert.False(t, worktrees[0].IsDetached)

	assert.Equal(t, "/path/to/worktree1", worktrees[1].Path)
	assert.Equal(t, "feature-branch", worktrees[1].Branch)
	assert.Equal(t, "bcdef2a", worktrees[1].Commit)
	assert.False(t, worktrees[1].IsDetached)

	assert.Equal(t, "/path/to/worktree2", worktrees[2].Path)
	assert.Equal(t, "cdef3ab", worktrees[2].Commit)
	assert.True(t, worktrees[2].IsDetached)
}

// Helper functions

func findWorktree(worktrees []domain.WorktreeInfo, path string) *domain.WorktreeInfo {
	for _, worktree := range worktrees {
		if worktree.Path == path {
			return &worktree
		}
	}
	return nil
}
