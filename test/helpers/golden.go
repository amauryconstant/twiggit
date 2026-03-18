package helpers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CompareGolden compares actual output with golden file content.
//
// If UPDATE_GOLDEN environment variable is set to "true", the golden file
// will be updated with the actual output instead of performing a comparison.
//
// Golden files should be located in test/golden/<category>/<name>.golden
// and the goldenFile parameter should be the path relative to the golden directory.
//
// Example usage:
//
//	helpers.CompareGolden(t, "list/basic_output.golden", actualOutput)
func CompareGolden(t *testing.T, goldenFile string, actual string) {
	t.Helper()

	// Build full path to golden file
	goldenPath := filepath.Join("test", "golden", goldenFile)

	// Check if UPDATE_GOLDEN is set to true
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		updateGoldenFile(t, goldenPath, actual)
		return
	}

	// Read golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Normalize line endings for cross-platform comparison
	expectedStr := normalizeLineEndings(string(expected))
	actualStr := normalizeLineEndings(actual)

	// Compare actual with expected
	if expectedStr != actualStr {
		// Generate diff
		diff := generateDiff(expectedStr, actualStr)
		t.Errorf("output does not match golden file %s\n\n%s\n\nTo update golden file, run: UPDATE_GOLDEN=true go test ./...", goldenPath, diff)
	}
}

// updateGoldenFile writes actual content to the golden file.
func updateGoldenFile(t *testing.T, goldenPath string, actual string) {
	t.Helper()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
		t.Fatalf("failed to create directory for golden file %s: %v", goldenPath, err)
	}

	// Write golden file
	if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", goldenPath, err)
	}

	t.Logf("updated golden file: %s", goldenPath)
}

// normalizeLineEndings converts all line endings to LF for consistent comparison.
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// generateDiff creates a human-readable diff between expected and actual strings.
func generateDiff(expected, actual string) string {
	var buf bytes.Buffer

	buf.WriteString("--- Expected\n")
	buf.WriteString("+++ Actual\n")

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	maxLines := max(len(expectedLines), len(actualLines))

	for i := 0; i < maxLines; i++ {
		expectedLine := ""
		actualLine := ""

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine == actualLine {
			// Line matches - show as context
			if expectedLine != "" {
				fmt.Fprintf(&buf, "  %s\n", expectedLine)
			}
		} else {
			// Line differs
			if i < len(expectedLines) && expectedLine != "" {
				fmt.Fprintf(&buf, "- %s\n", expectedLine)
			}
			if i < len(actualLines) && actualLine != "" {
				fmt.Fprintf(&buf, "+ %s\n", actualLine)
			}
		}
	}

	return buf.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
