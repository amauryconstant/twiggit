// Path: internal/iostreams/styles.go
package iostreams

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles are defined once and passed around; with color off every style is
// the identity, so piped output stays free of ANSI codes.
type Styles struct {
	Error   func(...string) string
	Success func(...string) string
	Warning func(...string) string
	Dim     func(...string) string
	Bold    func(...string) string
}

func NewStyles(color bool) *Styles {
	if !color {
		plain := func(s ...string) string { return strings.Join(s, " ") }
		return &Styles{plain, plain, plain, plain, plain}
	}
	return &Styles{
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render,
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render,
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render,
		Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render,
		Bold:    lipgloss.NewStyle().Bold(true).Render,
	}
}
