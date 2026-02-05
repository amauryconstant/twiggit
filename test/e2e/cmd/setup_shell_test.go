//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit setup-shell command.
// Tests validate shell wrapper installation for bash, zsh, and fish.
package cmde2e

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("setup-shell command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		configDir := fixture.Build()
		cli = cli.WithConfigDir(configDir)
		_ = helpers.NewTwiggitAssertions()
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	It("sets up bash shell with dry-run flag", func() {
		session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		Expect(output).To(ContainSubstring("Wrapper function:"))
		Expect(output).To(ContainSubstring("twiggit()"))
	})

	It("sets up zsh shell with dry-run flag", func() {
		session := cli.Run("setup-shell", "--shell=zsh", "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for zsh"))
		Expect(output).To(ContainSubstring("Wrapper function:"))
		Expect(output).To(ContainSubstring("twiggit()"))
	})

	It("sets up fish shell with dry-run flag", func() {
		session := cli.Run("setup-shell", "--shell=fish", "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for fish"))
		Expect(output).To(ContainSubstring("Wrapper function:"))
		Expect(output).To(ContainSubstring("function twiggit"))
	})

	It("forces reinstall with --force flag", func() {
		tempDir, err := os.MkdirTemp("", "twiggit-e2e-home-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		cli := cli.WithEnvironment("HOME", tempDir)

		session := cli.Run("setup-shell", "--shell=bash", "--force")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))
		Expect(output).To(ContainSubstring("To activate the wrapper"))
		Expect(output).To(ContainSubstring("twiggit cd <branch>"))
		Expect(output).To(ContainSubstring("builtin cd <path>"))
	})

	It("shows dry-run output without making changes", func() {
		session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		Expect(output).To(ContainSubstring("Wrapper function:"))
		Expect(output).NotTo(ContainSubstring("Shell wrapper installed"))
		Expect(output).NotTo(ContainSubstring("To activate the wrapper"))
	})

	It("skips installation when already installed", func() {
		tempDir, err := os.MkdirTemp("", "twiggit-e2e-home-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(tempDir)

		cli := cli.WithEnvironment("HOME", tempDir)

		session := cli.Run("setup-shell", "--shell=bash", "--force")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		session2 := cli.Run("setup-shell", "--shell=bash")
		cli.ShouldSucceed(session2)

		if session2.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session2.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper already installed for bash"))
		Expect(output).To(ContainSubstring("Use --force to reinstall"))
	})

	It("errors with invalid shell type", func() {
		session := cli.Run("setup-shell", "--shell=invalid")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "unsupported shell type: invalid")
		cli.ShouldErrorOutput(session, "(supported: bash, zsh, fish)")
	})
})
