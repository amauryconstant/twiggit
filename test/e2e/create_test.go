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

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Create Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for create command", func() {
		session := cli.Run("create", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit create [branch-name]"))
		Expect(output).To(ContainSubstring("Create a new Git worktree"))
	})

	It("shows help for create command with -h", func() {
		session := cli.Run("create", "-h")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("Examples:"))
	})

	It("requires at most one argument", func() {
		session := cli.Run("create", "branch1", "branch2")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Out.Contents())).To(ContainSubstring("❌"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("accepts at most 1 arg(s), received 2"))
	})

	It("errors when no branch name provided and interactive mode not implemented", func() {
		session := cli.Run("create")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Out.Contents())).To(ContainSubstring("❌"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("interactive mode not yet implemented"))
	})

	It("errors when empty branch name provided", func() {
		session := cli.Run("create", "")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Out.Contents())).To(ContainSubstring("❌"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("branch name is required"))
	})

	It("supports verbose flag", func() {
		session := cli.Run("create", "--help", "--verbose")
		Eventually(session).Should(gexec.Exit(0))
	})

	It("supports quiet flag", func() {
		session := cli.Run("create", "--help", "--quiet")
		Eventually(session).Should(gexec.Exit(0))
	})

	Describe("default_source_branch configuration", func() {
		var tempDir string
		var cleanup func()
		var configDir string
		var configPath string

		BeforeEach(func() {
			// Create a temporary directory for testing
			var err error
			tempDir, err = os.MkdirTemp("", "twiggit-config-test")
			Expect(err).NotTo(HaveOccurred())

			cleanup = func() {
				os.RemoveAll(tempDir)
			}

			configDir = filepath.Join(tempDir, "twiggit")
			configPath = filepath.Join(configDir, "config.toml")

			// Create config directory
			err = os.MkdirAll(configDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Set XDG_CONFIG_HOME to point to our temp directory so the config is found
			oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)

			// Update cleanup to restore environment variable
			originalCleanup := cleanup
			cleanup = func() {
				if oldXdgConfigHome != "" {
					os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
				} else {
					os.Unsetenv("XDG_CONFIG_HOME")
				}
				originalCleanup()
			}

			// Initialize a git repository for testing using the git helper
			t := &testing.T{}
			gitRepo := helpers.NewGitRepo(t, "twiggit-e2e-test")
			tempDir = gitRepo.Path

			// Update cleanup to include git repo cleanup
			currentCleanup := cleanup
			cleanup = func() {
				gitRepo.Cleanup()
				currentCleanup()
			}
		})

		AfterEach(func() {
			cleanup()
		})

		It("uses default_source_branch from config when no --source flag provided", func() {
			// Create config with custom default source branch
			configContent := `default_source_branch = "develop"`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Run create command with custom environment
			session := cli.RunWithDir(tempDir, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1)) // Will fail due to no git repo, but we can check the error message

			output := string(session.Out.Contents())
			// The error should mention 'develop' branch, not 'main'
			Expect(output).To(ContainSubstring("source branch 'develop' does not exist"))
			Expect(output).NotTo(ContainSubstring("source branch 'main' does not exist"))
		})

		It("overrides config default_source_branch with --source flag", func() {
			// Create config with default source branch
			configContent := `default_source_branch = "develop"`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Run create command with --source flag
			session := cli.RunWithDir(tempDir, "create", "--source", "main", "feature-branch")
			Eventually(session).Should(gexec.Exit(1)) // Will fail due to no git repo

			output := string(session.Out.Contents())
			// The error should mention 'main' branch (from --source flag), not 'develop' (from config)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output).NotTo(ContainSubstring("source branch 'develop' does not exist"))
		})

		It("falls back to 'main' when no config and no --source flag", func() {
			// No config file created

			// Run create command without any configuration
			session := cli.RunWithDir(tempDir, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1)) // Will fail due to no git repo

			output := string(session.Out.Contents())
			// The error should mention 'main' branch (default fallback)
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
		})

		It("respects configuration priority: --source flag > config > default", func() {
			// Create config with default source branch
			configContent := `default_source_branch = "develop"`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Test 1: Default behavior (should use config)
			session1 := cli.RunWithDir(tempDir, "create", "feature-branch")
			Eventually(session1).Should(gexec.Exit(1))
			output1 := string(session1.Out.Contents())
			Expect(output1).To(ContainSubstring("source branch 'develop' does not exist"))

			// Test 2: With --source flag (should override config)
			session2 := cli.RunWithDir(tempDir, "create", "--source", "main", "feature-branch")
			Eventually(session2).Should(gexec.Exit(1))
			output2 := string(session2.Out.Contents())
			Expect(output2).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output2).NotTo(ContainSubstring("source branch 'develop' does not exist"))
		})

		It("handles invalid default_source_branch in config", func() {
			// Create config with invalid branch name
			configContent := `default_source_branch = "invalid@branch#name"`
			err := os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Run create command - should fail at config loading
			session := cli.RunWithDir(tempDir, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Err.Contents())
			// Should show config validation error for invalid branch name
			Expect(output).To(ContainSubstring("Failed to load configuration"))
			Expect(output).To(ContainSubstring("invalid default source branch name"))
		})
	})

	Describe("validation order and error priority", func() {
		var tempDir string
		var cleanup func()
		var gitRepo *helpers.GitRepo

		BeforeEach(func() {
			t := &testing.T{}
			gitRepo = helpers.NewGitRepo(t, "twiggit-validation-test")
			tempDir = gitRepo.Path

			cleanup = func() {
				gitRepo.Cleanup()
			}
		})

		AfterEach(func() {
			cleanup()
		})

		It("validates branch name format before checking source branch existence", func() {
			// Create a git repo with only the default branch (no 'main' branch)
			// This simulates CI environment where 'main' might not exist

			// Run create command with invalid branch name
			session := cli.RunWithDir(tempDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			// Should fail with branch name validation error, not source branch error
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("validates branch name format before checking if we're in a git repository", func() {
			// Run create command with invalid branch name from a non-git directory
			nonGitDir, tempCleanup := helpers.TempDir(&testing.T{}, "non-git-dir")
			defer tempCleanup()

			session := cli.RunWithDir(nonGitDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			// Should fail with branch name validation error, not git repository error
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("not a git repository"))
		})

		It("checks source branch existence only after branch name validation passes", func() {
			// Use a valid branch name format but non-existent source branch
			session := cli.RunWithDir(tempDir, "create", "valid-branch-name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			// Should fail with source branch error, not branch name validation error
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
			Expect(output).NotTo(ContainSubstring("branch name format is invalid"))
		})

		It("handles multiple validation errors correctly", func() {
			// Create a git repo with a branch that has invalid characters
			t := &testing.T{}
			gitRepoWithBranches := helpers.NewGitRepoWithBranches(t, "twiggit-multi-validation", []string{"develop"})
			defer gitRepoWithBranches.Cleanup()

			// Test with invalid branch name - should only show branch name validation error
			session := cli.RunWithDir(gitRepoWithBranches.Path, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			// Should not mention other potential errors
			Expect(output).NotTo(ContainSubstring("source branch"))
		})
	})

	Describe("git repository state scenarios", func() {
		var tempDir string
		var cleanup func()

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "twiggit-git-state-test")
			Expect(err).NotTo(HaveOccurred())

			cleanup = func() {
				os.RemoveAll(tempDir)
			}
		})

		AfterEach(func() {
			cleanup()
		})

		It("works correctly in repository with only 'master' branch (no 'main')", func() {
			// Create a git repository and rename default branch to 'master'
			t := &testing.T{}
			gitRepo := helpers.NewGitRepo(t, "twiggit-master-branch")
			defer gitRepo.Cleanup()

			// Rename the default branch to 'master' instead of 'main'
			gitRepo.GitCmd(t, "branch", "-m", "master")

			// Test with valid branch name - should fail because source branch 'main' doesn't exist
			session := cli.RunWithDir(gitRepo.Path, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))

			// Test with --source flag pointing to 'master' - should work differently
			session2 := cli.RunWithDir(gitRepo.Path, "create", "--source", "master", "feature-branch")
			Eventually(session2).Should(gexec.Exit(1))

			output2 := string(session2.Out.Contents())
			// Should not fail with source branch error since 'master' exists
			Expect(output2).NotTo(ContainSubstring("source branch 'master' does not exist"))
			// But might fail with other errors (like worktree creation, which is expected)
		})

		It("works correctly in repository with custom default branch", func() {
			// Create a git repository with custom default branch name
			t := &testing.T{}
			gitRepo := helpers.NewGitRepo(t, "twiggit-custom-branch")
			defer gitRepo.Cleanup()

			// Rename the default branch to something custom
			gitRepo.GitCmd(t, "branch", "-m", "develop")

			// Test with valid branch name - should fail because source branch 'main' doesn't exist
			session := cli.RunWithDir(gitRepo.Path, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("source branch 'main' does not exist"))
		})

		It("works correctly in shallow clone (CI scenario)", func() {
			// Create a git repository and simulate shallow clone
			t := &testing.T{}
			gitRepo := helpers.NewGitRepo(t, "twiggit-shallow-clone")
			defer gitRepo.Cleanup()

			// Create a shallow clone (depth 1) to simulate CI environment
			shallowDir, shallowCleanup := helpers.TempDir(t, "shallow-clone")
			defer shallowCleanup()

			// Clone with depth 1
			gitRepo.GitCmd(t, "clone", "--depth", "1", "file://"+gitRepo.Path, shallowDir)

			// Test with invalid branch name - should still validate branch name first
			session := cli.RunWithDir(shallowDir, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})

		It("works correctly in detached HEAD state (CI scenario)", func() {
			// Create a git repository and simulate detached HEAD
			t := &testing.T{}
			gitRepo := helpers.NewGitRepo(t, "twiggit-detached-head")
			defer gitRepo.Cleanup()

			// Create a commit and checkout detached HEAD
			gitRepo.GitCmd(t, "checkout", "--detach", "HEAD")

			// Test with invalid branch name - should still validate branch name first
			session := cli.RunWithDir(gitRepo.Path, "create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
			Expect(output).NotTo(ContainSubstring("source branch"))
		})
	})

	Describe("-C/--change-dir flag", func() {
		var tempDir string
		var cleanup func()
		var gitRepo *helpers.GitRepo
		var configDir string
		var configPath string
		var workspaceDir string

		BeforeEach(func() {
			// Create a temporary directory for testing
			var err error
			tempDir, err = os.MkdirTemp("", "twiggit-change-dir-test")
			Expect(err).NotTo(HaveOccurred())

			cleanup = func() {
				os.RemoveAll(tempDir)
			}

			configDir = filepath.Join(tempDir, "twiggit")
			configPath = filepath.Join(configDir, "config.toml")
			workspaceDir = filepath.Join(tempDir, "workspaces")

			// Create config and workspace directories
			err = os.MkdirAll(configDir, 0755)
			Expect(err).NotTo(HaveOccurred())
			err = os.MkdirAll(workspaceDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Set XDG_CONFIG_HOME to point to our temp directory so the config is found
			oldXdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)

			// Update cleanup to restore environment variable
			originalCleanup := cleanup
			cleanup = func() {
				if oldXdgConfigHome != "" {
					os.Setenv("XDG_CONFIG_HOME", oldXdgConfigHome)
				} else {
					os.Unsetenv("XDG_CONFIG_HOME")
				}
				originalCleanup()
			}

			// Initialize a git repository for testing using the git helper
			t := &testing.T{}
			gitRepo = helpers.NewGitRepo(t, "twiggit-change-dir-test")
			tempDir = gitRepo.Path

			// Update cleanup to include git repo cleanup
			currentCleanup := cleanup
			cleanup = func() {
				gitRepo.Cleanup()
				currentCleanup()
			}

			// Ensure we have a 'main' branch for the tests
			// The NewGitRepo helper might create 'master' as default, so switch to 'main'
			gitRepo.GitCmd(t, "checkout", "-b", "main")

			// Create the project directory under workspace (worktree creator expects this structure)
			projectName := filepath.Base(gitRepo.Path)
			projectDir := filepath.Join(workspaceDir, projectName)
			err = os.MkdirAll(projectDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			// Create config with workspace path
			configContent := fmt.Sprintf(`workspaces_path = "%s"`, workspaceDir)
			err = os.WriteFile(configPath, []byte(configContent), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			cleanup()
		})

		It("shows -C/--change-dir flag in help", func() {
			session := cli.Run("create", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("-C, --change-dir"))
			Expect(output).To(ContainSubstring("Change to new worktree directory after creation"))
		})

		It("supports -C flag (short form)", func() {
			session := cli.Run("create", "-C", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})

		It("supports --change-dir flag (long form)", func() {
			session := cli.Run("create", "--change-dir", "--help")
			Eventually(session).Should(gexec.Exit(0))
		})

		It("outputs worktree path when -C flag is used", func() {
			session := cli.RunWithDir(tempDir, "create", "-C", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			// Should output the path that would be changed to
			Expect(output).To(ContainSubstring("/feature-branch"))
			// Should contain success message
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))
			// The path should be the last line (for shell wrapper consumption)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			Expect(lines[len(lines)-1]).To(ContainSubstring("/feature-branch"))
		})

		It("outputs worktree path when --change-dir flag is used", func() {
			session := cli.RunWithDir(tempDir, "create", "--change-dir", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			// Should output the path that would be changed to
			Expect(output).To(ContainSubstring("/feature-branch"))
			// Should contain success message
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))
			// The path should be the last line (for shell wrapper consumption)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			Expect(lines[len(lines)-1]).To(ContainSubstring("/feature-branch"))
		})

		It("does not output path when -C flag is not used", func() {
			session := cli.RunWithDir(tempDir, "create", "feature-branch")
			Eventually(session).Should(gexec.Exit(0))

			output := string(session.Out.Contents())
			// Should contain success message
			Expect(output).To(ContainSubstring("✅ Worktree created successfully"))
			// Should contain navigate message
			Expect(output).To(ContainSubstring("Navigate: cd"))
			// Should not output the path as a standalone line when flag is not used
			lines := strings.Split(strings.TrimSpace(output), "\n")
			// Last line should be the navigate message, not a standalone path
			lastLine := lines[len(lines)-1]
			Expect(lastLine).To(ContainSubstring("Navigate: cd"))
			// The last line should not be just the path (it should be the navigate message)
			expectedPath := strings.TrimSpace(filepath.Join(workspaceDir, filepath.Base(tempDir), "feature-branch"))
			Expect(lastLine).NotTo(Equal(expectedPath))
		})
	})
})
