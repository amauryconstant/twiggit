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

var _ = Describe("cleanup workflow", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *e2ehelpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = e2ehelpers.NewTwiggitCLI()
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	PIt("performs bulk delete operations with different flags", func() {
		projectName := "cleanup-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)

		feature1Branch := testID.BranchName("feature-1")
		feature2Branch := testID.BranchName("feature-2")
		feature3Branch := testID.BranchName("feature-3")
		feature1Path := filepath.Join(projectWorktreesDir, feature1Branch)
		feature2Path := filepath.Join(projectWorktreesDir, feature2Branch)
		feature3Path := filepath.Join(projectWorktreesDir, feature3Branch)

		session := ctxHelper.FromProjectDir(projectName, "create", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature1Path).To(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "create", feature2Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature2Path).To(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "create", feature3Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature3Path).To(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "delete", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature2Path).NotTo(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "delete", "--keep-branch", feature2Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature3Path).NotTo(BeADirectory())

		session = ctxHelper.FromWorktreeDir(projectName, feature3Branch, "delete", "-C", feature3Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(feature3Path).NotTo(BeADirectory())

		Expect(feature1Path).NotTo(BeADirectory())
		Expect(feature2Path).NotTo(BeADirectory())
		Expect(feature3Path).NotTo(BeADirectory())

		session = ctxHelper.FromProjectDir(projectName, "list")
		cli.ShouldSucceed(session)
		Eventually(session.Out).ShouldNot(gbytes.Say(feature1Branch))
		Eventually(session.Out).ShouldNot(gbytes.Say(feature2Branch))
		Eventually(session.Out).ShouldNot(gbytes.Say(feature3Branch))
	})
})
