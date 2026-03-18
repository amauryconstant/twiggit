package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
)

func setupHookRunnerTest(t *testing.T) (HookRunner, *MockCommandExecutor, string) {
	t.Helper()
	mockExec := NewMockCommandExecutor()
	runner := NewHookRunner(mockExec)
	tempDir := t.TempDir()
	return runner, mockExec, tempDir
}

func TestHookRunner_Run_NoConfigFile_ReturnsNotExecuted(t *testing.T) {
	runner, _, _ := setupHookRunnerTest(t)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   "/tmp/worktree",
		ConfigFilePath: "/nonexistent/.twiggit.toml",
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Executed)
	assert.True(t, result.Success)
	assert.Nil(t, result.Failures)
}

func TestHookRunner_Run_EmptyConfigFile_ReturnsNotExecuted(t *testing.T) {
	_, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	runner := NewHookRunner(mockExec)
	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Executed)
	assert.True(t, result.Success)
}

func TestHookRunner_Run_ConfigWithCommands_ExecutesCommands(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
commands = ["mise trust", "npm install"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, "sh", defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Twice()

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ProjectName:    "test-project",
		BranchName:     "feature",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, result.Executed)
	assert.True(t, result.Success)
	assert.Empty(t, result.Failures)
	mockExec.AssertExpectations(t)
}

func TestHookRunner_Run_CommandFailure_ContinuesAndCollectsFailures(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
commands = ["mise trust", "npm install", "echo done"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, "sh", defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, "sh", defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 1, Stdout: "npm error", Stderr: ""}, nil).Once()

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, "sh", defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.True(t, result.Executed)
	assert.False(t, result.Success)
	assert.Len(t, result.Failures, 1)
	assert.Equal(t, "npm install", result.Failures[0].Command)
	assert.Equal(t, 1, result.Failures[0].ExitCode)
}

func TestHookRunner_Run_MalformedTOML_LogsWarningAndReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create
commands = ["mise trust"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Executed)
	assert.True(t, result.Success)
}

func TestHookRunner_Run_MissingCommandsArray_ReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Executed)
	assert.True(t, result.Success)
}

func TestHookRunner_Run_EmptyCommandsArray_ReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
commands = []
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Executed)
	assert.True(t, result.Success)
}

func TestHookRunner_Run_EnvironmentVariablesSet(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
commands = ["echo test"]
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	var capturedArgs []string
	mockExec.On("ExecuteWithTimeout",
		mock.Anything, "/worktree/path", "sh", defaultTimeout(), mock.AnythingOfType("[]string"),
	).Run(func(args mock.Arguments) {
		capturedArgs = args.Get(4).([]string)
	}).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil)

	req := &HookRunRequest{
		HookType:       domain.HookPostCreate,
		WorktreePath:   "/worktree/path",
		ProjectName:    "my-project",
		BranchName:     "feature-branch",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: configPath,
	}

	_, err = runner.Run(context.Background(), req)
	require.NoError(t, err)

	require.Len(t, capturedArgs, 2)
	fullCmd := capturedArgs[1]

	assert.Contains(t, fullCmd, "TWIGGIT_WORKTREE_PATH")
	assert.Contains(t, fullCmd, "/worktree/path")
	assert.Contains(t, fullCmd, "TWIGGIT_PROJECT_NAME")
	assert.Contains(t, fullCmd, "my-project")
	assert.Contains(t, fullCmd, "TWIGGIT_BRANCH_NAME")
	assert.Contains(t, fullCmd, "feature-branch")
	assert.Contains(t, fullCmd, "TWIGGIT_SOURCE_BRANCH")
	assert.Contains(t, fullCmd, "main")
	assert.Contains(t, fullCmd, "TWIGGIT_MAIN_REPO_PATH")
	assert.Contains(t, fullCmd, "/repo/main")
}

func defaultTimeout() time.Duration {
	return 30 * time.Second
}
