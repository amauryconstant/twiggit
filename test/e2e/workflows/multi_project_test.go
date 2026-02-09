//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package workflowse2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"twiggit/test/e2e/fixtures"
	e2ehelpers "twiggit/test/e2e/helpers"
)

var _ = Describe("multi project workflow", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *e2ehelpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = e2ehelpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("switches between multiple projects and manages worktrees independently", func() {
		testID := fixture.GetTestID()
		project1Name := testID.ProjectNameWithSuffix("1")
		project2Name := testID.ProjectNameWithSuffix("2")

		fixture.SetupMultiProject()

		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		project1WorktreesDir := filepath.Join(worktreesDir, project1Name)
		project2WorktreesDir := filepath.Join(worktreesDir, project2Name)

		feature1Branch := testID.BranchName("feature-1")
		feature2Branch := testID.BranchName("feature-2")
		feature1Path := filepath.Join(project1WorktreesDir, feature1Branch)
		feature2Path := filepath.Join(project2WorktreesDir, feature2Branch)

		session := ctxHelper.FromProjectDir(project1Name, "create", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature1Path).To(BeADirectory())

		session = ctxHelper.FromProjectDir(project2Name, "create", feature2Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature2Path).To(BeADirectory())

		session = ctxHelper.FromProjectDir(project1Name, "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, feature1Branch)
		Eventually(session.Out).ShouldNot(gbytes.Say(feature2Branch))

		session = ctxHelper.FromProjectDir(project2Name, "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, feature2Branch)
		Eventually(session.Out).ShouldNot(gbytes.Say(feature1Branch))

		session = ctxHelper.FromWorktreeDir(project1Name, feature1Branch, "cd", project2Name+"/"+feature2Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, feature2Path)

		session = ctxHelper.FromProjectDir(project1Name, "delete", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature1Path).NotTo(BeADirectory())

		Expect(feature2Path).To(BeADirectory())
		session = ctxHelper.FromProjectDir(project2Name, "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, feature2Branch)
	})
})
