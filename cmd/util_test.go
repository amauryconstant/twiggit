package cmd

import (
	"strings"
	"testing"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerbosef_NilIOSafe asserts the helper tolerates a nil *IOStreams
// for early-startup call sites (matches the contract documented at
// cmd/util.go:verbosef). Without the guard, run paths that capture
// ios from cobra before PersistentPreRunE fires would panic.
func TestVerbosef_NilIOSafe(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		verbosef(nil, "no-op")
	})
}

// TestVerbosef_QuietIOSilent asserts the helper is a no-op when
// ios.Verbose is false (mirrors the predicate inside
// iostreams.IOStreams.Verbosef).
func TestVerbosef_QuietIOSilent(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = false

	verbosef(ios, "must not appear")

	assert.Empty(t, errOut.String())
}

// TestVerbosef_EmitsDimStderrOneLine asserts the helper, when given a
// Verbose=true ios, writes exactly one dim-styled line to ios.ErrOut
// terminated by a newline, per cli-verbose-output "Verbosef emits one
// line on -v" and "Stderr only".
func TestVerbosef_EmitsDimStderrOneLine(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	verbosef(ios, "Creating worktree for %s/%s", "demo", "feature-x")

	got := errOut.String()
	assert.Equal(t, "Creating worktree for demo/feature-x\n", got)
}

// TestVerbosef_OmitsDebugPrefixAndIndents asserts the rendered line
// carries no "DEBUG:" / "[VERBOSE]" prefix and no leading whitespace,
// per cli-verbose-output "No debug prefix" and the level-collapse
// design (the previous level-2 indentation rule is removed).
func TestVerbosef_OmitsDebugPrefixAndIndents(t *testing.T) {
	t.Parallel()

	ios, _, _, errOut := iostreams.Test()
	ios.Verbose = true

	verbosef(ios, "branch: %s", "feature-y")

	rendered := errOut.String()
	assert.NotContains(t, rendered, "DEBUG:")
	assert.NotContains(t, rendered, "[VERBOSE]")
	assert.False(t, strings.HasPrefix(rendered, "  "), "level-2 indentation must be gone")
	assert.Contains(t, rendered, "branch: feature-y")
}
