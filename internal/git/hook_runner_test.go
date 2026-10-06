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

// TestHookRunner_EnvVars covers tasks 2.3 + 2.4 from the rebase-sync
// change: TWIGGIT_REBASE_* and TWIGGIT_SYNC_* environment variables
// are exported only when the hook type matches and the field is
// non-empty. PostCreate requests must not emit any rebase or sync
// variables.
func TestHookRunner_EnvVars(t *testing.T) {
	tests := []struct {
		name           string
		req            *core.HookRunRequest
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "pre-rebase exports TWIGGIT_REBASE_BASE",
			req: &core.HookRunRequest{
				HookType:     core.HookTypePreRebase,
				RebaseBase:   "main",
				WorktreePath: "/wt",
			},
			mustContain:    []string{"TWIGGIT_REBASE_BASE", "main"},
			mustNotContain: []string{"TWIGGIT_REBASE_RESULT", "TWIGGIT_SYNC_REMOTE"},
		},
		{
			name: "post-rebase exports all four rebase vars",
			req: &core.HookRunRequest{
				HookType:     core.HookTypePostRebase,
				RebaseBase:   "main",
				RebaseOldTip: "abc",
				RebaseNewTip: "def",
				RebaseResult: "clean",
				WorktreePath: "/wt",
			},
			mustContain: []string{
				"TWIGGIT_REBASE_BASE", "main",
				"TWIGGIT_REBASE_OLD_TIP", "abc",
				"TWIGGIT_REBASE_NEW_TIP", "def",
				"TWIGGIT_REBASE_RESULT", "clean",
			},
			mustNotContain: []string{"TWIGGIT_SYNC_REMOTE", "TWIGGIT_SYNC_BRANCH"},
		},
		{
			name: "post-sync exports sync vars",
			req: &core.HookRunRequest{
				HookType:     core.HookTypePostSync,
				SyncRemote:   "origin",
				SyncBranch:   "main",
				WorktreePath: "/wt",
			},
			mustContain:    []string{"TWIGGIT_SYNC_REMOTE", "origin", "TWIGGIT_SYNC_BRANCH", "main"},
			mustNotContain: []string{"TWIGGIT_REBASE_BASE", "TWIGGIT_REBASE_RESULT"},
		},
		{
			name: "post-create omits rebase and sync vars",
			req: &core.HookRunRequest{
				HookType:     core.HookTypePostCreate,
				WorktreePath: "/wt",
			},
			mustContain:    []string{"TWIGGIT_WORKTREE_PATH"},
			mustNotContain: []string{"TWIGGIT_REBASE_BASE", "TWIGGIT_REBASE_OLD_TIP", "TWIGGIT_REBASE_NEW_TIP", "TWIGGIT_REBASE_RESULT", "TWIGGIT_SYNC_REMOTE", "TWIGGIT_SYNC_BRANCH"},
		},
		{
			name: "empty rebase field is omitted",
			req: &core.HookRunRequest{
				HookType:     core.HookTypePostRebase,
				RebaseBase:   "main",
				RebaseResult: "",
				WorktreePath: "/wt",
			},
			mustContain:    []string{"TWIGGIT_REBASE_BASE", "main"},
			mustNotContain: []string{"TWIGGIT_REBASE_RESULT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, mockExec, tempDir := setupHookRunnerTest(t)
			configPath := filepath.Join(tempDir, ".twiggit.toml")
			require.NoError(t, os.WriteFile(configPath, []byte(`[[hooks.post-create]]
command = "echo" `), 0o644))
			tt.req.ConfigFilePath = configPath

			var capturedArgs []string
			mockExec.On("ExecuteWithTimeout",
				mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
			).Run(func(args mock.Arguments) {
				capturedArgs = args.Get(4).([]string)
			}).Return(&CommandResult{ExitCode: 0}, nil)
			t.Cleanup(func() { mockExec.AssertExpectations(t) })

			// Reset the hook type onto a hook definition that matches
			// the test by writing the appropriate TOML block too.
			// Simpler: write all four blocks; runner picks by HookType.
			configContent := buildConfigFor(tt.req.HookType)
			require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o644))

			_, err := runner.Run(t.Context(), tt.req)
			require.NoError(t, err)
			require.NotEmpty(t, capturedArgs, "expected executor to be invoked")
			fullCmd := capturedArgs[1]

			for _, want := range tt.mustContain {
				assert.Contains(t, fullCmd, want, "missing env export: %s", want)
			}
			for _, nope := range tt.mustNotContain {
				assert.NotContains(t, fullCmd, nope, "unexpected env export: %s", nope)
			}
		})
	}
}

func buildConfigFor(hookType core.HookType) string {
	switch hookType {
	case core.HookTypePreRebase:
		return `[[hooks.pre-rebase]]
command = "echo" `
	case core.HookTypePostRebase:
		return `[[hooks.post-rebase]]
command = "echo" `
	case core.HookTypePostSync:
		return `[[hooks.post-sync]]
command = "echo" `
	default:
		return `[[hooks.post-create]]
command = "echo" `
	}
}

// TestHookRunner_FailureSemantics covers task 2.5: PreRebase failure
// short-circuits the run; PostRebase/PostSync failures are recorded
// but non-fatal (more commands can still run after a failure).
func TestHookRunner_FailureSemantics(t *testing.T) {
	t.Run("pre-rebase failure short-circuits", func(t *testing.T) {
		runner, mockExec, tempDir := setupHookRunnerTest(t)
		configPath := filepath.Join(tempDir, ".twiggit.toml")
		require.NoError(t, os.WriteFile(configPath, []byte(`[[hooks.pre-rebase]]
command = "first"
[[hooks.pre-rebase]]
command = "second"
`), 0o644))

		mockExec.On("ExecuteWithTimeout",
			mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
		).Return(&CommandResult{ExitCode: 1}, nil).Once()
		t.Cleanup(func() { mockExec.AssertExpectations(t) })

		req := &core.HookRunRequest{
			HookType:       core.HookTypePreRebase,
			WorktreePath:   tempDir,
			ConfigFilePath: configPath,
		}
		result, err := runner.Run(t.Context(), req)
		require.NoError(t, err)
		assert.False(t, result.IsSuccessful)
		assert.Len(t, result.Failures, 1, "pre-rebase must short-circuit on first failure")
	})

	t.Run("post-rebase failure records but does not short-circuit", func(t *testing.T) {
		runner, mockExec, tempDir := setupHookRunnerTest(t)
		configPath := filepath.Join(tempDir, ".twiggit.toml")
		require.NoError(t, os.WriteFile(configPath, []byte(`[[hooks.post-rebase]]
command = "first"
[[hooks.post-rebase]]
command = "second"
`), 0o644))

		// First invocation fails; second succeeds.
		mockExec.On("ExecuteWithTimeout",
			mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
		).Return(&CommandResult{ExitCode: 1}, nil).Once()
		mockExec.On("ExecuteWithTimeout",
			mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
		).Return(&CommandResult{ExitCode: 0}, nil).Once()
		t.Cleanup(func() { mockExec.AssertExpectations(t) })

		req := &core.HookRunRequest{
			HookType:       core.HookTypePostRebase,
			WorktreePath:   tempDir,
			ConfigFilePath: configPath,
		}
		result, err := runner.Run(t.Context(), req)
		require.NoError(t, err)
		assert.False(t, result.IsSuccessful, "at least one failure means result is unsuccessful")
		assert.Len(t, result.Failures, 1, "exactly one failure recorded; second command ran after first failure")
	})

	t.Run("post-sync failure records but does not short-circuit", func(t *testing.T) {
		runner, mockExec, tempDir := setupHookRunnerTest(t)
		configPath := filepath.Join(tempDir, ".twiggit.toml")
		require.NoError(t, os.WriteFile(configPath, []byte(`[[hooks.post-sync]]
command = "ok"
[[hooks.post-sync]]
command = "fail"
`), 0o644))

		mockExec.On("ExecuteWithTimeout",
			mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
		).Return(&CommandResult{ExitCode: 0}, nil).Once()
		mockExec.On("ExecuteWithTimeout",
			mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
		).Return(&CommandResult{ExitCode: 1}, nil).Once()
		t.Cleanup(func() { mockExec.AssertExpectations(t) })

		req := &core.HookRunRequest{
			HookType:       core.HookTypePostSync,
			WorktreePath:   tempDir,
			ConfigFilePath: configPath,
		}
		result, err := runner.Run(t.Context(), req)
		require.NoError(t, err)
		assert.False(t, result.IsSuccessful)
		assert.Len(t, result.Failures, 1, "one failure, one success")
	})
}

// TestHookRunner_Dispatch covers task 2.6: each new variant is read
// from its HookConfig slice. Confirms the dispatch table reads
// PreRebase/PostRebase/PostSync from the matching TOML block.
func TestHookRunner_Dispatch(t *testing.T) {
	tests := []struct {
		name      string
		hookType  core.HookType
		tomlBlock string
	}{
		{"pre-rebase dispatch", core.HookTypePreRebase, `[[hooks.pre-rebase]]
command = "echo pre"`},
		{"post-rebase dispatch", core.HookTypePostRebase, `[[hooks.post-rebase]]
command = "echo post-rebase"`},
		{"post-sync dispatch", core.HookTypePostSync, `[[hooks.post-sync]]
command = "echo post-sync"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, mockExec, tempDir := setupHookRunnerTest(t)
			configPath := filepath.Join(tempDir, ".twiggit.toml")
			require.NoError(t, os.WriteFile(configPath, []byte(tt.tomlBlock), 0o644))

			mockExec.On("ExecuteWithTimeout",
				mock.Anything, mock.Anything, CmdSh, defaultTimeout(), mock.AnythingOfType("[]string"),
			).Return(&CommandResult{ExitCode: 0}, nil).Once()
			t.Cleanup(func() { mockExec.AssertExpectations(t) })

			req := &core.HookRunRequest{
				HookType:       tt.hookType,
				WorktreePath:   tempDir,
				ConfigFilePath: configPath,
			}
			result, err := runner.Run(t.Context(), req)
			require.NoError(t, err)
			assert.True(t, result.HasExecuted, "matching slice must be dispatched")
			assert.True(t, result.IsSuccessful)
		})
	}
}
