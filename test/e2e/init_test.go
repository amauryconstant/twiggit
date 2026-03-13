//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit init command.
// Tests validate shell wrapper generation (stdout) and installation modes.
package e2e

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("init command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		configDir := fixture.Build()
		cli = cli.WithConfigDir(configDir)
		cli = cli.WithEnvironment("HOME", fixture.GetTempDir())
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			GinkgoT().Log(fixture.Inspect())
		}
		fixture.Cleanup()
	})

	// ========================================
	// Stdout Mode Tests (Default Behavior)
	// ========================================

	It("outputs wrapper to stdout by default", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")

		session := cli.Run("init")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("# Twiggit bash wrapper"))
		Expect(output).To(ContainSubstring("twiggit() {"))
		Expect(output).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
		Expect(output).To(ContainSubstring("### END TWIGGIT WRAPPER"))
		// Should NOT contain installation messages (stdout is clean for eval)
		Expect(output).NotTo(ContainSubstring("Shell wrapper installed"))
		Expect(output).NotTo(ContainSubstring("Config file:"))
	})

	It("outputs bash wrapper when shell specified", func() {
		session := cli.Run("init", "bash")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("# Twiggit bash wrapper"))
		Expect(output).To(ContainSubstring("twiggit() {"))
	})

	It("outputs zsh wrapper when shell specified", func() {
		session := cli.Run("init", "zsh")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("# Twiggit zsh wrapper"))
		Expect(output).To(ContainSubstring("twiggit() {"))
	})

	It("outputs fish wrapper when shell specified", func() {
		session := cli.Run("init", "fish")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("# Twiggit fish wrapper"))
		Expect(output).To(ContainSubstring("function twiggit"))
	})

	It("auto-detects shell from SHELL env when no arg provided", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/zsh")

		session := cli.Run("init")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("# Twiggit zsh wrapper"))
	})

	It("errors with invalid shell type in stdout mode", func() {
		session := cli.Run("init", "invalid")
		cli.ShouldFailWithExit(session, 5)

		cli.ShouldErrorOutput(session, "unsupported shell type")
	})

	It("errors when auto-detection fails in stdout mode", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/sh")

		session := cli.Run("init")
		cli.ShouldFailWithExit(session, 5)

		cli.ShouldErrorOutput(session, "shell auto-detection failed")
	})

	// ========================================
	// Install Mode Tests
	// ========================================

	It("installs to auto-detected config file with --install", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "--install")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
		Expect(output).To(ContainSubstring("To activate the wrapper"))

		content, err := os.ReadFile(bashrcPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
		Expect(string(content)).To(ContainSubstring("### END TWIGGIT WRAPPER"))
	})

	It("installs with explicit shell and --install", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		zshrcPath := filepath.Join(fixture.GetTempDir(), ".zshrc")
		Expect(os.WriteFile(zshrcPath, []byte("# Zsh config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "zsh", "--install")
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for zsh"))
		Expect(output).To(ContainSubstring("Config file: " + zshrcPath))
	})

	It("installs to custom config with --install --config", func() {
		customConfigPath := filepath.Join(fixture.GetTempDir(), "my-bash-config")
		Expect(os.WriteFile(customConfigPath, []byte("# Custom config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "bash", "--install", "--config", customConfigPath)
		cli.ShouldSucceed(session)

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))
		Expect(output).To(ContainSubstring("Config file: " + customConfigPath))

		content, err := os.ReadFile(customConfigPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
	})

	It("creates missing config file when installing", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/zsh")
		zshrcPath := filepath.Join(fixture.GetTempDir(), ".zshrc")

		session := cli.Run("init", "--install")
		cli.ShouldSucceed(session)

		_, err := os.Stat(zshrcPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("skips when wrapper already installed", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		// First install
		session1 := cli.Run("init", "--install")
		cli.ShouldSucceed(session1)

		// Second install (should skip)
		session2 := cli.Run("init", "--install")
		cli.ShouldSucceed(session2)

		output := string(session2.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper already installed for bash"))
		Expect(output).To(ContainSubstring("Use --force to reinstall"))
	})

	It("forces reinstall with --install --force", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		// First install
		session1 := cli.Run("init", "--install")
		cli.ShouldSucceed(session1)

		// Force reinstall
		session2 := cli.Run("init", "--install", "--force")
		cli.ShouldSucceed(session2)

		output := string(session2.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))

		// Verify only one wrapper block
		content, err := os.ReadFile(bashrcPath)
		Expect(err).NotTo(HaveOccurred())
		beginCount := strings.Count(string(content), "### BEGIN TWIGGIT WRAPPER")
		Expect(beginCount).To(Equal(1))
	})

	// ========================================
	// Flag Validation Tests
	// ========================================

	It("errors when --config used without --install", func() {
		session := cli.Run("init", "--config", "/custom/config")
		cli.ShouldFailWithExit(session, 2) // Usage error

		cli.ShouldErrorOutput(session, "--config requires --install")
	})

	It("errors when --force used without --install", func() {
		session := cli.Run("init", "--force")
		cli.ShouldFailWithExit(session, 2) // Usage error

		cli.ShouldErrorOutput(session, "--force requires --install")
	})

	It("errors with invalid shell type in install mode", func() {
		session := cli.Run("init", "invalid", "--install")
		cli.ShouldFailWithExit(session, 5)

		cli.ShouldErrorOutput(session, "unsupported shell type")
	})

	// ========================================
	// Verbose Output Tests
	// ========================================

	It("shows level 1 verbose output with -v flag in install mode", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "--install", "-v")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Setting up shell wrapper")
	})

	It("shows level 2 verbose output with -vv flag in install mode", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "--install", "-vv")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Setting up shell wrapper")
		cli.ShouldVerboseOutput(session, "  shell type: bash")
	})

	It("shows no verbose output by default in install mode", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "--install")
		cli.ShouldSucceed(session)
		cli.ShouldNotHaveVerboseOutput(session)
	})
})
