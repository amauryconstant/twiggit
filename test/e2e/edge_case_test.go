//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for edge case repository handling.
// Tests validate graceful error handling for unusual repository states.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("edge case handling", func() {
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
		fixture.Cleanup()
	})

	Context("corrupted repository", func() {
		It("handles list command gracefully", func() {
			fixture.SetupCorruptedProject("corrupted-project")

			session := ctxHelper.FromProjectDir("corrupted-project", "list")
			// Should handle gracefully - either succeed with empty output or fail
			Eventually(session).Should(Or(
				gexec.Exit(0),
				gexec.Exit(1),
			))
		})

		It("handles create command gracefully", func() {
			fixture.SetupCorruptedProject("corrupted-project")

			session := ctxHelper.FromProjectDir("corrupted-project", "create", "test-branch")
			// Should handle gracefully - either succeed or fail
			Eventually(session).Should(Or(
				gexec.Exit(0),
				gexec.Exit(1),
			))
		})
	})

	Context("bare repository", func() {
		It("handles list command gracefully", func() {
			fixture.SetupBareProject("bare-project")

			session := ctxHelper.FromProjectDir("bare-project", "list")
			// Bare repos don't have worktrees in the traditional sense
			// The command may succeed with empty output or fail gracefully
			Eventually(session).Should(Or(
				gexec.Exit(0), // Success with empty list
				gexec.Exit(1), // General error
				gexec.Exit(5), // Validation error (appropriate for bare repo)
			))
		})

		It("handles create command gracefully", func() {
			fixture.SetupBareProject("bare-project")

			session := ctxHelper.FromProjectDir("bare-project", "create", "test-branch")
			// Bare repos may not support worktree creation
			// Should handle gracefully - either succeed or fail
			Eventually(session).Should(Or(
				gexec.Exit(0),
				gexec.Exit(1),
			))
		})
	})

	Context("submodule repository", func() {
		It("handles list command in repo with submodules", func() {
			fixture.SetupSubmoduleProject("submodule-project")

			session := ctxHelper.FromProjectDir("submodule-project", "list")
			// Should succeed - submodules shouldn't break basic operations
			cli.ShouldSucceed(session)
		})

		It("detects submodule presence", func() {
			fixture.SetupSubmoduleProject("submodule-project")

			session := ctxHelper.FromProjectDir("submodule-project", "list")
			cli.ShouldSucceed(session)
			// Output should not contain error due to submodule
			Eventually(session.Out).ShouldNot(gbytes.Say("error"))
		})
	})

	Context("detached HEAD repository", func() {
		It("handles list command in detached HEAD state", func() {
			fixture.SetupDetachedProject("detached-project")

			session := ctxHelper.FromProjectDir("detached-project", "list")
			// Should succeed even in detached HEAD state
			cli.ShouldSucceed(session)
		})

		It("handles create command from detached HEAD", func() {
			fixture.SetupDetachedProject("detached-project")

			session := ctxHelper.FromProjectDir("detached-project", "create", "test-feature")
			// Creating from detached HEAD may behave differently
			// Either succeeds or fails gracefully
			Eventually(session).Should(Or(
				gexec.Exit(0),
				gexec.Exit(1),
			))
		})
	})
})
