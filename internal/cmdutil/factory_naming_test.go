package cmdutil_test

import (
	"reflect"
	"testing"
	"twiggit/internal/cmdutil"

	"github.com/stretchr/testify/assert"
)

// bannedPerRoleFields enumerates the role-interface accessor fields
// removed by cli-factory spec (Requirement: Factory exposes lazy
// function fields — Scenario: Factory does not expose per-role fields).
// Adding a field with one of these names back to cmdutil.Factory
// re-opens the door the spec closed; this test enforces the removal.
var bannedPerRoleFields = []string{
	"RepoOpener",
	"BranchReader",
	"RepositoryReader",
	"RemoteReader",
	"WorktreeWriter",
	"BranchWriter",
}

func TestFactory_NoPerRoleFields(t *testing.T) {
	t.Parallel()

	factoryType := reflect.TypeFor[cmdutil.Factory]()

	for _, name := range bannedPerRoleFields {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			_, ok := factoryType.FieldByName(name)
			is.False(ok,
				"cmdutil.Factory has banned field %q; cli-factory spec requires role-narrowed access via type assertion on f.GitClient(), not per-role fields", name)
		})
	}
}
