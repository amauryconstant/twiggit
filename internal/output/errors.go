package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"twiggit/internal/core"
	"twiggit/internal/iostreams"
)

// FormatError dispatches err to a per-type renderer that writes
// to w. Dispatch order matches spec 3.5: errors.As walks the chain
// to the first matching canonical type, so a wrapper that embeds a
// ValidationError in its Cause still renders via the wrapper's own
// Message + Op context.
//
//  1. core.OperationError  (runtime wrapper, dispatched first so it
//     catches shell/navigation/git wrappers before their inner
//     ValidationError gets a chance to match)
//  2. core.UsageError      (invocation-level usage failure)
//  3. core.NotFoundError   (resource missing)
//  4. core.ValidationError (leaf argument / input validation)
//
// Anything else falls through to a one-line "Error: <msg>"
// rendering. When TWIGGIT_DEBUG is set the full error chain is
// appended via fmt.Fprintf("%+v", err) so on-call debugging
// recovers the wrapped cause.
func FormatError(w io.Writer, err error, ios *iostreams.IOStreams) {
	if err == nil {
		return
	}
	switch {
	case errors.As(err, new(*core.OperationError)):
		var oe *core.OperationError
		_ = errors.As(err, &oe)
		formatOperationError(w, oe, ios)
	case errors.As(err, new(*core.UsageError)):
		var ue *core.UsageError
		_ = errors.As(err, &ue)
		formatUsageError(w, ue, ios)
	case errors.As(err, new(*core.NotFoundError)):
		var nf *core.NotFoundError
		_ = errors.As(err, &nf)
		formatNotFoundError(w, nf, ios)
	case errors.As(err, new(*core.ValidationError)):
		var ve *core.ValidationError
		_ = errors.As(err, &ve)
		formatValidationError(w, ve, ios)
	default:
		writeGenericError(w, err, ios)
	}
	if os.Getenv("TWIGGIT_DEBUG") != "" {
		_, _ = fmt.Fprintf(w, "%+v\n", err)
	}
}

func style(ios *iostreams.IOStreams) *iostreams.Styles {
	if ios == nil {
		return iostreams.NewStyles(false)
	}
	return ios.Styles()
}

func formatValidationError(w io.Writer, e *core.ValidationError, ios *iostreams.IOStreams) {
	st := style(ios)
	header := st.Error("Error:")
	msg := e.Message
	if msg == "" {
		msg = "validation failed"
	}
	_, _ = fmt.Fprintf(w, "%s %s\n", header, msg)
	if e.Op != "" || e.Entity != "" || e.Field != "" {
		var parts []string
		if e.Op != "" {
			parts = append(parts, "op="+e.Op)
		}
		if e.Entity != "" {
			parts = append(parts, "entity="+e.Entity)
		}
		if e.Field != "" {
			parts = append(parts, "field="+e.Field)
		}
		if e.Value != "" {
			parts = append(parts, "value="+e.Value)
		}
		if len(parts) > 0 {
			_, _ = fmt.Fprintf(w, "  %s\n", st.Hint(strings.Join(parts, " ")))
		}
	}
	writeSuggestions(w, e.Suggestions, st)
}

func formatNotFoundError(w io.Writer, e *core.NotFoundError, ios *iostreams.IOStreams) {
	st := style(ios)
	header := st.Error("Not found:")
	_, _ = fmt.Fprintf(w, "%s %s %s\n", header, e.Entity, e.Name)
}

func formatOperationError(w io.Writer, e *core.OperationError, ios *iostreams.IOStreams) {
	st := style(ios)
	header := st.Error("Error:")
	msg := e.Message
	if msg == "" {
		msg = "operation failed"
	}
	_, _ = fmt.Fprintf(w, "%s %s\n", header, msg)
	if e.Op != "" || e.Entity != "" || e.Field != "" {
		var parts []string
		if e.Op != "" {
			parts = append(parts, "op="+e.Op)
		}
		if e.Entity != "" {
			parts = append(parts, "entity="+e.Entity)
		}
		if e.Field != "" {
			parts = append(parts, "field="+e.Field)
		}
		if len(parts) > 0 {
			_, _ = fmt.Fprintf(w, "  %s\n", st.Hint(strings.Join(parts, " ")))
		}
	}
	if e.Cause != nil {
		_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("cause:"), e.Cause.Error())
	}
	writeSuggestions(w, e.Suggestions, st)
}

func formatUsageError(w io.Writer, e *core.UsageError, ios *iostreams.IOStreams) {
	st := style(ios)
	header := st.Error("Usage:")
	_, _ = fmt.Fprintf(w, "%s %s\n", header, e.Message)
}

func writeGenericError(w io.Writer, err error, ios *iostreams.IOStreams) {
	st := style(ios)
	_, _ = fmt.Fprintf(w, "%s %s\n", st.Error("Error:"), err.Error())
}

func writeSuggestions(w io.Writer, suggestions []string, st *iostreams.Styles) {
	if len(suggestions) == 0 {
		return
	}
	for _, s := range suggestions {
		_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("hint:"), s)
	}
}
