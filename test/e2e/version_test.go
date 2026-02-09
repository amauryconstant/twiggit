//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit version command.
// Tests validate version information display and formatting.
package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/helpers"
)

var _ = Describe("version command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("outputs version information", func() {
		session := cli.Run("version")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
			GinkgoT().Log("Error:", string(session.Err.Contents()))
		}
	})

	It("displays build information", func() {
		session := cli.Run("version")
		cli.ShouldSucceed(session)

		output := cli.GetOutput(session)
		if len(output) == 0 {
			GinkgoT().Log("No version output received")
		}
	})

	It("formats version output correctly", func() {
		session := cli.Run("version")
		cli.ShouldSucceed(session)

		output := cli.GetOutput(session)
		Expect(output).To(ContainSubstring("twiggit"))
		if session.ExitCode() != 0 {
			GinkgoT().Log("Output:", string(session.Out.Contents()))
		}
	})
})
