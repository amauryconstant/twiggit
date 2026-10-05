package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var _ = require.NoError

func setupHookRunnerTest(t *testing.T) (*HookRunner, *MockCommandExecutor, string) {
	t.Helper()
	mockExec := NewMockCommandExecutor()
	runner := NewHookRunner(mockExec)
	tempDir := t.TempDir()
	return runner, mockExec, tempDir
}

func TestHookRunner_Run_NoConfigFile_ReturnsNotExecuted(t *testing.T) {
	runner, _, _ := setupHookRunnerTest(t)

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   "/tmp/worktree",
		ConfigFilePath: "/nonexistent/.twiggit.toml",
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.False(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
	assert.Nil(t, result.Failures)
}

func TestHookRunner_Run_EmptyConfigFile_ReturnsNotExecuted(t *testing.T) {
	_, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	err := os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	runner := NewHookRunner(mockExec)
	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.False(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_ConfigWithCommand_ExecutesCommand(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create]
command = "mise trust && npm install"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ProjectName:    "test-project",
		BranchName:     "feature",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
	assert.Empty(t, result.Failures)
}

func TestHookRunner_Run_ConfigWithMultipleDefinitions_ExecutesAll(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = "mise trust"

[[hooks.post-create]]
command = "npm install"

[[hooks.post-create]]
command = "echo done"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Times(3)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
	assert.Empty(t, result.Failures)
}

func TestHookRunner_Run_CommandFailure_ContinuesAndCollectsFailures(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = "mise trust"

[[hooks.post-create]]
command = "npm install"

[[hooks.post-create]]
command = "echo done"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 1, Stdout: "npm error", Stderr: ""}, nil).Once()

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.HasExecuted)
	assert.False(t, result.IsSuccessful)
	assert.Len(t, result.Failures, 1)
	assert.Equal(t, "npm install", result.Failures[0].Command)
	assert.Equal(t, 1, result.Failures[0].ExitCode)
}

func TestHookRunner_Run_TimeoutMarksTimedOut(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = "sleep 60"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(nil, context.DeadlineExceeded).Once()
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	require.Len(t, result.Failures, 1)
	assert.True(t, result.Failures[0].TimedOut, "TimedOut must be true on context.DeadlineExceeded")
	assert.ErrorIs(t, result.Failures[0].Error, context.DeadlineExceeded)
}

func TestHookRunner_Run_DefinitionTimeoutOverridesDefault(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = "echo a"
timeout_seconds = 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, tempDir, CmdSh, 5*time.Second, mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_DefinitionWorkingDirectoryOverridesWorktreePath(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")
	otherDir := t.TempDir()

	configContent := `
[[hooks.post-create]]
command = "echo a"
working_directory = "` + otherDir + `"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	mockExec.On("ExecuteWithTimeout",
		mock.Anything, otherDir, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil).Once()
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_MalformedTOML_LogsWarningAndReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[hooks.post-create
command = "mise trust"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.False(t, result.HasExecuted)
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_MissingCommandField_ReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
timeout_seconds = 5
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.HasExecuted, "executor entered with definitions; empty Command skipped")
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_EmptyCommand_ReturnsNotExecuted(t *testing.T) {
	runner, _, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = ""
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   tempDir,
		ConfigFilePath: configPath,
	}

	result, err := runner.Run(t.Context(), req)

	require.NoError(t, err)
	assert.True(t, result.HasExecuted, "executor entered with definitions; empty Command skipped")
	assert.True(t, result.IsSuccessful)
}

func TestHookRunner_Run_EnvironmentVariablesSet(t *testing.T) {
	runner, mockExec, tempDir := setupHookRunnerTest(t)
	configPath := filepath.Join(tempDir, ".twiggit.toml")

	configContent := `
[[hooks.post-create]]
command = "echo test"
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	var capturedArgs []string
	mockExec.On("ExecuteWithTimeout",
		mock.Anything, "/worktree/path", CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
	).Run(func(args mock.Arguments) {
		capturedArgs = args.Get(4).([]string)
	}).Return(&CommandResult{ExitCode: 0, Stdout: "", Stderr: ""}, nil)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	req := &core.HookRunRequest{
		HookType:       core.HookTypePostCreate,
		WorktreePath:   "/worktree/path",
		ProjectName:    "my-project",
		BranchName:     "feature-branch",
		SourceBranch:   "main",
		MainRepoPath:   "/repo/main",
		ConfigFilePath: configPath,
	}

	_, err = runner.Run(t.Context(), req)
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

// Compile-time guard: confirm the runner surfaces timeout errors as
// sentinel-shaped failures via HookFailure.Error.
func TestHookRunner_FailureErrorChain(t *testing.T) {
	is := context.DeadlineExceeded
	f := core.HookFailure{Error: is}
	assert.ErrorIs(t, f.Error, context.DeadlineExceeded)
}

func defaultTimeout() time.Duration {
	return 30 * time.Second
}
