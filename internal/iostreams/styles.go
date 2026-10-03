package iostreams

import (
	"charm.land/lipgloss/v2"
)

// Styles groups the styled-render functions used by the cmd layer.
// When isColorEnabled is false every method is the identity function:
// strings pass through unchanged. When true each method renders the
// input through a lipgloss Style with the conventional 16-color
// terminal palette:
//
//   - Error:   color 9  (bright red)
//   - Success: color 10 (bright green)
//   - Hint:    color 11 (bright yellow)
//   - Header:  color 12 (bright blue)
//   - Dim:     color 8  (bright gray)
//
// lipgloss v2 detects color profile from the process environment
// via colorprofile.Detect, so non-TTY contexts (CI, pipes) emit
// plain text without ANSI codes regardless of the isColorEnabled
// flag on the IOStreams struct.
type Styles struct {
	Error   func(string) string
	Success func(string) string
	Hint    func(string) string
	Header  func(string) string
	Dim     func(string) string
}

// NewStyles builds a Styles set according to isColorEnabled. When
// false, every field is the identity function (string round-trip).
// When true, each field wraps a lipgloss style; the renderer is
// driven by lipgloss's global color profile detection.
func NewStyles(isColorEnabled bool) *Styles {
	if !isColorEnabled {
		return &Styles{
			Error:   identity,
			Success: identity,
			Hint:    identity,
			Header:  identity,
			Dim:     identity,
		}
	}
	return &Styles{
		Error: func(s string) string { return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render(s) },
		Success: func(s string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render(s)
		},
		Hint: func(s string) string { return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(s) },
		Header: func(s string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render(s)
		},
		Dim: func(s string) string { return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(s) },
	}
}

func identity(s string) string { return s }
