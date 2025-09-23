//go:build e2e
// +build e2e

package e2e

import (
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
})
