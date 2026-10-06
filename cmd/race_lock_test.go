package cmd

import (
	"sync"
	"twiggit/internal/cmdutil"

	"github.com/spf13/cobra"
)

// rootBuildMu serialises NewRootCommand invocations across the test
// binary. The cobra + carapace pair reads and writes to a global
// OnInitialize hook slice during carapace.Gen(cmd); without this
// lock the race detector flags every parallel test that constructs a
// root command, even when the same hooks are re-registered
// idempotently. The lock is a test-only artefact so the production
// binary is unaffected.
var rootBuildMu sync.Mutex

// RootBuildLock acquires the package mutex for callers that need to
// hold it across multiple NewRootCommand invocations.
func RootBuildLock() { rootBuildMu.Lock() }

// RootBuildUnlock releases the package mutex.
func RootBuildUnlock() { rootBuildMu.Unlock() }

var _ = func(f *cmdutil.Factory, mutate ...func(*cobra.Command)) *cobra.Command {
	rootBuildMu.Lock()
	defer rootBuildMu.Unlock()
	return newRootForTest(f, mutate...)
}
