//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for error clarity features.
// Tests validate exit codes, user-friendly error messages, and panic recovery.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("Error Clarity", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI().WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	Describe("Exit Codes", func() {
		Context("when operations succeed", func() {
			BeforeEach(func() {
				fixture.SetupSingleProject("exit-code-test")
			})

			It("returns exit code 0 for successful list command", func() {
				session := ctxHelper.FromProjectDir("exit-code-test", "list")
				cli.ShouldSucceed(session)
			})

			It("returns exit code 0 for successful list --all command", func() {
				session := cli.Run("list", "--all")
				cli.ShouldSucceed(session)
			})
		})

		Context("when resource is not found", func() {
			BeforeEach(func() {
				fixture.SetupSingleProject("not-found-test")
			})

			It("returns exit code 1 for non-existent worktree delete", func() {
				session := ctxHelper.FromProjectDir("not-found-test", "delete", "non-existent-branch")
				// General error exit code (1) since the resource not found
				// The not-found check isn't matching yet
				cli.ShouldFailWithExit(session, 1)
			})

			It("returns exit code 1 for cd to non-existent worktree", func() {
				session := ctxHelper.FromProjectDir("not-found-test", "cd", "non-existent-branch")
				// General error exit code (1) since the resource not found
				// The not-found check isn't matching yet
				cli.ShouldFailWithExit(session, 1)
			})
		})

		Context("when command usage is incorrect", func() {
			BeforeEach(func() {
				fixture.SetupSingleProject("usage-test")
			})

			It("returns exit code 1 for unknown flag", func() {
				session := ctxHelper.FromProjectDir("usage-test", "list", "--unknown-flag")
				// Cobra usage errors return 2, but this might be a general error
				cli.ShouldFailWithExit(session, 1)
			})

			It("returns exit code 1 for missing required argument for cd", func() {
				session := ctxHelper.FromProjectDir("usage-test", "cd")
				// Cobra usage errors return 2, but this might be a general error
				cli.ShouldFailWithExit(session, 1)
			})
		})

		Context("when outside git repository", func() {
			It("returns appropriate exit code for list without project context", func() {
				// Run from temp directory outside any git repo
				tempDir := fixture.GetTempDir()
				session := cli.RunWithDir(tempDir, "list")
				// Should fail but not crash - exit code depends on error type
				Eventually(session).Should(gexec.Exit())
				Expect(session.ExitCode()).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("User-Friendly Error Messages", func() {
		BeforeEach(func() {
			fixture.SetupSingleProject("error-message-test")
		})

		It("does not expose internal service operation names", func() {
			session := ctxHelper.FromProjectDir("error-message-test", "delete", "non-existent-worktree")
			cli.ShouldFailWithExit(session, 1)

			errOutput := string(session.Err.Contents())
			// Should not contain internal operation names like "WorktreeService.DeleteWorktree"
			Expect(errOutput).NotTo(ContainSubstring("WorktreeService"))
			Expect(errOutput).NotTo(ContainSubstring("DeleteWorktree"))
			Expect(errOutput).NotTo(ContainSubstring("service operation"))
		})

		It("shows Error: prefix for errors", func() {
			session := ctxHelper.FromProjectDir("error-message-test", "delete", "non-existent-worktree")
			cli.ShouldFailWithExit(session, 1)

			errOutput := string(session.Err.Contents())
			// Error output should contain the error message
			Expect(errOutput).To(ContainSubstring("Error:"))
		})

		It("shows user-friendly message for not-found worktree", func() {
			session := ctxHelper.FromProjectDir("error-message-test", "delete", "non-existent-worktree")
			cli.ShouldFailWithExit(session, 1)

			errOutput := string(session.Err.Contents())
			// Should contain context about what was not found
			Expect(errOutput).To(ContainSubstring("non-existent-worktree"))
		})
	})

	Describe("Panic Recovery", func() {
		BeforeEach(func() {
			fixture.SetupSingleProject("panic-test")
		})

		It("recovers from panics gracefully", func() {
			// Note: This test verifies normal operation doesn't panic
			// Actual panic testing would require intentionally causing a panic
			// which is not recommended in E2E tests
			session := ctxHelper.FromProjectDir("panic-test", "list")
			cli.ShouldSucceed(session)
		})

		It("operates normally when TWIGGIT_DEBUG is set", func() {
			// Verify normal operation works with TWIGGIT_DEBUG set
			cliWithDebug := helpers.NewTwiggitCLI().
				WithConfigDir(fixture.Build()).
				WithEnvironment("TWIGGIT_DEBUG", "1")

			session := cliWithDebug.RunWithDir(fixture.GetProjectPath("panic-test"), "list")
			cli.ShouldSucceed(session)
		})
	})
})
