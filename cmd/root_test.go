package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoot_HelpGroupsSubcommands pins the cli-command-groups spec
// scenario `--help shows four groups`: `twiggit --help` must list
// the four group headers (Core:, Navigation:, Setup:, Meta:) with
// their respective subcommands under each. The test drives cobra
// directly via SetArgs + Execute so it does not shell-out to bash
// or depend on a built binary.
//
// Cannot use t.Parallel: NewRootCommand triggers
// carapace.Gen → cobra.OnInitialize which mutates cobra's
// package-level initializer slice. Parallel tests that build a
// root tree race on that global state.
func TestRoot_HelpGroupsSubcommands(t *testing.T) {
	f := newTestFactory(t)
	root := newRootForTest(f)
	stdout := new(bytes.Buffer)
	root.SetOut(stdout)
	root.SetErr(stdout)
	root.SetArgs([]string{"--help"})

	require.NoError(t, root.Execute())

	out := stdout.String()

	cases := []struct {
		group   string
		members []string
	}{
		{group: "Core:", members: []string{"list", "create", "delete", "prune"}},
		{group: "Navigation:", members: []string{"cd"}},
		{group: "Setup:", members: []string{"init"}},
		{group: "Meta:", members: []string{"version", "completion"}},
	}

	for _, tc := range cases {
		t.Run(strings.TrimSuffix(tc.group, ":"), func(t *testing.T) {
			assert.Contains(t, out, tc.group,
				"--help output must contain group header %q; got %q", tc.group, out)
			for _, m := range tc.members {
				assert.Contains(t, out, m,
					"group %q must contain member %q; got %q", tc.group, m, out)
			}
		})
	}
}
