//go:build e2e
// +build e2e

package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

type TestIDGenerator struct {
	timestamp    string
	counter      int32
	shortName    string
	branchCache  map[string]string
	projectCache map[string]string
}

func NewTestIDGenerator() *TestIDGenerator {
	name := GinkgoT().Name()

	shortName := strings.ToLower(name)
	shortName = strings.ReplaceAll(shortName, " ", "-")
	shortName = strings.ReplaceAll(shortName, "_", "-")
	shortName = strings.ReplaceAll(shortName, "/", "-")
	if len(shortName) > 20 {
		shortName = shortName[:20]
	}

	return &TestIDGenerator{
		timestamp:    time.Now().Format("20060102-150405"),
		counter:      1,
		shortName:    shortName,
		branchCache:  make(map[string]string),
		projectCache: make(map[string]string),
	}
}

func (g *TestIDGenerator) generateRandomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", g.timestamp, atomic.AddInt32(&g.counter, 1))
	}
	return hex.EncodeToString(b)
}

func (g *TestIDGenerator) String() string {
	return fmt.Sprintf("%s-%s", g.shortName, g.generateRandomID())
}

func (g *TestIDGenerator) ProjectName() string {
	return g.String()
}

func (g *TestIDGenerator) ProjectNameWithSuffix(suffix string) string {
	if cached, ok := g.projectCache[suffix]; ok {
		return cached
	}
	result := fmt.Sprintf("%s-%s-%s", suffix, g.shortName, g.generateRandomID())
	g.projectCache[suffix] = result
	return result
}

func (g *TestIDGenerator) BranchName(branch string) string {
	// Use cache to ensure same branch name is returned for repeated calls
	if cached, ok := g.branchCache[branch]; ok {
		return cached
	}
	result := fmt.Sprintf("%s-%s", branch, g.generateRandomID())
	g.branchCache[branch] = result
	return result
}

func (g *TestIDGenerator) WorktreeName(branch string) string {
	return g.BranchName(branch)
}
