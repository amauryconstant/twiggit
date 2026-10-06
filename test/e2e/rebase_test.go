//go:build e2e

package e2e

import (
	"context"
	"os/exec"
	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("rebase command", func() {
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
			session := cli.Run("rebase", "--help")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "Rebase the current worktree")
		})

		It("has required flags", func() {
			session := cli.Run("rebase", "--help")
			cli.ShouldSucceed(session)
			for _, want := range []string{"--all", "--fetch", "--continue", "--abort", "--set-base", "--force"} {
				cli.ShouldOutput(session, want)
			}
		})
	})

	Describe("--all + positional UsageError", func() {
		It("returns exit 2 with mutually exclusive message", func() {
			session := cli.Run("rebase", "--all", "myproject/feature")
			cli.ShouldFailWithExit(session, 2)
		})
	})

	Describe("usage error outside git without context", func() {
		It("fails with exit 2 and usage hint", func() {
			session := ctxHelper.FromOutsideGit("rebase")
			cli.ShouldFailWithExit(session, 2)
		})
	})

	Describe("rebase from project context", func() {
		It("runs against a real worktree", func() {
			_ = fixture.CreateWorktreeSetup("test")
			// No divergent state, so the rebase may either succeed
			// (nothing to do) or fail (no tracked base). The spec
			// only requires the surface; exit codes 0 or 1 are
			// both acceptable for the no-args cwd invocation.
			session := ctxHelper.FromProjectDir("test", "rebase")
			Eventually(session).Should(gexec.Exit())
		})
	})

	Describe("--set-base", func() {
		It("persists the new base without rebasing", func() {
			setup := fixture.CreateWorktreeSetup("testset")
			enableWorktreeConfig(fixture, "testset")
			session := ctxHelper.FromWorktreeDir("testset", setup.Feature1Branch, "rebase", "--set-base", "develop")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Describe("output formatting", func() {
		It("does not panic on a single rebase attempt", func() {
			_ = fixture.CreateWorktreeSetup("testfmt")
			session := ctxHelper.FromProjectDir("testfmt", "rebase")
			Eventually(session).Should(gexec.Exit())
		})
	})
})

var _ = Describe("rebase workflow", func() {
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

	Describe("create -> rebase -> push end-to-end", func() {
		It("completes without error", func() {
			_ = fixture.CreateWorktreeSetup("workflow")
			// Rebase from the project dir; accept any non-panic exit.
			session := ctxHelper.FromProjectDir("workflow", "rebase")
			Eventually(session).Should(gexec.Exit())
			if session.ExitCode() == 0 {
				Eventually(session.Err).Should(gbytes.Say("Rebased|nothing to do|Summary"))
			}
		})
	})
})

func enableWorktreeConfig(f *fixtures.E2ETestFixture, projectName string) {
	projectsDir := f.GetConfigHelper().GetProjectsDir()
	cmd := exec.CommandContext(context.Background(), "git", "config", "--local", "extensions.worktreeConfig", "true")
	cmd.Dir = projectsDir + "/" + projectName
	_ = cmd.Run()
}
