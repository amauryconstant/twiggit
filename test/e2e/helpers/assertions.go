//go:build e2e
// +build e2e

package helpers

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// TwiggitAssertions provides domain-specific assertions for E2E tests
type TwiggitAssertions struct {
	cli *TwiggitCLI
}

// NewTwiggitAssertions creates a new TwiggitAssertions instance
func NewTwiggitAssertions() *TwiggitAssertions {
	return &TwiggitAssertions{
		cli: &TwiggitCLI{},
	}
}

// ShouldOutputWorktreeList asserts output matches worktree list format
func (a *TwiggitAssertions) ShouldOutputWorktreeList(session *gexec.Session, expectedBranches []string) {
	a.cli.ShouldSucceed(session)

	for _, branch := range expectedBranches {
		a.cli.ShouldOutput(session, branch)
	}
}

// ShouldHaveWorktree asserts worktree exists in filesystem
func (a *TwiggitAssertions) ShouldHaveWorktree(worktreePath string) {
	Expect(worktreePath).To(BeADirectory())
}

// ShouldCreateWorktree asserts worktree was created successfully
func (a *TwiggitAssertions) ShouldCreateWorktree(session *gexec.Session, branch string) {
	a.cli.ShouldSucceed(session)
	a.cli.ShouldOutput(session, branch)
}

// ShouldDeleteWorktree asserts worktree was deleted successfully
func (a *TwiggitAssertions) ShouldDeleteWorktree(session *gexec.Session, branch string) {
	a.cli.ShouldSucceed(session)
	a.cli.ShouldOutput(session, branch)
}

// ShouldListProjects asserts projects are listed correctly
func (a *TwiggitAssertions) ShouldListProjects(session *gexec.Session, expectedProjects []string) {
	a.cli.ShouldSucceed(session)

	for _, project := range expectedProjects {
		a.cli.ShouldOutput(session, project)
	}
}

// ShouldFailWithWorktreeError asserts command fails with worktree-related error
func (a *TwiggitAssertions) ShouldFailWithWorktreeError(session *gexec.Session, errorMsg string) {
	a.cli.ShouldFailWithExit(session, 1)
	a.cli.ShouldErrorOutput(session, errorMsg)
}

// ShouldOutputBranch asserts branch name appears in output
func (a *TwiggitAssertions) ShouldOutputBranch(session *gexec.Session, branch string) {
	a.cli.ShouldOutput(session, branch)
}

// ShouldChangeDirectory asserts cd command changes directory
func (a *TwiggitAssertions) ShouldChangeDirectory(session *gexec.Session, path string) {
	a.cli.ShouldSucceed(session)
	a.cli.ShouldOutput(session, path)
}
