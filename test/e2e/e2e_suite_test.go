//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTwiggitCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Twiggit CLI Suite")
}
