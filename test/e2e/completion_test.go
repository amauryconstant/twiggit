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

	It("emits powershell completion script", func() {
		session := cli.Run("completion", "powershell")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "using namespace System.Management.Automation")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits elvish completion script", func() {
		session := cli.Run("completion", "elvish")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "set edit:completion:arg-completer")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits nushell completion script", func() {
		session := cli.Run("completion", "nushell")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "let twiggit_completer")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits oil completion script", func() {
		session := cli.Run("completion", "oil")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "#!/bin/osh")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits tcsh completion script", func() {
		session := cli.Run("completion", "tcsh")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, `complete "twiggit"`)

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits xonsh completion script", func() {
		session := cli.Run("completion", "xonsh")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "from xonsh.completers.completer import add_one_completer")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("emits cmd-clink completion script", func() {
		session := cli.Run("completion", "cmd-clink")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "match_builder:setnosort()")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("rejects unsupported shell with ValidationError", func() {
		session := cli.Run("completion", "ksh")
		cli.ShouldFailWithExit(session, 1)
		cli.ShouldErrorOutput(session, `unsupported shell "ksh"`)
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
