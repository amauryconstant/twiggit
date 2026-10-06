//go:build e2e

package e2e

import (
	"path/filepath"
	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// rebase workflow e2e specs live in their own file per
// golang-testing file-naming convention (one source concern per test
// file). The flow exercises the user-visible contract:
//
//	twiggit create  →  twiggit rebase  →  twiggit cd
//
// The cd step proves the shell-wrapper navigation contract from
// cli-rebase/spec.md "Single-target rebase prints navigation path
// for the shell wrapper" + cli-cd: the printed path must equal the
// worktree directory the fixture just created.
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

	Describe("create -> rebase -> shell-wrapper cd", func() {
		It("prints the worktree path for the shell wrapper after a clean rebase", func() {
			setup := fixture.CreateWorktreeSetup("workflow")

			rebase := ctxHelper.FromProjectDir("workflow", "rebase")
			Eventually(rebase).Should(gexec.Exit())

			cd := ctxHelper.FromOutsideGit("cd", "workflow/"+setup.Feature1Branch)
			Eventually(cd).Should(gexec.Exit(0))

			wantPath := filepath.Join(
				fixture.GetConfigHelper().GetWorktreesDir(),
				"workflow",
				setup.Feature1Branch,
			)
			Expect(string(cd.Out.Contents())).To(ContainSubstring(wantPath),
				"twiggit cd must print the worktree path so the shell wrapper can navigate to it")
		})
	})
})
