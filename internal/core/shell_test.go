package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidShellType_TableDriven pins the contract of the pure
// derivation half of the shell-detect split (cli-functional-core-shell
// task 3.9 / infrastructure-shell-detect MODIFIED): `core.ShellType`
// + `core.IsValidShellType` live here, with the filesystem probing
// half moved to `internal/git/shell_detect.go`. The three documented
// shells are valid; everything else (including empty and unknown) is
// rejected.
func TestIsValidShellType_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		shellType ShellType
		want      bool
	}{
		{"bash valid", ShellBash, true},
		{"zsh valid", ShellZsh, true},
		{"fish valid", ShellFish, true},
		{"powershell invalid", ShellType("powershell"), false},
		{"sh invalid", ShellType("sh"), false},
		{"empty invalid", ShellType(""), false},
		{"uppercase rejected", ShellType("BASH"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			is.Equal(tt.want, IsValidShellType(tt.shellType))
		})
	}
}

// TestShellType_StringIdentity pins the wire format: shell_type values
// are exactly the strings `bash`, `zsh`, `fish`. Any drift here would
// break the env-driven detection in `internal/git/shell_detect.go` and
// the `cli-output` shell-wrapper rendering.
func TestShellType_StringIdentity(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bash", string(ShellBash))
	assert.Equal(t, "zsh", string(ShellZsh))
	assert.Equal(t, "fish", string(ShellFish))
}
