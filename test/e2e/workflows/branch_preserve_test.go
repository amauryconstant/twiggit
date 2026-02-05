//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package e2e

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

var _ = Describe("branch preserve workflow", func() {
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

	It("preserves branch when deleting worktree with --keep-branch flag", func() {
		projectName := "preserve-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		featureBranch := testID.BranchName("preserve-feature")
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

		testFile := filepath.Join(featurePath, "preserve.txt")
		err := os.WriteFile(testFile, []byte("Preserve this content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		repo, err := gitHelper.PlainOpen(featurePath)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Add("preserve.txt")
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Commit("Add preserve.txt", nil)
		Expect(err).NotTo(HaveOccurred())

		branches, err := gitHelper.ListBranches(projectPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(branches).To(ContainElement(featureBranch))

		session = ctxHelper.FromProjectDir(projectName, "delete", "--keep-branch", featureBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		Expect(featurePath).NotTo(BeADirectory())

		branches, err = gitHelper.ListBranches(projectPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(branches).To(ContainElement(featureBranch))

		newFeaturePath := filepath.Join(projectWorktreesDir, featureBranch)
		session = ctxHelper.FromProjectDir(projectName, "create", featureBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		Expect(newFeaturePath).To(BeADirectory())

		preservedFile := filepath.Join(newFeaturePath, "preserve.txt")
		content, err := os.ReadFile(preservedFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("Preserve this content"))
	})
})
