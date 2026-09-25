// Path: internal/output/errors.go
package output

import (
	"errors"
	"fmt"

	"github.com/you/myapp/internal/core"
	"github.com/you/myapp/internal/iostreams"
)

// FormatError writes err to stderr once: the full message (every wrap layer and
// every joined error), then the next steps carried by typed errors in the
// chain. main.go is the only caller (single-handling rule).
func FormatError(ios *iostreams.IOStreams, err error) {
	s := ios.ErrStyles()
	fmt.Fprintf(ios.ErrOut, "%s %v\n", s.Error("error:"), err)
	for _, hint := range suggestions(err) {
		fmt.Fprintf(ios.ErrOut, "  %s\n", s.Dim(hint))
	}
}

func suggestions(err error) []string {
	var (
		hints   []string
		invalid *core.ValidationError
		op      *core.OperationError
	)
	if errors.As(err, &invalid) {
		hints = append(hints, invalid.Suggestions...)
	}
	if errors.As(err, &op) {
		hints = append(hints, op.Suggestions...)
	}
	return hints
}
