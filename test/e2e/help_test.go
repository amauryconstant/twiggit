//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit help command.
// Tests validate help display for main command and subcommands.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"

	"twiggit/test/e2e/helpers"
)

var _ = Describe("help command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		configHelper := helpers.NewConfigHelper()
		cli = helpers.NewTwiggitCLI().WithConfigDir(configHelper.Build())
	})

	It("shows main help", func() {
		session := cli.Run("help")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "twiggit")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows help for list command", func() {
		session := cli.Run("help", "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "list")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows help for create command", func() {
		session := cli.Run("help", "create")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "create")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows help for delete command", func() {
		session := cli.Run("help", "delete")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "delete")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows help for cd command", func() {
		session := cli.Run("help", "cd")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "cd")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows ls alias in list command help", func() {
		session := cli.Run("help", "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "ls")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows rm alias in delete command help", func() {
		session := cli.Run("help", "delete")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "rm")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows -a short flag in list command help", func() {
		session := cli.Run("help", "list")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "-a")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("shows -y short flag in prune command help", func() {
		session := cli.Run("help", "prune")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "-y")
		cli.ShouldOutput(session, "--yes")

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})
})
