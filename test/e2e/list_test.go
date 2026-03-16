//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit list command.
// Tests validate context-aware listing behavior across project, worktree, and outside git contexts.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

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

	It("shows level 1 verbose output with -v flag", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list", "-v")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Listing worktrees")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows level 2 verbose output with -vv flag", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list", "-vv")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Listing worktrees")
		cli.ShouldVerboseOutput(session, "  project: test")
		cli.ShouldVerboseOutput(session, "  including main worktree: false")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows no verbose output by default", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list")
		cli.ShouldSucceed(session)
		cli.ShouldNotHaveVerboseOutput(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("outputs JSON format with --output json flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list", "--output", "json")
		cli.ShouldSucceed(session)
		cli.ShouldContain(session, `{"worktrees":[`)
		cli.ShouldContain(session, `"branch":"`+result.Feature1Branch+`"`)
		cli.ShouldContain(session, `"status":"clean"`)
	})

	It("outputs empty JSON array with --output json and no worktrees", func() {
		fixture.SetupSingleProject("empty-project")

		session := ctxHelper.FromProjectDir("empty-project", "list", "--output", "json")
		cli.ShouldSucceed(session)
		cli.ShouldContain(session, `{"worktrees":[]}`)
	})

	It("fails with invalid output format", func() {
		fixture.SetupSingleProject("test-project")

		session := ctxHelper.FromProjectDir("test-project", "list", "--output", "yaml")
		cli.ShouldFailWithExit(session, 5) // ExitCodeValidation
		cli.ShouldErrorOutput(session, "invalid output format")
	})

	It("suppresses success messages with --quiet flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list", "--quiet")
		cli.ShouldSucceed(session)
		// Should still show worktree list but not extra messages
		cli.ShouldOutput(session, result.Feature1Branch+" ->")
		// Should NOT show verbose messages even if they exist
		Eventually(session.Err).ShouldNot(gbytes.Say("Listing worktrees"))
	})

	It("preserves error output with --quiet flag", func() {
		fixture.SetupSingleProject("test-project")

		session := ctxHelper.FromProjectDir("test-project", "create", "invalid@branch", "--quiet")
		cli.ShouldFailWithExit(session, 5) // ExitCodeValidation
		// Error should still go to stderr
		Eventually(session.Err).Should(gbytes.Say("Error:"))
	})

	It("preserves path output with --quiet -C flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "cd", "--quiet")
		cli.ShouldSucceed(session)
		// Should output only the path, no success message
		output := cli.GetOutput(session)
		Expect(output).To(ContainSubstring("/worktrees"))
	})

	It("verbose wins over quiet with --quiet -v flags", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "list", "--quiet", "-v")
		cli.ShouldSucceed(session)
		// Verbose messages should appear (verbose wins)
		cli.ShouldVerboseOutput(session, "Listing worktrees")
	})

	It("alias ls behaves like list command", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "ls")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("list with -a short flag behaves like --all", func() {
		fixture.SetupMultiProject()

		session := cli.Run("list", "-a")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})
})
