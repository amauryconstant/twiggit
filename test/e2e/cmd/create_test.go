//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit create command.
// Tests validate worktree creation across project, worktree, and outside git contexts.
package cmde2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("create command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper
	var assertions *helpers.TwiggitAssertions

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		cli = cli.WithConfigDir(fixture.Build())
		assertions = helpers.NewTwiggitAssertions()
	})

	AfterEach(func() {
		fixture.Cleanup()
	})

	It("creates worktree from project context with default source", func() {
		fixture.SetupSingleProject("test-project")

		session := ctxHelper.FromProjectDir("test-project", "create", "feature-1")
		assertions.ShouldCreateWorktree(session, "feature-1")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		testID := fixture.GetTestID()
		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", testID.BranchName("feature-1"))
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("creates worktree with --source flag", func() {
		fixture.SetupSingleProject("test-project")
		projectPath := fixture.GetProjectPath("test-project")
		testID := fixture.GetTestID()
		gitHelper := fixture.GetGitHelper()

		err := gitHelper.CreateBranch(projectPath, "develop")
		Expect(err).NotTo(HaveOccurred())

		session := cli.Run("create", "test-project/"+testID.BranchName("feature-new"), "--source", "develop")
		assertions.ShouldCreateWorktree(session, "feature-new")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", testID.BranchName("feature-new"))
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("creates worktree from worktree context", func() {
		fixture.CreateWorktreeSetup("test")

		testID := fixture.GetTestID()
		session := ctxHelper.FromWorktreeDir("test", testID.BranchName("feature-1"), "create", "feature-2")
		assertions.ShouldCreateWorktree(session, "feature-2")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", testID.BranchName("feature-2"))
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("creates worktree from outside git with project/branch", func() {
		fixture.SetupSingleProject("external-project")

		session := ctxHelper.FromOutsideGit("create", "external-project/new-feature")
		assertions.ShouldCreateWorktree(session, "new-feature")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "external-project", "new-feature")
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("fails from outside git without project spec", func() {
		session := ctxHelper.FromOutsideGit("create", "feature-1")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "could not infer project name")
	})

	It("fails with invalid project/branch format", func() {
		fixture.SetupSingleProject("test")

		session := cli.Run("create", "invalid-format")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "invalid format: expected <project>/<branch>")
	})

	It("fails with reserved branch name", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "create", "HEAD")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "branch name format is invalid")
	})

	It("fails with invalid characters in branch name", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "create", "feature@branch#name")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "branch name format is invalid")
	})

	It("creates worktree when branch exists", func() {
		fixture.SetupSingleProject("test")
		projectPath := fixture.GetProjectPath("test")
		testID := fixture.GetTestID()
		gitHelper := fixture.GetGitHelper()

		branchName := testID.BranchName("existing-branch")
		err := gitHelper.CreateBranch(projectPath, branchName)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromProjectDir("test", "create", branchName)
		assertions.ShouldCreateWorktree(session, branchName)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", branchName)
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("fails when worktree already exists", func() {
		fixture.CreateWorktreeSetup("test")
		testID := fixture.GetTestID()
		branchName := testID.BranchName("feature-1")

		session := ctxHelper.FromProjectDir("test", "create", branchName)
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "worktree already exists")
	})

	It("fails with non-existent source branch", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "create", "new-feature", "--source", "nonexistent")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "source branch 'nonexistent' does not exist")
	})

	It("uses custom default branch from config", func() {
		fixture.WithConfig(func(ch *helpers.ConfigHelper) {
			ch.WithDefaultSourceBranch("develop")
		}).SetupSingleProject("test")

		projectPath := fixture.GetProjectPath("test")
		gitHelper := fixture.GetGitHelper()

		err := gitHelper.CreateBranch(projectPath, "develop")
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromProjectDir("test", "create", "feature")
		assertions.ShouldCreateWorktree(session, "feature")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", "feature")
		assertions.ShouldHaveWorktree(worktreePath)
	})

	PIt("outputs worktree path with --cd flag", func() {
		fixture.SetupSingleProject("test")

		session := ctxHelper.FromProjectDir("test", "create", "--cd", "feature-cd")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}

		output := cli.GetOutput(session)
		Expect(output).NotTo(BeEmpty())

		expectedPath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", "feature-cd")
		Expect(output).To(Equal(expectedPath))

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", "feature-cd")
		assertions.ShouldHaveWorktree(worktreePath)
	})
})
