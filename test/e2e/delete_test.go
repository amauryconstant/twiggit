//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit delete command.
// Tests validate worktree deletion with force, keep-branch, and merged-only flags.
package e2e

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("delete command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("deletes worktree from project context", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch)
		cli.ShouldSucceed(session)

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("deletes other worktree from worktree context", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktree2Path := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature2Branch)

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature2Branch)
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, result.Feature2Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktree2Path).NotTo(BeADirectory())
	})

	It("deletes with --force flag despite uncommitted changes", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		testFile := filepath.Join(worktreePath, "test.txt")
		err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "--force")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, result.Feature1Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("changes to main project with -C flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		projectPath := fixture.GetProjectPath("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "-C")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := cli.GetOutput(session)

		Expect(output).To(Equal(projectPath), "Should output only the project path for navigation")
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("with -C flag from worktree context outputs project path", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)
		projectPath := fixture.GetProjectPath("test")

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "-C")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := cli.GetOutput(session)

		Expect(output).To(Equal(projectPath), "Should output only the project path, not 'Deleted worktree' message")
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("with -C flag from project context outputs no navigation path", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch, "-C")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := cli.GetOutput(session)

		Expect(output).To(BeEmpty(), "Should output nothing when deleting from project context with -C")
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("with -C flag from outside git context outputs no navigation path", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromOutsideGit("delete", "test/"+result.Feature1Branch, "-C")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := cli.GetOutput(session)

		Expect(output).To(BeEmpty(), "Should output nothing when deleting from outside git context with -C")
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("with -f short form flag works correctly", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		testFile := filepath.Join(worktreePath, "test.txt")
		err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
		Expect(err).NotTo(HaveOccurred())

		session := ctxHelper.FromWorktreeDir("test", result.Feature1Branch, "delete", result.Feature1Branch, "-f")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, result.Feature1Branch)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("shows level 1 verbose output with -v flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch, "-v")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Deleting worktree at "+worktreePath)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("shows level 2 verbose output with -vv flag", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch, "-vv")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Deleting worktree at "+worktreePath)
		cli.ShouldVerboseOutput(session, "  project: test")
		cli.ShouldVerboseOutput(session, "  branch: "+result.Feature1Branch)
		cli.ShouldVerboseOutput(session, "  force: false")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("shows no verbose output by default", func() {
		result := fixture.CreateWorktreeSetup("test")

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test", result.Feature1Branch)

		session := ctxHelper.FromProjectDir("test", "delete", result.Feature1Branch)
		cli.ShouldSucceed(session)
		cli.ShouldNotHaveVerboseOutput(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		Expect(worktreePath).NotTo(BeADirectory())
	})
})
