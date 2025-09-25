//go:build e2e
// +build e2e

package e2e

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Cd Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for cd command", func() {
		session := cli.Run("cd", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit cd <project|project/branch>"))
		Expect(output).To(ContainSubstring("Change directory to a project repository or worktree"))
	})

	It("shows examples in help", func() {
		session := cli.Run("cd", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Examples:"))
		Expect(output).To(ContainSubstring("twiggit cd myproject"))
		Expect(output).To(ContainSubstring("twiggit cd myproject/feature-branch"))
		Expect(output).To(ContainSubstring("twiggit cd feature-branch"))
	})

	It("accepts at most one argument", func() {
		session := cli.Run("cd", "project1", "project2")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Out.Contents())).To(ContainSubstring("❌"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("accepts at most 1 arg(s), received 2"))
	})

	It("shows context help when no arguments provided", func() {
		tempDir, err := os.MkdirTemp("", "twiggit-cd-test")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		session := cli.RunWithDir(tempDir, "cd")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("Current context: Outside git repository"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("Available targets:"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("twiggit cd <project>"))
	})

	It("handles project switching format", func() {
		session := cli.Run("cd", "--help")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("twiggit cd <project|project/branch>"))
	})

	It("handles worktree switching format", func() {
		session := cli.Run("cd", "--help")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("project/branch"))
	})
})
