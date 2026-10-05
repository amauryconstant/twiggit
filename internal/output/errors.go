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
// to w. Dispatch order matches `cli-error-formatting/spec.md` §
// "Type-matched dispatch": errors.As walks the chain to the first
// matching canonical type, with the most specific matchers
// registered first.
//
//  1. core.ValidationError  (leaf argument / input validation)
//  2. core.NotFoundError     (resource missing)
//  3. core.OperationError    (runtime wrapper for shell/navigation/git)
//  4. core.UsageError        (invocation-level usage failure)
//
// A wrapper that embeds a more-specific type in its Cause still
// renders via the more-specific branch (e.g. OperationError wrapping
// a ValidationError renders as ValidationError).
//
// Anything else falls through to a one-line "Error: <msg>"
// rendering. When TWIGGIT_DEBUG is set the full error chain is
// appended via fmt.Fprintf("%+v", err) so on-call debugging
// recovers the wrapped cause.
func FormatError(w io.Writer, err error, ios *iostreams.IOStreams) {
	if err == nil {
		return
	}
	if ve, ok := errors.AsType[*core.ValidationError](err); ok {
		formatValidationError(w, ve, ios, err)
	} else if nf, ok := errors.AsType[*core.NotFoundError](err); ok {
		formatNotFoundError(w, nf, ios, err)
	} else if oe, ok := errors.AsType[*core.OperationError](err); ok {
		formatOperationError(w, oe, ios, err)
	} else if ue, ok := errors.AsType[*core.UsageError](err); ok {
		formatUsageError(w, ue, ios)
	} else {
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

// shouldEmitHints is the single source for the quiet gate. Hint
// lines render only when ios is non-nil AND ios.IsQuiet is false.
func shouldEmitHints(ios *iostreams.IOStreams) bool {
	return ios != nil && !ios.IsQuiet
}

// writeFieldContext renders the op=/entity=/field=/value= context
// line shared by ValidationError and OperationError. The Op and Value
// fields are gated on ios.DebugEnabled() so internal context stays
// out of user-facing errors by default; Entity / Field are always
// rendered because they help diagnose the input without leaking
// internals. Empty fields are omitted from the rendered output; the
// whole line is skipped when every field is empty.
func writeFieldContext(w io.Writer, op, entity, field, value string, st *iostreams.Styles, debugEnabled bool) {
	var parts []string
	if debugEnabled && op != "" {
		parts = append(parts, "op="+op)
	}
	if entity != "" {
		parts = append(parts, "entity="+entity)
	}
	if field != "" {
		parts = append(parts, "field="+field)
	}
	if debugEnabled && value != "" {
		parts = append(parts, "value="+value)
	}
	if len(parts) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "  %s\n", st.Hint(strings.Join(parts, " ")))
}

// hintFor returns the resource-specific hint for a NotFound sentinel
// reachable via errors.Is from err, or an empty string if no
// registered sentinel matches. The returned string is one of the
// four constants declared in `cli-error-formatting/spec.md`
// § Actionable hints.
func hintFor(err error) string {
	if nf, ok := errors.AsType[*core.NotFoundError](err); ok {
		switch nf.Entity {
		case "project":
			return "Use 'twiggit list --all' to see available projects"
		case "worktree":
			return "Use 'twiggit list' to see available worktrees"
		case "resolution", "navigation", "navigation target", "resolution target":
			return "Use 'twiggit list' to see available navigation targets"
		case "git repository", "repository", "git repo":
			return "Verify the repository path"
		}
	}
	switch {
	case errors.Is(err, core.ErrProjectNotFound):
		return "Use 'twiggit list --all' to see available projects"
	case errors.Is(err, core.ErrWorktreeNotFound):
		return "Use 'twiggit list' to see available worktrees"
	case errors.Is(err, core.ErrResolutionNotFound):
		return "Use 'twiggit list' to see available navigation targets"
	case errors.Is(err, core.ErrGitRepoNotFound):
		return "Verify the repository path"
	}
	return ""
}

func formatValidationError(w io.Writer, e *core.ValidationError, ios *iostreams.IOStreams, chain error) {
	st := style(ios)
	header := st.Error("Error:")
	msg := e.Message
	if msg == "" {
		msg = "validation failed"
	}
	_, _ = fmt.Fprintf(w, "%s %s\n", header, msg)
	writeFieldContext(w, e.Op, e.Entity, e.Field, e.Value, st, debugEnabled(ios))
	if shouldEmitHints(ios) {
		writeSuggestions(w, e.Suggestions, st)
		if hint := hintFor(chain); hint != "" {
			_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("hint:"), hint)
		}
	}
}

func formatNotFoundError(w io.Writer, e *core.NotFoundError, ios *iostreams.IOStreams, chain error) {
	st := style(ios)
	header := st.Error("Not found:")
	_, _ = fmt.Fprintf(w, "%s %s %s\n", header, e.Entity, e.Name)
	if shouldEmitHints(ios) {
		if hint := hintFor(chain); hint != "" {
			_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("hint:"), hint)
		}
	}
}

func formatOperationError(w io.Writer, e *core.OperationError, ios *iostreams.IOStreams, chain error) {
	st := style(ios)
	header := st.Error("Error:")
	msg := e.Message
	if msg == "" {
		msg = "operation failed"
	}
	_, _ = fmt.Fprintf(w, "%s %s\n", header, msg)
	writeFieldContext(w, e.Op, e.Entity, e.Field, "", st, debugEnabled(ios))
	if e.Cause != nil {
		_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("cause:"), e.Cause.Error())
	}
	if shouldEmitHints(ios) {
		writeSuggestions(w, e.Suggestions, st)
		if hint := hintFor(chain); hint != "" {
			_, _ = fmt.Fprintf(w, "  %s %s\n", st.Hint("hint:"), hint)
		}
	}
}

// debugEnabled is the single source for the debug-mode gate used by
// the formatters. Nil-safe: returns false when ios is not wired so
// unit tests that bypass the IOStreams wiring still render the
// non-debug default.
func debugEnabled(ios *iostreams.IOStreams) bool {
	return ios != nil && ios.DebugEnabled()
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
