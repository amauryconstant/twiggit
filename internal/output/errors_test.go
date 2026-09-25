package output_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"
	"twiggit/internal/output"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatError_ValidationError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	ios, _, _, _ := iostreams.Test()
	err := core.NewValidationError("branch", "feature/x", "invalid characters")
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "Error:")
	assert.Contains(t, out, "invalid characters")
	assert.Contains(t, out, "field=branch")
	assert.NotContains(t, out, "%", "no fmt placeholders leaked: %s", out)
	_ = ctx
}

func TestFormatError_ValidationErrorWithSuggestions(t *testing.T) {
	t.Parallel()

	err := core.NewValidationError("branch", "feature/x", "invalid characters")
	err.Suggestions = []string{"try 'feature-x'", "use letters and dashes only"}

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "hint: try 'feature-x'")
	assert.Contains(t, out, "hint: use letters and dashes only")
}

func TestFormatError_NotFoundError(t *testing.T) {
	t.Parallel()

	err := &core.NotFoundError{Entity: "worktree", Name: "ghost-branch"}
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "Not found:")
	assert.Contains(t, out, "worktree")
	assert.Contains(t, out, "ghost-branch")
}

func TestFormatError_NotFoundErrorEmptyName(t *testing.T) {
	t.Parallel()

	err := &core.NotFoundError{Entity: "worktree"}
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "Not found:")
	assert.Contains(t, out, "worktree")
}

func TestFormatError_OperationError(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	err := &core.OperationError{
		Op:      "git.worktree",
		Entity:  "feature/x",
		Message: "add failed",
		Cause:   cause,
	}
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "Error:")
	assert.Contains(t, out, "add failed")
	assert.Contains(t, out, "op=git.worktree")
	assert.Contains(t, out, "entity=feature/x")
	assert.Contains(t, out, "boom")
	assert.Contains(t, out, "cause:")
}

func TestFormatError_OperationErrorSuggestionsAfterMessage(t *testing.T) {
	t.Parallel()

	err := &core.OperationError{
		Op:          "git.worktree",
		Entity:      "feature/x",
		Message:     "add failed",
		Suggestions: []string{"check your branch name", "rerun with --force"},
	}

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	parts := strings.SplitN(buf.String(), "\n", 2)
	assert.Contains(t, parts[0], "add failed")

	// Suggestions must appear after the message line.
	idxMsg := strings.Index(buf.String(), "add failed")
	idxHint := strings.Index(buf.String(), "hint:")
	assert.Greater(t, idxHint, idxMsg, "suggestions must render after the message")
}

func TestFormatError_UsageError(t *testing.T) {
	t.Parallel()

	err := core.NewUsageError("missing required flag --branch", nil)
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	assert.Contains(t, out, "Usage:")
	assert.Contains(t, out, "missing required flag --branch")
}

func TestFormatError_GenericError(t *testing.T) {
	t.Parallel()

	err := errors.New("plain old error")
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	assert.Contains(t, buf.String(), "Error:")
	assert.Contains(t, buf.String(), "plain old error")
}

func TestFormatError_NilIsNoOp(t *testing.T) {
	t.Parallel()

	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, nil, ios)

	assert.Empty(t, buf.String())
}

func TestFormatError_TWIGGIT_DEBUG_RevealsChain(t *testing.T) {
	t.Setenv("TWIGGIT_DEBUG", "1")

	cause := errors.New("root cause")
	wrapped := fmt.Errorf("outer: %w", cause)
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, wrapped, ios)

	out := buf.String()
	assert.Contains(t, out, "outer:")
	assert.Contains(t, out, "root cause")
}

func TestFormatError_NoDebugByDefault(t *testing.T) {
	t.Setenv("TWIGGIT_DEBUG", "")

	err := errors.New("just an error")
	ios, _, _, _ := iostreams.Test()
	var buf bytes.Buffer

	output.FormatError(&buf, err, ios)

	out := buf.String()
	// Generic error: only one "Error: ..." line; %+v would expand the
	// chain for non-trivial errors. For a bare error there is nothing
	// to expand, but the contract is: no chain dump unless TWIGGIT_DEBUG
	// is set. The output should remain single-line.
	assert.Equal(t, 1, strings.Count(out, "\n"))
	require.NotEmpty(t, out)
}
