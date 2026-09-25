# Testing Strategy

The CLI-specific test pyramid. General Go idioms (named subtests, build tags, `t.Parallel()`, `goleak`, fuzzing) → See `samber/cc-skills-golang@golang-testing`; the `assert`/`require` API → See `samber/cc-skills-golang@golang-stretchr-testify`.

## Contents

- [The pyramid](#the-pyramid)
- [Core tests](#core-tests)
- [Command tests](#command-tests)
- [Golden files](#golden-files)
- [Integration tests](#integration-tests)
- [E2E tests](#e2e-tests)
- [TTY and interactive paths](#tty-and-interactive-paths)
- [Test doubles](#test-doubles)
- [Shared test helpers](#shared-test-helpers)
- [What to test per tier](#what-to-test-per-tier)

## The pyramid

```text
    ╱ E2E (few)          — the built binary: exit codes, streams, env vars
   ╱  Integration (some) — adapters against real tools, //go:build integration
  ╱   Command (many)     — runF injection and runX with fakes, captured output
 ╱    Core (most)        — table-driven, values in and out, no doubles
```

## Core tests

The core has no I/O, so its tests are table-driven functions over values: every validation rule and business rule gets its cases, and nothing needs a double.

```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {name: "valid", input: "feature/login"},
    {name: "empty", input: "", wantErr: true},
    {name: "reserved", input: "HEAD", wantErr: true},
}
```

## Command tests

Two levels, both in [create_test.go](../assets/examples/create_test.go):

- **Parse level** — `NewCmdCreate(f, runF)` with a `runF` that captures the options; `cmd.SetArgs(...)`, `cmd.Execute()`, then assert on the parsed fields. Nothing runs.
- **Run level** — call `runCreate` directly with an Options struct holding `iostreams.Test()` streams and a fake client; assert on stdout and on typed errors with `require.ErrorAs`.

## Golden files

A technique inside the command and E2E levels: capture stdout and stderr, compare to `testdata/<name>.golden`, refresh with `go test ./... -update`.

```go
ios, _, stdout, _ := iostreams.Test()
err := runList(&ListOptions{
    IO:     ios,
    Ctx:    context.Background(),
    Client: func() (worktreeLister, error) { return fakeLister{}, nil },
    Format: "json",
})
require.NoError(t, err)
golden.Assert(t, stdout.String(), "list-json.golden") // gotest.tools/v3/golden
```

## Integration tests

Adapters against the real tool or service, behind `//go:build integration`, with `t.TempDir()` for filesystem state. Helpers: `testcontainers-go` for services, `jarcoal/httpmock` for HTTP, `go-cmp` for diffs.

## E2E tests

Build the binary once (with `-cover` to collect coverage from the subprocess), run it with `os/exec`, and assert the exit code, stdout, and stderr:

```go
tests := []struct {
    name     string
    args     []string
    wantCode int
}{
    {name: "version", args: []string{"version"}, wantCode: 0},
    {name: "unknown command", args: []string{"bogus"}, wantCode: 2},
    {name: "bad flag", args: []string{"list", "--nope"}, wantCode: 2},
    {name: "runtime failure", args: []string{"get", "missing"}, wantCode: 1},
}
```

For many scenarios, `rogpeppe/go-internal/testscript` expresses each one as a `.txtar` script (`exec myapp bogus`, `! stdout .`, `stderr 'unknown command'`).

## TTY and interactive paths

- `iostreams.Test()` is non-TTY and colorless, so the scripted path is the default.
- `ios.SetTTY(true, true)` switches a test to the interactive path.
- Every prompting command gets both tests: flags supplied (no prompt) and missing input without a TTY (a `UsageError`). Bubble Tea programs take scripted input through `tea.WithInput`.

## Test doubles

| Interface | Double |
| --- | --- |
| Fewer than 5 methods, or needs state | Hand-written struct |
| 5+ methods | `matryer/moq` — generated function-field structs, no reflection |
| The core | None: test it with values |

## Shared test helpers

When the same setup repeats across 3+ command test files, move it to `internal/cmdutil/cmdtest`, imported by the tests only ([cmdtest.go](../assets/examples/cmdtest.go), with the fake):

```go
// Path: internal/cmdutil/cmdtest/cmdtest.go
type Env struct {
    IO             *iostreams.IOStreams
    Stdout, Stderr *bytes.Buffer
    Factory        *cmdutil.Factory
    Worktrees      *FakeWorktrees // hand-written, state-based; satisfies each command's consumer interface
}

func New(t *testing.T) *Env {
    t.Helper()
    ios, _, stdout, stderr := iostreams.Test()
    return &Env{IO: ios, Stdout: stdout, Stderr: stderr, Worktrees: &FakeWorktrees{},
        Factory: &cmdutil.Factory{
            IOStreams: ios,
            Config:    func() (*config.AppConfig, error) { return config.DefaultConfig(), nil },
        }}
}
```

Parse-level tests pass `env.Factory` to `NewCmdX`; run-level tests put `env.IO` and `env.Worktrees` into the options. Make it a package, not a `_test.go` file in `cmd/`: test files cannot be imported, so Tier 3 command packages could not share them.

## What to test per tier

| Tier | Core | Command | Integration | E2E |
| --- | --- | --- | --- | --- |
| 1 | `run()` table tests with buffers | — | Golden files on `run()` | Optional |
| 2 | Core unit tests | `runF` + golden output | Adapters | Exit codes for misuse |
| 3 | Core unit tests | Per command package | Component boundaries | Binary suite |
