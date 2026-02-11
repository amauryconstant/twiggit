//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit init command.
// Tests validate shell wrapper installation with inference, force, and config file scenarios.
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

	It("installs to existing config file with inferred shell type", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
		Expect(output).To(ContainSubstring("To activate the wrapper"))

		content, err := os.ReadFile(bashrcPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
		Expect(string(content)).To(ContainSubstring("### END TWIGGIT WRAPPER"))
	})

	It("installs to missing config file with inferred shell type", func() {
		configPath := filepath.Join(fixture.GetTempDir(), ".zshrc")

		session := cli.Run("init", configPath)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for zsh"))
		Expect(output).To(ContainSubstring("Config file: " + configPath))

		_, err := os.Stat(configPath)
		Expect(err).NotTo(HaveOccurred())

		content, err := os.ReadFile(configPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
		Expect(string(content)).To(ContainSubstring("### END TWIGGIT WRAPPER"))
	})

	It("infers shell type from .bashrc filename", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath, "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
	})

	It("infers shell type from .zshrc filename", func() {
		configDir := filepath.Join(fixture.GetTempDir(), ".config", "fish")
		Expect(os.MkdirAll(configDir, 0755)).To(Succeed())
		fishConfigPath := filepath.Join(configDir, "config.fish")
		Expect(os.WriteFile(fishConfigPath, []byte("# Fish config\n"), 0644)).To(Succeed())

		session := cli.Run("init", fishConfigPath, "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for fish"))
		Expect(output).To(ContainSubstring("Config file: " + fishConfigPath))
	})

	It("infers shell type from config.fish filename", func() {
		configDir := filepath.Join(fixture.GetTempDir(), ".config", "fish")
		Expect(os.MkdirAll(configDir, 0755)).To(Succeed())
		fishConfigPath := filepath.Join(configDir, "config.fish")
		Expect(os.WriteFile(fishConfigPath, []byte("# Fish config\n"), 0644)).To(Succeed())

		session := cli.Run("init", fishConfigPath, "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for fish"))
		Expect(output).To(ContainSubstring("Config file: " + fishConfigPath))
	})

	It("forces reinstall with block replacement", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session1 := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session1)

		if session1.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output1 := string(session1.Out.Contents())
		Expect(output1).To(ContainSubstring("Shell wrapper installed for bash"))

		session2 := cli.Run("init", bashrcPath, "--force")
		cli.ShouldSucceed(session2)

		if session2.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output2 := string(session2.Out.Contents())
		Expect(output2).To(ContainSubstring("Shell wrapper installed for bash"))

		content, err := os.ReadFile(bashrcPath)
		Expect(err).NotTo(HaveOccurred())

		beginCount := 0
		endCount := 0
		for _, line := range strings.Split(string(content), "\n") {
			if strings.Contains(line, "### BEGIN TWIGGIT WRAPPER") {
				beginCount++
			}
			if strings.Contains(line, "### END TWIGGIT WRAPPER") {
				endCount++
			}
		}
		Expect(beginCount).To(Equal(1))
		Expect(endCount).To(Equal(1))
	})

	It("skips when wrapper exists without force flag", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session1 := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session1)

		if session1.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		session2 := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session2)

		if session2.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session2.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper already installed for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
		Expect(output).To(ContainSubstring("Use --force to reinstall"))
	})

	It("shows dry-run output without making changes", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath, "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
		Expect(output).To(ContainSubstring("Wrapper function:"))
		Expect(output).To(ContainSubstring("### BEGIN TWIGGIT WRAPPER"))
		Expect(output).To(ContainSubstring("### END TWIGGIT WRAPPER"))
		Expect(output).NotTo(ContainSubstring("Shell wrapper installed"))
		Expect(output).NotTo(ContainSubstring("To activate the wrapper"))
	})

	It("errors with inference failure for custom path", func() {
		customConfigPath := filepath.Join(fixture.GetTempDir(), "config.txt")

		session := cli.Run("init", customConfigPath, "--dry-run")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "cannot infer shell type from path")
		cli.ShouldErrorOutput(session, customConfigPath)
		cli.ShouldErrorOutput(session, "use --shell to specify shell type")
	})

	It("accepts explicit --shell override", func() {
		customConfigPath := filepath.Join(fixture.GetTempDir(), "my-config")
		Expect(os.WriteFile(customConfigPath, []byte("# Custom config\n"), 0644)).To(Succeed())

		session := cli.Run("init", customConfigPath, "--shell=bash", "--dry-run")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		Expect(output).To(ContainSubstring("Config file: " + customConfigPath))
	})

	It("checks if wrapper is installed with --check flag", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session1 := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session1)

		session2 := cli.Run("init", bashrcPath, "--check")
		cli.ShouldSucceed(session2)

		if session2.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session2.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper is installed"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
	})

	It("shows not installed with --check flag when missing", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath, "--check")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper not installed"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
	})

	It("errors with invalid shell type", func() {
		configPath := filepath.Join(fixture.GetTempDir(), ".bashrc")

		session := cli.Run("init", configPath, "--shell=invalid")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		cli.ShouldErrorOutput(session, "unsupported shell type: invalid")
	})

	It("auto-detects shell and config file when no arguments provided", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for bash"))
		Expect(output).To(ContainSubstring("Config file: " + bashrcPath))
	})

	It("uses explicit --shell flag over auto-detection", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/bash")
		zshrcPath := filepath.Join(fixture.GetTempDir(), ".zshrc")
		Expect(os.WriteFile(zshrcPath, []byte("# Zsh config\n"), 0644)).To(Succeed())

		session := cli.Run("init", "--shell=zsh")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell wrapper installed for zsh"))
		Expect(output).To(ContainSubstring("Config file: " + zshrcPath))
	})

	It("errors when auto-detection fails without explicit shell", func() {
		cli = cli.WithEnvironment("SHELL", "/bin/sh")

		session := cli.Run("init")
		cli.ShouldFailWithExit(session, 1)

		if session.ExitCode() != 1 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Err.Contents())
		Expect(output).To(ContainSubstring("shell auto-detection failed"))
		Expect(output).To(ContainSubstring("unsupported shell detected"))
	})

	It("shows level 1 verbose output with -v flag", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath, "-v")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Setting up shell wrapper")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows level 2 verbose output with -vv flag", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath, "-vv")
		cli.ShouldSucceed(session)
		cli.ShouldVerboseOutput(session, "Setting up shell wrapper")
		cli.ShouldVerboseOutput(session, "  shell type: bash")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("shows no verbose output by default", func() {
		bashrcPath := filepath.Join(fixture.GetTempDir(), ".bashrc")
		Expect(os.WriteFile(bashrcPath, []byte("# Bash config\n"), 0644)).To(Succeed())

		session := cli.Run("init", bashrcPath)
		cli.ShouldSucceed(session)
		cli.ShouldNotHaveVerboseOutput(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})
})
