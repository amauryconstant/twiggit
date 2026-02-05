//go:build e2e
// +build e2e

package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// TwiggitCLI provides CLI execution utilities for E2E tests
type TwiggitCLI struct {
	binaryPath string
	env        map[string]string
}

// NewTwiggitCLI creates a new TwiggitCLI instance
func NewTwiggitCLI() *TwiggitCLI {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	if strings.HasSuffix(cwd, "test/e2e") {
		cwd = filepath.Dir(filepath.Dir(cwd))
	} else if strings.HasSuffix(cwd, "test/e2e/cmd") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	} else if strings.HasSuffix(cwd, "test/e2e/context") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	} else if strings.HasSuffix(cwd, "test/e2e/workflows") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	}

	binaryPath := filepath.Join(cwd, "bin", "twiggit-e2e")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		GinkgoT().Logf("Binary not found at %s, attempting to build...", binaryPath)
		BuildBinary()
	}

	_, err = os.Stat(binaryPath)
	Expect(err).NotTo(HaveOccurred(),
		"Twiggit binary not found at %s. Run 'mise run build:e2e' first", binaryPath)

	return &TwiggitCLI{
		binaryPath: binaryPath,
		env:        make(map[string]string),
	}
}

// WithEnvironment sets an environment variable for CLI execution
func (cli *TwiggitCLI) WithEnvironment(key, value string) *TwiggitCLI {
	cli.env[key] = value
	return cli
}

// WithConfigDir sets the XDG_CONFIG_HOME to point to a specific config directory
func (cli *TwiggitCLI) WithConfigDir(configDir string) *TwiggitCLI {
	return cli.WithEnvironment("XDG_CONFIG_HOME", configDir)
}

// WithProjectsDir sets the projects directory environment variable
func (cli *TwiggitCLI) WithProjectsDir(projectsDir string) *TwiggitCLI {
	return cli.WithEnvironment("TWIGGIT_PROJECTS_DIR", projectsDir)
}

// WithWorktreesDir sets the worktrees directory environment variable
func (cli *TwiggitCLI) WithWorktreesDir(worktreesDir string) *TwiggitCLI {
	return cli.WithEnvironment("TWIGGIT_WORKTREES_DIR", worktreesDir)
}

// Run executes the twiggit CLI with the given arguments
func (cli *TwiggitCLI) Run(args ...string) *gexec.Session {
	command := exec.Command(cli.binaryPath, args...)

	// Prepare environment
	env := os.Environ()
	for key, value := range cli.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	command.Env = env

	// Create and start the session
	session, err := gexec.Start(command, nil, nil)
	Expect(err).NotTo(HaveOccurred())

	return session
}

// RunWithDir executes the twiggit CLI from a specific directory
func (cli *TwiggitCLI) RunWithDir(dir string, args ...string) *gexec.Session {
	command := exec.Command(cli.binaryPath, args...)
	command.Dir = dir

	// Prepare environment
	env := os.Environ()
	for key, value := range cli.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	command.Env = env

	// Create and start the session
	session, err := gexec.Start(command, nil, nil)
	Expect(err).NotTo(HaveOccurred())

	return session
}

// GetOutput returns the stdout output from a session as a string
func (cli *TwiggitCLI) GetOutput(session *gexec.Session) string {
	return strings.TrimSpace(string(session.Out.Contents()))
}

// GetError returns the stderr output from a session as a string
func (cli *TwiggitCLI) GetError(session *gexec.Session) string {
	return strings.TrimSpace(string(session.Err.Contents()))
}

// Reset clears all environment variables
func (cli *TwiggitCLI) Reset() *TwiggitCLI {
	cli.env = make(map[string]string)
	return cli
}

// BuildBinary builds the twiggit binary for E2E tests
func BuildBinary() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	if strings.HasSuffix(cwd, "test/e2e") {
		cwd = filepath.Dir(filepath.Dir(cwd))
	} else if strings.HasSuffix(cwd, "test/e2e/cmd") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	} else if strings.HasSuffix(cwd, "test/e2e/context") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	} else if strings.HasSuffix(cwd, "test/e2e/workflows") {
		cwd = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	}

	cmd := exec.Command("go", "build", "-tags=e2e", "-o", filepath.Join(cwd, "bin", "twiggit-e2e"), "main.go")
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "Failed to build twiggit binary: %s", string(output))
}

// ShouldSucceed asserts the command succeeds with exit code 0
func (cli *TwiggitCLI) ShouldSucceed(session *gexec.Session) {
	Eventually(session).Should(gexec.Exit(0))
}

// ShouldFailWithExit asserts the command fails with specific exit code
func (cli *TwiggitCLI) ShouldFailWithExit(session *gexec.Session, exitCode int) {
	Eventually(session).Should(gexec.Exit(exitCode))
}

// ShouldOutput asserts stdout contains expected text
func (cli *TwiggitCLI) ShouldOutput(session *gexec.Session, expected string) {
	Eventually(session.Out).Should(gbytes.Say(expected))
}

// ShouldErrorOutput asserts stderr contains expected text
func (cli *TwiggitCLI) ShouldErrorOutput(session *gexec.Session, expected string) {
	Eventually(session.Err).Should(gbytes.Say(expected))
}

// ShouldOutputMatch asserts stdout matches regex pattern
func (cli *TwiggitCLI) ShouldOutputMatch(session *gexec.Session, pattern string) {
	Eventually(session.Out).Should(gbytes.Say(pattern))
}

// ShouldErrorOutputMatch asserts stderr matches regex pattern
func (cli *TwiggitCLI) ShouldErrorOutputMatch(session *gexec.Session, pattern string) {
	Eventually(session.Err).Should(gbytes.Say(pattern))
}

// DebugSession logs session details and fixture state for debugging
// Logs if session.ExitCode() != 0, otherwise does nothing
func (cli *TwiggitCLI) DebugSession(session *gexec.Session, fixtureInfo string) {
	if session.ExitCode() != 0 {
		GinkgoT().Log(fixtureInfo)
		GinkgoT().Log("Output:", string(session.Out.Contents()))
		GinkgoT().Log("Error:", string(session.Err.Contents()))
	}
}
