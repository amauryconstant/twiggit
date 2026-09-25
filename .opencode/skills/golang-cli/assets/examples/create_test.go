// Path: cmd/create_test.go
package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/core"
	"github.com/you/myapp/internal/iostreams"
)

// Parse level: runF captures the options; nothing touches git.
func TestNewCmdCreate(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	var got *CreateOptions
	cmd := NewCmdCreate(&cmdutil.Factory{IOStreams: ios}, func(o *CreateOptions) error {
		got = o
		return nil
	})

	cmd.SetArgs([]string{"feature/login", "--source", "develop"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "feature/login", got.Name)
	assert.Equal(t, "develop", got.Source)
}

type fakeCreator struct{}

func (fakeCreator) CreateWorktree(_ context.Context, name core.BranchName, _ string) (*core.Worktree, error) {
	return &core.Worktree{Name: name.String(), Path: "/wt/" + name.String()}, nil
}

// Run level: the real runCreate against a fake adapter, output captured.
func TestRunCreate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOut     string
		wantInvalid bool
	}{
		{name: "piped output is the bare path", input: "feature/login", wantOut: "/wt/feature/login\n"},
		{name: "invalid name is a validation error", input: "", wantInvalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, _ := iostreams.Test()
			err := runCreate(&CreateOptions{
				IO:     ios,
				Ctx:    context.Background(),
				Client: func() (worktreeCreator, error) { return fakeCreator{}, nil },
				Name:   tt.input,
			})

			if tt.wantInvalid {
				var invalid *core.ValidationError
				require.ErrorAs(t, err, &invalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOut, stdout.String())
		})
	}
}
