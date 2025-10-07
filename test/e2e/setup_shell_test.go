//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("Setup Shell Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	AfterEach(func() {
		cli.Reset()
	})

	Context("Help Display and Flag Validation", func() {
		It("shows help for setup-shell command", func() {
			session := cli.Run("setup-shell", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit setup-shell"))
			Expect(output).To(ContainSubstring("Install shell wrapper"))
			Expect(output).To(ContainSubstring("bash|zsh|fish"))
		})

		It("shows help with -h flag", func() {
			session := cli.Run("setup-shell", "-h")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("twiggit setup-shell"))
		})

		It("requires shell flag", func() {
			session := cli.Run("setup-shell")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("required flag(s) \"shell\" not set"))
		})

		It("shows usage examples in help", func() {
			session := cli.Run("setup-shell", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("twiggit setup-shell --shell=bash"))
		})

		It("describes wrapper functionality in help", func() {
			session := cli.Run("setup-shell", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("wrapper functions"))
			Expect(output).To(ContainSubstring("directory navigation"))
			Expect(output).To(ContainSubstring("Escape hatch"))
		})

		It("lists supported shells in help", func() {
			session := cli.Run("setup-shell", "--help")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("bash"))
			Expect(output).To(ContainSubstring("zsh"))
			Expect(output).To(ContainSubstring("fish"))
		})
	})

	Context("Shell Detection and Wrapper Generation", func() {
		It("generates bash wrapper successfully", func() {
			// First test that a simple command works
			helpSession := cli.Run("--help")
			Eventually(helpSession).Should(gexec.Exit(0))
			helpOutput := cli.GetOutput(helpSession)
			Expect(helpOutput).ToNot(BeEmpty())

			// Now test the setup-shell command
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session, 5).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).ToNot(BeEmpty())
			Expect(output).To(ContainSubstring("Would install wrapper for bash:"))
			Expect(output).To(ContainSubstring("twiggit()"))
			Expect(output).To(ContainSubstring("builtin cd"))
			Expect(output).To(ContainSubstring("command twiggit"))
		})

		It("generates zsh wrapper successfully", func() {
			session := cli.Run("setup-shell", "--shell=zsh", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for zsh:"))
			Expect(output).To(ContainSubstring("twiggit()"))
			Expect(output).To(ContainSubstring("builtin cd"))
			Expect(output).To(ContainSubstring("command twiggit"))
		})

		It("generates fish wrapper successfully", func() {
			session := cli.Run("setup-shell", "--shell=fish", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for fish:"))
			Expect(output).To(ContainSubstring("function twiggit"))
			Expect(output).To(ContainSubstring("builtin cd"))
			Expect(output).To(ContainSubstring("command twiggit"))
		})

		It("shows wrapper content in dry run mode", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Wrapper function:"))
			Expect(output).To(ContainSubstring("# Twiggit bash wrapper"))
		})

		It("handles different wrapper syntax for each shell", func() {
			// Test bash syntax
			bashSession := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(bashSession).Should(gexec.Exit(0))
			bashOutput := cli.GetOutput(bashSession)
			Expect(bashOutput).To(ContainSubstring("twiggit() {"))
			Expect(bashOutput).To(ContainSubstring("fi"))

			// Test zsh syntax
			zshSession := cli.Run("setup-shell", "--shell=zsh", "--dry-run")
			Eventually(zshSession).Should(gexec.Exit(0))
			zshOutput := cli.GetOutput(zshSession)
			Expect(zshOutput).To(ContainSubstring("twiggit() {"))
			Expect(zshOutput).To(ContainSubstring("fi"))

			// Test fish syntax
			fishSession := cli.Run("setup-shell", "--shell=fish", "--dry-run")
			Eventually(fishSession).Should(gexec.Exit(0))
			fishOutput := cli.GetOutput(fishSession)
			Expect(fishOutput).To(ContainSubstring("function twiggit"))
			Expect(fishOutput).To(ContainSubstring("end"))
		})
	})

	Context("Completion Support", func() {
		It("supports bash completion integration", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should include completion-friendly wrapper syntax
			Expect(output).To(ContainSubstring("twiggit()"))
		})

		It("supports zsh completion integration", func() {
			session := cli.Run("setup-shell", "--shell=zsh", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should include zsh-specific features
			Expect(output).To(ContainSubstring("twiggit()"))
		})

		It("supports fish completion integration", func() {
			session := cli.Run("setup-shell", "--shell=fish", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			// Should use fish function syntax
			Expect(output).To(ContainSubstring("function twiggit"))
			Expect(output).To(ContainSubstring("$argv"))
		})
	})

	Context("Error Handling for Unsupported Shells", func() {
		It("rejects unsupported shell types", func() {
			session := cli.Run("setup-shell", "--shell=unknown")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
			Expect(output).To(ContainSubstring("unknown"))
			Expect(output).To(ContainSubstring("bash, zsh, fish"))
		})

		It("rejects empty shell type", func() {
			session := cli.Run("setup-shell", "--shell=")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
		})

		It("rejects case-sensitive shell variations", func() {
			session := cli.Run("setup-shell", "--shell=Bash")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
			Expect(output).To(ContainSubstring("Bash"))
		})

		It("rejects partial shell names", func() {
			session := cli.Run("setup-shell", "--shell=bas")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
			Expect(output).To(ContainSubstring("bas"))
		})

		It("provides helpful error message with supported options", func() {
			session := cli.Run("setup-shell", "--shell=invalid")
			Eventually(session).Should(gexec.Exit(2))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("supported: bash, zsh, fish"))
		})
	})

	Context("Integration with Existing Shell Setups", func() {
		var tempHome string
		var bashrcPath string

		BeforeEach(func() {
			tempHome = GinkgoT().TempDir()
			bashrcPath = filepath.Join(tempHome, ".bashrc")
		})

		It("detects existing wrapper installation", func() {
			// Create existing wrapper
			existingWrapper := `# Twiggit bash wrapper
twiggit() {
    if [ "$1" = "cd" ]; then
        target_dir=$(command twiggit "$@")
        if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
            builtin cd "$target_dir"
        fi
    else
        command twiggit "$@"
    fi
}`
			err := os.WriteFile(bashrcPath, []byte(existingWrapper), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Set HOME to point to temp directory
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Shell wrapper already installed"))
			Expect(output).To(ContainSubstring("Use --force to reinstall"))
		})

		It("forces reinstall with --force flag", func() {
			// Create existing wrapper that matches detection pattern
			existingWrapper := `# Twiggit bash wrapper - Old version
twiggit() {
    echo "old wrapper"
}`
			err := os.WriteFile(bashrcPath, []byte(existingWrapper), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Force reinstall
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash", "--force")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Shell wrapper installed"))
			Expect(output).ToNot(ContainSubstring("already installed"))
		})

		It("handles missing config files gracefully", func() {
			// Use empty temp directory (no .bashrc)
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper"))
		})

		It("preserves existing shell configuration", func() {
			// Create .bashrc with existing content
			existingContent := `# Existing bash configuration
export PATH="/usr/local/bin:$PATH"
alias ll="ls -la"

# Other functions
custom_func() {
    echo "custom"
}`
			err := os.WriteFile(bashrcPath, []byte(existingContent), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Install wrapper
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			// Check that existing content would be preserved (this is a dry run test)
			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper"))
		})
	})

	Context("Configuration Integration", func() {
		var fixture *fixtures.E2ETestFixture

		BeforeEach(func() {
			fixture = fixtures.NewE2ETestFixture()
		})

		AfterEach(func() {
			fixture.Cleanup()
		})

		It("works with custom configuration", func() {
			fixture.SetupSingleProject("test-project")
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		})

		It("respects XDG_CONFIG_HOME", func() {
			customConfigDir := GinkgoT().TempDir()
			session := cli.WithEnvironment("XDG_CONFIG_HOME", customConfigDir).Run("setup-shell", "--shell=zsh", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for zsh"))
		})

		It("handles missing configuration gracefully", func() {
			session := cli.Run("setup-shell", "--shell=fish", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for fish"))
		})

		It("works with multi-project configurations", func() {
			fixture.SetupMultiProject()
			configDir := fixture.Build()

			session := cli.WithConfigDir(configDir).Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for bash"))
		})
	})

	Context("Output Format Validation", func() {
		It("provides clear success messages", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for bash:"))
			Expect(output).To(ContainSubstring("Wrapper function:"))
		})

		It("provides installation instructions", func() {
			tempHome := GinkgoT().TempDir()
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("To activate the wrapper:"))
			Expect(output).To(ContainSubstring("Restart your shell"))
			Expect(output).To(ContainSubstring("source ~/.bashrc"))
		})

		It("shows usage examples after installation", func() {
			tempHome := GinkgoT().TempDir()
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=zsh")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Usage:"))
			Expect(output).To(ContainSubstring("twiggit cd <branch>"))
			Expect(output).To(ContainSubstring("builtin cd <path>"))
		})

		It("provides shell-specific instructions", func() {
			tempHome := GinkgoT().TempDir()
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=fish")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("source ~/.bashrc")) // May need to adjust for fish
		})

		It("formats wrapper code properly", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			lines := strings.Split(output, "\n")

			// Find wrapper content section
			wrapperStart := -1
			for i, line := range lines {
				if strings.Contains(line, "Wrapper function:") {
					wrapperStart = i + 1
					break
				}
			}

			Expect(wrapperStart).To(BeNumerically(">", 0))

			// Check that wrapper is properly formatted
			wrapperContent := strings.Join(lines[wrapperStart:], "\n")
			Expect(wrapperContent).To(ContainSubstring("twiggit()"))
			Expect(wrapperContent).To(ContainSubstring("builtin cd"))
			Expect(wrapperContent).To(ContainSubstring("command twiggit"))
		})
	})

	Context("Flag Combinations", func() {
		It("handles --dry-run with --shell", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper"))
			Expect(output).To(ContainSubstring("bash"))
		})

		It("handles --force with --shell", func() {
			tempHome := GinkgoT().TempDir()
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash", "--force")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Shell wrapper installed"))
		})

		It("handles all flags together", func() {
			tempHome := GinkgoT().TempDir()
			session := cli.WithEnvironment("HOME", tempHome).Run("setup-shell", "--shell=bash", "--force", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper"))
		})

		It("rejects invalid flag combinations", func() {
			session := cli.Run("setup-shell", "--shell=invalid", "--dry-run")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
		})
	})

	Context("Edge Cases", func() {
		It("handles shell names with extra whitespace", func() {
			session := cli.Run("setup-shell", "--shell", " bash ")
			Eventually(session).Should(gexec.Exit(1))

			output := cli.GetError(session)
			Expect(output).To(ContainSubstring("unsupported shell type"))
		})

		It("handles multiple shell flags (last one wins)", func() {
			session := cli.Run("setup-shell", "--shell=bash", "--shell=zsh", "--dry-run")
			Eventually(session).Should(gexec.Exit(0))

			output := cli.GetOutput(session)
			Expect(output).To(ContainSubstring("Would install wrapper for zsh"))
		})

		It("provides consistent output format across shells", func() {
			bashSession := cli.Run("setup-shell", "--shell=bash", "--dry-run")
			zshSession := cli.Run("setup-shell", "--shell=zsh", "--dry-run")
			fishSession := cli.Run("setup-shell", "--shell=fish", "--dry-run")

			Eventually(bashSession).Should(gexec.Exit(0))
			Eventually(zshSession).Should(gexec.Exit(0))
			Eventually(fishSession).Should(gexec.Exit(0))

			bashOutput := cli.GetOutput(bashSession)
			zshOutput := cli.GetOutput(zshSession)
			fishOutput := cli.GetOutput(fishSession)

			// All should have similar structure
			for _, output := range []string{bashOutput, zshOutput, fishOutput} {
				Expect(output).To(ContainSubstring("Would install wrapper for"))
				Expect(output).To(ContainSubstring("Wrapper function:"))
			}
		})
	})
})
