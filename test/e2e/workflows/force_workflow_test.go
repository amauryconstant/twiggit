//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package workflowse2e

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"twiggit/test/e2e/fixtures"
	e2ehelpers "twiggit/test/e2e/helpers"
	twiggithelpers "twiggit/test/helpers"
)

var _ = Describe("force workflow", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *e2ehelpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper
	var gitHelper *twiggithelpers.GitTestHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = e2ehelpers.NewTwiggitCLI()
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		gitHelper = twiggithelpers.NewGitTestHelper(&testing.T{})
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("creates modified worktree and forces deletion with --force flag", func() {
		projectName := "force-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		forceBranch := testID.BranchName("force-feature")
		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)
		forcePath := filepath.Join(projectWorktreesDir, forceBranch)

		session := ctxHelper.FromProjectDir(projectName, "create", forceBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(forcePath).To(BeADirectory())

		testFile := filepath.Join(forcePath, "uncommitted.txt")
		err := os.WriteFile(testFile, []byte("Uncommitted changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		repo, err := gitHelper.PlainOpen(forcePath)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		status, err := wt.Status()
		Expect(err).NotTo(HaveOccurred())
		Expect(status.IsClean()).To(BeFalse())

		session = ctxHelper.FromProjectDir(projectName, "delete", forceBranch)
		cli.ShouldFailWithExit(session, 1)
		cli.ShouldErrorOutput(session, "uncommitted changes")

		Expect(forcePath).To(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "delete", "--force", forceBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		Expect(forcePath).NotTo(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "list")
		cli.ShouldSucceed(session)
		Eventually(session.Out).ShouldNot(gbytes.Say(forceBranch))
	})
})
