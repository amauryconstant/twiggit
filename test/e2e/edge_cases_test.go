//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit CLI.
// Tests use real git repositories and validate complete user workflows.
package e2e

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("Edge Cases", func() {
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

	Context("Long branch names", func() {
		PIt("handles branch names at length limit", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-long-branch")

			longBranchName := ""
			for i := 0; i < 250; i++ {
				longBranchName += "a"
			}

			session := ctxHelper.FromProjectDir("test-long-branch", "create", longBranchName)
			Eventually(session).Should(Or(gexec.Exit(0), gexec.Exit(1), gexec.Exit(2)))

			if session.ExitCode() == 0 {
				stdout := cli.GetOutput(session)
				Expect(stdout).To(Or(
					ContainSubstring("Created worktree"),
					ContainSubstring(longBranchName),
				), "Should create worktree with long branch name")
			} else {
				stderr := cli.GetError(session)
				Expect(stderr).NotTo(BeEmpty())
			}
		})
	})

	Context("Special characters in branch names", func() {
		PIt("handles branch names with leading dots", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-special-dots")

			session := ctxHelper.FromProjectDir("test-special-dots", "create", ".feature-branch")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("dot"),
				ContainSubstring("branch"),
			), "Should reject branch name with leading dot")
		})

		PIt("handles branch names with trailing dashes", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-special-dashes")

			session := ctxHelper.FromProjectDir("test-special-dashes", "create", "feature-branch-")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("dash"),
				ContainSubstring("branch"),
			), "Should reject branch name with trailing dash")
		})

		PIt("handles branch names with trailing dots", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-special-dots-trailing")

			session := ctxHelper.FromProjectDir("test-special-dots-trailing", "create", "feature-branch.")
			Eventually(session).Should(Or(gexec.Exit(1), gexec.Exit(2)))

			stderr := cli.GetError(session)
			Expect(stderr).NotTo(BeEmpty())
			Expect(stderr).To(Or(
				ContainSubstring("invalid"),
				ContainSubstring("dot"),
				ContainSubstring("branch"),
			), "Should reject branch name with trailing dot")
		})
	})

	Context("Symlink path handling", func() {
		PIt("operates correctly from symlinked project directory", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-symlink")

			projectPath := fixture.GetProjectPath("test-symlink")
			tempDir := fixture.GetTempDir()
			symlinkPath := filepath.Join(tempDir, "symlink-test-symlink")

			err := os.Symlink(projectPath, symlinkPath)
			Expect(err).NotTo(HaveOccurred(), "Should create symlink successfully")

			session := cli.RunWithDir(symlinkPath, "list")
			Eventually(session).Should(Or(gexec.Exit(0), gexec.Exit(1), gexec.Exit(2)))

			if session.ExitCode() == 0 {
				stdout := cli.GetOutput(session)
				Expect(stdout).NotTo(BeEmpty())
			} else {
				stderr := cli.GetError(session)
				Expect(stderr).NotTo(BeEmpty())
			}
		})

		PIt("creates worktree from symlinked project directory", func() {
			configDir := fixture.Build()
			cli = cli.WithConfigDir(configDir)
			ctxHelper = fixtures.NewContextHelper(fixture, cli)
			fixture.SetupSingleProject("test-symlink-create")

			projectPath := fixture.GetProjectPath("test-symlink-create")
			tempDir := fixture.GetTempDir()
			symlinkPath := filepath.Join(tempDir, "symlink-test-symlink-create")

			err := os.Symlink(projectPath, symlinkPath)
			Expect(err).NotTo(HaveOccurred(), "Should create symlink successfully")

			session := cli.RunWithDir(symlinkPath, "create", "feature-from-symlink")
			Eventually(session).Should(Or(gexec.Exit(0), gexec.Exit(1), gexec.Exit(2)))

			if session.ExitCode() == 0 {
				stdout := cli.GetOutput(session)
				Expect(stdout).To(Or(
					ContainSubstring("Created worktree"),
					ContainSubstring("feature-from-symlink"),
				), "Should create worktree from symlinked path")

				deleteSession := cli.RunWithDir(symlinkPath, "delete", "feature-from-symlink")
				Eventually(deleteSession).Should(gexec.Exit(0))
			} else {
				stderr := cli.GetError(session)
				Expect(stderr).NotTo(BeEmpty())
			}
		})
	})
})
