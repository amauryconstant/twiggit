package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookType_String(t *testing.T) {
	t.Run("unknown renders as unknown", func(t *testing.T) {
		is := assert.New(t)
		is.Equal("unknown", HookTypeUnknown.String())
		is.Equal("unknown", HookType(99).String())
	})
	t.Run("post-create wire form", func(t *testing.T) {
		is := assert.New(t)
		is.Equal("post-create", HookTypePostCreate.String())
	})
	t.Run("rebase and sync wire forms", func(t *testing.T) {
		is := assert.New(t)
		is.Equal("pre-rebase", HookTypePreRebase.String())
		is.Equal("post-rebase", HookTypePostRebase.String())
		is.Equal("post-sync", HookTypePostSync.String())
	})
}

func TestHookType_ZeroIsUnknown(t *testing.T) {
	is := assert.New(t)
	var h HookType
	is.Equal(HookTypeUnknown, h, "zero value of HookType must equal HookTypeUnknown per enum-unknown-zero")
	is.Equal("unknown", h.String())
}

func TestHookResult_Structure(t *testing.T) {
	t.Run("zero value of HookResult", func(t *testing.T) {
		is := assert.New(t)
		r := HookResult{}
		is.False(r.HasExecuted)
		is.False(r.IsSuccessful)
		is.Nil(r.Failures)
		is.Equal(HookTypeUnknown, r.HookType)
	})

	t.Run("populated fields round-trip", func(t *testing.T) {
		is := assert.New(t)
		failures := []HookFailure{{Command: "false", ExitCode: 1, Output: "nope"}}
		r := HookResult{
			HookType:     HookTypePostCreate,
			HasExecuted:  true,
			IsSuccessful: false,
			Failures:     failures,
		}
		is.Equal(HookTypePostCreate, r.HookType)
		is.True(r.HasExecuted)
		is.False(r.IsSuccessful)
		is.Equal(failures, r.Failures)
	})
}

func TestHookResult_NoOpResultShape(t *testing.T) {
	is := assert.New(t)

	r := HookResult{
		HookType:     HookTypePostCreate,
		HasExecuted:  false,
		IsSuccessful: true,
		Failures:     nil,
	}

	is.False(r.HasExecuted, "no commands run means HasExecuted is false")
	is.True(r.IsSuccessful, "vacuous success when nothing executed")
	is.Nil(r.Failures)
}

func TestHookResult_PartialFailureShape(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	r := HookResult{
		HookType:     HookTypePostCreate,
		HasExecuted:  true,
		IsSuccessful: false,
		Failures: []HookFailure{
			{Command: "npm install", ExitCode: 1, Output: "ENOENT"},
		},
	}

	must.Len(r.Failures, 1, "only one failed command recorded")
	is.True(r.HasExecuted)
	is.False(r.IsSuccessful)
	is.Equal("npm install", r.Failures[0].Command)
	is.Equal(1, r.Failures[0].ExitCode)
}

func TestHookFailure_ErrorEmbedding(t *testing.T) {
	is := assert.New(t)

	f := HookFailure{
		Command:  "false",
		ExitCode: 1,
		Output:   "boom",
	}

	wrapped := errors.New("wrapped hook failure: " + f.Output)
	is.Contains(wrapped.Error(), "boom")
}

func TestHookFailure_TimedOutField(t *testing.T) {
	is := assert.New(t)
	f := HookFailure{
		Command:  "sleep 60",
		ExitCode: -1,
		Output:   "killed",
		Error:    context.DeadlineExceeded,
		TimedOut: true,
	}
	is.True(f.TimedOut)
	is.ErrorIs(f.Error, context.DeadlineExceeded)
}

func TestHookDefinition_Fields(t *testing.T) {
	is := assert.New(t)
	def := HookDefinition{
		Command:          "mise trust && npm install",
		WorkingDirectory: "/tmp/wt",
		TimeoutSeconds:   60,
	}
	is.Equal("mise trust && npm install", def.Command)
	is.Equal("/tmp/wt", def.WorkingDirectory)
	is.Equal(60, def.TimeoutSeconds)
}

func TestHookConfig_PostCreateSlice(t *testing.T) {
	is := assert.New(t)
	cfg := HookConfig{
		PostCreate: []HookDefinition{
			{Command: "echo a"},
			{Command: "echo b", WorkingDirectory: "/p", TimeoutSeconds: 5},
		},
	}
	is.Len(cfg.PostCreate, 2)
	is.Equal("echo a", cfg.PostCreate[0].Command)
	is.Equal("/p", cfg.PostCreate[1].WorkingDirectory)
	is.Equal(5, cfg.PostCreate[1].TimeoutSeconds)
}

func TestHookConfig_NewVariants(t *testing.T) {
	is := assert.New(t)
	cfg := HookConfig{
		PreRebase:  []HookDefinition{{Command: "git stash"}},
		PostRebase: []HookDefinition{{Command: "go build ./..."}},
		PostSync:   []HookDefinition{{Command: "echo synced"}},
	}
	is.Len(cfg.PreRebase, 1)
	is.Equal("git stash", cfg.PreRebase[0].Command)
	is.Len(cfg.PostRebase, 1)
	is.Equal("go build ./...", cfg.PostRebase[0].Command)
	is.Len(cfg.PostSync, 1)
	is.Equal("echo synced", cfg.PostSync[0].Command)
	is.Empty(cfg.PostCreate, "post-create must remain empty when not set")
}

func TestHookRunRequest_RebaseFields(t *testing.T) {
	is := assert.New(t)
	req := HookRunRequest{
		HookType:     HookTypePostRebase,
		RebaseBase:   "main",
		RebaseOldTip: "abc",
		RebaseNewTip: "def",
		RebaseResult: "clean",
	}
	is.Equal("main", req.RebaseBase)
	is.Equal("abc", req.RebaseOldTip)
	is.Equal("def", req.RebaseNewTip)
	is.Equal("clean", req.RebaseResult)
}

func TestHookRunRequest_SyncFields(t *testing.T) {
	is := assert.New(t)
	req := HookRunRequest{
		HookType:   HookTypePostSync,
		SyncRemote: "origin",
		SyncBranch: "main",
	}
	is.Equal("origin", req.SyncRemote)
	is.Equal("main", req.SyncBranch)
}

func TestHookRunRequest_PostCreateCompilesWithoutRebaseFields(t *testing.T) {
	is := assert.New(t)
	req := HookRunRequest{
		HookType:     HookTypePostCreate,
		WorktreePath: "/wt",
		ProjectName:  "p",
		BranchName:   "feature",
		SourceBranch: "main",
		MainRepoPath: "/repo",
	}
	is.Empty(req.RebaseBase, "RebaseBase must default empty for PostCreate")
	is.Empty(req.SyncRemote, "SyncRemote must default empty for PostCreate")
}
