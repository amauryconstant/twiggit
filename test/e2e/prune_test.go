//go:build e2e
// +build e2e

package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("prune command", func() {
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
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	Describe("help and basic usage", func() {
		It("shows help without error", func() {
			session := cli.Run("prune", "--help")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Prune merged worktrees")
		})

		It("has required flags", func() {
			session := cli.Run("prune", "--help")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "--dry-run")
			cli.ShouldOutput(session, "--force")
			cli.ShouldOutput(session, "--delete-branches")
			cli.ShouldOutput(session, "--all")
		})
	})

	Describe("dry-run mode", func() {
		It("shows preview without error in project context", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "prune", "--dry-run")
			cli.ShouldSucceed(session)
		})

		It("shows preview across all projects with --all flag", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromOutsideGit("prune", "--dry-run", "--all")
			cli.ShouldSucceed(session)
		})
	})

	Describe("error handling", func() {
		It("fails with invalid worktree format", func() {
			session := ctxHelper.FromOutsideGit("prune", "invalid-format")
			cli.ShouldFailWithExit(session, 1)
		})

		It("fails when using --all with specific worktree", func() {
			session := ctxHelper.FromOutsideGit("prune", "--all", "test/feature-1")
			cli.ShouldFailWithExit(session, 1)
		})
	})

	Describe("context-aware behavior", func() {
		It("works from project context", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "prune", "--dry-run")
			cli.ShouldSucceed(session)
		})

		It("works from outside git context with --all", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromOutsideGit("prune", "--all", "--dry-run")
			cli.ShouldSucceed(session)
		})
	})

	Describe("force flag", func() {
		It("accepts --force flag", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "prune", "--dry-run", "--force")
			cli.ShouldSucceed(session)
		})
	})

	Describe("delete-branches flag", func() {
		It("accepts --delete-branches flag", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "prune", "--dry-run", "--delete-branches")
			cli.ShouldSucceed(session)
		})
	})

	Describe("protected branches", func() {
		It("protects main branch by default", func() {
			_ = fixture.CreateWorktreeSetup("test")

			session := ctxHelper.FromProjectDir("test", "prune", "--dry-run", "--force")
			cli.ShouldSucceed(session)
		})
	})

	Describe("bulk prune confirmation", func() {
		It("prompts for confirmation with --all --delete-branches and accepts y", func() {
			_ = fixture.CreateMergedWorktreeSetup("test1")

			stdin := strings.NewReader("y\n")
			session := ctxHelper.FromOutsideGitWithStdin(stdin, "prune", "--all", "--delete-branches")

			Eventually(session.Err).Should(gbytes.Say("This will prune merged worktrees across all projects"))
			Eventually(session.Err).Should(gbytes.Say("Continue"))
			Eventually(session).Should(gexec.Exit(0))
		})

		It("cancels prune when user declines confirmation with n", func() {
			_ = fixture.CreateMergedWorktreeSetup("test1")

			stdin := strings.NewReader("n\n")
			session := ctxHelper.FromOutsideGitWithStdin(stdin, "prune", "--all", "--delete-branches")

			Eventually(session.Err).Should(gbytes.Say("Continue"))
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Err).Should(gbytes.Say("Prune cancelled"))
		})

		It("cancels prune on empty response", func() {
			_ = fixture.CreateMergedWorktreeSetup("test1")

			stdin := strings.NewReader("\n")
			session := ctxHelper.FromOutsideGitWithStdin(stdin, "prune", "--all", "--delete-branches")

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Err).Should(gbytes.Say("Prune cancelled"))
		})
	})

	Describe("bulk prune with --force", func() {
		It("bypasses confirmation prompt with --force flag", func() {
			_ = fixture.CreateMergedWorktreeSetup("test1")

			session := ctxHelper.FromOutsideGit("prune", "--all", "--delete-branches", "--force")

			Consistently(session.Err).ShouldNot(gbytes.Say("Continue"))
			Consistently(session.Err).ShouldNot(gbytes.Say("This will prune merged worktrees across all projects"))
			Eventually(session).Should(gexec.Exit(0))
		})
	})
})
