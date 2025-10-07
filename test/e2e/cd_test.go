//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	e2ehelpers "twiggit/test/e2e/helpers"
	"twiggit/test/helpers"
)

var _ = Describe("CD Command", func() {
	var cli *e2ehelpers.TwiggitCLI

	BeforeEach(func() {
		cli = e2ehelpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		cli.Reset()
	})

	Context("Help and Usage", func() {
		It("shows help for cd command", func() {
			session := cli.Run("cd", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit cd"))
			Expect(output).To(ContainSubstring("Change directory"))
		})

		It("shows help with -h flag", func() {
			session := cli.Run("cd", "-h")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit cd"))
		})

		It("shows examples in help", func() {
			session := cli.Run("cd", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Examples:"))
		})

		It("accepts at most one argument", func() {
			session := cli.Run("cd", "project1", "project2")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("accepts at most 1 arg"))
		})

		It("handles project switching format", func() {
			session := cli.Run("cd", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("<project|project/branch>"))
		})
	})

	Context("Context Detection", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("shows context help when no arguments provided", func() {
			tempDir := GinkgoT().TempDir()

			session := cli.RunWithDir(tempDir, "cd")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Current context:"))
			Expect(output).To(ContainSubstring("Available targets:"))
		})

		It("detects project context correctly", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "cd")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Current context:"))
			Expect(output).To(ContainSubstring("project"))
		})

		It("detects outside git context correctly", func() {
			tempDir := GinkgoT().TempDir()

			session := cli.RunWithDir(tempDir, "cd")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Outside git repository"))
		})
	})

	Context("Navigation Functionality", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("navigates to project directory", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "project1")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should output the path to navigate to
			Expect(output).To(ContainSubstring("project1"))
		})

		It("navigates to worktree directory", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "test-project/feature-1")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("feature-1"))
		})

		It("handles relative project names", func() {
			fixture.SetupSingleProject("current-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("current-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "cd", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("feature-branch"))
		})
	})

	Context("Cross-Project Navigation", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("navigates between different projects", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			// Navigate from one project to another
			project1Path := fixture.GetProjectPath("project1")
			session := cli.WithConfigDir(configDir).RunWithDir(project1Path, "cd", "project2")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("project2"))
		})

		It("handles project with worktrees", func() {
			fixture.CreateWorktreeSetup("worktree-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "worktree-project")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("worktree-project"))
		})
	})

	Context("Error Handling", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles non-existent project", func() {
			fixture.SetupSingleProject("existing-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "non-existent-project")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles non-existent worktree", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "test-project/non-existent-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("provides helpful error messages", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "invalid-target")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("not found"))
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
			customProjectsDir := GinkgoT().TempDir()
			fixture.WithConfig(func(config *e2ehelpers.ConfigHelper) {
				config.WithProjectsDir(customProjectsDir)
			})
			fixture.SetupSingleProject("custom-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "custom-project")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("custom-project"))
		})

		It("handles missing configuration gracefully", func() {
			session := cli.Run("cd", "any-project")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Special Cases", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles main branch navigation", func() {
			fixture.SetupSingleProject("main-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "main-project/main")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("main-project"))
		})

		It("handles branch names with slashes", func() {
			fixture.SetupSingleProject("slash-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "slash-project/feature/sub-branch")
			Eventually(session).Should(gexec.Exit(0))

			// Should handle the slash appropriately
			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles project names similar to branch names", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			// Add a branch with same name as another project
			gitHelper := helpers.NewGitTestHelper(&testing.T{})
			project1Path := fixture.GetProjectPath("project1")
			err := gitHelper.CreateBranch(project1Path, "project2")
			Expect(err).NotTo(HaveOccurred())

			session := cli.WithConfigDir(configDir).Run("cd", "project1/project2")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("project2"))
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

		It("outputs clean paths for shell consumption", func() {
			fixture.SetupSingleProject("clean-output")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("cd", "clean-output")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should be a clean path without extra formatting
			Expect(output).ToNot(ContainSubstring("Error:"))
			Expect(output).ToNot(ContainSubstring("Warning:"))
		})

		It("provides context information when no target", func() {
			tempDir := GinkgoT().TempDir()

			session := cli.RunWithDir(tempDir, "cd")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Current context:"))
			Expect(output).To(ContainSubstring("Available targets:"))
		})
	})
})
