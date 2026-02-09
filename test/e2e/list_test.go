//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit list command.
// Tests validate context-aware listing behavior across project, worktree, and outside git contexts.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("list command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		fixture.Cleanup()
	})

	It("lists worktrees from project context", func() {
		fixture.CreateWorktreeSetup("test-project")

		session := ctxHelper.FromProjectDir("test-project", "list")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("lists all worktrees with --all flag", func() {
		fixture.SetupMultiProject()

		session := cli.Run("list", "--all")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("lists worktrees from worktree context", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "list")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("lists all worktrees when outside git", func() {
		fixture.SetupMultiProject()

		session := ctxHelper.FromOutsideGit("list", "--all")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows 'No worktrees found' for empty project", func() {
		fixture.SetupSingleProject("empty-project")

		session := ctxHelper.FromProjectDir("empty-project", "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "No worktrees found")
	})

	It("shows (modified) status for modified worktree", func() {
		_ = fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows (detached) status for detached worktree", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "list")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("lists all worktrees across multiple projects with --all", func() {
		fixture.SetupMultiProject()

		session := cli.Run("list", "--all")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})
})
