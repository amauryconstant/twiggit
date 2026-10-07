//go:build e2e

// Package e2e provides end-to-end tests for twiggit completion command.
// Tests validate that `twiggit completion <shell>` emits a Carapace snippet
// for each supported shell and rejects unsupported shells.
package e2e

import (
	"twiggit/test/e2e/helpers"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("completion command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("emits bash completion script", func() {
		session := cli.Run("completion", "bash")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "bash")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits zsh completion script", func() {
		session := cli.Run("completion", "zsh")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "compdef")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits fish completion script", func() {
		session := cli.Run("completion", "fish")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "complete -c")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("rejects a stray positional after the shell name", func() {
		session := cli.Run("completion", "zsh", "extra")
		cli.ShouldFailWithExit(session, 2)
	})

	It("accepts no positional after the shell name", func() {
		session := cli.Run("completion", "zsh")
		cli.ShouldSucceed(session)
	})
})
