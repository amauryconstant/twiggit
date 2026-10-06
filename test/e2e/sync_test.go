//go:build e2e

package e2e

import (
	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("sync command", func() {
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
			session := cli.Run("sync", "--help")
			cli.ShouldSucceed(session)
			cli.ShouldOutput(session, "tracking refs")
		})

		It("has required flags", func() {
			session := cli.Run("sync", "--help")
			cli.ShouldSucceed(session)
			for _, want := range []string{"--all", "--remote", "--branch", "--rebase", "--fetch-only"} {
				cli.ShouldOutput(session, want)
			}
		})
	})

	Describe("usage error outside git without context", func() {
		It("fails with exit 2", func() {
			session := ctxHelper.FromOutsideGit("sync")
			cli.ShouldFailWithExit(session, 2)
		})
	})

	Describe("sync from project context", func() {
		It("runs without panic", func() {
			_ = fixture.CreateWorktreeSetup("sync-test")
			session := ctxHelper.FromProjectDir("sync-test", "sync")
			Eventually(session).Should(gexec.Exit())
		})
	})

	Describe("--fetch-only", func() {
		It("skips the rebase walk", func() {
			_ = fixture.CreateWorktreeSetup("syncfetch")
			session := ctxHelper.FromProjectDir("syncfetch", "sync", "--fetch-only")
			Eventually(session).Should(gexec.Exit())
		})
	})
})
