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
	"twiggit/test/helpers"

	git "github.com/go-git/go-git/v5"
)

var _ = Describe("rebase workflow", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *e2ehelpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper
	var gitHelper *helpers.GitTestHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = e2ehelpers.NewTwiggitCLI()
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		gitHelper = helpers.NewGitTestHelper(&testing.T{})
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("completes rebase workflow: create worktree, commit changes, switch worktree, delete original", func() {
		projectName := "rebase-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		feature1Branch := testID.BranchName("feature-1")
		feature2Branch := testID.BranchName("feature-2")
		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)

		session := ctxHelper.FromProjectDir(projectName, "create", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		feature1Path := filepath.Join(projectWorktreesDir, feature1Branch)
		Expect(feature1Path).To(BeADirectory())

		GinkgoT().Log("Step 2: Create and commit changes in feature-1")
		testFile := filepath.Join(feature1Path, "feature1.txt")
		err := os.WriteFile(testFile, []byte("Feature 1 changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		repo, err := gitHelper.PlainOpen(feature1Path)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Add("feature1.txt")
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Commit("Add feature1.txt", &git.CommitOptions{})
		Expect(err).NotTo(HaveOccurred())

		session = ctxHelper.FromProjectDir(projectName, "create", feature2Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		feature2Path := filepath.Join(projectWorktreesDir, feature2Branch)
		Expect(feature2Path).To(BeADirectory())

		Expect(feature1Path).To(BeADirectory())

		session = ctxHelper.FromWorktreeDir(projectName, feature2Branch, "delete", feature1Branch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		Expect(feature1Path).NotTo(BeADirectory())

		Expect(feature2Path).To(BeADirectory())
	})
})
