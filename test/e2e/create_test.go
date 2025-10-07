//go:build e2e
// +build e2e

package e2e

import (
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

var _ = Describe("Create Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		cli.Reset()
	})

	Context("Help Display and Flag Validation", func() {
		It("shows help for create command", func() {
			session := cli.Run("create", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit create <project>/<branch> | <branch>"))
			Expect(output).To(ContainSubstring("Create a new worktree"))
		})

		It("shows help for create command with -h", func() {
			session := cli.Run("create", "-h")
			Eventually(session).Should(gexec.Exit(0))
			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Create a new worktree"))
		})

		It("requires at most one argument", func() {
			session := cli.Run("create", "branch1", "branch2")
			Eventually(session).Should(gexec.Exit(2))
			Expect(cli.GetError(session)).To(ContainSubstring("accepts 1 arg(s), received 2"))
		})

		It("errors when no branch name provided", func() {
			session := cli.Run("create")
			Eventually(session).Should(gexec.Exit(2))
			Expect(cli.GetError(session)).To(ContainSubstring("accepts 1 arg(s), received 0"))
		})

		It("errors when empty branch name provided", func() {
			session := cli.Run("create", "")
			Eventually(session).Should(gexec.Exit(2))
			Expect(cli.GetError(session)).To(ContainSubstring("failed to discover project"))
		})

		It("shows --cd flag in help", func() {
			session := cli.Run("create", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("--cd string"))
			Expect(output).To(ContainSubstring("Change directory after creation"))
		})

		It("supports --cd flag", func() {
			session := cli.Run("create", "--cd", "/tmp", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("Branch Name Validation", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("rejects branch names with invalid characters", func() {
			invalidNames := []string{
				"invalid@branch",
				"branch#name",
				"branch$name",
				"branch name",
				"branch/name",
				"branch\\name",
				"branch:name",
				"branch*name",
				"branch?name",
				"branch\"name",
				"branch<name>",
				"branch|name",
			}

			for _, name := range invalidNames {
				session := cli.Run("create", name)
				Eventually(session).Should(gexec.Exit(2))
				Expect(cli.GetError(session)).To(ContainSubstring("failed to discover project"))
			}
		})

		It("rejects branch names that start with invalid characters", func() {
			invalidNames := []string{
				"-branch",
				".branch",
				"@branch",
				"#branch",
			}

			for _, name := range invalidNames {
				session := cli.Run("create", name)
				Eventually(session).Should(gexec.Exit(1))
				Expect(cli.GetError(session)).To(ContainSubstring("branch name format is invalid"))
			}
		})

		It("rejects branch names that end with invalid characters", func() {
			invalidNames := []string{
				"branch-",
				"branch.",
				"branch@",
				"branch#",
			}

			for _, name := range invalidNames {
				session := cli.Run("create", name)
				Eventually(session).Should(gexec.Exit(1))
				Expect(cli.GetError(session)).To(ContainSubstring("branch name format is invalid"))
			}
		})

		It("accepts valid branch names", func() {
			validNames := []string{
				"feature-branch",
				"bugfix-123",
				"hotfix_v2",
				"release-1.0.0",
				"branch123",
				"a",
				"very-long-branch-name-with-many-parts",
				"branch_with_underscores",
				"123-branch",
			}

			for _, name := range validNames {
				// These will fail for other reasons (like no git repo), but not branch name validation
				session := cli.Run("create", name)
				Eventually(session).Should(gexec.Exit(1))
				errorOutput := cli.GetError(session)
				Expect(errorOutput).NotTo(ContainSubstring("branch name format is invalid"))
			}
		})

		It("rejects branch names that are too long", func() {
			// Create a very long branch name (over 255 characters)
			longName := strings.Repeat("a", 256)
			session := cli.Run("create", longName)
			Eventually(session).Should(gexec.Exit(1))
			Expect(cli.GetError(session)).To(ContainSubstring("branch name format is invalid"))
		})

		It("rejects branch names that are git reserved names", func() {
			reservedNames := []string{
				"HEAD",
				"head",
				"Master",
				"master",
				"Main",
				"main",
				"ORIG_HEAD",
				"FETCH_HEAD",
				"MERGE_HEAD",
			}

			for _, name := range reservedNames {
				session := cli.Run("create", name)
				Eventually(session).Should(gexec.Exit(1))
				errorOutput := cli.GetError(session)
				// Should fail with branch name validation or reserved name error
				Expect(errorOutput).To(SatisfyAny(
					ContainSubstring("branch name format is invalid"),
					ContainSubstring("reserved branch name"),
				))
			}
		})

		It("validates branch name before checking git repository", func() {
			nonGitDir := GinkgoT().TempDir()

			session := cli.RunWithDir(nonGitDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("not a git repository"))
		})
	})

	Context("Source Branch Resolution", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("uses default_source_branch from config when no --source flag provided", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithDefaultSourceBranch("develop")
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'develop' does not exist"))
			Expect(output).NotTo(ContainSubstring("source branch 'main' does not exist"))
		})

		It("overrides config default_source_branch with --source flag", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithDefaultSourceBranch("develop")
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "main", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output).NotTo(ContainSubstring("source branch 'develop' does not exist"))
		})

		It("falls back to 'main' when no config and no --source flag", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
		})

		It("respects configuration priority: --source flag > config > default", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithDefaultSourceBranch("develop")
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Test 1: Default behavior (should use config)
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session1).Should(gexec.Exit(1))
			output1 := cli.GetError(session1)
			Expect(output1).To(ContainSubstring("source branch 'develop' does not exist"))

			// Test 2: With --source flag (should override config)
			session2 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "main", "feature-branch")
			Eventually(session2).Should(gexec.Exit(1))
			output2 := cli.GetError(session2)
			Expect(output2).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output2).NotTo(ContainSubstring("source branch 'develop' does not exist"))
		})

		It("handles invalid default_source_branch in config", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithCustomConfig(`default_source_branch = "invalid@branch#name"`)
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("Failed to load configuration"))
			Expect(output).To(ContainSubstring("invalid default source branch name"))
		})

		It("validates --source flag branch name", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "invalid@branch", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch name format is invalid"))
		})
	})

	Context("Git Repository Scenarios", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("works correctly in repository with only 'master' branch (no 'main')", func() {
			fixture.CreateCustomBranchSetup("test-project", "master")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))

			// Test with --source flag pointing to 'master'
			session2 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "master", "feature-branch")
			Eventually(session2).Should(gexec.Exit(1))

			output2 := cli.GetError(session2)
			Expect(output2).NotTo(ContainSubstring("source branch 'master' does not exist"))
		})

		It("works correctly in repository with custom default branch", func() {
			fixture.CreateCustomBranchSetup("test-project", "develop")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
		})

		It("works correctly in shallow clone (CI scenario)", func() {
			// Create a regular repo first
			fixture.SetupSingleProject("test-project")
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

			// Test with invalid branch name - should still validate branch name first
			session := cli.WithConfigDir(configDir).RunWithDir(shallowDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("works correctly in detached HEAD state (CI scenario)", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create detached HEAD state
			gitHelper := testhelpers.NewGitTestHelper(&testing.T{})
			err := gitHelper.CreateDetachedHEAD(projectPath)
			if err != nil {
				Skip("Cannot create detached HEAD for testing: " + err.Error())
			}

			// Test with invalid branch name - should still validate branch name first
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("handles non-git directory gracefully", func() {
			nonGitDir := GinkgoT().TempDir()

			session := cli.RunWithDir(nonGitDir, "create", "valid-branch-name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})
	})

	Context("--cd Flag Behavior", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("outputs worktree path when --cd flag is used", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--cd", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("feature-branch"))
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))

			// The path should be the last line (for shell wrapper consumption)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			Expect(lines[len(lines)-1]).To(ContainSubstring("feature-branch"))
		})

		It("does not output path when --cd flag is not used", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))
			Expect(output).To(ContainSubstring("Navigate: cd"))

			// Should not output the path as a standalone line when flag is not used
			lines := strings.Split(strings.TrimSpace(output), "\n")
			lastLine := lines[len(lines)-1]
			Expect(lastLine).To(ContainSubstring("Navigate: cd"))
		})

		It("outputs absolute path with --cd flag", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--cd", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			pathLine := lines[len(lines)-1]

			// Should be an absolute path
			Expect(filepath.IsAbs(pathLine)).To(BeTrue())
		})
	})

	Context("Configuration Validation and Error Handling", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles missing config file gracefully", func() {
			fixture.SetupSingleProject("test-project")
			// Don't build config to simulate missing config

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("handles invalid TOML configuration", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithCustomConfig(`invalid toml content [`)
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("Failed to load configuration"))
		})

		It("handles invalid worktrees_dir in config", func() {
			fixture.WithConfig(func(config *helpers.ConfigHelper) {
				config.WithCustomConfig(`worktrees_dir = "/nonexistent/path"`)
			})
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).ToNot(BeEmpty())
		})

		It("respects XDG_CONFIG_HOME environment variable", func() {
			customConfigDir := GinkgoT().TempDir()
			configHelper := helpers.NewConfigHelper()
			configHelper.WithDefaultSourceBranch("develop")
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

			fixture.SetupSingleProject("test-project")
			projectPath := fixture.GetProjectPath("test-project")

			session := cli.WithConfigDir(customConfigDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'develop' does not exist"))
		})
	})

	Context("Validation Order and Error Priority", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("validates branch name format before checking source branch existence", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("validates branch name format before checking if we're in a git repository", func() {
			nonGitDir := GinkgoT().TempDir()

			session := cli.RunWithDir(nonGitDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("not a git repository"))
		})

		It("checks source branch existence only after branch name validation passes", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "valid-branch-name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output).NotTo(ContainSubstring("branch name format is invalid"))
		})

		It("handles multiple validation errors correctly", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			// Should not mention other potential errors
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("validates --source flag before checking source branch existence", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "--source", "invalid@branch", "valid-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("source branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch 'invalid@branch' does not exist"))
		})
	})

	Context("Integration with Existing Worktrees", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("handles existing worktree with same branch name", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")

			// Create first worktree
			session1 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session1).Should(gexec.Exit(0))

			// Try to create another worktree with same branch name
			session2 := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session2).Should(gexec.Exit(1))

			output := cli.GetError(session2)
			Expect(output).To(ContainSubstring("already exists"))
		})

		It("creates worktree in correct directory structure", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			projectPath := fixture.GetProjectPath("test-project")
			session := cli.WithConfigDir(configDir).RunWithDir(projectPath, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))

			// Check that worktree directory was created
			worktreesDir := fixture.GetConfigHelper().GetWorktreesDir()
			expectedWorktreePath := filepath.Join(worktreesDir, "test-project", "feature-branch")

			_, err := os.Stat(expectedWorktreePath)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
