//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestTwiggitCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Twiggit CLI Suite")
}

var _ = BeforeSuite(func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	if strings.HasSuffix(cwd, "test/e2e") {
		cwd = filepath.Dir(filepath.Dir(cwd))
	}

	cmd := exec.Command("go", "build", "-tags=e2e", "-o", filepath.Join(cwd, "bin", "twiggit-e2e"), "main.go")
	cmd.Dir = cwd

	session, err := gexec.Start(cmd, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
