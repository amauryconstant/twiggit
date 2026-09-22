package iostreams_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"twiggit/internal/iostreams"
)

func TestStyles_IdentityWhenColorDisabled(t *testing.T) {
	t.Parallel()

	s := iostreams.NewStyles(false)

	in := "msg"
	assert.Equal(t, in, s.Error(in))
	assert.Equal(t, in, s.Success(in))
	assert.Equal(t, in, s.Hint(in))
	assert.Equal(t, in, s.Header(in))
	assert.Equal(t, in, s.Dim(in))
}

func TestStyles_RendersWhenColorEnabled(t *testing.T) {
	t.Parallel()

	s := iostreams.NewStyles(true)
	in := "boom"

	for name, fn := range map[string]func(string) string{
		"Error":   s.Error,
		"Success": s.Success,
		"Hint":    s.Hint,
		"Header":  s.Header,
		"Dim":     s.Dim,
	} {
		fn := fn
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := fn(in)
			assert.NotEmpty(t, got, "styled output must not be empty")
			if got == in {
				// Either identity (which we just said is not the case
				// here) or terminal that drops styles entirely; lipgloss
				// always emits an escape sequence in color mode, so the
				// raw passthrough should be impossible. Belt + braces.
				t.Fatalf("expected lipgloss-rendered output for %s, got identity", name)
			}
			// All lipgloss-colored output begins with an ANSI escape.
			assert.True(t, strings.Contains(got, "\x1b["), "%s should contain ANSI escape", name)
		})
	}
}

func TestStyles_DimProducesDistinguishablyStyledOutput(t *testing.T) {
	t.Parallel()

	s := iostreams.NewStyles(true)
	in := "v"

	dim := s.Dim(in)
	header := s.Header(in)
	assert.NotEqual(t, dim, header, "different foregrounds must yield different sequences")
}
