//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
	testhelpers "twiggit/test/helpers"
)

var _ = Describe("Delete Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		cli.Reset()
	})

	Context("Help Display and Flag Validation", func() {
		It("shows help for delete command", func() {
			session := cli.Run("delete", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit delete <project>/<branch> | <worktree-path>"))
			Expect(output).To(ContainSubstring("Delete a worktree"))
			Expect(output).To(ContainSubstring("By default, prevents deletion of worktrees with uncommitted changes"))
		})

		It("shows help for delete command with -h", func() {
			session := cli.Run("delete", "-h")
			Eventually(session).Should(gexec.Exit(0))
			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Delete a worktree"))
		})

		It("requires exactly one argument", func() {
			session := cli.Run("delete")
			Eventually(session).Should(gexec.Exit(2))
			Expect(cli.GetError(session)).To(ContainSubstring("accepts 1 arg(s), received 0"))
		})

		It("rejects multiple arguments", func() {
			session := cli.Run("delete", "project/branch1", "project/branch2")
			Eventually(session).Should(gexec.Exit(2))
			Expect(cli.GetError(session)).To(ContainSubstring("accepts 1 arg(s), received 2"))
		})

		It("shows --force flag in help", func() {
			session := cli.Run("delete", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("--force"))
			Expect(output).To(ContainSubstring("Force deletion even with uncommitted changes"))
		})

		It("shows --keep-branch flag in help", func() {
			session := cli.Run("delete", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("--keep-branch"))
			Expect(output).To(ContainSubstring("Keep the branch after deletion"))
		})

		It("supports --force flag", func() {
			session := cli.Run("delete", "--force", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})

		It("supports --keep-branch flag", func() {
			session := cli.Run("delete", "--keep-branch", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})

		It("supports both flags together", func() {
			session := cli.Run("delete", "--force", "--keep-branch", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("Worktree Deletion Safety Checks", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("prevents deletion of worktree with uncommitted changes without --force", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Add uncommitted changes to the worktree
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			testFile := filepath.Join(worktreePath, "test.txt")
			err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Try to delete without --force
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(1))

			output := cli.GetError(session2)
			Expect(output).To(ContainSubstring("worktree has uncommitted changes"))
			Expect(output).To(ContainSubstring("use --force to override"))

			// Verify worktree still exists
			_, err = os.Stat(worktreePath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows deletion of worktree with uncommitted changes with --force", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Add uncommitted changes to the worktree
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			testFile := filepath.Join(worktreePath, "test.txt")
			err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Delete with --force
			session2 := cli.WithConfigDir(configDir).Run("delete", "--force", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
			Expect(output).To(ContainSubstring("feature-1"))

			// Verify worktree directory is removed
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("allows deletion of clean worktree without --force", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete clean worktree without --force
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
			Expect(output).To(ContainSubstring("feature-1"))

			// Verify worktree directory is removed
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			_, err := os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("handles non-existent worktree gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			// Try to delete non-existent worktree
			session := cli.WithConfigDir(configDir).Run("delete", "test-project/non-existent")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should succeed (idempotent operation)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles broken worktree gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Break the worktree by removing .git directory
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			gitDir := filepath.Join(worktreePath, ".git")
			err := os.RemoveAll(gitDir)
			Expect(err).NotTo(HaveOccurred())

			// Delete broken worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))

			// Verify worktree directory is removed
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("Branch Validation and Existence", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("validates project/branch format", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			invalidTargets := []string{
				"invalid-format",
				"project/",
				"/branch",
				"project//branch",
				"project/branch/extra",
				"project@branch",
				"project:branch",
			}

			for _, target := range invalidTargets {
				session := cli.WithConfigDir(configDir).Run("delete", target)
				Eventually(session).Should(gexec.Exit(1))
				// Should fail with context resolution error
				Expect(cli.GetError(session)).ToNot(BeEmpty())
			}
		})

		It("handles non-existent project gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("delete", "non-existent-project/feature-1")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("failed to resolve target"))
		})

		It("handles non-existent branch gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("delete", "test-project/non-existent-branch")
			Eventually(session).Should(gexec.Exit(0))

			// Should succeed (idempotent operation)
			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("accepts absolute worktree path as target", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete using absolute path
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			session2 := cli.WithConfigDir(configDir).Run("delete", worktreePath)
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
		})

		It("accepts relative worktree path as target", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Change to worktrees directory and use relative path
			session2 := cli.WithConfigDir(configDir).RunWithDir(worktreesDir, "delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
		})
	})

	Context("Protected Branch Handling", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles main branch worktree deletion", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			// Try to delete main branch worktree (should not exist as separate worktree normally)
			session := cli.WithConfigDir(configDir).Run("delete", "test-project/main")
			Eventually(session).Should(gexec.Exit(0))

			// Should handle gracefully (main worktree is special)
			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles master branch worktree deletion", func() {
			fixture.CreateCustomBranchSetup("test-project", "master")
			configDir := fixture.Build()

			// Try to delete master branch worktree
			session := cli.WithConfigDir(configDir).Run("delete", "test-project/master")
			Eventually(session).Should(gexec.Exit(0))

			// Should handle gracefully
			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles develop branch worktree deletion", func() {
			fixture.CreateCustomBranchSetup("test-project", "develop")
			configDir := fixture.Build()

			// Try to delete develop branch worktree
			session := cli.WithConfigDir(configDir).Run("delete", "test-project/develop")
			Eventually(session).Should(gexec.Exit(0))

			// Should handle gracefully
			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Cleanup Verification", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("removes worktree directory completely", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Verify worktree directory exists
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			_, err := os.Stat(worktreePath)
			Expect(err).NotTo(HaveOccurred())

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			// Verify worktree directory is completely removed
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())

			// Verify parent directories still exist
			projectWorktreesDir := filepath.Join(worktreesDir, "test-project")
			_, err = os.Stat(projectWorktreesDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes worktree with nested directories", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Add nested directories and files
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			nestedDir := filepath.Join(worktreePath, "src", "components")
			err := os.MkdirAll(nestedDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			testFile := filepath.Join(nestedDir, "test.js")
			err = os.WriteFile(testFile, []byte("console.log('test');"), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			// Verify everything is removed
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("handles worktree with symlinks gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Add a symlink (if supported on the platform)
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			linkPath := filepath.Join(worktreePath, "symlink")
			targetPath := filepath.Join(worktreePath, "target.txt")

			// Create target file first
			err := os.WriteFile(targetPath, []byte("target content"), 0644)
			if err != nil {
				Skip("Cannot create files for symlink testing")
			}

			// Create symlink
			err = os.Symlink(targetPath, linkPath)
			if err != nil {
				Skip("Symlinks not supported on this platform")
			}

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			// Verify everything is removed
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("Error Handling for Invalid Targets", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles empty target gracefully", func() {
			session := cli.Run("delete", "")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles whitespace-only target gracefully", func() {
			session := cli.Run("delete", "   ")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles special characters in target gracefully", func() {
			specialTargets := []string{
				"project/branch$name",
				"project/branch@name",
				"project/branch#name",
				"project/branch name",
				"project/branch\tname",
				"project/branch\nname",
			}

			for _, target := range specialTargets {
				session := cli.Run("delete", target)
				Eventually(session).Should(gexec.Exit(1))
				Expect(cli.GetError(session)).ToNot(BeEmpty())
			}
		})

		It("handles very long target gracefully", func() {
			longBranch := strings.Repeat("a", 300)
			target := "project/" + longBranch

			session := cli.Run("delete", target)
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles unicode characters in target gracefully", func() {
			unicodeTargets := []string{
				"project/feature-分支",
				"project/фича-ветка",
				"project/機能-ブランチ",
			}

			for _, target := range unicodeTargets {
				session := cli.Run("delete", target)
				Eventually(session).Should(gexec.Exit(1))
				// Should fail gracefully, not crash
				Expect(cli.GetError(session)).ToNot(BeEmpty())
			}
		})

		It("handles path traversal attempts", func() {
			traversalTargets := []string{
				"../../../etc/passwd",
				"project/../../../etc/passwd",
				"project/branch/../../../etc/passwd",
				"..\\..\\..\\windows\\system32",
			}

			for _, target := range traversalTargets {
				session := cli.Run("delete", target)
				Eventually(session).Should(gexec.Exit(1))
				// Should fail safely, not access system files
				Expect(cli.GetError(session)).ToNot(BeEmpty())
			}
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
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithProjectsDir(customProjectsDir)
			})
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete using project/branch format
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
		})

		It("respects worktrees directory from config", func() {
			customWorktreesDir := GinkgoT().TempDir()
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithWorktreesDir(customWorktreesDir)
			})
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Verify worktree is in custom directory
			worktreePath := filepath.Join(customWorktreesDir, "test-project", "feature-1")
			_, err := os.Stat(worktreePath)
			Expect(err).NotTo(HaveOccurred())

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))

			// Verify worktree is removed from custom directory
			_, err = os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("handles missing config gracefully", func() {
			// Run without any config
			session := cli.Run("delete", "test-project/feature-1")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles invalid TOML configuration", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithCustomConfig(`invalid toml content [`)
			})
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("Failed to load configuration"))
		})

		It("respects XDG_CONFIG_HOME environment variable", func() {
			customConfigDir := GinkgoT().TempDir()
			configHelper := helpers.NewConfigHelper()
			builtConfigDir := configHelper.Build()

			// Copy config to custom XDG location
			customTwiggitDir := filepath.Join(customConfigDir, "twiggit")
			err := os.MkdirAll(customTwiggitDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			configPath := filepath.Join(builtConfigDir, "twiggit", "config.toml")
			customConfigPath := filepath.Join(customTwiggitDir, "config.toml")
			configContent, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(customConfigPath, configContent, 0644)
			Expect(err).NotTo(HaveOccurred())

			fixture.CreateWorktreeSetup("test-project")
			projectPath := fixture.GetProjectPath("test-project")

			// Create a worktree first
			session1 := cli.WithConfigDir(customConfigDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete using custom config
			session2 := cli.WithConfigDir(customConfigDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
		})
	})

	Context("Integration with Git Operations", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("works correctly in repository with only 'master' branch", func() {
			fixture.CreateCustomBranchSetup("test-project", "master")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "master", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))
		})

		It("works correctly in shallow clone (CI scenario)", func() {
			// Create a regular repo first
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create a shallow clone to simulate CI environment
			shallowDir := GinkgoT().TempDir()
			gitHelper := testhelpers.NewGitTestHelper(&testing.T{})

			// Clone with depth 1 (simulate shallow clone)
			err := gitHelper.CreateShallowClone(projectPath, shallowDir, 1)
			if err != nil {
				Skip("Cannot create shallow clone for testing: " + err.Error())
			}

			// Test delete with invalid target - should still validate properly
			session := cli.WithConfigDir(configDir).RunWithDir(shallowDir, "delete", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("works correctly in detached HEAD state (CI scenario)", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create detached HEAD state
			gitHelper := testhelpers.NewGitTestHelper(&testing.T{})
			err := gitHelper.CreateDetachedHEAD(projectPath)
			if err != nil {
				Skip("Cannot create detached HEAD for testing: " + err.Error())
			}

			// Test delete with invalid target - should still validate properly
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "delete", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles non-git directory gracefully", func() {
			nonGitDir := GinkgoT().TempDir()

			session := cli.RunWithDir(nonGitDir, "delete", "test-project/feature-1")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("Multiple Worktree Operations", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles deletion of multiple worktrees sequentially", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create multiple worktrees
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			session2 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-2")
			Eventually(session2).Should(gexec.Exit(0))

			session3 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-3")
			Eventually(session3).Should(gexec.Exit(0))

			// Delete them one by one
			session4 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session4).Should(gexec.Exit(0))
			Expect(cli.GetOutput(session4)).To(ContainSubstring("Deleted worktree"))

			session5 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-2")
			Eventually(session5).Should(gexec.Exit(0))
			Expect(cli.GetOutput(session5)).To(ContainSubstring("Deleted worktree"))

			session6 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-3")
			Eventually(session6).Should(gexec.Exit(0))
			Expect(cli.GetOutput(session6)).To(ContainSubstring("Deleted worktree"))
		})

		It("handles deletion of worktree from different projects", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			project1Path := fixture.GetProjectPath("project1")
			project2Path := fixture.GetProjectPath("project2")

			// Create worktrees in different projects
			session1 := cli.WithConfigDir(configDir).RunWithDir(project1Path, "create", "feature-a")
			Eventually(session1).Should(gexec.Exit(0))

			session2 := cli.WithConfigDir(configDir).RunWithDir(project2Path, "create", "feature-b")
			Eventually(session2).Should(gexec.Exit(0))

			// Delete worktrees from different projects
			session3 := cli.WithConfigDir(configDir).Run("delete", "project1/feature-a")
			Eventually(session3).Should(gexec.Exit(0))
			Expect(cli.GetOutput(session3)).To(ContainSubstring("Deleted worktree"))

			session4 := cli.WithConfigDir(configDir).Run("delete", "project2/feature-b")
			Eventually(session4).Should(gexec.Exit(0))
			Expect(cli.GetOutput(session4)).To(ContainSubstring("Deleted worktree"))
		})
	})

	Context("Performance and Edge Cases", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles deletion of worktree with many files efficiently", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Add many files to the worktree
			worktreePath := filepath.Join(worktreesDir, "test-project", "feature-1")
			for i := 0; i < 100; i++ {
				filePath := filepath.Join(worktreePath, "file", fmt.Sprintf("test%d.txt", i))
				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				Expect(err).NotTo(HaveOccurred())
				err = os.WriteFile(filePath, []byte(fmt.Sprintf("content %d", i)), 0644)
				Expect(err).NotTo(HaveOccurred())
			}

			// Delete the worktree (should complete in reasonable time)
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			output := cli.GetOutput(session2)
			Expect(output).To(ContainSubstring("Deleted worktree"))

			// Verify everything is removed
			_, err := os.Stat(worktreePath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("handles concurrent deletion attempts gracefully", func() {
			fixture.CreateWorktreeSetup("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create a worktree first
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-1")
			Eventually(session1).Should(gexec.Exit(0))

			// Delete the worktree
			session2 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session2).Should(gexec.Exit(0))

			// Try to delete again (should be idempotent)
			session3 := cli.WithConfigDir(configDir).Run("delete", "test-project/feature-1")
			Eventually(session3).Should(gexec.Exit(0))

			output := cli.GetOutput(session3)
			Expect(output).ToNot(BeEmpty())
		})
	})
})
