package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitTestHelper_CreateRepoWithCommits(t *testing.T) {
	testCases := []struct {
		name         string
		commitCount  int
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid repository with commits",
			commitCount: 3,
			expectError: false,
		},
		{
			name:        "zero commits",
			commitCount: 0,
			expectError: false,
		},
		{
			name:         "negative commits",
			commitCount:  -1,
			expectError:  true,
			errorMessage: "commit count cannot be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewGitTestHelper(t)

			if tc.expectError {
				assert.Panics(t, func() {
					helper.CreateRepoWithCommits(tc.commitCount)
				})
			} else {
				repoPath := helper.CreateRepoWithCommits(tc.commitCount)
				assert.NotEmpty(t, repoPath)
				assert.DirExists(t, repoPath)

				// Verify git repository
				gitRepo, err := helper.PlainOpen(repoPath)
				require.NoError(t, err)
				assert.NotNil(t, gitRepo)
			}
		})
	}
}

func TestGitTestHelper_FunctionalComposition(t *testing.T) {
	helper := NewGitTestHelper(t)

	// Test functional composition
	repoPath := helper.WithCommits(3).WithBranch("feature-test").CreateRepoWithCommits(3)
	assert.NotEmpty(t, repoPath)
	assert.DirExists(t, repoPath)

	// Verify branch exists
	branches, err := helper.ListBranches(repoPath)
	require.NoError(t, err)
	assert.Contains(t, branches, "feature-test")
}

func TestGitTestHelper_CreateBranch(t *testing.T) {
	helper := NewGitTestHelper(t)
	repoPath := helper.CreateRepoWithCommits(1)

	err := helper.CreateBranch(repoPath, "feature-branch")
	require.NoError(t, err)

	// Verify branch exists
	branches, err := helper.ListBranches(repoPath)
	require.NoError(t, err)
	assert.Contains(t, branches, "feature-branch")
}

func TestGitTestHelper_ListBranches(t *testing.T) {
	helper := NewGitTestHelper(t)
	repoPath := helper.CreateRepoWithCommits(1)

	// Create additional branches
	err := helper.CreateBranch(repoPath, "feature-a")
	require.NoError(t, err)
	err = helper.CreateBranch(repoPath, "feature-b")
	require.NoError(t, err)

	branches, err := helper.ListBranches(repoPath)
	require.NoError(t, err)
	assert.Contains(t, branches, "master")
	assert.Contains(t, branches, "feature-a")
	assert.Contains(t, branches, "feature-b")
}

func TestRepoTestHelper_SetupTestRepo(t *testing.T) {
	testCases := []struct {
		name        string
		projectName string
		expectError bool
	}{
		{
			name:        "valid project name",
			projectName: "test-project",
			expectError: false,
		},
		{
			name:        "empty project name",
			projectName: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewRepoTestHelper(t)

			if tc.expectError {
				assert.Panics(t, func() {
					helper.SetupTestRepo(tc.projectName)
				})
			} else {
				repoPath := helper.SetupTestRepo(tc.projectName)
				assert.NotEmpty(t, repoPath)
				assert.DirExists(t, repoPath)
				assert.Contains(t, repoPath, tc.projectName)
			}
		})
	}
}

func TestRepoTestHelper_FunctionalComposition(t *testing.T) {
	helper := NewRepoTestHelper(t)

	// Test functional composition
	repoPath := helper.WithProject("test-project").WithCommits(2).SetupTestRepo("test-project")
	assert.NotEmpty(t, repoPath)
	assert.DirExists(t, repoPath)

	// Verify the repo is stored in helper
	storedPath := helper.GetRepoPath("test-project")
	assert.Equal(t, repoPath, storedPath)
}

func TestRepoTestHelper_GetRepoPath(t *testing.T) {
	helper := NewRepoTestHelper(t)
	repoPath := helper.SetupTestRepo("test-project")

	// Test getting existing repo
	storedPath := helper.GetRepoPath("test-project")
	assert.Equal(t, repoPath, storedPath)

	// Test getting non-existent repo
	assert.Panics(t, func() {
		helper.GetRepoPath("non-existent")
	})
}

func TestRepoTestHelper_Cleanup(t *testing.T) {
	helper := NewRepoTestHelper(t)
	repoPath := helper.SetupTestRepo("test-project")

	// Verify repo exists
	assert.DirExists(t, repoPath)

	// Cleanup
	helper.Cleanup()

	// Verify repo is cleaned up
	assert.NoDirExists(t, repoPath)
}

func TestShellTestHelper_ExecuteCommand(t *testing.T) {
	testCases := []struct {
		name         string
		command      string
		args         []string
		expectError  bool
		expectOutput string
	}{
		{
			name:         "successful echo command",
			command:      "echo",
			args:         []string{"hello", "world"},
			expectError:  false,
			expectOutput: "hello world",
		},
		{
			name:        "non-existent command",
			command:     "nonexistent-command-12345",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helper := NewShellTestHelper(t)

			output, err := helper.ExecuteCommand(tc.command, tc.args...)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, tc.expectOutput)
			}
		})
	}
}

func TestShellTestHelper_FunctionalComposition(t *testing.T) {
	helper := NewShellTestHelper(t)

	// Test functional composition
	output, err := helper.WithCommand("echo").WithArgs("test", "output").ExecuteCommand("echo", "test", "output")
	require.NoError(t, err)
	assert.Contains(t, output, "test output")
}

func TestShellTestHelper_WithWorkingDirectory(t *testing.T) {
	helper := NewShellTestHelper(t)
	tempDir := t.TempDir()

	output, err := helper.WithWorkingDirectory(tempDir).ExecuteCommand("pwd")
	require.NoError(t, err)
	assert.Contains(t, output, tempDir)
}

func TestShellTestHelper_WithEnvironment(t *testing.T) {
	helper := NewShellTestHelper(t)

	output, err := helper.WithEnvironment("TEST_VAR", "test_value").ExecuteCommand("sh", "-c", "echo $TEST_VAR")
	require.NoError(t, err)
	assert.Contains(t, output, "test_value")
}

func TestPerformanceTestHelper_MeasureFunction(t *testing.T) {
	helper := NewPerformanceTestHelper(t)

	// Test measuring a simple function
	duration, err := helper.MeasureFunction(func() {
		// Simulate some work
		for i := 0; i < 1000; i++ {
			_ = i * i
		}
	})

	require.NoError(t, err)
	assert.Greater(t, duration, time.Duration(0))
}

func TestPerformanceTestHelper_BenchmarkFunction(t *testing.T) {
	helper := NewPerformanceTestHelper(t)

	// Test benchmarking a function
	result, err := helper.BenchmarkFunction(10, func() interface{} {
		// Simulate some work and return a result
		sum := 0
		for i := 0; i < 100; i++ {
			sum += i
		}
		return sum
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 10, result.Iterations)
	assert.Greater(t, result.TotalDuration, time.Duration(0))
	assert.Greater(t, result.AvgDuration, time.Duration(0))
}

func TestPerformanceTestHelper_FunctionalComposition(t *testing.T) {
	helper := NewPerformanceTestHelper(t)

	// Test functional composition
	result, err := helper.WithIterations(5).WithWarmup(true).BenchmarkFunction(5, func() interface{} {
		return 42
	})

	require.NoError(t, err)
	assert.Equal(t, 5, result.Iterations)
	assert.Equal(t, 42, result.LastResult)
}

func TestPerformanceTestHelper_MemoryUsage(t *testing.T) {
	helper := NewPerformanceTestHelper(t)

	// Test memory usage measurement - just verify it works without asserting specific values
	before, after, err := helper.MeasureMemoryUsage(func() {
		// Simple function that doesn't allocate much
		sum := 0
		for i := 0; i < 100; i++ {
			sum += i
		}
		_ = sum
	})

	require.NoError(t, err)
	// Just verify we got measurements
	assert.Positive(t, before)
	assert.Positive(t, after)
}
