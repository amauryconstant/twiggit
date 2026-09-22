package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellWrapper(t *testing.T) {
	tests := []struct {
		name      string
		shellType ShellType
		wantErr   bool
		contains  []string
	}{
		{
			name:      "bash wrapper contains twiggit function definition",
			shellType: ShellBash,
			contains: []string{
				"### BEGIN TWIGGIT WRAPPER",
				"### END TWIGGIT WRAPPER",
				"twiggit() {",
				"esac",
				"### BEGIN TWIGGIT COMPLETION",
				"_carapace bash",
			},
		},
		{
			name:      "zsh wrapper contains twiggit function definition",
			shellType: ShellZsh,
			contains: []string{
				"### BEGIN TWIGGIT WRAPPER",
				"### END TWIGGIT WRAPPER",
				"twiggit() {",
				"esac",
				"### BEGIN TWIGGIT COMPLETION",
				"_carapace zsh",
			},
		},
		{
			name:      "fish wrapper contains fish function syntax",
			shellType: ShellFish,
			contains: []string{
				"### BEGIN TWIGGIT WRAPPER",
				"### END TWIGGIT WRAPPER",
				"function twiggit",
				"switch \"$argv[1]\"",
				"### BEGIN TWIGGIT COMPLETION",
				"_carapace fish | source",
			},
		},
		{
			name:      "unsupported shell type returns error",
			shellType: ShellType("powershell"),
			wantErr:   true,
		},
		{
			name:      "empty shell type returns error",
			shellType: ShellType(""),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			must := require.New(t)

			got, err := ShellWrapper(tt.shellType)
			if tt.wantErr {
				must.Error(err)
				is.Empty(got)
				return
			}

			must.NoError(err)
			must.NotEmpty(got)
			for _, want := range tt.contains {
				is.Contains(got, want)
			}

			is.NotContains(got, "{{SHELL_TYPE}}", "placeholders must be replaced")
			is.NotContains(got, "{{TIMESTAMP}}", "placeholders must be replaced")
		})
	}
}

func TestShellWrapper_PlaceholdersSubstituted(t *testing.T) {
	is := assert.New(t)

	got, err := ShellWrapper(ShellBash)
	require.NoError(t, err)
	is.Contains(got, "# Twiggit bash wrapper - Generated on")
}

func TestShellWrapper_FishSyntaxDiffers(t *testing.T) {
	is := assert.New(t)

	bash, err := ShellWrapper(ShellBash)
	require.NoError(t, err)
	fish, err := ShellWrapper(ShellFish)
	require.NoError(t, err)

	is.Contains(bash, "esac")
	is.NotContains(bash, "switch \"$argv[1]\"")
	is.Contains(fish, "switch \"$argv[1]\"")
	is.NotContains(fish, "esac")
}
