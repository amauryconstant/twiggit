//go:build e2e
// +build e2e

package helpers

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

type TestIDGenerator struct {
	testName  string
	randomID  string
	shortName string
}

func NewTestIDGenerator() *TestIDGenerator {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomID := fmt.Sprintf("%04x", rng.Uint32())

	testName := GinkgoT().Name()

	shortName := strings.ToLower(testName)
	shortName = strings.ReplaceAll(shortName, " ", "-")
	shortName = strings.ReplaceAll(shortName, "_", "-")
	shortName = strings.ReplaceAll(shortName, "/", "-")
	if len(shortName) > 20 {
		shortName = shortName[:20]
	}

	return &TestIDGenerator{
		testName:  testName,
		randomID:  randomID,
		shortName: shortName,
	}
}

func (g *TestIDGenerator) String() string {
	return fmt.Sprintf("%s-%s", g.shortName, g.randomID)
}

func (g *TestIDGenerator) ProjectName() string {
	return g.String()
}

func (g *TestIDGenerator) ProjectNameWithSuffix(suffix string) string {
	return fmt.Sprintf("%s-%s", g.String(), suffix)
}

func (g *TestIDGenerator) BranchName(branch string) string {
	return fmt.Sprintf("%s-%s", branch, g.randomID)
}

func (g *TestIDGenerator) WorktreeName(branch string) string {
	return g.BranchName(branch)
}
