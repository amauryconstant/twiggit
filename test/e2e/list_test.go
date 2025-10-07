//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("List Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		cli.Reset()
	})

	Context("Help and Usage", func() {
		It("shows help for list command", func() {
			session := cli.Run("list", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit list"))
			Expect(output).To(ContainSubstring("List worktrees"))
		})

		It("shows help with -h flag", func() {
			session := cli.Run("list", "-h")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit list"))
		})

		It("shows available flags in help", func() {
			session := cli.Run("list", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("--all"))
			Expect(output).To(ContainSubstring("List worktrees from all projects"))
		})

		It("rejects extra arguments", func() {
			session := cli.Run("list", "extra-arg")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unknown command"))
		})

		It("supports --all flag in help combination", func() {
			session := cli.Run("list", "--help", "--all")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("Basic Functionality", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("shows no worktrees for empty project", func() {
			// Setup single project
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			// Run list command from project directory
			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "list")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("No worktrees found"))
		})

		It("handles non-git directory gracefully", func() {
			tempDir := GinkgoT().TempDir()

			session := cli.RunWithDir(tempDir, "list")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			// Should show context detection error or similar
			Expect(output).ToNot(BeEmpty())
		})

		It("works with --all flag", func() {
			// Setup multiple projects
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("list", "--all")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should not error out, even if no worktrees exist
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Configuration Integration", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("respects projects directory from config", func() {
			// Setup with custom projects directory
			customProjectsDir := GinkgoT().TempDir()
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithProjectsDir(customProjectsDir)
			})
			fixture.SetupSingleProject("custom-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("list", "--all")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles missing config gracefully", func() {
			// Run without any config
			session := cli.Run("list")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Output Format", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("provides clear output format", func() {
			fixture.SetupSingleProject("format-test")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("format-test")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "list")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should contain either worktrees or "No worktrees found"
			Expect(output).To(SatisfyAny(
				ContainSubstring("No worktrees found"),
				MatchRegexp(`\w+.*->.*`),
			))
		})

		It("handles worktree status information", func() {
			// This test would be more meaningful with actual worktrees
			// For now, just ensure the command doesn't crash
			fixture.CreateWorktreeSetup("status-test")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("status-test")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "list")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Error Handling", func() {
		It("handles invalid flag combinations", func() {
			session := cli.Run("list", "--invalid-flag")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unknown flag"))
		})

		It("handles permission issues gracefully", func() {
			// Create a directory with restricted permissions
			tempDir := GinkgoT().TempDir()
			restrictedDir := tempDir + "/restricted"

			err := os.Mkdir(restrictedDir, 0000)
			if err != nil {
				Skip("Cannot create restricted directory for testing")
			}
			defer os.Chmod(restrictedDir, 0755) // Cleanup

			// Try to run list command from restricted directory
			session := cli.RunWithDir(restrictedDir, "list")
			Eventually(session).Should(gexec.Exit(1))

			// Should handle the permission error gracefully
			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Performance and Scalability", func() {
		It("handles large number of projects efficiently", func() {
			fixture := fixtures.NewE2ETestFixture()
			defer fixture.Cleanup()

			// Create multiple projects
			for i := 0; i < 5; i++ {
				projectName := fmt.Sprintf("project-%d", i)
				fixture.SetupSingleProject(projectName)
			}

			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("list", "--all")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})
	})
})
