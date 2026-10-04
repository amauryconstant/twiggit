package core_test

import (
	"reflect"
	"strings"
	"testing"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
)

// expectedRoleInterfaces mirrors the role set declared in
// internal/git/client.go. Adding a role here without a corresponding
// reflection entry below would silently leave the new role unguarded.
var expectedRoleInterfaces = []reflect.Type{
	reflect.TypeFor[core.RepositoryOpener](),
	reflect.TypeFor[core.BranchReader](),
	reflect.TypeFor[core.RepositoryReader](),
	reflect.TypeFor[core.RemoteReader](),
	reflect.TypeFor[core.WorktreeWriter](),
	reflect.TypeFor[core.BranchWriter](),
}

func TestRoleInterfaces_NoGetPrefix(t *testing.T) {
	t.Parallel()

	for _, roleType := range expectedRoleInterfaces {
		t.Run(roleType.Name(), func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			for method := range roleType.Methods() {
				name := method.Name
				is.False(strings.HasPrefix(name, "Get"),
					"role %s method %q carries a Get prefix; core-git spec bans Get* on read methods", roleType.Name(), name)
			}
		})
	}
}
