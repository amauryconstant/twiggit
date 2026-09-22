package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookResult_Structure(t *testing.T) {
	t.Run("zero value of HookResult", func(t *testing.T) {
		is := assert.New(t)
		r := HookResult{}
		is.False(r.HasExecuted)
		is.False(r.IsSuccessful)
		is.Nil(r.Failures)
		is.Empty(r.HookType)
	})

	t.Run("populated fields round-trip", func(t *testing.T) {
		is := assert.New(t)
		failures := []HookFailure{{Command: "false", ExitCode: 1, Output: "nope"}}
		r := HookResult{
			HookType:     HookPostCreate,
			HasExecuted:  true,
			IsSuccessful: false,
			Failures:     failures,
		}
		is.Equal(HookPostCreate, r.HookType)
		is.True(r.HasExecuted)
		is.False(r.IsSuccessful)
		is.Equal(failures, r.Failures)
	})
}

func TestHookResult_NoOpResultShape(t *testing.T) {
	is := assert.New(t)

	r := HookResult{
		HookType:     HookPostCreate,
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
	require := require.New(t)

	r := HookResult{
		HookType:     HookPostCreate,
		HasExecuted:  true,
		IsSuccessful: false,
		Failures: []HookFailure{
			{Command: "npm install", ExitCode: 1, Output: "ENOENT"},
		},
	}

	require.Len(r.Failures, 1, "only one failed command recorded")
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
