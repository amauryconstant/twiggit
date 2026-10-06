package git

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRebaser_BaseTracker_RoutesAreCli pins the git-rebaser spec
// requirement that every Rebaser and BaseTracker method on the
// composite *Client resolves through *cliClient (CLI-backed), not
// *reader (go-git-backed). Method-set membership is the proof: the
// composite embeds both halves, so a method must appear in the
// *cliClient method set (and must NOT appear in the *reader method
// set) for the routing to be correct. A future contributor who
// moves one of these methods onto *reader — even as a stub — would
// make *reader satisfy Rebaser/BaseTracker and trip this test.
//
// go-git lacks support for rebase, abort, continue, fetch, and
// per-worktree config read/write, so all six methods MUST route to
// the CLI adapter.
func TestRebaser_BaseTracker_RoutesAreCli(t *testing.T) {
	t.Parallel()

	clientType := reflect.TypeFor[*Client]()
	cliType := reflect.TypeFor[*cliClient]()
	readerType := reflect.TypeFor[*reader]()

	methods := []string{
		"Rebase",
		"Abort",
		"Continue",
		"Fetch",
		"SetTrackedBase",
		"GetTrackedBase",
	}

	for _, name := range methods {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, clientHas := clientType.MethodByName(name)
			require.True(t, clientHas, "*Client must expose %s (via embedded *cliClient)", name)

			_, cliHas := cliType.MethodByName(name)
			require.True(t, cliHas, "*cliClient must define %s", name)

			_, readerHas := readerType.MethodByName(name)
			assert.False(t, readerHas,
				"*reader must not define %s (go-git lacks support)", name)
		})
	}
}
