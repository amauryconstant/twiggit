//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/helpers"
)

var _ = Describe("Help Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help with --help flag", func() {
		session := cli.Run("--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit"))
		Expect(output).To(ContainSubstring("twiggit is a fast and intuitive tool for managing Git worktrees and projects"))
		Expect(output).To(ContainSubstring("Available Commands:"))
	})

	It("shows help with -h flag", func() {
		session := cli.Run("-h")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("Usage"))
	})

	It("shows version with --version", func() {
		session := cli.Run("--version")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("twiggit version"))
	})

	It("lists all available commands", func() {
		session := cli.Run("--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("cd"))
		Expect(output).To(ContainSubstring("list"))
		Expect(output).To(ContainSubstring("create"))
		Expect(output).To(ContainSubstring("delete"))
		Expect(output).To(ContainSubstring("setup-shell"))
	})

	It("shows command descriptions", func() {
		session := cli.Run("--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Change directory to a project or worktree"))
		Expect(output).To(ContainSubstring("List all available worktrees"))
		Expect(output).To(ContainSubstring("Create a new Git worktree"))
		Expect(output).To(ContainSubstring("Delete Git worktrees"))
		Expect(output).To(ContainSubstring("Setup shell integration"))
	})
})
