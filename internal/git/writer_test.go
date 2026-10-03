package git

import (
	"os"
	"os/exec"
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCLIClient_ParseWorktreeLine(t *testing.T) {
	testCases := []struct {
		name           string
		line           string
		expectedResult *core.Worktree
	}{
		{
			name:           "worktree line",
			line:           "worktree /path/to/worktree",
			expectedResult: &core.Worktree{Path: "/path/to/worktree"},
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
		t.Run(tt.name, func(t *testing.T) {
			result := parseWorktreeLine(tt.line)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCLIClient_BuildWorktreeAddArgs(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			args := buildWorktreeAddArgs(tt.branchExists, tt.branchName, tt.worktreePath, tt.sourceBranch)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestCLIClient_BuildWorktreeRemoveArgs(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			args := buildWorktreeRemoveArgs(tt.worktreePath, tt.force)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestCLIClient_CreateWorktree(t *testing.T) {
	worktreeDir := t.TempDir()
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"}).Return(&CommandResult{ExitCode: 1, Stdout: ""}, nil)
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "add", "-b", "feature", worktreeDir, "main"}).Return(func() (*CommandResult, error) {
		if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
			return nil, err
		}
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	}())
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(t.Context(), "/test/repo", "feature", "main", worktreeDir)
	assert.NoError(t, err)
}

func TestCLIClient_CreateWorktree_WithExistingBranch(t *testing.T) {
	worktreeDir := t.TempDir()
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"show-ref", "--verify", "--quiet", "refs/heads/existing-branch"}).Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil)
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "add", worktreeDir, "existing-branch"}).Return(func() (*CommandResult, error) {
		if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
			return nil, err
		}
		return &CommandResult{ExitCode: 0, Stdout: ""}, nil
	}())
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(t.Context(), "/test/repo", "existing-branch", "", worktreeDir)
	assert.NoError(t, err)
}

func TestCLIClient_CreateWorktree_Failure(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"}).Return(&CommandResult{ExitCode: 1, Stdout: ""}, nil)
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "add", "-b", "feature", "/path/to/worktree", "main"}).Return(&CommandResult{ExitCode: 1, Stderr: "fatal: Invalid path"}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(t.Context(), "/test/repo", "feature", "main", "/path/to/worktree")
	require.Error(t, err)
	var worktreeErr *core.OperationError
	require.ErrorAs(t, err, &worktreeErr)
	require.Equal(t, "git.worktree.create", worktreeErr.Op)
}

// TestCLIClient_NonZeroExit_PreservesExecError asserts the cause
// chain mandated by golang-error-handling rule 5: a non-zero git
// exit code must remain reachable via errors.As(err, &*exec.ExitError)
// after the writer wraps it as *core.OperationError. If a future
// change drops result.Err from the NewWorktreeError cause, this
// test fails loudly.
func TestCLIClient_NonZeroExit_PreservesExecError(t *testing.T) {
	// Produce a real *exec.ExitError to seed the chain.
	cmd := exec.Command("sh", "-c", "exit 7")
	runErr := cmd.Run()
	require.Error(t, runErr)
	var seed *exec.ExitError
	require.ErrorAs(t, runErr, &seed)

	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(&CommandResult{ExitCode: 1, Stderr: "boom", Err: seed}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(t.Context(), "/test/repo", "feature", "main", "/path/to/worktree")
	require.Error(t, err)

	var extracted *exec.ExitError
	require.ErrorAs(t, err, &extracted, "exec.ExitError must remain reachable through the cause chain")
}

func TestCLIClient_DeleteWorktree(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "remove", "/path/to/worktree"}).Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.DeleteWorktree(t.Context(), "/test/repo", "/path/to/worktree", false)
	assert.NoError(t, err)
}

func TestCLIClient_DeleteWorktree_WithForce(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "remove", "--force", "/path/to/worktree"}).Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.DeleteWorktree(t.Context(), "/test/repo", "/path/to/worktree", true)
	assert.NoError(t, err)
}

func TestCLIClient_ListWorktrees(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockOutput := `worktree /path/to/repo
HEAD abcdef1
branch refs/heads/main
worktree /path/to/worktree1
HEAD bcdef2a
branch refs/heads/feature-branch
worktree /path/to/worktree2
HEAD cdef3ab
detached`
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "list", "--porcelain"}).Return(&CommandResult{ExitCode: 0, Stdout: mockOutput}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	worktrees, err := client.ListWorktrees(t.Context(), "/test/repo")
	require.NoError(t, err)
	assert.Len(t, worktrees, 3)

	mainWorktree := findWorktree(worktrees, "/path/to/repo")
	require.NotNil(t, mainWorktree)
	assert.Equal(t, "main", mainWorktree.Branch)
	assert.Equal(t, "abcdef1", mainWorktree.Commit)
	assert.False(t, mainWorktree.IsDetached)

	featureWorktree := findWorktree(worktrees, "/path/to/worktree1")
	require.NotNil(t, featureWorktree)
	assert.Equal(t, "feature-branch", featureWorktree.Branch)
	assert.Equal(t, "bcdef2a", featureWorktree.Commit)
	assert.False(t, featureWorktree.IsDetached)

	detachedWorktree := findWorktree(worktrees, "/path/to/worktree2")
	require.NotNil(t, detachedWorktree)
	assert.Equal(t, "cdef3ab", detachedWorktree.Commit)
	assert.True(t, detachedWorktree.IsDetached)
}

func TestCLIClient_PruneWorktrees(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "prune"}).Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.PruneWorktrees(t.Context(), "/test/repo")
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
			mockExecutor := NewMockCommandExecutor()
			mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"branch", "--merged"}).Return(&CommandResult{ExitCode: 0, Stdout: tt.output}, nil)
			t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
			client := NewCLIClient(mockExecutor)

			isMerged, err := client.IsBranchMerged(t.Context(), "/test/repo", tt.branchName)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, isMerged)
			}
		})
	}
}

func TestCLIClient_DeleteBranch(t *testing.T) {
	tests := []struct {
		name        string
		repoPath    string
		branchName  string
		mockResult  *CommandResult
		mockError   error
		expectErr   bool
		errContains string
	}{
		{
			name:       "successful deletion",
			repoPath:   "/test/repo",
			branchName: "feature-branch",
			mockResult: &CommandResult{ExitCode: 0, Stdout: "", Stderr: ""},
			expectErr:  false,
		},
		{
			name:        "empty repo path",
			repoPath:    "",
			branchName:  "feature",
			expectErr:   true,
			errContains: "repository path cannot be empty",
		},
		{
			name:        "empty branch name",
			repoPath:    "/test/repo",
			branchName:  "",
			expectErr:   true,
			errContains: "branch name cannot be empty",
		},
		{
			name:       "branch not found - idempotent",
			repoPath:   "/test/repo",
			branchName: "non-existent",
			mockResult: &CommandResult{ExitCode: 1, Stdout: "", Stderr: "error: branch 'non-existent' not found"},
			expectErr:  false,
		},
		{
			name:        "git command fails",
			repoPath:    "/test/repo",
			branchName:  "feature",
			mockResult:  &CommandResult{ExitCode: 1, Stdout: "", Stderr: "error: some error"},
			expectErr:   true,
			errContains: "git branch -D failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := NewMockCommandExecutor()
			if tt.repoPath != "" && tt.branchName != "" {
				mockExecutor.On("ExecuteWithTimeout", mock.Anything, tt.repoPath, "git", mock.AnythingOfType("time.Duration"), []string{"branch", "-D", tt.branchName}).Return(tt.mockResult, tt.mockError)
			}
			t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
			client := NewCLIClient(mockExecutor)

			err := client.DeleteBranch(t.Context(), tt.repoPath, tt.branchName)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCLIClient_Timeout(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "list", "--porcelain"}).Return(&CommandResult{ExitCode: 0, Stdout: ""}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor, 5)

	_, err := client.ListWorktrees(t.Context(), "/test/repo")
	assert.NoError(t, err)
}

func TestCLIClient_ParseWorktreeList(t *testing.T) {
	client := NewCLIClient(nil)

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

func TestCLIClient_NilResultGuards(t *testing.T) {
	t.Run("CreateWorktree", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"show-ref", "--verify", "--quiet", "refs/heads/feature"}).Return(&CommandResult{ExitCode: 1}, nil)
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), []string{"worktree", "add", "-b", "feature", "/path/to/worktree", "main"}).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		err := client.CreateWorktree(t.Context(), "/test/repo", "feature", "main", "/path/to/worktree")
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for worktree create")
	})

	t.Run("DeleteWorktree", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		err := client.DeleteWorktree(t.Context(), "/test/repo", "/path/to/worktree", false)
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for worktree delete")
	})

	t.Run("ListWorktrees", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		_, err := client.ListWorktrees(t.Context(), "/test/repo")
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for worktree list")
	})

	t.Run("PruneWorktrees", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		err := client.PruneWorktrees(t.Context(), "/test/repo")
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for worktree prune")
	})

	t.Run("DeleteBranch", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		err := client.DeleteBranch(t.Context(), "/test/repo", "feature")
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for branch delete")
	})

	t.Run("IsBranchMerged", func(t *testing.T) {
		must := require.New(t)
		is := assert.New(t)
		mockExecutor := NewMockCommandExecutor()
		mockExecutor.On("ExecuteWithTimeout", mock.Anything, "/test/repo", "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(nil, nil)
		t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
		client := NewCLIClient(mockExecutor)
		_, err := client.IsBranchMerged(t.Context(), "/test/repo", "feature")
		must.Error(err)
		is.Contains(err.Error(), "command executor returned nil result for branch merge check")
	})
}

func findWorktree(worktrees []core.Worktree, path string) *core.Worktree {
	for _, worktree := range worktrees {
		if worktree.Path == path {
			return &worktree
		}
	}
	return nil
}

// TestNewCLIClient_ReturnsWriteSideConcrete pins git-client R3.S1:
// NewCLIClient returns a non-nil *CLIClient and the composite *Client
// satisfies both write-side roles (WorktreeWriter, BranchWriter) at the
// type level. Uses a typed-nil *Client for the interface check so no
// real git binary is required.
func TestNewCLIClient_ReturnsWriteSideConcrete(t *testing.T) {
	t.Parallel()

	executor := NewCommandExecutor(defaultCLITimeout)
	client := NewCLIClient(executor)
	require.NotNil(t, client)
	assert.IsType(t, &CLIClient{}, client)

	// Composite *Client must satisfy both write-side roles via embedded
	// promotion of *cliClient. Typed-nil pointer is the cheapest way to
	// ask "does this type implement this interface" without a real repo.
	var composite *Client
	assert.Implements(t, (*core.WorktreeWriter)(nil), composite)
	assert.Implements(t, (*core.BranchWriter)(nil), composite)
}

// TestWriteSideFailure_OpIsGitWorktree pins git-client R4.S2 for
// worktree ops: write-side failures must surface as *core.OperationError
// with the dot-concatenated Op "git.worktree.<method>" so callers can
// dispatch on the precise operation.
func TestWriteSideFailure_OpIsGitWorktree(t *testing.T) {
	t.Parallel()

	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, mock.Anything, "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(&CommandResult{ExitCode: 1, Stderr: "fatal: not a git repository"}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.CreateWorktree(t.Context(),
		"/non/existent/path", "feature", "", "/tmp/wt-feature")
	require.Error(t, err)

	var oe *core.OperationError
	require.ErrorAs(t, err, &oe)
	assert.Equal(t, "git.worktree.create", oe.Op,
		"worktree-op Op must follow \"git.worktree.<method>\" format")
}

// TestWriteSideFailure_OpIsGitBranch pins git-client R4.S2 for branch
// ops: DeleteBranch failures must surface as *core.OperationError with
// the dot-concatenated Op "git.branch.<method>". Closes the
// pre-existing misclassification (branch ops were tagged git.worktree).
func TestWriteSideFailure_OpIsGitBranch(t *testing.T) {
	t.Parallel()

	mockExecutor := NewMockCommandExecutor()
	mockExecutor.On("ExecuteWithTimeout", mock.Anything, mock.Anything, "git", mock.AnythingOfType("time.Duration"), mock.Anything).Return(&CommandResult{ExitCode: 1, Stderr: "fatal: not a git repository"}, nil)
	t.Cleanup(func() { mockExecutor.AssertExpectations(t) })
	client := NewCLIClient(mockExecutor)

	err := client.DeleteBranch(t.Context(),
		"/non/existent/path", "feature")
	require.Error(t, err)

	var oe *core.OperationError
	require.ErrorAs(t, err, &oe)
	assert.Equal(t, "git.branch.delete", oe.Op,
		"branch-op Op must follow \"git.branch.<method>\" format")
}
