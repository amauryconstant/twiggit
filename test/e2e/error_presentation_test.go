//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("Error Presentation", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		if fixture != nil {
			GinkgoT().Log(fixture.Inspect())
			fixture.Cleanup()
		}
	})

	Context("Validation errors", func() {
		PIt("includes helpful suggestions for invalid input", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-validation")

			session := ctxHelper.FromProjectDir("test-validation", "create", "invalid@branch#name")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty(), "Error output should be in stderr")
			Expect(stderr).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("format"),
			), "Error message should indicate invalid input")
			Expect(stderr).To(Or(
				ContainSubstring("💡"),
				ContainSubstring("Use"),
			), "Error should include helpful suggestions")
		})

		PIt("provides specific validation error messages", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-validation-specific")

			session := ctxHelper.FromProjectDir("test-validation-specific", "create")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("arg"),
				ContainSubstring("accepts"),
			), "Error should specify what's missing")
		})
	})

	Context("Context errors", func() {
		PIt("explains ambiguity clearly when running outside git", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)

			session := ctxHelper.FromOutsideGit("list")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("project"),
				ContainSubstring("context"),
			), "Error should explain missing context")
			Expect(stderr).To(Or(
				ContainSubstring("--all"),
			), "Error should suggest how to resolve ambiguity")
		})

		PIt("provides actionable messages for ambiguous project references", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)

			session := ctxHelper.FromOutsideGit("create", "nonexistent/feature")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("not found"),
				ContainSubstring("could not"),
			), "Error should explain the ambiguity clearly")
		})
	})

	Context("Filesystem errors", func() {
		PIt("provides actionable messages for permission errors", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-filesystem")

			session := ctxHelper.FromProjectDir("test-filesystem", "create", "/root/forbidden-worktree")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("permission"),
				ContainSubstring("cannot"),
				ContainSubstring("access"),
				ContainSubstring("denied"),
				ContainSubstring("branch name"),
			), "Error should explain the issue")
		})

		PIt("provides clear messages for worktree path conflicts", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-path-conflict")

			session := ctxHelper.FromProjectDir("test-path-conflict", "create", "test-path-conflict/nonexistent")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("path"),
				ContainSubstring("directory"),
				ContainSubstring("already"),
				ContainSubstring("not found"),
				ContainSubstring("does not exist"),
			), "Error should explain the path issue")
		})
	})

	Context("Git errors", func() {
		PIt("presents user-friendly git errors", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-git-error")

			session := ctxHelper.FromProjectDir("test-git-error", "cd", "nonexistent-branch")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).NotTo(ContainSubstring("fatal:"),
				"Error should not expose raw git error messages")
			Expect(stderr).To(Or(
				ContainSubstring("not found"),
				ContainSubstring("could not find"),
			), "Error should be user-friendly")
		})

		PIt("provides actionable suggestions for common git errors", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-git-suggestion")

			session := ctxHelper.FromProjectDir("test-git-suggestion", "create", "nonexistent-branch")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("does not exist"),
				ContainSubstring("not found"),
				ContainSubstring("source"),
			), "Error should explain the git error clearly")
		})
	})

	Context("Flag conflicts", func() {
		PIt("explains conflicting flags clearly", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-flag-conflict")

			session := ctxHelper.FromProjectDir("test-flag-conflict", "list", "--all", "test-flag-conflict")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("conflict"),
				ContainSubstring("cannot"),
				ContainSubstring("together"),
				ContainSubstring("argument"),
				ContainSubstring("unknown"),
			), "Error should explain the flag issue")
		})

		PIt("suggests valid flag combinations", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-flag-suggestion")

			session := ctxHelper.FromProjectDir("test-flag-suggestion", "create", "--source=main", "feature")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("does not exist"),
				ContainSubstring("not found"),
				ContainSubstring("source"),
			), "Error should explain the issue")
		})
	})
})
