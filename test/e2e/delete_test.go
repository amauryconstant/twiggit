//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Delete Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for delete command", func() {
		session := cli.Run("delete", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit delete"))
		Expect(output).To(ContainSubstring("Delete Git worktrees"))
	})

	It("shows safety features in help", func() {
		session := cli.Run("delete", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Safety features:"))
		Expect(output).To(ContainSubstring("Interactive confirmation"))
		Expect(output).To(ContainSubstring("Protection of main repositories"))
		Expect(output).To(ContainSubstring("Protection of current worktree"))
	})

	It("shows available flags", func() {
		session := cli.Run("delete", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("--dry-run"))
		Expect(output).To(ContainSubstring("--force"))
		Expect(output).To(ContainSubstring("--verbose"))
		Expect(output).To(ContainSubstring("Show what would be deleted"))
		Expect(output).To(ContainSubstring("Skip interactive confirmation"))
	})

	It("shows examples in help", func() {
		session := cli.Run("delete", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Examples:"))
		Expect(output).To(ContainSubstring("twiggit delete"))
		Expect(output).To(ContainSubstring("twiggit delete --dry-run"))
		Expect(output).To(ContainSubstring("twiggit delete --force"))
	})

	It("accepts no arguments", func() {
		session := cli.Run("delete", "extra-arg")
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(ContainSubstring("No worktrees found to delete"))
	})

	It("supports all flags", func() {
		session := cli.Run("delete", "--help", "--dry-run", "--force", "--verbose")
		Eventually(session).Should(gexec.Exit(0))
	})
})
