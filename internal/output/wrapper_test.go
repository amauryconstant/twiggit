package output_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"twiggit/internal/core"
	"twiggit/internal/output"
)

func TestComposeWrapper_EmptyTemplateUsesDefaultPerShell(t *testing.T) {
	t.Parallel()

	cases := []struct {
		shell core.ShellType
		token string
	}{
		{core.ShellBash, "bash"},
		{core.ShellZsh, "zsh"},
		{core.ShellFish, "fish"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.shell), func(t *testing.T) {
			t.Parallel()
			got := output.ComposeWrapper("", tc.shell)
			assert.Contains(t, got, tc.token, "default template must mention its shell name")
			// Defaults must reference the twiggit binary so the eval'd
			// wrapper actually runs something on cd.
			assert.Contains(t, got, "twiggit cd")
		})
	}
}

func TestComposeWrapper_SubstitutesShellNamePlaceholder(t *testing.T) {
	t.Parallel()

	template := "# twiggit wrapper for ${SHELL_NAME}\necho ${SHELL_NAME}\n"
	got := output.ComposeWrapper(template, core.ShellZsh)

	assert.Contains(t, got, "zsh")
	assert.NotContains(t, got, "${SHELL_NAME}")
	assert.NotContains(t, got, "${SHELL_TYPE}")
}

func TestComposeWrapper_SubstitutesShellTypePlaceholder(t *testing.T) {
	t.Parallel()

	template := "export SHELL_TYPE=${SHELL_TYPE}\nexport SHELL_NAME=${SHELL_NAME}\n"
	got := output.ComposeWrapper(template, core.ShellBash)

	assert.Contains(t, got, "export SHELL_TYPE=bash")
	assert.Contains(t, got, "export SHELL_NAME=bash")
}

func TestComposeWrapper_UnknownShellFallsBackToBashDefault(t *testing.T) {
	t.Parallel()

	got := output.ComposeWrapper("", core.ShellType("powershell"))
	// Bash default begins with the bash comment line.
	require.NotEmpty(t, got)
	assert.True(t, strings.HasPrefix(got, "# twiggit shell integration for bash"),
		"unknown shell should fall back to bash default, got: %q", got[:min(80, len(got))])
}

func TestComposeWrapper_NonEmptyTemplatePreservedApartFromSubstitutions(t *testing.T) {
	t.Parallel()

	template := "# keep me\necho hello\n# ${SHELL_NAME}\n"
	got := output.ComposeWrapper(template, core.ShellFish)

	assert.Contains(t, got, "# keep me")
	assert.Contains(t, got, "echo hello")
	assert.Contains(t, got, "fish")
	assert.NotContains(t, got, "${SHELL_NAME}")
}
