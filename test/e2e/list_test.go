//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("List Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for list command", func() {
		session := cli.Run("list", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit list"))
		Expect(output).To(ContainSubstring("List Git worktrees with intelligent auto-detection"))
	})

	It("shows available flags", func() {
		session := cli.Run("list", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("--all"))
		Expect(output).To(ContainSubstring("--sort"))
		Expect(output).To(ContainSubstring("Show worktrees from all projects"))
		Expect(output).To(ContainSubstring("Sort order"))
	})

	It("accepts no arguments", func() {
		session := cli.Run("list", "extra-arg")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("Worktrees in"))
	})

	It("supports --all flag", func() {
		session := cli.Run("list", "--help", "--all")
		Eventually(session).Should(gexec.Exit(0))
	})

	It("supports --sort flag with different values", func() {
		session := cli.Run("list", "--help", "--sort", "name")
		Eventually(session).Should(gexec.Exit(0))
	})

	It("shows examples in help", func() {
		session := cli.Run("list", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Examples:"))
		Expect(output).To(ContainSubstring("twiggit list"))
		Expect(output).To(ContainSubstring("twiggit list --all"))
		Expect(output).To(ContainSubstring("twiggit list --sort=date"))
	})
})
