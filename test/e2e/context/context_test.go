//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("context-aware behavior", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper
	var assertions *helpers.TwiggitAssertions

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		configDir := fixture.Build()
		cli = cli.WithConfigDir(configDir)
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
		assertions = helpers.NewTwiggitAssertions()
	})

	AfterEach(func() {
		if fixture != nil {
			fixture.Cleanup()
		}
	})

	It("creates worktree from project context", func() {
		fixture.SetupSingleProject("test-project")

		session := ctxHelper.FromProjectDir("test-project", "create", "feature-1")
		assertions.ShouldCreateWorktree(session, "feature-1")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", "feature-1")
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

	It("lists worktrees from project context", func() {
		fixture.CreateWorktreeSetup("test")

		testID := fixture.GetTestID()
		session := ctxHelper.FromProjectDir("test", "list")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldOutput(session, testID.BranchName("feature-1"))
		cli.ShouldOutput(session, testID.BranchName("feature-2"))
	})

	It("deletes worktree from worktree context", func() {
		fixture.CreateWorktreeSetup("test")

		testID := fixture.GetTestID()
		branchToDelete := testID.BranchName("feature-1")

		session := ctxHelper.FromWorktreeDir("test", branchToDelete, "delete", branchToDelete)
		assertions.ShouldDeleteWorktree(session, branchToDelete)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", branchToDelete)
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("changes directory from outside git", func() {
		fixture.CreateWorktreeSetup("test")

		testID := fixture.GetTestID()
		targetBranch := testID.BranchName("feature-1")

		session := ctxHelper.FromOutsideGit("cd", "test", targetBranch)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		expectedPath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", targetBranch)
		cli.ShouldOutput(session, expectedPath)
	})

	It("creates worktree for one project from another project", func() {
		fixture.SetupMultiProject()

		testID := fixture.GetTestID()
		project1Name := testID.ProjectNameWithSuffix("1")
		project2Name := testID.ProjectNameWithSuffix("2")

		session := ctxHelper.FromProjectDir(project1Name, "create", project2Name+"/cross-feature")
		assertions.ShouldCreateWorktree(session, "cross-feature")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), project2Name, "cross-feature")
		assertions.ShouldHaveWorktree(worktreePath)
	})

	It("fails to create worktree from outside git without project specification", func() {
		session := ctxHelper.FromOutsideGit("create", "feature-1")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "cannot infer project")
	})
})
