package git

import (
	"context"
	"errors"
	"testing"
	"time"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var _ = errors.Is

func newRebaserWithMock(t *testing.T) (*cliClient, *MockCommandExecutor) {
	t.Helper()
	mockExec := NewMockCommandExecutor()
	client := &cliClient{
		executor:       mockExec,
		defaultTimeout: 30 * time.Second,
	}
	return client, mockExec
}

func TestRebaser_StderrPhrases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		fn     func(string) bool
		input  string
		expect bool
	}{
		{"conflict-could-not-apply", isRebaseConflictStderr, "error: could not apply 123abc...", true},
		{"conflict-resolve-manually", isRebaseConflictStderr, "Resolve all conflicts manually, then run git rebase --continue", true},
		{"conflict-CONFLICT-token", isRebaseConflictStderr, "CONFLICT (content): Merge conflict in foo.txt", true},
		{"conflict-missing", isRebaseConflictStderr, "fatal: bad revision", false},

		{"nothing-to-do-stderr", isNothingToDoStderr, "Your branch is up to date with 'origin/main'", true},
		{"nothing-to-do-stdout", isNothingToDoStdout, "Current branch main is up to date.", true},
		{"nothing-to-do-missing", isNothingToDoStderr, "CONFLICT", false},

		{"no-rebase-in-progress", isNoRebaseInProgressStderr, "fatal: No rebase in progress", true},
		{"no-rebase-in-progress-lower", isNoRebaseInProgressStderr, "fatal: no rebase in progress", true},
		{"no-rebase-missing", isNoRebaseInProgressStderr, "nothing to commit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, tt.fn(tt.input))
		})
	}
}

func TestCliClient_Rebase_Clean(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "main"}).Return(&CommandResult{ExitCode: 0, Stdout: "First, rewinding head to replay your work on top of it...\nApplying: feat\n"}, nil)

	outcome, err := client.Rebase(context.Background(), "/wt", "main")
	require.NoError(t, err)
	assert.Equal(t, core.RebaseOutcomeClean, outcome)
}

func TestCliClient_Rebase_Conflict(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "main"}).Return(&CommandResult{ExitCode: 1, Stderr: "error: could not apply abc1234... feat"}, nil)

	outcome, err := client.Rebase(context.Background(), "/wt", "main")
	require.Error(t, err)
	assert.Equal(t, core.RebaseOutcomeConflicted, outcome)
	assert.ErrorIs(t, err, core.ErrRebaseConflict,
		"rebase conflict must wrap ErrRebaseConflict")
}

func TestCliClient_Rebase_NothingToDo(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "main"}).Return(&CommandResult{ExitCode: 0, Stdout: "Current branch main is up to date."}, nil)

	outcome, err := client.Rebase(context.Background(), "/wt", "main")
	require.NoError(t, err)
	assert.Equal(t, core.RebaseOutcomeNothingToDo, outcome)
}

func TestCliClient_Rebase_EmptyArgs(t *testing.T) {
	client, _ := newRebaserWithMock(t)

	_, err := client.Rebase(context.Background(), "", "main")
	require.Error(t, err)
	assert.NotErrorIs(t, err, core.ErrRebaseConflict, "empty wtPath should not be a conflict")

	_, err = client.Rebase(context.Background(), "/wt", "")
	require.Error(t, err)
}

func TestCliClient_Abort(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "--abort"}).Return(&CommandResult{ExitCode: 0}, nil)

	err := client.Abort(context.Background(), "/wt")
	require.NoError(t, err)
}

func TestCliClient_Abort_EmptyPath(t *testing.T) {
	client, _ := newRebaserWithMock(t)
	err := client.Abort(context.Background(), "")
	require.Error(t, err)
}

func TestCliClient_Continue_Clean(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "--continue"}).Return(&CommandResult{ExitCode: 0, Stdout: "Applying: feat\n"}, nil)

	outcome, err := client.Continue(context.Background(), "/wt")
	require.NoError(t, err)
	assert.Equal(t, core.RebaseOutcomeClean, outcome)
}

func TestCliClient_Continue_NoRebase(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"rebase", "--continue"}).Return(&CommandResult{ExitCode: 1, Stderr: "fatal: No rebase in progress"}, nil)

	outcome, err := client.Continue(context.Background(), "/wt")
	require.Error(t, err)
	assert.Equal(t, core.RebaseOutcomeAborted, outcome)
	assert.ErrorIs(t, err, core.ErrRebaseInProgress,
		"continue with no rebase must wrap ErrRebaseInProgress")
}

func TestCliClient_Fetch(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/repo", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"fetch", "origin", "main"}).Return(&CommandResult{ExitCode: 0}, nil)

	err := client.Fetch(context.Background(), "/repo", "origin", "main")
	require.NoError(t, err)
}

func TestCliClient_Fetch_EmptyArgs(t *testing.T) {
	client, _ := newRebaserWithMock(t)

	require.Error(t, client.Fetch(context.Background(), "", "origin", "main"))
	require.Error(t, client.Fetch(context.Background(), "/repo", "", "main"))
	require.Error(t, client.Fetch(context.Background(), "/repo", "origin", ""))
}

func TestCliClient_SetTrackedBase(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"config", "--worktree", "twiggit.tracked-base", "main"}).Return(&CommandResult{ExitCode: 0}, nil)

	err := client.SetTrackedBase(context.Background(), "/wt", "main")
	require.NoError(t, err)
}

func TestCliClient_SetTrackedBase_EmptyArgs(t *testing.T) {
	client, _ := newRebaserWithMock(t)

	require.Error(t, client.SetTrackedBase(context.Background(), "", "main"))
	require.Error(t, client.SetTrackedBase(context.Background(), "/wt", ""))
}

func TestCliClient_GetTrackedBase_Present(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"config", "--worktree", "--get", "twiggit.tracked-base"}).Return(&CommandResult{ExitCode: 0, Stdout: "main\n"}, nil)

	base, err := client.GetTrackedBase(context.Background(), "/wt")
	require.NoError(t, err)
	assert.Equal(t, "main", base)
}

func TestCliClient_GetTrackedBase_Absent(t *testing.T) {
	client, mockExec := newRebaserWithMock(t)
	t.Cleanup(func() { mockExec.AssertExpectations(t) })

	mockExec.On("ExecuteWithTimeout", mock.Anything, "/wt", CmdGit, mock.AnythingOfType("time.Duration"),
		[]string{"config", "--worktree", "--get", "twiggit.tracked-base"}).Return(&CommandResult{ExitCode: 1}, nil)

	base, err := client.GetTrackedBase(context.Background(), "/wt")
	require.NoError(t, err)
	assert.Empty(t, base, "absent tracked-base must return empty string, not error")
}

func TestCliClient_GetTrackedBase_EmptyPath(t *testing.T) {
	client, _ := newRebaserWithMock(t)
	_, err := client.GetTrackedBase(context.Background(), "")
	require.Error(t, err)
}
