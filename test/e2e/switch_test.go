//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Switch Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for switch command", func() {
		session := cli.Run("switch", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit switch <project|project/branch>"))
		Expect(output).To(ContainSubstring("Switch to a project repository or worktree"))
	})

	It("shows examples in help", func() {
		session := cli.Run("switch", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Examples:"))
		Expect(output).To(ContainSubstring("twiggit switch myproject"))
		Expect(output).To(ContainSubstring("twiggit switch myproject/feature-branch"))
		Expect(output).To(ContainSubstring("twiggit switch feature-branch"))
	})

	It("accepts at most one argument", func() {
		session := cli.Run("switch", "project1", "project2")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Err.Contents())).To(ContainSubstring("accepts at most 1 arg(s), received 2"))
	})

	It("shows context help when no arguments provided", func() {
		session := cli.RunWithDir("/tmp", "switch")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Err.Contents())).To(ContainSubstring("specify a target"))
	})

	It("handles project switching format", func() {
		session := cli.Run("switch", "--help")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("twiggit switch <project|project/branch>"))
	})

	It("handles worktree switching format", func() {
		session := cli.Run("switch", "--help")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("project/branch"))
	})
})
