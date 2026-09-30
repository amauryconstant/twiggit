package cmd

import (
	"testing"
	"twiggit/internal/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersion_NoOutputFlagRendersPlainText pins the cli-output-formats
// spec scenario `Non-list command without --output uses plain`: a
// non-list command without --output emits a plain human-readable
// success message on stdout — not a JSON array, not Tabular TSV, not
// an envelope object.
func TestVersion_NoOutputFlagRendersPlainText(t *testing.T) {
	t.Parallel()

	ios, _, stdout, _ := iostreams.Test()
	opts := &VersionOptions{
		IO:         ios,
		AppVersion: "1.27.1",
	}

	require.NoError(t, runVersion(opts))

	out := stdout.String()
	must := require.New(t)
	is := assert.New(t)

	must.NotEmpty(out, "version must emit something to stdout")

	is.NotContains(out, "[", "non-list default must not emit JSON array markers")
	is.NotContains(out, "]", "non-list default must not emit JSON array markers")
	is.NotRegexp(`^BRANCH\t`, out,
		"non-list default must not emit Tabular TSV header")
	is.NotRegexp(`\{".+":.+,"path":.+,"status":`, out,
		"non-list default must not emit bare-JSON worktree row")

	is.Regexp(`twiggit\s+1\.27\.1`, out,
		"non-list default renders plain human text mentioning the binary name and version")
}
