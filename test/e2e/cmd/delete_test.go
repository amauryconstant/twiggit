//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit delete command.
// Tests validate worktree deletion with force, keep-branch, and merged-only flags.
package cmde2e

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("delete command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper
	var assertions *helpers.TwiggitAssertions

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		assertions = helpers.NewTwiggitAssertions()
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("deletes worktree from project context", func() {
		result := fixture.CreateWorktreeSetup("test")

		// Update both CLI and ContextHelper after CreateWorktreeSetup rebuilt config
		configDir := fixture.GetConfigHelper().GetConfigDir()
		cli = cli.WithConfigDir(configDir)
		ctxHelper = fixtures.NewContextHelper(fixture, cli)

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("deletes other worktree from worktree context", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktree2Path := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature2Branch)

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature2Branch)
		assertions.ShouldDeleteWorktree(session, result.Feature2Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktree2Path).NotTo(BeADirectory())
	})

	PIt("fails with uncommitted changes", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		testFile := filepath.Join(worktreePath, "test.txt")
		err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch)
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "uncommitted changes")
	})

	It("deletes with --force flag despite uncommitted changes", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		testFile := filepath.Join(worktreePath, "test.txt")
		err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "--force")
		assertions.ShouldDeleteWorktree(session, result.Feature1Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	PIt("keeps branch with --keep-branch flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		projectPath := fixture.GetProjectPath("test")
		gitHelper := fixture.GetGitHelper()

		session := cli.Run("delete", "test/"+result.Feature1Branch, "--keep-branch")
		assertions.ShouldDeleteWorktree(session, result.Feature1Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())

		branches, err := gitHelper.ListBranches(projectPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(branches).To(ContainElement(result.Feature1Branch))
	})

	PIt("succeeds with --merged-only when branch is merged", func() {
		fixture.SetupSingleProject("test")
		projectPath := fixture.GetProjectPath("test")
		testID := fixture.GetTestID()
		gitHelper := fixture.GetGitHelper()

		err := gitHelper.CreateBranch(projectPath, testID.BranchName("merged-feature"))
		Expect(err).NotTo(HaveOccurred())

		worktreesDir := filepath.Join(fixture.GetTempDir(), "worktrees")
		err = os.MkdirAll(worktreesDir, 0755)
		Expect(err).NotTo(HaveOccurred())
		fixture.GetConfigHelper().WithWorktreesDir(worktreesDir)

		worktreePath := filepath.Join(worktreesDir, testID.BranchName("merged-feature"))
		err = fixture.CreateWorktree(projectPath, worktreePath, testID.BranchName("merged-feature"))
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromProjectDir("test", "delete", testID.BranchName("merged-feature"), "--merged-only")
		assertions.ShouldDeleteWorktree(session, testID.BranchName("merged-feature"))

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	PIt("fails with --merged-only when branch is not merged", func() {
		fixture.SetupSingleProject("test")
		projectPath := fixture.GetProjectPath("test")
		testID := fixture.GetTestID()
		gitHelper := fixture.GetGitHelper()

		err := gitHelper.CreateBranch(projectPath, testID.BranchName("unmerged"))
		Expect(err).NotTo(HaveOccurred())

		worktreesDir := filepath.Join(fixture.GetTempDir(), "worktrees")
		err = os.MkdirAll(worktreesDir, 0755)
		Expect(err).NotTo(HaveOccurred())
		fixture.GetConfigHelper().WithWorktreesDir(worktreesDir)

		worktreePath := filepath.Join(worktreesDir, testID.BranchName("unmerged"))
		err = fixture.CreateWorktree(projectPath, worktreePath, testID.BranchName("unmerged"))
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromProjectDir("test", "delete", testID.BranchName("unmerged"), "--merged-only")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "not merged")
	})

	It("changes to main project with -C flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "-C")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := cli.GetOutput(session)
		lines := strings.Split(output, "\n")
		Expect(len(lines)).To(BeNumerically(">=", 2))

		Expect(worktreePath).NotTo(BeADirectory())
	})

	PIt("fails to delete non-existent worktree", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "delete", "nonexistent")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "worktree not found")
	})

	PIt("gracefully handles already removed worktree", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		err := fixture.RemoveWorktree(worktreePath)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch)
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "worktree not found")
		Expect(worktreePath).NotTo(BeADirectory())
	})
})
