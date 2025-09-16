package main

import (
	"testing"

	"github.com/carapace-sh/carapace"
)

// TestCarapace verifies carapace configuration during build time
func TestCarapace(t *testing.T) {
	carapace.Test(t)
}
