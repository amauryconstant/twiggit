//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Global CLI Behavior", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("handles unknown commands gracefully", func() {
		session := cli.Run("unknown-command")
		Eventually(session).Should(gexec.Exit(1))
		Expect(string(session.Err.Contents())).To(ContainSubstring("unknown command \"unknown-command\" for \"twiggit\""))
	})

	It("supports global verbose flag", func() {
		session := cli.Run("--verbose", "--help")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit is a fast and intuitive tool"))
		Expect(output).To(ContainSubstring("--verbose"))
	})

	It("supports global quiet flag", func() {
		session := cli.Run("--quiet", "--help")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit is a fast and intuitive tool"))
		Expect(output).To(ContainSubstring("--quiet"))
	})

	It("shows consistent error format", func() {
		session := cli.Run("invalid-command")
		Eventually(session).Should(gexec.Exit(1))
		output := string(session.Err.Contents())
		Expect(output).To(ContainSubstring("Error:"))
		Expect(output).To(ContainSubstring("unknown command"))
	})

	It("handles missing arguments gracefully", func() {
		session := cli.Run()
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit is a fast and intuitive tool"))
		Expect(output).To(ContainSubstring("Usage:"))
	})

	It("shows command completion info", func() {
		session := cli.Run("__complete", "", "")
		Eventually(session).Should(gexec.Exit(0))
		output := string(session.Err.Contents())
		Expect(output).To(ContainSubstring("ShellCompDirective"))
	})
})
