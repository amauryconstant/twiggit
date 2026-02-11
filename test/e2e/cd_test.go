//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit cd command.
// Tests validate context-aware navigation between projects and worktrees.
package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("cd command", func() {
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

	It("cd from project to worktree", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "cd", result.Feature1Branch)
		Eventually(session).Should(gexec.Exit(0))

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		Expect(string(session.Out.Contents())).To(ContainSubstring(worktreePath))
	})

	It("cd from worktree to different worktree", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "cd", result.Feature2Branch)
		Eventually(session).Should(gexec.Exit(0))

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature2Branch)
		Expect(string(session.Out.Contents())).To(ContainSubstring(worktreePath))
	})

	It("cd from worktree to main project", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "cd", "main")
		Eventually(session).Should(gexec.Exit(0))

		projectPath := fixture.GetProjectPath("test")
		Expect(string(session.Out.Contents())).To(ContainSubstring(projectPath))
	})

	It("cd from outside git to project", func() {
		fixture.SetupSingleProject("test-project")

		session := ctxHelper.FromOutsideGit("cd", "test-project")
		Eventually(session).Should(gexec.Exit(0))

		projectPath := fixture.GetProjectPath("test-project")
		Expect(string(session.Out.Contents())).To(ContainSubstring(projectPath))
	})

	It("cd from outside git to cross-project worktree", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromOutsideGit("cd", "test/"+result.Feature1Branch)
		Eventually(session).Should(gexec.Exit(0))

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		Expect(string(session.Out.Contents())).To(ContainSubstring(worktreePath))
	})

	It("cd with no target from project context", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "cd")
		Eventually(session).Should(gexec.Exit(0))

		projectPath := fixture.GetProjectPath("test")
		Expect(string(session.Out.Contents())).To(ContainSubstring(projectPath))
	})

	It("cd with no target from worktree context", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "cd")
		Eventually(session).Should(gexec.Exit(0))

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		Expect(string(session.Out.Contents())).To(ContainSubstring(worktreePath))
	})

	It("cd to non-existent worktree (error)", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "cd", "nonexistent")
		Eventually(session).Should(gexec.Exit(1))
		cli.ShouldErrorOutput(session, "worktree 'nonexistent' not found")
	})

	It("shows level 1 verbose output with -v flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "cd", result.Feature1Branch, "-v")
		Eventually(session).Should(gexec.Exit(0))
		cli.ShouldVerboseOutput(session, "Navigating to worktree")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows level 2 verbose output with -vv flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "cd", result.Feature1Branch, "-vv")
		Eventually(session).Should(gexec.Exit(0))
		cli.ShouldVerboseOutput(session, "Navigating to worktree")
		cli.ShouldVerboseOutput(session, "  target: "+result.Feature1Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows no verbose output by default", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "cd", result.Feature1Branch)
		Eventually(session).Should(gexec.Exit(0))
		cli.ShouldNotHaveVerboseOutput(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})
})
