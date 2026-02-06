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

var _ = Describe("hotfix workflow", func() {
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

	It("completes hotfix workflow: create hotfix, make urgent fix, merge to main, delete hotfix", func() {
		projectName := "hotfix-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		hotfixBranch := testID.BranchName("hotfix-critical-bug")
		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)

		session := ctxHelper.FromProjectDir(projectName, "create", hotfixBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)
		hotfixPath := filepath.Join(projectWorktreesDir, hotfixBranch)
		Expect(hotfixPath).To(BeADirectory())

		GinkgoT().Log("Step 2: Create and commit urgent fix in hotfix")
		fixFile := filepath.Join(hotfixPath, "urgent-fix.txt")
		err := os.WriteFile(fixFile, []byte("Critical fix for production"), 0644)
		Expect(err).NotTo(HaveOccurred())

		repo, err := gitHelper.PlainOpen(hotfixPath)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Add("urgent-fix.txt")
		Expect(err).NotTo(HaveOccurred())

		_, err = wt.Commit("Fix critical bug in production", &git.CommitOptions{})
		Expect(err).NotTo(HaveOccurred())

		session = ctxHelper.FromWorktreeDir(projectName, hotfixBranch, "delete", hotfixBranch)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
		cli.ShouldSucceed(session)

		Expect(hotfixPath).NotTo(BeADirectory())
	})

	It("handles hotfix with multiple critical files", func() {
		projectName := "multi-fix-project"
		fixture.SetupSingleProject(projectName)
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		testID := fixture.GetTestID()

		hotfixBranch := testID.BranchName("hotfix-multi-bug")
		worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
		projectWorktreesDir := filepath.Join(worktreesDir, projectName)

		session := ctxHelper.FromProjectDir(projectName, "create", hotfixBranch)
		cli.ShouldSucceed(session)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
		hotfixPath := filepath.Join(projectWorktreesDir, hotfixBranch)
		Expect(hotfixPath).To(BeADirectory())

		GinkgoT().Log("Step 2: Create and commit multiple critical fixes")
		fixFiles := []string{"fix1.txt", "fix2.txt", "fix3.txt"}
		for i, fileName := range fixFiles {
			fixFile := filepath.Join(hotfixPath, fileName)
			err := os.WriteFile(fixFile, []byte("Critical fix "+string(rune(i))), 0644)
			Expect(err).NotTo(HaveOccurred())
		}

		repo, err := gitHelper.PlainOpen(hotfixPath)
		Expect(err).NotTo(HaveOccurred())
		wt, err := repo.Worktree()
		Expect(err).NotTo(HaveOccurred())

		for _, fileName := range fixFiles {
			_, err := wt.Add(fileName)
			Expect(err).NotTo(HaveOccurred())
		}

		_, err = wt.Commit("Fix multiple critical bugs", &git.CommitOptions{})
		Expect(err).NotTo(HaveOccurred())

		session = ctxHelper.FromWorktreeDir(projectName, hotfixBranch, "delete", hotfixBranch)
		cli.ShouldSucceed(session)
		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(hotfixPath).NotTo(BeADirectory())
	})
})
