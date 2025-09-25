//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/amaury/twiggit/test/helpers"
)

var _ = Describe("Setup-Shell Command", func() {
	var cli *helpers.TwiggitCLI

	BeforeEach(func() {
		cli = helpers.NewTwiggitCLI()
	})

	It("shows help for setup-shell command", func() {
		session := cli.Run("setup-shell", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit setup-shell"))
		Expect(output).To(ContainSubstring("Setup shell integration for twiggit"))
	})

	It("shows examples in help", func() {
		session := cli.Run("setup-shell", "--help")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Examples:"))
		Expect(output).To(ContainSubstring("twiggit setup-shell"))
		Expect(output).To(ContainSubstring("--shell bash"))
		Expect(output).To(ContainSubstring("--shell zsh"))
		Expect(output).To(ContainSubstring("--shell fish"))
	})

	It("accepts no arguments", func() {
		session := cli.Run("setup-shell")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Shell Integration Setup"))
		Expect(output).To(ContainSubstring("Detected shell:"))
	})

	It("supports --shell flag with bash", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Detected shell: bash"))
		Expect(output).To(ContainSubstring("bash"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for bash"))
	})

	It("supports --shell flag with zsh", func() {
		session := cli.Run("setup-shell", "--shell", "zsh")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Detected shell: zsh"))
		Expect(output).To(ContainSubstring("zsh"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for zsh"))
	})

	It("supports --shell flag with fish", func() {
		session := cli.Run("setup-shell", "--shell", "fish")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Detected shell: fish"))
		Expect(output).To(ContainSubstring("fish"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for fish"))
	})

	It("rejects invalid shell type", func() {
		session := cli.Run("setup-shell", "--shell", "invalid")
		Eventually(session).Should(gexec.Exit(1))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("❌"))
		Expect(output).To(ContainSubstring("unsupported shell type"))
		Expect(output).To(ContainSubstring("bash, zsh, fish"))
	})

	It("shows configuration files for bash", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring(".bashrc"))
		Expect(output).To(ContainSubstring(".bash_profile"))
		Expect(output).To(ContainSubstring(".profile"))
	})

	It("shows configuration files for zsh", func() {
		session := cli.Run("setup-shell", "--shell", "zsh")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring(".zshrc"))
		Expect(output).To(ContainSubstring(".zprofile"))
	})

	It("shows configuration files for fish", func() {
		session := cli.Run("setup-shell", "--shell", "fish")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("config.fish"))
	})

	It("shows wrapper function for bash", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit()"))
		Expect(output).To(ContainSubstring("builtin cd"))
		Expect(output).To(ContainSubstring("command twiggit"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for bash"))
	})

	It("shows wrapper function for zsh", func() {
		session := cli.Run("setup-shell", "--shell", "zsh")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("twiggit()"))
		Expect(output).To(ContainSubstring("builtin cd"))
		Expect(output).To(ContainSubstring("command twiggit"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for zsh"))
	})

	It("shows wrapper function for fish", func() {
		session := cli.Run("setup-shell", "--shell", "fish")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("function twiggit"))
		Expect(output).To(ContainSubstring("builtin cd"))
		Expect(output).To(ContainSubstring("command twiggit"))
		Expect(output).To(ContainSubstring("twiggit shell integration wrapper for fish"))
	})

	It("shows setup instructions", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Setup Instructions"))
		Expect(output).To(ContainSubstring("Add the wrapper function"))
		Expect(output).To(ContainSubstring("Reload your shell"))
		Expect(output).To(ContainSubstring("source ~/.bashrc"))
	})

	It("shows verification steps", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Verification"))
		Expect(output).To(ContainSubstring("twiggit cd --help"))
	})

	It("shows important notes", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Important Notes"))
		Expect(output).To(ContainSubstring("overrides the shell's built-in"))
		Expect(output).To(ContainSubstring("builtin cd"))
	})

	It("shows warning about shell built-in override", func() {
		session := cli.Run("setup-shell", "--shell", "bash")
		Eventually(session).Should(gexec.Exit(0))

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Warning"))
		Expect(output).To(ContainSubstring("overrides the shell's built-in"))
		Expect(output).To(ContainSubstring("builtin cd"))
	})
})
