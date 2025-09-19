// Package version provides version information for the twiggit CLI
package version

import (
	"os"
	"regexp"
	"strings"
)

// Version returns the current version from go.mod
func Version() string {
	// Read go.mod file
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "dev"
	}

	// Extract version from comment using regex
	re := regexp.MustCompile(`// Version: ([^\s]+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return "dev"
	}

	return strings.TrimSpace(matches[1])
}
