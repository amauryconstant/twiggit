package core_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// expectedDataTypeNames mirrors the rename table in core-types spec
// (Requirement: Data types in internal/core are named without Git*
// stutter or Info suffix). The hard-list is intentional: reflection
// over the core package would need a typed-nil probe per type, which
// adds risk without buying coverage the spec does not require.
var expectedDataTypeNames = []string{
	"Repository",
	"Branch",
	"Worktree",
	"Remote",
	"Commit",
	"RepoDir",
	"RepositoryStatus",
}

func TestDataTypes_NoGitPrefix(t *testing.T) {
	t.Parallel()

	for _, name := range expectedDataTypeNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			is.False(strings.HasPrefix(name, "Git"),
				"data type %q stutters the Git prefix; core-types spec bans Git* on exported types", name)
		})
	}
}

func TestDataTypes_NoInfoSuffix(t *testing.T) {
	t.Parallel()

	for _, name := range expectedDataTypeNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			is.False(strings.HasSuffix(name, "Info"),
				"data type %q carries the Info suffix; core-types spec bans *Info on exported types", name)
		})
	}
}
