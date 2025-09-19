//go:build e2e

package helpers

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type TwiggitCLI struct {
	binaryPath string
}

func NewTwiggitCLI() *TwiggitCLI {
	wd, _ := os.Getwd()
	return &TwiggitCLI{
		binaryPath: filepath.Join(wd, "..", "..", "bin", "twiggit"),
	}
}

func (c *TwiggitCLI) Run(args ...string) *gexec.Session {
	cmd := exec.Command(c.binaryPath, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func (c *TwiggitCLI) RunWithDir(dir string, args ...string) *gexec.Session {
	cmd := exec.Command(c.binaryPath, args...)
	cmd.Dir = dir
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
