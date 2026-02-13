//go:build e2e
// +build e2e

package fixtures

import (
	"io"
	"path/filepath"

	"github.com/onsi/gomega/gexec"

	e2ehelpers "twiggit/test/e2e/helpers"
)

// ContextHelper provides utilities for testing context-aware behavior
// Tests commands from different git contexts: project directory, worktree directory, outside git
type ContextHelper struct {
	fixture *E2ETestFixture
	cli     *e2ehelpers.TwiggitCLI
}

// NewContextHelper creates a new ContextHelper instance
func NewContextHelper(fixture *E2ETestFixture, cli *e2ehelpers.TwiggitCLI) *ContextHelper {
	return &ContextHelper{
		fixture: fixture,
		cli:     cli,
	}
}

// FromProjectDir runs command from within a project directory
// Use this to test commands that should infer project name from current directory
func (h *ContextHelper) FromProjectDir(projectName string, args ...string) *gexec.Session {
	projectPath := h.fixture.GetProjectPath(projectName)
	return h.cli.RunWithDir(projectPath, args...)
}

// FromWorktreeDir runs command from within a worktree directory
// Use this to test commands that should detect worktree context
func (h *ContextHelper) FromWorktreeDir(projectName, branch string, args ...string) *gexec.Session {
	worktreePath := filepath.Join(
		h.fixture.GetConfigHelper().GetWorktreesDir(),
		projectName,
		branch,
	)
	return h.cli.RunWithDir(worktreePath, args...)
}

// FromOutsideGit runs command from outside any git repository
// Use this to test commands that require explicit project names
func (h *ContextHelper) FromOutsideGit(args ...string) *gexec.Session {
	return h.cli.RunWithDir(h.fixture.GetTempDir(), args...)
}

// FromOutsideGitWithStdin runs command from outside git with stdin input
// Use this for testing interactive prompts like confirmation dialogs
func (h *ContextHelper) FromOutsideGitWithStdin(stdin io.Reader, args ...string) *gexec.Session {
	return h.cli.RunWithStdinAndDir(h.fixture.GetTempDir(), stdin, args...)
}

// WithConfigDir returns a new ContextHelper with the specified config directory
func (h *ContextHelper) WithConfigDir(configDir string) *ContextHelper {
	return &ContextHelper{
		fixture: h.fixture,
		cli:     h.cli.WithConfigDir(configDir),
	}
}

// FromProjectDirWithDebug runs command from within a project directory with debug logging
// Logs session details and fixture state if exit code != 0
func (h *ContextHelper) FromProjectDirWithDebug(projectName string, args ...string) *gexec.Session {
	session := h.FromProjectDir(projectName, args...)
	h.cli.DebugSession(session, h.fixture.Inspect())
	return session
}

// FromWorktreeDirWithDebug runs command from within a worktree directory with debug logging
// Logs session details and fixture state if exit code != 0
func (h *ContextHelper) FromWorktreeDirWithDebug(projectName, branch string, args ...string) *gexec.Session {
	session := h.FromWorktreeDir(projectName, branch, args...)
	h.cli.DebugSession(session, h.fixture.Inspect())
	return session
}

// FromOutsideGitWithDebug runs command from outside any git repository with debug logging
// Logs session details and fixture state if exit code != 0
func (h *ContextHelper) FromOutsideGitWithDebug(args ...string) *gexec.Session {
	session := h.FromOutsideGit(args...)
	h.cli.DebugSession(session, h.fixture.Inspect())
	return session
}
