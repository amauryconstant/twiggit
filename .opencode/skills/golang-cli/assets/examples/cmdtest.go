// Path: internal/cmdutil/cmdtest/cmdtest.go
// Shared command-test setup, imported by _test.go files only.
package cmdtest

import (
	"bytes"
	"context"
	"testing"

	"github.com/you/myapp/internal/cmdutil"
	"github.com/you/myapp/internal/config"
	"github.com/you/myapp/internal/core"
	"github.com/you/myapp/internal/iostreams"
)

type Env struct {
	IO             *iostreams.IOStreams
	Stdout, Stderr *bytes.Buffer
	Factory        *cmdutil.Factory
	Worktrees      *FakeWorktrees
}

func New(t *testing.T) *Env {
	t.Helper()
	ios, _, stdout, stderr := iostreams.Test()
	return &Env{IO: ios, Stdout: stdout, Stderr: stderr, Worktrees: &FakeWorktrees{},
		Factory: &cmdutil.Factory{
			IOStreams: ios,
			Config:    func() (*config.AppConfig, error) { return config.DefaultConfig(), nil },
		}}
}

// FakeWorktrees is hand-written and state-based; it satisfies each command's
// consumer interface (worktreeCreator in cmd/create.go).
type FakeWorktrees struct {
	Created []core.Worktree
	Err     error
}

func (f *FakeWorktrees) CreateWorktree(_ context.Context, name core.BranchName, source string) (*core.Worktree, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	wt := core.Worktree{Name: name.String(), Branch: source, Path: "/tmp/wt/" + name.String()}
	f.Created = append(f.Created, wt)
	return &wt, nil
}
