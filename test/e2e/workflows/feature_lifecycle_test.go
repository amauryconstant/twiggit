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

	"twiggit/test/e2e/fixtures"
	e2ehelpers "twiggit/test/e2e/helpers"
	twiggithelpers "twiggit/test/helpers"
)

var _ = Describe("feature lifecycle", func() {
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

	It("completes full feature lifecycle: create, modify, commit, merge, and delete worktree", func() {
		projectName := "lifecycle-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		featureBranch := testID.BranchName("new-feature")
		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)
		featurePath := filepath.Join(projectWorktreesDir, featureBranch)
		projectPath := fixture.GetProjectPath(projectName)

		session := ctxHelper.FromProjectDir(projectName, "create", featureBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(featurePath).To(BeADirectory())

		testFile := filepath.Join(featurePath, "feature.txt")
		err := os.WriteFile(testFile, []byte("Feature implementation"), 0644)
		Expect(err).NotTo(HaveOccurred())

		repo, err := gitHelper.PlainOpen(featurePath)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Add("feature.txt")
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Commit("Add feature.txt", nil)
		Expect(err).NotTo(HaveOccurred())

		session = ctxHelper.FromProjectDir(projectName, "merge", featureBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		branches, err := gitHelper.ListBranches(projectPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(branches).To(ContainElement(featureBranch))

		session = ctxHelper.FromProjectDir(projectName, "delete", featureBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		Expect(featurePath).NotTo(BeADirectory())

		branches, err = gitHelper.ListBranches(projectPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(branches).NotTo(ContainElement(featureBranch))
	})
})
