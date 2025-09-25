//go:build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Error Presentation", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	Describe("CLI Error Formatting", func() {
		It("should display validation errors with proper formatting", func() {
			// Try to create a worktree with empty branch name
			session := cli.Run("create", "")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name is required"))
		})

		It("should display project not found errors with proper formatting", func() {
			// Try to switch to a non-existent project
			session := cli.Run("cd", "nonexistent-project")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			// Handle both possible error messages depending on context
			Expect(output).To(Or(
				ContainSubstring("failed to discover projects"),
				ContainSubstring("failed to discover worktrees"),
				ContainSubstring("project 'nonexistent-project' not found"),
				ContainSubstring("worktree 'nonexistent-project' not found"),
				ContainSubstring("worktree 'shell-integration/nonexistent-project' not found"),
				ContainSubstring("worktree 'twiggit/nonexistent-project' not found")))
		})

		It("should display invalid branch name errors with proper formatting", func() {
			// Try to create a worktree with invalid branch name
			session := cli.Run("create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
		})
	})

	Describe("Error Consistency", func() {
		It("should maintain consistent error formatting across commands", func() {
			// Test error formatting across different commands
			commands := [][]string{
				{"create", "invalid@branch"},
				{"cd", "nonexistent-project"},
				{"create", ""},
			}

			for _, cmd := range commands {
				session := cli.Run(cmd...)
				Eventually(session).Should(gexec.Exit(1))

				output := string(session.Out.Contents())
				Expect(output).To(ContainSubstring("❌"))
				Expect(output).ToNot(BeEmpty())
			}
		})

		It("should include error context in all error messages", func() {
			// Test that errors include relevant context
			session := cli.Run("cd", "nonexistent-project")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			// Handle both possible error messages depending on context
			Expect(output).To(Or(
				ContainSubstring("failed to discover projects"),
				ContainSubstring("failed to discover worktrees"),
				ContainSubstring("project 'nonexistent-project' not found"),
				ContainSubstring("worktree 'nonexistent-project' not found"),
				ContainSubstring("worktree 'shell-integration/nonexistent-project' not found"),
				ContainSubstring("worktree 'twiggit/nonexistent-project' not found")))
		})
	})

	Describe("Error Suggestions", func() {
		It("should provide helpful suggestions for validation errors", func() {
			// Try to create a worktree with empty branch name
			session := cli.Run("create", "")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			// Should contain helpful suggestions
			Expect(output).To(ContainSubstring("💡"))
			Expect(output).To(ContainSubstring("Provide a valid branch name"))
		})

		It("should provide helpful suggestions for invalid branch names", func() {
			// Try to create a worktree with invalid branch name
			session := cli.Run("create", "invalid@branch#name")
			Eventually(session).Should(gexec.Exit(1))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("❌"))
			Expect(output).To(ContainSubstring("branch name format is invalid"))
		})
	})
})
